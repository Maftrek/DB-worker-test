package application

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/gorilla/mux"
	"template_1/middleware"
	"template_1/models"
	"template_1/prometheus"
	"template_1/service"
)

// Application struct
type Application struct {
	r       *mux.Router
	serv    *http.Server
	pr      *prometheus.Prometheus
	svc     service.Service
	logger  log.Logger
	hashSum []byte
}

// Options struct
type Options struct {
	Pr      *prometheus.Prometheus
	Serv    models.ServerOpt
	Svc     service.Service
	Logger  log.Logger
	HashSum []byte
}

// New return new application
func New(opt *Options) *Application {
	r := mux.NewRouter()
	return &Application{
		r: r,
		serv: &http.Server{
			Addr:         ":8120",
			ReadTimeout:  time.Duration(opt.Serv.ReadTimeout),
			IdleTimeout:  time.Duration(opt.Serv.IdleTimeout),
			WriteTimeout: time.Duration(opt.Serv.WriteTimeout),
			Handler:      r,
		},
		hashSum: opt.HashSum,
		svc:     opt.Svc,
		pr:      opt.Pr,
		logger:  opt.Logger,
	}
}

// Start application
func (app *Application) Start() {
	app.logger.Log("server", "start", "port", app.serv.Addr)
	// love & support
	app.r.HandleFunc("/health", app.HealthHandler)
	app.r.Handle("/metrics", promhttp.Handler())

	// создание новости
	app.r.HandleFunc("/api/one_news/", app.createOneNews).Methods("POST")
	app.r.HandleFunc("/api/many_news/", app.createManyNews).Methods("POST")

	// обновление новости
	app.r.HandleFunc("/api/news/", app.updateNews).Methods("PUT")

	// получение новости
	app.r.HandleFunc("/api/news/{title}", app.getNews).Methods("GET")
	app.r.HandleFunc("/api/news/", app.getNewsAll).Methods("GET")

	app.r.Use(middleware.Metrics(app.pr))

	app.svc.Subscribe()

	listenErr := make(chan error, 1)

	go func() {
		listenErr <- app.serv.ListenAndServe()
	}()

	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-listenErr:
		if err != nil {
			app.logger.Log("err", err)
			os.Exit(1)
		}
	case <-osSignals:
		app.serv.SetKeepAlivesEnabled(false)
		app.svc.Close()
		timeout := time.Second * 5
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		if err := app.serv.Shutdown(ctx); err != nil {
			app.logger.Log("err", err)
			os.Exit(1)
		}
		cancel()
	}
}
