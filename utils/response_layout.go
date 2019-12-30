package utils

import (
	"fmt"
	"github.com/nlopes/slack"
	"github.com/tigerteufel85/boggy/config"
	"gopkg.in/andygrunwald/go-jira.v1"
	"strings"
)

// Layout is a wrapper for layouting the Slack replies
type Layout struct {
	Option   string
	Project  string
	Assignee string
	Title    string
	Color    string
	List     string
}

// NewLayout holds all information for layouting a Slack reply
func NewLayout(config config.RegexConfig, input string) *Layout {
	return &Layout{
		Option:   ParseRegex(input, config.ReplyLayout),
		Project:  ParseRegex(input, config.JiraProject),
		Assignee: ParseRegex(input, config.JiraAssignee),
		Title:    ParseRegex(input, config.ReplyTitle),
		Color:    ParseRegex(input, config.ReplyColor),
		List:     ParseRegex(input, config.ReplyList),
	}
}

// BuildAttachment creates Slack Attachments for Slack replies
func (layout *Layout) BuildAttachment(replies config.ReplyConfig, config config.JiraConfig, amount int, issues []jira.Issue) slack.Attachment {
	option := layout.Option

	attachment := slack.Attachment{
		Color: replies.Colors.Grey,
		MarkdownIn: []string{
			"title",
			"text",
		},
	}

	var reply = replies.Jira["default"]
	if val, ok := replies.Jira[strings.ToLower(option)]; ok {
		reply = val
	}

	if reply.Text == "list" {
		attachment.Text = createList(config, layout.Project, issues, reply.OptionalFields)
	}

	switch reply.Parameter {
	case "amount":
		if reply.Title != "" {
			attachment.Title = fmt.Sprintf(reply.Title, amount)
		}
		if reply.Text != "list" {
			attachment.Text = fmt.Sprintf(reply.Text, amount)
		}
		break
	case "assignee":
		if reply.Title != "" {
			attachment.Title = fmt.Sprintf(reply.Title, layout.Assignee)
		}
		if reply.Text != "list" {
			attachment.Text = fmt.Sprintf(reply.Text, layout.Assignee)
		}
		break
	default:
		if reply.Title != "" {
			attachment.Title = reply.Title
		}
		if reply.Text != "list" {
			attachment.Text = reply.Text
		}
		break
	}

	attachment.Color = getReplyColor(reply.Color, layout.Color, replies, amount)

	// user defined layouts override predefined values
	if layout.Title != "" {
		attachment.Title = fmt.Sprintf(layout.Title+" %d", amount)
	}

	return attachment
}

// BuildAttachmentTicket creates a Slack Attachment with JIRA ticket information
func BuildAttachmentTicket(replies config.ReplyConfig, config config.JiraConfig, issue *jira.Issue) slack.Attachment {

	var fields []slack.AttachmentField
	var field slack.AttachmentField

	field.Title = "Type"
	field.Short = true
	field.Value = issue.Fields.Type.Name
	fields = append(fields, field)

	field.Title = "Status"
	field.Short = true
	field.Value = ""
	if issue.Fields.Status != nil {
		field.Value = issue.Fields.Status.Name
	}
	fields = append(fields, field)

	field.Title = "Components"
	field.Short = true
	field.Value = ""
	if len(issue.Fields.Components) > 0 {
		var components []string
		for _, component := range issue.Fields.Components {
			components = append(components, component.Name)
		}
		field.Value = strings.Join(components, ", ")
	}
	fields = append(fields, field)

	field.Title = "Assignee"
	field.Short = true
	field.Value = ""
	if issue.Fields.Assignee != nil {
		field.Value = " -> " + issue.Fields.Assignee.DisplayName
	}
	fields = append(fields, field)

	field.Title = "Description"
	field.Value = convertMarkdown(issue.Fields.Description)
	field.Short = false
	fields = append(fields, field)

	var title []string
	if issue.Fields.Priority != nil {
		title = append(title, getPriorityIcon(config.Priorities, issue.Fields.Priority.Name))
	}

	teamIcon := getTeamIcon(config.FeatureTeams, issue.Fields.Project.Key, issue.Fields)
	if teamIcon != "" {
		title = append(title, teamIcon)
	}

	title = append(title, issue.Fields.Summary)

	return slack.Attachment{
		Title:     strings.Join(title, " "),
		TitleLink: fmt.Sprintf("%s/browse/%s", config.Host, issue.Key),
		Color:     replies.Colors.Grey,
		Fields:    fields,
		MarkdownIn: []string{
			"text",
			"fields",
		},
	}
}

func createList(config config.JiraConfig, project string, issues []jira.Issue, optional bool) string {
	var output []string

	for _, issue := range issues {

		if issue.Fields.Priority != nil {
			output = append(output, getPriorityIcon(config.Priorities, issue.Fields.Priority.Name))
		}

		output = append(output, createJiraLink(config, issue.Key))

		teamIcon := getTeamIcon(config.FeatureTeams, project, issue.Fields)
		if teamIcon != "" {
			output = append(output, teamIcon)
		}

		output = append(output, issue.Fields.Summary)

		switch optional {
		case true:
			if issue.Fields.Status != nil {
				output = append(output, fmt.Sprintf("(%s)", issue.Fields.Status.Name))
			}
			if issue.Fields.Assignee != nil {
				output = append(output, fmt.Sprintf("\n-> %s", issue.Fields.Assignee.DisplayName))
			}
		case false:
			if len(issue.Fields.Components) > 0 {
				var components []string
				for _, component := range issue.Fields.Components {
					components = append(components, component.Name)
				}
				output = append(output, fmt.Sprintf("(%s)", strings.Join(components, ", ")))
			}
		}

		output = append(output, "\n")
	}

	return strings.Join(output, " ")
}

func convertMarkdown(content string) string {
	content = strings.Replace(content, "{code}", "```", -1)
	return content
}

func createJiraLink(config config.JiraConfig, issueKey string) string {
	return fmt.Sprintf("<%s/browse/%s|%s>", config.Host, issueKey, issueKey)
}

func getPriorityIcon(priorities map[string]config.Priority, id string) string {
	if val, ok := priorities[strings.ToLower(id)]; ok {
		return val.Icon
	}

	if val, ok := priorities["default"]; ok {
		return val.Icon
	}

	return ""
}

func getTeamIcon(config config.TeamConfig, project string, fields *jira.IssueFields) string {
	whitelist := make(map[string]string)
	for _, value := range config.Projects {
		whitelist[strings.ToLower(value)] = strings.ToLower(value)
	}
	if _, ok := whitelist[strings.ToLower(project)]; !ok {
		return ""
	}

	var team string
	featureTeam := fields.Unknowns[config.Field]
	if featureTeam != nil {
		team = featureTeam.(map[string]interface{})["value"].(string)
	}

	if val, ok := config.Teams[team]; ok {
		return val
	}

	return config.Default
}

func getReplyColor(replyColor string, layoutColor string, replies config.ReplyConfig, amount int) string {
	var attachmentColor = ""
	if replyColor != "" {

		if color, ok := replies.BugThresholds[replyColor]; ok {
			switch {
			case amount >= color.Danger:
				attachmentColor = replies.Colors.Red
				break
			case amount >= color.Warning:
				attachmentColor = replies.Colors.Yellow
				break
			default:
				attachmentColor = replies.Colors.Green
				break
			}
		} else {
			attachmentColor = replyColor
		}
	}

	// user defined layouts override predefined values
	if layoutColor != "" {
		attachmentColor = layoutColor
	}

	return attachmentColor
}
