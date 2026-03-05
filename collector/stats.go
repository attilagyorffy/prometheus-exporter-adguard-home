package collector

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	dnsQueriesDesc = prometheus.NewDesc(
		"adguard_dns_queries",
		"Total DNS queries in the configured stats window (from /control/stats num_dns_queries). Rolling window total, not a monotonic counter.",
		nil, nil,
	)
	blockedFilteringDesc = prometheus.NewDesc(
		"adguard_blocked_filtering",
		"DNS queries blocked by filtering rules in the stats window (from /control/stats num_blocked_filtering).",
		nil, nil,
	)
	replacedSafebrowsingDesc = prometheus.NewDesc(
		"adguard_replaced_safebrowsing",
		"DNS queries replaced by safe browsing in the stats window (from /control/stats num_replaced_safebrowsing).",
		nil, nil,
	)
	replacedSafesearchDesc = prometheus.NewDesc(
		"adguard_replaced_safesearch",
		"DNS queries replaced by safe search in the stats window (from /control/stats num_replaced_safesearch).",
		nil, nil,
	)
	replacedParentalDesc = prometheus.NewDesc(
		"adguard_replaced_parental",
		"DNS queries replaced by parental control in the stats window (from /control/stats num_replaced_parental).",
		nil, nil,
	)
	avgProcessingDesc = prometheus.NewDesc(
		"adguard_avg_processing_seconds",
		"Server-computed average DNS processing time in seconds for the stats window (from /control/stats avg_processing_time). Pre-computed average; raw observations not available from API.",
		nil, nil,
	)
	topQueriedDomainsDesc = prometheus.NewDesc(
		"adguard_top_queried_domains",
		"Query count for a top queried domain in the stats window (from /control/stats top_queried_domains).",
		[]string{"domain"}, nil,
	)
	topBlockedDomainsDesc = prometheus.NewDesc(
		"adguard_top_blocked_domains",
		"Block count for a top blocked domain in the stats window (from /control/stats top_blocked_domains).",
		[]string{"domain"}, nil,
	)
	topClientsDesc = prometheus.NewDesc(
		"adguard_top_clients",
		"Query count for a top client in the stats window (from /control/stats top_clients). The name label is resolved from persistent clients.",
		[]string{"client", "name"}, nil,
	)
	topUpstreamsResponsesDesc = prometheus.NewDesc(
		"adguard_top_upstreams_responses",
		"Response count for a DNS upstream in the stats window (from /control/stats top_upstreams_responses).",
		[]string{"upstream"}, nil,
	)
	topUpstreamsAvgTimeDesc = prometheus.NewDesc(
		"adguard_top_upstreams_avg_time_seconds",
		"Server-computed average response time in seconds for a DNS upstream (from /control/stats top_upstreams_avg_time). Pre-computed average.",
		[]string{"upstream"}, nil,
	)
)

func describeStats(ch chan<- *prometheus.Desc) {
	ch <- dnsQueriesDesc
	ch <- blockedFilteringDesc
	ch <- replacedSafebrowsingDesc
	ch <- replacedSafesearchDesc
	ch <- replacedParentalDesc
	ch <- avgProcessingDesc
	ch <- topQueriedDomainsDesc
	ch <- topBlockedDomainsDesc
	ch <- topClientsDesc
	ch <- topUpstreamsResponsesDesc
	ch <- topUpstreamsAvgTimeDesc
}

func collectStats(s *StatsResponse, clientMap map[string]string, topN int, ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(dnsQueriesDesc, prometheus.GaugeValue, s.NumDNSQueries)
	ch <- prometheus.MustNewConstMetric(blockedFilteringDesc, prometheus.GaugeValue, s.NumBlockedFiltering)
	ch <- prometheus.MustNewConstMetric(replacedSafebrowsingDesc, prometheus.GaugeValue, s.NumReplacedSafebrowsing)
	ch <- prometheus.MustNewConstMetric(replacedSafesearchDesc, prometheus.GaugeValue, s.NumReplacedSafesearch)
	ch <- prometheus.MustNewConstMetric(replacedParentalDesc, prometheus.GaugeValue, s.NumReplacedParental)
	ch <- prometheus.MustNewConstMetric(avgProcessingDesc, prometheus.GaugeValue, s.AvgProcessingTime)

	collectTopEntries(s.TopQueriedDomains, topQueriedDomainsDesc, topN, ch)
	collectTopEntries(s.TopBlockedDomains, topBlockedDomainsDesc, topN, ch)
	collectTopClients(s.TopClients, clientMap, topN, ch)
	collectTopEntries(s.TopUpstreamsResponses, topUpstreamsResponsesDesc, topN, ch)
	collectTopEntries(s.TopUpstreamsAvgTime, topUpstreamsAvgTimeDesc, topN, ch)
}

func collectTopEntries(entries []map[string]float64, desc *prometheus.Desc, topN int, ch chan<- prometheus.Metric) {
	for i, entry := range entries {
		if i >= topN {
			break
		}
		for key, value := range entry {
			ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, value, key)
		}
	}
}

func collectTopClients(entries []map[string]float64, clientMap map[string]string, topN int, ch chan<- prometheus.Metric) {
	for i, entry := range entries {
		if i >= topN {
			break
		}
		for ip, value := range entry {
			name := ip
			if clientMap != nil {
				if n, ok := clientMap[ip]; ok {
					name = n
				}
			}
			ch <- prometheus.MustNewConstMetric(topClientsDesc, prometheus.GaugeValue, value, ip, name)
		}
	}
}
