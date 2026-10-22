//go:build ignore

// Command specsync vendors the published OpusDNS OpenAPI specification into
// spec/ and records exactly which upstream revision it was taken from.
//
//	go run scripts/specsync.go                 # follow api-spec main
//	go run scripts/specsync.go -check          # report drift, write nothing
//	go run scripts/specsync.go -ref <sha>      # pin to one api-spec commit
//	go run scripts/specsync.go -spec-url <url> # any other source (debugging)
//
// Use `make spec-sync`, so every refresh runs with the same flags and writes
// identical bytes. Nothing in CI runs it: the sync is manual, see SPEC_SYNC.md.
// This command is the single canonical writer of spec/openapi.yaml and
// spec/version.yaml.
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
	"net/url"
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

	npmVersion, err := npmVersionFor(commit)
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
	// This media type makes the API answer with the bare sha as plain text.
	body, err := githubAPI("commits/"+ref, "application/vnd.github.sha")
	if err != nil {
		return "", err
	}
	sha := strings.TrimSpace(string(body))
	if len(sha) != 40 {
		return "", fmt.Errorf("unexpected sha %q", sha)
	}
	return sha, nil
}

// npmVersionFor returns the version of the npm package that ships the spec as
// it is at commit. api-spec commits a new spec first and runs `npm version
// minor` in a separate commit afterwards, so package.json at a spec commit
// still holds the previous number. The version that ships a spec is therefore
// the first package.json change at or after the last spec change.
//
// With -spec-url there is no commit; the version then comes from main, which
// is only right when the URL serves main's spec.
func npmVersionFor(commit string) (string, error) {
	if commit == "" {
		return packageVersion("main")
	}

	specChange, err := lastChange(commit, specRepoSrc)
	if err != nil {
		return "", err
	}
	pkgChange, err := lastChange(commit, specRepoPkg)
	if err != nil {
		return "", err
	}
	after, err := isAtOrAfter(pkgChange, specChange)
	if err != nil {
		return "", err
	}
	if after {
		return packageVersion(commit)
	}

	// The bump for this spec comes later. Walk main's package.json changes,
	// newest first, back to the oldest one that still follows the spec change.
	changes, err := changes("main", specRepoPkg)
	if err != nil {
		return "", err
	}
	bump, reached := "", false
	for _, c := range changes {
		after, err := isAtOrAfter(c, specChange)
		if err != nil {
			return "", err
		}
		if !after {
			reached = true
			break
		}
		bump = c
	}
	if bump == "" {
		return "", fmt.Errorf("the spec at %s has no npm release yet", commit[:7])
	}
	if !reached && len(changes) == 100 {
		return "", fmt.Errorf("more than 100 npm releases since %s; pass a newer -ref", commit[:7])
	}
	return packageVersion(bump)
}

// lastChange returns the newest commit at or before ref that touched path.
func lastChange(ref, path string) (string, error) {
	list, err := commitList(ref, path, 1)
	if err != nil {
		return "", err
	}
	if len(list) == 0 {
		return "", fmt.Errorf("no commit at or before %s touches %s", ref, path)
	}
	return list[0], nil
}

// changes returns the commits at or before ref that touched path, newest first.
func changes(ref, path string) ([]string, error) {
	return commitList(ref, path, 100)
}

func commitList(ref, path string, perPage int) ([]string, error) {
	q := url.Values{"sha": {ref}, "path": {path}, "per_page": {fmt.Sprint(perPage)}}
	body, err := githubAPI("commits?"+q.Encode(), "application/vnd.github+json")
	if err != nil {
		return nil, err
	}
	var list []struct {
		SHA string `json:"sha"`
	}
	if err := json.Unmarshal(body, &list); err != nil {
		return nil, err
	}
	shas := make([]string, len(list))
	for i, c := range list {
		shas[i] = c.SHA
	}
	return shas, nil
}

// isAtOrAfter reports whether commit a is b or a descendant of b. Commit dates
// cannot decide this: the spec commit and the version bump are made by one
// workflow run, often within the same second.
func isAtOrAfter(a, b string) (bool, error) {
	if a == b {
		return true, nil
	}
	body, err := githubAPI("compare/"+b+"..."+a, "application/vnd.github+json")
	if err != nil {
		return false, err
	}
	var cmp struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(body, &cmp); err != nil {
		return false, err
	}
	return cmp.Status == "ahead" || cmp.Status == "identical", nil
}

func githubAPI(path, accept string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://api.github.com/repos/%s/%s", specRepo, path), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", accept)
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return do(req)
}

func packageVersion(ref string) (string, error) {
	body, err := fetch(rawURL(ref, specRepoPkg))
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
