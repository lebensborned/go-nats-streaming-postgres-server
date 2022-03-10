package main

import (
	"log"
	"os"
	"os/signal"
	"runtime/debug"

	"github.com/lebensborned/go-nats-streaming-postgres-server/config"
	"github.com/lebensborned/go-nats-streaming-postgres-server/internal/db"
	"github.com/lebensborned/go-nats-streaming-postgres-server/internal/server"
	"github.com/lebensborned/go-nats-streaming-postgres-server/internal/streaming"
)

func main() {
	config.SetupEnv()
	// Dont let the service drops if panic occured and get stack info
	defer func() {
		if err := recover(); err != nil {
			debug.PrintStack()
		}
	}()

	// Create a repository for app
	repo := db.NewRepo()
	err := repo.Init(os.Getenv("DB_USERNAME"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"))
	if err != nil {
		log.Fatalf("cant init repo: %v", err)
	}
	// Starting a http server
	s := server.NewServer(repo)
	if err := s.Start(); err != nil {
		log.Fatalf("http server cant start: %v", err)
	}
	// Connects to nats-streaming server
	streamHandler, err := streaming.InitStreaming(os.Getenv("NATS_CLIENT_ID"), os.Getenv("NATS_CLUSTER_ID"), os.Getenv("NATS_URL"))
	if err != nil {
		log.Fatalf("init streaming failed: %v", err)
	}
	// Subscribe for nats-streaming channel
	err = streamHandler.Start(os.Getenv("NATS_SUBJECT"), os.Getenv("NATS_SUBJECT"), os.Getenv("NATS_DURABLE_NAME"), repo)
	if err != nil {
		log.Fatalf("cant start subscription: %v", err)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	<-signalChan
	log.Println("Received an interrupt signal (CTRL+C), finishing...")
	err = streamHandler.Finish()
	if err != nil {
		log.Fatal(err)
	}
	err = s.Close()
	if err != nil {
		log.Fatal(err)
	}
}
