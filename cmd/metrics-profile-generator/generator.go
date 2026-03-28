package main

import (
	"fmt"
	"os"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

// Generator collects metric definitions and outputs YAML.
type Generator struct {
	metrics []metricDefinition
}

// ---------------------------------------------------------------------------
// Core add methods
// ---------------------------------------------------------------------------

// Add appends a range metric definition.
func (g *Generator) Add(metricName string, query string) {
	g.metrics = append(g.metrics, metricDefinition{
		MetricName: metricName,
		Query:      query,
	})
}

// AddInstant appends an instant metric definition.
func (g *Generator) AddInstant(metricName string, query string, captureStart bool) {
	g.metrics = append(g.metrics, metricDefinition{
		MetricName:   metricName,
		Query:        query,
		Instant:      true,
		CaptureStart: captureStart,
	})
}

// AddQuery appends a metric from a Query builder.
func (g *Generator) AddQuery(metricName string, q *Query) {
	g.Add(metricName, q.String())
}

// AddQueryInstant appends an instant metric from a Query builder.
func (g *Generator) AddQueryInstant(metricName string, q *Query, captureStart bool) {
	g.AddInstant(metricName, q.String(), captureStart)
}

// ---------------------------------------------------------------------------
// Batch helpers
// ---------------------------------------------------------------------------

// HistogramQuantiles generates histogram_quantile queries for each percentile.
func (g *Generator) HistogramQuantiles(namePrefix string, metric Metric, percentiles []Percentile, filters string, groupBy []GroupBy) {
	for _, p := range percentiles {
		q := Q(metric, filters).Rate(Rate2m).Agg(AggSum, append(groupBy, GroupByLE)...).HistogramQuantile(p)
		g.Add(namePrefix+p.Label, q.String())
	}
}

// HistogramQuantilesIRate is like HistogramQuantiles but uses irate.
func (g *Generator) HistogramQuantilesIRate(namePrefix string, metric Metric, percentiles []Percentile, interval RateInterval, filters string, groupBy []GroupBy) {
	for _, p := range percentiles {
		q := Q(metric, filters).IRate(interval).Agg(AggSum, append(groupBy, GroupByLE)...).HistogramQuantile(p)
		g.Add(namePrefix+p.Label, q.String())
	}
}

// RateForMetrics generates sum(rate(...)) for each metric.
func (g *Generator) RateForMetrics(metrics []Metric, interval RateInterval, filters string, groupBy []GroupBy) {
	for _, m := range metrics {
		name := metricToName(m) + "Rate"
		q := Q(m, filters).Rate(interval).Agg(AggSum, groupBy...)
		g.Add(name, q.String())
	}
}

// IRateForMetrics generates sum(irate(...)) for each metric.
func (g *Generator) IRateForMetrics(metrics []Metric, interval RateInterval, filters string, groupBy []GroupBy) {
	for _, m := range metrics {
		name := metricToName(m) + "IRate"
		q := Q(m, filters).IRate(interval).Agg(AggSum, groupBy...)
		g.Add(name, q.String())
	}
}

// AggForMetrics generates agg(metric) for each metric.
func (g *Generator) AggForMetrics(fn AggFunc, metrics []Metric, filters string, groupBy []GroupBy) {
	for _, m := range metrics {
		name := metricToName(m)
		q := Q(m, filters).Agg(fn, groupBy...)
		g.Add(name, q.String())
	}
}

// ForNodeRoles repeats a query-building function for each node role.
func (g *Generator) ForNodeRoles(namePrefix string, roles []NodeRole, build func(role NodeRole) *Query) {
	for _, role := range roles {
		suffix := strings.ToUpper(string(role)[:1]) + string(role)[1:]
		g.Add(namePrefix+suffix, build(role).String())
	}
}

// ForNodeRolesInstant is like ForNodeRoles but produces instant queries.
func (g *Generator) ForNodeRolesInstant(namePrefix string, roles []NodeRole, captureStart bool, build func(role NodeRole) *Query) {
	for _, role := range roles {
		suffix := strings.ToUpper(string(role)[:1]) + string(role)[1:]
		g.AddInstant(namePrefix+suffix, build(role).String(), captureStart)
	}
}

// CustomTemplate generates metrics from Go template strings.
func (g *Generator) CustomTemplate(queryTmpl string, nameTmpl string, vars []map[string]string, instant bool) {
	qt := template.Must(template.New("query").Parse(queryTmpl))
	nt := template.Must(template.New("name").Parse(nameTmpl))

	for _, v := range vars {
		var qb, nb strings.Builder
		if err := qt.Execute(&qb, v); err != nil {
			fmt.Fprintf(os.Stderr, "Error executing query template: %v\n", err)
			os.Exit(1)
		}
		if err := nt.Execute(&nb, v); err != nil {
			fmt.Fprintf(os.Stderr, "Error executing name template: %v\n", err)
			os.Exit(1)
		}
		if instant {
			g.AddInstant(nb.String(), qb.String(), false)
		} else {
			g.Add(nb.String(), qb.String())
		}
	}
}

// ---------------------------------------------------------------------------
// Output
// ---------------------------------------------------------------------------

// Generate returns the metrics profile as YAML bytes.
func (g *Generator) Generate() ([]byte, error) {
	return yaml.Marshal(g.metrics)
}

// Count returns the number of generated metric definitions.
func (g *Generator) Count() int {
	return len(g.metrics)
}
