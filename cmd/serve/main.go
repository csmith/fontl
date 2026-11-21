package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"syscall"

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

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	log.Printf("Serving fonts from %s", *directory)
	if err := fontl.NewServer(storage, *port).Serve(ctx); err != nil {
		log.Fatalf("Server error: %v", err)
	}
	log.Println("Server stopped")
}
