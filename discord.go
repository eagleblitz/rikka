package rikka

import (
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
)

// The number of guilds supported by one shard.
const numGuildsPerShard = 2400

// DiscordServiceName is the service name for the Discord service.
const DiscordServiceName string = "Discord"

// DiscordMessage is a Message wrapper around discordgo.Message.
type DiscordMessage struct {
	Discord          *Discord
	DiscordgoMessage *discordgo.Message
	MessageType      MessageType
	Nick             *string
	Content          *string
}

// IsBot returns whether or not the message was sent from a bot
func (m *DiscordMessage) IsBot() bool {
	if m.DiscordgoMessage.Author.Bot {
		return m.DiscordgoMessage.Author.Bot
	}
	return false
}

// User returns the DiscordGo User type of a message
func (m *DiscordMessage) User() *discordgo.User {
	return m.DiscordgoMessage.Author
}

// Channel returns the channel id for this message.
func (m *DiscordMessage) Channel() string {
	return m.DiscordgoMessage.ChannelID
}

// UserName returns the user name for this message.
func (m *DiscordMessage) UserName() string {
	me := m.DiscordgoMessage
	if me.Author == nil {
		return ""
	}

	if m.Nick == nil {
		n := m.Discord.NicknameForID(me.Author.ID, me.Author.Username, me.ChannelID)
		m.Nick = &n
	}
	return *m.Nick
}

// UserID returns the user id for this message.
func (m *DiscordMessage) UserID() string {
	if m.DiscordgoMessage.Author == nil {
		return ""
	}

	return m.DiscordgoMessage.Author.ID
}

// UserAvatar returns the avatar url for this message.
func (m *DiscordMessage) UserAvatar() string {
	if m.DiscordgoMessage.Author == nil {
		return ""
	}

	return discordgo.EndpointUserAvatar(m.DiscordgoMessage.Author.ID, m.DiscordgoMessage.Author.Avatar)
}

// Message returns the message content for this message.
func (m *DiscordMessage) Message() string {
	if m.Content == nil {
		c := m.DiscordgoMessage.ContentWithMentionsReplaced()
		c = m.Discord.replaceRoleNames(m.DiscordgoMessage, c)
		c = m.Discord.replaceChannelNames(m.DiscordgoMessage, c)

		m.Content = &c
	}
	return *m.Content
}

// RawMessage returns the raw message content for this message.
func (m *DiscordMessage) RawMessage() string {
	return m.DiscordgoMessage.Content
}

// MessageID returns the message ID for this message.
func (m *DiscordMessage) MessageID() string {
	return m.DiscordgoMessage.ID
}

// Type returns the type of message.
func (m *DiscordMessage) Type() MessageType {
	return m.MessageType
}

// GuildID returns the guild ID of a message
func (m *DiscordMessage) GuildID() string {
	c, err := m.Discord.Channel(m.Channel())
	if err != nil {
		log.Println("error retrieving channel from state", err)
		return ""
	}
	g, err := m.Discord.Guild(c.GuildID)
	if err != nil {
		log.Println("error retrieving channel from state", err)
		return ""
	}
	return g.ID
}

// Mentions returns an array of mentions contained in a message
func (m *DiscordMessage) Mentions() []*discordgo.User {
	return m.DiscordgoMessage.Mentions
}

// Timestamp returns a parsed timestamp of a message
func (m *DiscordMessage) Timestamp() (time.Time, error) {
	return m.DiscordgoMessage.Timestamp.Parse()
}

// GuildName returns the name of the guild a message belongs to
func (m *DiscordMessage) GuildName() string {
	c, err := m.Discord.Channel(m.Channel())
	if err != nil {
		log.Println("error retrieving channel from state", err)
		return ""
	}
	g, err := m.Discord.Guild(c.GuildID)
	if err != nil {
		log.Println("error retrieving channel from state", err)
		return ""
	}
	return g.Name
}

// Guild returns the discordgo guild object of a message
func (m *DiscordMessage) Guild() *discordgo.Guild {
	c, err := m.Discord.Channel(m.Channel())
	if err != nil {
		log.Println("error retrieving channel from state", err)
		return nil
	}
	g, err := m.Discord.Guild(c.GuildID)
	if err != nil {
		log.Println("error retrieving channel from state", err)
		return nil
	}
	return g
}

