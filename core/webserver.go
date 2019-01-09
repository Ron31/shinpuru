package core

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/zekroTJA/shinpuru/util"
)

type WebServer struct {
	router *mux.Router
	config *Config
}

func (ws *WebServer) registerHandlers() {
	discordOAuth := util.NewDiscordOAuth(
		ws.config.WebServer.ClientID,
		ws.config.WebServer.ClientSecret,
		ws.config.WebServer.Domain+"/authorize")

	ws.router.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		discordOAuth.RedirectToAuth(w, r)
	})

	ws.router.HandleFunc("/authorize", func(w http.ResponseWriter, r *http.Request) {
		code := r.FormValue("code")
		if code == "" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(http.StatusText(http.StatusUnauthorized)))
			return
		}
		uid, err := discordOAuth.GetUserID(code))
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(http.StatusText(http.StatusUnauthorized)))
			return
		}
	})
}

func NewWebServer(config *Config) error {
	ws := &WebServer{
		router: mux.NewRouter(),
		config: config,
	}
	ws.registerHandlers()
	return http.ListenAndServe(":"+config.WebServer.Port, ws.router)
}
