package opusdns

// Drift checks between this hand-written client and the vendored OpenAPI
// specification in spec/. They run offline as part of `go test ./...`.
//
// TestSpecCoverage answers "which API operations does the client bind, and is
// every operation in the spec accounted for?". Every operation must appear in
// spec/coverage.yaml as implemented, deferred or excluded, so a newly published
// endpoint fails the build until somebody makes a decision about it.
//
// TestSpecOperations answers "does every implemented method send and read what
// its operation describes?": the request body, the query parameters and the
// response body, compared field by field and kind by kind.
//
// TestSpecModels compares the types listed under `models:` in the manifest with
// the schemas they mirror, and enforces `models.strict`.
//
// See SPEC_SYNC.md for the workflow around these tests.

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

const (
	specPath     = "../spec/openapi.yaml"
	specVersion  = "../spec/version.yaml"
	coveragePath = "../spec/coverage.yaml"

	statusImplemented = "implemented"
	statusDeferred    = "deferred"
	statusExcluded    = "excluded"

	// ignoreDirective opts a service method out of the route check, for the
	// rare method whose path cannot be read off a single BuildPath call. Write
	// it as a Go directive, on its own line with no space after the slashes, so
	// it stays out of the rendered documentation.
	ignoreDirective = "//speccheck:ignore"
)

// ---------------------------------------------------------------------------
// Vendored specification
// ---------------------------------------------------------------------------

type specDoc struct {
	Info struct {
		Version string `yaml:"version"`
	} `yaml:"info"`
	Paths      map[string]specPathItem `yaml:"paths"`
	Components struct {
		Schemas    map[string]*specSchema    `yaml:"schemas"`
		Parameters map[string]*specParameter `yaml:"parameters"`
	} `yaml:"components"`
}

type specPathItem struct {
	Parameters []*specParameter `yaml:"parameters"`
	Get        *specOperation   `yaml:"get"`
	Post       *specOperation   `yaml:"post"`
	Put        *specOperation   `yaml:"put"`
	Patch      *specOperation   `yaml:"patch"`
	Delete     *specOperation   `yaml:"delete"`
}

type specOperation struct {
	OperationID string                  `yaml:"operationId"`
	Tags        []string                `yaml:"tags"`
	Deprecated  bool                    `yaml:"deprecated"`
	Parameters  []*specParameter        `yaml:"parameters"`
	RequestBody *specRequestBody        `yaml:"requestBody"`
	Responses   map[string]specResponse `yaml:"responses"`

	// pathParameters are the parameters declared on the path item, which
	// apply to every operation on it.
	pathParameters []*specParameter
}

type specParameter struct {
	Ref      string      `yaml:"$ref"`
	Name     string      `yaml:"name"`
	In       string      `yaml:"in"`
	Required bool        `yaml:"required"`
	Schema   *specSchema `yaml:"schema"`
}

type specRequestBody struct {
	Required bool                 `yaml:"required"`
	Content  map[string]specMedia `yaml:"content"`
}

type specResponse struct {
	Content map[string]specMedia `yaml:"content"`
}

type specMedia struct {
	Schema *specSchema `yaml:"schema"`
}

const jsonMediaType = "application/json"

func (p specPathItem) byVerb() map[string]*specOperation {
	all := map[string]*specOperation{
		"GET": p.Get, "POST": p.Post, "PUT": p.Put, "PATCH": p.Patch, "DELETE": p.Delete,
	}
	out := make(map[string]*specOperation, len(all))
	for verb, op := range all {
		if op != nil {
			out[verb] = op
		}
	}
	return out
}

// operations keys every operation as "METHOD /path", using the path exactly as
// the spec spells it.
func (d *specDoc) operations() map[string]specOperation {
	ops := make(map[string]specOperation, len(d.Paths))
	for path, item := range d.Paths {
		for verb, op := range item.byVerb() {
			withPath := *op
			withPath.pathParameters = item.Parameters
			ops[verb+" "+path] = withPath
		}
	}
	return ops
}

// queryParameters returns the operation's query parameters by name, with
// $ref resolved against components.parameters.
func (d *specDoc) queryParameters(op specOperation) map[string]*specParameter {
	out := map[string]*specParameter{}
	for _, param := range append(append([]*specParameter{}, op.pathParameters...), op.Parameters...) {
		if param != nil && param.Ref != "" {
			param = d.Components.Parameters[strings.TrimPrefix(param.Ref, "#/components/parameters/")]
		}
		if param != nil && param.In == "query" {
			out[param.Name] = param
		}
	}
	return out
}

// requestSchema returns the JSON request body schema, if the operation takes one.
func (op specOperation) requestSchema() *specSchema {
	if op.RequestBody == nil {
		return nil
	}
	return op.RequestBody.Content[jsonMediaType].Schema
}

