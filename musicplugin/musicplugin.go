package musicplugin

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ThyLeader/rikka"
	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca"
)

type MusicPlugin struct {
	sync.Mutex

	discord *rikka.Discord

	VoiceConnections map[string]*voiceConnection
}

type voiceConnection struct {
	sync.Mutex
	debug bool

	GuildID      string
	ChannelID    string
	MaxQueueSize int
	Queue        []song
	Loop         bool
	Repeat       bool
	Announce     bool

	close   chan struct{}
	control chan controlMessage
	playing *song
	conn    *discordgo.VoiceConnection
}

type controlMessage int

var songsAdded int

const (
	Skip controlMessage = iota
	Pause
	Resume
)

type song struct {
	TextChannelID string
	AddedBy       string
	ID            string `json:"id"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	FullTitle     string `json:"full_title"`
	Thumbnail     string `json:"thumbnail"`
	URL           string `json:"webpage_url"`
	Duration      int    `json:"duration"`
	Likes         int    `json:"like_count"`
	Views         int    `json:"view_count"`
	Remaining     int
}

// New will create a new music plugin.
func New(discord *rikka.Discord) rikka.Plugin {

	p := &MusicPlugin{
		discord:          discord,
		VoiceConnections: make(map[string]*voiceConnection),
	}

	return p
}

// Name returns the name of the plugin.
func (p *MusicPlugin) Name() string {
	return "Music"
}

// Load will load plugin state from a byte array.
func (p *MusicPlugin) Load(bot *rikka.Bot, service rikka.Service, data []byte) (err error) {
	if service.Name() != rikka.DiscordServiceName {
		panic("Music Plugin only supports Discord.")
	}

	if data != nil {
		if err = json.Unmarshal(data, p); err != nil {
			log.Println("musicplugin: loading data err:", err)
		}
	}

	go p.init(service)

	return nil
}

func (p *MusicPlugin) init(service rikka.Service) {
	<-time.After(1 * time.Second)
	for _, s := range p.discord.Sessions {
		if !s.DataReady {
			go p.init(service)
			return
		}
	}
	p.ready(service)
}

func (p *MusicPlugin) ready(service rikka.Service) {
	// Join all registered voice channels and start the playback queue
	for _, v := range p.VoiceConnections {
		if v.ChannelID == "" {
			continue
		}
		vc, err := p.join(v.ChannelID)
		if err != nil {
			log.Println("musicplugin: join channel err:", err)
			continue
		}
		p.gostart(vc, service)
	}
}

// Save will save plugin state to a byte array.
func (p *MusicPlugin) Save() ([]byte, error) {
	return json.Marshal(p)
}

// Help returns a list of help strings that are printed when the user requests them.
func (p *MusicPlugin) Help(bot *rikka.Bot, service rikka.Service, message rikka.Message, detailed bool) []string {
	if service.IsPrivate(message) {
		return nil
	}

	help := []string{
		rikka.CommandHelp(service, "music", "<command>", fmt.Sprintf("Music, see `%shelp music`", service.CommandPrefix()))[0],
	}

	if detailed {
		help = append(help, []string{
			"Examples:",
			rikka.CommandHelp(service, "music", "join [channelid]", "Join your voice channel or the provided voice channel.")[0],
			rikka.CommandHelp(service, "music", "leave", "Leave current voice channel.")[0],
			rikka.CommandHelp(service, "music", "play/add [url | youtube search term]", "Start playing music and optionally enqueue provided url.")[0],
			rikka.CommandHelp(service, "music", "info", "Information about this plugin and the currently playing song.")[0],
			rikka.CommandHelp(service, "music", "pause", "Pause playback of current song.")[0],
			rikka.CommandHelp(service, "music", "resume", "Resume playback of current song.")[0],
			rikka.CommandHelp(service, "music", "skip", "Skip current song.")[0],
			rikka.CommandHelp(service, "music", "stop", "Stop playing music.")[0],
			rikka.CommandHelp(service, "music", "list/queue", "List contents of queue.")[0],
			rikka.CommandHelp(service, "music", "clear", "Clear all items from queue.")[0],
			rikka.CommandHelp(service, "music", "stats", "View stats about the music command.")[0],
			rikka.CommandHelp(service, "music", "loop", "Loops through the current queue.")[0],
			rikka.CommandHelp(service, "music", "repeat", "Repeats the current song.")[0],
			rikka.CommandHelp(service, "music", "announce", "Toggles 'now playing' announcements.")[0],
		}...)
	}

	return help
}

// Message handler.
func (p *MusicPlugin) Message(bot *rikka.Bot, service rikka.Service, message rikka.Message) {
	defer rikka.MessageRecover()

	if service.IsMe(message) {
		return
	}

	if !rikka.MatchesCommand(service, "music", message) && !rikka.MatchesCommand(service, "mu", message) && !rikka.MatchesCommand(service, "m", message) {
		return
	}

	if service.IsPrivate(message) {
		service.SendMessage(message.Channel(), "Sorry, this command doesn't work in private chat.")
		return
	}

	_, parts := rikka.ParseCommand(service, message)

	if len(parts) == 0 {
		service.SendMessage(message.Channel(), strings.Join(p.Help(bot, service, message, true), "\n"))
		return
	}

	// Get the Channel (and GuildID) for this channel because it's needed in
	// a few locations below
	channel, err := p.discord.Channel(message.Channel())
	if err != nil {
		log.Println("musicplugin: fetching channel err:", err.Error())
		return
	}

	// grab pointer to this channels voice connection, if exists.
	vc, vcok := p.VoiceConnections[channel.GuildID]

	switch parts[0] {

	case "help":
		// display extended help information
		service.SendMessage(message.Channel(), strings.Join(p.Help(bot, service, message, true), "\n"))

	case "join":
		// join the voice channel of the caller or the provided channel ID

		channelID := ""
		if len(parts) > 1 {
			channelID = parts[1]
		}

		if channelID == "" {
			messageUserID := message.UserID()
			for _, g := range p.discord.Guilds() {
				for _, v := range g.VoiceStates {
					if v.UserID == messageUserID {
						channelID = v.ChannelID
					}
				}
			}

			if channelID == "" {
				service.SendMessage(message.Channel(), "I couldn't find you in any voice channels, please join one.")
				return
			}
		}

		_, err := p.join(channelID)
		if err != nil {
			service.SendMessage(message.Channel(), err.Error())
			break
		}

		service.SendMessage(message.Channel(), "Now, let's play some music!")

	case "leave":
		// leave voice channel for this Guild
		if !vcok {
			service.SendMessage(message.Channel(), "There is no voice connection for this Guild.")
			return
		}

		err = vc.conn.Disconnect()
		if err != nil {
			log.Println("error disconnecting from vc", err.Error())
		}
		delete(p.VoiceConnections, channel.GuildID)
		service.SendMessage(message.Channel(), "Closed voice connection.")

	case "debug":
		// enable or disable debug
		if !vcok {
			service.SendMessage(message.Channel(), fmt.Sprintf("There is no voice connection for this Guild."))
			return
		}

		vc.Lock()
		vc.debug = !vc.debug
		service.SendMessage(message.Channel(), fmt.Sprintf("debug mode set to %v", vc.debug))
		vc.Unlock()

	//case "play":
	case "add", "play":
		// Start queue player and optionally enqueue provided songs
		if !vcok {
			service.SendMessage(message.Channel(), "There is no voice connection for this Guild.")
			return
		}

		p.gostart(vc, service)

		if len(parts[1:]) == 1 {
			u, err := url.ParseRequestURI(parts[1])
			if err != nil {
				service.Typing(message.Channel())
				err = p.enqueue(bot, vc, parts[1], service, message, true)
				if err != nil {
					service.SendMessage(message.Channel(), err.Error())
					return
				}
				return
			}
			err = p.enqueue(bot, vc, u.String(), service, message, false)
			if err != nil {
				service.SendMessage(message.Channel(), err.Error())
			}
			return
		}
		service.Typing(message.Channel())
		err = p.enqueue(bot, vc, strings.Join(parts[1:], " "), service, message, true)
		if err != nil {
			service.SendMessage(message.Channel(), err.Error())
		}

	case "stop":
		// stop the queue player
		if !vcok {
			service.SendMessage(message.Channel(), "There is no voice connection for this Guild.")
			return
		}

		if vc.close != nil {
			close(vc.close)
			vc.close = nil
		}

		if vc.control != nil {
			close(vc.control)
			vc.control = nil
		}

	case "skip":
		// skip current song
		if !vcok {
			service.SendMessage(message.Channel(), "There is no voice connection for this Guild.")
			return
		}

		if vc.control == nil {
			return
		}
		vc.control <- Skip

	case "pause":
		// pause the queue player
		if !vcok {
			service.SendMessage(message.Channel(), "There is no voice connection for this Guild.")
			return
		}

		if vc.control == nil {
			return
		}
		vc.control <- Pause

	case "resume":
		// resume the queue player
		if !vcok {
			service.SendMessage(message.Channel(), "There is no voice connection for this Guild.")
			return
		}

		if vc.control == nil {
			return
		}
		vc.control <- Resume

	case "info", "np":
		// report player settings, queue info, and current song

		if vc == nil {
			service.SendMessage(message.Channel(), "I'm not in a voice channel!")
			break
		}

		msg := fmt.Sprintf("`Voice Channel:` %s\n", vc.ChannelID)
		msg += fmt.Sprintf("`Queue Size:` %d\n", len(vc.Queue))

		if vc.playing == nil {
			service.SendMessage(message.Channel(), msg)
			break
		}

		msg += fmt.Sprintf("`Now Playing:`\n")
		msg += fmt.Sprintf("`ID:` %s\n", vc.playing.ID)
		msg += fmt.Sprintf("`Title:` %s\n", vc.playing.Title)
		msg += fmt.Sprintf("`Duration:` %ds\n", vc.playing.Duration)
		msg += fmt.Sprintf("`Remaining:` %ds\n", vc.playing.Remaining)
		msg += fmt.Sprintf("`Source URL:` <%s>\n", vc.playing.URL)
		msg += fmt.Sprintf("`Thumbnail:` %s\n", vc.playing.Thumbnail)
		service.SendMessage(message.Channel(), msg)

	case "stats":
		p.Lock()
		var l time.Duration
		var s int
		c := len(p.VoiceConnections)
		for _, e := range p.VoiceConnections {
			s += len(e.Queue)
			for _, q := range e.Queue {
				l += time.Duration(q.Duration)
			}
		}
		p.Unlock()
		msg := fmt.Sprintf("Music stats:\n")
		msg += fmt.Sprintf("`Total connections:`\t%v\n", c)
		msg += fmt.Sprintf("`Total songs queued:`\t%v\n", songsAdded)
		msg += fmt.Sprintf("`Current songs queued:`\t%v\n", s)
		msg += fmt.Sprintf("`Current time queued:`\t%v", time.Duration(l*time.Second).String())
		service.SendMessage(message.Channel(), msg)

	case "list", "queue":
		// list top items in the queue
		if !vcok {
			service.SendMessage(message.Channel(), "There is no voice connection for this Guild.")
			return
		}

		if len(vc.Queue) == 0 {
			service.SendMessage(message.Channel(), "The music queue is empty.")
			return
		}

		var msg string

		i := 1
		i2 := 0
		for k, v := range vc.Queue {
			np := ""
			if k == 0 {
				np = "**(Now Playing)**"
			}
			d := time.Duration(v.Duration) * time.Second
			msg += fmt.Sprintf("`%.3d:%.15s` **%s** [%s] - *%s* %s\n", k, v.ID, v.Title, d.String(), v.AddedBy, np)

			if i >= 15 {
				service.SendMessage(message.Channel(), msg)
				msg = ""
				i = 0
				i2++

				if i2 >= 8 {
					// limit response to 8 messages (120 songs)
					return
				}
			}
			i++
		}

		service.SendMessage(message.Channel(), msg)

	case "loop", "l":
		// loop the queue
		if !vcok {
			service.SendMessage(message.Channel(), "There is no voice connection for this Guild.")
			return
		}
		vc.Lock()
		defer vc.Unlock()
		if vc.Repeat {
			vc.Repeat = false
			vc.Loop = !vc.Loop
			go service.SendMessage(message.Channel(), fmt.Sprintf("Disabled repeat and set looping to `%v`", vc.Loop))
			return
		}

		vc.Loop = !vc.Loop
		go service.SendMessage(message.Channel(), fmt.Sprintf("Looping set to `%v`", vc.Loop))
		return

	case "repeat", "r":
		// repeat current song
		if !vcok {
			service.SendMessage(message.Channel(), "There is no voice connection for this Guild.")
			return
		}
		vc.Lock()
		defer vc.Unlock()
		if vc.Loop {
			vc.Loop = false
			vc.Repeat = !vc.Repeat
			go service.SendMessage(message.Channel(), fmt.Sprintf("Disabled looping and set repeat to `%v`", vc.Repeat))
			return
		}

		vc.Repeat = !vc.Repeat
		go service.SendMessage(message.Channel(), fmt.Sprintf("Repeat set to `%v`", vc.Repeat))
		return

	case "clear":
		// clear all items from the queue
		if !vcok {
			service.SendMessage(message.Channel(), "There is no voice connection for this Guild.")
			return
		}

		vc.Lock()
		vc.Queue = []song{}
		vc.Unlock()
		service.SendMessage(message.Channel(), "Queue cleared")

	case "announce":
		// toggle song announcements
		if !vcok {
			service.SendMessage(message.Channel(), "There is no voice connection for this Guild.")
			return
		}

		vc.Lock()
		vc.Announce = !vc.Announce
		go service.SendMessage(message.Channel(), fmt.Sprintf("Song announcements set to `%v`", vc.Announce))
		vc.Unlock()

	default:
		service.SendMessage(message.Channel(), "Unknown music command, try `help music`")
	}
}

// join a specific voice channel
func (p *MusicPlugin) join(cID string) (vc *voiceConnection, err error) {
	c, err := p.discord.Channel(cID)
	if err != nil {
		return
	}

	// 2 == GUILD_VOICE
	if c.Type != 2 {
		err = fmt.Errorf("not a voice channel")
		return
	}

	// Get or Create the VoiceConnection object
	p.Lock()
	vc, ok := p.VoiceConnections[c.GuildID]
	if !ok {
		vc = &voiceConnection{}
		vc.Announce = true
		p.VoiceConnections[c.GuildID] = vc
	}
	p.Unlock()

	// default song announcements to true
	vc.Announce = true

	guild, err := p.discord.Guild(c.GuildID)
	if err != nil {
		return
	}

	gID, err := strconv.Atoi(guild.ID)
	if err != nil {
		return
	}

	// NOTE: Setting mute to false, deaf to true.
	vc.conn, err = p.discord.Sessions[(gID>>22)%len(p.discord.Sessions)].ChannelVoiceJoin(c.GuildID, cID, false, true)
	if err != nil {
		return
	}

	vc.GuildID = c.GuildID
	vc.ChannelID = cID

	return
}

// enqueue a song/playlest to a VoiceConnections Queue
func (p *MusicPlugin) enqueue(bot *rikka.Bot, vc *voiceConnection, url string, service rikka.Service, message rikka.Message, search bool) (err error) {

	if vc == nil {
		return fmt.Errorf("Cannot enqueue to nil voice connection")
	}

	if url == "" {
		return fmt.Errorf("Cannot enqueue an empty string")
	}

	// TODO //////////////////////////////////////////////////////////////////
	// need to parse the url and have a way to know what we're doing
	// 1) option to queue local files
	// 2) option to queue saved playlists
	// 3) option to queue URL that can be passed directly to ffmpeg without youtube-dl
	// 4) option to queue youtube-dl playlist
	// 5) option to queue youtube-dl song
	// 6) option to queue youtube-dl search result

	// right now option 4 and 5 work, only.
	//////////////////////////////////////////////////////////////////////////

	var cmd *exec.Cmd
	if search {
		cmd = exec.Command("youtube-dl", "-i", "-j", "--youtube-skip-dash-manifest", fmt.Sprintf(`ytsearch5:%s`, url))
	} else {
		cmd = exec.Command("youtube-dl", "-i", "-j", "--youtube-skip-dash-manifest", url)
	}

	if vc.debug {
		cmd.Stderr = os.Stderr
	}

	output, err := cmd.StdoutPipe()
	if err != nil {
		log.Println(err)
		service.SendMessage(message.Channel(), fmt.Sprintf("Error adding song to playlist."))
		return
	}

	err = cmd.Start()
	if err != nil {
		log.Println(err)
		service.SendMessage(message.Channel(), fmt.Sprintf("Error adding song to playlist."))
		return
	}
	defer func() {
		go cmd.Wait()
	}()

	scanner := bufio.NewScanner(output)

	if search {
		// TODO: turn this all into a single function
		res := []song{}
		for scanner.Scan() {
			s := song{}
			err = json.Unmarshal(scanner.Bytes(), &s)
			if err != nil {
				log.Println(err)
				continue
			}

			s.TextChannelID = message.Channel()
			s.AddedBy = message.UserName()

			res = append(res, s)
		}
		if len(res) < 1 {
			service.SendMessage(message.Channel(), fmt.Sprintf("Your search term `%s` returned no results", url))
			return
		}
		if len(res) == 1 {
			vc.Lock()
			vc.Queue = append(vc.Queue, res[0])
			vcLen := len(vc.Queue)
			vc.Unlock()
			res[0].announceSongAdded(service, message.Channel(), vcLen)
			return
		}
		msg := []string{}
		msg = append(msg, []string{
			"```rb",
			"Please select the song you would like to play.\n",
		}...)
		for i, e := range res {
			i++
			msg = append(msg, fmt.Sprintf("[%v] # %s", i, e.Title))
		}
		msg = append(msg, []string{
			"\nType the appropriate number to select the song.",
			"Type 'exit' to leave the menu.",
			"```",
		}...)

		dd, _ := service.SendMessage(message.Channel(), strings.Join(msg, "\n"))
		defer service.DeleteMessage(message.Channel(), dd.ID)

		timeout := time.NewTicker(30 * time.Second)
		defer timeout.Stop()

		m, err := bot.MakeCallback(service, message.UserID())
		if err != nil {
			service.SendMessage(message.Channel(), "A menu already exists")
			return nil
		}
		defer bot.CloseCallback(service, message.UserID())
		e := 0
		for {
			select {
			case ms := <-m:
				if ms.Channel() != message.Channel() {
					continue
				}

				if strings.ToLower(ms.Message()) == "exit" {
					service.SendMessage(message.Channel(), "Exiting menu")
					return nil
				}

				n, err := strconv.Atoi(ms.Message())
				if e >= 5 {
					service.SendMessage(message.Channel(), "BAKA!! Seems you cant type a correct response. Exiting menu")
					return nil
				}
				if err != nil {
					service.SendMessage(message.Channel(), fmt.Sprintf("Please type a number between 1 and 5. You typed `%s`.", ms.Message()))
					e++
					continue
				}

				if n > 5 || n < 1 {
					service.SendMessage(message.Channel(), fmt.Sprintf("Please type a number between 1 and 5. You typed `%s`.", ms.Message()))
					e++
					continue
				}

				service.SendMessage(message.Channel(), fmt.Sprintf("You picked number %v.", n))
				s := res[n-1]
				if s.Duration > 18000 {
					service.SendMessage(message.Channel(), "Sorry, but Rikka does not currently allow songs longer than 5 hours")
					return nil
				}
				s.TextChannelID = message.Channel()
				s.AddedBy = message.UserName()

				vc.Lock()
				vc.Queue = append(vc.Queue, s)
				vcLen := len(vc.Queue)
				vc.Unlock()
				s.announceSongAdded(service, message.Channel(), vcLen)
				songsAdded++
				return nil
			case <-timeout.C:
				service.SendMessage(message.Channel(), "Menu timed out")
				return nil
			}
		}
	}

	for scanner.Scan() {
		s := song{}
		err = json.Unmarshal(scanner.Bytes(), &s)
		if err != nil {
			log.Println(err)
			continue
		}

		if s.Duration > 18000 {
			service.SendMessage(message.Channel(), "Sorry, but Rikka does not currently allow songs longer than 5 hours")
			return nil
		}

		s.TextChannelID = message.Channel()
		s.AddedBy = message.UserName()

		vc.Lock()
		vc.Queue = append(vc.Queue, s)
		vcLen := len(vc.Queue)
		vc.Unlock()
		s.announceSongAdded(service, message.Channel(), vcLen)
		songsAdded++
	}
	return
}

