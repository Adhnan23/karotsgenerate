# karotsgenerate

A CLI that scaffolds **fullstack Go projects** — and individual features within them —
following the architecture of the [go-learning](../learning) reference app:

> Echo v4 · templ · htmx + Alpine · sqlx + pgx · Postgres 17 · goose (embedded) ·
> go-playground/validator · JWT cookie auth (rotation + reuse detection) ·
> dual **web + JSON API** surfaces over one service · Swagger · layered testing.

## Contents

- [Requirements](#requirements)
- [Install](#install)
- [Commands at a glance](#commands-at-a-glance)
- [Quick start (full lifecycle)](#quick-start-full-lifecycle)
- [`new` — create a project](#new--create-a-project)
- [`add feature` — add a vertical slice](#add-feature--add-a-vertical-slice)
- [`remove feature` — delete a slice](#remove-feature--delete-a-slice)
- [`list feature` — inspect a project](#list-feature--inspect-a-project)
- [`doctor` — check your toolchain](#doctor--check-your-toolchain)
- [How it works](#how-it-works)
- [Generator layout](#generator-layout)

## Requirements

- **Go 1.26+** (to run the generator and build the projects it makes).
- For a generated project's `make` targets: **templ**, **swag**, **goose**, **air**, and
  **docker** (for Postgres). Run `karotsgenerate doctor` to see what's missing — or
  `make tools` inside a generated project to install the Go ones.

## Install

```sh
go install github.com/Adhnan23/karotsgenerate@latest
# or, from a clone:
go build -o bin/karotsgenerate . && sudo mv bin/karotsgenerate /usr/local/bin/
```

## Commands at a glance

| Command | What it does |
|---------|--------------|
| `karotsgenerate new` | Scaffold a new project (interactive form, or `-yes` + flags). |
| `karotsgenerate add feature <name>` | Generate a full vertical slice and wire it in. |
| `karotsgenerate remove feature <name>` | Delete a slice + its migration and un-wire it. |
| `karotsgenerate list feature` | List features with mount status + migration. |
| `karotsgenerate doctor` | Check that go/templ/swag/goose/git/docker are installed. |
| `karotsgenerate version` | Print the version. |

`--dry-run` works on `new`, `add feature`, and `remove feature`: it prints the planned
changes and writes nothing. Run any command with `-h` for its flags.

## Quick start (full lifecycle)

```sh
# 1. scaffold a project
karotsgenerate new -yes -module github.com/you/app -dir ./app

cd app
make tools           # install templ, swag, goose, air (one time)
make db-up           # start Postgres in docker
make run             # generate code, migrate, serve on :8080

# 2. add a feature with a real schema (in another terminal, from ./app)
karotsgenerate add feature product \
  -fields "name:string,price:decimal,in_stock:bool,details:text"
make generate        # render templ + swagger for the new feature
make migrate-up      # apply the new migration (make run does this on boot too)

# 3. inspect / remove
karotsgenerate list feature
karotsgenerate remove feature product
```

## `new` — create a project

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
`git init`s, and — if `templ`/`swag` are on your PATH — runs `templ generate` and
`swag init` so the project builds immediately. (The shipped `go.mod`/`go.sum` are already
correct, so it does **not** run `go mod tidy`.)

### Non-interactive

```sh
karotsgenerate new -yes \
  -module github.com/you/app \
  -dir ./app \
  -db-port 5433 \
  -note=false -docker=false
```

Run `karotsgenerate new -h` for every flag (`-name`, `-db-name`, `-app-port`,
`-jwt-secret`, `-swagger`, `-ci`, `-force`, `--dry-run`, …).

## `add feature` — add a vertical slice

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

### Custom fields

By default a feature gets `title` (string) + `body` (text) columns. Pass `-fields` to
define your own schema, which flows through the model struct, request/response DTOs (with
validation tags), the migration columns, the repository SQL, and the templ form:

```sh
karotsgenerate add feature product \
  -fields "name:string,price:decimal,in_stock:bool,details:text"
```

Supported field types:

| type | Go | Postgres | web input |
|------|----|----------|-----------|
| `string` | `string` | `TEXT` (validated `max=200`) | text |
| `text` | `string` | `TEXT` (validated `max=10000`) | textarea |
| `int` | `int64` | `BIGINT` | number |
| `bool` | `bool` | `BOOLEAN` | checkbox |
| `decimal` | `float64` | `NUMERIC` | number (step 0.01) |
| `time` | `time.Time` | `TIMESTAMPTZ` | datetime-local |

`id`, `user_id`, `created_at`, `updated_at` are added automatically and may not be
redefined. Field names must be `snake_case`.

### Flags

| Flag | Effect |
|------|--------|
| `-fields "name:type,..."` | columns for the feature (default: `title:string,body:text`) |
| `-no-api` | skip the JSON API handler |
| `-no-web` | skip the htmx web handler + views |
| `-table name` | override the DB table / route (default: naive plural of the name) |
| `-dir path` | project root (default `.`) |
| `--dry-run` | print what would be created, write nothing |

## `remove feature` — delete a slice

The inverse of `add feature`: deletes `internal/features/<name>/` and the feature's
migration, and un-wires the import + mount from `app.go`.

```sh
karotsgenerate remove feature product            # prints the plan, then asks to confirm
karotsgenerate remove feature product -force     # skip the confirmation
karotsgenerate remove feature product --dry-run  # show the plan only
```

Flags: `-table name` (to locate the migration if it isn't the naive plural), `-dir path`,
`-force`, `--dry-run`. The foundational `auth` feature cannot be removed.

## `list feature` — inspect a project

```sh
karotsgenerate list feature
```

Lists every feature under `internal/features/`, whether it's mounted in `app.go`, and the
migration file (if any) associated with it. Read-only. Accepts `-dir path`.

## `doctor` — check your toolchain

```sh
karotsgenerate doctor
```

Reports a ✓/✗ for `go` (and its version — needs 1.26+), `templ`, `swag`, `goose`, `git`,
and `docker`, with an install hint for anything missing. Exits non-zero only if Go itself
is missing or too old.

## How it works

- Templates are the go-learning tree, embedded (`//go:embed`) and tokenized. Rendering
  uses `text/template` with **custom delimiters `[[ ]]`** so Go `{}` and templ `{ }`/`@`
  syntax pass through untouched — only `[[.Field]]` / `[[if]]` / `[[range]]` markers are
  special.
- Conditional files (the `note` feature, Dockerfile, CI workflow) are skipped via
  `includeProjectFile` in `internal/scaffold/project.go`; mixed-concern files (`app.go`,
  `Makefile`, `docker-compose.yml`) use in-file `[[if]]` blocks.
- `add feature` reads the target's `go.mod` for the module path, parses `-fields` into a
  `[]Field` (each with its Go type, SQL type, validation tag, and form widget), renders
  the `templates/feature/*` set ranging over those fields, writes the next migration, and
  injects the import + mount lines at the markers. `remove feature` reverses the marker
  edit exactly.

## Generator layout

```
main.go            command dispatch (new | add | remove | list | doctor | version)
newcmd.go          `new`           — flags + interactive form + generate
addcmd.go          `add feature`   — flags (incl. -fields) + generate + wire
removecmd.go       `remove feature`— plan + confirm + delete + un-wire
listcmd.go         `list feature`  — report features + mount status
doctorcmd.go       `doctor`        — toolchain check
internal/
  prompt/          huh form + flag fallback -> scaffold.Answers
  scaffold/
    answers.go     Answers / FeatureOptions / FeatureData / Field (+ SQL/Go helpers)
    naming.go      pluralize / title / parseFields (field-type table)
    project.go     GenerateProject: walk + render the project tree (+ dry-run)
    feature.go     AddFeature: render a feature + migration, wire app.go
    remove.go      RemoveFeature + ListFeatures
    inject.go      marker-based import/mount injection + removal
    render.go      embedded FS + [[ ]] template execution
    *_test.go      unit tests + a golden generate/add/remove test
    templates/
      project/     the go-learning tree, tokenized (*.tmpl)
      feature/     one vertical slice, tokenized (*.tmpl)
.github/workflows/ci.yml   build/vet/gofmt/test + end-to-end generate smoke
```
