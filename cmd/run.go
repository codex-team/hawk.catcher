package cmd

import (
	"log"
	"time"

	"github.com/caarlos0/env"
	"github.com/codex-team/hawk.collector/server"
	"github.com/joho/godotenv"
)

// Execute Run server - Load configuration file and start server
func (x *RunCommand) Execute(args []string) error {
	if err := godotenv.Load(); err != nil {
		log.Println("File .env not found, reading configuration from ENV")
	}

	var cfg server.Config
	if err := env.Parse(&cfg); err != nil {
		log.Fatalln("Failed to parse ENV")
	}

	// Override env params with cli arguments
	if x.Host != "" {
		cfg.Host = x.Host
	}
	if x.Port != 0 {
		cfg.Port = x.Port
	}
	if x.BrokerURL != "" {
		cfg.BrokerURL = x.BrokerURL
	}

	s, err := server.New(&cfg)
	if err != nil {
		return err
	}

	// Try to connect to the Queue server several times until success or out of RetryNumber
	retry := cfg.RetryNumber
	err = s.Connect()
	for (err != nil) && (retry > 0) {
		time.Sleep(time.Second * time.Duration(cfg.RetryInterval))
		err = s.Connect()
		retry--
	}
	if err != nil {
		log.Fatalf("Could not connect to the queue server")
	}

	log.Printf("[AMQP] connected on %s", cfg.BrokerURL)

	// Run background workers
	s.RunWorkers()

	// Listen and serve
	s.Serve()

	return nil
}