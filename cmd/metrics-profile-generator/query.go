package main

import (
	"fmt"
	"strings"
)

// ---------------------------------------------------------------------------
// Query builder: fluent API for composing PromQL expressions
// ---------------------------------------------------------------------------

// Query is a fluent builder for PromQL expressions.
type Query struct {
	expr string
}

// Q starts a new query from a metric with optional filters.
func Q(metric Metric, filters string) *Query {
	return &Query{expr: string(metric) + wrapFilters(filters)}
}

// Raw starts a query from a raw PromQL string (escape hatch).
func Raw(expr string) *Query {
	return &Query{expr: expr}
}

// Rate wraps in rate(...)[interval].
func (q *Query) Rate(interval RateInterval) *Query {
	q.expr = fmt.Sprintf("rate(%s[%s])", q.expr, interval)
	return q
}

// IRate wraps in irate(...)[interval].
func (q *Query) IRate(interval RateInterval) *Query {
	q.expr = fmt.Sprintf("irate(%s[%s])", q.expr, interval)
	return q
}

// Agg wraps in an aggregation function with optional group-by labels.
func (q *Query) Agg(fn AggFunc, groupBy ...GroupBy) *Query {
	if len(groupBy) > 0 {
		q.expr = fmt.Sprintf("%s(%s) by (%s)", fn, q.expr, joinGroupBy(groupBy))
	} else {
		q.expr = fmt.Sprintf("%s(%s)", fn, q.expr)
	}
	return q
}

// TopK wraps in topk(k, ...).
func (q *Query) TopK(k int) *Query {
	q.expr = fmt.Sprintf("topk(%d, %s)", k, q.expr)
	return q
}

// HistogramQuantile wraps in histogram_quantile(quantile, ...).
func (q *Query) HistogramQuantile(p Percentile) *Query {
	q.expr = fmt.Sprintf("histogram_quantile(%s, %s)", p.Value, q.expr)
	return q
}

// Multiply scales by a factor.
func (q *Query) Multiply(factor string) *Query {
	q.expr = fmt.Sprintf("%s * %s", q.expr, factor)
	return q
}

// Gt appends > threshold.
func (q *Query) Gt(threshold string) *Query {
	q.expr = fmt.Sprintf("%s > %s", q.expr, threshold)
	return q
}

// Gte appends >= threshold.
func (q *Query) Gte(threshold string) *Query {
	q.expr = fmt.Sprintf("%s >= %s", q.expr, threshold)
	return q
}

// AndOn appends "and on (labels) rightExpr".
func (q *Query) AndOn(labels []GroupBy, right *Query) *Query {
	q.expr = fmt.Sprintf("%s and on (%s) %s", q.expr, joinGroupBy(labels), right.expr)
	return q
}

// And appends "and rightExpr".
func (q *Query) And(right *Query) *Query {
	q.expr = fmt.Sprintf("%s and %s", q.expr, right.expr)
	return q
}

// Or appends "or rightExpr".
func (q *Query) Or(right *Query) *Query {
	q.expr = fmt.Sprintf("%s or %s", q.expr, right.expr)
	return q
}

// Sub subtracts another query.
func (q *Query) Sub(right *Query) *Query {
	q.expr = fmt.Sprintf("%s - %s", q.expr, right.expr)
	return q
}

// Div divides by another query or constant.
func (q *Query) Div(right *Query) *Query {
	q.expr = fmt.Sprintf("%s / %s", q.expr, right.expr)
	return q
}

// OverTime wraps in a time aggregation over the elapsed job duration.
// Produces e.g. avg_over_time(...[{{.elapsed}}:])
func (q *Query) OverTime(fn TimeAggFunc) *Query {
	q.expr = fmt.Sprintf("%s(%s[{{.elapsed}}:])", fn, q.expr)
	return q
}

// OverTimeStep wraps in a time aggregation with a custom step.
// Produces e.g. avg_over_time(...[{{.elapsed}}:30s])
func (q *Query) OverTimeStep(fn TimeAggFunc, step string) *Query {
	q.expr = fmt.Sprintf("%s(%s[{{.elapsed}}:%s])", fn, q.expr, step)
	return q
}

// Delta wraps in delta(...)[range:step].
func (q *Query) Delta(rangeStr string, step string) *Query {
	q.expr = fmt.Sprintf("delta(%s[%s:%s])", q.expr, rangeStr, step)
	return q
}

// LabelReplace wraps in label_replace(...).
func (q *Query) LabelReplace(dst, replacement, src, regex string) *Query {
	q.expr = fmt.Sprintf(`label_replace(%s, "%s", "%s", "%s", "%s")`, q.expr, dst, replacement, src, regex)
	return q
}

// Paren wraps the expression in parentheses.
func (q *Query) Paren() *Query {
	q.expr = fmt.Sprintf("(%s)", q.expr)
	return q
}

// String returns the built PromQL expression.
func (q *Query) String() string {
	return q.expr
}

// ---------------------------------------------------------------------------
// Convenience constructors
// ---------------------------------------------------------------------------

// NodeRoleFilter returns a kube_node_role{role="..."} query for use with AndOn.
func NodeRoleFilter(role NodeRole) *Query {
	return Q(MetricKubeNodeRole, fmt.Sprintf(`role="%s"`, role))
}

// NodeRoleLabelReplace returns a label_replace that maps "node" -> "instance"
// for joining node-level metrics with kube_node_role.
func NodeRoleLabelReplace(role NodeRole) *Query {
	return Q(MetricKubeNodeRole, fmt.Sprintf(`role="%s"`, role)).
		LabelReplace("instance", "$1", "node", "(.+)")
}

// VectorZero returns vector(0) for use with Or as a default.
func VectorZero() *Query {
	return Raw("vector(0)")
}

// ---------------------------------------------------------------------------
// Utilities
// ---------------------------------------------------------------------------

func wrapFilters(filters string) string {
	if filters == "" {
		return "{}"
	}
	return "{" + filters + "}"
}

func joinGroupBy(groups []GroupBy) string {
	parts := make([]string, len(groups))
	for i, g := range groups {
		parts[i] = string(g)
	}
	return strings.Join(parts, ",")
}

func metricToName(m Metric) string {
	s := string(m)
	for _, suffix := range []string{"_total", "_bytes", "_seconds"} {
		s = strings.TrimSuffix(s, suffix)
	}
	parts := strings.Split(s, "_")
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
}
