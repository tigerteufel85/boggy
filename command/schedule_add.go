package command

import (
	"context"
	"fmt"
	"github.com/tigerteufel85/boggy/client"
	"github.com/tigerteufel85/boggy/config"
	"regexp"
	"strings"

	"github.com/nlopes/slack"
	"github.com/tigerteufel85/boggy/bot"
	"github.com/tigerteufel85/boggy/utils"
	"gopkg.in/andygrunwald/go-jira.v1"
)

type scheduleAdd struct {
	slackClient client.SlackClient
	jira        *jira.Client
	jiraCfg     config.JiraConfig
	jiraReplies config.ReplyConfig
	regex       config.RegexConfig
}

// NewScheduleAdd is a command to schedule a command via a cron
func NewScheduleAdd(slackClient client.SlackClient, jira *jira.Client, jiraCfg config.JiraConfig, jiraReplies config.ReplyConfig, regex config.RegexConfig) *scheduleAdd {
	return &scheduleAdd{
		slackClient,
		jira,
		jiraCfg,
		jiraReplies,
		regex,
	}
}

func (c *scheduleAdd) GetName() string {
	return "schedule add"
}

func (c *scheduleAdd) IsValid(b *bot.Bot, command string) bool {
	return false
}

func (c *scheduleAdd) getAllowedCommands() []bot.Command {
	return []bot.Command{
		NewJiraIssues(c.slackClient, c.jira, c.jiraCfg, c.jiraReplies, c.regex),
		NewJiraOverview(c.slackClient, c.jira, c.jiraCfg, c.jiraReplies, c.regex),
	}
}

func (c *scheduleAdd) isValidCommand(b *bot.Bot, command string) bool {
	for _, botCommand := range c.getAllowedCommands() {
		if botCommand.IsValid(b, command) {
			return true
		}
	}
	return false
}

func (c *scheduleAdd) Execute(ctx context.Context, b *bot.Bot, eventText string, event *slack.MessageEvent, user bot.User) bool {

	if !strings.HasPrefix(eventText, c.GetName()) {
		return false
	}

	// stop if user is not registered
	if user.Name == "" {
		c.slackClient.Respond(event, "I'm sorry, but it seems you are not allowed to add schedules")
		return true
	}

	_, err := utils.NewJQL(c.regex, eventText).BuildJqlQuery(c.jiraCfg)
	if err != nil {
		c.slackClient.Respond(event, err.Error())
		return true
	}

	project := utils.ParseRegex(eventText, c.regex.JiraProject)
	time := utils.ParseRegex(eventText, c.regex.CronTime)
	command := utils.ParseRegex(eventText, c.regex.CronCommand)

	var re = regexp.MustCompile("^([\\S]+)")
	time = re.ReplaceAllString(strings.Trim(time, " "), "0")

	if !b.IsValidSchedule(time) {
		c.slackClient.Respond(event, fmt.Sprintf("I'm sorry, but \"%s\" is not a valid schedule. Please check *@boggy help schedule>* for more information", time))
		return true
	}

	if !c.isValidCommand(b, command) {
		c.slackClient.Respond(event, fmt.Sprintf("I'm sorry, but \"%s\" is not a valid command. Please check *@boggy help <command>* for more information", command))
		return true
	}

	name := b.AddSchedule(b.NewSchedule(project, time, command, event.Channel, user))

	message := ""
	tasks := b.GetTasksForChannel(event.Channel)
	for _, value := range tasks {
		message += fmt.Sprintf("*%s* - `%s`\n```%s```\n", value.Name, value.Schedule.CronTime, value.Schedule.Command)
	}

	text := "The schedule was successfully added with id " + name + "!"
	if len(tasks) != 0 {
		text = "\nThe following schedules are currently active for this channel:"
	}

	attachment := slack.Attachment{
		Text: message,
	}

	c.slackClient.Respond(event, text, slack.MsgOptionAttachments(attachment))

	return true
}

func (c *scheduleAdd) GetHelp() []bot.Help {
	return []bot.Help{
		{
			"schedule add",
			"adds a new schedule by cron",
			[]string{
				"schedule add <cron:0 1/10 * * * *> <command>jira issues <project:foe> <type:bug> <prio>blocker</prio> <option:created> <time:10m></command>",
				"schedule add <cron:0 0 10 * * 1-5> <command>overview bugs <project:foe></command>",
			},
		},
	}
}
