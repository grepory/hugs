package main

import (
	"github.com/opsee/hugs/config"
	"github.com/opsee/hugs/service"
	log "github.com/sirupsen/logrus"
	"github.com/yeller/yeller-golang"
)

func main() {
	yeller.Start(config.GetConfig().YellerAPIKey)
	defer func() {
		if r := recover(); r != nil {
			yeller.NotifyPanic(r)
		}
	}()

	svc, err := service.NewService()
	if err != nil {
		log.Fatal("Unable to start service: ", err)
	}
	svc.Start()
}