// i had a bunch of different ones scattered around so hopefully this will clean things up in terms of consistency
func (s *song) announceSongAdded(service rikka.Service, channel string, vcLen int) {
	service.SendMessage(channel, fmt.Sprintf("Added *%s* to the queue as requested by %s.\nThere are now `%v` songs in the queue", s.Title, s.AddedBy, vcLen))
}

func (s *song) announceSongPlaying(service rikka.Service, channel string, vcLen int, timeLeft string) {
	service.SendMessage(channel, fmt.Sprintf("Now playing *%s* as requested by *%s*\nSongs left in queue: `%v` `[%s total]`", s.Title, s.AddedBy, vcLen, timeLeft))
}

// little wrapper function for start() to fire it off in a
// go routine if it is not already running.
func (p *MusicPlugin) gostart(vc *voiceConnection, service rikka.Service) (err error) {
	vc.Lock()

	if vc == nil {
		vc.Unlock()
		return fmt.Errorf("gostart cannot start a nil voice connection queue")
	}

	if vc.close != nil || vc.control != nil {
		vc.Unlock()
		return fmt.Errorf("gostart will not start a voice connection with non-nil control channels")
	}

	vc.close = make(chan struct{})
	vc.control = make(chan controlMessage)

	// TODO can this be moved lower?
	vc.Unlock()

	go p.start(vc, vc.close, vc.control, service)

	return
}

