package main

import (
	"github.com/opsee/hugs/service"
	log "github.com/sirupsen/logrus"
)

func main() {
	svc, err := service.NewService()
	if err != nil {
		log.Fatal("Unable to start service: ", err)
	}
	svc.Start()
}
