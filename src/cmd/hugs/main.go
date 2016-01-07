package main

import (
	"github.com/opsee/hugs/config"
	"github.com/opsee/hugs/sqsconsumer"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.Info("Starting SQS Consumer with ", config.GetConfig().MinWorkers, " workers (", config.GetConfig().MaxWorkers, " max).")
	foreman := sqsconsumer.NewForeman(10, 10, config.GetConfig().MaxWorkers, config.GetConfig().MinWorkers, config.GetConfig().SqsUrl, config.GetConfig().PostgresConn)
	foreman.Start()
}
