package opusdns

// Drift checks between this hand-written client and the vendored OpenAPI
// specification in spec/. They run offline as part of `go test ./...`.
//
// TestSpecCoverage answers "which API operations does the client bind, and is
// every operation in the spec accounted for?". Every operation must appear in
// spec/coverage.yaml as implemented, deferred or excluded, so a newly published
// endpoint fails the build until somebody makes a decision about it.
//
// TestSpecModels answers "do the hand-written response structs still match the
// schemas they were written against?".
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

	"github.com/opusdns/opusdns-go-client/models"
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

// modelRegistry names the response structs TestSpecModels can check. Go cannot
// look up a type by name at run time, so every type referenced from the
// `models.mappings` block of spec/coverage.yaml has to be listed here.
var modelRegistry = map[string]interface{}{
	"Contact":                models.Contact{},
	"ContactAttributeSet":    models.ContactAttributeSet{},
	"Domain":                 models.Domain{},
	"DomainForward":          models.DomainForward{},
	"DomainSummary":          models.DomainSummary{},
	"EmailForward":           models.EmailForward{},
	"Event":                  models.Event{},
	"Host":                   models.Host{},
	"IPRestriction":          models.IPRestriction{},
	"Invoice":                models.Invoice{},
	"JobBatchRetryResponse":  models.JobBatchRetryResponse{},
	"JobBatchStatusResponse": models.JobBatchStatusResponse{},
	"JobResponse":            models.JobResponse{},
	"Organization":           models.Organization{},
	"Report":                 models.Report{},
	"Tag":                    models.Tag{},
	"TokenResponse":          models.TokenResponse{},
	"User":                   models.User{},
	"VanityNameserverSet":    models.VanityNameserverSet{},
	"Whitelabel":             models.Whitelabel{},
	"Zone":                   models.Zone{},
	"ZoneSummary":            models.ZoneSummary{},
}

// ---------------------------------------------------------------------------
// Vendored specification
// ---------------------------------------------------------------------------

type specDoc struct {
	Info struct {
		Version string `yaml:"version"`
	} `yaml:"info"`
	Paths      map[string]specPathItem `yaml:"paths"`
	Components struct {
		Schemas map[string]specSchema `yaml:"schemas"`
	} `yaml:"components"`
}

type specPathItem struct {
	Get    *specOperation `yaml:"get"`
	Post   *specOperation `yaml:"post"`
	Put    *specOperation `yaml:"put"`
	Patch  *specOperation `yaml:"patch"`
	Delete *specOperation `yaml:"delete"`
}

type specOperation struct {
	OperationID string   `yaml:"operationId"`
	Tags        []string `yaml:"tags"`
	Deprecated  bool     `yaml:"deprecated"`
}

type specSchema struct {
	Properties map[string]yaml.Node `yaml:"properties"`
	AllOf      []struct {
		Ref string `yaml:"$ref"`
	} `yaml:"allOf"`
}

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
			ops[verb+" "+path] = *op
		}
	}
	return ops
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

		owner := entry.Method[0]
		claimed[owner] = key
		if ignoredRoutes[owner] {
			// The method opted out of the route check, typically because it
			// serves several spec paths that differ only in a path segment.
			continue
		}
		route, ok := routes[owner]
		if !ok {
			// A wrapper may legitimately own no BuildPath, but then some
			// method must; point at the mismatch rather than guessing.
			t.Errorf("MISSING   %s is mapped to %s, which builds no route of its own; "+
				"name the method that calls BuildPath first", key, owner)
			continue
		}
		if got, want := normalizeRoute(route.route), normalizeRoute(key); got != want {
			t.Errorf("MISMATCH  %s builds %q but coverage.yaml maps it to %q (%s)",
				owner, got, want, route.pos)
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

func TestSpecModels(t *testing.T) {
	spec := loadSpec(t)
	coverage := loadCoverage(t)

	strict := make(map[string]bool, len(coverage.Models.Strict))
	for _, name := range coverage.Models.Strict {
		if _, ok := coverage.Models.Mappings[name]; !ok {
			t.Errorf("%s: models.strict lists %q, which has no mapping", coveragePath, name)
		}
		strict[name] = true
	}

	for _, goName := range sortedStrings(coverage.Models.Mappings) {
		schemaName := coverage.Models.Mappings[goName]

		sample, ok := modelRegistry[goName]
		if !ok {
			t.Errorf("%s: models.mappings names %q, which is not in modelRegistry (%s)",
				coveragePath, goName, "opusdns/spec_coverage_test.go")
			continue
		}
		properties, ok := schemaProperties(spec.Components.Schemas, schemaName, 0)
		if !ok {
			t.Errorf("%s: schema %q (mapped from models.%s) is not in the spec",
				coveragePath, schemaName, goName)
			continue
		}
		if len(properties) == 0 {
			t.Logf("WARN      schema %s has no properties (a union type?); skipping models.%s", schemaName, goName)
			continue
		}

		fields := jsonFields(reflect.TypeOf(sample))
		for _, field := range sortedBoolKeys(fields) {
			if !properties[field] {
				t.Errorf("DROPPED   models.%s has json field %q, which %s no longer declares; "+
					"it now decodes to the zero value", goName, field, schemaName)
			}
		}
		for _, property := range sortedBoolKeys(properties) {
			if fields[property] {
				continue
			}
			if strict[goName] {
				t.Errorf("MISSING   %s declares %q, which models.%s does not carry", schemaName, property, goName)
				continue
			}
			t.Logf("WARN      %s declares %q, which models.%s does not carry", schemaName, property, goName)
		}
	}
}

// schemaProperties collects a schema's own properties plus those of the schemas
// it composes with allOf.
func schemaProperties(schemas map[string]specSchema, name string, depth int) (map[string]bool, bool) {
	schema, ok := schemas[name]
	if !ok {
		return nil, false
	}
	properties := make(map[string]bool, len(schema.Properties))
	for property := range schema.Properties {
		properties[property] = true
	}
	if depth < 2 {
		for _, composed := range schema.AllOf {
			ref := strings.TrimPrefix(composed.Ref, "#/components/schemas/")
			if ref == "" || ref == composed.Ref {
				continue
			}
			inherited, ok := schemaProperties(schemas, ref, depth+1)
			if !ok {
				continue
			}
			for property := range inherited {
				properties[property] = true
			}
		}
	}
	return properties, true
}

// jsonFields lists the json field names a struct decodes, following embedded
// structs.
func jsonFields(rt reflect.Type) map[string]bool {
	fields := make(map[string]bool)

	var walk func(reflect.Type)
	walk = func(rt reflect.Type) {
		for rt.Kind() == reflect.Pointer {
			rt = rt.Elem()
		}
		if rt.Kind() != reflect.Struct {
			return
		}
		for i := 0; i < rt.NumField(); i++ {
			field := rt.Field(i)
			tag := field.Tag.Get("json")
			if tag == "-" {
				continue
			}
			name := strings.Split(tag, ",")[0]
			if field.Anonymous && name == "" {
				walk(field.Type)
				continue
			}
			if field.PkgPath != "" {
				continue // unexported
			}
			if name == "" {
				name = field.Name
			}
			fields[name] = true
		}
	}
	walk(rt)
	return fields
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
