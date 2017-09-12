package pubgplugin

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"

	"github.com/ThyLeader/rikka"
	"github.com/albshin/go-pubg"
)

type pubgPlugin struct {
	sync.Mutex

	Client    *pubg.API
	Nicknames map[string]*userData
}

type player pubg.Player

type userData struct {
	PubgNick string
	SteamID  string
}

func (p *pubgPlugin) store(uid string, data *userData) error {
	p.Lock()
	p.Nicknames[uid] = data
	p.Unlock()
	return nil
}

func (p *pubgPlugin) check(uid string) (*userData, bool) {
	p.Lock()
	u, ok := p.Nicknames[uid]
	p.Unlock()
	if !ok {
		return nil, false
	}
	return u, true
}

func (p *pubgPlugin) Load(bot *rikka.Bot, service rikka.Service, data []byte) error {
	if data != nil {
		if err := json.Unmarshal(data, p); err != nil {
			log.Println("Error loading data", err)
			return err
		}
	}
	return nil
}

func (p *pubgPlugin) Message(bot *rikka.Bot, service rikka.Service, message rikka.Message) {
	if service.IsMe(message) {
		return
	}

	if !rikka.MatchesCommand(service, "pubg", message) {
		return
	}

	_, parts := rikka.ParseCommand(service, message)
	service.Typing(message.Channel())

	if len(parts) == 0 {
		u, err := p.search(message.UserID())
		if err != nil {
			service.SendMessage(message.Channel(), fmt.Sprintf("You have not set your PUBG nickname yet. Please try `%shelp pubg`", service.CommandPrefix()))
			return
		}
		s, err := p.nickSearch(u)
		if err != nil {
			service.SendMessage(message.Channel(), err.Error())
			return
		}
		service.SendMessage(message.Channel(), fmt.Sprintf("Nick: %s\n%+v", s.PlayerName, s.Stats[3].Stats))
		return
	}

	switch parts[0] {
	case "set":
		if len(parts) < 2 {
			service.SendMessage(message.Channel(), "Enter the PUBG nickname you would like to set\ne.g. `pubg set Summit1g`")
			return
		}
		p.store(message.UserID(), &userData{parts[1], ""})
		service.SendMessage(message.Channel(), fmt.Sprintf("Successfully set your PUBG nickname to `%s`", parts[1]))
		return
	case "check":
		u, ok := p.check(message.UserID())
		if !ok {
			service.SendMessage(message.Channel(), fmt.Sprintf("You have not set your PUBG nickname. To set it, type `%spubg set <nickname>`", service.CommandPrefix()))
			return
		}
		if n, _ := u.check(); n != "" {
			service.SendMessage(message.Channel(), fmt.Sprintf("Your PUBG nickname is currently set to `%s`", n))
			return
		}
		service.SendMessage(message.Channel(), fmt.Sprintf("You have not set your PUBG nickname. To set it, type `%spubg set <nickname>`", service.CommandPrefix()))
		return
	default:
		if len(parts) < 1 {
			u, ok := p.check(message.UserID())
			if !ok {
				service.SendMessage(message.Channel(), fmt.Sprintf("You have not set your PUBG nickname. To set it, type `%spubg set <nickname>`", service.CommandPrefix()))
				return
			}
			uC, t := u.check()
			if t != -1 {
				var d *pubg.Player
				var err error
				if t == 0 {
					d, err = p.nickSearch(uC)
				} else if t == 1 {
					d, err = p.steamSearch(uC)
				}

				if err != nil {
					service.SendMessage(message.Channel(), fmt.Sprintf("There was an error retrieving the statistics for %s\n`%s`", uC, err.Error()))
					return
				}
				service.SendMessage(message.Channel(), fmt.Sprintf("Name: %s \n%+v", d.PlayerName, d.Stats))
				return
			}

		}
		var d *pubg.Player
		var err, errP error
		_, err = strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			d, errP = p.nickSearch(parts[0])
		} else {
			d, errP = p.steamSearch(parts[0])
		}
		if errP != nil {
			service.SendMessage(message.Channel(), fmt.Sprintf("There was an error retrieving the statistics for %s\n`%s`", parts[0], err.Error()))
			return
		}
		service.SendMessage(message.Channel(), fmt.Sprintf("ID: %s\nName: %s", d.AccountID, d.PlayerName))
		return
	}
}

type newStats struct {
	Solo, Duo, Squad newStat
}

type newStat struct {
	Region string `json:"Region"`
	Season string `json:"Season"`
	Match  string `json:"Match"`
	Stats  []struct {
		Label        string      `json:"label"`
		Field        string      `json:"field"`
		Category     string      `json:"category"`
		ValueInt     interface{} `json:"ValueInt"`
		ValueDec     float64     `json:"ValueDec"`
		Value        string      `json:"value"`
		Rank         interface{} `json:"rank"`
		Percentile   int         `json:"percentile"`
		DisplayValue string      `json:"displayValue"`
	} `json:"Stats"`
}

func parse(p pubg.Player) string {
	var s1 []int
	for i := range p.Stats {
		if p.Stats[i].Region == "agg" {
			s1 = append(s1, i)
		}
	}

	var stats []string
	stats = append(stats, []string{
		fmt.Sprintf(""),
	}...)

	return strings.Join(stats, "\n")
}

func (p *pubgPlugin) nickSearch(uname string) (*pubg.Player, error) {
	return p.Client.GetPlayer(uname)
}

func (p *pubgPlugin) steamSearch(uname string) (*pubg.Player, error) {
	s, err := p.Client.GetSteamInfo(uname)
	if err != nil {
		return nil, err
	}
	return p.nickSearch(s.Nickname)
}

func (p *pubgPlugin) search(uid string) (string, error) {
	u, ok := p.check(uid)
	if !ok {
		return "", errors.New("User not set")
	}
	userC, _ := u.check()
	return userC, nil
}

func (u *userData) check() (string, int) {
	if u.PubgNick != "" {
		return u.PubgNick, 0
	} else if u.SteamID != "" {
		return u.SteamID, 1
	} else {
		return "", -1
	}
}

func (p *pubgPlugin) Help(bot *rikka.Bot, service rikka.Service, message rikka.Message, detailed bool) []string {
	if detailed {
		return nil
	}
	return rikka.CommandHelp(service, "pubg", "nick", "")
}

func (p *pubgPlugin) Name() string {
	return "PUBG"
}

func (p *pubgPlugin) Save() ([]byte, error) {
	return json.Marshal(p)
}

func (p *pubgPlugin) Stats(bot *rikka.Bot, service rikka.Service, message rikka.Message) []string {
	return nil
}

// New creates a new discordavatar plugin.
func New() rikka.Plugin {
	a, err := pubg.New("8075c8f3-b635-456c-961d-69287d4446c2")
	if err != nil {
		fmt.Println("error making new pubg api " + err.Error())
	}
	return &pubgPlugin{
		Nicknames: map[string]*userData{},
		Client:    a,
	}
}
