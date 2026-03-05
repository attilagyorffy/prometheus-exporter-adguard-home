package collector

import (
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	safebrowsingEnabledDesc = prometheus.NewDesc(
		"adguard_safebrowsing_enabled",
		"Whether safe browsing protection is enabled (from /control/safebrowsing/status).",
		nil, nil,
	)
	safesearchEnabledDesc = prometheus.NewDesc(
		"adguard_safesearch_enabled",
		"Whether safe search enforcement is enabled (from /control/safesearch/status).",
		nil, nil,
	)
	parentalEnabledDesc = prometheus.NewDesc(
		"adguard_parental_enabled",
		"Whether parental control is enabled (from /control/parental/status).",
		nil, nil,
	)
)

func describeProtection(ch chan<- *prometheus.Desc) {
	ch <- safebrowsingEnabledDesc
	ch <- safesearchEnabledDesc
	ch <- parentalEnabledDesc
}

func collectProtection(client *Client, ch chan<- prometheus.Metric) {
	if enabled, err := client.FetchEnabled("/control/safebrowsing/status"); err != nil {
		slog.Error("failed to fetch safebrowsing status", "error", err)
	} else {
		ch <- prometheus.MustNewConstMetric(safebrowsingEnabledDesc, prometheus.GaugeValue, boolToFloat(enabled))
	}

	if enabled, err := client.FetchEnabled("/control/safesearch/status"); err != nil {
		slog.Error("failed to fetch safesearch status", "error", err)
	} else {
		ch <- prometheus.MustNewConstMetric(safesearchEnabledDesc, prometheus.GaugeValue, boolToFloat(enabled))
	}

	if enabled, err := client.FetchEnabled("/control/parental/status"); err != nil {
		slog.Error("failed to fetch parental status", "error", err)
	} else {
		ch <- prometheus.MustNewConstMetric(parentalEnabledDesc, prometheus.GaugeValue, boolToFloat(enabled))
	}
}
