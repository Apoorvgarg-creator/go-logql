package logql_test

import (
	"fmt"
	"time"

	logql "github.com/Apoorvgarg-creator/go-logql"
)

func ExampleLogQuery_simple() {
	q := logql.NewLogQuery().
		Eq("job", "api").
		Eq("env", "prod").
		LineContains("error")

	fmt.Println(q.String())
	// Output: {job="api", env="prod"} |= "error"
}

func ExampleLogQuery_parsedAndFiltered() {
	q := logql.NewLogQuery().
		Eq("job", "api").
		JSON().
		LabelEqual("level", "error").
		LabelGreaterEq("status", "400")

	fmt.Println(q.String())
	// Output: {job="api"} | json | level = "error" | status >= 400
}

func ExampleMetricQuery_rate() {
	q := logql.NewLogQuery().
		Eq("job", "api").
		LineContains("error")

	m := logql.Rate(q, 5*time.Minute).
		Sum().
		By("job", "instance")

	fmt.Println(m.String())
	// Output: sum by (job, instance) (rate({job="api"} |= "error" [5m]))
}

func ExampleMetricQuery_quantile() {
	q := logql.NewLogQuery().
		Eq("job", "api").
		JSON().
		Unwrap("latency_ms")

	m := logql.QuantileOverTime(0.95, q, 5*time.Minute)

	fmt.Println(m.String())
	// Output: quantile_over_time(0.95, {job="api"} | json | unwrap latency_ms [5m])
}

func ExampleExpr_errorRate() {
	errors := logql.Rate(
		logql.NewLogQuery().Eq("job", "api").LineContains("error"),
		5*time.Minute,
	)
	total := logql.Rate(
		logql.NewLogQuery().Eq("job", "api"),
		5*time.Minute,
	)

	expr := logql.Mul(
		logql.Div(errors, total),
		&logql.Literal{Value: 100},
	)

	fmt.Println(expr.String())
	// Output: (rate({job="api"} |= "error" [5m]) / rate({job="api"} [5m])) * 100
}
