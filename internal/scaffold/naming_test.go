package scaffold

import "testing"

func TestPluralize(t *testing.T) {
	cases := map[string]string{
		"invoice":  "invoices",
		"category": "categories",
		"box":      "boxes",
		"class":    "classes",
		"dish":     "dishes",
		"day":      "days", // vowel before y -> just +s
		"note":     "notes",
	}
	for in, want := range cases {
		if got := pluralize(in); got != want {
			t.Errorf("pluralize(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestNormalizeAndTitle(t *testing.T) {
	if got := normalizeName("BlogPost"); got != "blogpost" {
		t.Errorf("normalizeName(BlogPost) = %q", got)
	}
	if got := normalizeName("user-profile"); got != "userprofile" {
		t.Errorf("normalizeName(user-profile) = %q", got)
	}
	if got := title("invoice"); got != "Invoice" {
		t.Errorf("title(invoice) = %q", got)
	}
}

func TestExportedFromColumn(t *testing.T) {
	cases := map[string]string{
		"price":      "Price",
		"in_stock":   "InStock",
		"created_by": "CreatedBy",
	}
	for in, want := range cases {
		if got := exportedFromColumn(in); got != want {
			t.Errorf("exportedFromColumn(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestParseFieldsDefault(t *testing.T) {
	fields, err := parseFields("")
	if err != nil {
		t.Fatalf("default parseFields: %v", err)
	}
	if len(fields) != 2 || fields[0].Column != "title" || fields[1].Column != "body" {
		t.Fatalf("default fields = %+v, want title/body", fields)
	}
	if fields[0].Validate != "required,max=200" {
		t.Errorf("default title validate = %q", fields[0].Validate)
	}
}

func TestParseFieldsValid(t *testing.T) {
	fields, err := parseFields("name:string, price:decimal , in_stock:bool")
	if err != nil {
		t.Fatalf("parseFields: %v", err)
	}
	if len(fields) != 3 {
		t.Fatalf("got %d fields, want 3", len(fields))
	}
	want := []Field{
		{Column: "name", GoName: "Name", GoType: "string", SQLType: "TEXT", Validate: "max=200", WebInput: "text"},
		{Column: "price", GoName: "Price", GoType: "float64", SQLType: "NUMERIC", Validate: "", WebInput: "number"},
		{Column: "in_stock", GoName: "InStock", GoType: "bool", SQLType: "BOOLEAN", Validate: "", WebInput: "checkbox"},
	}
	for i, w := range want {
		if fields[i] != w {
			t.Errorf("field[%d] = %+v, want %+v", i, fields[i], w)
		}
	}
}

func TestParseFieldsErrors(t *testing.T) {
	cases := []string{
		"name",           // no type
		"name:",          // empty type
		"name:wizard",    // unknown type
		"id:int",         // reserved
		"user_id:int",    // reserved
		"a:int,a:string", // duplicate
		"Name:string",    // not snake_case (uppercase)
		"1bad:int",       // starts with digit
	}
	for _, spec := range cases {
		if _, err := parseFields(spec); err == nil {
			t.Errorf("parseFields(%q) = nil error, want error", spec)
		}
	}
}