// "start" is a goroutine function that loops though the music queue and
// plays songs as they are added
func (p *MusicPlugin) start(vc *voiceConnection, close <-chan struct{}, control <-chan controlMessage, service rikka.Service) {

	if close == nil || control == nil || vc == nil {
		log.Println("musicplugin: start() exited due to nil channels")
		return
	}

	var i int
	var s song

	// main loop keeps this going until close
	for {

		// exit if close channel is closed
		select {
		case <-close:
			log.Println("musicplugin: start() exited due to close channel.")
			return
		default:
		}

		// loop until voice connection is ready and songs are in the queue.
		if vc.conn == nil || vc.conn.Ready == false || len(vc.Queue) < 1 {
			time.Sleep(1 * time.Second)
			continue
		}

		// Get song to play and store it in local Song var
		vc.Lock()
		if vc.Loop && vc.Repeat {
			service.SendMessage(vc.playing.TextChannelID, "There was an error. Resetting loop and repeat")
			vc.Loop = false
			vc.Repeat = false
		}

		vcLen := len(vc.Queue)
		if vcLen-1 >= i {
			s = vc.Queue[i]
		} else {
			i = 0
			vc.Unlock()
			continue
		}
		vc.Unlock()

		vc.playing = &s
		timeLeft := time.Duration(0)
		for _, v := range vc.Queue {
			timeLeft += time.Duration(v.Duration)
		}
		timeLeft *= time.Second
		if vc.Announce {
			s.announceSongPlaying(service, vc.playing.TextChannelID, vcLen, timeLeft.String())
		}
		p.play(vc, close, control, s)
		vc.playing = nil

		vc.Lock()
		if len(vc.Queue) > 0 {
			if !vc.Repeat && !vc.Loop {
				vc.Queue = append(vc.Queue[:i], vc.Queue[i+1:]...)
			}
			if vc.Repeat {
				// do nothing
			}
			if vc.Loop {
				var pop song
				pop, vc.Queue = vc.Queue[0], vc.Queue[1:]
				vc.Queue = append(vc.Queue, pop)
			}
		}
		vc.Unlock()
	}
}

