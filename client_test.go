package opusdns

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		want   *Config
	}{
		{
			name: "default values",
			config: &Config{
				APIKey: "opk_test123",
			},
			want: &Config{
				APIKey:      "opk_test123",
				APIEndpoint: DefaultAPIEndpoint,
				TTL:         DefaultTTL,
				HTTPTimeout: DefaultTimeout,
				MaxRetries:  DefaultMaxRetries,
			},
		},
		{
			name: "custom values",
			config: &Config{
				APIKey:      "opk_custom",
				APIEndpoint: "https://sandbox.opusdns.com",
				TTL:         120,
				HTTPTimeout: 60 * time.Second,
				MaxRetries:  5,
			},
			want: &Config{
				APIKey:      "opk_custom",
				APIEndpoint: "https://sandbox.opusdns.com",
				TTL:         120,
				HTTPTimeout: 60 * time.Second,
				MaxRetries:  5,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.config)
			assert.Equal(t, tt.want.APIKey, client.config.APIKey)
			assert.Equal(t, tt.want.APIEndpoint, client.config.APIEndpoint)
			assert.Equal(t, tt.want.TTL, client.config.TTL)
			assert.Equal(t, tt.want.HTTPTimeout, client.config.HTTPTimeout)
			assert.Equal(t, tt.want.MaxRetries, client.config.MaxRetries)
		})
	}
}

func TestListZones(t *testing.T) {
	tests := []struct {
		name       string
		response   interface{}
		statusCode int
		wantErr    bool
		wantZones  int
	}{
		{
			name: "successful single page",
			response: ZoneListResponse{
				Results: []Zone{
					{Name: "example.com", DNSSECStatus: "disabled"},
					{Name: "test.com", DNSSECStatus: "enabled"},
				},
				Pagination: Pagination{
					TotalPages:  1,
					CurrentPage: 1,
					HasNextPage: false,
				},
			},
			statusCode: http.StatusOK,
			wantErr:    false,
			wantZones:  2,
		},
		{
			name:       "server error",
			response:   map[string]string{"error": "internal server error"},
			statusCode: http.StatusInternalServerError,
			wantErr:    true,
		},
		{
			name:       "unauthorized",
			response:   map[string]string{"message": "invalid API key"},
			statusCode: http.StatusUnauthorized,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/v1/dns", r.URL.Path)
				assert.Equal(t, "opk_test123", r.Header.Get("X-Api-Key"))

				w.WriteHeader(tt.statusCode)
				_ = json.NewEncoder(w).Encode(tt.response)
			}))
			defer server.Close()

			client := NewClient(&Config{
				APIKey:      "opk_test123",
				APIEndpoint: server.URL,
			})

			zones, err := client.ListZones()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, zones, tt.wantZones)
			}
		})
	}
}

func TestFindZoneForFQDN(t *testing.T) {
	tests := []struct {
		name       string
		fqdn       string
		validZones map[string]bool // zones that exist
		wantZone   string
		wantErr    bool
	}{
		{
			name:       "subdomain match",
			fqdn:       "_acme-challenge.www.example.com",
			validZones: map[string]bool{"example.com": true},
			wantZone:   "example.com",
			wantErr:    false,
		},
		{
			name:       "longer zone match first",
			fqdn:       "_acme-challenge.sub.example.com",
			validZones: map[string]bool{"sub.example.com": true, "example.com": true},
			wantZone:   "sub.example.com",
			wantErr:    false,
		},
		{
			name:       "no match",
			fqdn:       "_acme-challenge.notfound.com",
			validZones: map[string]bool{"example.com": true},
			wantZone:   "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Extract zone from path: /v1/dns/{zone}
				path := strings.TrimPrefix(r.URL.Path, "/v1/dns/")
				if tt.validZones[path] {
					// Valid zone response with dnssec_status
					resp := map[string]interface{}{
						"name":          path + ".",
						"dnssec_status": "disabled",
					}
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(resp)
				} else {
					// Zone not found
					w.WriteHeader(http.StatusNotFound)
					_ = json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
				}
			}))
			defer server.Close()

			client := NewClient(&Config{
				APIKey:      "opk_test123",
				APIEndpoint: server.URL,
			})

			zone, err := client.FindZoneForFQDN(tt.fqdn)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantZone, zone)
			}
		})
	}
}

