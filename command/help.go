package command

import (
	"context"
	"fmt"
	"github.com/tigerteufel85/boggy/client"
	"strings"

	"github.com/nlopes/slack"
	"github.com/tigerteufel85/boggy/bot"
	"regexp"
	"sort"
)

type help struct {
	slackClient client.SlackClient
	commands    *bot.Commands
}

// NewHelp is a command to provide helpful information for various commands
func NewHelp(slackClient client.SlackClient, commands *bot.Commands) *help {
	return &help{
		slackClient,
		commands,
	}
}

func (c *help) GetName() string {
	return "help"
}

func (c *help) IsValid(b *bot.Bot, command string) bool {
	return false
}

func (c *help) Execute(ctx context.Context, b *bot.Bot, eventText string, event *slack.MessageEvent, user bot.User) bool {

	var re = regexp.MustCompile("^(?i:help)(.*)")
	match := re.FindAllStringSubmatch(eventText, 1)
	if len(match) == 0 {
		return false
	}
	eventText = strings.Trim(match[0][1], " ")

	help, names := c.buildHelpTree()
	text := ""
	if eventText == "" {
		auth, _ := c.slackClient.AuthTest()
		text = fmt.Sprintf("Hello <@%s>, I’m <@%s>. You want me to show you around?\n", event.User, auth.User)
		text += "I currently listen to the following commands:\n "
		for _, name := range names {
			text += fmt.Sprintf("- *%s*", name)
			if len(help[name].Description) > 0 {
				text += fmt.Sprintf(" _(%s)_", help[name].Description)
			}
			text += "\n"
		}

		text += fmt.Sprintf("With *_@%s help <command>_* I can provide you with more details!\n", auth.User)
		text += "More details can also be found <https://github.com/tigerteufel85/boggy|» here>"
	} else {
		commandHelp, ok := help[eventText]
		if !ok {
			c.slackClient.Respond(event, fmt.Sprintf("Invalid command: `%s`", eventText))
			return false
		}

		text += fmt.Sprintf("*%s command*:\n", commandHelp.Command)
		if len(commandHelp.Description) > 0 {
			text += commandHelp.Description + ": \n"
		}
		text += "*Some examples:*\n"
		for _, example := range commandHelp.Examples {
			text += fmt.Sprintf(" - %s\n", example)
		}
	}

	attachment := slack.Attachment{
		Text:  text,
		Color: "#121c99",
		MarkdownIn: []string{
			"text",
			"fields",
		},
	}

	c.slackClient.Respond(event, "", slack.MsgOptionAttachments(attachment))
	return true
}

func (c *help) buildHelpTree() (map[string]bot.Help, []string) {
	var names []string
	help := map[string]bot.Help{}
	for _, commandHelp := range c.commands.GetHelp() {
		if _, ok := help[commandHelp.Command]; ok {
			// main command already defined
			continue
		}
		help[commandHelp.Command] = commandHelp
		names = append(names, commandHelp.Command)
	}
	sort.Strings(names)

	return help, names
}

func (c *help) GetHelp() []bot.Help {
	return []bot.Help{
		{
			"help",
			"displays all available commands",
			[]string{
				"help",
				"help jira issue",
			},
		},
	}
}
