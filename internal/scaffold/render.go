package scaffold

import (
	"bytes"
	"embed"
	"fmt"
	"text/template"
)

// templatesFS holds the embedded template trees. `all:` makes go:embed include
// dotfiles (.env.example.tmpl, .gitignore, .github/...) which it skips by default.
//
//go:embed all:templates
var templatesFS embed.FS

// renderBytes executes a single template's content against data. We use custom
// delimiters [[ ]] so the templates can contain raw Go ({}) and templ ({ }, @...)
// syntax without escaping — only our own [[.Field]] / [[if]] markers are special.
func renderBytes(name string, content []byte, data any) ([]byte, error) {
	t, err := template.New(name).
		Delims("[[", "]]").
		Option("missingkey=error").
		Parse(string(content))
	if err != nil {
		return nil, fmt.Errorf("parse template %s: %w", name, err)
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("render template %s: %w", name, err)
	}
	return buf.Bytes(), nil
}
