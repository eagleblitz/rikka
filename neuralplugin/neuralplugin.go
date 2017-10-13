package neuralplugin

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/ThyLeader/rikka"
	"github.com/bwmarrin/discordgo"
)

type neuralPlugin struct {
	URL string
}

func (p *neuralPlugin) Load(bot *rikka.Bot, service rikka.Service, data []byte) error {
	return nil
}

func (p *neuralPlugin) Message(bot *rikka.Bot, service rikka.Service, message rikka.Message) {
	if service.IsMe(message) {
		return
	}

	if !rikka.MatchesCommand(service, "gen", message) && !rikka.MatchesCommand(service, "generate", message) {
		return
	}

	_, parts := rikka.ParseCommand(service, message)
	if len(parts) < 1 {
		service.SendMessage(message.Channel(), "Please provide something to generate. eg. `r.gen shakespeare`")
		return
	}

	r, err := p.requestData(parts[0], "200")
	if err != nil {
		service.SendMessage(message.Channel(), fmt.Sprintf("There was an error! %s", err.Error()))
		return
	}
	if r.Message != "" {
		service.SendMessage(message.Channel(), fmt.Sprintf("There was an error! %s", r.Message))
		return
	}
	service.SendMessageEmbed(message.Channel(), &discordgo.MessageEmbed{
		//Title: "In an alternate universe",
		Author: &discordgo.MessageEmbedAuthor{
			Name:    "Shakespeare",
			IconURL: "https://www.biography.com/.image/c_fill%2Ccs_srgb%2Cg_face%2Ch_300%2Cq_80%2Cw_300/MTE1ODA0OTcxNzgzMzkwNzMz/william-shakespeare-194895-1-402.jpg",
		},
		Color: 0xff0000,
		Fields: []*discordgo.MessageEmbedField{
			&discordgo.MessageEmbedField{
				Name:  "In an alternate universe...",
				Value: r.Data,
			},
		},
	})
}

type response struct {
	Message string `json:"message"`
	Data    string `json:"data"`
}

func (p *neuralPlugin) requestData(t, l string) (*response, error) {
	res, err := http.Get(fmt.Sprintf(p.URL, t))
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	d, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	r := response{}
	json.Unmarshal(d, &r)

	return &r, nil
}

func (p *neuralPlugin) Help(bot *rikka.Bot, service rikka.Service, message rikka.Message, detailed bool) []string {
	if detailed {
		return nil
	}
	return rikka.CommandHelp(service, "names", "[@username]", "See a user's past usernames")
}

func (p *neuralPlugin) Name() string {
	return "Neural"
}

func (p *neuralPlugin) Save() ([]byte, error) {
	return nil, nil
}

func (p *neuralPlugin) Stats(bot *rikka.Bot, service rikka.Service, message rikka.Message) []string {
	return nil
}

// New creates a new discordavatar plugin.
func New(url string) rikka.Plugin {
	return &neuralPlugin{
		URL: url,
	}
}
