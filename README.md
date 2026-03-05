# prometheus-exporter-adguard-home

Prometheus exporter for [AdGuard Home](https://adguard.com/en/adguard-home/overview.html), built following [official Prometheus exporter guidelines](https://prometheus.io/docs/instrumenting/writing_exporters/).

**Design principles:**

- **Synchronous pull-on-scrape** — fetches from all AdGuard Home API endpoints during each Prometheus scrape. No internal polling timers, no cache, no stale data.
- **Correct metric naming** — time-based metrics use `_seconds` suffixes (`adguard_avg_processing_seconds`, not `adguard_avg_processing_time`), byte-based metrics use `_bytes`, so Grafana applies units automatically.
- **`adguard_up` meta-metric** — reports scrape health (1 = up, 0 = down) so Prometheus can distinguish "target is down" from "metrics are unchanged".
- **Partial failure tolerance** — queries 8 API endpoints per scrape; only `/control/status` is fatal. A temporary failure in one endpoint doesn't zero out the entire scrape.
- **Client name resolution** — `adguard_top_clients` includes a `name` label resolved from AdGuard Home's persistent clients list, falling back to the IP address for unnamed clients.
- **Low cardinality** — top-N metrics are bounded by `ADGUARD_TOP_N` (default 10). No unbounded label dimensions.

## Configuration

| Variable | Default | Required | Description |
|----------|---------|----------|-------------|
| `ADGUARD_URL` | `http://localhost:3000` | no | AdGuard Home base URL |
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

Built and pushed to `ghcr.io/attilagyorffy/prometheus-exporter-adguard-home` via GitHub Actions on every push to `main`.

## Further reading

- [Comparison with existing exporters](docs/comparison.md) — why this exporter was built instead of using henrywhitaker3 or ebrianne
