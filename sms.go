package main

import (
	"log"

	"github.com/houaq/sms/api"
	"github.com/houaq/sms/config"
	"github.com/houaq/sms/database"
	"github.com/houaq/sms/modem"
	"github.com/houaq/sms/worker"
)

func main() {
	cfg, err := config.New("config.toml")
	if err != nil {
		log.Fatalf("main: Invalid config: %s", err.Error())
	}

	db, err := database.InitDB("db.sqlite")
	defer db.Close()
	if err != nil {
		log.Fatalf("main: Error initializing database: %s", err.Error())
	}

	err = modem.InitModem(cfg.ComPort, cfg.BaudRate)
	if err != nil {
		log.Fatalf("main: error initializing to modem. %s", err)
	}
	err = modem.Reset()
	if err != nil {
		log.Fatalf("main: error reseting modem. %s", err)
	}
	worker.InitWorker()
	err = api.InitServer(cfg.ServerHost, cfg.ServerPort)
	if err != nil {
		log.Fatalf("main: Error starting server: %s", err.Error())
	}
}
