package seenplugin

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"

	"github.com/ThyLeader/rikka"
	"github.com/go-redis/redis"
)

type seenPlugin struct {
}

var client *redis.Client

func (p *seenPlugin) Load(bot *rikka.Bot, service rikka.Service, data []byte) error {
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

	return nil
}

func (p *seenPlugin) Message(bot *rikka.Bot, service rikka.Service, message rikka.Message) {
	go p.Update(bot, service, message)
	if service.IsMe(message) {
		return
	}

	if !rikka.MatchesCommand(service, "seen", message) && !rikka.MatchesCommand(service, "lastseen", message) {
		return
	}

	mentions := message.Mentions()
	if len(mentions) > 1 {
		service.SendMessage(message.Channel(), "Please only mention one user")
		return
	}
	var id, name string
	var mentionedUser *discordgo.User
	if len(mentions) == 1 {
		mentionedUser = mentions[0]
		id = mentionedUser.ID
		name = mentionedUser.Username
	}

	if len(mentions) == 0 {
		_, parts := rikka.ParseCommand(service, message)
		switch len(parts) {
		case 1:
			id = parts[0]
			m, err := service.Member(message.GuildID(), id)
			if err != nil {
				service.SendMessage(message.Channel(), "There was an error!\n"+err.Error())
				log.Println(err.Error())
				return
			}
			mentionedUser = m.User
			name = mentionedUser.Username
		case 0:
			id = message.UserID()
			name = message.UserName()
		default:
			service.SendMessage(message.Channel(), "Please only search for one user at a time")
			return
		}
	}

	if id == "" || name == "" {
		service.SendMessage(message.Channel(), "tell thy he dun screwed up")
		return
	}

	lastSeen := p.getLastSeen(message.GuildID(), id)
	if lastSeen == "" {
		service.SendMessage(message.Channel(), fmt.Sprintf("%s (`%s`) has not sent a message yet or there was an error", name, id))
		return
	}

	service.SendMessage(message.Channel(), fmt.Sprintf("%s was last seen here %s", name, lastSeen))
}

func seenKey(gID string, uID string) string {
	return fmt.Sprintf("seen:%s:%s", gID, uID)
}

func (p *seenPlugin) Update(bot *rikka.Bot, service rikka.Service, message rikka.Message) {
	err := client.Set(seenKey(message.GuildID(), message.UserID()), time.Now().Format(time.UnixDate), 0).Err()
	if err != nil {
		fmt.Println("Error updating seen - ", err.Error())
	}
}

func (p *seenPlugin) getLastSeen(gID, uID string) string {
	s := client.Get(seenKey(gID, uID)).Val()
	if s == "" {
		return s
	}
	t, err := time.Parse(time.UnixDate, s)
	if err != nil {
		fmt.Println("Error parsing last seen - ", err.Error())
		return ""
	}
	return humanize.Time(t)
}

func (p *seenPlugin) Help(bot *rikka.Bot, service rikka.Service, message rikka.Message, detailed bool) []string {
	if detailed {
		return nil
	}
	return rikka.CommandHelp(service, "seen", "[@username]", "See the last time a user has typed in this guild")
}

func (p *seenPlugin) Name() string {
	return "Seen"
}

func (p *seenPlugin) Save() ([]byte, error) {
	return json.Marshal(p)
}

func (p *seenPlugin) Stats(bot *rikka.Bot, service rikka.Service, message rikka.Message) []string {
	return nil
}

// New creates a new discordavatar plugin.
func New() rikka.Plugin {
	return &seenPlugin{}
}