// responseSchema returns the JSON bodies of the operation's 2xx responses, as a
// union when there is more than one (renew answers 200 or 202 with different
// bodies), and the status codes they belong to.
func (op specOperation) responseSchema() (*specSchema, string) {
	codes := make([]string, 0, len(op.Responses))
	for code := range op.Responses {
		codes = append(codes, code)
	}
	sort.Strings(codes)
	var (
		schemas []*specSchema
		with    []string
	)
	for _, code := range codes {
		if !strings.HasPrefix(code, "2") {
			continue
		}
		if media, ok := op.Responses[code].Content[jsonMediaType]; ok && media.Schema != nil {
			schemas = append(schemas, media.Schema)
			with = append(with, code)
		}
	}
	return oneSchema(schemas), strings.Join(with, "/")
}

func loadSpec(t *testing.T) *specDoc {
	t.Helper()

	raw, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("read vendored spec: %v (run `make spec-sync`)", err)
	}
	var doc specDoc
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("parse %s: %v", specPath, err)
	}
	if len(doc.Paths) == 0 {
		t.Fatalf("%s has no paths", specPath)
	}
	return &doc
}

type versionStamp struct {
	NpmVersion    string `yaml:"npm_version"`
	InfoVersion   string `yaml:"info_version"`
	APISpecCommit string `yaml:"api_spec_commit"`
}

func loadVersionStamp(t *testing.T) versionStamp {
	t.Helper()

	raw, err := os.ReadFile(specVersion)
	if err != nil {
		t.Fatalf("read %s: %v (run `make spec-sync`)", specVersion, err)
	}
	var v versionStamp
	if err := yaml.Unmarshal(raw, &v); err != nil {
		t.Fatalf("parse %s: %v", specVersion, err)
	}
	return v
}

// ---------------------------------------------------------------------------
// Coverage manifest
// ---------------------------------------------------------------------------

type coverageDoc struct {
	Version    int                      `yaml:"version"`
	Operations map[string]coverageEntry `yaml:"operations"`
	Models     coverageModels           `yaml:"models"`
}

type coverageEntry struct {
	Status string     `yaml:"status"`
	Method methodList `yaml:"method"`
	Reason string     `yaml:"reason"`
	Since  string     `yaml:"since"`
}

type coverageModels struct {
	// Mappings is Go type name (package models) -> spec schema name.
	Mappings map[string]string `yaml:"mappings"`
	// Requests is Go request type name (package models) -> the request
	// schema it encodes.
	Requests map[string]string `yaml:"requests"`
	// Strict lists Go types for which a property present in the spec but
	// missing from the struct is an error rather than a warning.
	Strict []string `yaml:"strict"`
}

// methodList accepts either a single "Service.Method" string or a list of them.
// The first entry is the method that builds the path; further entries are
// convenience wrappers over it.
type methodList []string

func (m *methodList) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.ScalarNode:
		var one string
		if err := node.Decode(&one); err != nil {
			return err
		}
		*m = methodList{one}
		return nil
	case yaml.SequenceNode:
		var many []string
		if err := node.Decode(&many); err != nil {
			return err
		}
		*m = many
		return nil
	default:
		return fmt.Errorf("line %d: expected a string or a list of strings", node.Line)
	}
}

var (
	operationKeyRE = regexp.MustCompile(`^(GET|POST|PUT|PATCH|DELETE) /v1/`)
	methodRefRE    = regexp.MustCompile(`^[A-Z][A-Za-z0-9]*\.[A-Z][A-Za-z0-9]*$`)
	pathParamRE    = regexp.MustCompile(`\{[^}]*\}`)
)

func loadCoverage(t *testing.T) *coverageDoc {
	t.Helper()

	raw, err := os.ReadFile(coveragePath)
	if err != nil {
		t.Fatalf("read %s: %v", coveragePath, err)
	}
	var doc coverageDoc
	// KnownFields catches a typo in a status or method key instead of
	// silently treating the entry as empty.
	dec := yaml.NewDecoder(strings.NewReader(string(raw)))
	dec.KnownFields(true)
	if err := dec.Decode(&doc); err != nil {
		t.Fatalf("parse %s: %v", coveragePath, err)
	}

	for key, entry := range doc.Operations {
		where := fmt.Sprintf("%s: %q", coveragePath, key)
		if !operationKeyRE.MatchString(key) {
			t.Errorf(`%s: key must read "METHOD /v1/..." with an upper-case method`, where)
		}
		switch entry.Status {
		case statusImplemented:
			if len(entry.Method) == 0 {
				t.Errorf("%s: status %s requires a `method:` entry", where, statusImplemented)
			}
			for _, ref := range entry.Method {
				if !methodRefRE.MatchString(ref) {
					t.Errorf("%s: method %q must read Service.Method", where, ref)
				}
			}
		case statusDeferred, statusExcluded:
			if strings.TrimSpace(entry.Reason) == "" {
				t.Errorf("%s: status %s requires a `reason:`", where, entry.Status)
			}
			if len(entry.Method) > 0 {
				t.Errorf("%s: status %s must not name a method", where, entry.Status)
			}
		default:
			t.Errorf("%s: unknown status %q (want %s, %s or %s)",
				where, entry.Status, statusImplemented, statusDeferred, statusExcluded)
		}
	}
	return &doc
}