// Discord is a Service provider for Discord.
type Discord struct {
	args        []interface{}
	messageChan chan Message

	Shards int

	// The first session, used to send messages (and maintain backwards compatibility).
	Session             *discordgo.Session
	Sessions            []*discordgo.Session
	OwnerUserID         string
	ApplicationClientID string
}

// NewDiscord creates a new discord service.
func NewDiscord(args ...interface{}) *Discord {
	return &Discord{
		args:        args,
		messageChan: make(chan Message, 200),
	}
}

var channelIDRegex = regexp.MustCompile("<#[0-9]*>")

func (d *Discord) replaceChannelNames(message *discordgo.Message, content string) string {
	return channelIDRegex.ReplaceAllStringFunc(content, func(str string) string {
		c, err := d.Channel(str[2 : len(str)-1])
		if err != nil {
			return str
		}

		return "#" + c.Name
	})
}

var roleIDRegex = regexp.MustCompile("<@&[0-9]*>")

func (d *Discord) replaceRoleNames(message *discordgo.Message, content string) string {
	return roleIDRegex.ReplaceAllStringFunc(content, func(str string) string {
		roleID := str[3 : len(str)-1]

		c, err := d.Channel(message.ChannelID)
		if err != nil {
			return str
		}

		g, err := d.Guild(c.GuildID)
		if err != nil {
			return str
		}

		for _, r := range g.Roles {
			if r.ID == roleID {
				return "@" + r.Name
			}
		}

		return str
	})
}

func (d *Discord) onMessageCreate(s *discordgo.Session, message *discordgo.MessageCreate) {
	if message.Content == "" {
		return
	}

	d.messageChan <- &DiscordMessage{
		Discord:          d,
		DiscordgoMessage: message.Message,
		MessageType:      MessageTypeCreate,
	}
}

func (d *Discord) onMessageUpdate(s *discordgo.Session, message *discordgo.MessageUpdate) {
	if message.Content == "" {
		return
	}

	d.messageChan <- &DiscordMessage{
		Discord:          d,
		DiscordgoMessage: message.Message,
		MessageType:      MessageTypeUpdate,
	}
}

func (d *Discord) onMessageDelete(s *discordgo.Session, message *discordgo.MessageDelete) {
	d.messageChan <- &DiscordMessage{
		Discord:          d,
		DiscordgoMessage: message.Message,
		MessageType:      MessageTypeDelete,
	}
}

// Name returns the name of the service.
func (d *Discord) Name() string {
	return DiscordServiceName
}

// Open opens the service and returns a channel which all messages will be sent on.
func (d *Discord) Open() (<-chan Message, error) {
	shards := d.Shards
	if shards < 1 {
		shards = 1
	}

	d.Sessions = make([]*discordgo.Session, shards)

	for i := 0; i < shards; i++ {
		session, err := discordgo.New(d.args...)
		if err != nil {
			return nil, err
		}
		session.State.TrackPresences = false
		session.ShardCount = shards
		session.ShardID = i
		session.AddHandler(d.onMessageCreate)
		session.AddHandler(d.onMessageUpdate)
		session.AddHandler(d.onMessageDelete)

		d.Sessions[i] = session
	}

	d.Session = d.Sessions[0]

	for i := 0; i < len(d.Sessions); i++ {
		d.Sessions[i].Open()
	}

	return d.messageChan, nil
}

// IsMe returns whether or not a message was sent by the bot.
func (d *Discord) IsMe(message Message) bool {
	if d.Session.State.User == nil {
		return false
	}
	return message.UserID() == d.Session.State.User.ID
}

// SendMessage sends a message.
func (d *Discord) SendMessage(channel, message string) (*discordgo.Message, error) {
	if channel == "" {
		log.Println("Empty channel could not send message", message)
		return nil, nil
	}

	m, err := d.Session.ChannelMessageSend(channel, message)
	if err != nil {
		log.Println("Error sending discord message: ", err)
		return nil, err
	}

	return m, nil
}

