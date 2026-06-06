package main

import (
	"flag"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// toolCheck is a required/recommended external tool and the hint for installing it.
type toolCheck struct {
	name string
	hint string
}

var doctorTools = []toolCheck{
	{"go", "https://go.dev/dl/ (Go 1.26+ required)"},
	{"templ", "go install github.com/a-h/templ/cmd/templ@latest"},
	{"swag", "go install github.com/swaggo/swag/cmd/swag@latest"},
	{"goose", "go install github.com/pressly/goose/v3/cmd/goose@latest"},
	{"git", "https://git-scm.com/downloads"},
	{"docker", "https://docs.docker.com/get-docker/"},
}

// cmdDoctor reports which required tools are on PATH and whether Go is new enough. It only
// fails (non-zero exit) when Go itself is missing or too old; the rest are warnings.
func cmdDoctor(args []string) error {
	fs := flag.NewFlagSet("doctor", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}

	fmt.Println("karotsgenerate doctor — checking required tools:")
	fmt.Println()

	goOK := true
	for _, t := range doctorTools {
		path, err := exec.LookPath(t.name)
		if err != nil {
			fmt.Printf("  ✗ %-8s not found        — install: %s\n", t.name, t.hint)
			if t.name == "go" {
				goOK = false
			}
			continue
		}
		if t.name == "go" {
			ver, ok := goVersionOK()
			if !ok {
				fmt.Printf("  ✗ %-8s %s — need Go 1.26+\n", t.name, ver)
				goOK = false
				continue
			}
			fmt.Printf("  ✓ %-8s %s (%s)\n", t.name, path, ver)
			continue
		}
		fmt.Printf("  ✓ %-8s %s\n", t.name, path)
	}

	fmt.Println()
	if !goOK {
		return fmt.Errorf("Go 1.26+ is required")
	}
	fmt.Println("templ + swag are needed for `make generate`; goose for migrations; docker for `make db-up`.")
	return nil
}

var goVerRe = regexp.MustCompile(`go(\d+)\.(\d+)`)

// goVersionOK parses `go version` and reports whether it is >= 1.26.
func goVersionOK() (string, bool) {
	out, err := exec.Command("go", "version").Output()
	if err != nil {
		return "", false
	}
	m := goVerRe.FindStringSubmatch(strings.TrimSpace(string(out)))
	if m == nil {
		return strings.TrimSpace(string(out)), false
	}
	major, _ := strconv.Atoi(m[1])
	minor, _ := strconv.Atoi(m[2])
	ok := major > 1 || (major == 1 && minor >= 26)
	return fmt.Sprintf("go%s.%s", m[1], m[2]), ok
}