// normalizeRoute reduces path parameter names to a placeholder, so that the
// client's own naming does not have to match the spec's.
func normalizeRoute(route string) string {
	return pathParamRE.ReplaceAllString(route, "{}")
}

// ---------------------------------------------------------------------------
// Routes the client actually builds
// ---------------------------------------------------------------------------

type clientRoute struct {
	name  string // "Domains.RenewDomain"
	route string // "POST /v1/domains/{}/renew"
	pos   string // "service_domains.go:214"
}

var httpMethodConsts = map[string]string{
	"MethodGet":    "GET",
	"MethodPost":   "POST",
	"MethodPut":    "PUT",
	"MethodPatch":  "PATCH",
	"MethodDelete": "DELETE",
}

// extractClientRoutes reads every exported service method in this package and
// derives the route it calls, from its single BuildPath call and its single
// request call. Methods that build no path (pagination wrappers, helpers) are
// skipped.
func extractClientRoutes(t *testing.T) (map[string]clientRoute, map[string]bool) {
	t.Helper()

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, ".", func(fi os.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse package: %v", err)
	}

	routes := make(map[string]clientRoute)
	ignored := make(map[string]bool)
	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			for _, decl := range file.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok || fn.Body == nil {
					continue
				}
				service, recvVar, ok := serviceReceiver(fn)
				if !ok || !fn.Name.IsExported() {
					continue
				}
				if hasIgnoreDirective(fn.Doc) {
					ignored[service+"."+fn.Name.Name] = true
					continue
				}

				name := service + "." + fn.Name.Name
				pos := shortPos(fset, fn.Pos())
				route, err := routeOf(fn, recvVar)
				if err != nil {
					t.Errorf("%s (%s): %v", name, pos, err)
					continue
				}
				if route == "" {
					continue // builds no path of its own
				}
				routes[name] = clientRoute{name: name, route: route, pos: pos}
			}
		}
	}
	if len(routes) == 0 {
		t.Fatalf("no client routes found; the BuildPath convention may have changed")
	}
	return routes, ignored
}

// serviceReceiver reports the service name and receiver variable of a method
// declared on a *XxxService type.
func serviceReceiver(fn *ast.FuncDecl) (service, recvVar string, ok bool) {
	if fn.Recv == nil || len(fn.Recv.List) != 1 {
		return "", "", false
	}
	field := fn.Recv.List[0]
	star, ok := field.Type.(*ast.StarExpr)
	if !ok {
		return "", "", false
	}
	ident, ok := star.X.(*ast.Ident)
	if !ok || !strings.HasSuffix(ident.Name, "Service") {
		return "", "", false
	}
	if len(field.Names) != 1 {
		return "", "", false
	}
	return strings.TrimSuffix(ident.Name, "Service"), field.Names[0].Name, true
}

func routeOf(fn *ast.FuncDecl, recvVar string) (string, error) {
	httpExpr := recvVar + ".client.http"

	var (
		segments   []string
		buildPaths int
		verb       string
		verbCalls  int
	)
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || types.ExprString(sel.X) != httpExpr {
			return true
		}
		switch sel.Sel.Name {
		case "BuildPath":
			buildPaths++
			segments = pathSegments(call.Args)
		case "Get", "Post", "Put", "Patch", "Delete":
			verbCalls++
			verb = strings.ToUpper(sel.Sel.Name)
		case "Do":
			// The low-level escape hatch: the verb sits in the Request literal.
			if v := verbFromRequest(call.Args); v != "" {
				verbCalls++
				verb = v
			}
		}
		return true
	})

	switch {
	case buildPaths == 0 && verbCalls == 0:
		return "", nil
	case buildPaths > 1:
		return "", fmt.Errorf("calls BuildPath %d times; split the method or add a `%s` directive", buildPaths, ignoreDirective)
	case verbCalls > 1:
		return "", fmt.Errorf("issues %d requests; split the method or add a `%s` directive", verbCalls, ignoreDirective)
	case buildPaths == 0:
		return "", fmt.Errorf("issues a request without calling BuildPath; add a `%s` directive if that is intended", ignoreDirective)
	case verbCalls == 0:
		return "", fmt.Errorf("builds a path but issues no request; add a `%s` directive if that is intended", ignoreDirective)
	}

	return verb + " /" + DefaultAPIVersion + "/" + strings.Join(segments, "/"), nil
}

