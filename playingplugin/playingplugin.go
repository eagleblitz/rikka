package playingplugin

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/ThyLeader/rikka"
)

type playingPlugin struct {
	rikka.SimplePlugin
	Game string
	URL  string
}

// Name returns the name of the plugin.
func (p *playingPlugin) Name() string {
	return "Playing"
}

// Load will load plugin state from a byte array.
func (p *playingPlugin) Load(bot *rikka.Bot, service rikka.Service, data []byte) error {
	if data != nil {
		if err := json.Unmarshal(data, p); err != nil {
			log.Println("Error loading data", err)
		}
	}

	if p.Game != "" {
		err := service.(*rikka.Discord).Session.UpdateStreamingStatus(0, p.Game, p.URL)
		if err != nil {
			fmt.Println(err.Error())
		}
	} else {
		err := service.(*rikka.Discord).Session.UpdateStatus(0, p.Game)
		if err != nil {
			fmt.Println(err.Error())
		}
	}

	go p.Persist(bot, service)

	return nil
}

func (p *playingPlugin) Persist(bot *rikka.Bot, service rikka.Service) {
	t := time.Tick(1 * time.Hour)

	for {
		select {
		case <-t:
			if p.Game != "" {
				err := service.(*rikka.Discord).Session.UpdateStreamingStatus(0, p.Game, p.URL)
				if err != nil {
					fmt.Println(err.Error())
				}
			} else {
				err := service.(*rikka.Discord).Session.UpdateStatus(0, p.Game)
				if err != nil {
					fmt.Println(err.Error())
				}
			}
		}
	}
}

// Save will save plugin state to a byte array.
func (p *playingPlugin) Save() ([]byte, error) {
	return json.Marshal(p)
}

// Help returns a list of help strings that are printed when the user requests them.
func (p *playingPlugin) helpFunc(bot *rikka.Bot, service rikka.Service, message rikka.Message, detailed bool) []string {
	if detailed {
		return nil
	}

	if !service.IsBotOwner(message) {
		return nil
	}

	return rikka.CommandHelp(service, "playing", "<game>, <url>", fmt.Sprintf("Set which game %s is playing.", service.UserName()))
}

// Message handler.
func (p *playingPlugin) messageFunc(bot *rikka.Bot, service rikka.Service, message rikka.Message) {
	if service.IsMe(message) {
		return
	}

	if !rikka.MatchesCommand(service, "playing", message) {
		return
	}

	if !service.IsBotOwner(message) {
		fmt.Println("not owner")
		return
	}

	query, _ := rikka.ParseCommand(service, message)

	split := strings.Split(query, ",")

	p.Game = strings.Trim(split[0], " ")
	if len(split) > 1 {
		p.URL = strings.Trim(split[1], " ")
		err := service.(*rikka.Discord).Session.UpdateStreamingStatus(0, p.Game, p.URL)
		if err != nil {
			fmt.Println(err.Error())
		}
	} else {
		err := service.(*rikka.Discord).Session.UpdateStatus(0, p.Game)
		if err != nil {
			fmt.Println(err.Error())
		}
	}
}

// New will create a new top streamers plugin.
func New() rikka.Plugin {
	p := &playingPlugin{}
	p.MessageFunc = p.messageFunc
	p.HelpFunc = p.helpFunc
	return p
}
