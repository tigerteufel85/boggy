package config

import (
	"fmt"
	"github.com/imdario/mergo"
	"gopkg.in/andygrunwald/go-jira.v1"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
	"reflect"
)

// LoadPattern loads config yaml file(s) by a glob pattern
func LoadPattern(pattern string) (*Config, error) {
	fileNames, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	for _, fileName := range fileNames {
		newCfg, err := loadConfig(fileName)
		if err != nil {
			return nil, err
		}

		if err := mergo.Merge(cfg, newCfg, mergo.WithAppendSlice); err != nil {
			return nil, err
		}
	}

	err = verifyAuthConfig(cfg)
	if err != nil {
		return nil, err
	}

	err = verifyMandatoryFields(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func loadConfig(filename string) (*Config, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to load config file from %s: %s", filename, err)
	}

	cfg := &Config{}
	if err := yaml.UnmarshalStrict(content, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse configuration file: %s", err)
	}

	return cfg, nil
}

func verifyAuthConfig(cfg *Config) error {
	if cfg.Slack.Token == "" {
		return fmt.Errorf("slack config does not have a token")
	}

	jiraTransport := &jira.BasicAuthTransport{
		Username: cfg.Jira.Username,
		Password: cfg.Jira.Password,
	}
	client, _ := jira.NewClient(jiraTransport.Client(), cfg.Jira.Host)
	_, _, err := client.User.GetSelf()
	if err != nil {
		return fmt.Errorf("jira auth failed, please check your jira config which requires a host, username and password")
	}

	return nil
}

func verifyMandatoryFields(cfg *Config) error {
	for _, element := range []string{"default", "overviewall", "overviewmedium", "overviewlist"} {
		if _, ok := cfg.Replies.Jira[element]; !ok {
			return fmt.Errorf("could not find %s in jira replies configuration", element)
		}
	}

	v := reflect.ValueOf(cfg.Replies.Colors)
	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).CanInterface() && v.Field(i).Interface() == "" {
			return fmt.Errorf("could not find %s in replies color configuration", v.Type().Field(i).Name)
		}
	}

	v = reflect.ValueOf(cfg.Regex)
	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).CanInterface() && v.Field(i).Interface() == "" {
			return fmt.Errorf("could not find %s in regex configuration", v.Type().Field(i).Name)
		}
	}

	return nil
}
