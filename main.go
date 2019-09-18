package main

import (
	"DB-worker-test/application"
	"DB-worker-test/models"
	"DB-worker-test/provider"
	"DB-worker-test/repository"
	"DB-worker-test/service"
	"fmt"
	"os"

	"DB-worker-test/prometheus"
	"github.com/go-kit/kit/log"
)

var (
	appConfig models.Config
	pr        *prometheus.Prometheus
	logger    log.Logger
)

func init() {
	models.LoadConfig(&appConfig)
	fmt.Println(appConfig.NatsServer)
}

func main() {
	pr = prometheus.New("new-test")
	logger = log.With(
		log.NewJSONLogger(os.Stderr),
		"caller", log.DefaultCaller,
	)
	logger = prometheus.NewLogger(logger, pr)

	p := provider.New(&appConfig.SQLDataBase, &appConfig.NoSQLDataBase)
	err := p.Open("postgres", appConfig.NatsServer.Address)
	if err != nil {
		logger.Log("err", err)
		os.Exit(1)
	}

	repo := repository.New(p)

	svc := service.New(repo)

	app := application.New(&application.Options{
		Serv:    appConfig.ServerOpt,
		HashSum: appConfig.HashSum,
		Svc:     svc,
		Pr:      pr,
		Logger:  logger,
	})

	app.Start()
}