// SendMessageEmbed sends an embed.
func (d *Discord) SendMessageEmbed(channel string, embed *discordgo.MessageEmbed) (*discordgo.Message, error) {
	if channel == "" {
		log.Println("Empty channel could not send message")
		return nil, nil
	}

	m, err := d.Session.ChannelMessageSendEmbed(channel, embed)
	if err != nil {
		log.Println("Error sending discord message: ", err)
		return nil, err
	}

	return m, nil
}

// SendAction sends an action.
func (d *Discord) SendAction(channel, message string) (*discordgo.Message, error) {
	if channel == "" {
		log.Println("Empty channel could not send message", message)
		return nil, nil
	}

	p, err := d.UserChannelPermissions(d.UserID(), channel)
	if err != nil {
		return d.SendMessage(channel, message)
	}

	if p&discordgo.PermissionEmbedLinks == discordgo.PermissionEmbedLinks {
		if _, err := d.Session.ChannelMessageSendEmbed(channel, &discordgo.MessageEmbed{
			Color:       d.UserColor(d.UserID(), channel),
			Description: message,
		}); err != nil {
			return nil, err
		}
		return nil, nil
	}

	return d.SendMessage(channel, message)
}

// DeleteMessage deletes a message.
func (d *Discord) DeleteMessage(channel, messageID string) error {
	return d.Session.ChannelMessageDelete(channel, messageID)
}

// SendFile sends a file.
func (d *Discord) SendFile(channel, name string, r io.Reader) error {
	if _, err := d.Session.ChannelFileSend(channel, name, r); err != nil {
		log.Println("Error sending discord message: ", err)
		return err
	}
	return nil
}

// BanUser bans a user.
func (d *Discord) BanUser(channel, userID string, duration int) error {
	return d.Session.GuildBanCreate(channel, userID, 0)
}

// UnbanUser unbans a user.
func (d *Discord) UnbanUser(channel, userID string) error {
	return d.Session.GuildBanDelete(channel, userID)
}

// UserName returns the bots name.
func (d *Discord) UserName() string {
	if d.Session.State.User == nil {
		return ""
	}
	return d.Session.State.User.Username
}

// UserID returns the bots user id.
func (d *Discord) UserID() string {
	if d.Session.State.User == nil {
		return ""
	}
	return d.Session.State.User.ID
}

// Join accept an invite or return an error.
// If AlreadyJoinedError is return, @me has already accepted that invite.
func (d *Discord) Join(join string) error {
	if i, err := d.Session.Invite(join); err == nil {
		if _, err := d.Guild(i.Guild.ID); err == nil {
			return ErrAlreadyJoined
		}
	}

	if _, err := d.Session.InviteAccept(join); err != nil {
		return err
	}
	return nil
}

// Typing sets that the bot is typing.
func (d *Discord) Typing(channel string) error {
	return d.Session.ChannelTyping(channel)
}

// PrivateMessage will send a private message to a user.
func (d *Discord) PrivateMessage(userID, message string) (*discordgo.Message, error) {
	c, err := d.Session.UserChannelCreate(userID)
	if err != nil {
		return nil, err
	}
	return d.SendMessage(c.ID, message)
}

// SupportsPrivateMessages returns whether the service supports private messages.
func (d *Discord) SupportsPrivateMessages() bool {
	return true
}

// SupportsMultiline returns whether the service supports multiline messages.
func (d *Discord) SupportsMultiline() bool {
	return true
}

// CommandPrefix returns the command prefix for the service.
func (d *Discord) CommandPrefix() string {
	if len(os.Args) > 1 {
		return fmt.Sprintf("r.")
	}
	return fmt.Sprintf("rt.")
}

// IsBotOwner returns whether or not a message sender was the owner of the bot.
func (d *Discord) IsBotOwner(message Message) bool {
	return message.UserID() == d.OwnerUserID
}

// IsPrivate returns whether or not a message was private.
func (d *Discord) IsPrivate(message Message) bool {
	c, err := d.Channel(message.Channel())
	if err != nil {
		return false
	}
	return c.Type == 1
}

