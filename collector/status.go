package collector

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	buildInfoDesc = prometheus.NewDesc(
		"adguard_build_info",
		"AdGuard Home version and port information (from /control/status).",
		[]string{"version", "dns_port", "http_port"}, nil,
	)
	runningDesc = prometheus.NewDesc(
		"adguard_running",
		"Whether AdGuard Home is running (from /control/status running).",
		nil, nil,
	)
	protectionEnabledDesc = prometheus.NewDesc(
		"adguard_protection_enabled",
		"Whether DNS filtering protection is enabled (from /control/status protection_enabled).",
		nil, nil,
	)
)

func describeStatus(ch chan<- *prometheus.Desc) {
	ch <- buildInfoDesc
	ch <- runningDesc
	ch <- protectionEnabledDesc
}

func collectStatus(s *StatusResponse, ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(
		buildInfoDesc, prometheus.GaugeValue, 1,
		s.Version,
		fmt.Sprintf("%d", s.DNSPort),
		fmt.Sprintf("%d", s.HTTPPort),
	)
	ch <- prometheus.MustNewConstMetric(runningDesc, prometheus.GaugeValue, boolToFloat(s.Running))
	ch <- prometheus.MustNewConstMetric(protectionEnabledDesc, prometheus.GaugeValue, boolToFloat(s.ProtectionEnabled))
}
