package configs

import (
	"errors"
	"fmt"
	"os"
	"testing"

	cliBase "github.com/kahnwong/cli-base"
	"github.com/rs/zerolog/log"
)

type Calendar struct {
	Id    string `yaml:"id"`
	Color string `yaml:"color"`
}

type Account struct {
	Name        string     `yaml:"name"`
	Credentials string     `yaml:"credentials"`
	Calendars   []Calendar `yaml:"calendars"`
}
type Config struct {
	Accounts []Account `yaml:"accounts"`
}

var AppConfigBasePath string
var AppConfig *Config

func init() {
	var err error
	AppConfigBasePath, err = cliBase.ExpandHome("~/.config/gcal-tui")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to expand config path")
	}

	configPath := fmt.Sprintf("%s/config.yaml", AppConfigBasePath)
	AppConfig, err = cliBase.ReadYaml[Config](configPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) && testing.Testing() {
			return
		}
		log.Fatal().Err(err).Msg("failed to read config")
	}
}
