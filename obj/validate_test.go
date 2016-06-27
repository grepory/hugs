package obj

import (
	"testing"

	log "github.com/opsee/logrus"
)

func TestValidatorIsValid(t *testing.T) {
	n := Notification{
		Id:         1,
		CustomerId: "test",
		UserId:     1,
		CheckId:    "test",
		Value:      "test",
		Type:       "test",
	}
	if err := n.Validate(); err != nil {
		t.FailNow()
	}
}

func TestValidatorIsInvalid(t *testing.T) {
	n := Notification{
		CustomerId: "test",
		UserId:     0,
		Type:       "test",
	}

	if err := n.Validate(); err == nil {
		t.FailNow()
	} else {
		log.WithFields(log.Fields{"test": "TestValidatorIsInvalid", "error": err}).Debug("received error.")
	}
}
