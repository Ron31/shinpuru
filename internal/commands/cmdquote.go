package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/zekroTJA/shinpuru/internal/util"
	"github.com/zekroTJA/shinpuru/internal/util/static"
	"github.com/zekroTJA/shinpuru/pkg/discordutil"
)

type CmdQuote struct {
}

func (c *CmdQuote) GetInvokes() []string {
	return []string{"quote", "q"}
}

func (c *CmdQuote) GetDescription() string {
	return "quote a message from any chat"
}

func (c *CmdQuote) GetHelp() string {
	return "`quote <msgID/msgURL> (<comment>)`"
}

func (c *CmdQuote) GetGroup() string {
	return GroupChat
}

func (c *CmdQuote) GetDomainName() string {
	return "sp.chat.quote"
}

func (c *CmdQuote) GetSubPermissionRules() []SubPermission {
	return nil
}

func (c *CmdQuote) IsExecutableInDMChannels() bool {
	return false
}

func (c *CmdQuote) Exec(args *CommandArgs) error {
	if len(args.Args) < 1 {
		return util.SendEmbedError(args.Session, args.Channel.ID,
			"Please enter a message ID or URL which should be quoted.").
			DeleteAfter(8 * time.Second).Error()
	}

	if strings.HasPrefix(args.Args[0], "https://discordapp.com/channels/") ||
		strings.HasPrefix(args.Args[0], "https://discord.com/channels/") ||
		strings.HasPrefix(args.Args[0], "https://canary.discordapp.com/channels/") ||
		strings.HasPrefix(args.Args[0], "https://canary.discord.com/channels/") {
		urlSplit := strings.Split(args.Args[0], "/")
		args.Args[0] = urlSplit[len(urlSplit)-1]
	}

	comment := strings.Join(args.Args[1:], " ")

	msgSearchEmb := &discordgo.MessageEmbed{
		Color:       static.ColorEmbedGray,
		Description: fmt.Sprintf(":hourglass_flowing_sand:  Searching for message in channel <#%s>...", args.Channel.ID),
	}

	msgSearch, err := args.Session.ChannelMessageSendEmbed(args.Channel.ID, msgSearchEmb)
	if err != nil {
		return err
	}

	var textChans []*discordgo.Channel
	for _, ch := range args.Guild.Channels {
		if ch.Type == discordgo.ChannelTypeGuildText {
			textChans = append(textChans, ch)
		}
	}

	loopLen := len(textChans)
	results := make(chan *discordgo.Message, loopLen)
	timeout := make(chan bool, 1)
	timeOuted := false
	quoteMsg, _ := args.Session.ChannelMessage(args.Channel.ID, args.Args[0])
	if quoteMsg == nil {
		msgSearchEmb.Description = ":hourglass_flowing_sand:  Searching for message in other channels..."
		args.Session.ChannelMessageEditEmbed(args.Channel.ID, msgSearch.ID, msgSearchEmb)
		for _, ch := range textChans {
			go func(c *discordgo.Channel) {
				quoteMsg, _ := args.Session.ChannelMessage(c.ID, args.Args[0])
				results <- quoteMsg
			}(ch)
		}
		timeoutTimer := time.AfterFunc(10*time.Second, func() {
			timeout <- true
		})
		func() {
			i := 0
			for {
				select {
				case fmsg := <-results:
					i++
					if i >= loopLen {
						return
					}
					if fmsg != nil {
						quoteMsg = fmsg
						timeoutTimer.Stop()
						return
					}
				case <-timeout:
					timeOuted = true
					return
				}
			}
		}()
	}

	if timeOuted {
		msgSearchEmb.Description = "Searching worker timeout."
		msgSearchEmb.Color = static.ColorEmbedError
		_, err := args.Session.ChannelMessageEditEmbed(args.Channel.ID, msgSearch.ID, msgSearchEmb)
		discordutil.DeleteMessageLater(args.Session, msgSearch, 5*time.Second)
		return err
	}

	if quoteMsg == nil {
		msgSearchEmb.Description = "Could not find any message with this ID. :disappointed:"
		msgSearchEmb.Color = static.ColorEmbedError
		_, err := args.Session.ChannelMessageEditEmbed(args.Channel.ID, msgSearch.ID, msgSearchEmb)
		discordutil.DeleteMessageLater(args.Session, msgSearch, 5*time.Second)
		return err
	}

	if len(quoteMsg.Content) < 1 && len(quoteMsg.Attachments) < 1 {
		msgSearchEmb.Description = "Found messages content is empty. Maybe, it is an embed message itself, which can not be quoted."
		msgSearchEmb.Color = static.ColorEmbedError
		_, err := args.Session.ChannelMessageEditEmbed(args.Channel.ID, msgSearch.ID, msgSearchEmb)
		discordutil.DeleteMessageLater(args.Session, msgSearch, 8*time.Second)
		return err
	}

	quoteMsgChannel, err := args.Session.Channel(quoteMsg.ChannelID)
	if err != nil {
		return err
	}

	msgSearchEmb = &discordgo.MessageEmbed{
		Color: static.ColorEmbedDefault,
		Author: &discordgo.MessageEmbedAuthor{
			IconURL: quoteMsg.Author.AvatarURL(""),
			Name:    quoteMsg.Author.Username + "#" + quoteMsg.Author.Discriminator,
		},
		Description: quoteMsg.Content +
			fmt.Sprintf("\n\n*[jump to message](%s)*", discordutil.GetMessageLink(quoteMsg, args.Guild.ID)),
		Footer: &discordgo.MessageEmbedFooter{
			IconURL: args.User.AvatarURL("16"),
			Text:    fmt.Sprintf("#%s - quoted by: %s#%s", quoteMsgChannel.Name, args.User.Username, args.User.Discriminator),
		},
		Timestamp: string(quoteMsg.Timestamp),
	}

	if len(quoteMsg.Attachments) > 0 {
		att := quoteMsg.Attachments[0]
		msgSearchEmb.Image = &discordgo.MessageEmbedImage{
			URL:      att.URL,
			ProxyURL: att.ProxyURL,
			Height:   att.Height,
			Width:    att.Width,
		}
	}

	if comment != "" {
		args.Session.ChannelMessageEdit(args.Channel.ID, msgSearch.ID,
			fmt.Sprintf("**%s:**\n%s\n", args.User.String(), comment))
	}

	args.Session.ChannelMessageEditEmbed(args.Channel.ID, msgSearch.ID, msgSearchEmb)
	return nil
}
