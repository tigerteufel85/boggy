package command

import (
	"context"
	"github.com/tigerteufel85/boggy/client"

	"fmt"
	"github.com/slack-go/slack"
	"github.com/tigerteufel85/boggy/bot"
	"strings"
)

type scheduleList struct {
	slackClient client.SlackClient
}

// NewScheduleList is a command to list all active schedules in a channel or all
func NewScheduleList(slackClient client.SlackClient) *scheduleList {
	return &scheduleList{
		slackClient,
	}
}

func (c *scheduleList) GetName() string {
	return "schedule list"
}

func (c *scheduleList) IsValid(b *bot.Bot, command string) bool {
	return false
}

func (c *scheduleList) Execute(ctx context.Context, b *bot.Bot, eventText string, event *slack.MessageEvent, user bot.User) bool {

	if !strings.HasPrefix(eventText, c.GetName()) {
		return false
	}

	option := c.slackClient.TrimMessage(eventText, c.GetName())

	channel := event.Channel
	if user.Right == bot.UserRightAdmin && option == "all" {
		channel = ""
	}

	message := ""
	tasks := b.GetTasksForChannel(channel)
	for _, value := range tasks {
		if user.Right == bot.UserRightAdmin && option == "all" && client.Channels[value.Schedule.Channel] != "" {
			value.Schedule.Channel = "#" + client.Channels[value.Schedule.Channel]
			message += fmt.Sprintf("%s - `%s` (%s)\n```%s```\n", value.Name, value.Schedule.CronTime, value.Schedule.Channel, value.Schedule.Command)
		} else {
			message += fmt.Sprintf("%s - `%s`\n```%s```\n", value.Name, value.Schedule.CronTime, value.Schedule.Command)
		}
	}

	text := "The following schedules are currently active"
	if user.Right != bot.UserRightAdmin || option == "all" {
		text += " for this channel"
	}
	text += ":"
	if len(tasks) == 0 {
		text = "There are currently no active schedules"
	}

	attachment := slack.Attachment{Text: message}

	c.slackClient.Respond(event, text, slack.MsgOptionAttachments(attachment))
	return true
}

func (c *scheduleList) GetHelp() []bot.Help {
	return []bot.Help{
		{
			"schedule list",
			"lists all schedules set for the current channel",
			[]string{
				"schedule list",
			},
		},
	}
}
