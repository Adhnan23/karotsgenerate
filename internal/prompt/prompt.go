// Package prompt collects the answers that parameterize a new project, either from
// an interactive huh form or (with -yes) straight from command-line flags.
package prompt

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/charmbracelet/huh"

	"github.com/Adhnan23/karotsgenerate/internal/scaffold"
)

// Flags mirrors the `new` command-line flags. Empty string fields fall back to
// defaults; bool fields default to true (set by the flag package).
type Flags struct {
	Module    string
	Name      string
	Dir       string
	DBName    string
	DBPort    string
	AppPort   string
	JWTSecret string
	Note      bool
	Swagger   bool
	Docker    bool
	CI        bool
}

// A module path is one or more slash-separated segments. A single bare segment
// (e.g. "karots") is allowed for local/learning projects; a domain-style path
// (e.g. "github.com/you/app") is what you want if you'll publish it.
var moduleRe = regexp.MustCompile(`^[a-zA-Z0-9._~-]+(/[a-zA-Z0-9._~-]+)*$`)

// Collect turns flags (and, unless yes, interactive answers) into scaffold.Answers,
// applying defaults and validating required fields.
func Collect(f Flags, yes bool) (scaffold.Answers, error) {
	a := scaffold.Answers{
		ModulePath:     f.Module,
		ProjectName:    f.Name,
		TargetDir:      f.Dir,
		DBName:         orDefault(f.DBName, "app"),
		DBPort:         orDefault(f.DBPort, "5432"),
		AppPort:        orDefault(f.AppPort, "8080"),
		JWTSecret:      f.JWTSecret,
		IncludeNote:    f.Note,
		IncludeSwagger: f.Swagger,
		IncludeDocker:  f.Docker,
		IncludeCI:      f.CI,
	}

	if !yes {
		if err := runForm(&a); err != nil {
			return a, err
		}
	}

	if err := finalize(&a); err != nil {
		return a, err
	}
	return a, nil
}

// runForm shows the interactive huh form, mutating a in place. Each question is its
// own group so huh renders them one page at a time (Enter validates the page and
// advances to the next) instead of stacking a whole group on one screen.
func runForm(a *scaffold.Answers) error {
	field := func(f huh.Field) *huh.Group { return huh.NewGroup(f) }

	form := huh.NewForm(
		field(huh.NewInput().
			Title("Go module path").
			Description("e.g. github.com/you/app").
			Value(&a.ModulePath).
			Validate(validateModule)),
		field(huh.NewInput().
			Title("Project name").
			Description("blank = last segment of the module path").
			Value(&a.ProjectName)),
		field(huh.NewInput().
			Title("Target directory").
			Description("blank = ./<project name>").
			Value(&a.TargetDir)),
		field(huh.NewInput().Title("Postgres database name").Value(&a.DBName)),
		field(huh.NewInput().Title("Postgres host port").Value(&a.DBPort)),
		field(huh.NewInput().Title("HTTP server port").Value(&a.AppPort)),
		field(huh.NewConfirm().Title("Include the example 'note' feature?").Value(&a.IncludeNote)),
		field(huh.NewConfirm().Title("Include Swagger UI + OpenAPI docs?").Value(&a.IncludeSwagger)),
		field(huh.NewConfirm().Title("Include the app Dockerfile?").Value(&a.IncludeDocker)),
		field(huh.NewConfirm().Title("Include the GitHub Actions CI workflow?").Value(&a.IncludeCI)),
	)
	return form.Run()
}

// finalize fills derived defaults and validates the required module path.
func finalize(a *scaffold.Answers) error {
	if err := validateModule(a.ModulePath); err != nil {
		return err
	}
	if a.ProjectName == "" {
		a.ProjectName = path.Base(a.ModulePath)
	}
	if a.TargetDir == "" {
		a.TargetDir = "./" + a.ProjectName
	}
	abs, err := filepath.Abs(a.TargetDir)
	if err != nil {
		return err
	}
	a.TargetDir = abs

	if a.JWTSecret == "" {
		a.JWTSecret = randomSecret()
	}
	return nil
}

func validateModule(s string) error {
	if strings.TrimSpace(s) == "" {
		return fmt.Errorf("module path is required")
	}
	if !moduleRe.MatchString(s) {
		return fmt.Errorf(`module path must be slash-separated segments, e.g. "github.com/you/app" or just "app"`)
	}
	return nil
}

func orDefault(v, def string) string {
	if strings.TrimSpace(v) == "" {
		return def
	}
	return v
}

func randomSecret() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "dev-only-change-me-to-a-long-random-string"
	}
	return hex.EncodeToString(b)
}
