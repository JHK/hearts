package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JHK/hearts/internal/table"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

func main() {
	var (
		host    = flag.String("host", "127.0.0.1", "embedded NATS host")
		port    = flag.Int("port", 4222, "embedded NATS port")
		tableID = flag.String("table", "default", "table id")
	)
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	natsServer, err := server.NewServer(&server.Options{
		ServerName: "hearts-embedded",
		Host:       *host,
		Port:       *port,
		NoSigs:     true,
	})
	if err != nil {
		fatalf("failed to create embedded nats server: %v", err)
	}

	go natsServer.Start()
	if !natsServer.ReadyForConnections(10 * time.Second) {
		fatalf("embedded nats server did not become ready")
	}

	natsURL := natsServer.ClientURL()
	nc, err := nats.Connect(natsURL)
	if err != nil {
		fatalf("failed to connect to embedded nats server: %v", err)
	}
	defer nc.Close()

	svc := table.NewService(*tableID, nc)
	if err := svc.Register(); err != nil {
		fatalf("failed to register table service: %v", err)
	}

	slog.Info("hearts host started", "table", *tableID, "nats", natsURL)

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	<-shutdown

	slog.Info("shutting down")
	if err := nc.Drain(); err != nil {
		slog.Warn("nats drain failed", "error", err)
	}
	natsServer.Shutdown()
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
