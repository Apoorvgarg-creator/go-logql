package logql

import (
	"testing"
	"time"
)

func TestExpr(t *testing.T) {
	tests := []struct {
		name    string
		build   func() *Expr
		want    string
		wantErr bool
	}{
		{
			name: "addition",
			build: func() *Expr {
				a := Rate(NewLogQuery().Eq("job", "api"), 5*time.Minute)
				b := Rate(NewLogQuery().Eq("job", "web"), 5*time.Minute)
				return Add(a, b)
			},
			want: `rate({job="api"} [5m]) + rate({job="web"} [5m])`,
		},
		{
			name: "subtraction",
			build: func() *Expr {
				a := Rate(NewLogQuery().Eq("job", "api"), 5*time.Minute)
				b := Rate(NewLogQuery().Eq("job", "web"), 5*time.Minute)
				return Sub(a, b)
			},
			want: `rate({job="api"} [5m]) - rate({job="web"} [5m])`,
		},
		{
			name: "multiplication",
			build: func() *Expr {
				a := Rate(NewLogQuery().Eq("job", "api"), 5*time.Minute)
				return Mul(a, &Literal{Value: 100})
			},
			want: `rate({job="api"} [5m]) * 100`,
		},
		{
			name: "division",
			build: func() *Expr {
				a := Rate(NewLogQuery().Eq("job", "api").LineContains("error"), 5*time.Minute)
				b := Rate(NewLogQuery().Eq("job", "api"), 5*time.Minute)
				return Div(a, b)
			},
			want: `rate({job="api"} |= "error" [5m]) / rate({job="api"} [5m])`,
		},
		{
			name: "modulo",
			build: func() *Expr {
				a := Rate(NewLogQuery().Eq("job", "api"), 5*time.Minute)
				return Mod(a, &Literal{Value: 10})
			},
			want: `rate({job="api"} [5m]) % 10`,
		},
		{
			name: "power",
			build: func() *Expr {
				a := Rate(NewLogQuery().Eq("job", "api"), 5*time.Minute)
				return Pow(a, &Literal{Value: 2})
			},
			want: `rate({job="api"} [5m]) ^ 2`,
		},
		{
			name: "comparison equal",
			build: func() *Expr {
				a := Rate(NewLogQuery().Eq("job", "api"), 5*time.Minute)
				return CmpEq(a, &Literal{Value: 0})
			},
			want: `rate({job="api"} [5m]) == 0`,
		},
		{
			name: "comparison not equal",
			build: func() *Expr {
				a := Rate(NewLogQuery().Eq("job", "api"), 5*time.Minute)
				return CmpNeq(a, &Literal{Value: 0})
			},
			want: `rate({job="api"} [5m]) != 0`,
		},
		{
			name: "comparison greater",
			build: func() *Expr {
				a := Rate(NewLogQuery().Eq("job", "api"), 5*time.Minute)
				return CmpGt(a, &Literal{Value: 10})
			},
			want: `rate({job="api"} [5m]) > 10`,
		},
		{
			name: "comparison greater or equal",
			build: func() *Expr {
				a := Rate(NewLogQuery().Eq("job", "api"), 5*time.Minute)
				return CmpGte(a, &Literal{Value: 10})
			},
			want: `rate({job="api"} [5m]) >= 10`,
		},
		{
			name: "comparison less",
			build: func() *Expr {
				a := Rate(NewLogQuery().Eq("job", "api"), 5*time.Minute)
				return CmpLt(a, &Literal{Value: 10})
			},
			want: `rate({job="api"} [5m]) < 10`,
		},
		{
			name: "comparison less or equal",
			build: func() *Expr {
				a := Rate(NewLogQuery().Eq("job", "api"), 5*time.Minute)
				return CmpLte(a, &Literal{Value: 10})
			},
			want: `rate({job="api"} [5m]) <= 10`,
		},
		{
			name: "bool modifier",
			build: func() *Expr {
				a := Rate(NewLogQuery().Eq("job", "api"), 5*time.Minute)
				return CmpGt(a, &Literal{Value: 10}).Bool()
			},
			want: `rate({job="api"} [5m]) > bool 10`,
		},
		{
			name: "logical and",
			build: func() *Expr {
				a := Rate(NewLogQuery().Eq("job", "api"), 5*time.Minute)
				b := Rate(NewLogQuery().Eq("job", "web"), 5*time.Minute)
				return And(a, b)
			},
			want: `rate({job="api"} [5m]) and rate({job="web"} [5m])`,
		},
		{
			name: "logical or",
			build: func() *Expr {
				a := Rate(NewLogQuery().Eq("job", "api"), 5*time.Minute)
				b := Rate(NewLogQuery().Eq("job", "web"), 5*time.Minute)
				return Or(a, b)
			},
			want: `rate({job="api"} [5m]) or rate({job="web"} [5m])`,
		},
		{
			name: "logical unless",
			build: func() *Expr {
				a := Rate(NewLogQuery().Eq("job", "api"), 5*time.Minute)
				b := Rate(NewLogQuery().Eq("job", "web"), 5*time.Minute)
				return Unless(a, b)
			},
			want: `rate({job="api"} [5m]) unless rate({job="web"} [5m])`,
		},
		{
			name: "nested expression with parens",
			build: func() *Expr {
				a := Rate(NewLogQuery().Eq("job", "api"), 5*time.Minute)
				b := Rate(NewLogQuery().Eq("job", "web"), 5*time.Minute)
				c := &Literal{Value: 100}
				return Div(Add(a, b), c)
			},
			want: `(rate({job="api"} [5m]) + rate({job="web"} [5m])) / 100`,
		},
		{
			name: "error rate percentage",
			build: func() *Expr {
				errors := Rate(NewLogQuery().Eq("job", "api").LineContains("error"), 5*time.Minute)
				total := Rate(NewLogQuery().Eq("job", "api"), 5*time.Minute)
				return Mul(Div(errors, total), &Literal{Value: 100})
			},
			want: `(rate({job="api"} |= "error" [5m]) / rate({job="api"} [5m])) * 100`,
		},
		{
			name: "literal float",
			build: func() *Expr {
				a := Rate(NewLogQuery().Eq("job", "api"), 5*time.Minute)
				return Mul(a, &Literal{Value: 1.5})
			},
			want: `rate({job="api"} [5m]) * 1.5`,
		},
		{
			name: "left operand error propagates",
			build: func() *Expr {
				a := Rate(NewLogQuery(), 5*time.Minute) // no selectors
				b := Rate(NewLogQuery().Eq("job", "api"), 5*time.Minute)
				return Add(a, b)
			},
			wantErr: true,
		},
		{
			name: "right operand error propagates",
			build: func() *Expr {
				a := Rate(NewLogQuery().Eq("job", "api"), 5*time.Minute)
				b := Rate(NewLogQuery(), 5*time.Minute) // no selectors
				return Add(a, b)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.build().Build()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Build() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("Build() unexpected error: %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("Build() =\n  %q\nwant\n  %q", got, tt.want)
			}
		})
	}
}

