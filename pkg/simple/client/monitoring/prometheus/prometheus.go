package prometheus

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/api"
	apiv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"

	"captain/pkg/simple/client/monitoring"
)

const MeteringDefaultTimeout = 20 * time.Second

type prometheus struct {
	client apiv1.API
}

type roundTripper struct {
	auth      PrometheusAuth
	transport http.RoundTripper
}

func (r *roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if len(r.auth.Basic.Username) > 0 && len(r.auth.Basic.Password) > 0 {
		req.SetBasicAuth(r.auth.Basic.Username, r.auth.Basic.Password)
	}
	return r.transport.RoundTrip(req)
}

func NewPrometheus(options *Options) (monitoring.Interface, error) {
	cfg := api.Config{
		Address: options.Endpoint,
		RoundTripper: &roundTripper{
			auth:      options.Auth,
			transport: api.DefaultRoundTripper,
		},
	}

	client, err := api.NewClient(cfg)

	return &prometheus{client: apiv1.NewAPI(client)}, err
}

func (p *prometheus) GetNamedMetrics(metrics []string, ts time.Time, o monitoring.QueryOption) []monitoring.Metric {
	var res []monitoring.Metric
	var mtx sync.Mutex
	var wg sync.WaitGroup

	opts := monitoring.NewQueryOptions()
	o.Apply(opts)

	for _, metric := range metrics {
		wg.Add(1)
		go func(metric string) {
			parsedResp := monitoring.Metric{MetricName: metric}

			value, _, err := p.client.Query(context.Background(), makeExpr(metric, *opts), ts)
			if err != nil {
				parsedResp.Error = err.Error()
			} else {
				parsedResp.MetricData = parseQueryResp(value, genMetricFilter(o))
			}

			mtx.Lock()
			res = append(res, parsedResp)
			mtx.Unlock()

			wg.Done()
		}(metric)
	}

	wg.Wait()

	return res
}

func (p *prometheus) GetNamedMetricsOverTime(metrics []string, start, end time.Time, step time.Duration, o monitoring.QueryOption) []monitoring.Metric {
	var res []monitoring.Metric
	var mtx sync.Mutex
	var wg sync.WaitGroup

	opts := monitoring.NewQueryOptions()
	o.Apply(opts)

	timeRange := apiv1.Range{
		Start: start,
		End:   end,
		Step:  step,
	}

	for _, metric := range metrics {
		wg.Add(1)
		go func(metric string) {
			parsedResp := monitoring.Metric{MetricName: metric}

			value, _, err := p.client.QueryRange(context.Background(), makeExpr(metric, *opts), timeRange)
			if err != nil {
				parsedResp.Error = err.Error()
			} else {
				parsedResp.MetricData = parseQueryRangeResp(value, genMetricFilter(o))
			}

			mtx.Lock()
			res = append(res, parsedResp)
			mtx.Unlock()

			wg.Done()
		}(metric)
	}

	wg.Wait()

	return res
}

func parseQueryRangeResp(value model.Value, metricFilter func(metric model.Metric) bool) monitoring.MetricData {
	res := monitoring.MetricData{MetricType: monitoring.MetricTypeMatrix}

	data, _ := value.(model.Matrix)

	for _, v := range data {
		if metricFilter != nil && !metricFilter(v.Metric) {
			continue
		}
		mv := monitoring.MetricValue{
			Metadata: make(map[string]string),
		}

		for k, v := range v.Metric {
			mv.Metadata[string(k)] = string(v)
		}

		for _, k := range v.Values {
			mv.Series = append(mv.Series, monitoring.Point{float64(k.Timestamp) / 1000, float64(k.Value)})
		}

		res.MetricValues = append(res.MetricValues, mv)
	}

	return res
}

func parseQueryResp(value model.Value, metricFilter func(metric model.Metric) bool) monitoring.MetricData {
	var res monitoring.MetricData

	if value.Type() == model.ValVector {
		res = monitoring.MetricData{MetricType: monitoring.MetricTypeVector}
		data, _ := value.(model.Vector)
		for _, v := range data {
			if metricFilter != nil && !metricFilter(v.Metric) {
				continue
			}
			mv := monitoring.MetricValue{
				Metadata: make(map[string]string),
			}

			for k, v := range v.Metric {
				mv.Metadata[string(k)] = string(v)
			}

			mv.Sample = &monitoring.Point{float64(v.Timestamp) / 1000, float64(v.Value)}

			res.MetricValues = append(res.MetricValues, mv)
		}
	}

	return res
}

func genMetricFilter(o monitoring.QueryOption) func(metric model.Metric) bool {
	if o != nil {
		if po, ok := o.(monitoring.PodOption); ok {
			if po.NamespacedResourcesFilter != "" {
				namespacedPodsMap := make(map[string]struct{})
				for _, s := range strings.Split(po.NamespacedResourcesFilter, "|") {
					namespacedPodsMap[s] = struct{}{}
				}
				return func(metric model.Metric) bool {
					if len(metric) == 0 {
						return false
					}
					_, ok := namespacedPodsMap[string(metric["namespace"])+"/"+string(metric["pod"])]
					return ok
				}
			}
		}
	}
	return func(metric model.Metric) bool {
		return true
	}
}
