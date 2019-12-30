package bot

import (
	"github.com/nlopes/slack"
	"github.com/tigerteufel85/boggy/client"
	"time"
)

// ReloadUserFile handles updating the user file every 10 seconds
func (b *Bot) ReloadUserFile() {
	for range time.Tick(time.Second * 10) {
		b.AllowedUsers = getAllowedUsers()
	}
}

// ReloadSlackUsers handles updating the client.Users map, so new slack users will be added within 24 hours
func (b *Bot) ReloadSlackUsers() {
	for range time.Tick(time.Hour * 24) {
		users, _ := b.slackClient.GetUsers()
		client.Users = map[string]string{}

		for _, user := range users {
			client.Users[user.Name] = user.ID
			client.Users[user.ID] = user.Name
		}
	}
}

// ReloadSlackChannels handles updating client.Channels, so new channels will be added within an hour
func (b *Bot) ReloadSlackChannels() {
	for range time.Tick(time.Hour * 1) {
		params := slack.GetConversationsParameters{Types: []string{"private_channel", "public_channel"}}
		channels, nextCursor, err := b.slackClient.GetConversations(&params)
		if err != nil {
			continue
		}

		for _, channel := range channels {
			client.Channels[channel.ID] = channel.Name
			client.Channels[channel.Name] = channel.ID
		}
		for nextCursor != "" {
			params = slack.GetConversationsParameters{Cursor: nextCursor, Types: []string{"private_channel", "public_channel"}}
			channelPage, cursor, err := b.slackClient.GetConversations(&params)
			if err != nil {
				continue
			}

			nextCursor = cursor
			for _, channel := range channelPage {
				client.Channels[channel.ID] = channel.Name
				client.Channels[channel.Name] = channel.ID
			}
		}
	}
}
