package google

import (
	"encoding/gob"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"code.google.com/p/goauth2/oauth"
	"github.com/docker/machine/libmachine/log"
	raw "google.golang.org/api/compute/v1"
)

const (
	AuthURL      = "https://accounts.google.com/o/oauth2/auth"
	TokenURL     = "https://accounts.google.com/o/oauth2/token"
	ClientId     = "22738965389-8arp8bah3uln9eoenproamovfjj1ac33.apps.googleusercontent.com"
	ClientSecret = "qApc3amTyr5wI74vVrRWAfC_"
	RedirectURI  = "urn:ietf:wg:oauth:2.0:oob"
)

func newGCEService(storePath, authTokenPath string) (*raw.Service, error) {
	client := newOauthClient(storePath, authTokenPath)
	service, err := raw.New(client)
	return service, err
}

func newOauthClient(storePath, authTokenPath string) *http.Client {
	config := &oauth.Config{
		ClientId:     ClientId,
		ClientSecret: ClientSecret,
		Scope:        raw.ComputeScope,
		AuthURL:      AuthURL,
		TokenURL:     TokenURL,
	}

	token := token(storePath, authTokenPath, config)
	t := oauth.Transport{
		Token:     token,
		Config:    config,
		Transport: http.DefaultTransport,
	}
	return t.Client()
}

func token(storePath, authTokenPath string, config *oauth.Config) *oauth.Token {
	tokenPath := authTokenPath
	if authTokenPath == "" {
		tokenPath = filepath.Join(storePath, "gce_token")
	}
	log.Debugf("using auth token: %s", tokenPath)
	token, err := tokenFromCache(tokenPath)
	if err != nil {
		token = tokenFromWeb(config)
		saveToken(storePath, token)
	}
	return token
}

func tokenFromCache(tokenPath string) (*oauth.Token, error) {
	f, err := os.Open(tokenPath)
	if err != nil {
		return nil, err
	}
	token := new(oauth.Token)
	err = gob.NewDecoder(f).Decode(token)
	return token, err
}

func tokenFromWeb(config *oauth.Config) *oauth.Token {
	randState := fmt.Sprintf("st%d", time.Now().UnixNano())

	config.RedirectURL = RedirectURI
	authURL := config.AuthCodeURL(randState)

	log.Info("Opening auth URL in browser.")
	log.Info(authURL)
	log.Info("If the URL doesn't open please open it manually and copy the code here.")
	openURL(authURL)
	code := getCodeFromStdin()

	log.Infof("Got code: %s", code)

	t := &oauth.Transport{
		Config:    config,
		Transport: http.DefaultTransport,
	}
	_, err := t.Exchange(code)
	if err != nil {
		log.Fatalf("Token exchange error: %v", err)
	}
	return t.Token
}

func getCodeFromStdin() string {
	fmt.Print("Enter code: ")
	var code string
	fmt.Scanln(&code)
	return strings.Trim(code, "\n")
}

func openURL(url string) {
	try := []string{"xdg-open", "google-chrome", "open"}
	for _, bin := range try {
		err := exec.Command(bin, url).Run()
		if err == nil {
			return
		}
	}
}

func saveToken(storePath string, token *oauth.Token) {
	tokenPath := path.Join(storePath, "gce_token")
	log.Infof("Saving token in %v", tokenPath)
	f, err := os.Create(tokenPath)
	if err != nil {
		log.Infof("Warning: failed to cache oauth token: %v", err)
		return
	}
	defer f.Close()
	gob.NewEncoder(f).Encode(token)
}
