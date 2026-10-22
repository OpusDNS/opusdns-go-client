package opusdns

// Shape checks: does a Go type encode or decode the JSON a spec schema
// describes? They back TestSpecModels and TestSpecOperations in
// spec_coverage_test.go.
//
// The Go side is read with go/types rather than reflect, so any type in package
// models can be named from spec/coverage.yaml without a registry, and the
// operation checks can see which type a service method decodes into and which
// value it sends as the request body.

import (
	"bufio"
	"bytes"
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"os"
	"os/exec"
	"reflect"
	"sort"
	"strings"
	"sync"
	"testing"

	"gopkg.in/yaml.v3"
)

const modelsImportPath = "github.com/opusdns/opusdns-go-client/models"

// ---------------------------------------------------------------------------
// Schemas
// ---------------------------------------------------------------------------

// specSchema is the part of a JSON Schema the shape checks read. Keywords that
// only constrain values (pattern, maxLength, ...) are not modelled.
type specSchema struct {
	Ref                  string                 `yaml:"$ref"`
	Type                 schemaTypes            `yaml:"type"`
	Format               string                 `yaml:"format"`
	Properties           map[string]*specSchema `yaml:"properties"`
	Required             []string               `yaml:"required"`
	Items                *specSchema            `yaml:"items"`
	AdditionalProperties additionalProperties   `yaml:"additionalProperties"`
	AllOf                []*specSchema          `yaml:"allOf"`
	AnyOf                []*specSchema          `yaml:"anyOf"`
	OneOf                []*specSchema          `yaml:"oneOf"`
	Enum                 []interface{}          `yaml:"enum"`
	Const                yaml.Node              `yaml:"const"`
	MinLength            *int                   `yaml:"minLength"`
	Minimum              *float64               `yaml:"minimum"`
	ExclusiveMinimum     *float64               `yaml:"exclusiveMinimum"`
	Discriminator        struct {
		PropertyName string `yaml:"propertyName"`
	} `yaml:"discriminator"`
}

// schemaTypes accepts `type: string` as well as OpenAPI 3.1's `type: [string, "null"]`.
type schemaTypes []string

func (s *schemaTypes) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.ScalarNode:
		*s = schemaTypes{node.Value}
		return nil
	case yaml.SequenceNode:
		var many []string
		if err := node.Decode(&many); err != nil {
			return err
		}
		*s = many
		return nil
	default:
		return fmt.Errorf("line %d: type must be a string or a list of strings", node.Line)
	}
}

// additionalProperties is either a boolean or a schema for the values of a map.
type additionalProperties struct {
	schema *specSchema
}

func (a *additionalProperties) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind == yaml.ScalarNode && node.Tag == "!!bool" {
		return nil
	}
	var schema specSchema
	if err := node.Decode(&schema); err != nil {
		return err
	}
	a.schema = &schema
	return nil
}

const componentSchemaPrefix = "#/components/schemas/"

// resolve follows $ref until it reaches a schema with content, and reports the
// name of the last component it went through.
func resolveSchema(schemas map[string]*specSchema, s *specSchema) (*specSchema, string) {
	name := ""
	for i := 0; s != nil && s.Ref != "" && i < 16; i++ {
		ref := strings.TrimPrefix(s.Ref, componentSchemaPrefix)
		if ref == s.Ref {
			return nil, s.Ref
		}
		name = ref
		s = schemas[ref]
	}
	return s, name
}

// shape is a schema reduced to what JSON encoding and decoding depend on: the
// kinds of value it allows and, for objects and arrays, what they hold. Unions
// (anyOf/oneOf with more than one non-null alternative) are folded into one
// shape: a kind or property is allowed when any alternative allows it, and a
// property is required only when every object alternative requires it.
type shape struct {
	name     string
	kinds    map[string]bool // "string", "integer", "number", "boolean", "array", "object"
	nullable bool
	union    bool
	props    map[string][]*specSchema
	required map[string]bool
	items    []*specSchema
	values   []*specSchema

	// zeroRejected reports, per kind, whether the spec rejects that kind's zero
	// value ("", 0) in every alternative that allows the kind.
	zeroRejected map[string]bool

	// alternatives are the non-null alternatives of a top-level union. A
	// request body has to match one of them, not their merge.
	alternatives []*specSchema

	// discriminator names the property that selects the alternative. The
	// alternatives often give it a default, but the union cannot be resolved
	// without it, so it is required.
	discriminator string
}

func newShape() *shape {
	return &shape{
		kinds:        map[string]bool{},
		props:        map[string][]*specSchema{},
		required:     map[string]bool{},
		zeroRejected: map[string]bool{},
	}
}

