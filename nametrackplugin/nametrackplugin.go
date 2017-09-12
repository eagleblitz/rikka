package nametrackplugin

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/ThyLeader/rikka"
	"github.com/bwmarrin/discordgo"
)

type nameTrackPlugin struct {
	sync.Mutex

	Names map[string][]string
}

func (p *nameTrackPlugin) Load(bot *rikka.Bot, service rikka.Service, data []byte) error {
	if data != nil {
		if err := json.Unmarshal(data, p); err != nil {
			log.Println("Error loading data", err)
			return err
		}
	}
	go p.Run(bot, service)
	return nil
}

func (p *nameTrackPlugin) Message(bot *rikka.Bot, service rikka.Service, message rikka.Message) {
	if service.IsMe(message) {
		return
	}

	if !rikka.MatchesCommand(service, "names", message) {
		return
	}

	_, parts := rikka.ParseCommand(service, message)

	var user string
	if len(parts) < 1 {
		user = message.UserID()
	} else {
		if parts[0] == "scan" {
			p.testScan(message.Guild(), service)
			service.SendMessage(message.Channel(), "scanned guild "+message.Guild())
			return
		}
		user = parts[0]
	}

	u, ok := p.Names[user]
	if !ok {
		service.SendMessage(message.Channel(), fmt.Sprintf("User `%s` not found", user))
		return
	}
	service.SendMessage(message.Channel(), strings.Join(u, ", "))
}

func (p *nameTrackPlugin) Run(bot *rikka.Bot, service rikka.Service) {
	discord := service.(*rikka.Discord)

	discord.Session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		for _, g := range r.Guilds {
			fmt.Println("ready")
			for i, m := range g.Members {
				fmt.Println(i)
				p.update(m.User)
			}
		}
	})

	discord.Session.AddHandler(func(s *discordgo.Session, g *discordgo.GuildCreate) {
		if g.Unavailable {
			return
		}
		for _, m := range g.Members {
			p.update(m.User)
		}
	})

	discord.Session.AddHandler(func(s *discordgo.Session, r *discordgo.GuildMemberUpdate) {
		p.update(r.User)
	})

	discord.Session.AddHandler(func(s *discordgo.Session, r *discordgo.GuildMemberAdd) {
		p.update(r.User)
	})

	discord.Session.AddHandler(func(s *discordgo.Session, r *discordgo.UserUpdate) {
		p.update(r.User)
	})

	discord.Session.AddHandler(func(s *discordgo.Session, r *discordgo.PresenceUpdate) {
		p.update(r.User)
	})
}

func (p *nameTrackPlugin) testScan(gID string, service rikka.Service) {
	discord := service.(*rikka.Discord)

	g, _ := discord.Session.Guild(gID)

	for _, m := range g.Members {
		p.update(m.User)
	}
}

func (p *nameTrackPlugin) update(u *discordgo.User) {
	if u.Username == "" {
		return
	}
	_, ok := p.Names[u.ID]
	if !ok {
		p.Lock()
		p.Names[u.ID] = []string{u.Username}
		p.Unlock()
		fmt.Println("Updated " + u.Username)
		return
	}
	if !searchNicks(p.Names[u.ID], u.Username) {
		p.Lock()
		p.Names[u.ID] = append(p.Names[u.ID], u.Username)
		p.Unlock()
		fmt.Println("Updated " + u.Username)
		return
	}
}

// true if current nickname was already recorded, false if not
func searchNicks(nicks []string, c string) bool {
	if len(nicks) == 0 {
		return false
	}

	for _, n := range nicks {
		if n == c {
			return true
		}
	}
	return false
}

func (p *nameTrackPlugin) Help(bot *rikka.Bot, service rikka.Service, message rikka.Message, detailed bool) []string {
	if detailed {
		return nil
	}
	return rikka.CommandHelp(service, "pubg", "nick", "")
}

func (p *nameTrackPlugin) Name() string {
	return "NameTrack"
}

func (p *nameTrackPlugin) Save() ([]byte, error) {
	return json.Marshal(p)
}

func (p *nameTrackPlugin) Stats(bot *rikka.Bot, service rikka.Service, message rikka.Message) []string {
	return nil
}

// New creates a new discordavatar plugin.
func New() rikka.Plugin {
	return &nameTrackPlugin{
		Names: map[string][]string{},
	}
}
