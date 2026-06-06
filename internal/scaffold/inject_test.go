package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const appGoFixture = `package app

import (
	"log/slog"
	// karotsgenerate:imports
)

func Build(logger *slog.Logger) {
	authMod := auth.NewModule()
	srv.Mount(authMod)
	// karotsgenerate:mounts
	_ = logger
}
`

func writeFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "app.go")
	if err := os.WriteFile(path, []byte(appGoFixture), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestInjectThenRemoveRoundTrips(t *testing.T) {
	path := writeFixture(t)
	const mod = "github.com/x/d"
	fd := FeatureData{Lower: "widget"}

	if err := injectFeature(path, mod, fd); err != nil {
		t.Fatalf("inject: %v", err)
	}
	got, _ := os.ReadFile(path)
	src := string(got)
	for _, want := range []string{
		`"github.com/x/d/internal/features/widget"`,
		"widgetMod := widget.NewModule(",
		"srv.Mount(widgetMod)",
	} {
		if !strings.Contains(src, want) {
			t.Errorf("after inject, missing %q in:\n%s", want, src)
		}
	}

	removed, err := removeFeature(path, mod, fd)
	if err != nil {
		t.Fatalf("remove: %v", err)
	}
	if !removed {
		t.Error("removeFeature reported nothing removed")
	}
	got, _ = os.ReadFile(path)
	if string(got) != appGoFixture {
		t.Errorf("round-trip did not restore original:\n--- got ---\n%s\n--- want ---\n%s", got, appGoFixture)
	}
}

func TestInjectIsIdempotent(t *testing.T) {
	path := writeFixture(t)
	const mod = "github.com/x/d"
	fd := FeatureData{Lower: "widget"}

	if err := injectFeature(path, mod, fd); err != nil {
		t.Fatal(err)
	}
	after1, _ := os.ReadFile(path)
	if err := injectFeature(path, mod, fd); err != nil {
		t.Fatal(err)
	}
	after2, _ := os.ReadFile(path)
	if string(after1) != string(after2) {
		t.Error("second inject changed the file; want idempotent")
	}
}

func TestRemoveWhenAbsentIsNoOp(t *testing.T) {
	path := writeFixture(t)
	removed, err := removeFeature(path, "github.com/x/d", FeatureData{Lower: "ghost"})
	if err != nil {
		t.Fatal(err)
	}
	if removed {
		t.Error("removeFeature reported a removal for an absent feature")
	}
	got, _ := os.ReadFile(path)
	if string(got) != appGoFixture {
		t.Error("no-op remove modified the file")
	}
}

func TestInjectMissingMarkersErrors(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.go")
	_ = os.WriteFile(path, []byte("package app\n"), 0o644)
	if err := injectFeature(path, "github.com/x/d", FeatureData{Lower: "widget"}); err == nil {
		t.Error("expected error when markers are missing")
	}
}
