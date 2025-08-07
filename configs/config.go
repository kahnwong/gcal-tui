package configs

import (
	"fmt"

	cliBase "github.com/kahnwong/cli-base"
)

type Calendar struct {
	Id    string `yaml:"id"`
	Color string `yaml:"color"`
}
type Config struct {
	Accounts []struct {
		Name        string     `yaml:"name"`
		Credentials string     `yaml:"credentials"`
		Calendars   []Calendar `yaml:"calendars"`
	} `yaml:"accounts"`
}

var AppConfigBasePath = cliBase.ExpandHome("~/.config/gcal-tui")
var AppConfig = cliBase.ReadYaml[Config](fmt.Sprintf("%s/config.yaml", AppConfigBasePath)) // init
