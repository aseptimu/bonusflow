package main

import (
	"github.com/aseptimu/internal/server"
	"log/slog"
)

func main() {
	if err := server.Serve(); err != nil {
		slog.Error("Error running server", "error", err)
	}
}
