package logql

import (
	"fmt"
	"strconv"
	"strings"
)

// Buildable is anything that can be rendered to a LogQL string.
type Buildable interface {
	Build() (string, error)
}

// Literal wraps a numeric constant for use in binary expressions.
type Literal struct {
	Value float64
}

// Build renders the literal value.
func (l *Literal) Build() (string, error) {
	// Format without unnecessary trailing zeros
	s := strconv.FormatFloat(l.Value, 'f', -1, 64)
	return s, nil
}

// String renders the literal value, panicking on error.
func (l *Literal) String() string {
	s, err := l.Build()
	if err != nil {
		panic(err)
	}
	return s
}

// Expr represents a binary expression combining two Buildable operands.
type Expr struct {
	left     Buildable
	op       string
	right    Buildable
	boolMod  bool
	err      error
}

func (e *Expr) clone() *Expr {
	return &Expr{
		left:    e.left,
		op:      e.op,
		right:   e.right,
		boolMod: e.boolMod,
		err:     e.err,
	}
}

func newExpr(op string, left, right Buildable) *Expr {
	return &Expr{
		left:  left,
		op:    op,
		right: right,
	}
}

// Add creates an addition expression: left + right.
func Add(left, right Buildable) *Expr { return newExpr("+", left, right) }

// Sub creates a subtraction expression: left - right.
func Sub(left, right Buildable) *Expr { return newExpr("-", left, right) }

// Mul creates a multiplication expression: left * right.
func Mul(left, right Buildable) *Expr { return newExpr("*", left, right) }

// Div creates a division expression: left / right.
func Div(left, right Buildable) *Expr { return newExpr("/", left, right) }

// Mod creates a modulo expression: left % right.
func Mod(left, right Buildable) *Expr { return newExpr("%", left, right) }

// Pow creates a power expression: left ^ right.
func Pow(left, right Buildable) *Expr { return newExpr("^", left, right) }

// CmpEq creates an equality comparison: left == right.
func CmpEq(left, right Buildable) *Expr { return newExpr("==", left, right) }

// CmpNeq creates a not-equal comparison: left != right.
func CmpNeq(left, right Buildable) *Expr { return newExpr("!=", left, right) }

// CmpGt creates a greater-than comparison: left > right.
func CmpGt(left, right Buildable) *Expr { return newExpr(">", left, right) }

// CmpGte creates a greater-or-equal comparison: left >= right.
func CmpGte(left, right Buildable) *Expr { return newExpr(">=", left, right) }

// CmpLt creates a less-than comparison: left < right.
func CmpLt(left, right Buildable) *Expr { return newExpr("<", left, right) }

// CmpLte creates a less-or-equal comparison: left <= right.
func CmpLte(left, right Buildable) *Expr { return newExpr("<=", left, right) }

// And creates a logical AND expression: left and right.
func And(left, right Buildable) *Expr { return newExpr("and", left, right) }

// Or creates a logical OR expression: left or right.
func Or(left, right Buildable) *Expr { return newExpr("or", left, right) }

// Unless creates a logical UNLESS expression: left unless right.
func Unless(left, right Buildable) *Expr { return newExpr("unless", left, right) }

// Bool adds the bool modifier to a comparison expression.
func (e *Expr) Bool() *Expr {
	newE := e.clone()
	newE.boolMod = true
	return newE
}

// Build renders the binary expression string.
func (e *Expr) Build() (string, error) {
	if e.err != nil {
		return "", e.err
	}

	leftStr, err := e.left.Build()
	if err != nil {
		return "", fmt.Errorf("logql: left operand: %w", err)
	}

	rightStr, err := e.right.Build()
	if err != nil {
		return "", fmt.Errorf("logql: right operand: %w", err)
	}

	// Wrap sub-expressions in parentheses for clarity
	leftStr = wrapIfExpr(e.left, leftStr)
	rightStr = wrapIfExpr(e.right, rightStr)

	var sb strings.Builder
	sb.WriteString(leftStr)
	sb.WriteByte(' ')
	sb.WriteString(e.op)
	if e.boolMod {
		sb.WriteString(" bool")
	}
	sb.WriteByte(' ')
	sb.WriteString(rightStr)

	return sb.String(), nil
}

// wrapIfExpr wraps the rendered string in parentheses if the operand is an Expr.
func wrapIfExpr(b Buildable, s string) string {
	if _, ok := b.(*Expr); ok {
		return "(" + s + ")"
	}
	return s
}

// String renders the expression string, panicking on error.
func (e *Expr) String() string {
	s, err := e.Build()
	if err != nil {
		panic(err)
	}
	return s
}
