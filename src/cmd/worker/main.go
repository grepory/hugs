package main

import (
	"github.com/opsee/hugs/config"
	"github.com/opsee/hugs/sqsconsumer"
	"github.com/opsee/hugs/util"
	log "github.com/sirupsen/logrus"
)

func main() {
	// worker's ID, error threshold prior to idle
	worker, err := sqsconsumer.NewWorker(util.RandomString(5), 12, config.GetConfig().SqsUrl)
	if err != nil {
		log.Fatal(err)
	}
	worker.Start()
}
