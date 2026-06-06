package scaffold

import (
	"fmt"
	"os"
	"strings"
)

// Wiring markers placed in a generated project's internal/app/app.go. `add feature`
// inserts an import and a construct+mount block immediately above these lines.
const (
	markerImports = "// karotsgenerate:imports"
	markerMounts  = "// karotsgenerate:mounts"
)

// injectFeature edits app.go in place: it adds the feature's import path above the
// imports marker and its module construction + mount above the mounts marker. It is
// idempotent — if the feature is already wired, it returns without changing anything.
func injectFeature(appGoPath, modulePath string, fd FeatureData) error {
	raw, err := os.ReadFile(appGoPath)
	if err != nil {
		return fmt.Errorf("read %s: %w", appGoPath, err)
	}
	src := string(raw)

	if !strings.Contains(src, markerImports) || !strings.Contains(src, markerMounts) {
		return fmt.Errorf("%s is missing the karotsgenerate wiring markers; add it manually", appGoPath)
	}

	importLine := fmt.Sprintf("\t%q", modulePath+"/internal/features/"+fd.Lower)
	if strings.Contains(src, importLine) {
		return nil // already wired
	}

	mountBlock := fmt.Sprintf(
		"\t%sMod := %s.NewModule(logger, db, authMod.RequireAuth, authMod.RequireAuthWeb)\n"+
			"\tsrv.Mount(%sMod)\n",
		fd.Lower, fd.Lower, fd.Lower,
	)

	src = insertBefore(src, markerImports, importLine+"\n")
	src = insertBefore(src, markerMounts, mountBlock)

	return os.WriteFile(appGoPath, []byte(src), 0o644)
}

// insertBefore inserts text immediately before the line containing marker, preserving
// that line's leading indentation on the inserted text's first line is the caller's job.
func insertBefore(src, marker, text string) string {
	idx := strings.Index(src, marker)
	if idx < 0 {
		return src
	}
	// Back up to the start of the marker's line so we insert on its own line(s).
	lineStart := strings.LastIndexByte(src[:idx], '\n') + 1
	return src[:lineStart] + text + src[lineStart:]
}
