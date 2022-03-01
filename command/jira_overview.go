package command

import (
	"context"
	"github.com/tigerteufel85/boggy/client"
	"github.com/tigerteufel85/boggy/config"
	"strings"

	"github.com/slack-go/slack"
	"github.com/tigerteufel85/boggy/bot"
	"github.com/tigerteufel85/boggy/utils"
	"gopkg.in/andygrunwald/go-jira.v1"
)

type jiraOverview struct {
	slackClient client.SlackClient
	jira        *jira.Client
	jiraCfg     config.JiraConfig
	jiraReplies config.ReplyConfig
	regex       config.RegexConfig
}

// NewJiraOverview is a command to get an overview/report of bugs in a project
func NewJiraOverview(slackClient client.SlackClient, jira *jira.Client, jiraCfg config.JiraConfig, jiraReplies config.ReplyConfig, regex config.RegexConfig) *jiraOverview {
	return &jiraOverview{
		slackClient,
		jira,
		jiraCfg,
		jiraReplies,
		regex,
	}
}

func (c *jiraOverview) GetName() string {
	return "overview bugs"
}

func (c *jiraOverview) IsValid(b *bot.Bot, command string) bool {
	if !strings.HasPrefix(command, c.GetName()) {
		return false
	}

	_, err := utils.NewJQL(c.regex, command).BuildJqlQuery(c.jiraCfg)
	if err != nil {
		return false
	}

	return true
}

func (c *jiraOverview) Execute(ctx context.Context, b *bot.Bot, eventText string, event *slack.MessageEvent, user bot.User) bool {

	if !strings.HasPrefix(eventText, c.GetName()) {
		return false
	}

	auth, _ := c.slackClient.AuthTest()

	// get all bugs
	bugsAll := eventText + c.jiraCfg.BugOverview.All
	queryAll, err := utils.NewJQL(c.regex, bugsAll).BuildJqlQuery(c.jiraCfg)
	if err != nil {
		c.slackClient.Respond(event, err.Error())
		return true
	}
	issuesAll, responseAll, err := c.jira.Issue.Search(queryAll, &jira.SearchOptions{MaxResults: 50})
	if err != nil {
		if user.Name != auth.User {
			c.slackClient.Respond(event, err.Error())
		}
		return true
	}

	// get medium+ bugs
	bugsMedium := eventText + c.jiraCfg.BugOverview.Medium
	queryMedium, err := utils.NewJQL(c.regex, bugsMedium).BuildJqlQuery(c.jiraCfg)
	if err != nil {
		c.slackClient.Respond(event, err.Error())
		return true
	}
	issuesMedium, responseMedium, err := c.jira.Issue.Search(queryMedium, &jira.SearchOptions{MaxResults: 50})
	if err != nil {
		if user.Name != auth.User {
			c.slackClient.Respond(event, err.Error())
		}
		return true
	}

	jql := utils.NewJQL(c.regex, eventText)
	layout := utils.NewLayout(c.regex, eventText+"<layout:overviewlist>")

	var projectIssues = issuesMedium
	whitelist := make(map[string]string)
	for _, value := range c.jiraCfg.BugOverview.ListAll {
		whitelist[strings.ToLower(value)] = strings.ToLower(value)
	}
	if _, ok := whitelist[strings.ToLower(jql.Project)]; ok {
		projectIssues = issuesAll
	}

	attachments := slack.MsgOptionAttachments(
		utils.NewLayout(c.regex, eventText+"<layout:overviewall>").BuildAttachment(c.jiraReplies, c.jiraCfg, responseAll.Total, issuesAll),
		utils.NewLayout(c.regex, eventText+"<layout:overviewmedium>").BuildAttachment(c.jiraReplies, c.jiraCfg, responseMedium.Total, issuesMedium),
		layout.BuildAttachment(c.jiraReplies, c.jiraCfg, 0, projectIssues),
	)

	c.slackClient.Respond(event, "", attachments)
	return true
}

func (c *jiraOverview) GetHelp() []bot.Help {
	return []bot.Help{
		{
			"overview bugs",
			"provides an overview of open bugs",
			[]string{
				"overview bugs <project:foe>",
			},
		},
	}
}
