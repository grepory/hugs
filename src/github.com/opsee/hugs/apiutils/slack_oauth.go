package apiutils

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

var SlackOAuthEndpoint = "https://slack.com/api/oauth.access"

// Oath response for slack token
type SlackOAuthResponse struct {
	AccessToken     string                `json:"access_token" db:"access_token"`
	Scope           string                `json:"scope" db:"scope"`
	TeamName        string                `json:"team_name" db:"team_name"`
	TeamID          string                `json:"team_id" db:"team_id"`
	IncomingWebhook *SlackIncomingWebhook `json:"incoming_webhook" db:"incoming_webhook"`
	Bot             *SlackBotCreds        `json:"bot" db:"bot"`
}

type SlackIncomingWebhook struct {
	URL              string `json:"url" db:"url"`
	Channel          string `json:"channel" db:"channel"`
	ConfigurationURL string `json:"configuration_url" db:"configuration_url"`
}

type SlackBotCreds struct {
	BotUserID      string `json:"bot_user_id" db:"bot_user_id"`
	BotAccessToken string `json:"bot_access_token" db:"bot_access_token"`
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

	values := url.Values{
		"client_id":     {this.ClientID},
		"client_secret": {this.ClientSecret},
		"code":          {this.Code},
		"redirect_uri":  {this.RedirectURI},
	}

	resp, err := http.PostForm(endpoint, values)
	if err != nil {
		return nil, err
	}

	slackResponse := &SlackOAuthResponse{}

	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&slackResponse)
	if err != nil {
		fmt.Printf("%T\n%s\n%#v\n", err, err, err)
	}

	return slackResponse, nil
}
