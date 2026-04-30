// initial code is from google sdk docs
package gcal

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/kahnwong/gcal-tui/configs"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

func ReadOauthClientID(path string) (*oauth2.Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("unable to read client secret file: %w", err)
	}

	config, err := google.ConfigFromJSON(b, calendar.CalendarReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file to config: %w", err)
	}

	return config, nil
}

func GetClient(accountName string, config *oauth2.Config) (*http.Client, error) {
	tokFile := fmt.Sprintf("%s/%s-token.json", configs.AppConfigBasePath, accountName)
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		slog.Info("No valid token found, requesting new token from web")
		tok, err = getTokenFromWeb(accountName, config)
		if err != nil {
			return nil, fmt.Errorf("failed to get token from web: %w", err)
		}
		if err = saveToken(tokFile, tok); err != nil {
			return nil, fmt.Errorf("failed to save token: %w", err)
		}
	} else if !tok.Valid() || tok.Expiry.Before(time.Now().Add(5*time.Minute)) {
		// Token is invalid, expired, or expires within 5 minutes, try to refresh it
		slog.Info("Token expired or expiring soon, attempting to refresh")
		if tok.RefreshToken == "" {
			slog.Warn("No refresh token available, requesting new token from web")
			tok, err = getTokenFromWeb(accountName, config)
			if err != nil {
				return nil, fmt.Errorf("failed to get token from web: %w", err)
			}
		} else {
			tok, err = refreshToken(config, tok)
			if err != nil {
				slog.Warn("Failed to refresh token, requesting new token from web", "error", err)
				tok, err = getTokenFromWeb(accountName, config)
				if err != nil {
					return nil, fmt.Errorf("failed to get token from web: %w", err)
				}
			} else {
				slog.Info("Token refreshed successfully")
			}
		}
		if err = saveToken(tokFile, tok); err != nil {
			return nil, fmt.Errorf("failed to save token: %w", err)
		}
	} else {
		slog.Debug("Using existing valid token")
	}
	return config.Client(context.Background(), tok), nil
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer func(f *os.File) {
		if err := f.Close(); err != nil {
			slog.Warn("Unable to close token file", "error", err)
		}
	}(f)
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

func getTokenFromWeb(accountName string, config *oauth2.Config) (*oauth2.Token, error) {
	fmt.Printf("Account Name: %s\n", accountName)
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return nil, fmt.Errorf("unable to read authorization code: %w", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve token from web: %w", err)
	}
	return tok, nil
}

func saveToken(path string, token *oauth2.Token) error {
	slog.Debug(fmt.Sprintf("Saving credential file to: %s", path))
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("unable to cache oauth token: %w", err)
	}
	defer func(f *os.File) {
		if err := f.Close(); err != nil {
			slog.Warn("Error closing oauth token file", "error", err)
		}
	}(f)
	err = json.NewEncoder(f).Encode(token)
	if err != nil {
		return fmt.Errorf("unable to encode oauth token: %w", err)
	}
	return nil
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

	//slog.Debug("Successfully refreshed OAuth token")
	return newToken, nil
}
