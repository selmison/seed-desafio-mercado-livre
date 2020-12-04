package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/selmison/seed-desafio-mercado-livre/mercadolivre"
)

func main() {
	logger := mercadolivre.NewLogger(mercadolivre.DebugLevel)

	driverName := "postgres"
	db, err := sql.Open(driverName, "host=localhost port=5433 dbname=mercadolivre user=postgres password=postgres sslmode=disable")
	if err != nil {
		log.Fatalf("failed to initialize db: %v\n", err)
	}

	svc, err := mercadolivre.NewService(db, "postgres", logger)
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
