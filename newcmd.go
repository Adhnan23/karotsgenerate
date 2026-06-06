package main

import (
	"flag"
	"fmt"

	"github.com/Adhnan23/karotsgenerate/internal/prompt"
	"github.com/Adhnan23/karotsgenerate/internal/scaffold"
)

// cmdNew handles `karotsgenerate new`. Without -yes it opens an interactive form
// (pre-filled with any flags given); with -yes it runs non-interactively from flags.
func cmdNew(args []string) error {
	fs := flag.NewFlagSet("new", flag.ContinueOnError)

	var f prompt.Flags
	fs.StringVar(&f.Module, "module", "", "Go module path, e.g. github.com/you/app (required)")
	fs.StringVar(&f.Name, "name", "", "project name (defaults to the last module path segment)")
	fs.StringVar(&f.Dir, "dir", "", "target directory (defaults to ./<name>)")
	fs.StringVar(&f.DBName, "db-name", "app", "Postgres database name")
	fs.StringVar(&f.DBPort, "db-port", "5432", "Postgres host port")
	fs.StringVar(&f.AppPort, "app-port", "8080", "HTTP server port")
	fs.StringVar(&f.JWTSecret, "jwt-secret", "", "JWT signing secret (default: randomly generated)")
	fs.BoolVar(&f.Note, "note", true, "include the example 'note' feature")
	fs.BoolVar(&f.Swagger, "swagger", true, "include Swagger UI + OpenAPI docs")
	fs.BoolVar(&f.Docker, "docker", true, "include the app Dockerfile + compose app service")
	fs.BoolVar(&f.CI, "ci", true, "include the GitHub Actions workflow")

	yes := fs.Bool("yes", false, "skip the interactive form; use flags and defaults")
	force := fs.Bool("force", false, "allow generating into a non-empty directory")
	if err := fs.Parse(args); err != nil {
		return err
	}

	answers, err := prompt.Collect(f, *yes)
	if err != nil {
		return err
	}

	if err := scaffold.GenerateProject(answers, *force); err != nil {
		return err
	}

	fmt.Printf(`
✓ Project generated at %s

Next steps:
  cd %s
  make tools         # install templ, swag, goose, air (one time)
  make db-up         # start Postgres in docker
  make run           # generate code, migrate, serve on :%s

Add a feature later with:
  karotsgenerate add feature <name>
`, answers.TargetDir, answers.TargetDir, answers.AppPort)
	return nil
}
