package store

import (
	"testing"

	"github.com/opsee/hugs/obj"
	log "github.com/opsee/logrus"
)

func TestStorePutSlackOAuthResponse(t *testing.T) {
	slackOAuthResponse := &obj.SlackOAuthResponse{
		AccessToken: "test",
		Scope:       "test",
		TeamName:    "test",
		TeamId:      "test",
		IncomingWebhook: &obj.SlackIncomingWebhook{
			URL:              "test",
			Channel:          "test",
			ConfigurationURL: "test",
		},
		Bot: &obj.SlackBotCreds{
			BotUserId:      "test",
			BotAccessToken: "test",
		},
	}

	err := Common.DBStore.PutSlackOAuthResponse(Common.User, slackOAuthResponse)
	if err != nil {
		log.Error(err)
		t.FailNow()
	}
}

func TestStoreUpdateSlackOAuthResponse(t *testing.T) {
	slackOAuthResponse := &obj.SlackOAuthResponse{
		AccessToken: "test",
		Scope:       "test",
		TeamName:    "feck",
		TeamId:      "test",
		IncomingWebhook: &obj.SlackIncomingWebhook{
			URL:              "test",
			Channel:          "test",
			ConfigurationURL: "test",
		},
		Bot: &obj.SlackBotCreds{
			BotUserId:      "test",
			BotAccessToken: "test",
		},
	}

	err := Common.DBStore.UpdateSlackOAuthResponse(Common.User, slackOAuthResponse)
	if err != nil {
		log.Error(err)
		t.FailNow()
	}
}

func TestStoreGetSlackOAuthResponse(t *testing.T) {
	response, err := Common.DBStore.GetSlackOAuthResponse(Common.User)
	if err != nil {
		log.Error(err)
		t.FailNow()
	}
	log.Info("Got OAuthResponse: ", response)
}

func TestStoreGetSlackOAuthResponses(t *testing.T) {
	responses, err := Common.DBStore.GetSlackOAuthResponses(Common.User)
	if err != nil {
		log.Error(err)
		t.FailNow()
	}

	if len(responses) == 0 {
		t.FailNow()
	}

	log.Info("Got OAuthResponse: ", responses[0])
}
