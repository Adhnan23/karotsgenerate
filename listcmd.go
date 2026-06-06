package main

import (
	"flag"
	"fmt"

	"github.com/Adhnan23/karotsgenerate/internal/scaffold"
)

// cmdList handles `karotsgenerate list feature`, reporting each feature, whether it is
// mounted in app.go, and a migration file if one is found. Read-only.
func cmdList(args []string) error {
	// The "feature" word is optional sugar: `list` and `list feature` are equivalent.
	if len(args) > 0 && args[0] == "feature" {
		args = args[1:]
	}

	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	dir := fs.String("dir", ".", "project root directory")
	if err := fs.Parse(args); err != nil {
		return err
	}

	infos, err := scaffold.ListFeatures(*dir)
	if err != nil {
		return err
	}
	if len(infos) == 0 {
		fmt.Println("no features found")
		return nil
	}

	fmt.Println("Features:")
	for _, fi := range infos {
		mounted := "mounted"
		if !fi.Mounted {
			mounted = "NOT mounted"
		}
		mig := fi.Migration
		if mig == "" {
			mig = "(no migration)"
		}
		fmt.Printf("  • %-14s %-12s %s\n", fi.Name, mounted, mig)
	}
	return nil
}
