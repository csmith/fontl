package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/csmith/envflag/v2"
	"github.com/csmith/fontl"
)

var (
	port      = flag.Int("port", 8080, "Port to listen on")
	directory = flag.String("dir", ".", "Directory containing fonts")
)

func main() {
	envflag.Parse()

	storage := fontl.NewStorage(*directory)
	if err := storage.Load(); err != nil {
		log.Fatalf("Failed to load fonts: %v", err)
	}

	server := fontl.NewServer(storage, *port)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("Received shutdown signal, gracefully shutting down...")
		cancel()
	}()

	go func() {
		log.Printf("Starting server on port %d, serving fonts from %s", *port, *directory)
		if err := server.Start(); err != nil {
			log.Printf("Server error: %v", err)
			cancel()
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}
	log.Println("Server stopped")
}
