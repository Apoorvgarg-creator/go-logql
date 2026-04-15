// Example program demonstrating go-logql query builder with a local Loki instance.
//
// Prerequisites:
//
//	docker compose up -d   (starts Loki on :3100 and Grafana on :3000)
//
// Run:
//
//	go run main.go
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	logql "github.com/Apoorvgarg-creator/go-logql"
)

const lokiURL = "http://localhost:3100"

func main() {
	// ------------------------------------------------------------------
	// Step 1: Push sample logs to Loki
	// ------------------------------------------------------------------
	fmt.Println("=== Pushing sample logs to Loki ===")
	if err := pushSampleLogs(); err != nil {
		log.Fatalf("Failed to push logs: %v", err)
	}
	fmt.Println("Logs pushed successfully.")

	// Give Loki a moment to index
	time.Sleep(2 * time.Second)

	// ------------------------------------------------------------------
	// Step 2: Build and execute queries
	// ------------------------------------------------------------------

	// Query 1: Simple log query — find all error logs from the api job
	fmt.Println("\n=== Query 1: Error logs from api job ===")
	q1 := logql.NewLogQuery().
		Eq("job", "api").
		LineContains("error")
	runLogQuery(q1)

	// Query 2: Parsed and filtered — JSON parse, then filter by status >= 500
	fmt.Println("\n=== Query 2: Parsed logs with status >= 500 ===")
	q2 := logql.NewLogQuery().
		Eq("job", "api").
		JSON().
		LabelGreaterEq("status", "500")
	runLogQuery(q2)

	// Query 3: Regex line filter — match error or warn
	fmt.Println("\n=== Query 3: Logs matching error|warn ===")
	q3 := logql.NewLogQuery().
		Eq("job", "api").
		LineMatch("error|warn")
	runLogQuery(q3)

	// Query 4: Multiple stream selectors with label filter
	fmt.Println("\n=== Query 4: Prod api logs with level=error ===")
	q4 := logql.NewLogQuery().
		Eq("job", "api").
		Eq("env", "prod").
		JSON().
		LabelEqual("level", "error")
	runLogQuery(q4)

	// Query 5: Metric query — rate of errors over 10 minutes
	fmt.Println("\n=== Query 5: Rate of api errors [10m] ===")
	m1 := logql.Rate(
		logql.NewLogQuery().Eq("job", "api").LineContains("error"),
		10*time.Minute,
	)
	runMetricQuery(m1)

	// Query 6: Aggregated metric — sum rate by job
	fmt.Println("\n=== Query 6: Sum rate by job [10m] ===")
	m2 := logql.Rate(
		logql.NewLogQuery().Re("job", "api|web"),
		10*time.Minute,
	).Sum().By("job")
	runMetricQuery(m2)

	// Query 7: count_over_time
	fmt.Println("\n=== Query 7: Count of all logs over 10m ===")
	m3 := logql.CountOverTime(
		logql.NewLogQuery().Eq("job", "api"),
		10*time.Minute,
	)
	runMetricQuery(m3)

	// Query 8: TopK
	fmt.Println("\n=== Query 8: Top 3 instances by rate [10m] ===")
	m4 := logql.Rate(
		logql.NewLogQuery().Re("job", "api|web"),
		10*time.Minute,
	).TopK(3)
	runMetricQuery(m4)

	// Query 9: Binary expression — error rate percentage
	fmt.Println("\n=== Query 9: Error rate percentage ===")
	errors := logql.Rate(
		logql.NewLogQuery().Eq("job", "api").LineContains("error"),
		10*time.Minute,
	)
	total := logql.Rate(
		logql.NewLogQuery().Eq("job", "api"),
		10*time.Minute,
	)
	expr := logql.Mul(
		logql.Div(errors, total),
		&logql.Literal{Value: 100},
	)
	runExprQuery(expr)

	// Query 10: P95 latency with unwrap
	fmt.Println("\n=== Query 10: P95 latency_ms [10m] ===")
	m5 := logql.QuantileOverTime(0.95,
		logql.NewLogQuery().
			Eq("job", "api").
			JSON().
			Unwrap("latency_ms"),
		10*time.Minute,
	)
	runMetricQuery(m5)

	// Query 11: Builder reuse (immutability demo)
	fmt.Println("\n=== Query 11: Immutability — reuse a base query ===")
	base := logql.NewLogQuery().Eq("job", "api").Eq("env", "prod")
	errQ := base.LineContains("error")
	warnQ := base.LineContains("warn")
	fmt.Printf("  Base:     %s\n", base.String())
	fmt.Printf("  Errors:   %s\n", errQ.String())
	fmt.Printf("  Warnings: %s\n", warnQ.String())

	fmt.Println("\nDone. Open Grafana at http://localhost:3000 to explore further.")
}

