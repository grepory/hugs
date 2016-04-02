package config

import (
	"testing"

	log "github.com/sirupsen/logrus"
)

func TestGetConfig(t *testing.T) {
	config := GetConfig()
	log.WithFields(log.Fields{"test": "TestGetConfig", "config": config}).Info("Success.")
}
