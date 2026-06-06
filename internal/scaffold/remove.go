package scaffold

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// RemoveResult reports what RemoveFeature did (or, in dry-run, would do).
type RemoveResult struct {
	Files   []string // paths deleted, relative to the project root
	Unwired bool     // whether app.go was (or would be) modified
	DryRun  bool
}

// FilesList renders the deleted files as an indented bullet list for CLI output.
func (r RemoveResult) FilesList() string {
	var b strings.Builder
	for _, f := range r.Files {
		fmt.Fprintf(&b, "  • %s\n", f)
	}
	return b.String()
}

// RemoveFeature deletes a feature's vertical slice + its migration and un-wires it from
// internal/app/app.go. With opts.DryRun set it computes the plan without touching disk.
func RemoveFeature(opts FeatureOptions) (RemoveResult, error) {
	res := RemoveResult{DryRun: opts.DryRun}
	if normalizeName(opts.Name) == "" {
		return res, fmt.Errorf("invalid feature name %q", opts.Name)
	}

	modulePath, err := readModulePath(opts.ProjectDir)
	if err != nil {
		return res, err
	}
	fd, err := newFeatureData(modulePath, opts)
	if err != nil {
		return res, err
	}
	if fd.Lower == "auth" {
		return res, fmt.Errorf("refusing to remove the auth feature (it is foundational)")
	}

	featureDir := filepath.Join(opts.ProjectDir, "internal", "features", fd.Lower)
	if _, err := os.Stat(featureDir); err != nil {
		return res, fmt.Errorf("feature %q not found at %s", fd.Lower, featureDir)
	}

	// Collect the slice's files for reporting.
	_ = fs.WalkDir(os.DirFS(featureDir), ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		res.Files = append(res.Files, filepath.Join("internal", "features", fd.Lower, p))
		return nil
	})
	// Migration(s) belonging to this feature (loose match on table/name).
	migs := featureMigrations(opts.ProjectDir, fd)
	for _, m := range migs {
		res.Files = append(res.Files, relTo(opts.ProjectDir, m))
	}
	sort.Strings(res.Files)

	// In dry-run we also predict the un-wire without modifying app.go.
	appGo := filepath.Join(opts.ProjectDir, "internal", "app", "app.go")
	if opts.DryRun {
		if b, err := os.ReadFile(appGo); err == nil {
			importLine, mountBlock := featureWiring(modulePath, fd)
			res.Unwired = strings.Contains(string(b), importLine) || strings.Contains(string(b), mountBlock)
		}
		return res, nil
	}

	if err := os.RemoveAll(featureDir); err != nil {
		return res, fmt.Errorf("remove feature dir: %w", err)
	}
	for _, m := range migs {
		if err := os.Remove(m); err != nil {
			return res, fmt.Errorf("remove migration %s: %w", m, err)
		}
	}
	unwired, err := removeFeature(appGo, modulePath, fd)
	if err != nil {
		return res, err
	}
	res.Unwired = unwired

	if _, err := exec.LookPath("gofmt"); err == nil {
		_ = exec.Command("gofmt", "-w", appGo).Run()
	}
	return res, nil
}

// featureMigrations finds migration files that belong to a feature. It matches both the
// generated `NNNN_create_<table>.sql` form and looser `NNNN_<name>...sql` forms (e.g. the
// example note feature ships as `0002_notes.sql`).
func featureMigrations(projectDir string, fd FeatureData) []string {
	migDir := filepath.Join(projectDir, "migrations")
	entries, err := os.ReadDir(migDir)
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range entries {
		name := e.Name()
		if !strings.HasSuffix(name, ".sql") {
			continue
		}
		if strings.Contains(name, "create_"+fd.Table) || strings.Contains(name, "_"+fd.Table+".") {
			out = append(out, filepath.Join(migDir, name))
		}
	}
	return out
}

// FeatureInfo describes a discovered feature for `list feature`.
type FeatureInfo struct {
	Name      string
	Mounted   bool
	Migration string // a representative migration filename, or "" if none found
}

// ListFeatures scans internal/features/* and reports each feature, whether it is mounted
// in app.go, and a migration file if one is found. Read-only.
func ListFeatures(projectDir string) ([]FeatureInfo, error) {
	if _, err := readModulePath(projectDir); err != nil {
		return nil, err
	}
	featuresDir := filepath.Join(projectDir, "internal", "features")
	entries, err := os.ReadDir(featuresDir)
	if err != nil {
		return nil, fmt.Errorf("read features dir: %w", err)
	}

	appSrc := ""
	if b, err := os.ReadFile(filepath.Join(projectDir, "internal", "app", "app.go")); err == nil {
		appSrc = string(b)
	}
	migEntries, _ := os.ReadDir(filepath.Join(projectDir, "migrations"))

	var infos []FeatureInfo
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		info := FeatureInfo{
			Name:    name,
			Mounted: strings.Contains(appSrc, fmt.Sprintf("srv.Mount(%sMod)", name)),
		}
		for _, m := range migEntries {
			if strings.Contains(m.Name(), name) {
				info.Migration = m.Name()
				break
			}
		}
		infos = append(infos, info)
	}
	return infos, nil
}
