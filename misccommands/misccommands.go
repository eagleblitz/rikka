package misccommands

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/ThyLeader/rikka"
)

var pepe = `:frog::frog::frog::frog::frog::frog::frog:
_:frog::frog::frog::frog::frog::frog::frog::frog::frog:
__:frog::frog::frog::frog::frog::frog::frog::frog::frog::frog::frog:
:frog::frog::frog::frog::frog::frog::frog::frog::frog::frog::frog::frog::frog:
:frog::frog::white_circle::black_circle::black_circle::white_circle::frog::frog::frog::white_circle::black_circle::black_circle::white_circle:
:frog::white_circle::black_circle::black_circle::white_circle::black_circle::white_circle::frog::white_circle::black_circle::black_circle::white_circle::black_circle::white_circle:
:frog::white_circle::black_circle::white_circle::black_circle::black_circle::white_circle::frog::white_circle::black_circle::white_circle::black_circle::black_circle::white_circle:
:frog::frog::white_circle::black_circle::white_circle::white_circle::frog::frog::frog::white_circle::black_circle::white_circle::white_circle:
:frog::frog::frog::frog::frog::frog::frog::frog::frog::frog::frog::frog::frog:
:red_circle::red_circle::frog::frog::frog::frog::frog::frog::frog::frog::frog::frog::frog:
:frog::red_circle::red_circle::frog::frog::frog::frog::frog::frog::frog::frog::frog:
:frog::frog::red_circle::red_circle::red_circle::red_circle::red_circle::red_circle::red_circle::red_circle::red_circle::red_circle::red_circle:
:frog::frog::frog::red_circle::red_circle::red_circle::red_circle::red_circle::red_circle::red_circle::red_circle::red_circle::red_circle:
:frog::frog::frog::frog::frog::frog::frog::frog::frog::frog::frog:
:frog::frog::frog::frog::frog::frog::frog::frog::frog::frog:
:frog::frog::frog::frog::frog::frog::frog::frog::frog:`

func MessagePeepo(bot *rikka.Bot, service rikka.Service, message rikka.Message, command string, parts []string) {
	if service.IsMe(message) {
		return
	}

	if !rikka.MatchesCommand(service, "pepe", message) {
		return
	}
	service.SendMessage(message.Channel(), pepe)
}

var HelpPeepo = rikka.NewCommandHelp("", "Sends a pepe.")

var userIDRegex = regexp.MustCompile("<@!?([0-9]*)>")
var chanIDRegex = regexp.MustCompile("<#!?([0-9]*)>")

func MessageIDTS(bot *rikka.Bot, service rikka.Service, message rikka.Message, command string, parts []string) {
	if service.IsMe(message) {
		return
	}

	if !rikka.MatchesCommand(service, "ts", message) {
		return
	}

	query := strings.Join(strings.Split(message.RawMessage(), " ")[1:], " ")
	var id string
	if len(parts) == 0 {
		id = message.UserID()
	} else {
		if q := userIDRegex.FindStringSubmatch(query); q != nil {
			id = q[1]
		}
		if q := chanIDRegex.FindStringSubmatch(query); q != nil {
			id = q[1]
		}
		if id == "" {
			id = parts[0]
		}
	}
	t, err := service.TimestampForID(id)
	if err != nil {
		service.SendMessage(message.Channel(), "Incorrect snowflake")
		return
	}
	service.SendMessage(message.Channel(), fmt.Sprintf("`%s`", t.UTC().Format(time.UnixDate)))
}

var HelpIDTS = rikka.NewCommandHelp("[@username]", "Parses a snowflake (id) and returns a timestamp.")

func MessageSupport(bot *rikka.Bot, service rikka.Service, message rikka.Message, command string, parts []string) {
	service.SendMessage(message.Channel(), "You can join the support server here: https://discord.gg/CB5sXP")
}

var HelpSupport = rikka.NewCommandHelp("", "Gives an invite link to join the support server.")

// MessagePing is the command handler for the ping command
func MessagePing(bot *rikka.Bot, service rikka.Service, message rikka.Message, command string, parts []string) {
	now := time.Now()
	p, _ := service.SendMessage(message.Channel(), "Pong!")
	t, err := p.Timestamp.Parse()
	if err != nil {
		service.SendMessage(message.Channel(), "There was an error parsing the timestamp "+err.Error())
		return
	}
	service.EditMessage(message.Channel(), p.ID, fmt.Sprintf("Pong! - `%s`", t.Sub(now).String()))
}

// HelpPing is the help text for the ping command
var HelpPing = rikka.NewCommandHelp("", "Shows bot latency.")
