package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"

	"lazy-lagoon/pkg/types"
	"github.com/kelseyhightower/envconfig"
)

/*
	Validate the env variables
*/
type GoogleDriveEnv struct {
	GoogleID           string `envconfig:"GOOGLE_ID"`
	GoogleSecret       string `envconfig:"GOOGLE_SECRET"`
	GoogleDeveloperKey string `envconfig:"GOOGLE_DEVELOPER_KEY"`
	GoogleRedirectURI  string `envconfig:"GOOGLE_REDIRECT_URI"`
}

var googleDriveEnv GoogleDriveEnv

func init() {
	if err := envconfig.Process("", &googleDriveEnv); err != nil {
		panic(err)
	}
}

type Credentials struct {
	ClientID                string   `json:"client_id"`
	ProjectID               string   `json:"project_id"`
	AuthURI                 string   `json:"auth_uri"`
	TokenURI                string   `json:"token_uri"`
	AuthProviderX509CertURL string   `json:"auth_provider_x509_cert_url"`
	ClientSecret            string   `json:"clientSecret"`
	RedirectUris            []string `json:"redirect_uris"`
}

type WebCredentials struct {
	Web Credentials `json:"web"`
}

func getClient(config *oauth2.Config, accessToken string) *http.Client {
	return config.Client(context.Background(), &oauth2.Token{AccessToken: accessToken})
}

/*
	Downloads content from Google Drive
*/
func DownloadFromGoogleDrive(input types.Input) ([]byte, error) {
	ctx := context.Background()

	creds := WebCredentials{
		Web: Credentials{
			ClientID:                googleDriveEnv.GoogleID,
			AuthURI:                 "https://accounts.google.com/o/oauth2/auth",
			TokenURI:                "https://oauth2.googleapis.com/token",
			AuthProviderX509CertURL: "https://www.googleapis.com/oauth2/v1/certs",
			ClientSecret:            googleDriveEnv.GoogleSecret,
			RedirectUris:            []string{googleDriveEnv.GoogleRedirectURI},
		},
	}

	jsonCreds, err := json.Marshal(creds)
	if err != nil {
		return nil, err
	}

	config, err := google.ConfigFromJSON(jsonCreds, drive.DriveMetadataReadonlyScope, drive.DriveFileScope)
	if err != nil {
		return nil, err
	}
	client := getClient(config, input.Credential.Secrets.AccessToken)

	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}

	resp, err := srv.Files.Get(input.Reference.Id).Download()
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	fileContents, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return fileContents, nil
}
