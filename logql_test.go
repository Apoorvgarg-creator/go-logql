package logql

import (
	"sync"
	"testing"
)

func TestLogQuery(t *testing.T) {
	tests := []struct {
		name    string
		query   func() *LogQuery
		want    string
		wantErr bool
	}{
		{
			name:  "single label selector",
			query: func() *LogQuery { return NewLogQuery().Eq("job", "api") },
			want:  `{job="api"}`,
		},
		{
			name: "multiple labels",
			query: func() *LogQuery {
				return NewLogQuery().Eq("job", "api").Eq("env", "prod")
			},
			want: `{job="api", env="prod"}`,
		},
		{
			name:  "not equal match",
			query: func() *LogQuery { return NewLogQuery().Neq("env", "dev") },
			want:  `{env!="dev"}`,
		},
		{
			name:  "regex match",
			query: func() *LogQuery { return NewLogQuery().Re("job", "api|web") },
			want:  `{job=~"api|web"}`,
		},
		{
			name:  "not regex match",
			query: func() *LogQuery { return NewLogQuery().Nre("job", "debug.*") },
			want:  `{job!~"debug.*"}`,
		},
		{
			name: "all four match ops",
			query: func() *LogQuery {
				return NewLogQuery().
					Eq("job", "api").
					Neq("env", "dev").
					Re("instance", "10\\.0\\..*").
					Nre("method", "OPTIONS|HEAD")
			},
			want: `{job="api", env!="dev", instance=~"10\.0\..*", method!~"OPTIONS|HEAD"}`,
		},
		{
			name: "line contains filter",
			query: func() *LogQuery {
				return NewLogQuery().Eq("job", "api").LineContains("error")
			},
			want: `{job="api"} |= "error"`,
		},
		{
			name: "line not contains filter",
			query: func() *LogQuery {
				return NewLogQuery().Eq("job", "api").LineNotContains("debug")
			},
			want: `{job="api"} != "debug"`,
		},
		{
			name: "line match filter",
			query: func() *LogQuery {
				return NewLogQuery().Eq("job", "api").LineMatch("error|warn")
			},
			want: `{job="api"} |~ "error|warn"`,
		},
		{
			name: "line not match filter",
			query: func() *LogQuery {
				return NewLogQuery().Eq("job", "api").LineNotMatch("debug|trace")
			},
			want: `{job="api"} !~ "debug|trace"`,
		},
		{
			name: "json parser no args",
			query: func() *LogQuery {
				return NewLogQuery().Eq("job", "api").JSON()
			},
			want: `{job="api"} | json`,
		},
		{
			name: "json parser with args",
			query: func() *LogQuery {
				return NewLogQuery().Eq("job", "api").JSON("field1", "field2")
			},
			want: `{job="api"} | json field1, field2`,
		},
		{
			name: "logfmt parser",
			query: func() *LogQuery {
				return NewLogQuery().Eq("job", "api").Logfmt()
			},
			want: `{job="api"} | logfmt`,
		},
		{
			name: "logfmt parser with labels",
			query: func() *LogQuery {
				return NewLogQuery().Eq("job", "api").Logfmt("level", "msg")
			},
			want: `{job="api"} | logfmt level, msg`,
		},
		{
			name: "regexp parser",
			query: func() *LogQuery {
				return NewLogQuery().Eq("job", "api").Regexp(`(?P<method>\w+) (?P<path>\S+)`)
			},
			want: `{job="api"} | regexp "(?P<method>\w+) (?P<path>\S+)"`,
		},
		{
			name: "pattern parser",
			query: func() *LogQuery {
				return NewLogQuery().Eq("job", "api").Pattern("<method> <path> <status>")
			},
			want: `{job="api"} | pattern "<method> <path> <status>"`,
		},
		{
			name: "unpack parser",
			query: func() *LogQuery {
				return NewLogQuery().Eq("job", "api").Unpack()
			},
			want: `{job="api"} | unpack`,
		},
		{
			name: "unpack parser with labels",
			query: func() *LogQuery {
				return NewLogQuery().Eq("job", "api").Unpack("field1", "field2")
			},
			want: `{job="api"} | unpack field1, field2`,
		},
		{
			name: "label equal filter",
			query: func() *LogQuery {
				return NewLogQuery().Eq("job", "api").JSON().LabelEqual("level", "error")
			},
			want: `{job="api"} | json | level == "error"`,
		},
		{
			name: "label not equal filter",
			query: func() *LogQuery {
				return NewLogQuery().Eq("job", "api").JSON().LabelNotEqual("level", "debug")
			},
			want: `{job="api"} | json | level != "debug"`,
		},
		{
			name: "label greater filter",
			query: func() *LogQuery {
				return NewLogQuery().Eq("job", "api").JSON().LabelGreater("status", "400")
			},
			want: `{job="api"} | json | status > 400`,
		},
		{
			name: "label greater eq filter",
			query: func() *LogQuery {
				return NewLogQuery().Eq("job", "api").JSON().LabelGreaterEq("status", "400")
			},
			want: `{job="api"} | json | status >= 400`,
		},
		{
			name: "label less filter",
			query: func() *LogQuery {
				return NewLogQuery().Eq("job", "api").JSON().LabelLess("duration", "5s")
			},
			want: `{job="api"} | json | duration < 5s`,
		},
		{
			name: "label less eq filter",
			query: func() *LogQuery {
				return NewLogQuery().Eq("job", "api").JSON().LabelLessEq("duration", "10s")
			},
			want: `{job="api"} | json | duration <= 10s`,
		},
		{
			name: "label regex filter",
			query: func() *LogQuery {
				return NewLogQuery().Eq("job", "api").JSON().LabelRe("method", "GET|POST")
			},
			want: `{job="api"} | json | method =~ "GET|POST"`,
		},
		{
			name: "label not regex filter",
			query: func() *LogQuery {
				return NewLogQuery().Eq("job", "api").JSON().LabelNre("path", "/health.*")
			},
			want: `{job="api"} | json | path !~ "/health.*"`,
		},
		{
			name: "line format",
			query: func() *LogQuery {
				return NewLogQuery().Eq("job", "api").LineFormat("{{.msg}}")
			},
			want: `{job="api"} | line_format "{{.msg}}"`,
		},
		{
			name: "label format single",
			query: func() *LogQuery {
				return NewLogQuery().Eq("job", "api").LabelFormatEntry("dst", "src")
			},
			want: `{job="api"} | label_format dst=src`,
		},
		{
			name: "drop labels",
			query: func() *LogQuery {
				return NewLogQuery().Eq("job", "api").Drop("label1", "label2")
			},
			want: `{job="api"} | drop label1, label2`,
		},
		{
			name: "keep labels",
			query: func() *LogQuery {
				return NewLogQuery().Eq("job", "api").Keep("label1", "label2")
			},
			want: `{job="api"} | keep label1, label2`,
		},
		{
			name: "decolorize",
			query: func() *LogQuery {
				return NewLogQuery().Eq("job", "api").Decolorize()
			},
			want: `{job="api"} | decolorize`,
		},
		{
			name: "unwrap stage",
			query: func() *LogQuery {
				return NewLogQuery().Eq("job", "api").JSON().Unwrap("latency_ms")
			},
			want: `{job="api"} | json | unwrap latency_ms`,
		},
		{
			name: "complex pipeline",
			query: func() *LogQuery {
				return NewLogQuery().
					Eq("job", "api").
					Eq("env", "prod").
					LineContains("error").
					JSON().
					LabelEqual("level", "error").
					LabelGreaterEq("status", "400").
					LineFormat("{{.msg}}")
			},
			want: `{job="api", env="prod"} |= "error" | json | level == "error" | status >= 400 | line_format "{{.msg}}"`,
		},
		{
			name:    "empty selector error",
			query:   func() *LogQuery { return NewLogQuery() },
			wantErr: true,
		},
		{
			name:    "empty label name error",
			query:   func() *LogQuery { return NewLogQuery().Eq("", "value") },
			wantErr: true,
		},
		{
			name: "invalid regex in Re",
			query: func() *LogQuery {
				return NewLogQuery().Re("job", "[invalid")
			},
			wantErr: true,
		},
		{
			name: "invalid regex in Nre",
			query: func() *LogQuery {
				return NewLogQuery().Nre("job", "[invalid")
			},
			wantErr: true,
		},
		{
			name: "invalid regex in LineMatch",
			query: func() *LogQuery {
				return NewLogQuery().Eq("job", "api").LineMatch("[invalid")
			},
			wantErr: true,
		},
		{
			name: "invalid regex in LineNotMatch",
			query: func() *LogQuery {
				return NewLogQuery().Eq("job", "api").LineNotMatch("[invalid")
			},
			wantErr: true,
		},
		{
			name: "invalid regex in LabelRe",
			query: func() *LogQuery {
				return NewLogQuery().Eq("job", "api").LabelRe("method", "[invalid")
			},
			wantErr: true,
		},
		{
			name: "invalid regex in LabelNre",
			query: func() *LogQuery {
				return NewLogQuery().Eq("job", "api").LabelNre("path", "[invalid")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.query().Build()
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

func TestLogQueryImmutability(t *testing.T) {
	base := NewLogQuery().Eq("job", "api").Eq("env", "prod")

	errors := base.LineContains("error")
	warnings := base.LineContains("warning")

	baseStr, err := base.Build()
	if err != nil {
		t.Fatalf("base.Build() error: %v", err)
	}
	errStr, err := errors.Build()
	if err != nil {
		t.Fatalf("errors.Build() error: %v", err)
	}
	warnStr, err := warnings.Build()
	if err != nil {
		t.Fatalf("warnings.Build() error: %v", err)
	}

	if baseStr != `{job="api", env="prod"}` {
		t.Errorf("base mutated: %q", baseStr)
	}
	if errStr != `{job="api", env="prod"} |= "error"` {
		t.Errorf("errors wrong: %q", errStr)
	}
	if warnStr != `{job="api", env="prod"} |= "warning"` {
		t.Errorf("warnings wrong: %q", warnStr)
	}
}

func TestLogQueryConcurrency(t *testing.T) {
	base := NewLogQuery().Eq("job", "api")

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			q := base.LineContains("error").JSON().LabelEqual("level", "error")
			_, err := q.Build()
			if err != nil {
				t.Errorf("Build() error in goroutine: %v", err)
			}
		}()
	}
	wg.Wait()
}

func TestLogQueryString(t *testing.T) {
	q := NewLogQuery().Eq("job", "api")
	s := q.String()
	if s != `{job="api"}` {
		t.Errorf("String() = %q, want %q", s, `{job="api"}`)
	}
}

func TestLogQueryStringPanicsOnError(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("String() should panic on error, but didn't")
		}
	}()
	_ = NewLogQuery().String()
}