// play an individual song
func (p *MusicPlugin) play(vc *voiceConnection, close <-chan struct{}, control <-chan controlMessage, s song) {
	var err error

	if close == nil || control == nil || vc == nil || vc.conn == nil {
		log.Println("musicplugin: play exited because [close|control|vc|vc.conn] is nil.")
		return
	}

	options := dca.StdEncodeOptions
	options.RawOutput = true
	options.Bitrate = 64
	options.Application = "lowdelay"

	ytdl := exec.Command("youtube-dl", "-v", "-f", "bestaudio", "-o", "-", s.URL)
	ytdlout, err := ytdl.StdoutPipe()
	if err != nil {
		log.Println("ytdl StdoutPipe err:", err)
		return
	}
	ytdlbuf := bufio.NewReaderSize(ytdlout, 16384)

	encodingSession, err := dca.EncodeMem(ytdlbuf, options)
	if err != nil {
		fmt.Println("error creating encoding session", err.Error())
		return
	}
	defer encodingSession.Cleanup()

	err = ytdl.Start()
	if err != nil {
		log.Println("ytdl Start err:", err)
		return
	}
	defer func() {
		go ytdl.Wait()
	}()

	vc.conn.Speaking(true)
	defer vc.conn.Speaking(false)

	d := make(chan error)
	stream := dca.NewStream(encodingSession, vc.conn, d)

	start := time.Now()
	for {
		select {
		case <-close:
			log.Println("musicplugin: play() exited due to close channel.")
			return
		case err = <-d:
			if err == io.EOF {
				return
			}
			if err != nil {
				log.Println(err.Error())
				return
			}
		default:
		}
		select {
		case ctl := <-control:
			switch ctl {
			case Skip:
				return
			case Pause:
				stream.SetPaused(true)
				b := false
				for {
					ctl, ok := <-control
					if !ok {
						return
					}
					switch ctl {
					case Skip:
						return
					case Resume:
						stream.SetPaused(false)
						b = true
						break
					}
					if b == true {
						break
					}
				}
			}
		default:
		}

		vc.playing.Remaining = (vc.playing.Duration - int(time.Since(start).Seconds()))
		time.Sleep(500 * time.Millisecond)
	}
}

// Stats will return the stats for a plugin.
func (p *MusicPlugin) Stats(bot *rikka.Bot, service rikka.Service, message rikka.Message) []string {
	return nil
}
