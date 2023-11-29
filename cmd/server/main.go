package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/jkeddari/jarvis/api"
	"github.com/jkeddari/jarvis/client"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
)

func main() {
	fmt.Println("Starting Chains Tracker...")
	godotenv.Load()

	url := os.Getenv("RPC_URL")
	if url == "" {
		log.Fatal("no url found")
	}

	ethDBPath := os.Getenv("DB_ETH_PATH")
	if ethDBPath == "" {
		log.Fatal("no ethereum db path found")
	}

	maxConnEnv := os.Getenv("MAX_CONN")
	maxConn, err := strconv.Atoi(maxConnEnv)
	if err != nil {
		log.Println(err)
	}

	logger := zerolog.New(
		zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC1123Z},
	).Level(zerolog.TraceLevel).With().Timestamp().Logger()

	conf := api.Config{
		ETHConfig: client.Config{
			DBPath:           ethDBPath,
			URL:              url,
			ConcurrentNumber: maxConn,
			Logger:           &logger,
		},
	}

	app, err := api.NewServer(conf)
	if err != nil {
		log.Fatal(err)
	}

	app.Logger = logger
	app.Listen("localhost:8080")
}
