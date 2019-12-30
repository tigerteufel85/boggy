package command

import (
	"context"
	"fmt"
	"github.com/tigerteufel85/boggy/client"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/nlopes/slack"
	"github.com/tigerteufel85/boggy/bot"
)

type adminRemoveUser struct {
	slackClient client.SlackClient
}

// NewAdminRemoveUser is an admin command to remove users to be able to schedule crons
func NewAdminRemoveUser(slackClient client.SlackClient) *adminRemoveUser {
	return &adminRemoveUser{
		slackClient,
	}
}

func (c *adminRemoveUser) GetName() string {
	return "user remove"
}

func (c *adminRemoveUser) IsValid(b *bot.Bot, command string) bool {
	return false
}

func (c *adminRemoveUser) Execute(ctx context.Context, b *bot.Bot, eventText string, event *slack.MessageEvent, user bot.User) bool {

	// stop if it is not about removing a user
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

	// check if user exists
	remName := textPart[0]
	remUser := b.GetAllowedUser(remName)
	if remUser.Name == "" {
		c.slackClient.Respond(event, fmt.Sprintf("Sorry, but user %s does not exist!", remName))
		return true
	}

	file, err := os.Open(b.GetUserFile())
	if err != nil {
		log.Fatal(err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(file)

	outFile, _ := ioutil.ReadAll(file)
	re := regexp.MustCompile("(?m)[\r\n]+^.*" + remName + ".*$")
	message := re.ReplaceAllString(string(outFile), "")

	out := []byte(message)
	if err := ioutil.WriteFile(b.GetUserFile(), out, 0666); err != nil {
		log.Fatal(err)
	}

	c.slackClient.Respond(event, fmt.Sprintf("User %s was successfully removed", remName))
	return true
}

func (c *adminRemoveUser) GetHelp() []bot.Help {
	return []bot.Help{}
}
