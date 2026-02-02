package logql

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// stage is an internal interface for pipeline stages.
type stage interface {
	String() string
}

// lineFilter represents a line filter stage: |= "error"
type lineFilter struct {
	op    string
	value string
}

func (s lineFilter) String() string {
	return fmt.Sprintf(`%s "%s"`, s.op, s.value)
}

// parserStage represents a parser stage: | json, | logfmt, | regexp "pat", | pattern "pat", | unpack
type parserStage struct {
	parser string
	params []string
	quoted bool // whether params should be quoted (regexp, pattern)
}

func (s parserStage) String() string {
	if len(s.params) == 0 {
		return "| " + s.parser
	}
	if s.quoted {
		return fmt.Sprintf(`| %s "%s"`, s.parser, s.params[0])
	}
	return fmt.Sprintf("| %s %s", s.parser, strings.Join(s.params, ", "))
}

// labelFilterStage represents a label filter stage: | status >= 400
type labelFilterStage struct {
	filter LabelFilter
}

func (s labelFilterStage) String() string {
	// For string comparison operators that use regex or equality, quote the value
	switch s.filter.Op {
	case FilterEqual, FilterNotEqual:
		return fmt.Sprintf(`| %s %s "%s"`, s.filter.Label, string(s.filter.Op), s.filter.Value)
	case FilterRegexp, FilterNotRegexp:
		return fmt.Sprintf(`| %s %s "%s"`, s.filter.Label, string(s.filter.Op), s.filter.Value)
	default:
		// Numeric comparisons: render value as-is
		return fmt.Sprintf("| %s %s %s", s.filter.Label, string(s.filter.Op), s.filter.Value)
	}
}

// lineFormatStage represents: | line_format "template"
type lineFormatStage struct {
	template string
}

func (s lineFormatStage) String() string {
	return fmt.Sprintf(`| line_format "%s"`, s.template)
}

// labelFormatEntry represents a single dst=src mapping.
type labelFormatEntry struct {
	dst string
	src string
}

// labelFormatStage represents: | label_format dst=src, dst2=src2
type labelFormatStage struct {
	entries []labelFormatEntry
}

func (s labelFormatStage) String() string {
	parts := make([]string, len(s.entries))
	for i, e := range s.entries {
		parts[i] = e.dst + "=" + e.src
	}
	return "| label_format " + strings.Join(parts, ", ")
}

// dropStage represents: | drop label1, label2
type dropStage struct {
	labels []string
}

func (s dropStage) String() string {
	return "| drop " + strings.Join(s.labels, ", ")
}

// keepStage represents: | keep label1, label2
type keepStage struct {
	labels []string
}

func (s keepStage) String() string {
	return "| keep " + strings.Join(s.labels, ", ")
}

// decolorizeStage represents: | decolorize
type decolorizeStage struct{}

func (s decolorizeStage) String() string {
	return "| decolorize"
}

// unwrapStage represents: | unwrap label
type unwrapStage struct {
	label string
}

func (s unwrapStage) String() string {
	return "| unwrap " + s.label
}

// LogQuery builds a LogQL log query: {selectors} | pipeline stages...
type LogQuery struct {
	matchers []LabelMatcher
	stages   []stage
	err      error
}

// NewLogQuery creates a new empty LogQuery builder.
func NewLogQuery() *LogQuery {
	return &LogQuery{}
}

func (q *LogQuery) clone() *LogQuery {
	newQ := &LogQuery{
		err: q.err,
	}
	newQ.matchers = make([]LabelMatcher, len(q.matchers))
	copy(newQ.matchers, q.matchers)
	newQ.stages = make([]stage, len(q.stages))
	copy(newQ.stages, q.stages)
	return newQ
}

// Eq adds an exact match stream selector: {label="value"}.
func (q *LogQuery) Eq(label, value string) *LogQuery {
	newQ := q.clone()
	if label == "" {
		newQ.err = errors.New("logql: label name cannot be empty")
		return newQ
	}
	newQ.matchers = append(newQ.matchers, LabelMatcher{Label: label, Op: MatchEqual, Value: value})
	return newQ
}

// Neq adds a not-equal stream selector: {label!="value"}.
func (q *LogQuery) Neq(label, value string) *LogQuery {
	newQ := q.clone()
	if label == "" {
		newQ.err = errors.New("logql: label name cannot be empty")
		return newQ
	}
	newQ.matchers = append(newQ.matchers, LabelMatcher{Label: label, Op: MatchNotEqual, Value: value})
	return newQ
}

// Re adds a regex match stream selector: {label=~"pattern"}.
func (q *LogQuery) Re(label, pattern string) *LogQuery {
	newQ := q.clone()
	if label == "" {
		newQ.err = errors.New("logql: label name cannot be empty")
		return newQ
	}
	if _, err := regexp.Compile(pattern); err != nil {
		newQ.err = fmt.Errorf("logql: invalid regex pattern %q: %w", pattern, err)
		return newQ
	}
	newQ.matchers = append(newQ.matchers, LabelMatcher{Label: label, Op: MatchRegexp, Value: pattern})
	return newQ
}