// unconstrained reports whether the schema allows any JSON value.
func (sh *shape) unconstrained() bool {
	return len(sh.kinds) == 0 && !sh.nullable
}

func (sh *shape) label() string {
	if sh.name != "" {
		return sh.name
	}
	return "the schema"
}

func (sh *shape) kindList() string {
	kinds := sortedBoolKeys(sh.kinds)
	if sh.nullable {
		kinds = append(kinds, "null")
	}
	if len(kinds) == 0 {
		return "anything"
	}
	return strings.Join(kinds, " or ")
}

func (sh *shape) allows(goKind string) bool {
	switch goKind {
	case "number":
		return sh.kinds["number"] || sh.kinds["integer"]
	default:
		return sh.kinds[goKind]
	}
}

func oneSchema(list []*specSchema) *specSchema {
	switch len(list) {
	case 0:
		return nil
	case 1:
		return list[0]
	default:
		return &specSchema{AnyOf: list}
	}
}

type schemaIndex struct {
	schemas map[string]*specSchema
}

func (x *schemaIndex) shapeOf(s *specSchema) *shape {
	return x.shapeAt(s, 0)
}

func (x *schemaIndex) shapeAt(s *specSchema, depth int) *shape {
	sh := newShape()
	if s == nil || depth > 32 {
		return sh
	}
	if s.Ref != "" {
		resolved, name := resolveSchema(x.schemas, s)
		inner := x.shapeAt(resolved, depth+1)
		if inner.name == "" {
			inner.name = name
		}
		return inner
	}

	for _, kind := range s.Type {
		if kind == "null" {
			sh.nullable = true
			continue
		}
		sh.kinds[kind] = true
	}
	for name, property := range s.Properties {
		sh.props[name] = append(sh.props[name], property)
	}
	for _, name := range s.Required {
		sh.required[name] = true
	}
	if s.Items != nil {
		sh.items = append(sh.items, s.Items)
	}
	if s.AdditionalProperties.schema != nil {
		sh.values = append(sh.values, s.AdditionalProperties.schema)
	}
	if len(s.Type) == 0 {
		switch {
		case len(s.Properties) > 0 || len(s.Required) > 0 || s.AdditionalProperties.schema != nil:
			sh.kinds["object"] = true
		case s.Items != nil:
			sh.kinds["array"] = true
		}
		if s.Const.Kind != 0 {
			addValueKind(sh, s.Const.Tag)
		}
		for _, value := range s.Enum {
			addValueKind(sh, enumTag(value))
		}
	}
	if (s.MinLength != nil && *s.MinLength > 0) || (len(s.Enum) > 0 && !enumHasEmptyString(s.Enum)) ||
		s.Const.Kind != 0 && s.Const.Value != "" {
		sh.zeroRejected["string"] = true
	}
	if (s.Minimum != nil && *s.Minimum > 0) || s.ExclusiveMinimum != nil && *s.ExclusiveMinimum >= 0 {
		sh.zeroRejected["integer"] = true
		sh.zeroRejected["number"] = true
	}

	for _, part := range s.AllOf {
		sh.mergeAll(x.shapeAt(part, depth+1))
	}

	alternatives := append(append([]*specSchema{}, s.AnyOf...), s.OneOf...)
	if len(alternatives) > 0 {
		var (
			concrete []*shape
			schemas  []*specSchema
		)
		for _, alternative := range alternatives {
			alt := x.shapeAt(alternative, depth+1)
			if len(alt.kinds) == 0 && alt.nullable {
				sh.nullable = true
				continue
			}
			if alt.nullable {
				sh.nullable = true
			}
			concrete = append(concrete, alt)
			schemas = append(schemas, alternative)
		}
		switch len(concrete) {
		case 0:
		case 1:
			if sh.name == "" {
				sh.name = concrete[0].name
			}
			sh.mergeAll(concrete[0])
		default:
			sh.mergeUnion(concrete)
			sh.alternatives = schemas
		}
	}
	if name := s.Discriminator.PropertyName; name != "" {
		sh.discriminator = name
		sh.required[name] = true
	}
	return sh
}

// mergeAll folds in a schema the value must satisfy as well (allOf, or the
// single non-null alternative of a nullable anyOf).
func (sh *shape) mergeAll(other *shape) {
	for kind := range other.kinds {
		sh.kinds[kind] = true
	}
	sh.nullable = sh.nullable || other.nullable
	sh.union = sh.union || other.union
	for name, list := range other.props {
		sh.props[name] = append(sh.props[name], list...)
	}
	for name := range other.required {
		sh.required[name] = true
	}
	sh.items = append(sh.items, other.items...)
	sh.values = append(sh.values, other.values...)
	for kind := range other.zeroRejected {
		sh.zeroRejected[kind] = true
	}
}

