package opusdns

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
				APIKey:          "opk_test123",
				APIEndpoint:     DefaultAPIEndpoint,
				TTL:             DefaultTTL,
				HTTPTimeout:     DefaultTimeout,
				MaxRetries:      DefaultMaxRetries,
				PollingInterval: DefaultPollingInterval,
				PollingTimeout:  DefaultPollingTimeout,
				DNSResolvers:    []string{"ns1.opusdns.com:53", "ns2.opusdns.net:53"},
			},
		},
		{
			name: "custom values",
			config: &Config{
				APIKey:          "opk_custom",
				APIEndpoint:     "https://sandbox.opusdns.com",
				TTL:             120,
				HTTPTimeout:     60 * time.Second,
				MaxRetries:      5,
				PollingInterval: 10 * time.Second,
				PollingTimeout:  120 * time.Second,
				DNSResolvers:    []string{"1.1.1.1:53"},
			},
			want: &Config{
				APIKey:          "opk_custom",
				APIEndpoint:     "https://sandbox.opusdns.com",
				TTL:             120,
				HTTPTimeout:     60 * time.Second,
				MaxRetries:      5,
				PollingInterval: 10 * time.Second,
				PollingTimeout:  120 * time.Second,
				DNSResolvers:    []string{"1.1.1.1:53"},
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
			assert.Equal(t, tt.want.PollingInterval, client.config.PollingInterval)
			assert.Equal(t, tt.want.PollingTimeout, client.config.PollingTimeout)
			assert.Equal(t, tt.want.DNSResolvers, client.config.DNSResolvers)
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
				json.NewEncoder(w).Encode(tt.response)
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
		name      string
		fqdn      string
		zones     []Zone
		wantZone  string
		wantErr   bool
	}{
		{
			name: "exact match",
			fqdn: "example.com",
			zones: []Zone{
				{Name: "example.com"},
			},
			wantZone: "example.com",
			wantErr:  false,
		},
		{
			name: "subdomain match",
			fqdn: "_acme-challenge.www.example.com",
			zones: []Zone{
				{Name: "example.com"},
			},
			wantZone: "example.com",
			wantErr:  false,
		},
		{
			name: "longest match",
			fqdn: "_acme-challenge.sub.example.com",
			zones: []Zone{
				{Name: "example.com"},
				{Name: "sub.example.com"},
			},
			wantZone: "sub.example.com",
			wantErr:  false,
		},
		{
			name: "no match",
			fqdn: "notfound.com",
			zones: []Zone{
				{Name: "example.com"},
			},
			wantZone: "",
			wantErr:  true,
		},
		{
			name:     "no zones",
			fqdn:     "example.com",
			zones:    []Zone{},
			wantZone: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				resp := ZoneListResponse{
					Results: tt.zones,
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
		zones      []Zone
		wantErr    bool
		wantOp     string
		wantName   string
		wantType   string
		wantRData  string
	}{
		{
			name:  "successful upsert",
			fqdn:  "_acme-challenge.example.com",
			value: "challenge-value",
			zones: []Zone{
				{Name: "example.com"},
			},
			wantErr:   false,
			wantOp:    "upsert",
			wantName:  "_acme-challenge",
			wantType:  "TXT",
			wantRData: "\"challenge-value\"",
		},
		{
			name:  "value already quoted",
			fqdn:  "_acme-challenge.example.com",
			value: "\"challenge-value\"",
			zones: []Zone{
				{Name: "example.com"},
			},
			wantErr:   false,
			wantOp:    "upsert",
			wantName:  "_acme-challenge",
			wantType:  "TXT",
			wantRData: "\"challenge-value\"",
		},
		{
			name:  "subdomain",
			fqdn:  "_acme-challenge.sub.example.com",
			value: "test",
			zones: []Zone{
				{Name: "example.com"},
			},
			wantErr:   false,
			wantOp:    "upsert",
			wantName:  "_acme-challenge.sub",
			wantType:  "TXT",
			wantRData: "\"test\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var receivedReq RRSetPatchRequest

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/v1/dns" {
					resp := ZoneListResponse{
						Results: tt.zones,
						Pagination: Pagination{
							TotalPages:  1,
							CurrentPage: 1,
							HasNextPage: false,
						},
					}
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(resp)
					return
				}

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
		zones      []Zone
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
			zones: []Zone{
				{Name: "example.com"},
			},
			wantErr:   false,
			wantOp:    "remove",
			wantName:  "_acme-challenge",
			wantType:  "TXT",
			wantTTL:   DefaultTTL,
			wantRData: "\"test-value\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var receivedReq RRSetPatchRequest

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/v1/dns" {
					resp := ZoneListResponse{
						Results: tt.zones,
						Pagination: Pagination{
							TotalPages:  1,
							CurrentPage: 1,
							HasNextPage: false,
						},
					}
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(resp)
					return
				}

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

func TestZoneCache(t *testing.T) {
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

	// Second call should use cache
	zones2, err := client.ListZones()
	assert.NoError(t, err)
	assert.Len(t, zones2, 1)
	assert.Equal(t, 1, calls, "should still be 1 due to cache")
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
