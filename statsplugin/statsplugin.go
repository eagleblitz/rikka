package statsplugin

import (
	"fmt"
	"runtime"
	"time"

	"github.com/ThyLeader/rikka"
	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
)

var statsStartTime = time.Now()

func getDurationString(duration time.Duration) string {
	return fmt.Sprintf(
		"%0.2d:%02d:%02d",
		int(duration.Hours()),
		int(duration.Minutes())%60,
		int(duration.Seconds())%60,
	)
}

// StatsCommand returns bot statistics.
func StatsCommand(bot *rikka.Bot, service rikka.Service, message rikka.Message, command string, parts []string) {
	var users, channels int
	discord := service.(*rikka.Discord)
	stats := runtime.MemStats{}
	runtime.ReadMemStats(&stats)

	for _, e := range discord.Session.State.Ready.Guilds {
		users += e.MemberCount
		channels += len(e.Channels)
	}

	guilds := service.ChannelCount()

	embed := &discordgo.MessageEmbed{
		Title: "Bot stats",
		Fields: []*discordgo.MessageEmbedField{
			&discordgo.MessageEmbedField{Name: "GoLang | DiscordGo", Value: fmt.Sprintf("%s | %s", runtime.Version(), discordgo.VERSION), Inline: true},
			&discordgo.MessageEmbedField{Name: "Uptime", Value: fmt.Sprintf("%s", getDurationString(time.Now().Sub(statsStartTime))), Inline: true},
			&discordgo.MessageEmbedField{Name: "Memory used", Value: fmt.Sprintf("%s / %s (%s garbage collected)", humanize.Bytes(stats.Alloc), humanize.Bytes(stats.Sys), humanize.Bytes(stats.TotalAlloc)), Inline: true},
			&discordgo.MessageEmbedField{Name: "Concurrent tasks", Value: fmt.Sprintf("%d", runtime.NumGoroutine()), Inline: true},
			&discordgo.MessageEmbedField{Name: "Users | Channels | Guilds", Value: fmt.Sprintf("%d | %d | %d", users, channels, guilds), Inline: true},
			&discordgo.MessageEmbedField{Name: "Total Shards | Current Shard", Value: fmt.Sprintf("%d | %d", discord.Session.ShardCount, discord.Session.ShardID+1), Inline: true},
		},
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: discordgo.EndpointUserAvatar(discord.Session.State.User.ID, discord.Session.State.User.Avatar),
		},
		Color:     0x79c879,
		Timestamp: fmt.Sprintf("%s", time.Now().Format(time.RFC3339)),
	}

	_, err := discord.Session.ChannelMessageSendEmbed(message.Channel(), embed)
	if err != nil {
		service.SendMessage(message.Channel(), ":octagonal_sign: : Error getting Bot info - "+err.Error())
	}
}

// StatsHelp is the help for the stats command.
var StatsHelp = rikka.NewCommandHelp("", "Lists bot statistics.")