// Nre adds a not-regex stream selector: {label!~"pattern"}.
func (q *LogQuery) Nre(label, pattern string) *LogQuery {
	newQ := q.clone()
	if label == "" {
		newQ.err = errors.New("logql: label name cannot be empty")
		return newQ
	}
	if _, err := regexp.Compile(pattern); err != nil {
		newQ.err = fmt.Errorf("logql: invalid regex pattern %q: %w", pattern, err)
		return newQ
	}
	newQ.matchers = append(newQ.matchers, LabelMatcher{Label: label, Op: MatchNotRegexp, Value: pattern})
	return newQ
}

// LineContains adds a line contains filter: |= "text".
func (q *LogQuery) LineContains(text string) *LogQuery {
	newQ := q.clone()
	newQ.stages = append(newQ.stages, lineFilter{op: "|=", value: text})
	return newQ
}

// LineNotContains adds a line not-contains filter: != "text".
func (q *LogQuery) LineNotContains(text string) *LogQuery {
	newQ := q.clone()
	newQ.stages = append(newQ.stages, lineFilter{op: "!=", value: text})
	return newQ
}

// LineMatch adds a line regex match filter: |~ "pattern".
func (q *LogQuery) LineMatch(pattern string) *LogQuery {
	newQ := q.clone()
	if _, err := regexp.Compile(pattern); err != nil {
		newQ.err = fmt.Errorf("logql: invalid regex pattern %q: %w", pattern, err)
		return newQ
	}
	newQ.stages = append(newQ.stages, lineFilter{op: "|~", value: pattern})
	return newQ
}

// LineNotMatch adds a line not-regex match filter: !~ "pattern".
func (q *LogQuery) LineNotMatch(pattern string) *LogQuery {
	newQ := q.clone()
	if _, err := regexp.Compile(pattern); err != nil {
		newQ.err = fmt.Errorf("logql: invalid regex pattern %q: %w", pattern, err)
		return newQ
	}
	newQ.stages = append(newQ.stages, lineFilter{op: "!~", value: pattern})
	return newQ
}

// JSON adds a JSON parser stage: | json or | json label1, label2.
func (q *LogQuery) JSON(labels ...string) *LogQuery {
	newQ := q.clone()
	newQ.stages = append(newQ.stages, parserStage{parser: "json", params: labels})
	return newQ
}

// Logfmt adds a logfmt parser stage: | logfmt or | logfmt label1, label2.
func (q *LogQuery) Logfmt(labels ...string) *LogQuery {
	newQ := q.clone()
	newQ.stages = append(newQ.stages, parserStage{parser: "logfmt", params: labels})
	return newQ
}

// Regexp adds a regexp parser stage: | regexp "pattern".
func (q *LogQuery) Regexp(pattern string) *LogQuery {
	newQ := q.clone()
	newQ.stages = append(newQ.stages, parserStage{parser: "regexp", params: []string{pattern}, quoted: true})
	return newQ
}

// Pattern adds a pattern parser stage: | pattern "pattern".
func (q *LogQuery) Pattern(pattern string) *LogQuery {
	newQ := q.clone()
	newQ.stages = append(newQ.stages, parserStage{parser: "pattern", params: []string{pattern}, quoted: true})
	return newQ
}

// Unpack adds an unpack parser stage: | unpack or | unpack label1, label2.
func (q *LogQuery) Unpack(labels ...string) *LogQuery {
	newQ := q.clone()
	newQ.stages = append(newQ.stages, parserStage{parser: "unpack", params: labels})
	return newQ
}

// LabelEqual adds a label equality filter: | label == "value".
func (q *LogQuery) LabelEqual(label, value string) *LogQuery {
	newQ := q.clone()
	newQ.stages = append(newQ.stages, labelFilterStage{filter: LabelFilter{Label: label, Op: FilterEqual, Value: value}})
	return newQ
}

// LabelNotEqual adds a label not-equal filter: | label != "value".
func (q *LogQuery) LabelNotEqual(label, value string) *LogQuery {
	newQ := q.clone()
	newQ.stages = append(newQ.stages, labelFilterStage{filter: LabelFilter{Label: label, Op: FilterNotEqual, Value: value}})
	return newQ
}

// LabelGreater adds a label greater-than filter: | label > value.
func (q *LogQuery) LabelGreater(label, value string) *LogQuery {
	newQ := q.clone()
	newQ.stages = append(newQ.stages, labelFilterStage{filter: LabelFilter{Label: label, Op: FilterGreater, Value: value}})
	return newQ
}

// LabelGreaterEq adds a label greater-or-equal filter: | label >= value.
func (q *LogQuery) LabelGreaterEq(label, value string) *LogQuery {
	newQ := q.clone()
	newQ.stages = append(newQ.stages, labelFilterStage{filter: LabelFilter{Label: label, Op: FilterGreaterEqual, Value: value}})
	return newQ
}

