package main

import (
	"flag"
	"github.com/tigerteufel85/boggy/bot"
	"github.com/tigerteufel85/boggy/client"
	"github.com/tigerteufel85/boggy/command"
	"github.com/tigerteufel85/boggy/config"
	"log"
)

func main() {
	log.SetFlags(log.Lshortfile)

	configFile := flag.String(
		"config",
		"config/config.yaml",
		"Path to config.yaml. Can be a glob like config/*.yaml",
	)
	flag.Parse()

	cfg, err := config.LoadPattern(*configFile)
	checkError(err)

	jiraClient, err := client.GetJiraClient(cfg.Jira)
	checkError(err)

	slackClient := client.GetSlackClient(cfg.Slack)

	commands := command.GetDefaultCommands(slackClient, jiraClient, *cfg)

	b := bot.NewBot(slackClient, commands)
	err = b.Init(cfg)
	checkError(err)
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}
