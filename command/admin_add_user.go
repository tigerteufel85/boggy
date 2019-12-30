package command

import (
	"context"
	"fmt"
	"github.com/tigerteufel85/boggy/client"
	"log"
	"os"
	"strings"

	"github.com/nlopes/slack"
	"github.com/tigerteufel85/boggy/bot"
)

type adminAddUser struct {
	slackClient client.SlackClient
}

// NewAdminAddUser is a command to allow a user to create new schedules
func NewAdminAddUser(slackClient client.SlackClient) *adminAddUser {
	return &adminAddUser{
		slackClient,
	}
}

func (c *adminAddUser) GetName() string {
	return "user add"
}

func (c *adminAddUser) IsValid(b *bot.Bot, command string) bool {
	return false
}

func (c *adminAddUser) Execute(ctx context.Context, b *bot.Bot, eventText string, event *slack.MessageEvent, user bot.User) bool {

	// stop if it is not about adding a user
	if !strings.HasPrefix(eventText, c.GetName()) {
		return false
	}
	eventText = c.slackClient.TrimMessage(eventText, c.GetName())

	// stop if user is not registered
	if user.Name == "" || user.Right != bot.UserRightAdmin {
		c.slackClient.Respond(event, bot.NotRegistered)
		return true
	}

	textPart := strings.Split(eventText, " ")
	if len(textPart) == 0 {
		return false
	}

	// check if user was already added
	addName := textPart[0]
	addUser := b.GetAllowedUser(addName)
	if addUser.Name != "" {
		c.slackClient.Respond(event, fmt.Sprintf("Sorry, but %s is already registered", addName))
		return true
	}

	// check if user exists in slack
	if _, ok := client.Users[addName]; !ok {
		c.slackClient.Respond(event, fmt.Sprintf("Sorry, but user %s does not seem to exist in Slack", addName))
		return true
	}

	// check if right is valid
	if len(textPart) > 1 && textPart[1] != bot.UserRightAdmin {
		c.slackClient.Respond(event, fmt.Sprintf("Sorry, %s is not a valid right", textPart[2]))
		return true
	}
	addRights := ""
	if len(textPart) > 1 {
		addRights = textPart[1]
	}

	// add user to file
	file, err := os.OpenFile(b.GetUserFile(), os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(file)

	message := strings.ToLower(fmt.Sprintf("%s\n", addName))
	if addRights != "" {
		message = strings.ToLower(fmt.Sprintf("%s;%s\n", addName, addRights))
	}
	if _, err := file.Write([]byte(message)); err != nil {
		log.Fatal(err)
	}

	c.slackClient.Respond(event, fmt.Sprintf("User %s was successfully added", addName))
	return true
}

func (c *adminAddUser) GetHelp() []bot.Help {
	return []bot.Help{}
}
