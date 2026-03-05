package collector

import (
	"log/slog"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	upDesc = prometheus.NewDesc(
		"adguard_up",
		"Whether the AdGuard Home instance is reachable (from /control/status).",
		nil, nil,
	)
	scrapeDurationDesc = prometheus.NewDesc(
		"adguard_scrape_duration_seconds",
		"Time taken to scrape AdGuard Home API.",
		nil, nil,
	)
)

// Collector implements prometheus.Collector for AdGuard Home metrics.
type Collector struct {
	client *Client
	topN   int
}

// New creates a new Collector.
func New(adguardURL, username, password string, topN int) *Collector {
	return &Collector{
		client: NewClient(adguardURL, username, password),
		topN:   topN,
	}
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- upDesc
	ch <- scrapeDurationDesc
	describeStatus(ch)
	describeStats(ch)
	describeDNSInfo(ch)
	describeFiltering(ch)
	describeProtection(ch)
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	start := time.Now()

	// /control/status — determines adguard_up
	status, err := c.client.FetchStatus()
	if err != nil {
		slog.Error("failed to fetch status", "error", err)
		ch <- prometheus.MustNewConstMetric(upDesc, prometheus.GaugeValue, 0)
		ch <- prometheus.MustNewConstMetric(scrapeDurationDesc, prometheus.GaugeValue, time.Since(start).Seconds())
		return
	}
	ch <- prometheus.MustNewConstMetric(upDesc, prometheus.GaugeValue, 1)
	collectStatus(status, ch)

	// /control/clients — build IP-to-name map for top_clients enrichment
	clientMap := c.client.BuildClientMap()

	// /control/stats
	stats, err := c.client.FetchStats()
	if err != nil {
		slog.Error("failed to fetch stats", "error", err)
	} else {
		collectStats(stats, clientMap, c.topN, ch)
	}

	// /control/dns_info
	dnsInfo, err := c.client.FetchDNSInfo()
	if err != nil {
		slog.Error("failed to fetch dns info", "error", err)
	} else {
		collectDNSInfo(dnsInfo, ch)
	}

	// /control/filtering/status
	filtering, err := c.client.FetchFilteringStatus()
	if err != nil {
		slog.Error("failed to fetch filtering status", "error", err)
	} else {
		collectFiltering(filtering, ch)
	}

	// safebrowsing + safesearch + parental
	collectProtection(c.client, ch)

	ch <- prometheus.MustNewConstMetric(scrapeDurationDesc, prometheus.GaugeValue, time.Since(start).Seconds())
}
