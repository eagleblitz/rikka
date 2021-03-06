package carbonitexplugin

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/ThyLeader/rikka"
)

type carbonitexPlugin struct {
	rikka.SimplePlugin
	key string
}

func (p *carbonitexPlugin) carbonitexPluginLoadFunc(bot *rikka.Bot, service rikka.Service, data []byte) error {
	if service.Name() != rikka.DiscordServiceName {
		panic("Carbonitex Plugin only supports Discord.")
	}

	go p.Run(bot, service)
	return nil
}

func (p *carbonitexPlugin) Run(bot *rikka.Bot, service rikka.Service) {
	for {
		<-time.After(5 * time.Minute)

		http.PostForm("https://www.carbonitex.net/discord/data/botdata.php", url.Values{"key": {p.key}, "servercount": {fmt.Sprintf("%d", service.ChannelCount())}})

		<-time.After(55 * time.Minute)
	}

}

// New will create a new carbonitex plugin.
// This plugin reports the server count to the carbonitex service.
func New(key string) rikka.Plugin {
	p := &carbonitexPlugin{
		SimplePlugin: *rikka.NewSimplePlugin("Carbonitex"),
		key:          key,
	}
	p.LoadFunc = p.carbonitexPluginLoadFunc
	return p
}