// pathSegments keeps string literals verbatim and reduces everything else
// (identifiers, url.PathEscape calls, conversions) to a path parameter.
func pathSegments(args []ast.Expr) []string {
	segments := make([]string, 0, len(args))
	for _, arg := range args {
		lit, ok := arg.(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			segments = append(segments, "{}")
			continue
		}
		value, err := strconv.Unquote(lit.Value)
		if err != nil {
			segments = append(segments, "{}")
			continue
		}
		segments = append(segments, value)
	}
	return segments
}

// verbFromRequest reads Method out of a `&Request{Method: http.MethodPost, ...}`
// argument.
func verbFromRequest(args []ast.Expr) string {
	for _, arg := range args {
		if unary, ok := arg.(*ast.UnaryExpr); ok {
			arg = unary.X
		}
		lit, ok := arg.(*ast.CompositeLit)
		if !ok {
			continue
		}
		for _, elt := range lit.Elts {
			kv, ok := elt.(*ast.KeyValueExpr)
			if !ok {
				continue
			}
			if key, ok := kv.Key.(*ast.Ident); !ok || key.Name != "Method" {
				continue
			}
			switch value := kv.Value.(type) {
			case *ast.SelectorExpr:
				if verb, ok := httpMethodConsts[value.Sel.Name]; ok {
					return verb
				}
			case *ast.BasicLit:
				if verb, err := strconv.Unquote(value.Value); err == nil {
					return strings.ToUpper(verb)
				}
			}
		}
	}
	return ""
}

func hasIgnoreDirective(doc *ast.CommentGroup) bool {
	if doc == nil {
		return false
	}
	// Not doc.Text(): that strips directives, which is exactly what this is.
	for _, comment := range doc.List {
		if strings.HasPrefix(comment.Text, ignoreDirective) {
			return true
		}
	}
	return false
}

func shortPos(fset *token.FileSet, pos token.Pos) string {
	p := fset.Position(pos)
	return fmt.Sprintf("%s:%d", p.Filename, p.Line)
}

// ---------------------------------------------------------------------------
// TestSpecCoverage
// ---------------------------------------------------------------------------

