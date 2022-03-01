package bot

import (
	"context"
	"github.com/slack-go/slack"
)

// Commands is a wrapper of a list of commands. Only the first matched command will be executed
type Commands struct {
	Commands []Command
}

// Command is the main command struct which needs to provide the actual executed action, validation, a name and a help for the user
type Command interface {
	// return true in case command did a response
	Execute(ctx context.Context, b *Bot, eventText string, event *slack.MessageEvent, user User) bool

	// return true in case command can be scheduled and passed a basic check
	IsValid(b *Bot, command string) bool

	// each command has a name
	GetName() string

	// information on how to use the command with examples
	GetHelp() []Help
}

// GetHelp returns the help for ALL included commands
func (c *Commands) GetHelp() []Help {
	help := make([]Help, 0)

	for _, command := range c.Commands {
		help = append(help, command.GetHelp()...)
	}

	return help
}
