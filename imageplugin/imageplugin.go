package imageplugin

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/ThyLeader/rikka"
	"github.com/bwmarrin/discordgo"
)

type response struct {
	Status   int           `json:"status"`
	Message  string        `json:"message,omitempty"`
	ID       string        `json:"id"`
	Type     string        `json:"type"`
	BaseType string        `json:"baseType"`
	Nsfw     bool          `json:"nsfw"`
	FileType string        `json:"fileType"`
	MimeType string        `json:"mimeType"`
	Tags     []interface{} `json:"tags"`
	URL      string        `json:"url"`
	Hidden   bool          `json:"hidden"`
	Account  string        `json:"account"`
}

type imagePlugin struct {
	Client     *http.Client
	Key        string
	Categories []string
	Tags       []string
}

func (i *imagePlugin) Message(bot *rikka.Bot, service rikka.Service, message rikka.Message) {
	if service.IsMe(message) {
		return
	}

	for _, e := range i.Categories {
		if rikka.MatchesCommand(service, e, message) {
			service.Typing(message.Channel())
			var r response
			req, _ := http.NewRequest("GET", "https://api.weeb.sh/images/random/?type="+e, nil)
			req.Header.Set("Authorization", "Bearer "+i.Key)
			res, err := i.Client.Do(req)
			if err != nil {
				service.SendMessage(message.Channel(), "Error getting the image - "+err.Error())
				return
			}
			defer res.Body.Close()

			body, err := ioutil.ReadAll(res.Body)
			if err != nil {
				return
			}
			json.Unmarshal(body, &r)

			if r.Status == 404 {
				service.SendMessage(message.Channel(), "Unable to find the requested image (404)")
				return
			}

			service.SendMessageEmbed(message.Channel(), &discordgo.MessageEmbed{
				Image: &discordgo.MessageEmbedImage{
					URL: r.URL,
				},
				Footer: &discordgo.MessageEmbedFooter{
					Text: "Powered by weeb.sh",
				},
			})
		}
	}
}

func (i *imagePlugin) Help(bot *rikka.Bot, service rikka.Service, message rikka.Message, detailed bool) (help []string) {
	if detailed {
		help = append(help, "Available image commands:\n")
		index := 0
		tmp := []string{}
		for _, e := range i.Categories {
			if index == 6 {
				help = append(help, strings.Join(tmp, ", "))
				tmp = []string{}
				index = 0
			}
			tmp = append(tmp, fmt.Sprintf("`r.%s`", e))
			index++
		}
		help = append(help, strings.Join(tmp, ", "))
		return help
	}

	help = []string{
		rikka.CommandHelp(service, "images", "", fmt.Sprintf("Images, see `%shelp images`", service.CommandPrefix()))[0],
	}
	return help
}

func (i *imagePlugin) Load(bot *rikka.Bot, service rikka.Service, data []byte) error {
	return nil
}

func (i *imagePlugin) Name() string {
	return "Images"
}

func (i *imagePlugin) Save() ([]byte, error) {
	return nil, nil
}

func (i *imagePlugin) Stats(bot *rikka.Bot, service rikka.Service, message rikka.Message) []string {
	return nil
}

// New creates a new math plugin.
func New(key string) rikka.Plugin {
	return &imagePlugin{
		Client:     &http.Client{},
		Key:        key,
		Categories: []string{"awoo", "bang", "blush", "clagwimoth", "cry", "cuddle", "dance", "hug", "insult", "jojo", "kiss", "lewd", "lick", "megumin", "neko", "nom", "owo", "pat", "poke", "pout", "rem", "shrug", "slap", "sleepy", "smile", "teehee", "smug", "stare", "thumbsup", "triggered", "wag", "waifu_insult", "wasted", "sumfuk", "dab", "tickle", "highfive", "banghead", "bite", "discord_memes", "nani", "initial_d"},
		Tags:       []string{"nuzzle", "cuddle", "momiji inubashiri", "wan", "astolfo", "facedesk", "everyone"},
	}
}
