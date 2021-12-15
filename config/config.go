package config

// Config contains the full config structure of this bot
type Config struct {
	Slack   SlackConfig
	Jira    JiraConfig
	Replies ReplyConfig
	Regex   RegexConfig
}

// SlackConfig contains the credentials of the Slack client
type SlackConfig struct {
	Token string
}

// JiraConfig contains the credentials and configuration of the JIRA client
type JiraConfig struct {
	Host         string
	Username     string
	Password     string
	Projects     []string            `yaml:",flow"`
	Statuses     map[string]string   `yaml:",flow"`
	Priorities   map[string]Priority `yaml:",flow"`
	FeatureTeams TeamConfig
	TimeFormat   string
	BugOverview  struct {
		ListAll []string `yaml:",flow"`
		All     string
		Medium  string
	}
	Sorting map[string]string `yaml:",flow"`
}

// ReplyConfig contains the configuration for replies via Slack
type ReplyConfig struct {
	Jira map[string]struct {
		Title          string
		Text           string // use list for jira ticket list
		Parameter      string // amount/assignee can be replaced in title or text with %d/%s
		Color          string
		OptionalFields bool
	}
	BugThresholds map[string]struct {
		Danger  int
		Warning int
	}
	Colors struct {
		Red    string
		Yellow string
		Green  string
		Blue   string
		Grey   string
	}
}

// Priority contains the mapping of JIRA priority to Slack emoji
type Priority struct {
	Value string
	Icon  string
}

// TeamConfig contains the configuration for Feature Teams in JIRA
type TeamConfig struct {
	Field    string
	Projects []string
	Default  string
	Teams    map[string]string `yaml:",flow"`
}

// RegexConfig contains the various regex expressions to parse Slack messages for commands
type RegexConfig struct {
	JiraAssignee    string
	JiraCustom      string
	JiraIssueType   string
	JiraOption      string
	JiraPriority    string
	JiraProject     string
	JiraSorting     string
	JiraStatus      string
	JiraTime        string
	JiraOffsetTime  string
	JiraOffsetField string
	ReplyColor      string
	ReplyLayout     string
	ReplyList       string
	ReplyTitle      string
	CronCommand     string
	CronTime        string
}
