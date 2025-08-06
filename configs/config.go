package configs

import (
	"fmt"
	cliBase "github.com/kahnwong/cli-base"
)

type Config struct {
	Accounts []struct {
		Name        string   `yaml:"name"`
		Credentials string   `yaml:"credentials"`
		Calendars   []string `yaml:"calendars"`
	} `yaml:"accounts"`
}

var AppConfigBasePath = cliBase.ExpandHome("~/.config/gcal-tui")
var AppConfig = cliBase.ReadYaml[Config](fmt.Sprintf("%s/config.yaml", AppConfigBasePath)) // init
