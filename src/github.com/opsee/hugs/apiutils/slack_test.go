package apiutils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	log "github.com/sirupsen/logrus"
)

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

func init() {
	go StartSlackAPIEmulator()
}

func TestSlack(t *testing.T) {

	blah := &SlackOAuthRequest{
		ClientID:     "test",
		ClientSecret: "test",
		Code:         "test",
	}

	response, err := blah.Do("http://localhost:7766/api/oath.access")
	if err != nil {
		t.FailNow()
	}
	log.Info("Got response: ", response)
}
