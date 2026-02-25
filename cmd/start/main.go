package main

import (
	"context"
	"os/signal"
	"syscall"

	"clean-arch-go/adapters"
	"clean-arch-go/bootstrap"
	"clean-arch-go/core/logs"
	"clean-arch-go/core/tracer"

	"github.com/sirupsen/logrus"
)

func main() {
	logs.Init()
	logger := logrus.NewEntry(logrus.StandardLogger())

	mysqlDB, err := adapters.NewMySQLConnection()
	if err != nil {
		logger.Fatalln("Can not connect to mysql", err)
	}

	traceProvider, err := tracer.SetupTracer()
	if err != nil {
		logger.Fatalln("Failed to setup tracer", err)
	}

	svc, err := bootstrap.New(mysqlDB, traceProvider, logger)
	if err != nil {
		logger.Fatalln("Failed to create service", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := svc.Run(ctx); err != nil {
		logger.WithError(err).Error("Service run error")
	}

	logger.Info("Service stopped")
}
