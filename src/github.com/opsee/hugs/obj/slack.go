package obj

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/jmoiron/sqlx/types"
	"github.com/opsee/hugs/util"
)

type SlackChannel struct {
	ID   string `json:"id" required:"true"`
	Name string `json:"name" required:"true"`
}

func (this *SlackChannel) Validate() error {
	validator := &util.Validator{}
	return validator.Validate(this)
}

type SlackChannels struct {
	Channels []*SlackChannel `json:"channels" required:"true"`
}

func (this *SlackChannels) Validate() error {
	for i, _ := range this.Channels {
		if err := this.Channels[i].Validate(); err != nil {
			return err
		}
	}
	return nil
}

type SlackOAuthResponseDBWrapper struct {
	ID         int            `json:"id" db:"id"`
	CustomerID string         `json:"customer_id" db:"customer_id" required:"true"`
	Data       types.JSONText `json:"data" db:"data" `
}

func (this *SlackOAuthResponseDBWrapper) Validate() error {
	validator := &util.Validator{}
	return validator.Validate(this)
}

// Oath response for slack token
type SlackOAuthResponse struct {
	AccessToken     string                `json:"access_token" db:"access_token" required:"true"`
	Scope           string                `json:"scope" db:"scope"`
	TeamName        string                `json:"team_name" db:"team_name"`
	TeamID          string                `json:"team_id" db:"team_id"`
	IncomingWebhook *SlackIncomingWebhook `json:"incoming_webhook" db:"incoming_webhook"`
	Bot             *SlackBotCreds        `json:"bot" db:"bot"`
	OK              bool                  `json:"ok" db:"ok"`
	Error           string                `json:"error" db:"error"`
}

func (this *SlackOAuthResponse) Validate() error {
	validator := &util.Validator{}
	return validator.Validate(this)
}

type SlackIncomingWebhook struct {
	URL              string `json:"url" db:"url" required:"true"`
	Channel          string `json:"channel" db:"channel" required:"true"`
	ConfigurationURL string `json:"configuration_url" db:"configuration_url" required:"true"`
}

func (this *SlackIncomingWebhook) Validate() error {
	validator := &util.Validator{}
	return validator.Validate(this)
}

type SlackBotCreds struct {
	BotUserID      string `json:"bot_user_id" db:"bot_user_id" required:"true"`
	BotAccessToken string `json:"bot_access_token" db:"bot_access_token" required:"true"`
}

func (this *SlackBotCreds) Validate() error {
	validator := &util.Validator{}
	return validator.Validate(this)
}

type SlackOAuthRequest struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Code         string `json:"code" required:"true"`
	RedirectURI  string `json:"redirect_uri" required:"true"`
}

func (this *SlackOAuthRequest) Validate() error {
	validator := &util.Validator{}
	return validator.Validate(this)
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
		return nil, err
	}

	return slackResponse, nil
}
