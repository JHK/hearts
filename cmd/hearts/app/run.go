package app

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/JHK/hearts/internal/webui"
)

func Run() {
	defaultPort := 8080
	if s := os.Getenv("BEANS_WORKSPACE_PORT"); s != "" {
		defaultPort, _ = strconv.Atoi(s)
	}
	addr := flag.String("addr", fmt.Sprintf(":%d", defaultPort), "web listen address")
	logLevel := flag.String("log-level", "", "log level (debug, info, warn, error); overrides LOG_LEVEL env var")
	dev := flag.Bool("dev", false, "enable dev mode (exposes bot-hand debug endpoint and debugBot() console helper)")
	flag.Parse()

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: resolveLogLevel(*logLevel)})))

	if *dev {
		slog.Warn("dev mode enabled — bot hands are exposed via /api/debug/bot-hands")
	}

	slog.Info("Hearts web server starting", "addr", *addr)
	if err := webui.Run(webui.Config{Addr: *addr, Dev: *dev}); err != nil {
		slog.Error("web server failed", "err", err)
		os.Exit(1)
	}
}

func resolveLogLevel(flagValue string) slog.Level {
	s := strings.ToLower(strings.TrimSpace(flagValue))
	if s == "" {
		s = strings.ToLower(strings.TrimSpace(os.Getenv("LOG_LEVEL")))
	}
	switch s {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
