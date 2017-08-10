package playedplugin

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/ThyLeader/rikka"
	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
	"github.com/go-redis/redis"
)

func init() {
	client := redis.NewClient(&redis.Options{
		Addr:     "192.168.1.121:6379",
		Password: "",
		DB:       0,
	})

	pong, err := client.Ping().Result()
	if err != nil {
		return
	}
	fmt.Println(pong)
}

type playedEntry struct {
	Name     string
	Duration time.Duration
}

type byDuration []*playedEntry

func (a byDuration) Len() int           { return len(a) }
func (a byDuration) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byDuration) Less(i, j int) bool { return a[i].Duration >= a[j].Duration }

type playedUser struct {
	sync.RWMutex
	Entries     map[string]*playedEntry
	Current     string
	LastChanged time.Time
	FirstSeen   time.Time
}

func (p *playedUser) Update(name string, now time.Time) {
	p.Lock()
	defer p.Unlock()
	if p.Current != "" {
		pe := p.Entries[p.Current]
		if pe == nil {
			pe = &playedEntry{
				Name: p.Current,
			}
			p.Entries[p.Current] = pe
		}
		pe.Duration += now.Sub(p.LastChanged)
	}

	p.Current = name
	p.LastChanged = now
}

type playedPlugin struct {
	sync.RWMutex
	Users map[string]*playedUser
}

// Load will load plugin state from a byte array.
func (p *playedPlugin) Load(bot *rikka.Bot, service rikka.Service, data []byte) error {
	if service.Name() != rikka.DiscordServiceName {
		panic("Played Plugin only supports Discord.")
	}

	if data != nil {
		if err := json.Unmarshal(data, p); err != nil {
			log.Println("Error loading data", err)
		}
	}

	go p.Run(bot, service)
	return nil
}

// Save will save plugin state to a byte array.
func (p *playedPlugin) Save() ([]byte, error) {
	p.Lock()
	defer p.Unlock()
	return json.Marshal(p)
}

func (p *playedPlugin) Update(user string, entry string) {
	p.Lock()
	defer p.Unlock()

	t := time.Now()

	u := p.Users[user]
	if u == nil {
		u = &playedUser{
			Entries:     map[string]*playedEntry{},
			Current:     entry,
			LastChanged: t,
			FirstSeen:   t,
		}
		p.Users[user] = u
	}
	u.Update(entry, t)
}

// Run is the background go routine that executes for the life of the plugin.
func (p *playedPlugin) Run(bot *rikka.Bot, service rikka.Service) {
	discord := service.(*rikka.Discord)

	discord.Session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		for _, g := range r.Guilds {
			for _, pu := range g.Presences {
				e := ""
				if pu.Game != nil {
					e = pu.Game.Name
				}
				p.Update(pu.User.ID, e)
			}
		}
	})

	discord.Session.AddHandler(func(s *discordgo.Session, g *discordgo.GuildCreate) {
		if g.Unavailable {
			return
		}

		for _, pu := range g.Presences {
			e := ""
			if pu.Game != nil {
				e = pu.Game.Name
			}
			p.Update(pu.User.ID, e)
		}

		// if s.Token == "Bot Mjc2MTIxNDk1MTIwNjQyMDQ4.C3cBnQ.rG3L8KQWVBRs0l3fYsDG_MbymOM" {
		// 	return
		// }

		t, err := g.JoinedAt.Parse()
		if err != nil {
			fmt.Println("fuccccc")
			return
		}
		if t.Before(time.Now().Add(-1 * time.Minute)) {
			fmt.Println("gotem coach")
			return
		}

		guildOwner, err := discord.Session.State.Member(g.ID, g.OwnerID)
		if err != nil {
			discord.Session.ChannelMessageSend("340326362432798720", "Unable to retrieve information on the guild owner")
			return
		}
		gc, _ := discordgo.Timestamp(guildOwner.JoinedAt).Parse()

		var userCount float32
		var botCount float32
		for _, e := range g.Members {
			if e.User.Bot {
				botCount++
			} else {
				userCount++
			}
		}
		percent := botCount / (userCount + botCount) * 100

		discord.Session.ChannelMessageSendEmbed("340326362432798720", &discordgo.MessageEmbed{
			Color: discord.UserColor(service.UserID(), "340326362432798720"),
			Title: "Rikka joined a guild",
			Fields: []*discordgo.MessageEmbedField{
				&discordgo.MessageEmbedField{Name: "Name", Value: g.Name, Inline: true},
				&discordgo.MessageEmbedField{Name: "ID", Value: g.ID, Inline: true},
				&discordgo.MessageEmbedField{Name: "Owner name", Value: guildOwner.User.Username + "#" + guildOwner.User.Discriminator, Inline: true},
				&discordgo.MessageEmbedField{Name: "Owner ID", Value: guildOwner.User.ID, Inline: true},
				&discordgo.MessageEmbedField{Name: "Users", Value: fmt.Sprintf("%v", userCount), Inline: true},
				&discordgo.MessageEmbedField{Name: "Bots", Value: fmt.Sprintf("%v", botCount), Inline: true},
				&discordgo.MessageEmbedField{Name: "Percent", Value: fmt.Sprintf("%v", int(percent)) + "%", Inline: true},
				&discordgo.MessageEmbedField{Name: "Created", Value: humanize.Time(gc), Inline: true},
			},
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: discordgo.EndpointGuildIcon(g.ID, g.Icon), //"https://discordapp.com/api/guilds/" + parts[0] + "/icons/" + guild.Icon + ".jpg",
			},
		})
		//service.SendMessage("286247106749136910", "Thyy has joined "+g.Name+" (")
	})

	discord.Session.AddHandler(func(s *discordgo.Session, pr *discordgo.PresencesReplace) {
		for _, pu := range *pr {
			e := ""
			if pu.Game != nil {
				e = pu.Game.Name
			}
			p.Update(pu.User.ID, e)
		}
	})

	discord.Session.AddHandler(func(s *discordgo.Session, pu *discordgo.PresenceUpdate) {
		e := ""
		if pu.Game != nil {
			e = pu.Game.Name
		}
		p.Update(pu.User.ID, e)
	})
}

