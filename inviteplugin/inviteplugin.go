package inviteplugin

import (
	"fmt"

	"github.com/ThyLeader/rikka"
)

// InviteHelp will return the help text for the invite command.
func InviteHelp(bot *rikka.Bot, service rikka.Service, message rikka.Message) (string, string) {
	switch service.Name() {
	case rikka.DiscordServiceName:
		discord := service.(*rikka.Discord)

		if discord.ApplicationClientID != "" {
			return "", fmt.Sprintf("Returns a URL to add %s to your server.", service.UserName())
		}
		return "<discordinvite>", "Joins the provided Discord server."
	}
	return "<channel>", "Joins the provided channel."
}

// InviteCommand is a command for accepting an invite to a channel.
func InviteCommand(bot *rikka.Bot, service rikka.Service, message rikka.Message, command string, parts []string) {
	if service.Name() == rikka.DiscordServiceName {
		discord := service.(*rikka.Discord)

		if discord.ApplicationClientID != "" {
			service.SendMessage(message.Channel(), fmt.Sprintf("Please visit https://discordapp.com/oauth2/authorize?client_id=%s&scope=bot to add %s to your server.", discord.ApplicationClientID, service.UserName()))
			return
		}
	}
}
