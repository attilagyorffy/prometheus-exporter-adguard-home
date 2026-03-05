# prometheus-exporter-adguard-home

Custom Prometheus exporter for [AdGuard Home](https://adguard.com/en/adguard-home/overview.html).

## Why a custom exporter?

Several community AdGuard Home exporters exist — most notably [henrywhitaker3/adguard-exporter](https://github.com/henrywhitaker3/adguard-exporter) and [ebrianne/adguard-exporter](https://github.com/ebrianne/adguard-exporter). We ran henrywhitaker3's exporter for months before replacing it. It works, but it has design choices that conflict with Prometheus best practices in ways that cause real operational problems.

### Timer-based polling instead of pull-on-scrape

Both henrywhitaker3 and ebrianne poll AdGuard Home on an internal timer (default 30s and 10s respectively) and cache the results. Prometheus then scrapes stale cached data rather than fresh metrics. This violates the [Prometheus exporter guidelines](https://prometheus.io/docs/instrumenting/writing_exporters/#scheduling), which state that an exporter should fetch data synchronously on each scrape — not on its own timer. The practical consequence: metrics drift out of sync with scrape timestamps, and you can't control freshness via Prometheus's `scrape_interval`.

This exporter fetches from every AdGuard Home API endpoint synchronously during each Prometheus scrape. No timers, no cache, no stale data.

### Incorrect metric units

The henrywhitaker3 exporter exposes `adguard_avg_processing_time` without a unit suffix (the value is in seconds but the name doesn't say so). Prometheus naming conventions require a `_seconds` suffix for time-based metrics so that tools like Grafana can apply the correct unit automatically. This exporter uses `adguard_avg_processing_seconds` and `adguard_top_upstreams_avg_time_seconds`.

### Missing `_up` meta-metric

A well-behaved exporter should report its own health via an `_up` gauge (1 = scrape succeeded, 0 = target unreachable). The henrywhitaker3 exporter has no such metric — if AdGuard Home goes down, the exporter simply stops updating its cached gauges and Prometheus has no way to distinguish "target is down" from "metrics are unchanged". This exporter emits `adguard_up` and `adguard_scrape_duration_seconds` on every scrape.

### Partial failure tolerance

This exporter queries 8 AdGuard Home API endpoints per scrape. Only `/control/status` is mandatory — if any other endpoint fails, the exporter still returns the metrics it could collect and sets `adguard_up` to 1. A temporary failure in one endpoint doesn't zero out the entire scrape.

### Client name resolution

The `adguard_top_clients` metric includes both the `client` (IP address) and `name` labels. The name is resolved from AdGuard Home's persistent clients list (`/control/clients`). Clients without a persistent client entry fall back to the IP address, so the label is always populated and Grafana legend entries are always readable.

## Configuration

| Variable | Default | Required | Description |
|----------|---------|----------|-------------|
| `ADGUARD_URL` | `https://adguard.gyorffy.network` | no | AdGuard Home base URL |
| `ADGUARD_USERNAME` | — | yes | Basic auth username |
| `ADGUARD_PASSWORD` | — | yes | Basic auth password |
| `LISTEN_PORT` | `9617` | no | Exporter listen port |
| `ADGUARD_TOP_N` | `10` | no | Number of top domains/clients/upstreams |
| `LOG_LEVEL` | `info` | no | Log level (debug, info, warn, error) |

## Metrics

### Scrape meta

| Metric | Type | Description |
|--------|------|-------------|
| `adguard_up` | gauge | Whether the last scrape succeeded (1 = up, 0 = down) |
| `adguard_scrape_duration_seconds` | gauge | Duration of the last scrape |

### Instance info (`/control/status`)

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `adguard_build_info` | gauge | `version`, `dns_port`, `http_port` | Build information (always 1) |
| `adguard_running` | gauge | — | Whether AdGuard Home is running |
| `adguard_protection_enabled` | gauge | — | Whether DNS protection is enabled |

### Query statistics (`/control/stats`)

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `adguard_dns_queries` | gauge | — | Total DNS queries (24h rolling window) |
| `adguard_blocked_filtering` | gauge | — | Queries blocked by filters |
| `adguard_replaced_safebrowsing` | gauge | — | Queries replaced by safe browsing |
| `adguard_replaced_safesearch` | gauge | — | Queries replaced by safe search |
| `adguard_replaced_parental` | gauge | — | Queries replaced by parental control |
| `adguard_avg_processing_seconds` | gauge | — | Average query processing time |
| `adguard_top_queried_domains` | gauge | `domain` | Top queried domains by count |
| `adguard_top_blocked_domains` | gauge | `domain` | Top blocked domains by count |
| `adguard_top_clients` | gauge | `client`, `name` | Top clients by query count |
| `adguard_top_upstreams_responses` | gauge | `upstream` | Top upstreams by response count |
| `adguard_top_upstreams_avg_time_seconds` | gauge | `upstream` | Top upstreams by average response time |

### DNS config (`/control/dns_info`)

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `adguard_dns_config_info` | gauge | `upstream_mode`, `blocking_mode` | DNS configuration (always 1) |
| `adguard_dns_ratelimit` | gauge | — | DNS rate limit |
| `adguard_dns_cache_size_bytes` | gauge | — | DNS cache size in bytes |
| `adguard_dns_cache_enabled` | gauge | — | Whether DNS cache is enabled |
| `adguard_dns_cache_optimistic` | gauge | — | Whether optimistic caching is enabled |
| `adguard_dnssec_enabled` | gauge | — | Whether DNSSEC is enabled |

### Filtering (`/control/filtering/status`)

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `adguard_filtering_enabled` | gauge | — | Whether filtering is enabled |
| `adguard_filtering_update_interval_hours` | gauge | — | Filter update interval |
| `adguard_filtering_rules_total` | gauge | — | Total filter rules across all lists |
| `adguard_filtering_lists_total` | gauge | — | Total number of filter lists |
| `adguard_filtering_lists_enabled` | gauge | — | Number of enabled filter lists |
| `adguard_filtering_user_rules_total` | gauge | — | Number of user-defined rules |
| `adguard_filtering_list_rules` | gauge | `name`, `enabled` | Rules per filter list |
| `adguard_filtering_list_last_updated_timestamp_seconds` | gauge | `name` | Last update time per filter list |

### Protection (3 endpoints)

| Metric | Type | Description |
|--------|------|-------------|
| `adguard_safebrowsing_enabled` | gauge | Whether safe browsing is enabled |
| `adguard_safesearch_enabled` | gauge | Whether safe search is enabled |
| `adguard_parental_enabled` | gauge | Whether parental control is enabled |

## Running locally

```sh
make run    # Starts exporter (fetches creds from 1Password)
make query  # Curls the /metrics endpoint
make test   # Runs all tests
```

## Deployment

Built and pushed to `ghcr.io/attilagyorffy/prometheus-exporter-adguard-home` via GitHub Actions. Deployed to TrueNAS via Gitea CD pipeline (mirror sync triggers redeploy).
