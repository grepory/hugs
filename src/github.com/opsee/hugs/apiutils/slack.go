package apiutils

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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
	http.HandleFunc("/api/oauth.access", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, string(oaResponseData), r.URL.Path)
	})

	log.Fatal(http.ListenAndServe(":7766", nil))
}
