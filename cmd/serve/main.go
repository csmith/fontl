package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/csmith/envflag/v2"
	"github.com/csmith/fontl"
	"github.com/csmith/slogflags"
)

var (
	port      = flag.Int("port", 8080, "Port to listen on")
	directory = flag.String("dir", ".", "Directory containing fonts")
)

func main() {
	envflag.Parse()
	_ = slogflags.Logger(slogflags.WithSetDefault(true))

	storage := fontl.NewStorage(*directory)
	if err := storage.Load(); err != nil {
		slog.Error("Failed to load fonts", "error", err)
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	slog.Info("Serving fonts", "dir", *directory)
	if err := fontl.NewServer(storage, *port).Serve(ctx); err != nil {
		slog.Error("Server error", "error", err)
		os.Exit(1)
	}
	slog.Info("Server stopped")
}
