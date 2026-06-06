package scaffold

import (
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
func newFeatureData(modulePath string, opts FeatureOptions) FeatureData {
	lower := normalizeName(opts.Name)
	table := opts.Table
	if table == "" {
		table = pluralize(lower)
	}
	return FeatureData{
		ModulePath:  modulePath,
		Lower:       lower,
		Title:       title(lower),
		TitlePlural: title(pluralize(lower)),
		Table:       table,
		IncludeAPI:  opts.IncludeAPI,
		IncludeWeb:  opts.IncludeWeb,
	}
}
