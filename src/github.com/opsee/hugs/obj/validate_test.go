package obj

import (
	"testing"

	log "github.com/sirupsen/logrus"
)

func TestValidatorIsValid(t *testing.T) {
	n := Notification{
		ID:         1,
		CustomerID: "test",
		UserID:     1,
		CheckID:    "test",
		Value:      "test",
		Type:       "test",
	}
	if err := n.Validate(); err != nil {
		t.FailNow()
	}
}

func TestValidatorIsInvalid(t *testing.T) {
	n := Notification{
		CustomerID: "test",
		UserID:     0,
		Type:       "test",
	}

	if err := n.Validate(); err == nil {
		t.FailNow()
	} else {
		log.WithFields(log.Fields{"test": "TestValidatorIsInvalid", "error": err}).Debug("received error.")
	}
}
