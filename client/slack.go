package client

import (
	"fmt"
	"github.com/nlopes/slack"
	"github.com/tigerteufel85/boggy/config"
	"strings"
)

// Users holds all slack users mapped from name -> id and vice versa from id -> name
var Users map[string]string

// Channels holds all slack channels mapped from name -> id and vice vers from id -> name
var Channels map[string]string

// Slack is a wrapper for the RTM connection
type Slack struct {
	slack.RTM
}

// GetSlackClient establishes a RTM connection to the slack server
func GetSlackClient(cfg config.SlackConfig) *Slack {
	rtm := slack.New(cfg.Token).NewRTM()
	slackClient := &Slack{RTM: *rtm}

	return slackClient
}

// SlackClient is the main slack interface
type SlackClient interface {
	Respond(event *slack.MessageEvent, text string, options ...slack.MsgOption) string
	TrimMessage(msg string, trim string) string
	AuthTest() (response *slack.AuthTestResponse, error error)
}

// Respond handles messages sent to Slack
func (s Slack) Respond(event *slack.MessageEvent, text string, options ...slack.MsgOption) string {
	if event.Channel == "" {
		return ""
	}

	if len(options) == 0 {
		if text == "" {
			return ""
		}
		options = make([]slack.MsgOption, 0)
	}

	defaultOptions := []slack.MsgOption{
		slack.MsgOptionTS(event.ThreadTimestamp), // send in current thread by default
		slack.MsgOptionAsUser(true),
		slack.MsgOptionText(text, false),
		slack.MsgOptionEnableLinkUnfurl(),
	}

	options = append(defaultOptions, options...)
	_, msgTimestamp, err := s.PostMessage(
		event.Channel,
		options...,
	)
	if err != nil {
		fmt.Println(event.Channel, text, err)
	}
	return msgTimestamp
}

// TrimMessage removes a specified from the message start and removes white spaces and line breaks in the beginning and end of the message
func (s Slack) TrimMessage(msg string, trim string) string {
	msg = strings.TrimPrefix(msg, trim)
	msg = strings.Trim(msg, " :\n")

	return msg
}
