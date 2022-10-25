package monitoring

import (
	"time"

	"captain/pkg/simple/client/monitoring"
)

type MonitoringOperator interface {
	GetNamedMetrics(metrics []string, time time.Time, opt monitoring.QueryOption) Metrics
	GetNamedMetricsOverTime(metrics []string, start, end time.Time, step time.Duration, opt monitoring.QueryOption) Metrics
}

type monitoringOperator struct {
	prometheus monitoring.Interface
}

func NewMonitoringOperator(monitoringClient monitoring.Interface) MonitoringOperator {
	return &monitoringOperator{
		prometheus: monitoringClient,
	}
}

func (mo monitoringOperator) GetNamedMetrics(metrics []string, time time.Time, opt monitoring.QueryOption) Metrics {
	ress := mo.prometheus.GetNamedMetrics(metrics, time, opt)
	return Metrics{Results: ress}
}

func (mo monitoringOperator) GetNamedMetricsOverTime(metrics []string, start, end time.Time, step time.Duration, opt monitoring.QueryOption) Metrics {
	ress := mo.prometheus.GetNamedMetricsOverTime(metrics, start, end, step, opt)
	return Metrics{Results: ress}
}
