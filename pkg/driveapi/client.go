package driveapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// GetDriveService initializes and returns a Google Drive service client using OAuth 2.0.
func GetDriveService(ctx context.Context) (*drive.Service, error) {
	client, err := getOAuthClient(ctx, drive.DriveScope)
	if err != nil {
		return nil, fmt.Errorf("failed to get OAuth client: %w", err)
	}

	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Printf("Unable to retrieve Drive client: %v", err)
		return nil, fmt.Errorf("unable to retrieve Drive client: %w", err)
	}

	return srv, nil
}

const (
	oauthClientSecretPath = "/app/secrets/Oauth.json" // Path where Oauth.json will be mounted in Docker
	tokenFilePath         = "/app/data/token.json"    // Path where token.json will be stored persistently
)

// getOAuthClient retrieves a token, or asks the user to authorize if needed.
func getOAuthClient(ctx context.Context, scope string) (*http.Client, error) {
	b, err := ioutil.ReadFile(oauthClientSecretPath)
	if err != nil {
		log.Printf("Unable to read client secret file from '%s': %v", oauthClientSecretPath, err)
		return nil, fmt.Errorf("unable to read client secret file from '%s': %w", oauthClientSecretPath, err)
	}

	config, err := google.ConfigFromJSON(b, scope)
	if err != nil {
		log.Printf("Unable to parse client secret file to config: %v", err)
		return nil, fmt.Errorf("unable to parse client secret file to config: %w", err)
	}

	// Ensure the directory for token.json exists
	tokenDir := filepath.Dir(tokenFilePath)
	if _, err := os.Stat(tokenDir); os.IsNotExist(err) {
		if err := os.MkdirAll(tokenDir, 0700); err != nil {
			return nil, fmt.Errorf("unable to create token directory '%s': %w", tokenDir, err)
		}
	}

	// Try to read the token from a file
	tok, err := tokenFromFile(tokenFilePath)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokenFilePath, tok)
	}
	return config.Client(ctx, tok), nil
}

// getTokenFromWeb uses a code to get a token from the web.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(oauth2.NoContext, authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
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
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache OAuth client token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}
