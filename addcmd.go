package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/Adhnan23/karotsgenerate/internal/scaffold"
)

// cmdAdd handles `karotsgenerate add feature <name>`. It must be run from the root
// of a generated project (one containing internal/app/app.go with the wiring markers).
func cmdAdd(args []string) error {
	if len(args) == 0 || args[0] != "feature" {
		return fmt.Errorf("usage: karotsgenerate add feature <name> [flags]")
	}

	fs := flag.NewFlagSet("add feature", flag.ContinueOnError)
	noAPI := fs.Bool("no-api", false, "skip the JSON API handler for this feature")
	noWeb := fs.Bool("no-web", false, "skip the htmx web handler + views for this feature")
	table := fs.String("table", "", "DB table / route name (default: pluralized feature name)")
	dir := fs.String("dir", ".", "project root directory")

	// Separate the positional <name> from flags so they can appear in any order
	// (the stdlib flag package otherwise stops at the first positional argument).
	name, flagArgs, err := splitNameAndFlags(args[1:])
	if err != nil {
		return err
	}
	if err := fs.Parse(flagArgs); err != nil {
		return err
	}
	if name == "" {
		return fmt.Errorf("usage: karotsgenerate add feature <name> [flags]")
	}

	opts := scaffold.FeatureOptions{
		Name:       name,
		Table:      *table,
		IncludeAPI: !*noAPI,
		IncludeWeb: !*noWeb,
		ProjectDir: *dir,
	}
	res, err := scaffold.AddFeature(opts)
	if err != nil {
		return err
	}

	fmt.Printf(`
✓ Feature %q generated:
%s  wired into internal/app/app.go

Next steps:
  make generate      # render templ + swagger for the new feature
  make migrate-up    # apply the new migration (or `+"`make run`"+` does it on boot)
`, name, res.FilesList())
	return nil
}

// valueFlags are the add-feature flags that consume the following argument when
// written in the separated form (e.g. "-table foo" rather than "-table=foo").
var valueFlags = map[string]bool{
	"-table": true,
	"-dir":   true,
	"--table": true,
	"--dir":   true,
}

// splitNameAndFlags separates the positional feature <name> from flag arguments,
// allowing the name to appear in any position. The stdlib flag package stops
// parsing at the first non-flag token, so without this a command like
// `add feature widget -no-api` would treat "-no-api" as a stray positional.
//
// It returns the first non-flag token as name and everything else (preserving
// order) as flagArgs for flag.FlagSet.Parse.
func splitNameAndFlags(args []string) (name string, flagArgs []string, err error) {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "-") {
			flagArgs = append(flagArgs, arg)
			// A value-taking flag in separated form consumes the next token.
			if valueFlags[arg] && i+1 < len(args) {
				flagArgs = append(flagArgs, args[i+1])
				i++
			}
			continue
		}
		// First bare token is the feature name; any further bare token is an error.
		if name == "" {
			name = arg
			continue
		}
		return "", nil, fmt.Errorf("unexpected argument %q (only one feature name is allowed)", arg)
	}
	return name, flagArgs, nil
}
