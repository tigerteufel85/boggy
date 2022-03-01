package bot

import (
	"bufio"
	"fmt"
	"github.com/slack-go/slack"
	"github.com/tigerteufel85/boggy/client"
	"github.com/tigerteufel85/boggy/config"
	"log"
	"os"
	"strings"
)

const (
	// NotRegistered is the default reply when rights don't match
	NotRegistered = "I'm sorry, but you are not allowed to use this command."

	// UserRightUser is the default user right for any user
	UserRightUser = "user"
	// UserRightAdmin is the user right for admins who can use some hidden features
	UserRightAdmin = "admin"
)

var userList = "config/user.list"

// User is a wrapper with the user name and right
type User struct {
	Name  string
	Right string
}

// GetAllowedUser returns all users which are allowed to handle schedules
func (b *Bot) GetAllowedUser(slackUser string) User {
	var botUser User
	for _, user := range b.AllowedUsers {
		if slackUser == user.Name {
			botUser = user
		}
	}
	return botUser
}

// GetUserFile returns the path to the user file where the allowed users are stored
func (b *Bot) GetUserFile() string {
	return userList
}

func getAllowedUsers() []User {
	file, err := os.Open(userList)
	if err != nil {
		panic(err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(file)

	var users []User
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.Split(scanner.Text(), ";")

		right := UserRightUser
		if len(line) > 1 {
			right = line[1]
		}

		botUser := User{
			Name:  line[0],
			Right: right,
		}

		users = append(users, botUser)
	}
	return users
}

// load the public and private channels and list of all users from current space
func (b *Bot) loadChannelsUsersAndProjects(config *config.Config) error {
	fmt.Println("...Loading Channels")
	var err error
	var cursor string
	var channels []slack.Channel
	client.Channels = make(map[string]string)

	for {
		params := slack.GetConversationsParameters{
			Limit: 1000,
			Cursor: cursor,
			ExcludeArchived: true,
			Types: []string{"private_channel", "public_channel"},
		}

		channels, cursor, err = b.slackClient.GetConversations(&params)
		if err != nil {
			return err
		}

		for _, channel := range channels {
			client.Channels[channel.ID] = channel.Name
			client.Channels[channel.Name] = channel.ID
		}

		if cursor == "" {
			break
		}
	}

	fmt.Println("...Loading Users")
	users, err := b.slackClient.GetUsers()
	if err != nil {
		return err
	}
	client.Users = make(map[string]string)
	for _, user := range users {
		client.Users[user.Name] = user.ID
		client.Users[user.ID] = user.Name
	}

	fmt.Println("...Loading Projects")
	client.Projects = make(map[string]string)
	for _, project := range config.Jira.Projects {
		client.Projects[project] = project
		client.Projects[strings.ToLower(project)] = project
	}
	return nil
}
