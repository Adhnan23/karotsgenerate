# karotsgenerate

A CLI that scaffolds **fullstack Go projects** — and individual features within them —
following the architecture of the [go-learning](../learning) reference app:

> Echo v4 · templ · htmx + Alpine · sqlx + pgx · Postgres 17 · goose (embedded) ·
> go-playground/validator · JWT cookie auth (rotation + reuse detection) ·
> dual **web + JSON API** surfaces over one service · Swagger · layered testing.

## Install

```sh
go install github.com/Adhnan23/karotsgenerate@latest
# or, from a clone:
go build -o bin/karotsgenerate . && sudo mv bin/karotsgenerate /usr/local/bin/
```

## Create a project

```sh
karotsgenerate new
```

An interactive form (powered by `charmbracelet/huh`) asks for:

| Prompt | Default |
|--------|---------|
| Go module path | *(required)* e.g. `github.com/you/app` |
| Project name | last segment of the module path |
| Target directory | `./<project name>` |
| Postgres database name / host port | `app` / `5432` |
| HTTP server port | `8080` |
| Include the example `note` feature? | yes |
| Include Swagger UI + OpenAPI docs? | yes |
| Include the app Dockerfile? | yes |
| Include the GitHub Actions CI workflow? | yes |

`auth` and the whole `internal/platform` layer are always generated — they're the
foundation. The JSON API surface is always generated (it's the project's reason to
exist); Swagger is the optional *docs UI* on top of it.

After writing the files it best-effort runs `gofmt`, seeds `.env` from `.env.example`,
`git init`s, and runs `go mod tidy`. Then:

```sh
cd <project>
make tools     # templ, swag, goose, air
make db-up     # Postgres in docker
make run       # generate code, migrate, serve
```

### Non-interactive

```sh
karotsgenerate new -yes \
  -module github.com/you/app \
  -dir ./app \
  -db-port 5433 \
  -note=false -docker=false
```

Run `karotsgenerate new -h` for every flag.

## Add a feature

From the **root of a generated project**:

```sh
karotsgenerate add feature invoice
```

This generates a full vertical slice under `internal/features/invoice/`
(`model.go`, `repository.go`, `service.go`, `module.go`, `web_handler.go`,
`api_handler.go`, `views.templ`), creates the next sequential migration
(`migrations/NNNN_create_invoices.sql`), and **auto-wires the module** into
`internal/app/app.go` (via the `// karotsgenerate:imports` / `:mounts` markers).

Then:

```sh
make generate    # render templ + swagger for the new feature
make migrate-up  # apply the new migration (make run does it on boot too)
```

Flags:

| Flag | Effect |
|------|--------|
| `-no-api` | skip the JSON API handler |
| `-no-web` | skip the htmx web handler + views |
| `-table name` | override the DB table / route (default: naive plural of the name) |
| `-dir path` | project root (default `.`) |

## How it works

- Templates are the go-learning tree, embedded (`//go:embed`) and tokenized. Rendering
  uses `text/template` with **custom delimiters `[[ ]]`** so Go `{}` and templ `{ }`/`@`
  syntax pass through untouched — only `[[.Field]]` / `[[if]]` markers are special.
- Conditional files (the `note` feature, Dockerfile, CI workflow) are skipped via the
  manifest in `internal/scaffold/project.go`; mixed-concern files (`app.go`, `Makefile`,
  `docker-compose.yml`) use in-file `[[if]]` blocks.
- `add feature` reads the target's `go.mod` for the module path, renders the
  `templates/feature/*` set, and injects the import + mount lines at the markers.

## Layout

```
main.go            command dispatch (new | add feature | version)
newcmd.go          `new`  — flags + interactive form + generate
addcmd.go          `add feature` — flags + generate + wire
internal/
  prompt/          huh form + flag fallback -> scaffold.Answers
  scaffold/
    answers.go     Answers / FeatureOptions / FeatureData
    project.go     GenerateProject: walk + render the project tree
    feature.go     AddFeature: render a feature + migration, wire app.go
    inject.go      marker-based import/mount injection
    naming.go      pluralize / title helpers
    render.go      embedded FS + [[ ]] template execution
    templates/
      project/     the go-learning tree, tokenized (*.tmpl)
      feature/     one vertical slice, tokenized (*.tmpl)
```
