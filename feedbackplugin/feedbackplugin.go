package feedbackplugin

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/ThyLeader/rikka"
)

type feedbackPlugin struct {
	sync.Mutex

	Banned map[string]bool
}

func (p *feedbackPlugin) Load(bot *rikka.Bot, service rikka.Service, data []byte) error {
	if data != nil {
		if err := json.Unmarshal(data, p); err != nil {
			log.Println("Error loading data", err)
			return err
		}
	}

	return nil
}

func (p *feedbackPlugin) Ban(uID string) error {
	p.Lock()
	b, ok := p.Banned[uID]
	if !ok {
		p.Banned[uID] = true
	} else {
		if b {
			p.Unlock()
			return errors.New("User already banned")
		}
		p.Banned[uID] = true
	}
	p.Unlock()
	return nil
}

func (p *feedbackPlugin) Unban(uID string) error {
	p.Lock()
	b, ok := p.Banned[uID]
	if !ok {
		p.Banned[uID] = false
	} else {
		if !b {
			p.Unlock()
			return errors.New("User isn't banned")
		}
		p.Banned[uID] = false
	}
	p.Unlock()
	return nil
}

func (p *feedbackPlugin) Message(bot *rikka.Bot, service rikka.Service, message rikka.Message) {
	if service.IsMe(message) {
		return
	}

	if !rikka.MatchesCommand(service, "feedback", message) {
		return
	}

	m, parts := rikka.ParseCommand(service, message)
	if service.IsBotOwner(message) {
		if len(parts) > 0 {
			switch parts[0] {
			case "ban":
				if len(parts) > 1 {
					err := p.Ban(parts[1])
					if err != nil {
						service.SendMessage(message.Channel(), "Unable to ban user - "+err.Error())
					}
					service.SendMessage(message.Channel(), "Banned user "+parts[1])
					return
				}
				service.SendMessage(message.Channel(), "supply a userid you idiot")
				return
			case "unban":
				if len(parts) > 1 {
					err := p.Unban(parts[1])
					if err != nil {
						service.SendMessage(message.Channel(), "Unable to unban user - "+err.Error())
					}
					service.SendMessage(message.Channel(), "Unbanned user "+parts[1])
					return
				}
				service.SendMessage(message.Channel(), "supply a userid you idiot")
				return
			default:
				// service.SendMessage(message.Channel(), "tf are u tryna do")
				// return
			}
		}
	}

	p.Lock()
	b, ok := p.Banned[message.UserID()]
	p.Unlock()
	if ok {
		if b {
			service.SendMessage(message.Channel(), "Sorry, but you are banned from sending feedback because of abuse")
			return
		}
	}
	service.SendMessage("359902628055875585", fmt.Sprintf("%s (`%s`) left some feedback from guild %s (`%s`)\n```%s```", message.User().Username, message.UserID(), message.GuildName(), message.GuildID(), m))
	service.SendMessage(message.Channel(), "Feedback left!\nOur support server is here: <https://rikka.xyz>")
}

func (p *feedbackPlugin) Help(bot *rikka.Bot, service rikka.Service, message rikka.Message, detailed bool) []string {
	if detailed {
		return nil
	}
	return rikka.CommandHelp(service, "feedback", "<constructive criticism>", "Sends a message to the devs with your thoughts")
}

func (p *feedbackPlugin) Name() string {
	return "Feedback"
}

func (p *feedbackPlugin) Save() ([]byte, error) {
	return json.Marshal(p)
}

func (p *feedbackPlugin) Stats(bot *rikka.Bot, service rikka.Service, message rikka.Message) []string {
	return nil
}

// New creates a new discordavatar plugin.
func New() rikka.Plugin {
	return &feedbackPlugin{
		Banned: make(map[string]bool, 0),
	}
}
