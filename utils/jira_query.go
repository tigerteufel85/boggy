package utils

import (
	"fmt"
	"github.com/tigerteufel85/boggy/client"
	"github.com/tigerteufel85/boggy/config"
	"html"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// JQL is a wrapper with all information for a JIRA JQL
type JQL struct {
	Project     string
	Type        string
	Priority    string
	Status      string
	Option      string
	Time        string
	OffsetTime  string
	OffsetField string
	Custom      string
	Assignee    string
	Sorting     string
}

// NewJQL provides all information needed for a JIRA JQL
func NewJQL(config config.RegexConfig, input string) *JQL {
	return &JQL{
		Project:     ParseRegex(input, config.JiraProject),
		Type:        ParseRegex(input, config.JiraIssueType),
		Priority:    ParseRegex(input, config.JiraPriority),
		Status:      ParseRegex(input, config.JiraStatus),
		Option:      ParseRegex(input, config.JiraOption),
		Time:        ParseRegex(input, config.JiraTime),
		OffsetTime:  ParseRegex(input, config.JiraOffsetTime),
		OffsetField: ParseRegex(input, config.JiraOffsetField),
		Custom:      ParseRegex(input, config.JiraCustom),
		Assignee:    ParseRegex(input, config.JiraAssignee),
		Sorting:     ParseRegex(input, config.JiraSorting),
	}
}

// BuildJqlQuery creates a JIRA JQL based on the given information
func (jql *JQL) BuildJqlQuery(config config.JiraConfig) (string, error) {
	var results []string

	// verify project key
	if err := verifyProject(jql.Project); err != nil {
		return "", err
	}
	results = append(results, fmt.Sprintf("project = \"%s\"", jql.Project))

	// verify issue priorities
	if jql.Priority != "" {
		if err := verifyPriorities(config.Priorities, jql.Priority); err != nil {
			return "", err
		}
		results = append(results, fmt.Sprintf("priority in (%s)", jql.Priority))
	}

	// get resolutions
	if jql.Status != "" {
		results = append(results, fmt.Sprintf("resolution in (%s)", getStatus(config.Statuses, jql.Status)))
	}

	// prepare issue types
	if jql.Type != "" {
		results = append(results, fmt.Sprintf("issuetype in (\"%s\")", jql.Type))
	}

	// prepare option with time
	if jql.Option != "" && jql.Time != "" {
		switch strings.ToLower(jql.Option) {
		case "resolved", "created":
			results = append(results, fmt.Sprintf("%s >= -%s", jql.Option, jql.Time))
		}
	}

	// prepare field with offset
	if jql.OffsetField != "" && jql.OffsetTime != "" && jql.Time != "" {
		loc, _ := time.LoadLocation("Europe/Berlin")
		timein := time.Now().In(loc)
		timein = addTime(timein, strings.ToLower(jql.OffsetTime))
		timeout := addTime(timein, strings.ToLower(jql.Time))

		results = append(results, fmt.Sprintf(
			"\"%s\" >= -%s AND \"%s\" <= %s",
			jql.OffsetField,
			timein.Format(config.TimeFormat),
			jql.OffsetField,
			timeout.Format(config.TimeFormat),
		))
	}

	// prepare assignee with time
	if jql.Assignee != "" && jql.Time != "" {
		results = append(results, fmt.Sprintf("assignee changed TO %s DURING (-%s,now())", jql.Assignee, jql.Time))
	}

	// add custom jql
	if jql.Custom != "" {
		results = append(results, html.UnescapeString(jql.Custom))
	}

	result := strings.Join(results, " AND ")

	// apply sorting
	if val, ok := config.Sorting[jql.Sorting]; ok {
		return strings.Join([]string{result, val}, " "), nil
	}
	if val, ok := config.Sorting["default"]; ok {
		return strings.Join([]string{result, val}, " "), nil
	}
	return result, nil
}

func verifyProject(input string) error {
	if client.Projects[input] == "" {
		return fmt.Errorf("the provided project \"%s\" is invalid", input)
	}

	return nil
}

func verifyPriorities(priorities map[string]config.Priority, input string) error {
	verify := strings.Split(input, ",")
	for i := range verify {
		priority := strings.TrimSpace(verify[i])
		if _, found := priorities[priority]; !found {
			return fmt.Errorf("the provided priority \"%s\" is invalid", priority)
		}
	}

	return nil
}

func getStatus(statuses map[string]string, input string) string {
	if val, ok := statuses[strings.ToLower(input)]; ok {
		return val
	}

	return ""
}

func addTime(startTime time.Time, offset string) time.Time {
	timeMinutes := regexp.MustCompile(`^[0-9]*m$`)
	timeHours := regexp.MustCompile(`^[0-9]*h$`)
	timeDays := regexp.MustCompile(`^[0-9]*d$`)

	switch {
	case timeMinutes.MatchString(offset):
		timeOffset, _ := strconv.Atoi(strings.ReplaceAll(offset, "m", ""))
		startTime.Add(time.Minute * time.Duration(timeOffset))
	case timeHours.MatchString(offset):
		timeOffset, _ := strconv.Atoi(strings.ReplaceAll(offset, "h", ""))
		startTime.Add(time.Hour * time.Duration(timeOffset))
	case timeDays.MatchString(offset):
		timeOffset, _ := strconv.Atoi(strings.ReplaceAll(offset, "d", ""))
		startTime.Add(time.Hour * 24 * time.Duration(timeOffset))
	}

	return startTime
}