// LabelLess adds a label less-than filter: | label < value.
func (q *LogQuery) LabelLess(label, value string) *LogQuery {
	newQ := q.clone()
	newQ.stages = append(newQ.stages, labelFilterStage{filter: LabelFilter{Label: label, Op: FilterLess, Value: value}})
	return newQ
}

// LabelLessEq adds a label less-or-equal filter: | label <= value.
func (q *LogQuery) LabelLessEq(label, value string) *LogQuery {
	newQ := q.clone()
	newQ.stages = append(newQ.stages, labelFilterStage{filter: LabelFilter{Label: label, Op: FilterLessEqual, Value: value}})
	return newQ
}

// LabelRe adds a label regex filter: | label =~ "pattern".
func (q *LogQuery) LabelRe(label, pattern string) *LogQuery {
	newQ := q.clone()
	if _, err := regexp.Compile(pattern); err != nil {
		newQ.err = fmt.Errorf("logql: invalid regex pattern %q: %w", pattern, err)
		return newQ
	}
	newQ.stages = append(newQ.stages, labelFilterStage{filter: LabelFilter{Label: label, Op: FilterRegexp, Value: pattern}})
	return newQ
}

// LabelNre adds a label not-regex filter: | label !~ "pattern".
func (q *LogQuery) LabelNre(label, pattern string) *LogQuery {
	newQ := q.clone()
	if _, err := regexp.Compile(pattern); err != nil {
		newQ.err = fmt.Errorf("logql: invalid regex pattern %q: %w", pattern, err)
		return newQ
	}
	newQ.stages = append(newQ.stages, labelFilterStage{filter: LabelFilter{Label: label, Op: FilterNotRegexp, Value: pattern}})
	return newQ
}

// LineFormat adds a line_format stage: | line_format "template".
func (q *LogQuery) LineFormat(template string) *LogQuery {
	newQ := q.clone()
	newQ.stages = append(newQ.stages, lineFormatStage{template: template})
	return newQ
}

// LabelFormatEntry adds a label_format stage with the given dst=src mappings.
// Each call adds a new label_format stage. To add multiple entries in one stage,
// pass multiple dst/src pairs using LabelFormatEntries.
func (q *LogQuery) LabelFormatEntry(dst, src string) *LogQuery {
	newQ := q.clone()
	newQ.stages = append(newQ.stages, labelFormatStage{entries: []labelFormatEntry{{dst: dst, src: src}}})
	return newQ
}

// LabelFormatEntries adds a label_format stage with multiple dst=src mappings.
func (q *LogQuery) LabelFormatEntries(entries map[string]string) *LogQuery {
	newQ := q.clone()
	lfEntries := make([]labelFormatEntry, 0, len(entries))
	for dst, src := range entries {
		lfEntries = append(lfEntries, labelFormatEntry{dst: dst, src: src})
	}
	newQ.stages = append(newQ.stages, labelFormatStage{entries: lfEntries})
	return newQ
}

// Drop adds a drop stage: | drop label1, label2.
func (q *LogQuery) Drop(labels ...string) *LogQuery {
	newQ := q.clone()
	newQ.stages = append(newQ.stages, dropStage{labels: labels})
	return newQ
}

// Keep adds a keep stage: | keep label1, label2.
func (q *LogQuery) Keep(labels ...string) *LogQuery {
	newQ := q.clone()
	newQ.stages = append(newQ.stages, keepStage{labels: labels})
	return newQ
}

// Decolorize adds a decolorize stage: | decolorize.
func (q *LogQuery) Decolorize() *LogQuery {
	newQ := q.clone()
	newQ.stages = append(newQ.stages, decolorizeStage{})
	return newQ
}

// Unwrap adds an unwrap stage: | unwrap label.
func (q *LogQuery) Unwrap(label string) *LogQuery {
	newQ := q.clone()
	newQ.stages = append(newQ.stages, unwrapStage{label: label})
	return newQ
}

// Build renders the full LogQL query string.
func (q *LogQuery) Build() (string, error) {
	if q.err != nil {
		return "", q.err
	}
	if len(q.matchers) == 0 {
		return "", errors.New("logql: at least one stream selector is required")
	}

	var sb strings.Builder

	sb.WriteByte('{')
	for i, m := range q.matchers {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(m.Label)
		sb.WriteString(string(m.Op))
		sb.WriteByte('"')
		sb.WriteString(m.Value)
		sb.WriteByte('"')
	}
	sb.WriteByte('}')

	for _, s := range q.stages {
		sb.WriteString(" ")
		sb.WriteString(s.String())
	}

	return sb.String(), nil
}

// String renders the query string, panicking on error.
func (q *LogQuery) String() string {
	s, err := q.Build()
	if err != nil {
		panic(err)
	}
	return s
}
