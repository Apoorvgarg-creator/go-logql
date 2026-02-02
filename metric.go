package logql

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// MetricQuery wraps a LogQuery in a metric aggregation.
type MetricQuery struct {
	inner        *LogQuery
	rangeAgg     string
	duration     time.Duration
	quantile     float64
	offset       time.Duration
	aggOp        string
	groupBy      []string
	groupWithout []string
	topBottomK   int
	err          error
}

func (m *MetricQuery) clone() *MetricQuery {
	newM := &MetricQuery{
		inner:    m.inner,
		rangeAgg: m.rangeAgg,
		duration: m.duration,
		quantile: m.quantile,
		offset:   m.offset,
		aggOp:    m.aggOp,
		topBottomK: m.topBottomK,
		err:      m.err,
	}
	if m.groupBy != nil {
		newM.groupBy = make([]string, len(m.groupBy))
		copy(newM.groupBy, m.groupBy)
	}
	if m.groupWithout != nil {
		newM.groupWithout = make([]string, len(m.groupWithout))
		copy(newM.groupWithout, m.groupWithout)
	}
	return newM
}

// formatDuration formats a time.Duration as a LogQL duration string.
func formatDuration(d time.Duration) string {
	if d >= time.Hour && d%time.Hour == 0 {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	if d >= time.Minute && d%time.Minute == 0 {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d >= time.Second && d%time.Second == 0 {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	return fmt.Sprintf("%dms", d.Milliseconds())
}

func newRangeAgg(name string, q *LogQuery, duration time.Duration) *MetricQuery {
	m := &MetricQuery{
		inner:    q,
		rangeAgg: name,
		duration: duration,
	}
	if duration <= 0 {
		m.err = errors.New("logql: duration must be positive")
	}
	return m
}

// Rate creates a rate range aggregation: rate({...} [duration]).
func Rate(q *LogQuery, duration time.Duration) *MetricQuery {
	return newRangeAgg("rate", q, duration)
}

// CountOverTime creates a count_over_time range aggregation.
func CountOverTime(q *LogQuery, duration time.Duration) *MetricQuery {
	return newRangeAgg("count_over_time", q, duration)
}

// BytesRate creates a bytes_rate range aggregation.
func BytesRate(q *LogQuery, duration time.Duration) *MetricQuery {
	return newRangeAgg("bytes_rate", q, duration)
}

// BytesOverTime creates a bytes_over_time range aggregation.
func BytesOverTime(q *LogQuery, duration time.Duration) *MetricQuery {
	return newRangeAgg("bytes_over_time", q, duration)
}

// AbsentOverTime creates an absent_over_time range aggregation.
func AbsentOverTime(q *LogQuery, duration time.Duration) *MetricQuery {
	return newRangeAgg("absent_over_time", q, duration)
}

// FirstOverTime creates a first_over_time range aggregation.
func FirstOverTime(q *LogQuery, duration time.Duration) *MetricQuery {
	return newRangeAgg("first_over_time", q, duration)
}

// LastOverTime creates a last_over_time range aggregation.
func LastOverTime(q *LogQuery, duration time.Duration) *MetricQuery {
	return newRangeAgg("last_over_time", q, duration)
}

// SumOverTime creates a sum_over_time unwrap range aggregation.
func SumOverTime(q *LogQuery, duration time.Duration) *MetricQuery {
	return newRangeAgg("sum_over_time", q, duration)
}

// AvgOverTime creates an avg_over_time unwrap range aggregation.
func AvgOverTime(q *LogQuery, duration time.Duration) *MetricQuery {
	return newRangeAgg("avg_over_time", q, duration)
}

// MaxOverTime creates a max_over_time unwrap range aggregation.
func MaxOverTime(q *LogQuery, duration time.Duration) *MetricQuery {
	return newRangeAgg("max_over_time", q, duration)
}

// MinOverTime creates a min_over_time unwrap range aggregation.
func MinOverTime(q *LogQuery, duration time.Duration) *MetricQuery {
	return newRangeAgg("min_over_time", q, duration)
}

// StddevOverTime creates a stddev_over_time unwrap range aggregation.
func StddevOverTime(q *LogQuery, duration time.Duration) *MetricQuery {
	return newRangeAgg("stddev_over_time", q, duration)
}

// StdvarOverTime creates a stdvar_over_time unwrap range aggregation.
func StdvarOverTime(q *LogQuery, duration time.Duration) *MetricQuery {
	return newRangeAgg("stdvar_over_time", q, duration)
}

// QuantileOverTime creates a quantile_over_time unwrap range aggregation.
func QuantileOverTime(quantile float64, q *LogQuery, duration time.Duration) *MetricQuery {
	m := newRangeAgg("quantile_over_time", q, duration)
	m.quantile = quantile
	if quantile < 0 || quantile > 1 {
		m.err = errors.New("logql: quantile must be between 0 and 1")
	}
	return m
}

// Sum wraps the metric query in a sum aggregation.
func (m *MetricQuery) Sum() *MetricQuery {
	newM := m.clone()
	newM.aggOp = "sum"
	return newM
}

// Avg wraps the metric query in an avg aggregation.
func (m *MetricQuery) Avg() *MetricQuery {
	newM := m.clone()
	newM.aggOp = "avg"
	return newM
}

// Min wraps the metric query in a min aggregation.
func (m *MetricQuery) Min() *MetricQuery {
	newM := m.clone()
	newM.aggOp = "min"
	return newM
}

// Max wraps the metric query in a max aggregation.
func (m *MetricQuery) Max() *MetricQuery {
	newM := m.clone()
	newM.aggOp = "max"
	return newM
}

// Stddev wraps the metric query in a stddev aggregation.
func (m *MetricQuery) Stddev() *MetricQuery {
	newM := m.clone()
	newM.aggOp = "stddev"
	return newM
}

// Stdvar wraps the metric query in a stdvar aggregation.
func (m *MetricQuery) Stdvar() *MetricQuery {
	newM := m.clone()
	newM.aggOp = "stdvar"
	return newM
}

// Count wraps the metric query in a count aggregation.
func (m *MetricQuery) Count() *MetricQuery {
	newM := m.clone()
	newM.aggOp = "count"
	return newM
}

// TopK wraps the metric query in a topk aggregation.
func (m *MetricQuery) TopK(k int) *MetricQuery {
	newM := m.clone()
	newM.aggOp = "topk"
	newM.topBottomK = k
	if k <= 0 {
		newM.err = errors.New("logql: topk k must be > 0")
	}
	return newM
}

// BottomK wraps the metric query in a bottomk aggregation.
func (m *MetricQuery) BottomK(k int) *MetricQuery {
	newM := m.clone()
	newM.aggOp = "bottomk"
	newM.topBottomK = k
	if k <= 0 {
		newM.err = errors.New("logql: bottomk k must be > 0")
	}
	return newM
}

// Sort wraps the metric query in a sort aggregation.
func (m *MetricQuery) Sort() *MetricQuery {
	newM := m.clone()
	newM.aggOp = "sort"
	return newM
}

// SortDesc wraps the metric query in a sort_desc aggregation.
func (m *MetricQuery) SortDesc() *MetricQuery {
	newM := m.clone()
	newM.aggOp = "sort_desc"
	return newM
}

// By sets the "by" grouping labels for the aggregation.
func (m *MetricQuery) By(labels ...string) *MetricQuery {
	newM := m.clone()
	newM.groupBy = make([]string, len(labels))
	copy(newM.groupBy, labels)
	newM.groupWithout = nil
	return newM
}

// Without sets the "without" grouping labels for the aggregation.
func (m *MetricQuery) Without(labels ...string) *MetricQuery {
	newM := m.clone()
	newM.groupWithout = make([]string, len(labels))
	copy(newM.groupWithout, labels)
	newM.groupBy = nil
	return newM
}

// Offset sets the offset modifier on the range aggregation.
func (m *MetricQuery) Offset(d time.Duration) *MetricQuery {
	newM := m.clone()
	newM.offset = d
	return newM
}

// Build renders the full metric query string.
func (m *MetricQuery) Build() (string, error) {
	if m.err != nil {
		return "", m.err
	}

	innerStr, err := m.inner.Build()
	if err != nil {
		return "", err
	}

	var sb strings.Builder

	// Build the range aggregation: range_agg(inner [duration])
	rangeExpr := m.buildRangeExpr(innerStr)

	// Apply offset if set
	if m.offset > 0 {
		rangeExpr = rangeExpr + " offset " + formatDuration(m.offset)
	}

	// Wrap in aggregation operator if set
	if m.aggOp != "" {
		m.buildAggExpr(&sb, rangeExpr)
	} else {
		sb.WriteString(rangeExpr)
	}

	return sb.String(), nil
}

func (m *MetricQuery) buildRangeExpr(innerStr string) string {
	var sb strings.Builder
	if m.rangeAgg == "quantile_over_time" {
		sb.WriteString(fmt.Sprintf("quantile_over_time(%s, ", formatQuantile(m.quantile)))
	} else {
		sb.WriteString(m.rangeAgg)
		sb.WriteByte('(')
	}
	sb.WriteString(innerStr)
	sb.WriteString(" [")
	sb.WriteString(formatDuration(m.duration))
	sb.WriteString("])")
	return sb.String()
}

func (m *MetricQuery) buildAggExpr(sb *strings.Builder, rangeExpr string) {
	switch m.aggOp {
	case "topk", "bottomk":
		sb.WriteString(fmt.Sprintf("%s(%d, %s)", m.aggOp, m.topBottomK, rangeExpr))
	case "sort", "sort_desc":
		sb.WriteString(fmt.Sprintf("%s(%s)", m.aggOp, rangeExpr))
	default:
		sb.WriteString(m.aggOp)
		if len(m.groupBy) > 0 {
			sb.WriteString(" by (")
			sb.WriteString(strings.Join(m.groupBy, ", "))
			sb.WriteString(")")
		} else if len(m.groupWithout) > 0 {
			sb.WriteString(" without (")
			sb.WriteString(strings.Join(m.groupWithout, ", "))
			sb.WriteString(")")
		}
		sb.WriteString(" (")
		sb.WriteString(rangeExpr)
		sb.WriteByte(')')
	}
}

func formatQuantile(q float64) string {
	s := fmt.Sprintf("%g", q)
	return s
}

// String renders the metric query string, panicking on error.
func (m *MetricQuery) String() string {
	s, err := m.Build()
	if err != nil {
		panic(err)
	}
	return s
}
