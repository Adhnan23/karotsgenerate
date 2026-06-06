package scaffold

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
	IncludeAPI bool
	IncludeWeb bool
	ProjectDir string // project root (default ".")
}

// FeatureData is what the feature templates render against.
type FeatureData struct {
	ModulePath  string // taken from the target project's go.mod
	Lower       string // package + variable base, e.g. "invoice"
	Title       string // exported type base, e.g. "Invoice"
	TitlePlural string // exported plural for view names, e.g. "Invoices"
	Table       string // DB table + route segment, e.g. "invoices"
	IncludeAPI  bool
	IncludeWeb  bool
}
