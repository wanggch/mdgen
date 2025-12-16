package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/example/mdgen/internal/config"
	"github.com/example/mdgen/internal/logger"
	"github.com/example/mdgen/internal/server"
)

func main() {
	cfgPath := flag.String("config", "", "path to config file")
	listen := flag.String("listen", "", "listen address (overrides config)")
	output := flag.String("output", "", "output directory (overrides config)")
	apiKey := flag.String("api-key", "", "api key for requests")
	logLevel := flag.String("log-level", "", "log level")
	maxBody := flag.Int64("max-body", 0, "max body size")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		panic(err)
	}

	if *listen != "" {
		cfg.ListenAddr = *listen
	}
	if *output != "" {
		cfg.OutputDir = *output
	}
	if *apiKey != "" {
		cfg.APIKey = *apiKey
	}
	if *logLevel != "" {
		cfg.LogLevel = *logLevel
	}
	if *maxBody > 0 {
		cfg.MaxBodySize = *maxBody
	}

	log := logger.New(cfg.LogLevel)

	srv := server.New(cfg, log)

	log.Info("starting server", "listen", cfg.ListenAddr, "output_dir", cfg.OutputDir)
	if err := http.ListenAndServe(cfg.ListenAddr, srv.Handler()); err != nil {
		log.Error("server stopped", slog.Any("error", err))
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}