// IsChannelOwner returns whether or not the sender of a message is a moderator.
func (d *Discord) IsChannelOwner(message Message) bool {
	c, err := d.Channel(message.Channel())
	if err != nil {
		return false
	}
	g, err := d.Guild(c.GuildID)
	if err != nil {
		return false
	}
	return g.OwnerID == message.UserID() || d.IsBotOwner(message)
}

// IsModerator returns whether or not the sender of a message is a moderator.
func (d *Discord) IsModerator(message Message) bool {
	p, err := d.UserChannelPermissions(message.UserID(), message.Channel())
	if err == nil {
		if p&discordgo.PermissionAdministrator == discordgo.PermissionAdministrator || p&discordgo.PermissionManageChannels == discordgo.PermissionManageChannels || p&discordgo.PermissionManageServer == discordgo.PermissionManageServer {
			return true
		}
	}

	return d.IsChannelOwner(message)
}

// ChannelCount returns the number of channels the bot is in.
func (d *Discord) ChannelCount() int {
	return len(d.Guilds())
}

// GuildList returns an array of the current guilds and a separate array of the member count.
func (d *Discord) GuildList() ([]string, []int) {
	var total []string
	var memcount []int
	guilds := d.Guilds()

	for _, e := range guilds {
		total = append(total, e.Name)
		memcount = append(memcount, len(e.Members))
	}
	return total, memcount
}

// SupportsMessageHistory returns if the service supports message history.
func (d *Discord) SupportsMessageHistory() bool {
	return true
}

// MessageHistory returns the message history for a channel.
func (d *Discord) MessageHistory(channel string) []Message {
	c, err := d.Channel(channel)
	if err != nil {
		return nil
	}

	messages := make([]Message, len(c.Messages))
	for i := 0; i < len(c.Messages); i++ {
		messages[i] = &DiscordMessage{
			Discord:          d,
			DiscordgoMessage: c.Messages[i],
			MessageType:      MessageTypeCreate,
		}
	}

	return messages
}

func (d *Discord) Channel(channelID string) (channel *discordgo.Channel, err error) {
	for _, s := range d.Sessions {
		channel, err = s.State.Channel(channelID)
		if err == nil {
			return channel, nil
		}
	}
	return
}

func (d *Discord) Guild(guildID string) (guild *discordgo.Guild, err error) {
	for _, s := range d.Sessions {
		guild, err = s.State.Guild(guildID)
		if err == nil {
			return guild, nil
		}
	}
	return
}

func (d *Discord) Guilds() []*discordgo.Guild {
	guilds := []*discordgo.Guild{}
	for _, s := range d.Sessions {
		guilds = append(guilds, s.State.Guilds...)
	}
	return guilds
}

func (d *Discord) UserChannelPermissions(userID, channelID string) (apermissions int, err error) {
	for _, s := range d.Sessions {
		apermissions, err = s.State.UserChannelPermissions(userID, channelID)
		if err == nil {
			return apermissions, nil
		}
	}
	return
}

func (d *Discord) UserColor(userID, channelID string) int {
	for _, s := range d.Sessions {
		color := s.State.UserColor(userID, channelID)
		if color != 0 {
			return color
		}
	}
	return 0
}

func (d *Discord) Nickname(message Message) string {
	return d.NicknameForID(message.UserID(), message.UserName(), message.Channel())
}

func (d *Discord) NicknameForID(userID, userName, channelID string) string {
	c, err := d.Channel(channelID)
	if err == nil {
		g, err := d.Guild(c.GuildID)
		if err == nil {
			for _, m := range g.Members {
				if m.User.ID == userID {
					if m.Nick != "" {
						return m.Nick
					}
					break
				}
			}
		}
	}
	return userName
}

func (d *Discord) Member(gID, uID string) (*discordgo.Member, error) {
	return d.Session.GuildMember(gID, uID)
}

// TimestampForID takes a Discord ID and parses a timestamp from it
func (d *Discord) TimestampForID(id string) (time.Time, error) {
	_id, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return time.Unix(0, 0), err
	}
	return time.Unix(((_id>>22)+1420070400000)/1000, 0), nil
}

//
func (d *Discord) EditMessage(cID, mID, content string) (*discordgo.Message, error) {
	return d.Session.ChannelMessageEdit(cID, mID, content)
}