func TestSpecCoverage(t *testing.T) {
	// The route prefix is baked into the comparison, so a changed default
	// would silently invalidate every route.
	if got := NewConfig().APIVersion; got != DefaultAPIVersion {
		t.Fatalf("default APIVersion is %q, want %q", got, DefaultAPIVersion)
	}

	spec := loadSpec(t)
	stamp := loadVersionStamp(t)
	coverage := loadCoverage(t)
	routes, ignoredRoutes := extractClientRoutes(t)

	if stamp.InfoVersion != spec.Info.Version {
		t.Errorf("%s records info_version %q but %s says %q; re-run `make spec-sync`",
			specVersion, stamp.InfoVersion, specPath, spec.Info.Version)
	}

	specOps := spec.operations()

	// Index the client's routes by normalized route so unmapped operations can
	// be matched against code that already exists.
	byRoute := make(map[string][]clientRoute, len(routes))
	for _, route := range routes {
		key := normalizeRoute(route.route)
		byRoute[key] = append(byRoute[key], route)
	}

	// (a) Operations in the spec that nobody has triaged.
	var unmapped []string
	for key := range specOps {
		if _, ok := coverage.Operations[key]; !ok {
			unmapped = append(unmapped, key)
		}
	}
	sort.Strings(unmapped)
	for _, key := range unmapped {
		op := specOps[key]
		t.Errorf("UNMAPPED  %s  (tag=%s, operationId=%s)", key, strings.Join(op.Tags, ","), op.OperationID)
	}

	// (b) Manifest entries whose operation no longer exists upstream.
	var stale []string
	for key := range coverage.Operations {
		if _, ok := specOps[key]; !ok {
			stale = append(stale, key)
		}
	}
	sort.Strings(stale)
	for _, key := range stale {
		entry := coverage.Operations[key]
		t.Errorf("STALE     %s  (status=%s%s) is no longer in the spec; drop the entry and any code behind it",
			key, entry.Status, methodSuffix(entry.Method))
	}

	// (c) Implemented entries must name methods that exist and build that route.
	claimed := make(map[string]string, len(coverage.Operations))
	implemented := 0
	for _, key := range sortedKeys(coverage.Operations) {
		entry := coverage.Operations[key]
		if entry.Status != statusImplemented {
			continue
		}
		implemented++

		if len(entry.Method) == 0 {
			// loadCoverage already reported the missing `method:`; the checks
			// below all need an owning method, so skip rather than panic.
			continue
		}

		if op, ok := specOps[key]; ok && op.Deprecated {
			t.Logf("NOTE      %s is deprecated upstream but still implemented (%s)", key, entry.Method[0])
		}
		for _, ref := range entry.Method {
			if err := assertServiceMethod(ref); err != nil {
				t.Errorf("%s: %v", key, err)
			}
		}

		// Every method on the entry is claimed by it. An operation is served
		// by more than one method when its request body is discriminated and
		// the client splits it into one method per variant.
		owner := entry.Method[0]
		for _, ref := range entry.Method {
			claimed[ref] = key
		}
		if ignoredRoutes[owner] {
			// The method opted out of the route check, typically because it
			// serves several spec paths that differ only in a path segment.
			continue
		}
		if _, ok := routes[owner]; !ok {
			// A wrapper may legitimately own no BuildPath, but then some
			// method must; point at the mismatch rather than guessing.
			t.Errorf("MISSING   %s is mapped to %s, which builds no route of its own; "+
				"name the method that calls BuildPath first", key, owner)
			continue
		}
		for _, ref := range entry.Method {
			if ignoredRoutes[ref] {
				continue
			}
			route, ok := routes[ref]
			if !ok {
				continue // a wrapper that delegates to another method
			}
			if got, want := normalizeRoute(route.route), normalizeRoute(key); got != want {
				t.Errorf("MISMATCH  %s builds %q but coverage.yaml maps it to %q (%s)",
					ref, got, want, route.pos)
			}
		}
	}

	// (d) Routes the client builds that the spec does not describe.
	for _, name := range sortedRouteNames(routes) {
		route := routes[name]
		if _, ok := claimed[name]; ok {
			continue
		}
		if hasSpecRoute(specOps, route.route) {
			t.Errorf("UNCLAIMED %s builds %q, which the spec has but coverage.yaml maps to no method (%s)",
				name, route.route, route.pos)
			continue
		}
		t.Errorf("ORPHAN    %s builds %q, which is not in the spec (%s)", name, route.route, route.pos)
	}

	// (e) A method that opted out of the route check is still bound to the
	// operations it serves through the manifest; only the route comparison is
	// skipped. Without an entry nothing would check it at all.
	for _, name := range sortedBoolKeys(ignoredRoutes) {
		if _, ok := claimed[name]; !ok {
			t.Errorf("UNCLAIMED %s opts out of the route check with %s, but no implemented coverage.yaml entry names it; "+
				"add it under `method:` of every operation it serves", name, ignoreDirective)
		}
	}

	if len(unmapped) > 0 {
		writeStubs(t, unmapped, specOps, byRoute)
	}

	deferredCount, excludedCount := 0, 0
	for _, entry := range coverage.Operations {
		switch entry.Status {
		case statusDeferred:
			deferredCount++
		case statusExcluded:
			excludedCount++
		}
	}
	t.Logf("spec %s (npm %s): %d operations; %d implemented, %d deferred, %d excluded, %d unmapped, %d stale",
		spec.Info.Version, stamp.NpmVersion, len(specOps),
		implemented, deferredCount, excludedCount, len(unmapped), len(stale))
}

// writeStubs prints ready-to-paste manifest entries, pre-filled with the method
// that already implements the operation where one exists.
//
// It writes to stdout rather than through t.Log on purpose: the test log
// indents every line, which would corrupt the YAML for whoever pastes it. The
// markers let `make spec-stubs` cut the block out verbatim.
func writeStubs(t *testing.T, unmapped []string, specOps map[string]specOperation, byRoute map[string][]clientRoute) {
	t.Helper()

	var b strings.Builder
	b.WriteString("\n--- STUBS (paste under `operations:` in " + coveragePath + ")\n")
	for _, key := range unmapped {
		fmt.Fprintf(&b, "  %q:\n", key)
		if matches := byRoute[normalizeRoute(key)]; len(matches) == 1 {
			fmt.Fprintf(&b, "    status: %s\n", statusImplemented)
			fmt.Fprintf(&b, "    method: %s\n", matches[0].name)
			continue
		}
		fmt.Fprintf(&b, "    status: %s\n", statusDeferred)
		fmt.Fprintf(&b, "    reason: TODO  # %s\n", strings.Join(specOps[key].Tags, ","))
	}
	b.WriteString("--- END STUBS\n")
	fmt.Print(b.String())
}

