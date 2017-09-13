package emojiplugin

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/ThyLeader/rikka"
)

func emojiFile(base, s string) string {
	found := ""
	filename := ""
	for _, r := range s {
		if filename != "" {
			filename = fmt.Sprintf("%s-%x", filename, r)
		} else {
			filename = fmt.Sprintf("%x", r)
		}

		if _, err := os.Stat(fmt.Sprintf("%s/%s.png", base, filename)); err == nil {
			found = filename
		} else if found != "" {
			return found
		}
	}
	return found
}

var discordRegex = regexp.MustCompile("<:.*?:(.*?)>")

func emojiMessageFunc(bot *rikka.Bot, service rikka.Service, message rikka.Message) {
	if service.IsMe(message) {
		return
	}

	if !rikka.MatchesCommand(service, "emoji", message) {
		return
	}

	base := "emoji/twitter/72x72"
	_, parts := rikka.ParseCommand(service, message)
	if len(parts) == 1 {
		submatches := discordRegex.FindStringSubmatch(parts[0])
		if len(submatches) != 0 {
			h, err := http.Get("https://cdn.discordapp.com/emojis/" + submatches[1] + ".png")
			if err != nil {
				service.SendMessage(message.Channel(), ":octagonal_sign: : Error getting emoji from Discord CDN - "+err.Error())
				return
			}

			service.SendFile(message.Channel(), "emoji.png", h.Body)
			h.Body.Close()
			return
		}

		s := strings.TrimSpace(parts[0])
		for i := range s {
			filename := emojiFile(base, s[i:])
			if filename != "" {
				if f, err := os.Open(fmt.Sprintf("%s/%s.png", base, filename)); err == nil {
					defer f.Close()
					service.SendFile(message.Channel(), "emoji.png", f)
					return
				} else {
					service.SendMessage(message.Channel(), ":octagonal_sign: : Error opening emoji - "+err.Error())
					return
				}
			}
		}
	}
}

func emojiHelpFunc(bot *rikka.Bot, service rikka.Service, message rikka.Message, detailed bool) []string {
	help := rikka.CommandHelp(service, "emoji", "<emoji>", "Returns a big version of an emoji.")

	if detailed {
		return nil
	}

	return help
}

// New creates a new emoji plugin.
func New() rikka.Plugin {
	p := rikka.NewSimplePlugin("Emoji")
	p.MessageFunc = emojiMessageFunc
	p.HelpFunc = emojiHelpFunc
	return p
}
