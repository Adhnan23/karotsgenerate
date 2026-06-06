package scaffold

import (
	"fmt"
	"strings"
)

// Answers is the full set of choices that parameterize a new project. It is the data
// object passed to every project template (referenced as [[.Field]] inside *.tmpl).
type Answers struct {
	ModulePath  string // e.g. github.com/you/app  — replaces the reference module path
	ProjectName string // human/dir name, e.g. "app"
	TargetDir   string // absolute directory the project is written to
	DBName      string // Postgres database name
	DBPort      string // Postgres host port
	AppPort     string // HTTP server port
	JWTSecret   string // signing secret seeded into .env.example

	IncludeNote    bool // the example 'note' feature
	IncludeSwagger bool // Swagger UI + OpenAPI docs
	IncludeDocker  bool // app Dockerfile + compose 'app' service
	IncludeCI      bool // GitHub Actions workflow
}

// FeatureOptions configures `add feature`.
type FeatureOptions struct {
	Name       string // raw feature name as typed, e.g. "invoice" or "BlogPost"
	Table      string // optional explicit table/route name; defaults to plural of Name
	Fields     string // optional "col:type,col:type" spec; empty -> title:string,body:text
	IncludeAPI bool
	IncludeWeb bool
	ProjectDir string // project root (default ".")
	DryRun     bool   // when true, report planned changes without writing
}

// Field is one user-defined column of a generated feature. Everything the templates
// need to emit consistent Go, SQL, validation tags, and HTML inputs for the column is
// precomputed here so the templates can stay declarative ([[range .Fields]]).
type Field struct {
	Column   string // snake_case DB column + form/json name, e.g. "in_stock"
	GoName   string // exported Go field name, e.g. "InStock"
	GoType   string // Go type, e.g. "int64", "string", "bool"
	SQLType  string // Postgres column type, e.g. "BIGINT"
	Validate string // go-playground/validator tag contents (may be empty)
	WebInput string // logical input kind: text | textarea | number | checkbox | datetime
}

// IsBool/IsText/etc. let the templ form pick the right widget without leaking the
// WebInput string literal into multiple templates.
func (f Field) IsBool() bool     { return f.WebInput == "checkbox" }
func (f Field) IsText() bool     { return f.WebInput == "textarea" }
func (f Field) IsNumber() bool   { return f.WebInput == "number" }
func (f Field) IsDateTime() bool { return f.WebInput == "datetime" }

// FeatureData is what the feature templates render against.
type FeatureData struct {
	ModulePath  string  // taken from the target project's go.mod
	Lower       string  // package + variable base, e.g. "invoice"
	Title       string  // exported type base, e.g. "Invoice"
	TitlePlural string  // exported plural for view names, e.g. "Invoices"
	Table       string  // DB table + route segment, e.g. "invoices"
	Fields      []Field // user-defined columns (defaults to title/body)
	IncludeAPI  bool
	IncludeWeb  bool
}

// --- Computed strings used by repository.go.tmpl, where building column lists inline
// with [[range]] + comma handling would be noisy. Kept as methods so the data stays the
// single source of truth. ---

// SelectColumns is the full ordered column list for SELECT/RETURNING, e.g.
// "id, user_id, name, price, created_at, updated_at".
func (d FeatureData) SelectColumns() string {
	cols := []string{"id", "user_id"}
	for _, f := range d.Fields {
		cols = append(cols, f.Column)
	}
	cols = append(cols, "created_at", "updated_at")
	return joinComma(cols)
}

// InsertColumns is the user columns inserted on create, e.g. "name, price".
func (d FeatureData) InsertColumns() string {
	cols := make([]string, len(d.Fields))
	for i, f := range d.Fields {
		cols[i] = f.Column
	}
	return joinComma(cols)
}

// InsertPlaceholders are the positional params for the user columns after user_id ($1),
// e.g. "$2, $3".
func (d FeatureData) InsertPlaceholders() string {
	ph := make([]string, len(d.Fields))
	for i := range d.Fields {
		ph[i] = fmt.Sprintf("$%d", i+2) // $1 is user_id
	}
	return joinComma(ph)
}

// UpdateAssignments are the SET clauses for update, e.g. "name = $2, price = $3".
func (d FeatureData) UpdateAssignments() string {
	parts := make([]string, len(d.Fields))
	for i, f := range d.Fields {
		parts[i] = fmt.Sprintf("%s = $%d", f.Column, i+2) // $1 is id
	}
	return joinComma(parts)
}

// CreateArgs is the Go argument list reading from the request DTO, e.g.
// "in.Name, in.Price".
func (d FeatureData) CreateArgs() string {
	args := make([]string, len(d.Fields))
	for i, f := range d.Fields {
		args[i] = "in." + f.GoName
	}
	return joinComma(args)
}

func joinComma(parts []string) string { return strings.Join(parts, ", ") }
