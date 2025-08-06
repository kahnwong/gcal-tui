// initial code is from google sdk docs
package gcal

import (
	"context"
	"encoding/json"
	"fmt"
	cliBase "github.com/kahnwong/cli-base"
	"github.com/kahnwong/gcal-tui/configs"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"net/http"
	"os"
	"time"
)

func ReadOauthClientID() *oauth2.Config {
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
	tokFile := fmt.Sprintf("%s/%s-token.json", configs.AppConfigBasePath, configs.AppConfig.Accounts[0].Name) // [TODO] loop through all accounts
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		log.Info().Msg("No valid token found, requesting new token from web")
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	} else if !tok.Valid() || tok.Expiry.Before(time.Now().Add(5*time.Minute)) {
		// Token is invalid, expired, or expires within 5 minutes, try to refresh it
		log.Info().Msg("Token expired or expiring soon, attempting to refresh")
		if tok.RefreshToken == "" {
			log.Warn().Msg("No refresh token available, requesting new token from web")
			tok = getTokenFromWeb(config)
		} else {
			tok, err = refreshToken(config, tok)
			if err != nil {
				log.Warn().Err(err).Msg("Failed to refresh token, requesting new token from web")
				tok = getTokenFromWeb(config)
			} else {
				log.Info().Msg("Token refreshed successfully")
			}
		}
		saveToken(tokFile, tok)
	} else {
		log.Debug().Msg("Using existing valid token")
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
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline, oauth2.ApprovalForce)
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

func refreshToken(config *oauth2.Config, token *oauth2.Token) (*oauth2.Token, error) {
	if token.RefreshToken == "" {
		return nil, fmt.Errorf("no refresh token available - please re-authenticate to grant offline access")
	}

	tokenSource := config.TokenSource(context.Background(), token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	// Ensure we preserve the refresh token if it's not included in the response
	if newToken.RefreshToken == "" && token.RefreshToken != "" {
		newToken.RefreshToken = token.RefreshToken
	}

	//log.Debug().Msg("Successfully refreshed OAuth token")
	return newToken, nil
}
