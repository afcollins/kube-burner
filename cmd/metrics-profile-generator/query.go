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

// Q starts a new query from a metric with filters.
// Pass "" for no filters (emits metric{}).
func Q(metric Metric, filters string) *Query {
	return &Query{expr: string(metric) + wrapFilters(filters)}
}

// QRaw starts a query from a metric name without appending {}.
// Use for recording rules like "cluster:memory_usage:ratio" or bare metrics like "kube_node_role".
func QRaw(metric Metric) *Query {
	return &Query{expr: string(metric)}
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

// BucketRate wraps in rate() with _bucket appended to the metric name (before any filters).
// Handles both "metric{}" → "metric_bucket{}" and "metric{filters}" → "metric_bucket{filters}".
func (q *Query) BucketRate(interval RateInterval) *Query {
	q.expr = fmt.Sprintf("rate(%s[%s])", insertBucket(q.expr), interval)
	return q
}

// BucketIRate wraps in irate() with _bucket appended to the metric name (before any filters).
func (q *Query) BucketIRate(interval RateInterval) *Query {
	q.expr = fmt.Sprintf("irate(%s[%s])", insertBucket(q.expr), interval)
	return q
}

// insertBucket inserts "_bucket" before the first "{" in the expression,
// or appends it if there's no "{".
func insertBucket(expr string) string {
	if idx := strings.Index(expr, "{"); idx >= 0 {
		return expr[:idx] + "_bucket" + expr[idx:]
	}
	return expr + "_bucket"
}

// Agg wraps in an aggregation function with optional group-by labels.
// Produces: fn(expr) by (labels)
func (q *Query) Agg(fn AggFunc, groupBy ...GroupBy) *Query {
	if len(groupBy) > 0 {
		q.expr = fmt.Sprintf("%s(%s) by (%s)", fn, q.expr, joinGroupBy(groupBy))
	} else {
		q.expr = fmt.Sprintf("%s(%s)", fn, q.expr)
	}
	return q
}

// AggBy wraps in an aggregation with "by" before the expression.
// Produces: fn by (labels) (expr)
// TODO: Remove after all existing profiles are regenerated; only exists to match legacy PromQL syntax.
func (q *Query) AggBy(fn AggFunc, groupBy ...GroupBy) *Query {
	if len(groupBy) > 0 {
		q.expr = fmt.Sprintf("%s by (%s)(%s)", fn, joinGroupBy(groupBy), q.expr)
	} else {
		q.expr = fmt.Sprintf("%s(%s)", fn, q.expr)
	}
	return q
}

// AggBySpaced wraps in an aggregation with "by" before the expression and spaces inside parens.
// Produces: fn by (labels) ( expr )
// TODO: Remove after all existing profiles are regenerated; only exists to match legacy whitespace.
func (q *Query) AggBySpaced(fn AggFunc, groupBy ...GroupBy) *Query {
	q.expr = fmt.Sprintf("%s by (%s) ( %s )", fn, joinGroupBy(groupBy), q.expr)
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

// Div divides by another query.
func (q *Query) Div(right *Query) *Query {
	q.expr = fmt.Sprintf("%s / %s", q.expr, right.expr)
	return q
}

// DivConst divides by a constant.
func (q *Query) DivConst(divisor string) *Query {
	q.expr = fmt.Sprintf("%s/%s", q.expr, divisor)
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

// By appends a bare "by (labels)" clause (for use after Agg without labels, or after Paren).
func (q *Query) By(groupBy ...GroupBy) *Query {
	q.expr = fmt.Sprintf("%s by (%s)", q.expr, joinGroupBy(groupBy))
	return q
}

// SpacedBy appends "by (labels)" with leading space inside sum( ... ) patterns.
// Produces e.g. "sum( expr ) by (mode)"
// TODO: Remove after all existing profiles are regenerated; only exists to match legacy whitespace.
func (q *Query) SpacedBy(groupBy ...GroupBy) *Query {
	q.expr = fmt.Sprintf("%s by (%s)", q.expr, joinGroupBy(groupBy))
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

// NodeRoleFilterExclude returns kube_node_role{role="x",role!="y"} for excluding a role.
func NodeRoleFilterExclude(include NodeRole, exclude NodeRole) *Query {
	return Q(MetricKubeNodeRole, fmt.Sprintf(`role="%s",role!="%s"`, include, exclude))
}

// NodeRoleLabelReplace returns a label_replace that maps "node" -> "instance"
// for joining node-level metrics with kube_node_role.
func NodeRoleLabelReplace(role NodeRole) *Query {
	return Q(MetricKubeNodeRole, fmt.Sprintf(`role="%s"`, role)).
		LabelReplace("instance", "$1", "node", "(.+)")
}

// NodeRoleLabelReplaceExclude returns label_replace with role include + exclude.
func NodeRoleLabelReplaceExclude(include NodeRole, exclude NodeRole) *Query {
	return Q(MetricKubeNodeRole, fmt.Sprintf(`role="%s",role!="%s"`, include, exclude)).
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
