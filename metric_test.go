package logql

import (
	"testing"
	"time"
)

func TestMetricQuery(t *testing.T) {
	tests := []struct {
		name    string
		build   func() *MetricQuery
		want    string
		wantErr bool
	}{
		{
			name: "rate 5m",
			build: func() *MetricQuery {
				q := NewLogQuery().Eq("job", "api")
				return Rate(q, 5*time.Minute)
			},
			want: `rate({job="api"} [5m])`,
		},
		{
			name: "count_over_time 1h",
			build: func() *MetricQuery {
				q := NewLogQuery().Eq("job", "api")
				return CountOverTime(q, 1*time.Hour)
			},
			want: `count_over_time({job="api"} [1h])`,
		},
		{
			name: "bytes_rate 5m",
			build: func() *MetricQuery {
				q := NewLogQuery().Eq("job", "api")
				return BytesRate(q, 5*time.Minute)
			},
			want: `bytes_rate({job="api"} [5m])`,
		},
		{
			name: "bytes_over_time 1h",
			build: func() *MetricQuery {
				q := NewLogQuery().Eq("job", "api")
				return BytesOverTime(q, 1*time.Hour)
			},
			want: `bytes_over_time({job="api"} [1h])`,
		},
		{
			name: "absent_over_time 5m",
			build: func() *MetricQuery {
				q := NewLogQuery().Eq("job", "api")
				return AbsentOverTime(q, 5*time.Minute)
			},
			want: `absent_over_time({job="api"} [5m])`,
		},
		{
			name: "first_over_time 5m",
			build: func() *MetricQuery {
				q := NewLogQuery().Eq("job", "api")
				return FirstOverTime(q, 5*time.Minute)
			},
			want: `first_over_time({job="api"} [5m])`,
		},
		{
			name: "last_over_time 5m",
			build: func() *MetricQuery {
				q := NewLogQuery().Eq("job", "api")
				return LastOverTime(q, 5*time.Minute)
			},
			want: `last_over_time({job="api"} [5m])`,
		},
		{
			name: "sum_over_time with unwrap",
			build: func() *MetricQuery {
				q := NewLogQuery().Eq("job", "api").JSON().Unwrap("duration")
				return SumOverTime(q, 5*time.Minute)
			},
			want: `sum_over_time({job="api"} | json | unwrap duration [5m])`,
		},
		{
			name: "avg_over_time with unwrap",
			build: func() *MetricQuery {
				q := NewLogQuery().Eq("job", "api").JSON().Unwrap("latency")
				return AvgOverTime(q, 5*time.Minute)
			},
			want: `avg_over_time({job="api"} | json | unwrap latency [5m])`,
		},
		{
			name: "max_over_time with unwrap",
			build: func() *MetricQuery {
				q := NewLogQuery().Eq("job", "api").JSON().Unwrap("bytes")
				return MaxOverTime(q, 5*time.Minute)
			},
			want: `max_over_time({job="api"} | json | unwrap bytes [5m])`,
		},
		{
			name: "min_over_time with unwrap",
			build: func() *MetricQuery {
				q := NewLogQuery().Eq("job", "api").JSON().Unwrap("latency")
				return MinOverTime(q, 5*time.Minute)
			},
			want: `min_over_time({job="api"} | json | unwrap latency [5m])`,
		},
		{
			name: "stddev_over_time with unwrap",
			build: func() *MetricQuery {
				q := NewLogQuery().Eq("job", "api").JSON().Unwrap("latency")
				return StddevOverTime(q, 5*time.Minute)
			},
			want: `stddev_over_time({job="api"} | json | unwrap latency [5m])`,
		},
		{
			name: "stdvar_over_time with unwrap",
			build: func() *MetricQuery {
				q := NewLogQuery().Eq("job", "api").JSON().Unwrap("latency")
				return StdvarOverTime(q, 5*time.Minute)
			},
			want: `stdvar_over_time({job="api"} | json | unwrap latency [5m])`,
		},
		{
			name: "quantile_over_time 0.95",
			build: func() *MetricQuery {
				q := NewLogQuery().Eq("job", "api").JSON().Unwrap("latency_ms")
				return QuantileOverTime(0.95, q, 5*time.Minute)
			},
			want: `quantile_over_time(0.95, {job="api"} | json | unwrap latency_ms [5m])`,
		},
		{
			name: "sum aggregation",
			build: func() *MetricQuery {
				q := NewLogQuery().Eq("job", "api")
				return Rate(q, 5*time.Minute).Sum()
			},
			want: `sum (rate({job="api"} [5m]))`,
		},
		{
			name: "avg aggregation",
			build: func() *MetricQuery {
				q := NewLogQuery().Eq("job", "api")
				return Rate(q, 5*time.Minute).Avg()
			},
			want: `avg (rate({job="api"} [5m]))`,
		},
		{
			name: "count aggregation",
			build: func() *MetricQuery {
				q := NewLogQuery().Eq("job", "api")
				return Rate(q, 5*time.Minute).Count()
			},
			want: `count (rate({job="api"} [5m]))`,
		},
		{
			name: "min aggregation",
			build: func() *MetricQuery {
				q := NewLogQuery().Eq("job", "api")
				return Rate(q, 5*time.Minute).Min()
			},
			want: `min (rate({job="api"} [5m]))`,
		},
		{
			name: "max aggregation",
			build: func() *MetricQuery {
				q := NewLogQuery().Eq("job", "api")
				return Rate(q, 5*time.Minute).Max()
			},
			want: `max (rate({job="api"} [5m]))`,
		},
		{
			name: "stddev aggregation",
			build: func() *MetricQuery {
				q := NewLogQuery().Eq("job", "api")
				return Rate(q, 5*time.Minute).Stddev()
			},
			want: `stddev (rate({job="api"} [5m]))`,
		},
		{
			name: "stdvar aggregation",
			build: func() *MetricQuery {
				q := NewLogQuery().Eq("job", "api")
				return Rate(q, 5*time.Minute).Stdvar()
			},
			want: `stdvar (rate({job="api"} [5m]))`,
		},
		{
			name: "topk 5",
			build: func() *MetricQuery {
				q := NewLogQuery().Eq("job", "api")
				return Rate(q, 5*time.Minute).TopK(5)
			},
			want: `topk(5, rate({job="api"} [5m]))`,
		},
		{
			name: "bottomk 3",
			build: func() *MetricQuery {
				q := NewLogQuery().Eq("job", "api")
				return Rate(q, 5*time.Minute).BottomK(3)
			},
			want: `bottomk(3, rate({job="api"} [5m]))`,
		},
		{
			name: "sort",
			build: func() *MetricQuery {
				q := NewLogQuery().Eq("job", "api")
				return Rate(q, 5*time.Minute).Sort()
			},
			want: `sort(rate({job="api"} [5m]))`,
		},
		{
			name: "sort_desc",
			build: func() *MetricQuery {
				q := NewLogQuery().Eq("job", "api")
				return Rate(q, 5*time.Minute).SortDesc()
			},
			want: `sort_desc(rate({job="api"} [5m]))`,
		},
		{
			name: "sum by labels",
			build: func() *MetricQuery {
				q := NewLogQuery().Eq("job", "api").LineContains("error")
				return Rate(q, 5*time.Minute).Sum().By("job", "instance")
			},
			want: `sum by (job, instance) (rate({job="api"} |= "error" [5m]))`,
		},
		{
			name: "sum without labels",
			build: func() *MetricQuery {
				q := NewLogQuery().Eq("job", "api")
				return Rate(q, 5*time.Minute).Sum().Without("instance")
			},
			want: `sum without (instance) (rate({job="api"} [5m]))`,
		},
		{
			name: "offset 1h",
			build: func() *MetricQuery {
				q := NewLogQuery().Eq("job", "api")
				return Rate(q, 5*time.Minute).Offset(1 * time.Hour)
			},
			want: `rate({job="api"} [5m]) offset 1h`,
		},
		{
			name: "nested aggregation + grouping + offset",
			build: func() *MetricQuery {
				q := NewLogQuery().Eq("job", "api").LineContains("error")
				return Rate(q, 5*time.Minute).Offset(1*time.Hour).Sum().By("job")
			},
			want: `sum by (job) (rate({job="api"} |= "error" [5m]) offset 1h)`,
		},
		{
			name: "rate with line filter and parser",
			build: func() *MetricQuery {
				q := NewLogQuery().Eq("job", "api").LineContains("error").JSON()
				return Rate(q, 5*time.Minute)
			},
			want: `rate({job="api"} |= "error" | json [5m])`,
		},
		{
			name: "duration seconds",
			build: func() *MetricQuery {
				q := NewLogQuery().Eq("job", "api")
				return Rate(q, 30*time.Second)
			},
			want: `rate({job="api"} [30s])`,
		},
		{
			name: "duration milliseconds",
			build: func() *MetricQuery {
				q := NewLogQuery().Eq("job", "api")
				return Rate(q, 500*time.Millisecond)
			},
			want: `rate({job="api"} [500ms])`,
		},
		{
			name: "invalid duration zero",
			build: func() *MetricQuery {
				q := NewLogQuery().Eq("job", "api")
				return Rate(q, 0)
			},
			wantErr: true,
		},
		{
			name: "invalid duration negative",
			build: func() *MetricQuery {
				q := NewLogQuery().Eq("job", "api")
				return Rate(q, -5*time.Minute)
			},
			wantErr: true,
		},
		{
			name: "invalid quantile > 1",
			build: func() *MetricQuery {
				q := NewLogQuery().Eq("job", "api").JSON().Unwrap("latency")
				return QuantileOverTime(1.5, q, 5*time.Minute)
			},
			wantErr: true,
		},
		{
			name: "invalid quantile < 0",
			build: func() *MetricQuery {
				q := NewLogQuery().Eq("job", "api").JSON().Unwrap("latency")
				return QuantileOverTime(-0.1, q, 5*time.Minute)
			},
			wantErr: true,
		},
		{
			name: "topk invalid k",
			build: func() *MetricQuery {
				q := NewLogQuery().Eq("job", "api")
				return Rate(q, 5*time.Minute).TopK(0)
			},
			wantErr: true,
		},
		{
			name: "bottomk invalid k",
			build: func() *MetricQuery {
				q := NewLogQuery().Eq("job", "api")
				return Rate(q, 5*time.Minute).BottomK(-1)
			},
			wantErr: true,
		},
		{
			name: "inner query error propagates",
			build: func() *MetricQuery {
				q := NewLogQuery() // no selectors
				return Rate(q, 5*time.Minute)
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

func TestMetricQueryImmutability(t *testing.T) {
	q := NewLogQuery().Eq("job", "api")
	base := Rate(q, 5*time.Minute)

	summed := base.Sum()
	avg := base.Avg()

	baseStr, _ := base.Build()
	sumStr, _ := summed.Build()
	avgStr, _ := avg.Build()

	if baseStr != `rate({job="api"} [5m])` {
		t.Errorf("base mutated: %q", baseStr)
	}
	if sumStr != `sum (rate({job="api"} [5m]))` {
		t.Errorf("sum wrong: %q", sumStr)
	}
	if avgStr != `avg (rate({job="api"} [5m]))` {
		t.Errorf("avg wrong: %q", avgStr)
	}
}

func TestMetricQueryString(t *testing.T) {
	q := NewLogQuery().Eq("job", "api")
	m := Rate(q, 5*time.Minute)
	s := m.String()
	if s != `rate({job="api"} [5m])` {
		t.Errorf("String() = %q, want %q", s, `rate({job="api"} [5m])`)
	}
}

func TestMetricQueryStringPanicsOnError(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("String() should panic on error, but didn't")
		}
	}()
	q := NewLogQuery().Eq("job", "api")
	_ = Rate(q, 0).String()
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{5 * time.Minute, "5m"},
		{1 * time.Hour, "1h"},
		{30 * time.Second, "30s"},
		{500 * time.Millisecond, "500ms"},
		{2 * time.Hour, "2h"},
		{90 * time.Second, "90s"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := formatDuration(tt.d)
			if got != tt.want {
				t.Errorf("formatDuration(%v) = %q, want %q", tt.d, got, tt.want)
			}
		})
	}
}
