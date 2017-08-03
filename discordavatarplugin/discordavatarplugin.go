package discordavatarplugin

import (
	"regexp"
	"strings"

	"github.com/ThyLeader/rikka"
	"github.com/bwmarrin/discordgo"
)

var userIDRegex = regexp.MustCompile("<@!?([0-9]*)>")

func avatarLoadFunc(bot *rikka.Bot, service rikka.Service, data []byte) error {
	if service.Name() != rikka.DiscordServiceName {
		panic("DiscordAvatar plugin only supports Discord.")
	}
	return nil
}

func avatarMessageFunc(bot *rikka.Bot, service rikka.Service, message rikka.Message) {
	if service.IsMe(message) {
		return
	}

	if !rikka.MatchesCommand(service, "avatar", message) {
		return
	}

	query := strings.Join(strings.Split(message.RawMessage(), " ")[1:], " ")

	id := message.UserID()
	match := userIDRegex.FindStringSubmatch(query)
	if match != nil {
		id = match[1]
	}

	discord := service.(*rikka.Discord)

	u, err := discord.Session.User(id)
	if err != nil {
		return
	}

	service.SendMessage(message.Channel(), discordgo.EndpointUserAvatar(u.ID, u.Avatar))
}

func avatarHelpFunc(bot *rikka.Bot, service rikka.Service, message rikka.Message, detailed bool) []string {
	if detailed {
		return nil
	}
	return rikka.CommandHelp(service, "avatar", "[@username]", "Returns a big version of your avatar, or a users avatar if provided.")
}

// New creates a new discordavatar plugin.
func New() rikka.Plugin {
	p := rikka.NewSimplePlugin("DiscordAvatar")
	p.LoadFunc = avatarLoadFunc
	p.MessageFunc = avatarMessageFunc
	p.HelpFunc = avatarHelpFunc
	return p
}
