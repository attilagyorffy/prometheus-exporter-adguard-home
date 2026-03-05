package collector

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	dnsConfigInfoDesc = prometheus.NewDesc(
		"adguard_dns_config_info",
		"DNS configuration metadata (from /control/dns_info).",
		[]string{"upstream_mode", "blocking_mode"}, nil,
	)
	dnsRatelimitDesc = prometheus.NewDesc(
		"adguard_dns_ratelimit",
		"DNS query rate limit per client per second (from /control/dns_info ratelimit).",
		nil, nil,
	)
	dnsCacheSizeDesc = prometheus.NewDesc(
		"adguard_dns_cache_size_bytes",
		"Configured DNS cache size in bytes (from /control/dns_info cache_size).",
		nil, nil,
	)
	dnsCacheEnabledDesc = prometheus.NewDesc(
		"adguard_dns_cache_enabled",
		"Whether DNS caching is enabled (from /control/dns_info cache_enabled).",
		nil, nil,
	)
	dnsCacheOptimisticDesc = prometheus.NewDesc(
		"adguard_dns_cache_optimistic",
		"Whether optimistic DNS caching is enabled (from /control/dns_info cache_optimistic).",
		nil, nil,
	)
	dnssecEnabledDesc = prometheus.NewDesc(
		"adguard_dnssec_enabled",
		"Whether DNSSEC validation is enabled (from /control/dns_info dnssec_enabled).",
		nil, nil,
	)
)

func describeDNSInfo(ch chan<- *prometheus.Desc) {
	ch <- dnsConfigInfoDesc
	ch <- dnsRatelimitDesc
	ch <- dnsCacheSizeDesc
	ch <- dnsCacheEnabledDesc
	ch <- dnsCacheOptimisticDesc
	ch <- dnssecEnabledDesc
}

func collectDNSInfo(d *DNSInfoResponse, ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(dnsConfigInfoDesc, prometheus.GaugeValue, 1, d.UpstreamMode, d.BlockingMode)
	ch <- prometheus.MustNewConstMetric(dnsRatelimitDesc, prometheus.GaugeValue, d.Ratelimit)
	ch <- prometheus.MustNewConstMetric(dnsCacheSizeDesc, prometheus.GaugeValue, d.CacheSize)
	ch <- prometheus.MustNewConstMetric(dnsCacheEnabledDesc, prometheus.GaugeValue, boolToFloat(d.CacheEnabled))
	ch <- prometheus.MustNewConstMetric(dnsCacheOptimisticDesc, prometheus.GaugeValue, boolToFloat(d.CacheOptimistic))
	ch <- prometheus.MustNewConstMetric(dnssecEnabledDesc, prometheus.GaugeValue, boolToFloat(d.DNSSECEnabled))
}
