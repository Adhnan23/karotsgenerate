package main

import (
	"reflect"
	"testing"
)

func TestSplitNameAndFlags(t *testing.T) {
	cases := []struct {
		name      string
		args      []string
		wantName  string
		wantFlags []string
		wantErr   bool
	}{
		{"bare name", []string{"widget"}, "widget", nil, false},
		{"name then bool flag", []string{"widget", "-no-api"}, "widget", []string{"-no-api"}, false},
		{"bool flag then name", []string{"-no-api", "widget"}, "widget", []string{"-no-api"}, false},
		{"value flag separated", []string{"widget", "-table", "foos"}, "widget", []string{"-table", "foos"}, false},
		{"value flag equals", []string{"widget", "-table=foos"}, "widget", []string{"-table=foos"}, false},
		{"fields flag consumes value", []string{"-fields", "a:int,b:string", "widget"}, "widget", []string{"-fields", "a:int,b:string"}, false},
		{"two bare tokens error", []string{"a", "b"}, "", nil, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			name, flags, err := splitNameAndFlags(tc.args)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("want error, got name=%q flags=%v", name, flags)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if name != tc.wantName {
				t.Errorf("name = %q, want %q", name, tc.wantName)
			}
			if !reflect.DeepEqual(flags, tc.wantFlags) {
				t.Errorf("flags = %v, want %v", flags, tc.wantFlags)
			}
		})
	}
}
