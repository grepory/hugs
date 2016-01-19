package main

import (
	"github.com/opsee/hugs/config"
	"github.com/opsee/hugs/service"
	"github.com/opsee/hugs/sqsconsumer"
	"github.com/sirupsen/logrus"
)

func main() {

	go func() {
		logrus.Info("Starting SQS Consumer with ", config.GetConfig().MinWorkers, " workers (", config.GetConfig().MaxWorkers, " max).")
		foreman := sqsconsumer.NewForeman(10, 10, config.GetConfig().MaxWorkers, config.GetConfig().MinWorkers, config.GetConfig().SqsUrl, config.GetConfig().PostgresConn)
		foreman.Start()
	}()

	svc, err := service.NewService()
	if err != nil {
		logrus.Fatal("Unable to start service.")
	}
	svc.Start()
}
