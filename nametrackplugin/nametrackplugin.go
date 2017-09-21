package nametrackplugin

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"

	"github.com/ThyLeader/rikka"
	"github.com/bwmarrin/discordgo"
	"github.com/go-redis/redis"
)

type nameTrackPlugin struct {
	sync.Mutex

	Names map[string][]string
}

var client *redis.Client

func (p *nameTrackPlugin) Load(bot *rikka.Bot, service rikka.Service, data []byte) error {
	if data != nil {
		if err := json.Unmarshal(data, p); err != nil {
			log.Println("Error loading data", err)
			return err
		}
	}
	client = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	go p.Run(bot, service)
	return nil
}

func (p *nameTrackPlugin) Message(bot *rikka.Bot, service rikka.Service, message rikka.Message) {
	if service.IsMe(message) {
		return
	}

	if !rikka.MatchesCommand(service, "names", message) && !rikka.MatchesCommand(service, "nicks", message) {
		return
	}

	_, parts := rikka.ParseCommand(service, message)

	var user string
	if len(parts) < 1 {
		user = message.UserID()
	} else {
		if parts[0] == "scan" {
			if !service.IsBotOwner(message) {
				service.SendMessage(message.Channel(), "Only the bot owner can use this feature")
				return
			}
			p.testScan(message.GuildID(), service)
			service.SendMessage(message.Channel(), "scanned guild "+message.GuildID())
			return
		}
		if parts[0] == "scanall" {
			if !service.IsBotOwner(message) {
				service.SendMessage(message.Channel(), "Only the bot owner can use this feature")
				return
			}
			p.scanAll(service)
			service.SendMessage(message.Channel(), "scanning all..")
			return
		}

		userIDRegex := regexp.MustCompile("<@!?([0-9]*)>")
		query := strings.Join(strings.Split(message.RawMessage(), " ")[1:], " ")
		if m := userIDRegex.FindStringSubmatch(query); m != nil {
			user = m[1]
		}

		if user == "" {
			user = parts[0]
		}
	}

	u := client.SMembers("names:" + user).Val()
	if len(u) < 1 {
		service.SendMessage(message.Channel(), fmt.Sprintf("User `%s` not found\nPlease use the user's ID or mention them. Username searches coming soon:tm:", user))
		return
	}
	/*embed := &discordgo.MessageEmbed{
		Title:       title,
		Description: fmt.Sprintf("*First seen %s, last update %s*", humanize.Time(u.FirstSeen), lc),
		Fields: []*discordgo.MessageEmbedField{
			&discordgo.MessageEmbedField{Name: "Games", Value: statuses, Inline: false},
		},
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: url,
		},
		Color: 0x79c879,
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Data valid as of %s", time.Now().Format(time.UnixDate)),
		},
	}*/
	service.SendMessage(message.Channel(), strings.Join(u, ", "))
}

func (p *nameTrackPlugin) Run(bot *rikka.Bot, service rikka.Service) {
	discord := service.(*rikka.Discord)

	discord.Session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		fmt.Println("ready")
		for _, g := range r.Guilds {
			for _, m := range g.Members {
				p.update(m.User, m.Nick)
			}
		}

		p.scanAll(service)
	})

	discord.Session.AddHandler(func(s *discordgo.Session, g *discordgo.GuildCreate) {
		if g.Unavailable {
			return
		}
		for _, m := range g.Members {
			p.update(m.User, m.Nick)
		}
	})

	discord.Session.AddHandler(func(s *discordgo.Session, r *discordgo.GuildMemberUpdate) {
		p.update(r.User, r.Nick)
	})

	discord.Session.AddHandler(func(s *discordgo.Session, r *discordgo.GuildMemberAdd) {
		p.update(r.User, r.Nick)
	})

	discord.Session.AddHandler(func(s *discordgo.Session, r *discordgo.UserUpdate) {
		p.update(r.User, "")
	})

	discord.Session.AddHandler(func(s *discordgo.Session, r *discordgo.PresenceUpdate) {
		p.update(r.User, r.Nick)
	})
}

func (p *nameTrackPlugin) testScan(gID string, service rikka.Service) {
	discord := service.(*rikka.Discord)

	g, _ := discord.Session.Guild(gID)

	for _, m := range g.Members {
		p.update(m.User, m.Nick)
	}
}

func (p *nameTrackPlugin) scanAll(service rikka.Service) {
	discord := service.(*rikka.Discord)

	fmt.Println("scanning all")
	for _, g := range discord.Session.State.Ready.Guilds {
		if g.Unavailable {
			fmt.Println("guild unavailable")
			continue
		}
		for _, m := range g.Members {
			p.update(m.User, m.Nick)
		}
	}
}

func (p *nameTrackPlugin) update(u *discordgo.User, nick string) {
	if u.Username != "" {
		client.SAdd("names:"+u.ID, u.Username)
	}

	if nick != "" {
		client.SAdd("nicks:"+u.ID, nick)
	}

	// p.Lock()
	// n, ok := p.Names[u.ID]
	// if !ok {
	// 	p.Names[u.ID] = []string{u.Username}
	// 	fmt.Println("update new " + u.Username)
	// 	return
	// }
	// if !searchNicks(n, u.Username) {
	// 	p.Names[u.ID] = append(p.Names[u.ID], u.Username)
	// 	fmt.Println("update changed " + u.Username)
	// 	return
	// }
	// p.Unlock()
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
	return rikka.CommandHelp(service, "names", "[@username]", "See a user's past usernames")
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