func assertServiceMethod(ref string) error {
	parts := strings.SplitN(ref, ".", 2)
	if len(parts) != 2 {
		return fmt.Errorf("method %q must read Service.Method", ref)
	}
	field, ok := reflect.TypeOf(Client{}).FieldByName(parts[0])
	if !ok {
		return fmt.Errorf("Client has no service field %q", parts[0])
	}
	if _, ok := field.Type.MethodByName(parts[1]); !ok {
		return fmt.Errorf("%s has no method %s", field.Type, parts[1])
	}
	return nil
}

func hasSpecRoute(specOps map[string]specOperation, route string) bool {
	want := normalizeRoute(route)
	for key := range specOps {
		if normalizeRoute(key) == want {
			return true
		}
	}
	return false
}

func methodSuffix(methods methodList) string {
	if len(methods) == 0 {
		return ""
	}
	return ", method=" + strings.Join(methods, ",")
}

func sortedKeys(m map[string]coverageEntry) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func sortedRouteNames(m map[string]clientRoute) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// ---------------------------------------------------------------------------
// TestSpecModels
// ---------------------------------------------------------------------------

// TestSpecModels checks the types named under `models.mappings`. Most types are
// already reached through TestSpecOperations; a mapping pins a type to its
// schema explicitly, and is what `models.strict` hangs off.
func TestSpecModels(t *testing.T) {
	spec := loadSpec(t)
	coverage := loadCoverage(t)
	pkgs := loadGoPackages(t)
	index := &schemaIndex{schemas: spec.Components.Schemas}
	models := manifestModels(t, coverage, pkgs)

	decode := newShapeChecker(index, decoding, models, "")
	encode := newShapeChecker(index, encoding, models, "")
	for _, section := range []struct {
		name     string
		mappings map[string]string
		checker  *shapeChecker
	}{
		{"models.mappings", coverage.Models.Mappings, decode},
		{"models.requests", coverage.Models.Requests, encode},
	} {
		for _, goName := range sortedStrings(section.mappings) {
			schemaName := section.mappings[goName]
			goType, ok := pkgs.lookupModel(goName)
			if !ok {
				t.Errorf("%s: %s names %q, which is not a type in package models", coveragePath, section.name, goName)
				continue
			}
			if _, ok := spec.Components.Schemas[schemaName]; !ok {
				t.Errorf("%s: schema %q (mapped from models.%s) is not in the spec", coveragePath, schemaName, goName)
				continue
			}
			section.checker.where = "models." + goName + " ~ " + schemaName
			section.checker.check(goType, &specSchema{Ref: componentSchemaPrefix + schemaName}, "models."+goName)
		}
	}
	report(t, decode)
	report(t, encode)
}

// manifestModels reads models.mappings and models.strict, keyed by qualified
// Go type name.
func manifestModels(t *testing.T, coverage *coverageDoc, pkgs *goPackages) modelSchemas {
	t.Helper()
	models := modelSchemas{
		canonical:  make(map[string]string, len(coverage.Models.Mappings)),
		strict:     make(map[string]bool, len(coverage.Models.Strict)),
		deprecated: pkgs.deprecated,
	}
	for goName, schemaName := range coverage.Models.Mappings {
		models.canonical["models."+goName] = schemaName
	}
	for _, name := range coverage.Models.Strict {
		if _, ok := coverage.Models.Mappings[name]; !ok {
			t.Errorf("%s: models.strict lists %q, which has no mapping", coveragePath, name)
		}
		models.strict["models."+name] = true
	}
	return models
}

func report(t *testing.T, c *shapeChecker) {
	t.Helper()
	for _, message := range c.errors {
		t.Error(message)
	}
	for _, message := range c.warns {
		t.Log(message)
	}
}

// ---------------------------------------------------------------------------
// TestSpecOperations
// ---------------------------------------------------------------------------

// TestSpecOperations checks every method an implemented manifest entry names
// against its operation: the request body it sends, the query parameters it
// sets and the type it decodes the response into.
func TestSpecOperations(t *testing.T) {
	spec := loadSpec(t)
	coverage := loadCoverage(t)
	pkgs := loadGoPackages(t)
	index := &schemaIndex{schemas: spec.Components.Schemas}
	models := manifestModels(t, coverage, pkgs)
	specOps := spec.operations()

	decode := newShapeChecker(index, decoding, models, "")
	encode := newShapeChecker(index, encoding, models, "")
	checked := 0
	for _, key := range sortedKeys(coverage.Operations) {
		entry := coverage.Operations[key]
		op, ok := specOps[key]
		if entry.Status != statusImplemented || !ok {
			continue // TestSpecCoverage reports stale and malformed entries
		}
		for _, ref := range entry.Method {
			fn := pkgs.methods[ref]
			if fn == nil {
				continue // TestSpecCoverage reports methods that do not exist
			}
			call := pkgs.requestOf(fn)
			if call == nil {
				continue // a wrapper; the method that issues the request is checked
			}
			checked++
			where := key + " " + ref
			checkRequestBody(t, encode, pkgs, op, call, where)
			checkQuery(t, spec, op, call, where)
			checkResponse(t, decode, op, call, where)
		}
	}
	report(t, decode)
	report(t, encode)
	if checked == 0 {
		t.Fatalf("no service method issues a request; the request helpers may have been renamed")
	}
	t.Logf("checked %d methods against the request and response of their operations", checked)
}