// pushSampleLogs sends sample log entries to Loki's push API.
func pushSampleLogs() error {
	now := time.Now()
	entries := make([]string, 0, 20)

	type logLine struct {
		Level     string `json:"level"`
		Status    int    `json:"status"`
		Msg       string `json:"msg"`
		LatencyMs int    `json:"latency_ms"`
	}

	samples := []struct {
		job      string
		env      string
		instance string
		lines    []logLine
	}{
		{
			job: "api", env: "prod", instance: "api-1",
			lines: []logLine{
				{"error", 500, "internal server error", 120},
				{"info", 200, "request completed", 15},
				{"error", 502, "bad gateway", 5000},
				{"warn", 429, "rate limited", 2},
				{"info", 200, "health check ok", 1},
				{"error", 500, "database connection timeout", 3000},
				{"info", 200, "user login successful", 45},
				{"info", 201, "resource created", 30},
				{"warn", 400, "invalid request body", 5},
				{"error", 503, "service unavailable", 10000},
			},
		},
		{
			job: "api", env: "prod", instance: "api-2",
			lines: []logLine{
				{"info", 200, "request completed", 12},
				{"info", 200, "request completed", 8},
				{"error", 500, "null pointer exception", 150},
				{"warn", 429, "rate limited", 3},
				{"info", 200, "cache hit", 2},
			},
		},
		{
			job: "web", env: "prod", instance: "web-1",
			lines: []logLine{
				{"info", 200, "page rendered", 45},
				{"error", 500, "template error", 200},
				{"info", 200, "static asset served", 5},
				{"info", 304, "not modified", 1},
				{"warn", 404, "page not found", 10},
			},
		},
	}

	streams := make([]string, 0, len(samples))
	for _, s := range samples {
		values := make([]string, 0, len(s.lines))
		for i, line := range s.lines {
			ts := now.Add(time.Duration(i) * time.Millisecond)
			lineJSON, _ := json.Marshal(line)
			values = append(values, fmt.Sprintf(`["%d", %s]`, ts.UnixNano(), string(mustJSON(string(lineJSON)))))
		}
		stream := fmt.Sprintf(`{"stream":{"job":%s,"env":%s,"instance":%s},"values":[%s]}`,
			mustJSONStr(s.job), mustJSONStr(s.env), mustJSONStr(s.instance),
			strings.Join(values, ","))
		streams = append(streams, stream)
	}
	_ = entries

	body := fmt.Sprintf(`{"streams":[%s]}`, strings.Join(streams, ","))

	resp, err := http.Post(lokiURL+"/loki/api/v1/push", "application/json", strings.NewReader(body))
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Loki returned %d: %s", resp.StatusCode, string(respBody))
	}
	return nil
}

func mustJSON(s string) []byte {
	b, _ := json.Marshal(s)
	return b
}

