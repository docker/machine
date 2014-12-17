package google

import (
	"encoding/gob"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"code.google.com/p/goauth2/oauth"
	log "github.com/Sirupsen/logrus"
	raw "google.golang.org/api/compute/v1"
)

const (
	AuthURL      = "https://accounts.google.com/o/oauth2/auth"
	TokenURL     = "https://accounts.google.com/o/oauth2/token"
	ClientId     = "22738965389-8arp8bah3uln9eoenproamovfjj1ac33.apps.googleusercontent.com"
	ClientSecret = "qApc3amTyr5wI74vVrRWAfC_"
)

func newGCEService(storePath string) (*raw.Service, error) {
	client := newOauthClient(storePath)
	service, err := raw.New(client)
	return service, err
}

func newOauthClient(storePath string) *http.Client {
	config := &oauth.Config{
		ClientId:     ClientId,
		ClientSecret: ClientSecret,
		Scope:        raw.ComputeScope,
		AuthURL:      AuthURL,
		TokenURL:     TokenURL,
	}
	token := token(storePath, config)
	t := oauth.Transport{
		Token:     token,
		Config:    config,
		Transport: http.DefaultTransport,
	}
	return t.Client()
}

func token(storePath string, config *oauth.Config) *oauth.Token {
	token, err := tokenFromCache(storePath)
	if err != nil {
		token = tokenFromWeb(config)
		saveToken(storePath, token)
	}
	return token
}

func tokenFromCache(storePath string) (*oauth.Token, error) {
	tokenPath := path.Join(storePath, "gce_token")
	f, err := os.Open(tokenPath)
	if err != nil {
		return nil, err
	}
	token := new(oauth.Token)
	err = gob.NewDecoder(f).Decode(token)
	return token, err
}

func tokenFromWeb(config *oauth.Config) *oauth.Token {
	ch := make(chan string)
	randState := fmt.Sprintf("st%d", time.Now().UnixNano())
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/favicon.ico" {
			http.Error(rw, "", 404)
			return
		}
		if req.FormValue("state") != randState {
			log.Debugf("State doesn't match: req = %#v", req)
			http.Error(rw, "", 500)
			return
		}
		if code := req.FormValue("code"); code != "" {
			fmt.Fprintf(rw, "<h1>Success</h1>Authorized. Code: %v", code)
			rw.(http.Flusher).Flush()
			ch <- code
			return
		}
		log.Fatalf("no code")
		http.Error(rw, "", 500)
	}))
	defer ts.Close()

	config.RedirectURL = ts.URL
	authURL := config.AuthCodeURL(randState)

	log.Info("Opening auth URL in browser.")
	log.Info(authURL)
	log.Info("If the URL doesn't open, please open it manually and copy the code here.")
	go openURL(authURL)
	go getCodeFromStdin(ch)

	code := <-ch
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

func getCodeFromStdin(ch chan<- string) {
	fmt.Print("Enter code: ")
	var code string
	fmt.Scanln(&code)
	code = strings.Trim(code, "\n")
	// Under 'go test', stdin is /dev/null.
	if code != "" {
		ch <- code
	}
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