// Help returns a list of help strings that are printed when the user requests them.
func (p *playedPlugin) Help(bot *rikka.Bot, service rikka.Service, message rikka.Message, detailed bool) []string {
	if detailed {
		return nil
	}

	return rikka.CommandHelp(service, "played", "[@username]", "Returns your most played games, or a users most played games if provided.")
}

func (p *playedPlugin) Message(bot *rikka.Bot, service rikka.Service, message rikka.Message) {
	defer rikka.MessageRecover()
	if service.Name() != rikka.DiscordServiceName {
		return
	}

	if service.IsMe(message) {
		return
	}

	if !rikka.MatchesCommand(service, "played", message) {
		return
	}

	mentions := message.Mentions()
	if len(mentions) > 1 {
		service.SendMessage(message.Channel(), "Please only mention one user")
		return
	}
	var mentionedUser *discordgo.User
	if len(mentions) == 1 {
		mentionedUser = mentions[0]
	}

	var id string
	if len(mentions) == 0 {
		_, parts := rikka.ParseCommand(service, message)
		switch len(parts) {
		case 1:
			id = parts[0]
			m, err := service.Member(message.Guild(), id)
			if err != nil {
				service.SendMessage(message.Channel(), "There was an error! Please report this to the devs\n"+err.Error())
				log.Println(err.Error())
				return
			}
			mentionedUser = m.User
		case 0:
			id = message.UserID()
		default:
			service.SendMessage(message.Channel(), "Please only search for one user at a time")
			return
		}
	}

	if id == "" {
		service.SendMessage(message.Channel(), "tell thy he dun screwed up")
		return
	}

	p.Lock()
	defer p.Unlock()

	u := p.Users[id]
	if u == nil {
		service.SendMessage(message.Channel(), fmt.Sprintf("I haven't seen user %s.", id))
		return
	}

	if len(u.Entries) == 0 {
		service.SendMessage(message.Channel(), fmt.Sprintf("I do not have anything recorded for user %s.", id))
		return
	}

	lc := humanize.Time(u.LastChanged)
	u.Update(u.Current, time.Now())

	pes := make(byDuration, len(u.Entries))
	i := 0
	for _, pe := range u.Entries {
		pes[i] = pe
		i++
	}

	sort.Sort(pes)
	discord := service.(*rikka.Discord)
	var statuses string

	for i = 0; i < len(pes) && i < 5; i++ {
		pe := pes[i]

		du := pe.Duration

		ds := ""
		hours := int(du / time.Hour)
		if hours > 0 {
			ds += fmt.Sprintf("%dh ", hours)
			du -= time.Duration(hours) * time.Hour
		}

		minutes := int(du / time.Minute)
		if minutes > 0 || len(ds) > 0 {
			ds += fmt.Sprintf("%dm ", minutes)
			du -= time.Duration(minutes) * time.Minute
		}

		seconds := int(du / time.Second)
		ds += fmt.Sprintf("%ds", seconds)

		//statuses = append(statuses, &discordgo.MessageEmbedField{Name: pe.Name, Value: ds, Inline: false})
		statuses += fmt.Sprintf("• **%s**: %s\n", pe.Name, ds)
	}

	var title, url string
	if mentionedUser != nil {
		title = mentionedUser.Username
		url = discordgo.EndpointUserAvatar(mentionedUser.ID, mentionedUser.Avatar)
	} else {
		user := message.User()
		title = message.UserName()
		url = discordgo.EndpointUserAvatar(user.ID, user.Avatar)
	}

	embed := &discordgo.MessageEmbed{
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
	}

	_, err := discord.Session.ChannelMessageSendEmbed(message.Channel(), embed)
	if err != nil {
		service.SendMessage(message.Channel(), "Unable to send embed "+err.Error())
	}
	//service.SendMessage(message.Channel(), messageText)
}

// Name returns the name of the plugin.
func (p *playedPlugin) Name() string {
	return "Played"
}

// Stats will return the stats for a plugin.
func (p *playedPlugin) Stats(bot *rikka.Bot, service rikka.Service, message rikka.Message) []string {
	return []string{fmt.Sprintf("Users playing: \t%d\n", len(p.Users))}
}

// New will create a played plugin.
func New() rikka.Plugin {
	return &playedPlugin{
		Users: map[string]*playedUser{},
	}
}
