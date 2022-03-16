package main

import (
	"flag"
	"github.com/lipandr/yandex-practicum-diploma/internal/types"
	"log"

	"github.com/caarlos0/env/v6"
	"github.com/lipandr/yandex-practicum-diploma/internal/app"
	"github.com/lipandr/yandex-practicum-diploma/internal/client"
	"github.com/lipandr/yandex-practicum-diploma/internal/config"
	"github.com/lipandr/yandex-practicum-diploma/internal/dao"
	"github.com/lipandr/yandex-practicum-diploma/internal/service"
)

func main() {
	var cfg config.Config
	if err := env.Parse(&cfg); err != nil {
		log.Fatal(err)
	}
	flag.StringVar(&cfg.RunAddress, "a",
		cfg.RunAddress, "Address and port to start the service")
	flag.StringVar(&cfg.DatabaseURI, "d",
		cfg.DatabaseURI, "Database connection address")
	flag.StringVar(&cfg.AccrualSystemAddress, "r",
		cfg.AccrualSystemAddress, "Address of the accrual system")
	flag.Parse()

	db, err := dao.NewDAO(cfg.DatabaseURI)
	if err != nil {
		log.Fatal("Can't start application:", err)
	}

	cl := client.NewAccrualProcessor(db, cfg.AccrualSystemAddress, types.WorkersPoolSize)
	cl.Run()

	svc, err := service.NewService(db)
	if err != nil {
		log.Fatal("Can't start application:", err)
	}
	urlApp := app.NewApp(cfg, svc)

	log.Fatal(urlApp.Run())
}