func mustJSONStr(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

// runLogQuery builds and executes a log query, printing results.
func runLogQuery(q *logql.LogQuery) {
	queryStr, err := q.Build()
	if err != nil {
		log.Printf("  Build error: %v", err)
		return
	}
	fmt.Printf("  Query: %s\n", queryStr)

	result, err := queryLoki(queryStr)
	if err != nil {
		log.Printf("  Loki error: %v", err)
		return
	}
	printResult(result)
}

// runMetricQuery builds and executes a metric query, printing results.
func runMetricQuery(m *logql.MetricQuery) {
	queryStr, err := m.Build()
	if err != nil {
		log.Printf("  Build error: %v", err)
		return
	}
	fmt.Printf("  Query: %s\n", queryStr)

	result, err := queryLoki(queryStr)
	if err != nil {
		log.Printf("  Loki error: %v", err)
		return
	}
	printResult(result)
}

// runExprQuery builds and executes an expression query, printing results.
func runExprQuery(e *logql.Expr) {
	queryStr, err := e.Build()
	if err != nil {
		log.Printf("  Build error: %v", err)
		return
	}
	fmt.Printf("  Query: %s\n", queryStr)

	result, err := queryLoki(queryStr)
	if err != nil {
		log.Printf("  Loki error: %v", err)
		return
	}
	printResult(result)
}

// queryLoki sends a query to Loki's query_range API and returns the raw JSON response.
func queryLoki(query string) (map[string]interface{}, error) {
	now := time.Now()
	params := url.Values{}
	params.Set("query", query)
	params.Set("start", fmt.Sprintf("%d", now.Add(-1*time.Hour).UnixNano()))
	params.Set("end", fmt.Sprintf("%d", now.Add(1*time.Minute).UnixNano()))
	params.Set("limit", "20")

	reqURL := fmt.Sprintf("%s/loki/api/v1/query_range?%s", lokiURL, params.Encode())
	resp, err := http.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Loki returned %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parsing JSON: %w", err)
	}
	return result, nil
}

// printResult formats and prints a Loki query response.
func printResult(result map[string]interface{}) {
	data, ok := result["data"].(map[string]interface{})
	if !ok {
		fmt.Println("  (no data)")
		return
	}

	resultType, _ := data["resultType"].(string)

	switch resultType {
	case "streams":
		streams, _ := data["result"].([]interface{})
		if len(streams) == 0 {
			fmt.Println("  (no log lines returned)")
			return
		}
		total := 0
		for _, s := range streams {
			stream, _ := s.(map[string]interface{})
			labels, _ := stream["stream"].(map[string]interface{})
			values, _ := stream["values"].([]interface{})
			fmt.Printf("  Stream %v: %d lines\n", labels, len(values))
			for i, v := range values {
				if i >= 3 {
					fmt.Printf("    ... and %d more\n", len(values)-3)
					break
				}
				pair, _ := v.([]interface{})
				if len(pair) == 2 {
					fmt.Printf("    %s\n", pair[1])
				}
			}
			total += len(values)
		}
		fmt.Printf("  Total: %d lines across %d stream(s)\n", total, len(streams))

	case "matrix":
		series, _ := data["result"].([]interface{})
		if len(series) == 0 {
			fmt.Println("  (no metric data returned)")
			return
		}
		for _, s := range series {
			ser, _ := s.(map[string]interface{})
			metric, _ := ser["metric"].(map[string]interface{})
			values, _ := ser["values"].([]interface{})
			if len(values) > 0 {
				lastVal := values[len(values)-1]
				pair, _ := lastVal.([]interface{})
				if len(pair) == 2 {
					fmt.Printf("  %v => %s\n", metric, pair[1])
				}
			}
		}

	case "vector":
		vectors, _ := data["result"].([]interface{})
		if len(vectors) == 0 {
			fmt.Println("  (no instant data returned)")
			return
		}
		for _, v := range vectors {
			vec, _ := v.(map[string]interface{})
			metric, _ := vec["metric"].(map[string]interface{})
			value, _ := vec["value"].([]interface{})
			if len(value) == 2 {
				fmt.Printf("  %v => %s\n", metric, value[1])
			}
		}

	default:
		// Dump raw JSON for unknown types
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("  ", "  ")
		enc.Encode(result)
	}
}
