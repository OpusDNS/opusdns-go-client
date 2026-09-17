package opusdns

import (
	"context"
	"net/http"
	"net/http/httptest"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildInfoWithDep(path, version string) func() (*debug.BuildInfo, bool) {
	return func() (*debug.BuildInfo, bool) {
		return &debug.BuildInfo{
			Main: debug.Module{Path: "example.com/consumer", Version: "v1.0.0"},
			Deps: []*debug.Module{{Path: path, Version: version}},
		}, true
	}
}

func TestResolveVersion(t *testing.T) {
	noBuildInfo := func() (*debug.BuildInfo, bool) { return nil, false }

	t.Run("ldflags version wins", func(t *testing.T) {
		assert.Equal(t, "v1.2.3", resolveVersion("v1.2.3", buildInfoWithDep(modulePath, "v0.9.0")))
	})

	t.Run("falls back to the module version recorded in the consumer", func(t *testing.T) {
		assert.Equal(t, "v0.9.0", resolveVersion("dev", buildInfoWithDep(modulePath, "v0.9.0")))
	})

	t.Run("reads the main module when this module is the binary", func(t *testing.T) {
		readBuildInfo := func() (*debug.BuildInfo, bool) {
			return &debug.BuildInfo{Main: debug.Module{Path: modulePath, Version: "v2.0.0"}}, true
		}
		assert.Equal(t, "v2.0.0", resolveVersion("dev", readBuildInfo))
	})

	t.Run("prefers the replacement of a replaced module", func(t *testing.T) {
		readBuildInfo := func() (*debug.BuildInfo, bool) {
			return &debug.BuildInfo{
				Deps: []*debug.Module{{
					Path:    modulePath,
					Version: "v0.9.0",
					Replace: &debug.Module{Path: modulePath, Version: "v0.9.1"},
				}},
			}, true
		}
		assert.Equal(t, "v0.9.1", resolveVersion("dev", readBuildInfo))
	})

	// The API drops a version outside its charset, so sending one costs the whole version
	// dimension for that request rather than degrading it.
	t.Run("rejects a version the API would not store", func(t *testing.T) {
		for name, version := range map[string]string{
			"unversioned build": "(devel)",
			"empty":             "",
			"spaces":            "1.0 beta",
			"markup":            "<script>",
			"too long":          strings.Repeat("9", 65),
		} {
			t.Run(name, func(t *testing.T) {
				assert.Equal(t, unknownVersion, resolveVersion("dev", buildInfoWithDep(modulePath, version)))
				assert.Equal(t, unknownVersion, resolveVersion(version, noBuildInfo))
			})
		}
	})

	t.Run("falls back when there is no build info", func(t *testing.T) {
		assert.Equal(t, unknownVersion, resolveVersion("dev", noBuildInfo))
	})

	t.Run("falls back when this module is not in the build", func(t *testing.T) {
		assert.Equal(t, unknownVersion, resolveVersion("dev", buildInfoWithDep("example.com/other", "v1.0.0")))
	})
}

func TestClientTokenDefaults(t *testing.T) {
	client, err := NewClient(WithAPIKey("opk_test"))
	require.NoError(t, err)

	assert.Equal(t, ProductName+"/"+ResolvedVersion(), client.Config.ClientToken)
	assert.True(t, strings.HasPrefix(client.Config.ClientToken, ProductName+"/"))
	// The token must not be derived from the User-Agent: replacing that is a supported thing
	// to do and would otherwise silently drop the attribution.
	assert.Equal(t, GetClientToken(), NewConfig(WithUserAgent("something-else/1.0")).ClientToken)
}

func TestClientHeaderIsSent(t *testing.T) {
	for name, tc := range map[string]struct {
		options  []Option
		expected string
	}{
		"by default": {
			expected: GetClientToken(),
		},
		"overridden by an embedding product": {
			options:  []Option{WithClientToken("acme-provisioner/2.1")},
			expected: "acme-provisioner/2.1",
		},
		"omitted when cleared": {
			options:  []Option{WithClientToken("")},
			expected: "",
		},
	} {
		t.Run(name, func(t *testing.T) {
			var got string
			var userAgent string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				got = r.Header.Get("X-OpusDNS-Client")
				userAgent = r.Header.Get("User-Agent")
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{}`))
			}))
			defer server.Close()

			options := append([]Option{WithAPIKey("opk_test"), WithAPIEndpoint(server.URL)}, tc.options...)
			client, err := NewClient(options...)
			require.NoError(t, err)

			_, err = client.HTTPClient().Get(context.Background(), "/tlds", nil)
			require.NoError(t, err)

			assert.Equal(t, tc.expected, got)
			// The User-Agent is independent of the channel token and always sent.
			assert.Equal(t, GetUserAgent(), userAgent)
		})
	}
}
