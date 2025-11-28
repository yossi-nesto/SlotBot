package calendar

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

// getOAuthClient retrieves a token, saves the token, then returns the generated client.
func getOAuthClient(config *oauth2.Config) (*http.Client, error) {
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok), nil
}

// getTokenFromWeb uses Config to request a Token and returns the retrieved Token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("\n🔐 Go to the following link in your browser:\n%v\n\n", authURL)
	fmt.Printf("After authorization, paste the code here: ")

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		slog.Error("Unable to read authorization code", "error", err)
		return nil
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		slog.Error("Unable to retrieve token from web", "error", err)
		return nil
	}
	return tok
}

// tokenFromFile retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// saveToken saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	slog.Info("Saving credential file", "path", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		slog.Error("Unable to cache oauth token", "error", err)
		return
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

// getOAuthConfig creates OAuth2 config from credentials file
func getOAuthConfig(credentialsFile string) (*oauth2.Config, error) {
	b, err := os.ReadFile(credentialsFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read client secret file: %w", err)
	}

	config, err := google.ConfigFromJSON(b, calendar.CalendarScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file to config: %w", err)
	}

	return config, nil
}
