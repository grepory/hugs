package apiutils

import (
	"testing"

	log "github.com/sirupsen/logrus"
)

func TestSlack(t *testing.T) {
	go StartSlackAPIEmulator()

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
