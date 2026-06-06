package scaffold

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// normalizeName lowercases and strips a raw feature name to a safe identifier base,
// e.g. "BlogPost" -> "blogpost", "user-profile" -> "userprofile".
func normalizeName(raw string) string {
	var b strings.Builder
	for _, r := range raw {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(unicode.ToLower(r))
		}
	}
	return b.String()
}

// title returns the exported form: "invoice" -> "Invoice".
func title(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

// pluralize applies naive English pluralization rules — enough for table/route names.
// Override with the -table flag when it guesses wrong.
func pluralize(s string) string {
	switch {
	case s == "":
		return s
	case strings.HasSuffix(s, "y") && !endsInVowelY(s):
		return s[:len(s)-1] + "ies"
	case strings.HasSuffix(s, "s"),
		strings.HasSuffix(s, "x"),
		strings.HasSuffix(s, "z"),
		strings.HasSuffix(s, "ch"),
		strings.HasSuffix(s, "sh"):
		return s + "es"
	default:
		return s + "s"
	}
}

// endsInVowelY reports whether the char before a trailing "y" is a vowel (e.g. "day"),
// in which case we add "s" rather than "ies".
func endsInVowelY(s string) bool {
	if len(s) < 2 {
		return false
	}
	switch s[len(s)-2] {
	case 'a', 'e', 'i', 'o', 'u':
		return true
	}
	return false
}

// newFeatureData builds the template data from options + the project's module path.
// It returns an error if the -fields spec is malformed (callers surface it to the user).
func newFeatureData(modulePath string, opts FeatureOptions) (FeatureData, error) {
	lower := normalizeName(opts.Name)
	table := opts.Table
	if table == "" {
		table = pluralize(lower)
	}
	fields, err := parseFields(opts.Fields)
	if err != nil {
		return FeatureData{}, err
	}
	return FeatureData{
		ModulePath:  modulePath,
		Lower:       lower,
		Title:       title(lower),
		TitlePlural: title(pluralize(lower)),
		Table:       table,
		Fields:      fields,
		IncludeAPI:  opts.IncludeAPI,
		IncludeWeb:  opts.IncludeWeb,
	}, nil
}

// fieldType describes how a -fields type token maps onto Go, SQL, validation, and the
// web form input. Keep this table the single source of truth for supported types.
type fieldType struct {
	goType   string
	sqlType  string
	validate string
	webInput string
}

var fieldTypes = map[string]fieldType{
	"string":  {"string", "TEXT", "max=200", "text"},
	"text":    {"string", "TEXT", "max=10000", "textarea"},
	"int":     {"int64", "BIGINT", "", "number"},
	"bool":    {"bool", "BOOLEAN", "", "checkbox"},
	"decimal": {"float64", "NUMERIC", "", "number"},
	"time":    {"time.Time", "TIMESTAMPTZ", "", "datetime"},
}

// columnRe matches a snake_case SQL/identifier column name.
var columnRe = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// reservedColumns are managed by the scaffold itself and may not be redefined.
var reservedColumns = map[string]bool{
	"id": true, "user_id": true, "created_at": true, "updated_at": true,
}

// parseFields turns a "col:type,col:type" spec into ordered Field values. An empty spec
// yields the default title/body pair, which reproduces the original note-style feature.
func parseFields(spec string) ([]Field, error) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return []Field{
			{Column: "title", GoName: "Title", GoType: "string", SQLType: "TEXT", Validate: "required,max=200", WebInput: "text"},
			{Column: "body", GoName: "Body", GoType: "string", SQLType: "TEXT", Validate: "max=10000", WebInput: "textarea"},
		}, nil
	}

	var fields []Field
	seen := map[string]bool{}
	for _, part := range strings.Split(spec, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		name, typ, ok := strings.Cut(part, ":")
		name = strings.TrimSpace(name)
		typ = strings.TrimSpace(strings.ToLower(typ))
		if !ok || name == "" || typ == "" {
			return nil, fmt.Errorf("invalid field %q (want name:type, e.g. price:decimal)", part)
		}
		if !columnRe.MatchString(name) {
			return nil, fmt.Errorf("invalid field name %q (use snake_case: a-z, 0-9, _)", name)
		}
		if reservedColumns[name] {
			return nil, fmt.Errorf("field name %q is reserved (id/user_id/created_at/updated_at are added automatically)", name)
		}
		if seen[name] {
			return nil, fmt.Errorf("duplicate field %q", name)
		}
		ft, known := fieldTypes[typ]
		if !known {
			return nil, fmt.Errorf("unknown type %q for field %q (supported: string, text, int, bool, decimal, time)", typ, name)
		}
		seen[name] = true
		fields = append(fields, Field{
			Column:   name,
			GoName:   exportedFromColumn(name),
			GoType:   ft.goType,
			SQLType:  ft.sqlType,
			Validate: ft.validate,
			WebInput: ft.webInput,
		})
	}
	if len(fields) == 0 {
		return nil, fmt.Errorf("no valid fields parsed from %q", spec)
	}
	return fields, nil
}

// exportedFromColumn converts a snake_case column to an exported CamelCase Go name,
// e.g. "in_stock" -> "InStock", "price" -> "Price".
func exportedFromColumn(col string) string {
	parts := strings.Split(col, "_")
	var b strings.Builder
	for _, p := range parts {
		if p == "" {
			continue
		}
		b.WriteString(title(p))
	}
	return b.String()
}
