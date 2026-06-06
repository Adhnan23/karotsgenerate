package scaffold

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const featureRoot = "templates/feature"

// AddResult reports what AddFeature created.
type AddResult struct {
	Files  []string // paths written, relative to the project root
	DryRun bool     // when true, nothing was actually written
}

// FilesList renders the created files as an indented bullet list for CLI output.
func (r AddResult) FilesList() string {
	var b strings.Builder
	for _, f := range r.Files {
		fmt.Fprintf(&b, "  • %s\n", f)
	}
	return b.String()
}

// AddFeature generates a new vertical-slice feature inside an existing project and
// wires it into internal/app/app.go. The project root is opts.ProjectDir.
func AddFeature(opts FeatureOptions) (AddResult, error) {
	var res AddResult
	if normalizeName(opts.Name) == "" {
		return res, fmt.Errorf("invalid feature name %q", opts.Name)
	}
	if !opts.IncludeAPI && !opts.IncludeWeb {
		return res, fmt.Errorf("a feature needs at least one of web or api (don't pass both -no-web and -no-api)")
	}

	modulePath, err := readModulePath(opts.ProjectDir)
	if err != nil {
		return res, err
	}
	fd, err := newFeatureData(modulePath, opts)
	if err != nil {
		return res, err
	}

	featureDir := filepath.Join(opts.ProjectDir, "internal", "features", fd.Lower)
	if _, err := os.Stat(featureDir); err == nil {
		return res, fmt.Errorf("feature %q already exists at %s", fd.Lower, featureDir)
	}

	// Render every feature template, honoring the web/api toggles.
	err = fs.WalkDir(templatesFS, featureRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		base := strings.TrimSuffix(filepath.Base(path), ".tmpl")
		if base == "api_handler.go" && !fd.IncludeAPI {
			return nil
		}
		if (base == "web_handler.go" || base == "views.templ") && !fd.IncludeWeb {
			return nil
		}
		if base == "migration.sql" {
			return nil // handled separately (needs a sequence number + final name)
		}

		content, err := templatesFS.ReadFile(path)
		if err != nil {
			return err
		}
		rendered, err := renderBytes(path, content, fd)
		if err != nil {
			return err
		}
		dest := filepath.Join(featureDir, base)
		res.Files = append(res.Files, relTo(opts.ProjectDir, dest))
		if opts.DryRun {
			return nil
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return err
		}
		return os.WriteFile(dest, rendered, 0o644)
	})
	if err != nil {
		return res, err
	}

	// Migration with the next sequence number.
	migPath, err := writeMigration(opts.ProjectDir, fd, opts.DryRun)
	if err != nil {
		return res, err
	}
	res.Files = append(res.Files, relTo(opts.ProjectDir, migPath))

	if opts.DryRun {
		res.DryRun = true
		return res, nil
	}

	// Wire into app.go.
	appGo := filepath.Join(opts.ProjectDir, "internal", "app", "app.go")
	if err := injectFeature(appGo, modulePath, fd); err != nil {
		return res, err
	}

	// Best-effort gofmt of the new feature + the edited app.go.
	if _, err := exec.LookPath("gofmt"); err == nil {
		_ = exec.Command("gofmt", "-w", featureDir, appGo).Run()
	}

	return res, nil
}

// readModulePath returns the module path from the project's go.mod.
func readModulePath(projectDir string) (string, error) {
	b, err := os.ReadFile(filepath.Join(projectDir, "go.mod"))
	if err != nil {
		return "", fmt.Errorf("not a Go project (no go.mod in %s): %w", projectDir, err)
	}
	for _, line := range strings.Split(string(b), "\n") {
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module")), nil
		}
	}
	return "", fmt.Errorf("no module directive in go.mod")
}

var migrationSeq = regexp.MustCompile(`^(\d+)_`)

// writeMigration renders the migration template into migrations/NNNN_create_<table>.sql,
// choosing NNNN as one past the highest existing numbered migration. With dryRun set it
// computes and returns the destination path without writing.
func writeMigration(projectDir string, fd FeatureData, dryRun bool) (string, error) {
	migDir := filepath.Join(projectDir, "migrations")
	entries, err := os.ReadDir(migDir)
	if err != nil {
		return "", fmt.Errorf("read migrations dir: %w", err)
	}
	max := 0
	for _, e := range entries {
		if m := migrationSeq.FindStringSubmatch(e.Name()); m != nil {
			if n, _ := strconv.Atoi(m[1]); n > max {
				max = n
			}
		}
	}
	name := fmt.Sprintf("%04d_create_%s.sql", max+1, fd.Table)
	dest := filepath.Join(migDir, name)

	content, err := templatesFS.ReadFile(featureRoot + "/migration.sql.tmpl")
	if err != nil {
		return "", err
	}
	rendered, err := renderBytes("migration.sql", content, fd)
	if err != nil {
		return "", err
	}
	if dryRun {
		return dest, nil
	}
	if err := os.WriteFile(dest, rendered, 0o644); err != nil {
		return "", err
	}
	return dest, nil
}

func relTo(base, path string) string {
	if r, err := filepath.Rel(base, path); err == nil {
		return r
	}
	return path
}
