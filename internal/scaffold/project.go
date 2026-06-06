package scaffold

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const projectRoot = "templates/project"

// GenerateProject writes a complete new project described by a into a.TargetDir.
// It walks the embedded project template tree, skips files excluded by the answers
// (e.g. the note feature when IncludeNote is false), renders *.tmpl files, and copies
// everything else verbatim. Finally it best-effort runs gofmt, seeds .env, inits git,
// and tidies modules.
func GenerateProject(a Answers, force bool) error {
	if a.ModulePath == "" {
		return errors.New("module path is required")
	}
	if a.TargetDir == "" {
		return errors.New("target directory is required")
	}

	if err := ensureWritableDir(a.TargetDir, force); err != nil {
		return err
	}

	err := fs.WalkDir(templatesFS, projectRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel := strings.TrimPrefix(path, projectRoot+"/")
		destRel := strings.TrimSuffix(rel, ".tmpl")

		if !includeProjectFile(destRel, a) {
			return nil
		}

		content, err := templatesFS.ReadFile(path)
		if err != nil {
			return err
		}
		if strings.HasSuffix(rel, ".tmpl") {
			content, err = renderBytes(rel, content, a)
			if err != nil {
				return err
			}
		}

		dest := filepath.Join(a.TargetDir, filepath.FromSlash(destRel))
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return err
		}
		return os.WriteFile(dest, content, 0o644)
	})
	if err != nil {
		return fmt.Errorf("write project: %w", err)
	}

	postGenerate(a)
	return nil
}

// includeProjectFile applies the conditional-inclusion rules. destRel is the path
// relative to the project root with any .tmpl suffix already stripped.
func includeProjectFile(destRel string, a Answers) bool {
	switch {
	case strings.HasPrefix(destRel, "internal/features/note/"),
		destRel == "e2e/notes_test.go",
		destRel == "migrations/0002_notes.sql":
		return a.IncludeNote
	case destRel == "Dockerfile":
		return a.IncludeDocker
	case destRel == ".github/workflows/ci.yml":
		return a.IncludeCI
	}
	return true
}

// ensureWritableDir creates the target directory, erroring if it already has files
// and force is not set.
func ensureWritableDir(dir string, force bool) error {
	entries, err := os.ReadDir(dir)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return os.MkdirAll(dir, 0o755)
	case err != nil:
		return err
	case len(entries) > 0 && !force:
		return fmt.Errorf("directory %s is not empty (use -force to override)", dir)
	}
	return nil
}

// postGenerate runs the best-effort finishing touches. Failures are reported but do
// not abort generation — the project files are already written. The shipped go.mod +
// go.sum are already correct, so we don't run `go mod tidy` (which would fail until
// codegen has produced *_templ.go and docs/swagger).
func postGenerate(a Answers) {
	// gofmt the generated Go (fixes whitespace from [[if]] blocks).
	if _, err := exec.LookPath("gofmt"); err == nil {
		_ = run(a.TargetDir, "gofmt", "-w", ".")
	}
	// Seed .env from the generated .env.example.
	src := filepath.Join(a.TargetDir, ".env.example")
	if b, err := os.ReadFile(src); err == nil {
		_ = os.WriteFile(filepath.Join(a.TargetDir, ".env"), b, 0o644)
	}
	// Initialize a git repo (ignore if git is missing or already a repo).
	if _, err := exec.LookPath("git"); err == nil {
		if _, statErr := os.Stat(filepath.Join(a.TargetDir, ".git")); statErr != nil {
			_ = run(a.TargetDir, "git", "init", "-q")
		}
	}
	// Run codegen if the tools are on PATH, so the project builds immediately.
	if _, err := exec.LookPath("templ"); err == nil {
		_ = run(a.TargetDir, "templ", "generate")
	}
	if a.IncludeSwagger {
		if _, err := exec.LookPath("swag"); err == nil {
			_ = run(a.TargetDir, "swag", "init", "-g", "cmd/web/main.go", "-o", "docs/swagger", "--parseInternal")
		}
	}
}

func run(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
