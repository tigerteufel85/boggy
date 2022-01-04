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

type jiraIssuesSingle struct {
	slackClient client.SlackClient
	jira        *jira.Client
	jiraCfg     config.JiraConfig
	jiraReplies config.ReplyConfig
	regex       config.RegexConfig
}

func NewJiraIssuesSingle(slackClient client.SlackClient, jira *jira.Client, jiraCfg config.JiraConfig, jiraReplies config.ReplyConfig, regex config.RegexConfig) *jiraIssuesSingle {
	return &jiraIssuesSingle{
		slackClient,
		jira,
		jiraCfg,
		jiraReplies,
		regex,
	}
}

func (c *jiraIssuesSingle) GetName() string {
	return "jira single"
}

func (c *jiraIssuesSingle) IsValid(b *bot.Bot, command string) bool {
	if !strings.HasPrefix(command, c.GetName()) {
		return false
	}

	_, err := utils.NewJQL(c.regex, command).BuildJqlQuery(c.jiraCfg)
	if err != nil {
		return false
	}

	return true
}

func (c *jiraIssuesSingle) Execute(ctx context.Context, b *bot.Bot, eventText string, event *slack.MessageEvent, user bot.User) bool {
	if !strings.HasPrefix(eventText, c.GetName()) {
		return false
	}

	// Create JQL Query
	jql := utils.NewJQL(c.regex, eventText)
	query, err := jql.BuildJqlQuery(c.jiraCfg)
	if err != nil {
		c.slackClient.Respond(event, err.Error())
		return true
	}

	// Search issues on JIRA
	issues, searchResponse, err := c.jira.Issue.Search(query, &jira.SearchOptions{MaxResults: 50, Expand: "names"})
	if err != nil {
		auth, _ := c.slackClient.AuthTest()
		if user.Name != auth.User {
			c.slackClient.Respond(event, err.Error())
		}
		return true
	}

	if searchResponse.Total == 0 {
		return true
	}

	// Get all fields
	allFields, _, err := c.jira.Field.GetList()
	if err != nil {
		auth, _ := c.slackClient.AuthTest()
		if user.Name != auth.User {
			c.slackClient.Respond(event, err.Error())
		}
		return true
	}

	for _, issue := range issues {
		customFields, _, err := c.jira.Issue.GetCustomFields(issue.ID)
		if err != nil {
			auth, _ := c.slackClient.AuthTest()
			if user.Name != auth.User {
				c.slackClient.Respond(event, err.Error())
			}
			return true
		}

		responseText := utils.NewLayout(c.regex, eventText).BuildSimpleTextResponse(c.jiraCfg, issue, customFields, allFields, jql)

		c.slackClient.Respond(event, responseText)
	}

	return true
}

func (c *jiraIssuesSingle) GetHelp() []bot.Help {
	return []bot.Help{
		{
			"jira issues single",
			"creates a slack response for each crm sale which will start",
			[]string{
				"jira issues single<project:crm><offset-field:Start Time><offset-time:1h><jql>“Campaign Category” = Sale</jql><time:10m>",
			},
		},
	}
}
