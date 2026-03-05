package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/attilagyorffy/prometheus-exporter-adguard-home/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	port := envOrDefault("LISTEN_PORT", "9617")
	adguardURL := envOrDefault("ADGUARD_URL", "https://adguard.gyorffy.network")
	adguardUsername := os.Getenv("ADGUARD_USERNAME")
	adguardPassword := os.Getenv("ADGUARD_PASSWORD")
	topN := envOrDefaultInt("ADGUARD_TOP_N", 10)
	logLevel := envOrDefault("LOG_LEVEL", "info")

	setupLogging(logLevel)

	if adguardUsername == "" || adguardPassword == "" {
		slog.Error("ADGUARD_USERNAME and ADGUARD_PASSWORD environment variables are required")
		os.Exit(1)
	}

	c := collector.New(adguardURL, adguardUsername, adguardPassword, topN)
	prometheus.MustRegister(c)

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "AdGuard Home Exporter\n\nVisit /metrics for Prometheus metrics.\n")
	})

	addr := ":" + port
	slog.Info("starting exporter", "address", addr, "adguard_url", adguardURL, "top_n", topN)
	if err := http.ListenAndServe(addr, nil); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func envOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func envOrDefaultInt(key string, defaultVal int) int {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		slog.Warn("invalid integer for env var, using default", "key", key, "value", v, "default", defaultVal)
		return defaultVal
	}
	return n
}

func setupLogging(level string) {
	var l slog.Level
	switch level {
	case "debug":
		l = slog.LevelDebug
	case "warn":
		l = slog.LevelWarn
	case "error":
		l = slog.LevelError
	default:
		l = slog.LevelInfo
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: l})))
}
