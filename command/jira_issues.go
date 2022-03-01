package command

import (
	"context"
	"github.com/slack-go/slack"
	"github.com/tigerteufel85/boggy/bot"
	"github.com/tigerteufel85/boggy/client"
	"github.com/tigerteufel85/boggy/config"
	"github.com/tigerteufel85/boggy/utils"
	"gopkg.in/andygrunwald/go-jira.v1"
	"strings"
)

type jiraIssues struct {
	slackClient client.SlackClient
	jira        *jira.Client
	jiraCfg     config.JiraConfig
	jiraReplies config.ReplyConfig
	regex       config.RegexConfig
}

// NewJiraIssues is a command to get information to a list of JIRA tickets
func NewJiraIssues(slackClient client.SlackClient, jira *jira.Client, jiraCfg config.JiraConfig, jiraReplies config.ReplyConfig, regex config.RegexConfig) *jiraIssues {
	return &jiraIssues{
		slackClient,
		jira,
		jiraCfg,
		jiraReplies,
		regex,
	}
}

func (c *jiraIssues) GetName() string {
	return "jira issues"
}

func (c *jiraIssues) IsValid(b *bot.Bot, command string) bool {
	if !strings.HasPrefix(command, c.GetName()) {
		return false
	}

	_, err := utils.NewJQL(c.regex, command).BuildJqlQuery(c.jiraCfg)
	if err != nil {
		return false
	}

	return true
}

func (c *jiraIssues) Execute(ctx context.Context, b *bot.Bot, eventText string, event *slack.MessageEvent, user bot.User) bool {

	if !strings.HasPrefix(eventText, c.GetName()) {
		return false
	}

	query, err := utils.NewJQL(c.regex, eventText).BuildJqlQuery(c.jiraCfg)
	if err != nil {
		c.slackClient.Respond(event, err.Error())
		return true
	}

	issues, response, err := c.jira.Issue.Search(query, &jira.SearchOptions{MaxResults: 50})
	if err != nil {
		auth, _ := c.slackClient.AuthTest()
		if user.Name != auth.User {
			c.slackClient.Respond(event, err.Error())
		}
		return true
	}

	if response.Total == 0 {
		return true
	}

	attachments := slack.MsgOptionAttachments(utils.NewLayout(c.regex, eventText).BuildAttachment(c.jiraReplies, c.jiraCfg, response.Total, issues))

	c.slackClient.Respond(event, "", attachments)

	return true
}

func (c *jiraIssues) GetHelp() []bot.Help {
	return []bot.Help{
		{
			"jira issues",
			"lists jira tickets matching parameters",
			[]string{
				"jira issues <project:foe>",
				"jira issues <project:foe> <type:bug>",
				"jira issues <project:foe> <status:open>",
				"jira issues <project:foe> <prio>blocker,critical,major</prio>",
				"jira issues <project:foe> <assignee:tigerteufel> <time:10m>",
				"jira issues <project:foe> <option:created> <time:10m>",
				"jira issues <project:foe> <jql>issuetype = bug AND created >=-100m</jql>",
			},
		},
	}
}
