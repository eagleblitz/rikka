package mathplugin

import (
	"fmt"

	"github.com/Knetic/govaluate"
	"github.com/ThyLeader/rikka"
)

func messageFunc(bot *rikka.Bot, service rikka.Service, message rikka.Message) {
	if service.IsMe(message) {
		return
	}
	if !rikka.MatchesCommand(service, "math", message) && !rikka.MatchesCommand(service, "eval", message) {
		return
	}
	defer mathRecover(service, message.Channel())
	t, _ := rikka.ParseCommand(service, message)

	expression, err := govaluate.NewEvaluableExpression(t)
	result, err := expression.Evaluate(nil)
	if err != nil {
		service.SendMessage(message.Channel(), fmt.Sprintf("There was an error\n%s", err.Error()))
		return
	}

	service.SendMessage(message.Channel(), fmt.Sprintf("**%s**, your expression evaluates to `%v`", message.UserName(), result))
}

func helpFunc(bot *rikka.Bot, service rikka.Service, message rikka.Message, detailed bool) []string {
	if detailed {
		return []string{
			"Check out <https://gist.github.com/ThyLeader/d0a98d2c15a824997513a92c911b0fab> for a detailed explanation.",
			"An edited down version coming soon",
		}
	}
	return rikka.CommandHelp(service, "math/eval", "<expression>", "Evaluates an expression. see `help math` for all possible operators")
}

func mathRecover(service rikka.Service, cID string) {
	if r := recover(); r != nil {
		service.SendMessage(cID, "There was an error processing the request")
	}
}

// New creates a new math plugin.
func New() rikka.Plugin {
	p := rikka.NewSimplePlugin("Math")
	p.MessageFunc = messageFunc
	p.HelpFunc = helpFunc
	return p
}
