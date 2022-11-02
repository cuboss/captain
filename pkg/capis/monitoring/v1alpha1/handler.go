package v1alpha1

import (
	"regexp"
	"strings"

	"github.com/emicklei/go-restful"

	"captain/pkg/api"
	model "captain/pkg/models/monitoring"
	"captain/pkg/simple/client/monitoring"
)

type handler struct {
	mo model.MonitoringOperator
}

func NewHandler(monitoringClient monitoring.Interface) *handler {
	return &handler{
		mo: model.NewMonitoringOperator(monitoringClient),
	}
}

func (h handler) handleClusterMetricsQuery(req *restful.Request, resp *restful.Response) {
	params := parseRequestParams(req)
	opt, err := h.makeQueryOptions(params, monitoring.LevelCluster)
	if err != nil {
		api.HandleBadRequest(resp, nil, err)
		return
	}
	h.handleNamedMetricsQuery(resp, opt)
}

func (h handler) handleNodeMetricsQuery(req *restful.Request, resp *restful.Response) {
	params := parseRequestParams(req)
	opt, err := h.makeQueryOptions(params, monitoring.LevelNode)
	if err != nil {
		api.HandleBadRequest(resp, nil, err)
		return
	}
	h.handleNamedMetricsQuery(resp, opt)
}

func (h handler) handleWorkloadMetricsQuery(req *restful.Request, resp *restful.Response) {
	params := parseRequestParams(req)
	opt, err := h.makeQueryOptions(params, monitoring.LevelWorkload)
	if err != nil {
		if err.Error() == ErrNoHit {
			res := handleNoHit(opt.namedMetrics)
			resp.WriteAsJson(res)
			return
		}

		api.HandleBadRequest(resp, nil, err)
		return
	}
	h.handleNamedMetricsQuery(resp, opt)
}

func (h handler) handlePodMetricsQuery(req *restful.Request, resp *restful.Response) {
	params := parseRequestParams(req)
	opt, err := h.makeQueryOptions(params, monitoring.LevelPod)
	if err != nil {
		if err.Error() == ErrNoHit {
			res := handleNoHit(opt.namedMetrics)
			resp.WriteAsJson(res)
			return
		}

		api.HandleBadRequest(resp, nil, err)
		return
	}
	h.handleNamedMetricsQuery(resp, opt)
}

func (h handler) handleContainerMetricsQuery(req *restful.Request, resp *restful.Response) {
	params := parseRequestParams(req)
	opt, err := h.makeQueryOptions(params, monitoring.LevelContainer)
	if err != nil {
		if err.Error() == ErrNoHit {
			res := handleNoHit(opt.namedMetrics)
			resp.WriteAsJson(res)
			return
		}

		api.HandleBadRequest(resp, nil, err)
		return
	}
	h.handleNamedMetricsQuery(resp, opt)
}

func handleNoHit(namedMetrics []string) model.Metrics {
	var res model.Metrics
	for _, metic := range namedMetrics {
		res.Results = append(res.Results, monitoring.Metric{
			MetricName: metic,
			MetricData: monitoring.MetricData{},
		})
	}
	return res
}

func (h handler) handleNamedMetricsQuery(resp *restful.Response, q queryOptions) {
	var res model.Metrics

	var metrics []string
	for _, metric := range q.namedMetrics {
		if strings.HasPrefix(metric, model.MetricMeterPrefix) {
			// skip meter metric
			continue
		}
		ok, _ := regexp.MatchString(q.metricFilter, metric)
		if ok {
			metrics = append(metrics, metric)
		}
	}
	if len(metrics) == 0 {
		resp.WriteAsJson(res)
		return
	}

	if q.isRangeQuery() {
		res = h.mo.GetNamedMetricsOverTime(metrics, q.start, q.end, q.step, q.option)
	} else {
		res = h.mo.GetNamedMetrics(metrics, q.time, q.option)
		res = *res.Sort(q.target, q.order, q.identifier).Page(q.page, q.limit)
		if q.shouldSort() {
			res = *res.Sort(q.target, q.order, q.identifier).Page(q.page, q.limit)
		}
	}
	resp.WriteAsJson(res)
}
