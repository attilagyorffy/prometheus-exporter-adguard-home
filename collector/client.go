package collector

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// Client talks to the AdGuard Home REST API.
type Client struct {
	baseURL    string
	username   string
	password   string
	httpClient *http.Client
}

// NewClient creates an AdGuard Home API client with basic auth.
func NewClient(baseURL, username, password string) *Client {
	return &Client{
		baseURL:  baseURL,
		username: username,
		password: password,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) fetchJSON(path string, target any) error {
	req, err := http.NewRequest("GET", c.baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.SetBasicAuth(c.username, c.password)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request %s: %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d from %s", resp.StatusCode, path)
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("decode response from %s: %w", path, err)
	}
	return nil
}

// StatusResponse represents /control/status.
type StatusResponse struct {
	Version           string `json:"version"`
	DNSPort           int    `json:"dns_port"`
	HTTPPort          int    `json:"http_port"`
	ProtectionEnabled bool   `json:"protection_enabled"`
	Running           bool   `json:"running"`
}

// StatsResponse represents /control/stats.
type StatsResponse struct {
	NumDNSQueries           float64              `json:"num_dns_queries"`
	NumBlockedFiltering     float64              `json:"num_blocked_filtering"`
	NumReplacedSafebrowsing float64              `json:"num_replaced_safebrowsing"`
	NumReplacedSafesearch   float64              `json:"num_replaced_safesearch"`
	NumReplacedParental     float64              `json:"num_replaced_parental"`
	AvgProcessingTime       float64              `json:"avg_processing_time"`
	TopQueriedDomains       []map[string]float64 `json:"top_queried_domains"`
	TopBlockedDomains       []map[string]float64 `json:"top_blocked_domains"`
	TopClients              []map[string]float64 `json:"top_clients"`
	TopUpstreamsResponses   []map[string]float64 `json:"top_upstreams_responses"`
	TopUpstreamsAvgTime     []map[string]float64 `json:"top_upstreams_avg_time"`
}

// DNSInfoResponse represents /control/dns_info.
type DNSInfoResponse struct {
	Ratelimit       float64 `json:"ratelimit"`
	BlockingMode    string  `json:"blocking_mode"`
	DNSSECEnabled   bool    `json:"dnssec_enabled"`
	CacheSize       float64 `json:"cache_size"`
	CacheEnabled    bool    `json:"cache_enabled"`
	CacheOptimistic bool    `json:"cache_optimistic"`
	UpstreamMode    string  `json:"upstream_mode"`
}

// FilteringStatusResponse represents /control/filtering/status.
type FilteringStatusResponse struct {
	Filters   []Filter `json:"filters"`
	UserRules []string `json:"user_rules"`
	Interval  float64  `json:"interval"`
	Enabled   bool     `json:"enabled"`
}

// Filter represents a single filter list in /control/filtering/status.
type Filter struct {
	Name        string  `json:"name"`
	LastUpdated string  `json:"last_updated"`
	RulesCount  float64 `json:"rules_count"`
	Enabled     bool    `json:"enabled"`
}

// EnabledResponse represents simple {"enabled": bool} responses.
type EnabledResponse struct {
	Enabled bool `json:"enabled"`
}

// ClientsResponse represents /control/clients.
type ClientsResponse struct {
	Clients []PersistentClient `json:"clients"`
}

// PersistentClient represents a persistent client entry.
type PersistentClient struct {
	Name string   `json:"name"`
	IDs  []string `json:"ids"`
}

// FetchStatus retrieves /control/status.
func (c *Client) FetchStatus() (*StatusResponse, error) {
	var resp StatusResponse
	if err := c.fetchJSON("/control/status", &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// FetchStats retrieves /control/stats.
func (c *Client) FetchStats() (*StatsResponse, error) {
	var resp StatsResponse
	if err := c.fetchJSON("/control/stats", &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// FetchDNSInfo retrieves /control/dns_info.
func (c *Client) FetchDNSInfo() (*DNSInfoResponse, error) {
	var resp DNSInfoResponse
	if err := c.fetchJSON("/control/dns_info", &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// FetchFilteringStatus retrieves /control/filtering/status.
func (c *Client) FetchFilteringStatus() (*FilteringStatusResponse, error) {
	var resp FilteringStatusResponse
	if err := c.fetchJSON("/control/filtering/status", &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// FetchEnabled retrieves a simple {"enabled": bool} endpoint.
func (c *Client) FetchEnabled(path string) (bool, error) {
	var resp EnabledResponse
	if err := c.fetchJSON(path, &resp); err != nil {
		return false, err
	}
	return resp.Enabled, nil
}

// BuildClientMap fetches /control/clients and returns a map of identifier to
// persistent client name for IP-to-name resolution.
func (c *Client) BuildClientMap() map[string]string {
	var resp ClientsResponse
	if err := c.fetchJSON("/control/clients", &resp); err != nil {
		slog.Warn("failed to fetch client list for name resolution", "error", err)
		return nil
	}
	m := make(map[string]string)
	for _, cl := range resp.Clients {
		for _, id := range cl.IDs {
			m[id] = cl.Name
		}
	}
	return m
}
