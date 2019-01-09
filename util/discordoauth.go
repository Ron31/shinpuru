package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

type DiscordOAuth struct {
	ClientID     string
	ClientSecret string
	CallbackURI  string
}

func NewDiscordOAuth(clientID, clientSecret, callbackURI string) *DiscordOAuth {
	return &DiscordOAuth{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		CallbackURI:  callbackURI,
	}
}

func (d *DiscordOAuth) RedirectToAuth(w http.ResponseWriter, r *http.Request) {
	redirectURL := url.QueryEscape(d.CallbackURI)
	url := fmt.Sprintf(
		"https://discordapp.com/api/oauth2/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=identify",
		d.ClientID, redirectURL)
	http.Redirect(w, r,
		url, http.StatusSeeOther)
}

func (d *DiscordOAuth) GetUserID(authCode string) (string, error) {
	payload := url.Values{
		"client_id":     {d.ClientID},
		"client_secret": {d.ClientSecret},
		"grant_type":    {"authorization_code"},
		"code":          {authCode},
		"redirect_uri":  {d.CallbackURI},
		"scope":         {"identify"},
	}

	resp, err := http.PostForm("https://discordapp.com/api/oauth2/token", payload)
	if err != nil {
		return "", err
	}

	var res map[string]interface{}

	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&res)
	if err != nil {
		return "", err
	}

	accessToken, ok := res["access_token"]
	if !ok {
		return "", errors.New("unauthorized")
	}

	req, err := http.NewRequest("GET", "https://discordapp.com/api/users/@me", nil)
	if err != nil {
		return "", err
	}

	req.Header.Add("Authorization", "Bearer "+accessToken.(string))

	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		return "", err
	}

	dec = json.NewDecoder(resp.Body)
	err = dec.Decode(&res)
	if err != nil {
		return "", err
	}

	id, ok := res["id"]
	if err != nil {
		return "", errors.New("unauthorized")
	}
	return id.(string), nil
}
