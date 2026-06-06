// Command karotsgenerate scaffolds new fullstack Go projects — and individual
// features within them — that follow the architecture of the go-learning reference
// app (Echo · templ · htmx/Alpine · sqlx/pgx · Postgres · goose · JWT cookie auth ·
// dual web+API surfaces · Swagger · layered testing).
//
// Usage:
//
//	karotsgenerate new                  # interactive: scaffold a brand-new project
//	karotsgenerate add feature <name>   # add a vertical-slice feature to a project
//	karotsgenerate version
//
// Run `karotsgenerate <cmd> -h` for the flags each command accepts.
package main

import (
	"fmt"
	"os"
)

// version is bumped manually; surfaced via `karotsgenerate version`.
const version = "0.1.0"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		usage()
		return nil
	}
	switch args[0] {
	case "new":
		return cmdNew(args[1:])
	case "add":
		return cmdAdd(args[1:])
	case "version", "-v", "--version":
		fmt.Println("karotsgenerate", version)
		return nil
	case "help", "-h", "--help":
		usage()
		return nil
	default:
		usage()
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func usage() {
	fmt.Print(`karotsgenerate — fullstack Go project & feature scaffolder

Commands:
  new                    Scaffold a new project (interactive, or pass -yes + flags)
  add feature <name>     Add a vertical-slice feature to the current project
  version                Print the version
  help                   Show this help

Examples:
  karotsgenerate new
  karotsgenerate new -yes -module github.com/you/app -dir ./app
  karotsgenerate add feature invoice
  karotsgenerate add feature comment -no-api

Run "karotsgenerate new -h" or "karotsgenerate add feature -h" for flags.
`)
}
