// initial code is from google sdk d ocs
package gcal

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	cliBase "github.com/kahnwong/cli-base"
	"github.com/kahnwong/gcal-tui/configs"
	_ "github.com/kahnwong/gcal-tui/internal/logger"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

func ReadOauthClientIDJSON() *oauth2.Config {
	b, err := os.ReadFile(cliBase.ExpandHome(configs.AppConfig.Accounts[0].Credentials)) // [TODO] loop through all accounts
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to read client secret file")
	}

	config, err := google.ConfigFromJSON(b, calendar.CalendarReadonlyScope)
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to parse client secret file to config")
	}

	return config
}

func GetClient(config *oauth2.Config) *http.Client {
	tokFile := fmt.Sprintf("%s/%s-token.json", cliBase.ExpandHome("~/.config/gcal-tui"), configs.AppConfig.Accounts[0].Name) // [TODO] loop through all accounts
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Fatal().Err(err).Msg("Unable to close token file")
		}
	}(f)
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatal().Err(err).Msg("Unable to read authorization code")
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to retrieve token from web")
	}
	return tok
}

func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to cache oauth token")
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Fatal().Err(err).Msg("Error closing oauth token file")
		}
	}(f)
	err = json.NewEncoder(f).Encode(token)
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to cache oauth token")
	}
}
