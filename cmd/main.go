package main

import (
	"fmt"
	"log"

	"github.com/selmison/seed-desafio-mercado-livre/mercadolivre"
)

func main() {
	logger := mercadolivre.NewLogger(mercadolivre.DebugLevel)
	svc, err := mercadolivre.NewService(
		"postgres",
		"host=localhost port=5433 dbname=mercadolivre user=postgres password=postgres sslmode=disable",
		logger,
	)
	if err != nil {
		log.Fatalf("failed to initialize service: %v\n", err)
	}
	if err := mercadolivre.NewHTTPServer(svc, logger, &mercadolivre.Config{
		Host: "localhost",
		Port: 3333,
	}); err != nil {
		log.Fatal(fmt.Sprintf("Error starting http server: %s", err))
	}
}
