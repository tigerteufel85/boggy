package command

import (
	"github.com/tigerteufel85/boggy/bot"
	"github.com/tigerteufel85/boggy/client"
	"github.com/tigerteufel85/boggy/config"
	"gopkg.in/andygrunwald/go-jira.v1"
)

// GetDefaultCommands returns a list of all commands
func GetDefaultCommands(slackClient client.SlackClient, jira *jira.Client, config config.Config) bot.Commands {
	var commands bot.Commands

	commands = bot.Commands{
		Commands: []bot.Command{
			NewHelp(slackClient, &commands),
			NewJiraIssues(slackClient, jira, config.Jira, config.Replies, config.Regex),
			NewJiraIssuesSingle(slackClient, jira, config.Jira, config.Replies, config.Regex),
			NewJiraIssue(slackClient, jira, config.Jira, config.Replies),
			NewJiraOverview(slackClient, jira, config.Jira, config.Replies, config.Regex),
			NewAdminAddUser(slackClient),
			NewAdminRemoveUser(slackClient),
			NewScheduleAdd(slackClient, jira, config.Jira, config.Replies, config.Regex),
			NewScheduleDelete(slackClient),
			NewScheduleList(slackClient),
		},
	}

	return commands
}
