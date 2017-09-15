package main

import (
	"fmt"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"time"

	"github.com/spf13/viper"

	"github.com/ThyLeader/rikka"
	"github.com/ThyLeader/rikka/discordavatarplugin"
	"github.com/ThyLeader/rikka/emojiplugin"
	"github.com/ThyLeader/rikka/imageplugin"
	"github.com/ThyLeader/rikka/inviteplugin"
	"github.com/ThyLeader/rikka/mathplugin"
	"github.com/ThyLeader/rikka/misccommands"
	"github.com/ThyLeader/rikka/musicplugin"
	"github.com/ThyLeader/rikka/nametrackplugin"
	"github.com/ThyLeader/rikka/playedplugin"
	"github.com/ThyLeader/rikka/playingplugin"
	"github.com/ThyLeader/rikka/pubgplugin"
	"github.com/ThyLeader/rikka/reminderplugin"
	"github.com/ThyLeader/rikka/statsplugin"
)

var discordToken string
var discordApplicationClientID string
var discordOwnerUserID string
var discordShards int
var carbonitexKey string

func init() {
	rand.Seed(time.Now().UnixNano())

	viper.AddConfigPath(".")
	viper.SetConfigType("json")
	viper.SetConfigName("config")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	if len(os.Args) > 1 {
		if t := viper.GetString("token1"); t != "" {
			discordToken = t
		} else {
			panic("token1 not set")
		}
	} else {
		if t := viper.GetString("token2"); t != "" {
			discordToken = t
		} else {
			panic("token2 not set")
		}
	}
	if u := viper.GetString("ownerid"); u != "" {
		discordOwnerUserID = u
	} else {
		panic("ownerid not set")
	}
	if c := viper.GetString("clientid"); c != "" {
		discordApplicationClientID = c
	} else {
		panic("clientid not set")
	}
}

func main() {
	q := make(chan bool)

	// Set our variables.
	bot := rikka.NewBot()

	// Generally CommandPlugins don't hold state, so we share one instance of the command plugin for all services.
	cp := rikka.NewCommandPlugin()
	cp.AddCommand("invite", inviteplugin.InviteCommand, inviteplugin.InviteHelp)
	cp.AddCommand("join", inviteplugin.InviteCommand, nil)
	cp.AddCommand("stats", statsplugin.StatsCommand, statsplugin.StatsHelp)
	cp.AddCommand("info", statsplugin.StatsCommand, nil)
	cp.AddCommand("stat", statsplugin.StatsCommand, nil)
	cp.AddCommand("guilds", statsplugin.GuildsCommand, nil)
	cp.AddCommand("pepe", misccommands.MessagePeepo, nil)
	cp.AddCommand("ts", misccommands.MessageIDTS, misccommands.HelpIDTS)
	cp.AddCommand("support", misccommands.MessageSupport, misccommands.HelpSupport)
	cp.AddCommand("server", misccommands.MessageSupport, nil)
	cp.AddCommand("ping", misccommands.MessagePing, misccommands.HelpPing)

	cp.AddCommand("quit", func(bot *rikka.Bot, service rikka.Service, message rikka.Message, args string, parts []string) {
		if service.IsBotOwner(message) {
			q <- true
		}
	}, nil)

	var discord *rikka.Discord
	discord = rikka.NewDiscord(discordToken)

	discord.ApplicationClientID = discordApplicationClientID
	discord.OwnerUserID = discordOwnerUserID
	discord.Shards = discordShards
	bot.RegisterService(discord)

	bot.RegisterPlugin(discord, cp)

	//bot.RegisterPlugin(discord, darkthemetextplugin.New())
	bot.RegisterPlugin(discord, discordavatarplugin.New())
	bot.RegisterPlugin(discord, musicplugin.New(discord))
	bot.RegisterPlugin(discord, playedplugin.New())
	bot.RegisterPlugin(discord, playingplugin.New())
	bot.RegisterPlugin(discord, reminderplugin.New())
	bot.RegisterPlugin(discord, mathplugin.New())
	bot.RegisterPlugin(discord, imageplugin.New())
	bot.RegisterPlugin(discord, pubgplugin.New())
	bot.RegisterPlugin(discord, nametrackplugin.New())
	//bot.RegisterPlugin(discord, callbacktesting.New())
	bot.RegisterPlugin(discord, emojiplugin.New())

	// Start all our services.
	bot.Open()
	fmt.Println("bot running")
	// Wait for a termination signal, while saving the bot state every minute. Save on close.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	t := time.Tick(1 * time.Minute)

	go http.ListenAndServe("localhost:6060", nil)

out:
	for {
		select {
		case <-q:
			break out
		case <-c:
			break out
		case <-t:
			bot.Save()
		}
	}

	bot.Save()
}
