package command

import (
	"context"
	"github.com/nlopes/slack"
	"github.com/tigerteufel85/boggy/bot"
	"github.com/tigerteufel85/boggy/client"
	"github.com/tigerteufel85/boggy/config"
	"github.com/tigerteufel85/boggy/utils"
	"gopkg.in/andygrunwald/go-jira.v1"
	"strings"
)

type jiraIssue struct {
	slackClient client.SlackClient
	jira        *jira.Client
	jiraCfg     config.JiraConfig
	jiraReplies config.ReplyConfig
}

// NewJiraIssue is a command to get information on a JIRA ticket
func NewJiraIssue(slackClient client.SlackClient, jira *jira.Client, jiraCfg config.JiraConfig, jiraReplies config.ReplyConfig) *jiraIssue {
	return &jiraIssue{
		slackClient,
		jira,
		jiraCfg,
		jiraReplies,
	}
}

func (c *jiraIssue) GetName() string {
	return "jira issue"
}

func (c *jiraIssue) IsValid(b *bot.Bot, command string) bool {
	return false
}

func (c *jiraIssue) Execute(ctx context.Context, b *bot.Bot, eventText string, event *slack.MessageEvent, user bot.User) bool {

	if !strings.HasPrefix(eventText, c.GetName()) {
		return false
	}
	eventText = c.slackClient.TrimMessage(eventText, c.GetName())

	issue, _, err := c.jira.Issue.Get(eventText, nil)
	if err != nil {
		auth, _ := c.slackClient.AuthTest()
		if user.Name != auth.User {
			c.slackClient.Respond(event, err.Error())
		}
		return true
	}

	attachments := slack.MsgOptionAttachments(utils.BuildAttachmentTicket(c.jiraReplies, c.jiraCfg, issue))

	c.slackClient.Respond(event, "", attachments)
	return true
}

func (c *jiraIssue) GetHelp() []bot.Help {
	return []bot.Help{
		{
			"jira issue",
			"lists information about a specific issue",
			[]string{
				"issue FOE-40000",
			},
		},
	}
}
