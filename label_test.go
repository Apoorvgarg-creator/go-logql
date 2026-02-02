package logql

import "testing"

func TestMatchOpValues(t *testing.T) {
	tests := []struct {
		name string
		op   MatchOp
		want string
	}{
		{"Equal", MatchEqual, "="},
		{"NotEqual", MatchNotEqual, "!="},
		{"Regexp", MatchRegexp, "=~"},
		{"NotRegexp", MatchNotRegexp, "!~"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.op) != tt.want {
				t.Errorf("MatchOp %s = %q, want %q", tt.name, string(tt.op), tt.want)
			}
		})
	}
}

func TestFilterOpValues(t *testing.T) {
	tests := []struct {
		name string
		op   FilterOp
		want string
	}{
		{"Equal", FilterEqual, "=="},
		{"NotEqual", FilterNotEqual, "!="},
		{"Greater", FilterGreater, ">"},
		{"GreaterEqual", FilterGreaterEqual, ">="},
		{"Less", FilterLess, "<"},
		{"LessEqual", FilterLessEqual, "<="},
		{"Regexp", FilterRegexp, "=~"},
		{"NotRegexp", FilterNotRegexp, "!~"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.op) != tt.want {
				t.Errorf("FilterOp %s = %q, want %q", tt.name, string(tt.op), tt.want)
			}
		})
	}
}

func TestLabelMatcher(t *testing.T) {
	m := LabelMatcher{Label: "job", Op: MatchEqual, Value: "api"}
	if m.Label != "job" {
		t.Errorf("Label = %q, want %q", m.Label, "job")
	}
	if m.Op != MatchEqual {
		t.Errorf("Op = %q, want %q", m.Op, MatchEqual)
	}
	if m.Value != "api" {
		t.Errorf("Value = %q, want %q", m.Value, "api")
	}
}

func TestLabelFilter(t *testing.T) {
	f := LabelFilter{Label: "status", Op: FilterGreaterEqual, Value: "400"}
	if f.Label != "status" {
		t.Errorf("Label = %q, want %q", f.Label, "status")
	}
	if f.Op != FilterGreaterEqual {
		t.Errorf("Op = %q, want %q", f.Op, FilterGreaterEqual)
	}
	if f.Value != "400" {
		t.Errorf("Value = %q, want %q", f.Value, "400")
	}
}
