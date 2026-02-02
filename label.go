package logql

// MatchOp represents a label matching operator used in stream selectors.
type MatchOp string

const (
	MatchEqual     MatchOp = "="
	MatchNotEqual  MatchOp = "!="
	MatchRegexp    MatchOp = "=~"
	MatchNotRegexp MatchOp = "!~"
)

// LabelMatcher represents a stream selector label matcher, e.g. {job="api"}.
type LabelMatcher struct {
	Label string
	Op    MatchOp
	Value string
}

// FilterOp represents a label filter comparison operator used in pipeline stages.
type FilterOp string

const (
	FilterEqual        FilterOp = "=="
	FilterNotEqual     FilterOp = "!="
	FilterGreater      FilterOp = ">"
	FilterGreaterEqual FilterOp = ">="
	FilterLess         FilterOp = "<"
	FilterLessEqual    FilterOp = "<="
	FilterRegexp       FilterOp = "=~"
	FilterNotRegexp    FilterOp = "!~"
)

// LabelFilter represents a pipeline label filter, e.g. | status >= 400.
type LabelFilter struct {
	Label string
	Op    FilterOp
	Value string
}
