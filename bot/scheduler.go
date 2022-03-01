package bot

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/slack-go/slack"
	"gopkg.in/robfig/cron.v2"
)

var scheduleList = "config/schedule.list"

// Schedule is a wrapper for the command to be scheduled
type Schedule struct {
	Project  string
	Creator  User
	CronTime string
	Channel  string
	Command  string
}

// CronTask is a wrapper for the scheduled command
type CronTask struct {
	Name     string
	Schedule *Schedule
	Cron     *cron.Cron
}

// NewSchedule holds all information for creating a new schedule
func (b *Bot) NewSchedule(project string, time string, command string, channel string, user User) *Schedule {
	return &Schedule{
		Creator:  user,
		Channel:  channel,
		Project:  project,
		CronTime: time,
		Command:  command,
	}
}

// InitTasks initializes all cron commands
func (b *Bot) InitTasks() {
	file, err := os.Open(scheduleList)
	if err != nil {
		panic(err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(file)

	b.Crons = map[string]*CronTask{}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.Split(scanner.Text(), "|")
		schedule := &Schedule{
			CronTime: line[1],
			Project:  line[2],
			Command:  line[5],
			Channel:  line[4],
			Creator:  b.GetAllowedUser(line[3]),
		}
		b.Crons[line[0]] = b.createSchedule(line[0], schedule)
	}
}

func (b *Bot) resetTasks() {
	for _, value := range b.Crons {
		value.Cron.Stop()
	}
	b.InitTasks()
}

// AddSchedule is used to add a new schedule via a cron
func (b *Bot) AddSchedule(schedule *Schedule) string {
	name := fmt.Sprintf("%d", time.Now().UnixNano())

	file, err := os.OpenFile(scheduleList, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(file)

	message := fmt.Sprintf("%s|%s|%s|%s|%s|%s\n", name, schedule.CronTime, schedule.Project,
		schedule.Creator.Name, schedule.Channel, schedule.Command)
	if _, err := file.Write([]byte(message)); err != nil {
		log.Fatal(err)
	}

	b.resetTasks()
	return name
}

func (b *Bot) createSchedule(name string, schedule *Schedule) *CronTask {
	c := cron.New()
	_, err := c.AddFunc("TZ=Europe/Berlin "+schedule.CronTime, func() {
		createCommand(b, b.GetAllowedUser("boggy"), schedule.Command, schedule.Channel)
	})
	if err != nil {
		log.Fatal(err)
	}
	c.Start()

	return &CronTask{
		Name:     name,
		Schedule: schedule,
		Cron:     c,
	}
}

// DeleteSchedule removes a schedule from the active crons and from the schedule list
func (b *Bot) DeleteSchedule(name string) {
	// remove from file
	file, err := os.Open(scheduleList)
	if err != nil {
		log.Fatal(err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(file)

	outFile, _ := ioutil.ReadAll(file)
	re := regexp.MustCompile(name + ".*[\\r\\n]+")
	message := re.ReplaceAllString(string(outFile), "")

	out := []byte(message)
	if err := ioutil.WriteFile(scheduleList, out, 0666); err != nil {
		log.Fatal(err)
	}

	b.resetTasks()
}

// IsValidSchedule checks if a schedule can be parsed
func (b *Bot) IsValidSchedule(spec string) bool {
	_, err := cron.Parse(spec)
	if err != nil {
		fmt.Println(err)
		return false
	}

	return true
}

// GetTasksForChannel is returning all tasks which are active in the given channel
func (b *Bot) GetTasksForChannel(channel string) []*CronTask {
	var tasks []*CronTask
	for _, value := range b.Crons {
		if channel == value.Schedule.Channel || channel == "" {
			tasks = append(tasks, value)
		}
	}
	return tasks
}

// creates an actual command, user is the bot with admin rights to avoid side effects
func createCommand(b *Bot, user User, command string, channel string) {
	ctx := context.Background()

	event := new(slack.MessageEvent)
	event.Channel = channel

	auth, _ := b.slackClient.AuthTest()
	user.Name = auth.User
	user.Right = UserRightAdmin

	for i := range b.commands.Commands {
		if b.commands.Commands[i].Execute(ctx, b, command, event, user) {
			// command was handled
			return
		}
	}
}
