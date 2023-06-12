package main

import (
	"github.com/caarlos0/env/v8"
	"log"
	"tcpChat/internal/server"
)

func main() {
	cfg := server.Config{}

	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("failed to retrieve environment variables, %s\n", err)
	}
	s := server.New()
	if err := s.Run(cfg); err != nil {
		log.Fatal(err)
	}
}
