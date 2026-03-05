package collector

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

// newTestServer creates a mock AdGuard Home API that responds to all endpoints.
func newTestServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/control/status":
			json.NewEncoder(w).Encode(StatusResponse{
				Version:           "v0.107.72",
				DNSPort:           53,
				HTTPPort:          3000,
				ProtectionEnabled: true,
				Running:           true,
			})
		case "/control/stats":
			json.NewEncoder(w).Encode(StatsResponse{
				NumDNSQueries:           100000,
				NumBlockedFiltering:     15000,
				NumReplacedSafebrowsing: 0,
				NumReplacedSafesearch:   0,
				NumReplacedParental:     0,
				AvgProcessingTime:       0.025,
				TopQueriedDomains:       []map[string]float64{{"example.com": 500}},
				TopBlockedDomains:       []map[string]float64{{"ads.example.com": 300}},
				TopClients:              []map[string]float64{{"10.0.0.3": 50000}, {"10.0.0.100": 1200}},
				TopUpstreamsResponses:   []map[string]float64{{"127.0.0.1:5300": 90000}},
				TopUpstreamsAvgTime:     []map[string]float64{{"127.0.0.1:5300": 0.05}},
			})
		case "/control/dns_info":
			json.NewEncoder(w).Encode(DNSInfoResponse{
				Ratelimit:       20,
				BlockingMode:    "default",
				DNSSECEnabled:   false,
				CacheSize:       4194304,
				CacheEnabled:    true,
				CacheOptimistic: false,
				UpstreamMode:    "parallel",
			})
		case "/control/filtering/status":
			json.NewEncoder(w).Encode(FilteringStatusResponse{
				Enabled:  true,
				Interval: 24,
				Filters: []Filter{
					{Name: "AdGuard DNS filter", LastUpdated: "2026-03-04T23:34:26+01:00", RulesCount: 155880, Enabled: true},
					{Name: "Disabled filter", LastUpdated: "2026-03-01T00:00:00+01:00", RulesCount: 1000, Enabled: false},
				},
				UserRules: []string{"rule1", "rule2"},
			})
		case "/control/clients":
			json.NewEncoder(w).Encode(ClientsResponse{
				Clients: []PersistentClient{
					{Name: "TrueNAS", IDs: []string{"10.0.0.3"}},
				},
			})
		case "/control/safebrowsing/status":
			json.NewEncoder(w).Encode(EnabledResponse{Enabled: false})
		case "/control/safesearch/status":
			json.NewEncoder(w).Encode(EnabledResponse{Enabled: false})
		case "/control/parental/status":
			json.NewEncoder(w).Encode(EnabledResponse{Enabled: false})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestCollectorMetaMetrics(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c := New(srv.URL, "admin", "secret", 10)

	expected := `
		# HELP adguard_up Whether the AdGuard Home instance is reachable (from /control/status).
		# TYPE adguard_up gauge
		adguard_up 1
	`
	if err := testutil.CollectAndCompare(c, strings.NewReader(expected), "adguard_up"); err != nil {
		t.Error(err)
	}
}

func TestCollectorUpDown(t *testing.T) {
	// Server that always returns 500
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := New(srv.URL, "admin", "secret", 10)

	expected := `
		# HELP adguard_up Whether the AdGuard Home instance is reachable (from /control/status).
		# TYPE adguard_up gauge
		adguard_up 0
	`
	if err := testutil.CollectAndCompare(c, strings.NewReader(expected), "adguard_up"); err != nil {
		t.Error(err)
	}
}

func TestCollectorBuildInfo(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c := New(srv.URL, "admin", "secret", 10)

	expected := `
		# HELP adguard_build_info AdGuard Home version and port information (from /control/status).
		# TYPE adguard_build_info gauge
		adguard_build_info{dns_port="53",http_port="3000",version="v0.107.72"} 1
	`
	if err := testutil.CollectAndCompare(c, strings.NewReader(expected), "adguard_build_info"); err != nil {
		t.Error(err)
	}
}

func TestCollectorDNSQueries(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c := New(srv.URL, "admin", "secret", 10)

	expected := `
		# HELP adguard_dns_queries Total DNS queries in the configured stats window (from /control/stats num_dns_queries). Rolling window total, not a monotonic counter.
		# TYPE adguard_dns_queries gauge
		adguard_dns_queries 100000
	`
	if err := testutil.CollectAndCompare(c, strings.NewReader(expected), "adguard_dns_queries"); err != nil {
		t.Error(err)
	}
}

func TestCollectorTopClients(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c := New(srv.URL, "admin", "secret", 10)

	expected := `
		# HELP adguard_top_clients Query count for a top client in the stats window (from /control/stats top_clients). The name label is resolved from persistent clients.
		# TYPE adguard_top_clients gauge
		adguard_top_clients{client="10.0.0.100",name="10.0.0.100"} 1200
		adguard_top_clients{client="10.0.0.3",name="TrueNAS"} 50000
	`
	if err := testutil.CollectAndCompare(c, strings.NewReader(expected), "adguard_top_clients"); err != nil {
		t.Error(err)
	}
}

func TestCollectorFiltering(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c := New(srv.URL, "admin", "secret", 10)

	expected := `
		# HELP adguard_filtering_rules_total Total number of rules across all enabled filter lists (computed from /control/filtering/status filters).
		# TYPE adguard_filtering_rules_total gauge
		adguard_filtering_rules_total 155880
	`
	if err := testutil.CollectAndCompare(c, strings.NewReader(expected), "adguard_filtering_rules_total"); err != nil {
		t.Error(err)
	}
}

func TestCollectorFilteringLists(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c := New(srv.URL, "admin", "secret", 10)

	expected := `
		# HELP adguard_filtering_lists_enabled Number of enabled filter lists (from /control/filtering/status filters).
		# TYPE adguard_filtering_lists_enabled gauge
		adguard_filtering_lists_enabled 1
		# HELP adguard_filtering_lists_total Total number of configured filter lists (from /control/filtering/status filters).
		# TYPE adguard_filtering_lists_total gauge
		adguard_filtering_lists_total 2
	`
	if err := testutil.CollectAndCompare(c, strings.NewReader(expected), "adguard_filtering_lists_total", "adguard_filtering_lists_enabled"); err != nil {
		t.Error(err)
	}
}

func TestCollectorProtection(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c := New(srv.URL, "admin", "secret", 10)

	expected := `
		# HELP adguard_safebrowsing_enabled Whether safe browsing protection is enabled (from /control/safebrowsing/status).
		# TYPE adguard_safebrowsing_enabled gauge
		adguard_safebrowsing_enabled 0
	`
	if err := testutil.CollectAndCompare(c, strings.NewReader(expected), "adguard_safebrowsing_enabled"); err != nil {
		t.Error(err)
	}
}

func TestCollectorDNSConfig(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	c := New(srv.URL, "admin", "secret", 10)

	expected := `
		# HELP adguard_dns_cache_size_bytes Configured DNS cache size in bytes (from /control/dns_info cache_size).
		# TYPE adguard_dns_cache_size_bytes gauge
		adguard_dns_cache_size_bytes 4.194304e+06
	`
	if err := testutil.CollectAndCompare(c, strings.NewReader(expected), "adguard_dns_cache_size_bytes"); err != nil {
		t.Error(err)
	}
}
