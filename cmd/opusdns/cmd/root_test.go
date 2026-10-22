package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeAPI records the requests the CLI sends and answers them with handler.
type fakeAPI struct {
	mu       sync.Mutex
	requests []string
}

func (f *fakeAPI) record(r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.requests = append(f.requests, r.Method+" "+r.URL.Path)
}

func (f *fakeAPI) seen() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.requests...)
}

// newFakeAPI starts a test server and points the CLI at it.
func newFakeAPI(t *testing.T, handler http.HandlerFunc) *fakeAPI {
	t.Helper()
	api := &fakeAPI{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		api.record(r)
		if handler != nil {
			handler(w, r)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(server.Close)

	t.Setenv("OPUSDNS_API_KEY", "opk_test")
	t.Setenv("OPUSDNS_API_ENDPOINT", server.URL)
	t.Setenv("OPUSDNS_DEBUG", "")
	return api
}

// resetCLI puts the package state and every flag back to its default, so
// that each test runs the command tree as if it were a fresh process.
func resetCLI() {
	apiKey = ""
	debug = false
	client = nil

	reset := func(f *pflag.Flag) {
		if sv, ok := f.Value.(pflag.SliceValue); ok {
			_ = sv.Replace(nil)
		} else {
			_ = f.Value.Set(f.DefValue)
		}
		f.Changed = false
	}
	var walk func(c *cobra.Command)
	walk = func(c *cobra.Command) {
		// PersistentPreRunE sets this on the command that ran.
		c.SilenceUsage = false
		c.Flags().VisitAll(reset)
		c.PersistentFlags().VisitAll(reset)
		for _, sub := range c.Commands() {
			walk(sub)
		}
	}
	walk(rootCmd)
}

// execute runs the CLI with args and stdin. It returns the error Execute
// returns and what cobra itself wrote to its output streams.
func execute(t *testing.T, stdin string, args ...string) (string, error) {
	t.Helper()
	resetCLI()

	var out bytes.Buffer
	rootCmd.SetArgs(args)
	rootCmd.SetIn(strings.NewReader(stdin))
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	t.Cleanup(func() {
		rootCmd.SetArgs(nil)
		rootCmd.SetIn(nil)
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
	})

	err := Execute()
	return out.String(), err
}

func TestCompletionRunsWithoutAPIKey(t *testing.T) {
	t.Setenv("OPUSDNS_API_KEY", "")

	out, err := execute(t, "", "completion", "bash")
	require.NoError(t, err)
	assert.Contains(t, out, "bash completion")

	_, err = execute(t, "", cobra.ShellCompRequestCmd, "zones", "")
	require.NoError(t, err)
}

func TestConfirmation(t *testing.T) {
	tests := []struct {
		name       string
		stdin      string
		args       []string
		wantDelete bool
	}{
		{name: "no input", stdin: "", wantDelete: false},
		{name: "answer no", stdin: "no\n", wantDelete: false},
		{name: "answer yes", stdin: "yes\n", wantDelete: true},
		{name: "answer yes without newline", stdin: "yes", wantDelete: true},
		{name: "force", stdin: "", args: []string{"--force"}, wantDelete: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := newFakeAPI(t, nil)

			args := append([]string{"zones", "delete", "example.com"}, tt.args...)
			_, err := execute(t, tt.stdin, args...)

			if tt.wantDelete {
				require.NoError(t, err)
				assert.Equal(t, []string{"DELETE /v1/dns/example.com"}, api.seen())
				return
			}
			require.ErrorIs(t, err, errNotConfirmed)
			assert.Empty(t, api.seen())
		})
	}
}

func TestUsersRoleSetRejectsRoleWithClear(t *testing.T) {
	api := newFakeAPI(t, nil)

	_, err := execute(t, "", "users", "role", "set", "user_01h45ytscbebyvny4gc8cr8ma2", "admin", "--clear")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not both")
	assert.Empty(t, api.seen())
}

func TestHostsRequireIP(t *testing.T) {
	api := newFakeAPI(t, nil)

	for _, sub := range []string{"create", "update"} {
		t.Run(sub, func(t *testing.T) {
			_, err := execute(t, "", "hosts", sub, "ns1.example.com")
			require.Error(t, err)
			assert.Contains(t, err.Error(), `required flag(s) "ip" not set`)
		})
	}
	assert.Empty(t, api.seen())
}

func TestRolesUpdateWithoutChanges(t *testing.T) {
	api := newFakeAPI(t, nil)

	_, err := execute(t, "", "roles", "update", "support")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nothing to update")
	assert.Empty(t, api.seen())
}

// TestTimeoutIsPerRequest lists three pages that each take longer than half
// the timeout. Under a deadline for the whole command, the listing failed.
func TestTimeoutIsPerRequest(t *testing.T) {
	const pages = 3
	api := newFakeAPI(t, func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(150 * time.Millisecond)
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"results": []map[string]interface{}{{"name": fmt.Sprintf("zone%d.com", page)}},
			"pagination": map[string]interface{}{
				"current_page":  page,
				"total_pages":   pages,
				"has_next_page": page < pages,
			},
		})
	})

	_, err := execute(t, "", "zones", "list", "--timeout", "300ms")
	require.NoError(t, err)
	assert.Len(t, api.seen(), pages)
	assert.Equal(t, 300*time.Millisecond, getClient().Config.HTTPTimeout)
}

func TestDebugFlag(t *testing.T) {
	tests := []struct {
		name string
		env  string
		args []string
		want bool
	}{
		{name: "default", want: false},
		{name: "env", env: "1", want: true},
		{name: "flag", args: []string{"--debug"}, want: true},
		{name: "flag overrides env", env: "true", args: []string{"--debug=false"}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newFakeAPI(t, func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(`{"results":[],"pagination":{"has_next_page":false}}`))
			})
			t.Setenv("OPUSDNS_DEBUG", tt.env)

			_, err := execute(t, "", append([]string{"zones", "list"}, tt.args...)...)
			require.NoError(t, err)
			assert.Equal(t, tt.want, getClient().Config.Debug)
			assert.NotNil(t, getClient().Config.Logger, "debug output must not go to stdout")
		})
	}
}

func TestErrorOutput(t *testing.T) {
	t.Run("API error prints no usage and no error", func(t *testing.T) {
		newFakeAPI(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		})

		out, err := execute(t, "", "zones", "get", "example.com")
		require.Error(t, err)
		// main prints the error once; cobra must print nothing itself.
		assert.Empty(t, out)
	})

	t.Run("wrong arguments still print usage", func(t *testing.T) {
		out, err := execute(t, "", "zones", "delete")
		require.Error(t, err)
		assert.Contains(t, out, "Usage:")
		assert.NotContains(t, out, "Error:")
	})
}