// methodCall is what one service method sends and reads, taken from its body.
type methodCall struct {
	verb      string
	body      ast.Expr // nil when the request carries no body
	bodyKnown bool     // false when the body could not be read off the call
	queryKeys map[string]bool
	decodes   []types.Type
}

// requestOf reads the request a service method issues, or returns nil when it
// issues none itself.
func (p *goPackages) requestOf(fn *ast.FuncDecl) *methodCall {
	_, recvVar, ok := serviceReceiver(fn)
	if !ok {
		return nil
	}
	httpExpr := recvVar + ".client.http"

	call := &methodCall{bodyKnown: true}
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		expr, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := expr.Fun.(*ast.SelectorExpr)
		if !ok || types.ExprString(sel.X) != httpExpr {
			return true
		}
		switch sel.Sel.Name {
		case "Get", "Delete":
			call.verb = strings.ToUpper(sel.Sel.Name)
		case "Post", "Put", "Patch":
			call.verb = strings.ToUpper(sel.Sel.Name)
			if len(expr.Args) == 3 {
				call.body = expr.Args[2]
			}
		case "Do":
			call.verb = verbFromRequest(expr.Args)
			call.body, call.bodyKnown = bodyFromRequest(expr.Args)
		case "DecodeResponse":
			if len(expr.Args) == 2 && !isNilLiteral(expr.Args[1]) {
				if target := p.info.TypeOf(expr.Args[1]); target != nil {
					call.decodes = append(call.decodes, derefType(target))
				}
			}
		}
		return true
	})
	if call.verb == "" {
		return nil
	}
	call.queryKeys = p.queryKeys(fn)
	return call
}

// bodyFromRequest reads Body out of a `&Request{...}` argument.
func bodyFromRequest(args []ast.Expr) (ast.Expr, bool) {
	for _, arg := range args {
		if unary, ok := arg.(*ast.UnaryExpr); ok {
			arg = unary.X
		}
		lit, ok := arg.(*ast.CompositeLit)
		if !ok {
			continue
		}
		for _, elt := range lit.Elts {
			kv, ok := elt.(*ast.KeyValueExpr)
			if !ok {
				continue
			}
			if key, ok := kv.Key.(*ast.Ident); ok && key.Name == "Body" {
				return kv.Value, true
			}
		}
		return nil, true
	}
	return nil, false
}

// queryKeys collects the literal keys a method sets on url.Values, in its own
// body and in the package functions it calls (query builders such as
// objectLogsQuery), but not inside the HTTP client.
//
// A builder shared by two operations may set a key only for one of them,
// behind a bool parameter (usageQuery's withGranularity). When the method
// passes the constant false for such a parameter, the branches it guards are
// skipped.
func (p *goPackages) queryKeys(fn *ast.FuncDecl) map[string]bool {
	keys := map[string]bool{}
	visited := map[*ast.FuncDecl]bool{}

	var walk func(node ast.Node, falseParams map[types.Object]bool)
	visit := func(decl *ast.FuncDecl, call *ast.CallExpr) {
		if visited[decl] {
			return
		}
		visited[decl] = true
		falseParams := map[types.Object]bool{}
		if call != nil {
			i := 0
			for _, field := range decl.Type.Params.List {
				for _, name := range field.Names {
					if i < len(call.Args) && isConstFalse(p.info, call.Args[i]) {
						falseParams[p.info.Defs[name]] = true
					}
					i++
				}
			}
		}
		walk(decl.Body, falseParams)
	}
	walk = func(node ast.Node, falseParams map[types.Object]bool) {
		ast.Inspect(node, func(n ast.Node) bool {
			switch n := n.(type) {
			case *ast.IfStmt:
				if isFalseUnder(p.info, n.Cond, falseParams) {
					if n.Else != nil {
						walk(n.Else, falseParams)
					}
					return false
				}
			case *ast.CallExpr:
				var ident *ast.Ident
				switch fun := n.Fun.(type) {
				case *ast.Ident:
					ident = fun
				case *ast.SelectorExpr:
					ident = fun.Sel
					if (ident.Name == "Set" || ident.Name == "Add") && len(n.Args) == 2 && isURLValues(p.info.TypeOf(fun.X)) {
						if lit, ok := n.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
							if key, err := strconv.Unquote(lit.Value); err == nil {
								keys[key] = true
							}
						}
					}
				}
				if ident == nil {
					return true
				}
				callee, ok := p.info.Uses[ident].(*types.Func)
				if !ok || isHTTPClientMethod(callee) {
					return true
				}
				if decl := p.decls[callee]; decl != nil {
					visit(decl, n)
				}
			}
			return true
		})
	}
	visit(fn, nil)
	return keys
}