func TestExprString(t *testing.T) {
	a := Rate(NewLogQuery().Eq("job", "api"), 5*time.Minute)
	e := CmpGt(a, &Literal{Value: 10})
	s := e.String()
	if s != `rate({job="api"} [5m]) > 10` {
		t.Errorf("String() = %q", s)
	}
}

func TestExprStringPanicsOnError(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("String() should panic on error, but didn't")
		}
	}()
	a := Rate(NewLogQuery(), 5*time.Minute)
	_ = Add(a, &Literal{Value: 1}).String()
}

func TestLiteral(t *testing.T) {
	tests := []struct {
		value float64
		want  string
	}{
		{100, "100"},
		{0, "0"},
		{1.5, "1.5"},
		{0.001, "0.001"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			l := &Literal{Value: tt.value}
			got, err := l.Build()
			if err != nil {
				t.Errorf("Build() error: %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("Build() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExprBoolImmutability(t *testing.T) {
	a := Rate(NewLogQuery().Eq("job", "api"), 5*time.Minute)
	base := CmpGt(a, &Literal{Value: 10})
	withBool := base.Bool()

	baseStr, _ := base.Build()
	boolStr, _ := withBool.Build()

	if baseStr != `rate({job="api"} [5m]) > 10` {
		t.Errorf("base mutated: %q", baseStr)
	}
	if boolStr != `rate({job="api"} [5m]) > bool 10` {
		t.Errorf("bool wrong: %q", boolStr)
	}
}
