package obj

import (
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"strings"

	"github.com/jmoiron/sqlx/types"
	"github.com/nlopes/slack"
	"github.com/opsee/hugs/util"
	log "github.com/sirupsen/logrus"
)

// helper message to escape
func escapeMessage(message string) string {
	message = html.UnescapeString(message)
	replacer := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")
	return replacer.Replace(message)
}

type SlackChannel struct {
	Id   string `json:"id" required:"true"`
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

type SlackResponse struct {
	OK    bool   `json:"ok" db:"ok"`
	Error string `json:"error" db:"error"`
}

type SlackOAuthResponseDBWrapper struct {
	Id         int            `json:"id" db:"id"`
	CustomerId string         `json:"customer_id" db:"customer_id" required:"true"`
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
	TeamId          string                `json:"team_id" db:"team_id"`
	TeamDomain      string                `json:"team_domain" db:"team_domain"`
	IncomingWebhook *SlackIncomingWebhook `json:"incoming_webhook" db:"incoming_webhook"`
	Bot             *SlackBotCreds        `json:"bot" db:"bot"`
	SlackResponse
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
	BotUserId      string `json:"bot_user_id" db:"bot_user_id" required:"true"`
	BotAccessToken string `json:"bot_access_token" db:"bot_access_token" required:"true"`
}

func (this *SlackBotCreds) Validate() error {
	validator := &util.Validator{}
	return validator.Validate(this)
}

type SlackOAuthRequest struct {
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Code         string `json:"code" required:"true"`
	RedirectURI  string `json:"redirect_uri"`
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
		"client_id":     {this.ClientId},
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

	if !slackResponse.OK {
		err = fmt.Errorf("Slack Error: %s", slackResponse.Error)
		log.WithError(err).Error("Slack oauth.access failed.")
		return nil, err
	}

	return slackResponse, nil
}

type SlackPostChatMessageResponse struct {
	SlackResponse
}

type SlackPostChatMessageRequest struct {
	Token       string             `json:"token"`
	Channel     string             `json:"channel"`
	Text        string             `json:"text"`
	Username    string             `json:"username"`
	AsUser      bool               `json:"as_user"`
	Parse       string             `json:"parse"`
	LinkNames   int                `json:"link_names"`
	Attachments []slack.Attachment `json:"attachments"`
	UnfurlLinks bool               `json:"unfurl_links"`
	UnfurlMedia bool               `json:"unfurl_media"`
	IconURL     string             `json:"icon_url"`
	IconEmoji   string             `json:"icon_emoji"`
	Markdown    bool               `json:"mrkdwn,omitempty"`
	EscapeText  bool               `json:"escape_text"`
}

// unescapes for chars like ' " etc, escapes slack control characters
func (this *SlackPostChatMessageRequest) prepareText() {

	this.Text = escapeMessage(this.Text)
	for i, attachment := range this.Attachments {
		this.Attachments[i].Title = escapeMessage(attachment.Title)
		this.Attachments[i].Text = escapeMessage(attachment.Text)
		this.Attachments[i].Pretext = escapeMessage(attachment.Pretext)
	}
}

func (this *SlackPostChatMessageRequest) Do(endpoint string) (*SlackPostChatMessageResponse, error) {
	/*
		https://slack.com/api/chat.postMessage
	*/
	values := url.Values{
		"token":   {this.Token},
		"channel": {this.Channel},
	}

	this.prepareText()
	values.Set("username", string(this.Username))
	values.Set("as_user", "false")
	if this.Attachments != nil {
		attachments, err := json.Marshal(this.Attachments)
		if err != nil {
			return nil, err
		}
		values.Set("attachments", string(attachments))
	}
	values.Set("icon_url", this.IconURL)
	values.Set("parse", "full")

	resp, err := http.PostForm(endpoint, values)
	if err != nil {
		return nil, err
	}

	slackResponse := &SlackPostChatMessageResponse{}

	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&slackResponse)
	if err != nil {
		return nil, err
	}

	if !slackResponse.OK {
		err = fmt.Errorf("Slack Error: %s", slackResponse.Error)
		log.WithError(err).Error("Slack chat.postMessage failed.")
		return nil, err
	}

	return slackResponse, nil
}
