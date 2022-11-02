package monitoring

import "time"

type Interface interface {
	GetNamedMetrics(metrics []string, time time.Time, opt QueryOption) []Metric
	GetNamedMetricsOverTime(metrics []string, start, end time.Time, step time.Duration, opt QueryOption) []Metric
}
