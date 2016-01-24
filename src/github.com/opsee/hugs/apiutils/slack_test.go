package apiutils

import (
	"testing"

	"github.com/opsee/hugs/config"
	log "github.com/sirupsen/logrus"
)

func TestSlack(t *testing.T) {
	go StartSlackAPIEmulator()

	blah := &SlackOAuthRequest{
		ClientID:     config.GetConfig().SlackTestClientID,
		ClientSecret: config.GetConfig().SlackTestClientSecret,
		Code:         "test",
	}

	response, err := blah.Do("http://localhost:7766/api/oauth.access")
	if err != nil {
		t.FailNow()
	}

	log.Info("Test Slack got response: ", response)
}
