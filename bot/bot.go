package bot

import (
	"context"
	"fmt"
	"github.com/nlopes/slack"
	"github.com/tigerteufel85/boggy/client"
	"github.com/tigerteufel85/boggy/config"
	"strings"
	"time"
)

// Bot is a wrapper for the authentification, slack client, all commands, running crons and users allowed to schedule crons
type Bot struct {
	slackClient  *client.Slack
	auth         *slack.AuthTestResponse
	commands     Commands
	Crons        map[string]*CronTask
	AllowedUsers []User
}

// NewBot created main bot struct which holds the slack connection and dispatch messages to commands
func NewBot(slackClient *client.Slack, commands Commands) *Bot {
	return &Bot{
		slackClient: slackClient,
		commands:    commands,
	}
}

// Init establishes the slack connection, loads slack and user information
func (b *Bot) Init(config *config.Config) (err error) {
	fmt.Println(time.Now().UnixNano())

	b.auth, err = b.slackClient.AuthTest()
	if err != nil {
		return err
	}

	go b.slackClient.ManageConnection()

	err = b.loadChannelsUsersAndProjects(config)
	if err != nil {
		return err
	}

	fmt.Println(fmt.Sprintf("Loaded %d users and %d channels", len(client.Users)/2, len(client.Channels)/2))
	fmt.Println(fmt.Sprintf("Initialized %s with ID: %s", b.auth.User, b.auth.UserID))

	go b.InitTasks()

	go b.ReloadUserFile()
	go b.ReloadSlackUsers()
	go b.ReloadSlackChannels()

	for {
		select {
		case msg := <-b.slackClient.IncomingEvents:
			switch message := msg.Data.(type) {
			case *slack.MessageEvent:
				go b.handleMessage(message)
			}
		}
	}
}

// handleMessage processes the incoming message and responds appropriately
func (b *Bot) handleMessage(event *slack.MessageEvent) {
	if event.BotID != "" || event.User == "" || event.SubType == "bot_message" {
		return
	}
	eventText := strings.Trim(event.Text, " \n\r")
	if !b.isBotMessage(event, eventText) {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Hour)

	userName := client.Users[event.User]
	fmt.Println(fmt.Sprintf("%s channel: %s - user: %s - message: %s", time.Now().Format(time.RFC1123), event.Channel, userName, event.Text))

	// send "bot is typing message"
	go b.newTypingMessage(ctx, event.Channel, cancel)

	eventText = b.trimBot(eventText)
	for i := range b.commands.Commands {
		if b.commands.Commands[i].Execute(ctx, b, eventText, event, b.GetAllowedUser(userName)) {
			cancel()
			return
		}
	}
	b.slackClient.Respond(event, fmt.Sprintf("Oops! To err is human, for assistance try *_@%s help_*", b.auth.User))
	cancel()

}

// checks whether a message is meant for boggy either in a channel or direct
func (b *Bot) isBotMessage(event *slack.MessageEvent, eventText string) bool {
	if event.SubType == "internal" {
		return true
	}

	// Bot was mentioned in a public channel
	if strings.Contains(eventText, "<@"+b.auth.UserID+">") {
		return true
	}

	// Direct message channels always starts with 'D'
	if strings.HasPrefix(event.Channel, "D") {
		return true
	}
	return false
}

// sends a typing indicator for boggy
func (b *Bot) newTypingMessage(ctx context.Context, channel string, cancel context.CancelFunc) {
	for {
		b.slackClient.SendMessage(b.slackClient.NewTypingMessage(channel))

		select {
		case <-ctx.Done():
			cancel()
			return
		case <-time.After(time.Second * 1):
			// wait for 1 second
		}
	}
}

// remove @bot prefix of message and cleanup
func (b *Bot) trimBot(msg string) string {
	msg = strings.Replace(msg, "<@"+b.auth.UserID+"> ", "", 1)
	msg = strings.Replace(msg, "‘", "'", -1)
	msg = strings.Replace(msg, "’", "'", -1)

	return strings.Trim(msg, " ")
}
