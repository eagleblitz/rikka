package callbacktesting

import (
	"encoding/json"

	"github.com/ThyLeader/rikka"
)

type nameTrackPlugin struct{}

func (p *nameTrackPlugin) Load(bot *rikka.Bot, service rikka.Service, data []byte) error {
	return nil
}

func (p *nameTrackPlugin) Message(bot *rikka.Bot, service rikka.Service, message rikka.Message) {
	if service.IsMe(message) {
		return
	}

	if !rikka.MatchesCommand(service, "cb", message) {
		return
	}
	service.SendMessage(message.Channel(), "starting callback")

	m := bot.MakeCallback(service, message.UserID())

	for {
		ms := <-m
		service.SendMessage(message.Channel(), ms.Message())
	}
}

func (p *nameTrackPlugin) Help(bot *rikka.Bot, service rikka.Service, message rikka.Message, detailed bool) []string {
	if detailed {
		return nil
	}
	return rikka.CommandHelp(service, "", "", "")
}

func (p *nameTrackPlugin) Name() string {
	return "callbacks"
}

func (p *nameTrackPlugin) Save() ([]byte, error) {
	return json.Marshal(p)
}

func (p *nameTrackPlugin) Stats(bot *rikka.Bot, service rikka.Service, message rikka.Message) []string {
	return nil
}

// New creates a new discordavatar plugin.
func New() rikka.Plugin {
	return &nameTrackPlugin{}
}