func isConstFalse(info *types.Info, expr ast.Expr) bool {
	tv, ok := info.Types[expr]
	return ok && tv.Value != nil && tv.Value.String() == "false"
}

// isFalseUnder reports whether cond is false whenever the parameters in
// falseParams are false.
func isFalseUnder(info *types.Info, cond ast.Expr, falseParams map[types.Object]bool) bool {
	switch cond := cond.(type) {
	case *ast.ParenExpr:
		return isFalseUnder(info, cond.X, falseParams)
	case *ast.Ident:
		return falseParams[info.Uses[cond]] || isConstFalse(info, cond)
	case *ast.BinaryExpr:
		return cond.Op == token.LAND && (isFalseUnder(info, cond.X, falseParams) || isFalseUnder(info, cond.Y, falseParams))
	}
	return isConstFalse(info, cond)
}

func isURLValues(t types.Type) bool {
	return t != nil && types.TypeString(derefType(t), nil) == "net/url.Values"
}

func isHTTPClientMethod(fn *types.Func) bool {
	sig, ok := fn.Type().(*types.Signature)
	if !ok || sig.Recv() == nil {
		return false
	}
	named, ok := derefType(sig.Recv().Type()).(*types.Named)
	return ok && named.Obj().Name() == "HTTPClient"
}

func isNilLiteral(expr ast.Expr) bool {
	ident, ok := expr.(*ast.Ident)
	return ok && ident.Name == "nil"
}

func checkRequestBody(t *testing.T, encode *shapeChecker, pkgs *goPackages, op specOperation, call *methodCall, where string) {
	t.Helper()
	if !call.bodyKnown {
		t.Logf("NOTE      %s: request body not read off the call; not checked", where)
		return
	}
	required := op.RequestBody != nil && op.RequestBody.Required
	if call.body == nil || isNilLiteral(call.body) {
		if required {
			t.Errorf("NOBODY    %s: the operation requires a request body, but the method sends none", where)
		}
		return
	}
	bodyType := pkgs.info.TypeOf(call.body)
	if bodyType == nil {
		return
	}
	if op.RequestBody == nil {
		t.Errorf("BODY      %s: sends a request body (%s), but the operation takes none", where, typeName(bodyType))
		return
	}
	schema := op.requestSchema()
	if schema == nil {
		return // not a JSON body
	}
	encode.where = where + " request"
	encode.check(bodyType, schema, typeName(derefType(bodyType)))
}

func checkQuery(t *testing.T, spec *specDoc, op specOperation, call *methodCall, where string) {
	t.Helper()
	params := spec.queryParameters(op)
	for _, key := range sortedBoolKeys(call.queryKeys) {
		if _, ok := params[key]; !ok {
			t.Errorf("QUERY     %s: sends query parameter %q, which the operation does not declare, so the API ignores it", where, key)
		}
	}
	names := make([]string, 0, len(params))
	for name := range params {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if params[name].Required && !call.queryKeys[name] {
			t.Errorf("QUERY     %s: the operation requires query parameter %q, which the method never sends", where, name)
		}
	}
}

func checkResponse(t *testing.T, decode *shapeChecker, op specOperation, call *methodCall, where string) {
	t.Helper()
	if len(call.decodes) == 0 {
		return // the method returns no body, or streams it
	}
	schema, code := op.responseSchema()
	if schema == nil {
		t.Errorf("RESPONSE  %s: decodes into %s, but the operation returns no JSON body", where, typeName(call.decodes[0]))
		return
	}
	if len(call.decodes) == 1 {
		decode.where = where + " response " + code
		decode.check(call.decodes[0], schema, typeName(call.decodes[0]))
		return
	}
	// Several targets, e.g. a fallback for another body shape: one of them has
	// to read the documented response.
	var failures []string
	for _, target := range call.decodes {
		trial := newShapeChecker(decode.index, decoding, decode.models, where+" response "+code)
		trial.check(target, schema, typeName(target))
		if len(trial.errors) == 0 {
			return
		}
		failures = append(failures, trial.errors...)
	}
	for _, failure := range failures {
		t.Error(failure)
	}
}

func sortedStrings(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func sortedBoolKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