func TestUpsertTXTRecord(t *testing.T) {
	tests := []struct {
		name       string
		fqdn       string
		value      string
		validZones map[string]bool
		wantErr    bool
		wantOp     string
		wantName   string
		wantType   string
		wantRData  string
	}{
		{
			name:       "successful upsert",
			fqdn:       "_acme-challenge.example.com",
			value:      "challenge-value",
			validZones: map[string]bool{"example.com": true},
			wantErr:    false,
			wantOp:     "upsert",
			wantName:   "_acme-challenge",
			wantType:   "TXT",
			wantRData:  "\"challenge-value\"",
		},
		{
			name:       "value already quoted",
			fqdn:       "_acme-challenge.example.com",
			value:      "\"challenge-value\"",
			validZones: map[string]bool{"example.com": true},
			wantErr:    false,
			wantOp:     "upsert",
			wantName:   "_acme-challenge",
			wantType:   "TXT",
			wantRData:  "\"challenge-value\"",
		},
		{
			name:       "subdomain",
			fqdn:       "_acme-challenge.sub.example.com",
			value:      "test",
			validZones: map[string]bool{"example.com": true},
			wantErr:    false,
			wantOp:     "upsert",
			wantName:   "_acme-challenge.sub",
			wantType:   "TXT",
			wantRData:  "\"test\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var receivedReq RRSetPatchRequest

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Handle zone detection GET requests
				if r.Method == "GET" && strings.HasPrefix(r.URL.Path, "/v1/dns/") {
					zone := strings.TrimPrefix(r.URL.Path, "/v1/dns/")
					if tt.validZones[zone] {
						resp := map[string]interface{}{
							"name":          zone + ".",
							"dnssec_status": "disabled",
						}
						w.WriteHeader(http.StatusOK)
						json.NewEncoder(w).Encode(resp)
					} else {
						w.WriteHeader(http.StatusNotFound)
						json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
					}
					return
				}

				// Handle PATCH request
				assert.Equal(t, "PATCH", r.Method)
				assert.Equal(t, "opk_test123", r.Header.Get("X-Api-Key"))

				err := json.NewDecoder(r.Body).Decode(&receivedReq)
				require.NoError(t, err)

				w.WriteHeader(http.StatusNoContent)
			}))
			defer server.Close()

			client := NewClient(&Config{
				APIKey:      "opk_test123",
				APIEndpoint: server.URL,
				TTL:         60,
			})

			err := client.UpsertTXTRecord(tt.fqdn, tt.value)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				require.Len(t, receivedReq.Ops, 1)
				assert.Equal(t, tt.wantOp, receivedReq.Ops[0].Op)
				assert.Equal(t, tt.wantName, receivedReq.Ops[0].Record.Name)
				assert.Equal(t, tt.wantType, receivedReq.Ops[0].Record.Type)
				assert.Equal(t, 60, receivedReq.Ops[0].Record.TTL)
				assert.Equal(t, tt.wantRData, receivedReq.Ops[0].Record.RData)
			}
		})
	}
}

