package main

import (
	"github.com/opsee/hugs/config"
	"github.com/opsee/hugs/consumer/nsq"
	"github.com/opsee/hugs/util"
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

	// worker's Id, error threshold prior to idle
	worker, err := nsq.NewWorker(util.RandomString(5))
	if err != nil {
		log.Fatal(err)
	}

	err = worker.Start()
	if err != nil {
		log.Fatal(err)
	}
}
