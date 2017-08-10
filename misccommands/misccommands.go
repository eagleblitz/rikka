package misccommands

import "github.com/ThyLeader/rikka"

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

// func HelpPeepo(bot *rikka.Bot, service rikka.Service, message rikka.Message, detailed bool) []string {
// 	if detailed {
// 		return nil
// 	}
// 	return rikka.CommandHelp(service, "pepe", "", "Sends a pepe.")
// }

var HelpPeepo = rikka.NewCommandHelp("", "Sends a pepe.")