// mergeUnion folds in the alternatives of an anyOf/oneOf.
func (sh *shape) mergeUnion(alternatives []*shape) {
	sh.union = true
	objects := 0
	requiredCount := map[string]int{}
	zeroAllowed := map[string]bool{}
	for _, alt := range alternatives {
		for kind := range alt.kinds {
			sh.kinds[kind] = true
			if !alt.zeroRejected[kind] {
				zeroAllowed[kind] = true
			}
		}
		for name, list := range alt.props {
			sh.props[name] = append(sh.props[name], list...)
		}
		if alt.kinds["object"] {
			objects++
			for name := range alt.required {
				requiredCount[name]++
			}
		}
		sh.items = append(sh.items, alt.items...)
		sh.values = append(sh.values, alt.values...)
	}
	for name, count := range requiredCount {
		if count == objects {
			sh.required[name] = true
		}
	}
	for kind := range sh.kinds {
		if !zeroAllowed[kind] {
			sh.zeroRejected[kind] = true
		}
	}
}

func addValueKind(sh *shape, tag string) {
	switch tag {
	case "!!str":
		sh.kinds["string"] = true
	case "!!int":
		sh.kinds["integer"] = true
	case "!!float":
		sh.kinds["number"] = true
	case "!!bool":
		sh.kinds["boolean"] = true
	case "!!null":
		sh.nullable = true
	}
}

func enumTag(value interface{}) string {
	switch value.(type) {
	case string:
		return "!!str"
	case int, int64, uint64:
		return "!!int"
	case float64:
		return "!!float"
	case bool:
		return "!!bool"
	case nil:
		return "!!null"
	}
	return ""
}

