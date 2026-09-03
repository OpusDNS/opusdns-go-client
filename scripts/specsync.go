//go:build ignore

// Command specsync vendors the published OpusDNS OpenAPI specification into
// spec/ and records exactly which upstream revision it was taken from.
//
//	go run scripts/specsync.go                 # follow api-spec main
//	go run scripts/specsync.go -check          # report drift, write nothing
//	go run scripts/specsync.go -ref <sha>      # pin to one api-spec commit
//	go run scripts/specsync.go -spec-url <url> # any other source (debugging)
//
// Use `make spec-sync`, which is what CI runs, so a local refresh and an
// automated one write identical bytes. This command is the single canonical
// writer of spec/openapi.yaml and spec/version.yaml.
//
// The spec is stored byte-for-byte as published: a diff against a previous
// sync is then exactly the upstream diff, and openapi-changes reads the YAML
// directly. Nothing here converts, reformats or filters the document.
//
// The version stamp deliberately carries no timestamp of its own, so running
// this twice against an unchanged upstream leaves the working tree clean.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	// specRepo is the source of truth. Note that the Scalar-rendered document
	// on developers.opusdns.com lags behind by days and must not be used.
	specRepo    = "OpusDNS/api-spec"
	specRepoSrc = "src/openapi.yaml"
	specRepoPkg = "package.json"

	outSpec    = "spec/openapi.yaml"
	outVersion = "spec/version.yaml"
)

type versionFile struct {
	NpmVersion    string `yaml:"npm_version"`
	InfoVersion   string `yaml:"info_version"`
	APISpecCommit string `yaml:"api_spec_commit"`
}

func main() {
	ref := flag.String("ref", "main", "api-spec git ref (branch, tag or commit sha) to vendor the spec from")
	specURL := flag.String("spec-url", "", "fetch the spec from this URL instead of api-spec (skips commit pinning)")
	check := flag.Bool("check", false, "report whether the vendored spec is behind upstream and exit non-zero if it is; writes nothing")
	flag.Parse()

	before := readVersionFile()

	var (
		commit string
		source string
		err    error
	)
	if *specURL != "" {
		source = *specURL
	} else {
		commit, err = resolveCommit(*ref)
		if err != nil {
			fatal(fmt.Errorf("resolve api-spec ref %q: %w", *ref, err))
		}
		source = rawURL(commit, specRepoSrc)
	}

	spec, err := fetch(source)
	if err != nil {
		fatal(fmt.Errorf("fetch spec: %w", err))
	}
	infoVersion, err := specInfoVersion(spec)
	if err != nil {
		fatal(err)
	}

	// The npm version always comes from main: api-spec commits the new spec
	// first and runs `npm version minor` afterwards, so package.json at the
	// dispatched commit still holds the previous number.
	npmVersion, err := packageVersion()
	if err != nil {
		fatal(fmt.Errorf("read api-spec package.json: %w", err))
	}

	after := versionFile{NpmVersion: npmVersion, InfoVersion: infoVersion, APISpecCommit: commit}

	if *check {
		reportCheck(before, after, spec)
		return
	}

	if err := writeFiles(spec, after); err != nil {
		fatal(err)
	}

	fmt.Printf("source:       %s\n", source)
	fmt.Printf("npm_version:  %s\n", change(before.NpmVersion, after.NpmVersion))
	fmt.Printf("info_version: %s\n", change(before.InfoVersion, after.InfoVersion))
	fmt.Printf("commit:       %s\n", change(before.APISpecCommit, after.APISpecCommit))
	fmt.Printf("wrote %s (%d bytes) and %s\n", outSpec, len(spec), outVersion)
}

// reportCheck compares the vendored spec with what upstream publishes and exits
// non-zero when they differ, so it can gate a scheduled job. It writes nothing,
// so it is safe to run against a dirty working tree.
func reportCheck(before, after versionFile, spec []byte) {
	current, err := os.ReadFile(outSpec)
	if err != nil {
		fatal(fmt.Errorf("read %s: %w (run `make spec-sync` first)", outSpec, err))
	}

	if bytes.Equal(current, spec) && before.InfoVersion == after.InfoVersion {
		fmt.Printf("up to date with api-spec: npm %s, spec %s\n", before.NpmVersion, before.InfoVersion)
		return
	}

	fmt.Printf("api-spec has moved on:\n")
	fmt.Printf("  npm_version:  %s\n", change(before.NpmVersion, after.NpmVersion))
	fmt.Printf("  info_version: %s\n", change(before.InfoVersion, after.InfoVersion))
	fmt.Printf("  commit:       %s\n", change(before.APISpecCommit, after.APISpecCommit))
	fmt.Printf("\nRun `make spec-sync` to vendor it, then `make spec-check` for what needs code.\n")
	os.Exit(1)
}

// resolveCommit turns a branch, tag or sha into a full commit sha, so the
// vendored spec always records an immutable source revision.
func resolveCommit(ref string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/commits/%s", specRepo, ref)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	// This media type makes the API answer with the bare sha as plain text.
	req.Header.Set("Accept", "application/vnd.github.sha")
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	body, err := do(req)
	if err != nil {
		return "", err
	}
	sha := strings.TrimSpace(string(body))
	if len(sha) != 40 {
		return "", fmt.Errorf("unexpected sha %q", sha)
	}
	return sha, nil
}

func packageVersion() (string, error) {
	body, err := fetch(rawURL("main", specRepoPkg))
	if err != nil {
		return "", err
	}
	var pkg struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(body, &pkg); err != nil {
		return "", err
	}
	if pkg.Version == "" {
		return "", fmt.Errorf("no version field")
	}
	return pkg.Version, nil
}

func specInfoVersion(spec []byte) (string, error) {
	var doc struct {
		Info struct {
			Version string `yaml:"version"`
		} `yaml:"info"`
	}
	if err := yaml.Unmarshal(spec, &doc); err != nil {
		return "", fmt.Errorf("parse spec: %w", err)
	}
	if doc.Info.Version == "" {
		return "", fmt.Errorf("spec has no info.version")
	}
	return doc.Info.Version, nil
}

func writeFiles(spec []byte, v versionFile) error {
	if err := os.MkdirAll(filepath.Dir(outSpec), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(outSpec, spec, 0o644); err != nil {
		return err
	}

	var buf strings.Builder
	buf.WriteString("# Which published OpenAPI specification spec/openapi.yaml was taken from.\n")
	buf.WriteString("# Written by `make spec-sync`; do not edit by hand.\n")
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(v); err != nil {
		return err
	}
	if err := enc.Close(); err != nil {
		return err
	}
	return os.WriteFile(outVersion, []byte(buf.String()), 0o644)
}

func readVersionFile() versionFile {
	var v versionFile
	raw, err := os.ReadFile(outVersion)
	if err != nil {
		return v
	}
	_ = yaml.Unmarshal(raw, &v)
	return v
}

func rawURL(ref, path string) string {
	return fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s", specRepo, ref, path)
}

func fetch(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return do(req)
}

func do(req *http.Request) ([]byte, error) {
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: %s", req.URL, resp.Status)
	}
	return io.ReadAll(resp.Body)
}

func change(before, after string) string {
	if before == "" {
		return after
	}
	if before == after {
		return after + " (unchanged)"
	}
	return before + " → " + after
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "specsync:", err)
	os.Exit(1)
}
