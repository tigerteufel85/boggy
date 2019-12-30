package client

import (
	"github.com/tigerteufel85/boggy/config"
	"gopkg.in/andygrunwald/go-jira.v1"
)

// Projects is a map of all projects which should be accessible by boggy via Slack
// This doesn't handle the permissions which need to be granted via JIRA
var Projects map[string]string

// GetJiraClient establishes a connection to the JIRA server
func GetJiraClient(cfg config.JiraConfig) (*jira.Client, error) {
	jiraTransport := &jira.BasicAuthTransport{
		Username: cfg.Username,
		Password: cfg.Password,
	}

	return jira.NewClient(jiraTransport.Client(), cfg.Host)
}
