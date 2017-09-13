package imageplugin

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/ThyLeader/rikka"
)

type response struct {
	Path string `json:"path"`
	ID   string `json:"id"`
	Type string `json:"type"`
	Nsfw bool   `json:"nsfw"`
}

func messageFunc(bot *rikka.Bot, service rikka.Service, message rikka.Message) {
	if service.IsMe(message) {
		return
	}
	if rikka.MatchesCommand(service, "lewd", message) {
		service.Typing(message.Channel())
		var r response
		h, err := http.Get("https://rra.ram.moe/i/r?type=lewd")
		if err != nil {
			service.SendMessage(message.Channel(), "Error getting the image - "+err.Error())
			return
		}
		body, err := ioutil.ReadAll(h.Body)
		if err != nil {
			return
		}
		json.Unmarshal(body, &r)

		i, err := http.Get("https://rra.ram.moe" + r.Path)
		if err != nil {
			service.SendMessage(message.Channel(), "Error getting the image - "+err.Error())
			return
		}
		service.SendFile(message.Channel(), "lewd."+strings.Split(r.Path, ".")[1], i.Body)
		h.Body.Close()
		i.Body.Close()
		return
	}

	if rikka.MatchesCommand(service, "kiss", message) {
		service.Typing(message.Channel())
		var r response
		h, err := http.Get("https://rra.ram.moe/i/r?type=kiss")
		if err != nil {
			service.SendMessage(message.Channel(), "Error getting the image - "+err.Error())
			return
		}
		body, err := ioutil.ReadAll(h.Body)
		if err != nil {
			return
		}
		json.Unmarshal(body, &r)

		i, err := http.Get("https://rra.ram.moe" + r.Path)
		if err != nil {
			service.SendMessage(message.Channel(), "Error getting the image - "+err.Error())
			return
		}
		service.SendFile(message.Channel(), "lewd."+strings.Split(r.Path, ".")[1], i.Body)
		h.Body.Close()
		i.Body.Close()
		return
	}
}

func helpFunc(bot *rikka.Bot, service rikka.Service, message rikka.Message, detailed bool) []string {
	if detailed {
		return nil
	}
	return rikka.CommandHelp(service, "lewd", "", ":eyes:")
}

// New creates a new math plugin.
func New() rikka.Plugin {
	p := rikka.NewSimplePlugin("Images")
	p.MessageFunc = messageFunc
	p.HelpFunc = helpFunc
	return p
}
