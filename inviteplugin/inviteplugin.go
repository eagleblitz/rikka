package inviteplugin

import (
	"fmt"
	"log"
	"strings"

	"github.com/ThyLeader/rikka"
)

func discordInviteID(id string) string {
	id = strings.Replace(id, "://discordapp.com/invite/", "://discord.gg/", -1)
	id = strings.Replace(id, "https://discord.gg/", "", -1)
	id = strings.Replace(id, "http://discord.gg/", "", -1)
	return id
}

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

	if len(parts) == 1 {
		join := parts[0]
		if service.Name() == rikka.DiscordServiceName {
			join = discordInviteID(join)
		}
		if err := service.Join(join); err != nil {
			if service.Name() == rikka.DiscordServiceName && err == rikka.ErrAlreadyJoined {
				service.PrivateMessage(message.UserID(), "I have already joined that server.")
				return
			}
			log.Println("Error joining %s %v", service.Name(), err)
		} else if service.Name() == rikka.DiscordServiceName {
			service.PrivateMessage(message.UserID(), "I have joined that server.")
		}
	}
}
