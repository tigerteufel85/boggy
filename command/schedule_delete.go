package command

import (
	"context"
	"fmt"
	"github.com/tigerteufel85/boggy/client"

	"github.com/slack-go/slack"
	"github.com/tigerteufel85/boggy/bot"
	"strings"
)

type scheduleDelete struct {
	slackClient client.SlackClient
}

// NewScheduleDelete is a command to delete a cron schedule
func NewScheduleDelete(slackClient client.SlackClient) *scheduleDelete {
	return &scheduleDelete{
		slackClient,
	}
}

func (c *scheduleDelete) GetName() string {
	return "schedule delete"
}

func (c *scheduleDelete) IsValid(b *bot.Bot, command string) bool {
	return false
}

func (c *scheduleDelete) Execute(ctx context.Context, b *bot.Bot, eventText string, event *slack.MessageEvent, user bot.User) bool {

	if !strings.HasPrefix(eventText, c.GetName()) {
		return false
	}
	cronId := c.slackClient.TrimMessage(eventText, c.GetName())

	cron := b.Crons[cronId]
	if cron == nil {
		c.slackClient.Respond(event, "I'm sorry, but I couldn't find a schedule with ID: "+cronId)
		return true
	}

	if b.Crons[cronId].Schedule.Channel != event.Channel && user.Right != bot.UserRightAdmin {
		c.slackClient.Respond(event, "I'm sorry, but you can only delete schedules from the current channel")
		return true
	}

	b.DeleteSchedule(cronId)

	message := fmt.Sprintf("The scheduled task has been deleted:\n```%s```", cron.Schedule.Command)

	attachment := slack.Attachment{Text: message}

	c.slackClient.Respond(event, "", slack.MsgOptionAttachments(attachment))
	return true
}

func (c *scheduleDelete) GetHelp() []bot.Help {
	return []bot.Help{
		{
			"schedule delete",
			"deletes a schedule with the provided id",
			[]string{
				"schedule delete 1511987987428336381",
			},
		},
	}
}
