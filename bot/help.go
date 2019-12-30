package bot

// Help can be provided by a command to add information within "help" command
type Help struct {
	Command     string
	Description string
	Examples    []string
}