func TestRemoveTXTRecord(t *testing.T) {
	tests := []struct {
		name       string
		fqdn       string
		value      string
		validZones map[string]bool
		wantErr    bool
		wantOp     string
		wantName   string
		wantType   string
		wantTTL    int
		wantRData  string
	}{
		{
			name:       "successful remove",
			fqdn:       "_acme-challenge.example.com",
			value:      "test-value",
			validZones: map[string]bool{"example.com": true},
			wantErr:    false,
			wantOp:     "remove",
			wantName:   "_acme-challenge",
			wantType:   "TXT",
			wantTTL:    DefaultTTL,
			wantRData:  "\"test-value\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var receivedReq RRSetPatchRequest

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Handle zone detection GET requests
				if r.Method == "GET" && strings.HasPrefix(r.URL.Path, "/v1/dns/") {
					zone := strings.TrimPrefix(r.URL.Path, "/v1/dns/")
					if tt.validZones[zone] {
						resp := map[string]interface{}{
							"name":          zone + ".",
							"dnssec_status": "disabled",
						}
						w.WriteHeader(http.StatusOK)
						json.NewEncoder(w).Encode(resp)
					} else {
						w.WriteHeader(http.StatusNotFound)
						json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
					}
					return
				}

				// Handle PATCH request
				assert.Equal(t, "PATCH", r.Method)
				err := json.NewDecoder(r.Body).Decode(&receivedReq)
				require.NoError(t, err)

				w.WriteHeader(http.StatusNoContent)
			}))
			defer server.Close()

			client := NewClient(&Config{
				APIKey:      "opk_test123",
				APIEndpoint: server.URL,
			})

			err := client.RemoveTXTRecord(tt.fqdn, tt.value)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				require.Len(t, receivedReq.Ops, 1)
				assert.Equal(t, tt.wantOp, receivedReq.Ops[0].Op)
				assert.Equal(t, tt.wantName, receivedReq.Ops[0].Record.Name)
				assert.Equal(t, tt.wantType, receivedReq.Ops[0].Record.Type)
				assert.Equal(t, tt.wantTTL, receivedReq.Ops[0].Record.TTL)
				assert.Equal(t, tt.wantRData, receivedReq.Ops[0].Record.RData)
			}
		})
	}
}

func TestRetryLogic(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]string{"error": "rate limited"})
			return
		}
		resp := ZoneListResponse{
			Results: []Zone{
				{Name: "example.com"},
			},
			Pagination: Pagination{
				TotalPages:  1,
				CurrentPage: 1,
				HasNextPage: false,
			},
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(&Config{
		APIKey:      "opk_test123",
		APIEndpoint: server.URL,
		MaxRetries:  3,
	})

	zones, err := client.ListZones()
	assert.NoError(t, err)
	assert.Len(t, zones, 1)
	assert.Equal(t, 3, attempts)
}

func TestZoneOperations(t *testing.T) {
	// Tests for ListZones now makes multiple calls (no cache)
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		resp := ZoneListResponse{
			Results: []Zone{
				{Name: "example.com"},
			},
			Pagination: Pagination{
				TotalPages:  1,
				CurrentPage: 1,
				HasNextPage: false,
			},
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(&Config{
		APIKey:      "opk_test123",
		APIEndpoint: server.URL,
	})

	// First call should hit the API
	zones1, err := client.ListZones()
	assert.NoError(t, err)
	assert.Len(t, zones1, 1)
	assert.Equal(t, 1, calls)

	// Second call should also hit the API (no cache)
	zones2, err := client.ListZones()
	assert.NoError(t, err)
	assert.Len(t, zones2, 1)
	assert.Equal(t, 2, calls, "should be 2 as there is no cache")
}

func TestPagination(t *testing.T) {
	page := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page++
		
		var resp ZoneListResponse
		if page == 1 {
			resp = ZoneListResponse{
				Results: []Zone{
					{Name: "zone1.com"},
					{Name: "zone2.com"},
				},
				Pagination: Pagination{
					TotalPages:  2,
					CurrentPage: 1,
					HasNextPage: true,
				},
			}
		} else {
			resp = ZoneListResponse{
				Results: []Zone{
					{Name: "zone3.com"},
				},
				Pagination: Pagination{
					TotalPages:  2,
					CurrentPage: 2,
					HasNextPage: false,
				},
			}
		}
		
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(&Config{
		APIKey:      "opk_test123",
		APIEndpoint: server.URL,
	})

	zones, err := client.ListZones()
	assert.NoError(t, err)
	assert.Len(t, zones, 3)
	assert.Equal(t, "zone1.com", zones[0].Name)
	assert.Equal(t, "zone2.com", zones[1].Name)
	assert.Equal(t, "zone3.com", zones[2].Name)
}
