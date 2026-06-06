package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/Adhnan23/karotsgenerate/internal/scaffold"
)

// cmdRemove handles `karotsgenerate remove feature <name>`. It prints the deletion plan,
// then (unless -force) asks for confirmation before deleting the slice + migration and
// un-wiring app.go. With --dry-run it only prints the plan.
func cmdRemove(args []string) error {
	if len(args) == 0 || args[0] != "feature" {
		return fmt.Errorf("usage: karotsgenerate remove feature <name> [flags]")
	}

	fs := flag.NewFlagSet("remove feature", flag.ContinueOnError)
	table := fs.String("table", "", "DB table/route name (to locate the migration)")
	dir := fs.String("dir", ".", "project root directory")
	force := fs.Bool("force", false, "delete without confirmation")
	dryRun := fs.Bool("dry-run", false, "show what would be removed, delete nothing")

	name, flagArgs, err := splitNameAndFlags(args[1:])
	if err != nil {
		return err
	}
	if err := fs.Parse(flagArgs); err != nil {
		return err
	}
	if name == "" {
		return fmt.Errorf("usage: karotsgenerate remove feature <name> [flags]")
	}

	base := scaffold.FeatureOptions{Name: name, Table: *table, ProjectDir: *dir}

	// Compute and show the plan first.
	plan, err := scaffold.RemoveFeature(withDryRun(base))
	if err != nil {
		return err
	}
	fmt.Printf("The following will be removed for feature %q:\n%s", name, plan.FilesList())
	if plan.Unwired {
		fmt.Println("  • un-wire import + mount in internal/app/app.go")
	}

	if *dryRun {
		fmt.Println("\n(dry run — nothing deleted)")
		return nil
	}
	if !*force && !confirm("\nProceed? [y/N] ") {
		fmt.Println("aborted")
		return nil
	}

	res, err := scaffold.RemoveFeature(base)
	if err != nil {
		return err
	}
	fmt.Printf("\n✓ Removed feature %q (%d files)\n", name, len(res.Files))
	return nil
}

func withDryRun(o scaffold.FeatureOptions) scaffold.FeatureOptions {
	o.DryRun = true
	return o
}

// confirm reads a yes/no answer from stdin, defaulting to no.
func confirm(prompt string) bool {
	fmt.Print(prompt)
	r := bufio.NewReader(os.Stdin)
	line, _ := r.ReadString('\n')
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "y", "yes":
		return true
	default:
		return false
	}
}
