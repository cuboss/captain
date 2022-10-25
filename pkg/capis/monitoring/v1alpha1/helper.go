package v1alpha1

import (
	"time"

	"github.com/emicklei/go-restful"

	model "captain/pkg/models/monitoring"
	"captain/pkg/simple/client/monitoring"
)

const (
	DefaultFilter = ".*"
)

type reqParams struct {
	metricFilter string
	time         string
	start        string
	end          string
	step         string
}

func parseRequestParams(req *restful.Request) reqParams {
	var r reqParams
	r.time = req.QueryParameter("time")
	r.start = req.QueryParameter("start")
	r.end = req.QueryParameter("end")
	r.step = req.QueryParameter("step")
	r.metricFilter = req.QueryParameter("metrics_filter")
	return r
}

type queryOptions struct {
	metricFilter string
	namedMetrics []string

	start time.Time
	end   time.Time
	time  time.Time
	step  time.Duration

	target     string
	identifier string
	order      string
	page       int
	limit      int

	option monitoring.QueryOption
}

func (q queryOptions) isRangeQuery() bool {
	return q.time.IsZero()
}

func (q queryOptions) shouldSort() bool {
	return q.target != "" && q.identifier != ""
}

func (h handler) makeQueryOptions(r reqParams, lvl monitoring.Level) (q queryOptions, err error) {

	q.metricFilter = r.metricFilter
	if r.metricFilter == "" {
		q.metricFilter = DefaultFilter
	}

	switch lvl {
	case monitoring.LevelCluster:
		q.option = monitoring.ClusterOption{}
		q.namedMetrics = model.ClusterMetrics
	}

	// Parse time params
	if r.start != "" && r.end != "" {

	} else if r.start == "" && r.end == "" {
		if r.time == "" {
			q.time = time.Now()
		}
	}
	return q, nil
}
