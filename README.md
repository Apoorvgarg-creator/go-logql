# go-logql

A pure Go query builder for [Grafana Loki's LogQL](https://grafana.com/docs/loki/latest/query/) query language.

- **Fluent, immutable builder pattern** (like [squirrel](https://github.com/Masterminds/squirrel) for SQL)
- **Zero external dependencies** — only the Go standard library
- **Generates LogQL strings only** — bring your own HTTP client
- **Safe for concurrent use** — every builder method returns a new instance

## Installation

```bash
go get github.com/Apoorvgarg-creator/go-logql
```

## Quick Start

```go
package main

import (
    "fmt"
    "time"

    logql "github.com/Apoorvgarg-creator/go-logql"
)

func main() {
    q := logql.NewLogQuery().
        Eq("job", "api").
        Eq("env", "prod").
        LineContains("error")

    fmt.Println(q.String())
    // {job="api", env="prod"} |= "error"

    m := logql.Rate(q, 5*time.Minute).
        Sum().
        By("job", "instance")

    fmt.Println(m.String())
    // sum by (job, instance) (rate({job="api", env="prod"} |= "error" [5m]))
}
```

## Usage

### Log Queries

#### Stream Selectors

```go
q := logql.NewLogQuery().
    Eq("job", "api").          // {job="api"}
    Neq("env", "dev").         // {env!="dev"}
    Re("instance", "10\\..*"). // {instance=~"10\..*"}
    Nre("method", "OPTIONS")   // {method!~"OPTIONS"}
```

#### Line Filters

```go
q := logql.NewLogQuery().
    Eq("job", "api").
    LineContains("error").       // |= "error"
    LineNotContains("debug").    // != "debug"
    LineMatch("error|warn").     // |~ "error|warn"
    LineNotMatch("trace|debug")  // !~ "trace|debug"
```

#### Parsers

```go
// JSON parser
q := logql.NewLogQuery().Eq("job", "api").JSON()
// {job="api"} | json

// JSON with specific fields
q = logql.NewLogQuery().Eq("job", "api").JSON("status", "method")
// {job="api"} | json status, method

// Logfmt parser
q = logql.NewLogQuery().Eq("job", "api").Logfmt()
// {job="api"} | logfmt

// Regexp parser
q = logql.NewLogQuery().Eq("job", "api").Regexp(`(?P<method>\w+) (?P<path>\S+)`)
// {job="api"} | regexp "(?P<method>\w+) (?P<path>\S+)"

// Pattern parser
q = logql.NewLogQuery().Eq("job", "api").Pattern("<method> <path> <status>")
// {job="api"} | pattern "<method> <path> <status>"

// Unpack
q = logql.NewLogQuery().Eq("job", "api").Unpack()
// {job="api"} | unpack
```

#### Label Filters

```go
q := logql.NewLogQuery().
    Eq("job", "api").
    JSON().
    LabelEqual("level", "error").     // | level = "error"
    LabelNotEqual("method", "GET").   // | method != "GET"
    LabelGreater("status", "400").    // | status > 400
    LabelGreaterEq("status", "400"). // | status >= 400
    LabelLess("duration", "5s").      // | duration < 5s
    LabelLessEq("duration", "10s").  // | duration <= 10s
    LabelRe("method", "GET|POST").    // | method =~ "GET|POST"
    LabelNre("path", "/health.*")     // | path !~ "/health.*"
```

#### Formatting and Pipeline Stages

```go
q := logql.NewLogQuery().
    Eq("job", "api").
    JSON().
    LineFormat("{{.msg}}").              // | line_format "{{.msg}}"
    LabelFormatEntry("dst", "src").      // | label_format dst=src
    Drop("internal_id", "trace_id").     // | drop internal_id, trace_id
    Keep("level", "msg").                // | keep level, msg
    Decolorize()                         // | decolorize
```

### Immutability

Every builder method returns a new instance. The original is never modified:

```go
base := logql.NewLogQuery().Eq("job", "api").Eq("env", "prod")

errors := base.LineContains("error")
warnings := base.LineContains("warning")

fmt.Println(base.String())     // {job="api", env="prod"}
fmt.Println(errors.String())   // {job="api", env="prod"} |= "error"
fmt.Println(warnings.String()) // {job="api", env="prod"} |= "warning"
```

### Metric Queries

#### Range Aggregations

```go
q := logql.NewLogQuery().Eq("job", "api")

logql.Rate(q, 5*time.Minute)            // rate({job="api"} [5m])
logql.CountOverTime(q, 1*time.Hour)      // count_over_time({job="api"} [1h])
logql.BytesRate(q, 5*time.Minute)        // bytes_rate({job="api"} [5m])
logql.BytesOverTime(q, 1*time.Hour)      // bytes_over_time({job="api"} [1h])
logql.AbsentOverTime(q, 5*time.Minute)   // absent_over_time({job="api"} [5m])
logql.FirstOverTime(q, 5*time.Minute)    // first_over_time({job="api"} [5m])
logql.LastOverTime(q, 5*time.Minute)     // last_over_time({job="api"} [5m])
```

#### Unwrap Range Aggregations

These require an `| unwrap <label>` stage in the log query:

```go
q := logql.NewLogQuery().Eq("job", "api").JSON().Unwrap("latency_ms")

logql.SumOverTime(q, 5*time.Minute)      // sum_over_time({...} | json | unwrap latency_ms [5m])
logql.AvgOverTime(q, 5*time.Minute)      // avg_over_time(...)
logql.MaxOverTime(q, 5*time.Minute)      // max_over_time(...)
logql.MinOverTime(q, 5*time.Minute)      // min_over_time(...)
logql.StddevOverTime(q, 5*time.Minute)   // stddev_over_time(...)
logql.StdvarOverTime(q, 5*time.Minute)   // stdvar_over_time(...)
logql.QuantileOverTime(0.95, q, 5*time.Minute) // quantile_over_time(0.95, ...)
```

#### Aggregation Operators

```go
q := logql.NewLogQuery().Eq("job", "api")
r := logql.Rate(q, 5*time.Minute)

r.Sum()                        // sum (rate(...))
r.Avg()                        // avg (rate(...))
r.Min()                        // min (rate(...))
r.Max()                        // max (rate(...))
r.Count()                      // count (rate(...))
r.Stddev()                     // stddev (rate(...))
r.Stdvar()                     // stdvar (rate(...))
r.TopK(5)                      // topk(5, rate(...))
r.BottomK(3)                   // bottomk(3, rate(...))
r.Sort()                       // sort(rate(...))
r.SortDesc()                   // sort_desc(rate(...))
```

#### Grouping and Offset

```go
r := logql.Rate(
    logql.NewLogQuery().Eq("job", "api"),
    5*time.Minute,
)

r.Sum().By("job", "instance")       // sum by (job, instance) (rate(...))
r.Sum().Without("instance")         // sum without (instance) (rate(...))
r.Offset(1 * time.Hour)             // rate(...) offset 1h
```

### Binary Expressions

```go
errors := logql.Rate(
    logql.NewLogQuery().Eq("job", "api").LineContains("error"),
    5*time.Minute,
)
total := logql.Rate(
    logql.NewLogQuery().Eq("job", "api"),
    5*time.Minute,
)

// Error rate as percentage
expr := logql.Mul(logql.Div(errors, total), &logql.Literal{Value: 100})
fmt.Println(expr.String())
// (rate({job="api"} |= "error" [5m]) / rate({job="api"} [5m])) * 100

// Comparison with bool modifier
alert := logql.CmpGt(
    logql.Rate(logql.NewLogQuery().Eq("job", "api").LineContains("error"), 5*time.Minute),
    &logql.Literal{Value: 10},
).Bool()
fmt.Println(alert.String())
// rate({job="api"} |= "error" [5m]) > bool 10
```

Available operators: `Add`, `Sub`, `Mul`, `Div`, `Mod`, `Pow`, `CmpEq`, `CmpNeq`, `CmpGt`, `CmpGte`, `CmpLt`, `CmpLte`, `And`, `Or`, `Unless`.

### Error Handling

`Build()` returns `(string, error)` and validates:

- At least one stream selector is required
- Label names cannot be empty
- Regex patterns must be valid
- Duration must be positive for range aggregations
- Quantile must be between 0 and 1
- `topk`/`bottomk` k must be > 0

```go
q := logql.NewLogQuery() // no selectors
_, err := q.Build()
// err: "logql: at least one stream selector is required"
```

`String()` calls `Build()` and panics on error — use it only when the query is known to be valid.

---

## Try It — Loki + Grafana + Live Logs (Docker)

The `examples/` directory ships with everything you need to see `go-logql` in action: a local Loki instance, Grafana with the Loki datasource pre-configured, and a **continuously running log simulator** that pushes realistic JSON logs from multiple services so you always have live data to query.

### What's Inside

```
examples/
├── docker-compose.yml                      # Loki + Grafana + log-simulator
├── loki-config.yml                         # Minimal Loki config
├── log-simulator.sh                        # Generates logs every 2s
├── grafana/provisioning/datasources/
│   └── loki.yml                            # Auto-configures Loki datasource
├── main.go                                 # Example Go program using go-logql
└── go.mod
```

### Services

| Container | Port | Description |
|---|---|---|
| **loki** | `localhost:3100` | Grafana Loki log aggregation backend |
| **grafana** | `localhost:3000` | Grafana UI (auto-login, no password needed) |
| **log-simulator** | — | Pushes JSON logs to Loki every 2 seconds |

### Simulated Log Streams

The log simulator generates realistic JSON log lines for these services:

| Stream labels | Sample log fields |
|---|---|
| `{job="api", env="prod", instance="api-1\|api-2"}` | `level`, `status`, `msg`, `latency_ms`, `method`, `path` |
| `{job="api", env="staging", instance="api-staging-1"}` | Same fields as prod api |
| `{job="web", env="prod", instance="web-1"}` | `level`, `status`, `msg`, `latency_ms` |
| `{job="auth", env="prod", instance="auth-1"}` | `level`, `status`, `msg`, `latency_ms` |
| `{job="worker", env="prod", instance="worker-1"}` | `level`, `msg`, `latency_ms`, `queue` |

Each batch includes a mix of `info`, `warn`, and `error` level logs with realistic HTTP status codes and latency values.

### 1. Start Everything

```bash
cd examples
docker compose up -d
```

Wait for all three containers to be healthy (Grafana waits for Loki automatically, the simulator waits for Loki too):

```bash
docker compose ps
```

### 2. Open Grafana Explore

1. Go to http://localhost:3000
2. Click **Explore** in the left sidebar (compass icon)
3. The **Loki** datasource is already selected
4. You should see logs flowing in immediately

### 3. Try These Queries in Grafana

Copy-paste any of these into the Explore query editor. Switch between **Logs** and **Metric** visualization modes to see different views.

**Browse all API logs:**
```
{job="api"}
```

**Filter errors only:**
```
{job="api"} |= "error"
```

**Parse JSON and filter by status:**
```
{job="api"} | json | status >= 500
```

**Errors across all services:**
```
{job=~".+"} | json | level = "error"
```

**Rate of logs per service (metric):**
```
sum by (job) (rate({job=~".+"} [1m]))
```

**Error rate per service (metric):**
```
sum by (job) (rate({job=~".+"} |~ "error" [1m]))
```

**P95 latency by service (unwrap metric):**
```
quantile_over_time(0.95, {job=~".+"} | json | unwrap latency_ms [1m]) by (job)
```

**Error rate percentage (binary expression):**
```
(sum(rate({job="api"} |= "error" [5m])) / sum(rate({job="api"} [5m]))) * 100
```

**Top 3 noisiest instances:**
```
topk(3, rate({job=~".+"} [5m]))
```

**Worker queue logs only:**
```
{job="worker"} | json | queue == "payments"
```

### 4. Run the Example Go Program

The example program uses `go-logql` to build queries programmatically and executes them against your local Loki:

```bash
cd examples
go run main.go
```

This will:
- Push an additional batch of sample logs
- Build 11 different LogQL queries using the `go-logql` builder
- Execute each query against Loki and print results

### 5. Watch Simulator Logs

To see the simulator working in real time:

```bash
docker compose logs -f log-simulator
```

### 6. Cleanup

```bash
cd examples
docker compose down -v
```

---

## API Reference

### Log Query Builder

| Method | Output |
|---|---|
| `NewLogQuery()` | Creates a new empty builder |
| `.Eq(label, value)` | `{label="value"}` |
| `.Neq(label, value)` | `{label!="value"}` |
| `.Re(label, pattern)` | `{label=~"pattern"}` |
| `.Nre(label, pattern)` | `{label!~"pattern"}` |
| `.LineContains(text)` | `\|= "text"` |
| `.LineNotContains(text)` | `!= "text"` |
| `.LineMatch(pattern)` | `\|~ "pattern"` |
| `.LineNotMatch(pattern)` | `!~ "pattern"` |
| `.JSON(labels...)` | `\| json [labels]` |
| `.Logfmt(labels...)` | `\| logfmt [labels]` |
| `.Regexp(pattern)` | `\| regexp "pattern"` |
| `.Pattern(pattern)` | `\| pattern "pattern"` |
| `.Unpack(labels...)` | `\| unpack [labels]` |
| `.LabelEqual(label, value)` | `\| label = "value"` |
| `.LabelNotEqual(label, value)` | `\| label != "value"` |
| `.LabelGreater(label, value)` | `\| label > value` |
| `.LabelGreaterEq(label, value)` | `\| label >= value` |
| `.LabelLess(label, value)` | `\| label < value` |
| `.LabelLessEq(label, value)` | `\| label <= value` |
| `.LabelRe(label, pattern)` | `\| label =~ "pattern"` |
| `.LabelNre(label, pattern)` | `\| label !~ "pattern"` |
| `.LineFormat(template)` | `\| line_format "template"` |
| `.LabelFormatEntry(dst, src)` | `\| label_format dst=src` |
| `.Drop(labels...)` | `\| drop l1, l2` |
| `.Keep(labels...)` | `\| keep l1, l2` |
| `.Decolorize()` | `\| decolorize` |
| `.Unwrap(label)` | `\| unwrap label` |
| `.Build()` | `(string, error)` |
| `.String()` | `string` (panics on error) |

### Metric Query Constructors

`Rate`, `CountOverTime`, `BytesRate`, `BytesOverTime`, `AbsentOverTime`, `FirstOverTime`, `LastOverTime`, `SumOverTime`, `AvgOverTime`, `MaxOverTime`, `MinOverTime`, `StddevOverTime`, `StdvarOverTime`, `QuantileOverTime`

### Metric Query Methods

`Sum`, `Avg`, `Min`, `Max`, `Count`, `Stddev`, `Stdvar`, `TopK`, `BottomK`, `Sort`, `SortDesc`, `By`, `Without`, `Offset`, `Build`, `String`

### Expression Constructors

`Add`, `Sub`, `Mul`, `Div`, `Mod`, `Pow`, `CmpEq`, `CmpNeq`, `CmpGt`, `CmpGte`, `CmpLt`, `CmpLte`, `And`, `Or`, `Unless`

### Expression Methods

`Bool`, `Build`, `String`

## License

MIT
