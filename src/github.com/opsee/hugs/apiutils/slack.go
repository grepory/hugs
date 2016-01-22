package apiutils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
)

var SlackOAuthEndpoint = "https://slack.com/api/oauth.access"

// Oath response for slack token
type SlackOAuthResponse struct {
	AccessToken     string                `json:"access_token"`
	Scope           string                `json:"scope"`
	TeamName        string                `json:"team_name"`
	TeamID          string                `json:"team_id"`
	IncomingWebhook *SlackIncomingWebhook `json:"incoming_webhook"`
	Bot             *SlackBotCreds        `json:"bot"`
}

type SlackIncomingWebhook struct {
	URL              string `json:"url"`
	Channel          string `json:"channel"`
	ConfigurationURL string `json:"configuration_url"`
}

type SlackBotCreds struct {
	BotUserID      string `json:"bot_user_id"`
	BotAccessToken string `json:"bot_access_token"`
}

// Oauth request for slack token
type SlackOAuthRequest struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Code         string `json:"code"`
	RedirectURI  string `json:"redirect_uri"`
}

// NOTE: THis is for the router decoder validation...
func (this *SlackOAuthRequest) Validate() error {
	if this.Code == "" {
		return errors.New("There must at least be a code.")
	}
	return nil
}

func (this *SlackOAuthRequest) Do(endpoint string) (*SlackOAuthResponse, error) {
	/*
		https://slack.com/api/oauth.access
		client_id     - issued when you created your app (required)
		client_secret - issued when you created your app (required)
		code          - the code param (required)
		redirect_uri  - must match the originally submitted URI (if one was sent)
	*/
	datjson, err := json.Marshal(this)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(datjson))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	slackResponse := &SlackOAuthResponse{}
	dec := json.NewDecoder(resp.Body)
	dec.Decode(slackResponse)

	return slackResponse, nil
}

func StartSlackAPIEmulator() {
	oaResponse := &SlackOAuthResponse{
		AccessToken: "test",
		Scope:       "test",
		TeamName:    "test",
		TeamID:      "test",
		IncomingWebhook: &SlackIncomingWebhook{
			URL:              "test",
			Channel:          "test",
			ConfigurationURL: "test",
		},
		Bot: &SlackBotCreds{
			BotUserID:      "test",
			BotAccessToken: "test",
		},
	}
	oaResponseData, err := json.Marshal(oaResponse)
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("api/oath.access", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, string(oaResponseData), r.URL.Path)
	})

	log.Fatal(http.ListenAndServe(":7766", nil))
}
