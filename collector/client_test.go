package collector

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/control/status" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		user, pass, ok := r.BasicAuth()
		if !ok || user != "admin" || pass != "secret" {
			t.Errorf("unexpected auth: ok=%v user=%s", ok, user)
		}
		json.NewEncoder(w).Encode(StatusResponse{
			Version:           "v0.107.72",
			DNSPort:           53,
			HTTPPort:          3000,
			ProtectionEnabled: true,
			Running:           true,
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "admin", "secret")
	status, err := c.FetchStatus()
	if err != nil {
		t.Fatalf("FetchStatus: %v", err)
	}
	if status.Version != "v0.107.72" {
		t.Errorf("version = %q, want v0.107.72", status.Version)
	}
	if !status.Running {
		t.Error("running = false, want true")
	}
}

func TestFetchStatusUnauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "admin", "wrong")
	_, err := c.FetchStatus()
	if err == nil {
		t.Fatal("expected error for 401 response")
	}
}

func TestFetchStats(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(StatsResponse{
			NumDNSQueries:       100000,
			NumBlockedFiltering: 15000,
			AvgProcessingTime:   0.025,
			TopQueriedDomains:   []map[string]float64{{"example.com": 500}},
			TopClients:          []map[string]float64{{"10.0.0.3": 50000}},
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "admin", "secret")
	stats, err := c.FetchStats()
	if err != nil {
		t.Fatalf("FetchStats: %v", err)
	}
	if stats.NumDNSQueries != 100000 {
		t.Errorf("num_dns_queries = %f, want 100000", stats.NumDNSQueries)
	}
	if len(stats.TopQueriedDomains) != 1 {
		t.Errorf("top_queried_domains count = %d, want 1", len(stats.TopQueriedDomains))
	}
}

func TestBuildClientMap(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(ClientsResponse{
			Clients: []PersistentClient{
				{Name: "TrueNAS", IDs: []string{"10.0.0.3", "04:42:1a:0d:cf:fe"}},
				{Name: "Laptop", IDs: []string{"10.0.0.11"}},
			},
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "admin", "secret")
	m := c.BuildClientMap()
	if m == nil {
		t.Fatal("client map is nil")
	}
	if m["10.0.0.3"] != "TrueNAS" {
		t.Errorf("10.0.0.3 = %q, want TrueNAS", m["10.0.0.3"])
	}
	if m["04:42:1a:0d:cf:fe"] != "TrueNAS" {
		t.Errorf("MAC = %q, want TrueNAS", m["04:42:1a:0d:cf:fe"])
	}
	if m["10.0.0.11"] != "Laptop" {
		t.Errorf("10.0.0.11 = %q, want Laptop", m["10.0.0.11"])
	}
	if m["10.0.0.99"] != "" {
		t.Errorf("unknown IP = %q, want empty", m["10.0.0.99"])
	}
}

func TestBuildClientMapServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "admin", "secret")
	m := c.BuildClientMap()
	if m != nil {
		t.Errorf("expected nil map on error, got %v", m)
	}
}
