package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jkeddari/jarvis/api"
	"github.com/jkeddari/jarvis/client"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
)

func main() {
	fmt.Println("Starting Chains Tracker...")
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	url := os.Getenv("RPC_URL")
	if url == "" {
		log.Fatal("no url found")
	}
	streamURL := os.Getenv("STREAM_URL")
	if url == "" {
		log.Fatal("no url found")
	}
	ethDBPath := os.Getenv("DB_ETH_PATH")
	if ethDBPath == "" {
		log.Fatal("no ethereum db path found")
	}

	logger := zerolog.New(
		zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC1123Z},
	).Level(zerolog.TraceLevel).With().Timestamp().Logger()

	conf := api.Config{
		ETHConfig: client.Config{
			DBPath:           ethDBPath,
			URL:              url,
			StreamURL:        streamURL,
			ConcurrentNumber: 1000,
			Logger:           &logger,
		},
	}

	app, err := api.NewServer(conf)
	app.Logger = logger
	if err != nil {
		log.Fatal(err)
	}
	app.Listen("localhost:8080")
}
