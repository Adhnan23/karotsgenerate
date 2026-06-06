package scaffold

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func fullAnswers(dir string) Answers {
	return Answers{
		ModulePath:     "github.com/example/golden",
		ProjectName:    "golden",
		TargetDir:      dir,
		DBName:         "golden",
		DBPort:         "5432",
		AppPort:        "8080",
		JWTSecret:      "test-secret",
		IncludeNote:    true,
		IncludeSwagger: true,
		IncludeDocker:  true,
		IncludeCI:      true,
	}
}

// gofmtClean fails the test if any generated .go file under dir is unparseable or
// unformatted — a fast syntax gate that catches template breakage without a full build.
func gofmtClean(t *testing.T, dir string) {
	t.Helper()
	out, err := exec.Command("gofmt", "-l", dir).CombinedOutput()
	if err != nil {
		t.Fatalf("gofmt error: %v\n%s", err, out)
	}
	if strings.TrimSpace(string(out)) != "" {
		t.Fatalf("gofmt reported unformatted/unparseable files:\n%s", out)
	}
}

func mustExist(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected file to exist: %s", path)
	}
}

func mustNotExist(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err == nil {
		t.Errorf("expected file to be absent: %s", path)
	}
}

func TestGenerateProjectFull(t *testing.T) {
	dir := t.TempDir()
	a := fullAnswers(dir)
	if _, err := GenerateProject(a, false, false); err != nil {
		t.Fatalf("GenerateProject: %v", err)
	}
	mustExist(t, filepath.Join(dir, "go.mod"))
	mustExist(t, filepath.Join(dir, "internal", "app", "app.go"))
	mustExist(t, filepath.Join(dir, "internal", "features", "note", "model.go"))
	mustExist(t, filepath.Join(dir, "Dockerfile"))
	mustExist(t, filepath.Join(dir, ".github", "workflows", "ci.yml"))
	gofmtClean(t, dir)
}

func TestGenerateProjectMinimal(t *testing.T) {
	dir := t.TempDir()
	a := fullAnswers(dir)
	a.IncludeNote = false
	a.IncludeSwagger = false
	a.IncludeDocker = false
	a.IncludeCI = false
	if _, err := GenerateProject(a, false, false); err != nil {
		t.Fatalf("GenerateProject: %v", err)
	}
	mustExist(t, filepath.Join(dir, "internal", "features", "auth", "service.go"))
	mustNotExist(t, filepath.Join(dir, "internal", "features", "note"))
	mustNotExist(t, filepath.Join(dir, "Dockerfile"))
	mustNotExist(t, filepath.Join(dir, ".github", "workflows", "ci.yml"))
	gofmtClean(t, dir)
}

func TestGenerateProjectDryRunWritesNothing(t *testing.T) {
	dir := t.TempDir()
	a := fullAnswers(dir)
	count, err := GenerateProject(a, false, true)
	if err != nil {
		t.Fatalf("GenerateProject dry-run: %v", err)
	}
	if count == 0 {
		t.Error("dry-run reported 0 files")
	}
	if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
		t.Error("dry-run wrote go.mod")
	}
}

func TestAddAndRemoveFeatureWithFields(t *testing.T) {
	dir := t.TempDir()
	if _, err := GenerateProject(fullAnswers(dir), false, false); err != nil {
		t.Fatalf("GenerateProject: %v", err)
	}

	add, err := AddFeature(FeatureOptions{
		Name:       "order",
		Fields:     "sku:string,qty:int",
		IncludeAPI: true,
		IncludeWeb: true,
		ProjectDir: dir,
	})
	if err != nil {
		t.Fatalf("AddFeature: %v", err)
	}
	gofmtClean(t, dir)

	model, _ := os.ReadFile(filepath.Join(dir, "internal", "features", "order", "model.go"))
	for _, want := range []string{"Sku string", "Qty int64"} {
		if !strings.Contains(string(model), want) {
			t.Errorf("order model.go missing %q", want)
		}
	}

	appGo, _ := os.ReadFile(filepath.Join(dir, "internal", "app", "app.go"))
	for _, want := range []string{
		`"github.com/example/golden/internal/features/order"`,
		"srv.Mount(orderMod)",
	} {
		if !strings.Contains(string(appGo), want) {
			t.Errorf("app.go missing wiring %q", want)
		}
	}

	migs, _ := filepath.Glob(filepath.Join(dir, "migrations", "*_create_orders.sql"))
	if len(migs) != 1 {
		t.Errorf("expected 1 orders migration, found %d", len(migs))
	}
	_ = add

	// Now remove it and confirm reversal.
	if _, err := RemoveFeature(FeatureOptions{Name: "order", ProjectDir: dir}); err != nil {
		t.Fatalf("RemoveFeature: %v", err)
	}
	mustNotExist(t, filepath.Join(dir, "internal", "features", "order"))
	appGo, _ = os.ReadFile(filepath.Join(dir, "internal", "app", "app.go"))
	if strings.Contains(string(appGo), "srv.Mount(orderMod)") {
		t.Error("app.go still references orderMod after removal")
	}
	if m, _ := filepath.Glob(filepath.Join(dir, "migrations", "*_create_orders.sql")); len(m) != 0 {
		t.Error("orders migration not deleted")
	}
}