func enumHasEmptyString(values []interface{}) bool {
	for _, value := range values {
		if s, ok := value.(string); ok && s == "" {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// Go types
// ---------------------------------------------------------------------------

// goPackages holds the type-checked models package and this package (without
// its tests), plus the function declarations of both, so the checks can read a
// method's body and know the type of every expression in it.
type goPackages struct {
	fset       *token.FileSet
	models     *types.Package
	client     *types.Package
	info       *types.Info
	decls      map[types.Object]*ast.FuncDecl
	methods    map[string]*ast.FuncDecl // "Domains.RenewDomain"
	deprecated map[types.Object]bool    // struct fields documented as Deprecated
}

var (
	goPackagesOnce   sync.Once
	goPackagesResult *goPackages
	goPackagesErr    error
)

func loadGoPackages(t *testing.T) *goPackages {
	t.Helper()
	goPackagesOnce.Do(func() {
		goPackagesResult, goPackagesErr = typeCheckPackages()
	})
	if goPackagesErr != nil {
		t.Fatalf("type-check the client: %v", goPackagesErr)
	}
	return goPackagesResult
}

func typeCheckPackages() (*goPackages, error) {
	fset := token.NewFileSet()
	modelFiles, err := parseNonTest(fset, "../models")
	if err != nil {
		return nil, err
	}
	clientFiles, err := parseNonTest(fset, ".")
	if err != nil {
		return nil, err
	}

	std, err := stdlibImporter(fset, append(append([]*ast.File{}, modelFiles...), clientFiles...))
	if err != nil {
		return nil, err
	}

	info := &types.Info{
		Types: map[ast.Expr]types.TypeAndValue{},
		Defs:  map[*ast.Ident]types.Object{},
		Uses:  map[*ast.Ident]types.Object{},
	}
	models, err := (&types.Config{Importer: std}).Check(modelsImportPath, fset, modelFiles, info)
	if err != nil {
		return nil, fmt.Errorf("package models: %w", err)
	}
	conf := &types.Config{Importer: importerFunc(func(path string) (*types.Package, error) {
		if path == modelsImportPath {
			return models, nil
		}
		return std.Import(path)
	})}
	client, err := conf.Check("github.com/opusdns/opusdns-go-client/opusdns", fset, clientFiles, info)
	if err != nil {
		return nil, fmt.Errorf("package opusdns: %w", err)
	}

	pkgs := &goPackages{
		fset:       fset,
		models:     models,
		client:     client,
		info:       info,
		decls:      map[types.Object]*ast.FuncDecl{},
		methods:    map[string]*ast.FuncDecl{},
		deprecated: map[types.Object]bool{},
	}
	for _, file := range append(append([]*ast.File{}, modelFiles...), clientFiles...) {
		ast.Inspect(file, func(n ast.Node) bool {
			field, ok := n.(*ast.Field)
			if !ok || field.Doc == nil || !strings.Contains(field.Doc.Text(), "Deprecated:") {
				return true
			}
			for _, name := range field.Names {
				if obj := info.Defs[name]; obj != nil {
					pkgs.deprecated[obj] = true
				}
			}
			return true
		})
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			if obj := info.Defs[fn.Name]; obj != nil {
				pkgs.decls[obj] = fn
			}
			if service, _, ok := serviceReceiver(fn); ok {
				pkgs.methods[service+"."+fn.Name.Name] = fn
			}
		}
	}
	return pkgs, nil
}

type importerFunc func(path string) (*types.Package, error)

func (f importerFunc) Import(path string) (*types.Package, error) { return f(path) }

func parseNonTest(fset *token.FileSet, dir string) ([]*ast.File, error) {
	pkgs, err := parser.ParseDir(fset, dir, func(fi os.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	var files []*ast.File
	for _, pkg := range pkgs {
		names := make([]string, 0, len(pkg.Files))
		for name := range pkg.Files {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			files = append(files, pkg.Files[name])
		}
	}
	return files, nil
}

// stdlibImporter reads compiler export data for the standard library packages
// the files import. It asks the go command for all of them in one call, which
// is much faster than type-checking net/http and its dependencies from source.
func stdlibImporter(fset *token.FileSet, files []*ast.File) (types.Importer, error) {
	seen := map[string]bool{}
	var paths []string
	for _, file := range files {
		for _, spec := range file.Imports {
			path := strings.Trim(spec.Path.Value, `"`)
			if path == modelsImportPath || seen[path] {
				continue
			}
			seen[path] = true
			paths = append(paths, path)
		}
	}
	sort.Strings(paths)

	args := append([]string{"list", "-export", "-deps", "-f", "{{.ImportPath}} {{.Export}}"}, paths...)
	out, err := exec.Command("go", args...).Output()
	if err != nil {
		var stderr []byte
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr = exitErr.Stderr
		}
		return nil, fmt.Errorf("go list -export: %v\n%s", err, stderr)
	}
	exports := map[string]string{}
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		parts := strings.SplitN(scanner.Text(), " ", 2)
		if len(parts) == 2 && parts[1] != "" {
			exports[parts[0]] = parts[1]
		}
	}
	return importer.ForCompiler(fset, "gc", func(path string) (io.ReadCloser, error) {
		file, ok := exports[path]
		if !ok {
			return nil, fmt.Errorf("no export data for %s", path)
		}
		return os.Open(file)
	}), nil
}

// lookupModel finds a named type in package models.
func (p *goPackages) lookupModel(name string) (types.Type, bool) {
	obj, ok := p.models.Scope().Lookup(name).(*types.TypeName)
	if !ok {
		return nil, false
	}
	return obj.Type(), true
}

// unalias resolves a type alias where the toolchain represents one (go/types
// does from Go 1.22 on, with the gotypesalias setting).
func unalias(t types.Type) types.Type {
	for i := 0; i < 16; i++ {
		alias, ok := t.(interface{ Rhs() types.Type })
		if !ok {
			return t
		}
		t = alias.Rhs()
	}
	return t
}

func derefType(t types.Type) types.Type {
	t = unalias(t)
	for {
		ptr, ok := t.(*types.Pointer)
		if !ok {
			return t
		}
		t = unalias(ptr.Elem())
	}
}

func typeName(t types.Type) string {
	return types.TypeString(t, func(pkg *types.Package) string { return pkg.Name() })
}

type direction int

const (
	decoding direction = iota // response bodies
	encoding                  // request bodies
)

// jsonKind reports the kind of JSON value encoding/json reads or writes for t:
// one of the schema kinds, "any" for values it cannot judge (interfaces,
// json.RawMessage) and "custom" for types with their own JSON methods.
func jsonKind(t types.Type, dir direction) string {
	t = derefType(t)
	if named, ok := t.(*types.Named); ok {
		obj := named.Obj()
		if obj.Pkg() != nil {
			switch obj.Pkg().Path() + "." + obj.Name() {
			case "time.Time":
				return "string"
			case "encoding/json.RawMessage":
				return "any"
			case "encoding/json.Number":
				return "number"
			}
		}
		jsonMethod, textMethod := "UnmarshalJSON", "UnmarshalText"
		if dir == encoding {
			jsonMethod, textMethod = "MarshalJSON", "MarshalText"
		}
		methods := types.NewMethodSet(types.NewPointer(named))
		if methods.Lookup(nil, jsonMethod) != nil {
			return "custom"
		}
		if methods.Lookup(nil, textMethod) != nil {
			return "string"
		}
	}
	switch u := t.Underlying().(type) {
	case *types.Basic:
		switch info := u.Info(); {
		case info&types.IsString != 0:
			return "string"
		case info&types.IsBoolean != 0:
			return "boolean"
		case info&types.IsInteger != 0:
			return "integer"
		case info&types.IsFloat != 0:
			return "number"
		}
	case *types.Slice:
		if basic, ok := u.Elem().Underlying().(*types.Basic); ok && basic.Kind() == types.Byte {
			return "string"
		}
		return "array"
	case *types.Array:
		return "array"
	case *types.Map, *types.Struct:
		return "object"
	}
	return "any"
}

type goField struct {
	name      string
	typ       types.Type
	omitEmpty bool
	asString  bool
	owner     types.Type // the embedded struct that declares a promoted field, else nil
	obj       *types.Var
}

// jsonFields lists the fields encoding/json reads and writes for a struct,
// promoting the fields of embedded structs.
func jsonFields(st *types.Struct) []goField {
	return jsonFieldsOf(st, nil)
}

func jsonFieldsOf(st *types.Struct, owner types.Type) []goField {
	var fields []goField
	for i := 0; i < st.NumFields(); i++ {
		field := st.Field(i)
		tag := reflect.StructTag(st.Tag(i)).Get("json")
		if tag == "-" {
			continue
		}
		parts := strings.Split(tag, ",")
		name := parts[0]
		if field.Embedded() && name == "" {
			embeddedType := derefType(field.Type())
			if embedded, ok := embeddedType.Underlying().(*types.Struct); ok {
				fields = append(fields, jsonFieldsOf(embedded, embeddedType)...)
				continue
			}
		}
		if !field.Exported() {
			continue
		}
		if name == "" {
			name = field.Name()
		}
		f := goField{name: name, typ: field.Type(), owner: owner, obj: field}
		for _, option := range parts[1:] {
			switch option {
			case "omitempty":
				f.omitEmpty = true
			case "string":
				f.asString = true
			}
		}
		fields = append(fields, f)
	}
	sort.Slice(fields, func(i, j int) bool { return fields[i].name < fields[j].name })
	return fields
}

// ---------------------------------------------------------------------------
// Comparison
// ---------------------------------------------------------------------------

// modelSchemas is what the manifest says about Go types: the schema each one
// mirrors (models.mappings) and which ones must mirror it completely
// (models.strict). Keys are qualified Go names, "models.Zone".
type modelSchemas struct {
	canonical map[string]string
	strict    map[string]bool

	// deprecated holds the struct fields documented as Deprecated. A response
	// field kept that way for compatibility is a warning, not DROPPED.
	deprecated map[types.Object]bool
}

// shapeChecker compares Go types with schemas and collects what it finds.
// Errors are drift that loses or corrupts data; warnings are spec properties
// the client simply does not mirror (yet), or fields a shared type only
// receives from some operations.
type shapeChecker struct {
	index  *schemaIndex
	dir    direction
	models modelSchemas
	where  string // prefix for every message, e.g. the operation

	visited map[string]bool
	errors  []string
	warns   []string

	// forceRequired adds required properties to the next struct checked: the
	// discriminator of the union an alternative came from.
	forceRequired map[string]bool
}

func newShapeChecker(index *schemaIndex, dir direction, models modelSchemas, where string) *shapeChecker {
	return &shapeChecker{index: index, dir: dir, models: models, where: where, visited: map[string]bool{}}
}

func (c *shapeChecker) errorf(label, format string, args ...interface{}) {
	c.errors = append(c.errors, fmt.Sprintf("%-9s %s: %s", label, c.where, fmt.Sprintf(format, args...)))
}

func (c *shapeChecker) warnf(format string, args ...interface{}) {
	c.warns = append(c.warns, fmt.Sprintf("WARN      %s: %s", c.where, fmt.Sprintf(format, args...)))
}

// check compares t with schema s. path names the value in messages, e.g.
// "models.Zone.rrsets[]".
func (c *shapeChecker) check(t types.Type, s *specSchema, path string) {
	t = derefType(t)
	sh := c.index.shapeOf(s)
	if sh.unconstrained() {
		return
	}
	key := fmt.Sprintf("%s|%p", types.TypeString(t, nil), s)
	if c.visited[key] {
		return
	}
	c.visited[key] = true

	kind := jsonKind(t, c.dir)
	switch kind {
	case "any", "custom":
		return
	}
	if !sh.allows(kind) {
		if len(sh.kinds) == 0 {
			c.errorf("KIND", "%s is %s (%s), but %s only allows null", path, kind, typeName(t), sh.label())
			return
		}
		c.errorf("KIND", "%s is %s (%s), but %s is %s", path, kind, typeName(t), sh.label(), sh.kindList())
		return
	}

	switch u := t.Underlying().(type) {
	case *types.Slice:
		c.check(u.Elem(), oneSchema(sh.items), path+"[]")
	case *types.Array:
		c.check(u.Elem(), oneSchema(sh.items), path+"[]")
	case *types.Map:
		c.check(u.Elem(), oneSchema(sh.values), path+"{}")
	case *types.Struct:
		if c.dir == encoding && len(sh.alternatives) > 1 {
			c.checkStructAlternatives(t, sh, path)
			return
		}
		c.checkStruct(t, u, sh, path)
	}
}

// checkStructAlternatives checks a request struct against a union. The struct
// encodes one JSON object, so it has to match one alternative on its own; the
// merged shape would accept a mix of fields no alternative takes. The closest
// alternative is reported.
func (c *shapeChecker) checkStructAlternatives(t types.Type, sh *shape, path string) {
	var best *shapeChecker
	for _, alternative := range sh.alternatives {
		trial := newShapeChecker(c.index, c.dir, c.models, c.where)
		if sh.discriminator != "" {
			trial.forceRequired = map[string]bool{sh.discriminator: true}
		}
		trial.check(t, alternative, path)
		if best == nil || len(trial.errors) < len(best.errors) {
			best = trial
		}
		if len(trial.errors) == 0 {
			break
		}
	}
	c.errors = append(c.errors, best.errors...)
	c.warns = append(c.warns, best.warns...)
}

func (c *shapeChecker) checkStruct(t types.Type, st *types.Struct, sh *shape, path string) {
	if len(sh.props) == 0 && len(sh.values) > 0 {
		return // a map-like object; a struct reads it by name
	}
	goName := typeName(t)
	required := sh.required
	if len(c.forceRequired) > 0 {
		// The discriminator of the union this alternative was picked from.
		required = map[string]bool{}
		for name := range sh.required {
			required[name] = true
		}
		for name := range c.forceRequired {
			required[name] = true
		}
		c.forceRequired = nil
	}
	fields := jsonFields(st)
	// A shared type may get a subset of its canonical schema from some
	// operations. A schema that declares none of its fields is no subset: the
	// method decodes the wrong type, and every field is DROPPED.
	subset := false
	for _, field := range fields {
		if len(sh.props[field.name]) > 0 {
			subset = true
			break
		}
	}
	carried := make(map[string]bool, len(fields))
	for _, field := range fields {
		carried[field.name] = true
		fieldPath := path + "." + field.name
		schemas := sh.props[field.name]
		if len(schemas) == 0 {
			owner := goName
			if field.owner != nil {
				owner = typeName(field.owner)
			}
			switch canonical := c.canonicalSchema(owner, sh); {
			case canonical != nil && subset && len(canonical.props[field.name]) > 0:
				c.warnf("%s: %s does not send %q (only %s does), so it stays zero here", fieldPath, sh.label(), field.name, canonical.name)
			case c.dir == decoding && c.models.deprecated[field.obj]:
				c.warnf("%s: %s declares no %q; the field is deprecated and stays zero", fieldPath, sh.label(), field.name)
			case c.dir == decoding:
				c.errorf("DROPPED", "%s: %s declares no %q, so the field always decodes to its zero value; remove or rename it",
					fieldPath, sh.label(), field.name)
			default:
				c.errorf("UNKNOWN", "%s: %s declares no %q, so the API ignores or rejects it; remove or rename it",
					fieldPath, sh.label(), field.name)
			}
			continue
		}
		property := oneSchema(schemas)
		if field.asString {
			continue // `json:",string"` quotes the value; the kind check does not apply
		}
		c.check(field.typ, property, fieldPath)
		if c.dir == encoding {
			c.checkEncodedField(field, property, required[field.name], fieldPath)
		}
	}

	for _, name := range sortedSchemaKeys(sh.props) {
		if carried[name] {
			continue
		}
		switch {
		case c.dir == encoding && required[name]:
			c.errorf("MISSING", "%s requires %q, which %s does not carry", sh.label(), name, goName)
		case c.dir == decoding && c.models.strict[goName] && c.models.canonical[goName] == sh.name:
			c.errorf("MISSING", "%s declares %q, which %s (listed under models.strict) does not carry", sh.label(), name, goName)
		case c.dir == decoding:
			c.warnf("%s declares %q, which %s does not carry", sh.label(), name, goName)
		}
	}
}

// canonicalSchema returns the schema models.mappings names for a Go type, when
// a response is checked against a different one. A type shared between
// operations mirrors its canonical schema; other operations may send a subset.
func (c *shapeChecker) canonicalSchema(goName string, sh *shape) *shape {
	name := c.models.canonical[goName]
	if c.dir != decoding || name == "" || name == sh.name {
		return nil
	}
	canonical := c.index.shapeOf(&specSchema{Ref: componentSchemaPrefix + name})
	canonical.name = name
	return canonical
}

// checkEncodedField catches fields whose Go encoding sends a value the API
// rejects, or never sends one it needs.
func (c *shapeChecker) checkEncodedField(field goField, property *specSchema, required bool, path string) {
	t := unalias(field.typ)
	if _, isPointer := t.(*types.Pointer); isPointer {
		return
	}
	kind := jsonKind(t, encoding)
	if required && field.omitEmpty && kind == "boolean" {
		c.errorf("OMITTED", "%s is required, but omitempty on a bool means false is never sent; drop omitempty", path)
		return
	}
	if required || field.omitEmpty {
		return
	}
	switch kind {
	case "string", "integer", "number":
		if c.index.shapeOf(property).zeroRejected[kind] {
			c.errorf("ZERO", "%s is optional, but without omitempty an unset field is sent as its zero value, which the spec rejects; add omitempty or use a pointer", path)
		}
	}
}

func sortedSchemaKeys(m map[string][]*specSchema) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// ---------------------------------------------------------------------------
// The checker against the drift it exists to catch
// ---------------------------------------------------------------------------

// The fixtures reproduce, in small, bugs these checks were written after:
// renew sending period as a number, a wrapper object decoded from a bare
// array, a required bool behind omitempty, an empty auth code sent as "".
const shapeFixtureSchemas = `
DomainPeriod:
  type: object
  properties: {value: {type: integer}, unit: {type: string}}
  required: [value, unit]
RenewRequest:
  type: object
  properties:
    period: {$ref: '#/components/schemas/DomainPeriod'}
    current_expiry_date: {type: string, format: date-time}
  required: [period, current_expiry_date]
Item:
  type: object
  properties: {name: {type: string}}
ItemList:
  type: array
  items: {$ref: '#/components/schemas/Item'}
WithdrawRequest:
  type: object
  properties: {zone_delete: {type: boolean}}
  required: [zone_delete]
TransferRequest:
  type: object
  properties:
    name: {type: string}
    auth_code:
      anyOf: [{type: string, minLength: 6}, {type: 'null'}]
  required: [name]
Renewed:
  type: object
  properties: {name: {type: string}, expires_on: {type: string}}
TransferPending:
  type: object
  properties: {name: {type: string}, request_id: {type: string}}
RenewResponse:
  anyOf: [{$ref: '#/components/schemas/Renewed'}, {$ref: '#/components/schemas/TransferPending'}]
Command:
  oneOf: [{$ref: '#/components/schemas/CreateCommand'}, {$ref: '#/components/schemas/DeleteCommand'}]
  discriminator: {propertyName: command}
CreateCommand:
  type: object
  properties: {command: {const: create, type: string}, payload: {type: object}}
  required: [payload]
DeleteCommand:
  type: object
  properties: {command: {const: delete, type: string}, id: {type: string}}
  required: [id]
ZoneFull:
  type: object
  properties: {name: {type: string}, zone_id: {type: string}, tags: {type: array, items: {type: string}}}
ZoneSummary:
  type: object
  properties: {name: {type: string}, zone_id: {type: string}}
ZoneChanges:
  type: object
  properties: {changeset_id: {type: string}, zone_name: {type: string}}
`

const shapeFixtureSource = `package fixture

type Period struct {
	Value int    ` + "`json:\"value\"`" + `
	Unit  string ` + "`json:\"unit\"`" + `
}

type RenewInt struct {
	Period int ` + "`json:\"period\"`" + `
}

type RenewOK struct {
	Period            Period ` + "`json:\"period\"`" + `
	CurrentExpiryDate string ` + "`json:\"current_expiry_date\"`" + `
}

type Item struct {
	Name  string ` + "`json:\"name\"`" + `
	Color string ` + "`json:\"color\"`" + `
}

type ItemWrapper struct {
	Results []Item ` + "`json:\"results\"`" + `
}

type WithdrawOmit struct {
	ZoneDelete bool ` + "`json:\"zone_delete,omitempty\"`" + `
}

type TransferEmpty struct {
	Name     string ` + "`json:\"name\"`" + `
	AuthCode string ` + "`json:\"auth_code\"`" + `
}

type TransferOK struct {
	Name     string ` + "`json:\"name\"`" + `
	AuthCode string ` + "`json:\"auth_code,omitempty\"`" + `
}

type RenewResult struct {
	Name      string ` + "`json:\"name\"`" + `
	ExpiresOn string ` + "`json:\"expires_on\"`" + `
	RequestID string ` + "`json:\"request_id\"`" + `
}

type CreateCommand struct {
	Command string      ` + "`json:\"command\"`" + `
	Payload interface{} ` + "`json:\"payload\"`" + `
}

type MixedCommand struct {
	Command string      ` + "`json:\"command\"`" + `
	Payload interface{} ` + "`json:\"payload\"`" + `
	ID      string      ` + "`json:\"id\"`" + `
}

type Zone struct {
	Name   string   ` + "`json:\"name\"`" + `
	ZoneID string   ` + "`json:\"zone_id\"`" + `
	Tags   []string ` + "`json:\"tags\"`" + `
}
`

func TestShapeCheckerCatchesDrift(t *testing.T) {
	var schemas map[string]*specSchema
	if err := yaml.Unmarshal([]byte(shapeFixtureSchemas), &schemas); err != nil {
		t.Fatalf("parse fixture schemas: %v", err)
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "fixture.go", shapeFixtureSource, 0)
	if err != nil {
		t.Fatalf("parse fixture source: %v", err)
	}
	pkg, err := (&types.Config{}).Check("fixture", fset, []*ast.File{file}, nil)
	if err != nil {
		t.Fatalf("type-check fixture source: %v", err)
	}
	named := func(name string) types.Type { return pkg.Scope().Lookup(name).Type() }

	zoneMapped := modelSchemas{canonical: map[string]string{"fixture.Zone": "ZoneFull"}}

	cases := []struct {
		name   string
		dir    direction
		goType types.Type
		schema string
		want   []string
		models modelSchemas
	}{
		{"period sent as a number", encoding, named("RenewInt"), "RenewRequest", []string{"KIND", "MISSING"}, modelSchemas{}},
		{"renew request as the spec has it", encoding, named("RenewOK"), "RenewRequest", nil, modelSchemas{}},
		{"wrapper object for a bare array", decoding, named("ItemWrapper"), "ItemList", []string{"KIND"}, modelSchemas{}},
		{"field the schema does not declare", decoding, types.NewSlice(named("Item")), "ItemList", []string{"DROPPED"}, modelSchemas{}},
		{"required bool behind omitempty", encoding, named("WithdrawOmit"), "WithdrawRequest", []string{"OMITTED"}, modelSchemas{}},
		{"empty auth code sent as \"\"", encoding, named("TransferEmpty"), "TransferRequest", []string{"ZERO"}, modelSchemas{}},
		{"auth code omitted when empty", encoding, named("TransferOK"), "TransferRequest", nil, modelSchemas{}},
		{"one struct for both renew shapes", decoding, named("RenewResult"), "RenewResponse", nil, modelSchemas{}},
		{"one alternative of a discriminated union", encoding, named("CreateCommand"), "Command", nil, modelSchemas{}},
		{"fields of two alternatives mixed", encoding, named("MixedCommand"), "Command", []string{"UNKNOWN"}, modelSchemas{}},
		{"shared type, subset of its schema", decoding, named("Zone"), "ZoneSummary", nil, zoneMapped},
		{"shared type, wrong schema", decoding, named("Zone"), "ZoneChanges", []string{"DROPPED", "DROPPED", "DROPPED"}, zoneMapped},
	}
	for _, tc := range cases {
		checker := newShapeChecker(&schemaIndex{schemas: schemas}, tc.dir, tc.models, tc.name)
		checker.check(tc.goType, &specSchema{Ref: componentSchemaPrefix + tc.schema}, typeName(tc.goType))
		var got []string
		for _, message := range checker.errors {
			got = append(got, strings.Fields(message)[0])
		}
		sort.Strings(got)
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%s: got %v, want %v\n%s", tc.name, got, tc.want, strings.Join(checker.errors, "\n"))
		}
	}
}
