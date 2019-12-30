package utils

import (
	"html"
	"regexp"
)

const (
	// RegexQuotes contains quotes to be replaced by regex which are sent via Slack
	RegexQuotes = "([“”])"
)

// ParseRegex replaces text by a specified regex and fixes Slack quotes
func ParseRegex(text string, regex string) string {
	var re = regexp.MustCompile(regex)

	text = html.UnescapeString(text)
	text = replaceRegex(text, RegexQuotes)

	match := re.FindAllStringSubmatch(text, 1)

	result := ""
	if len(match) > 0 {
		result = match[0][1]
	}

	return result
}

func replaceRegex(text string, regex string) string {
	var re = regexp.MustCompile(regex)
	return re.ReplaceAllLiteralString(text, "\"")
}
