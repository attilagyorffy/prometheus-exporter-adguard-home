package collector

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	filteringEnabledDesc = prometheus.NewDesc(
		"adguard_filtering_enabled",
		"Whether DNS filtering is globally enabled (from /control/filtering/status enabled).",
		nil, nil,
	)
	filteringUpdateIntervalDesc = prometheus.NewDesc(
		"adguard_filtering_update_interval_hours",
		"Filter list auto-update interval in hours (from /control/filtering/status interval).",
		nil, nil,
	)
	filteringRulesTotalDesc = prometheus.NewDesc(
		"adguard_filtering_rules_total",
		"Total number of rules across all enabled filter lists (computed from /control/filtering/status filters).",
		nil, nil,
	)
	filteringListsTotalDesc = prometheus.NewDesc(
		"adguard_filtering_lists_total",
		"Total number of configured filter lists (from /control/filtering/status filters).",
		nil, nil,
	)
	filteringListsEnabledDesc = prometheus.NewDesc(
		"adguard_filtering_lists_enabled",
		"Number of enabled filter lists (from /control/filtering/status filters).",
		nil, nil,
	)
	filteringUserRulesTotalDesc = prometheus.NewDesc(
		"adguard_filtering_user_rules_total",
		"Number of user-defined filtering rules (from /control/filtering/status user_rules).",
		nil, nil,
	)
	filteringListRulesDesc = prometheus.NewDesc(
		"adguard_filtering_list_rules",
		"Number of rules in a filter list (from /control/filtering/status filters[].rules_count).",
		[]string{"name", "enabled"}, nil,
	)
	filteringListLastUpdatedDesc = prometheus.NewDesc(
		"adguard_filtering_list_last_updated_timestamp_seconds",
		"Unix timestamp of last filter list update (from /control/filtering/status filters[].last_updated).",
		[]string{"name"}, nil,
	)
)

func describeFiltering(ch chan<- *prometheus.Desc) {
	ch <- filteringEnabledDesc
	ch <- filteringUpdateIntervalDesc
	ch <- filteringRulesTotalDesc
	ch <- filteringListsTotalDesc
	ch <- filteringListsEnabledDesc
	ch <- filteringUserRulesTotalDesc
	ch <- filteringListRulesDesc
	ch <- filteringListLastUpdatedDesc
}

func collectFiltering(f *FilteringStatusResponse, ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(filteringEnabledDesc, prometheus.GaugeValue, boolToFloat(f.Enabled))
	ch <- prometheus.MustNewConstMetric(filteringUpdateIntervalDesc, prometheus.GaugeValue, f.Interval)

	var totalRules float64
	var enabledLists float64
	for _, filter := range f.Filters {
		if filter.Enabled {
			totalRules += filter.RulesCount
			enabledLists++
		}

		ch <- prometheus.MustNewConstMetric(
			filteringListRulesDesc, prometheus.GaugeValue, filter.RulesCount,
			filter.Name, fmt.Sprintf("%t", filter.Enabled),
		)

		if t, err := time.Parse(time.RFC3339, filter.LastUpdated); err == nil {
			ch <- prometheus.MustNewConstMetric(
				filteringListLastUpdatedDesc, prometheus.GaugeValue, float64(t.Unix()),
				filter.Name,
			)
		} else {
			slog.Warn("failed to parse filter last_updated", "filter", filter.Name, "value", filter.LastUpdated, "error", err)
		}
	}

	ch <- prometheus.MustNewConstMetric(filteringRulesTotalDesc, prometheus.GaugeValue, totalRules)
	ch <- prometheus.MustNewConstMetric(filteringListsTotalDesc, prometheus.GaugeValue, float64(len(f.Filters)))
	ch <- prometheus.MustNewConstMetric(filteringListsEnabledDesc, prometheus.GaugeValue, enabledLists)
	ch <- prometheus.MustNewConstMetric(filteringUserRulesTotalDesc, prometheus.GaugeValue, float64(len(f.UserRules)))
}
