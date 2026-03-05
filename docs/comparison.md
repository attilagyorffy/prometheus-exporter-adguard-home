# Comparison with existing exporters

Several community AdGuard Home exporters exist — most notably [henrywhitaker3/adguard-exporter](https://github.com/henrywhitaker3/adguard-exporter) and [ebrianne/adguard-exporter](https://github.com/ebrianne/adguard-exporter). We ran henrywhitaker3's exporter for months before replacing it. It works, but it has design choices that conflict with [Prometheus exporter guidelines](https://prometheus.io/docs/instrumenting/writing_exporters/) in ways that cause real operational problems.

## Timer-based polling instead of pull-on-scrape

Both henrywhitaker3 and ebrianne poll AdGuard Home on an internal timer (default 30s and 10s respectively) and cache the results. Prometheus then scrapes stale cached data rather than fresh metrics. The exporter guidelines [explicitly state](https://prometheus.io/docs/instrumenting/writing_exporters/#scheduling) that an exporter should fetch data synchronously on each scrape — not on its own timer. The practical consequence: metrics drift out of sync with scrape timestamps, and you can't control freshness via Prometheus's `scrape_interval`.

This exporter fetches from every AdGuard Home API endpoint synchronously during each Prometheus scrape. No timers, no cache, no stale data.

## Incorrect metric units

The henrywhitaker3 exporter exposes `adguard_avg_processing_time` without a unit suffix (the value is in seconds but the name doesn't say so). Prometheus naming conventions require a `_seconds` suffix for time-based metrics so that tools like Grafana can apply the correct unit automatically. This exporter uses `adguard_avg_processing_seconds` and `adguard_top_upstreams_avg_time_seconds`.

## Missing `_up` meta-metric

A well-behaved exporter should report its own health via an `_up` gauge (1 = scrape succeeded, 0 = target unreachable). The henrywhitaker3 exporter has no such metric — if AdGuard Home goes down, the exporter simply stops updating its cached gauges and Prometheus has no way to distinguish "target is down" from "metrics are unchanged". This exporter emits `adguard_up` and `adguard_scrape_duration_seconds` on every scrape.
