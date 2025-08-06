package configs

import (
	cliBase "github.com/kahnwong/cli-base"
)

type Config struct {
	Accounts []struct {
		Name        string   `yaml:"name"`
		Credentials string   `yaml:"credentials"`
		Calendars   []string `yaml:"calendars"`
	} `yaml:"accounts"`
}

var AppConfig = cliBase.ReadYaml[Config]("~/.config/gcal-tui/config.yaml") // init
