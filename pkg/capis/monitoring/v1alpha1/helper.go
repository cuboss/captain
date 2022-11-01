package v1alpha1

import (
	"strconv"
	"strings"
	"time"

	"github.com/emicklei/go-restful"
	"github.com/pkg/errors"

	model "captain/pkg/models/monitoring"
	"captain/pkg/simple/client/monitoring"
)

const (
	DefaultStep   = 10 * time.Minute
	DefaultFilter = ".*"
	DefaultOrder  = model.OrderDescending
	DefaultPage   = 1
	DefaultLimit  = 5

	ErrNoHit             = "'end' or 'time' must be after the namespace creation time."
	ErrParamConflict     = "'time' and the combination of 'start' and 'end' are mutually exclusive."
	ErrInvalidStartEnd   = "'start' must be before 'end'."
	ErrInvalidPage       = "Invalid parameter 'page'."
	ErrInvalidLimit      = "Invalid parameter 'limit'."
	ErrParameterNotfound = "Parmameter [%s] not found"
)

type reqParams struct {
	time  string
	start string
	end   string
	step  string

	target string
	order  string
	page   string
	limit  string

	metricFilter              string
	resourceFilter            string
	namespacedResourcesFilter string

	nodeName      string
	namespaceName string
	workloadKind  string
	workloadName  string
	podName       string
	containerName string
}

func parseRequestParams(req *restful.Request) reqParams {
	var r reqParams
	r.time = req.QueryParameter("time")
	r.start = req.QueryParameter("start")
	r.end = req.QueryParameter("end")
	r.step = req.QueryParameter("step")

	r.target = req.QueryParameter("sort_metric")
	r.order = req.QueryParameter("sort_type")
	r.page = req.QueryParameter("page")
	r.limit = req.QueryParameter("limit")

	r.resourceFilter = req.QueryParameter("resources_filter")
	r.metricFilter = req.QueryParameter("metrics_filter")
	// namespacedResourcesFilter supports only <namespace>/<pod_name>|<namespace>/<pod_name> format
	// which is different from resources_filter or metrics_filter, so wipe off the possible $ at the end.
	r.namespacedResourcesFilter = strings.TrimRight(req.QueryParameter("namespaced_resources_filter"), "$")

	r.nodeName = req.PathParameter("node")
	r.namespaceName = req.PathParameter("namespace")
	r.workloadKind = req.PathParameter("kind")
	r.workloadName = req.PathParameter("workload")
	//will be overide if "pod" in the path parameter.
	r.podName = req.QueryParameter("pod")
	r.podName = req.PathParameter("pod")
	r.containerName = req.PathParameter("container")
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
	if r.resourceFilter == "" {
		r.resourceFilter = DefaultFilter
	}

	q.metricFilter = r.metricFilter
	if r.metricFilter == "" {
		q.metricFilter = DefaultFilter
	}

	switch lvl {
	case monitoring.LevelCluster:
		q.option = monitoring.ClusterOption{}
		q.namedMetrics = model.ClusterMetrics

	case monitoring.LevelNode:
		q.identifier = model.IdentifierNode
		q.option = monitoring.NodeOption{
			ResourceFilter: r.resourceFilter,
			NodeName:       r.nodeName,
		}
		q.namedMetrics = model.NodeMetrics

	case monitoring.LevelWorkload:
		q.identifier = model.IdentifierWorkload
		q.option = monitoring.WorkloadOption{
			ResourceFilter: r.resourceFilter,
			NamespaceName:  r.namespaceName,
			WorkloadKind:   r.workloadKind,
		}
		q.namedMetrics = model.WorkloadMetrics

	case monitoring.LevelPod:
		q.identifier = model.IdentifierPod
		q.option = monitoring.PodOption{
			NamespacedResourcesFilter: r.namespacedResourcesFilter,
			ResourceFilter:            r.resourceFilter,
			NodeName:                  r.nodeName,
			NamespaceName:             r.namespaceName,
			WorkloadKind:              r.workloadKind,
			WorkloadName:              r.workloadName,
			PodName:                   r.podName,
		}
		q.namedMetrics = model.PodMetrics

	case monitoring.LevelContainer:
		q.identifier = model.IdentifierContainer
		q.option = monitoring.ContainerOption{
			ResourceFilter: r.resourceFilter,
			NamespaceName:  r.namespaceName,
			PodName:        r.podName,
			ContainerName:  r.containerName,
		}
		q.namedMetrics = model.ContainerMetrics
	}

	// Parse time params
	if r.start != "" && r.end != "" {
		startInt, err := strconv.ParseInt(r.start, 10, 64)
		if err != nil {
			return q, err
		}
		q.start = time.Unix(startInt, 0)

		endInt, err := strconv.ParseInt(r.end, 10, 64)
		if err != nil {
			return q, err
		}
		q.end = time.Unix(endInt, 0)

		if r.step == "" {
			q.step = DefaultStep
		} else {
			q.step, err = time.ParseDuration(r.step)
			if err != nil {
				return q, err
			}
		}

		if q.start.After(q.end) {
			return q, errors.New(ErrInvalidStartEnd)
		}
	} else if r.start == "" && r.end == "" {
		if r.time == "" {
			q.time = time.Now()
		} else {
			timeInt, err := strconv.ParseInt(r.time, 10, 64)
			if err != nil {
				return q, err
			}
			q.time = time.Unix(timeInt, 0)
		}
	} else {
		return q, errors.Errorf(ErrParamConflict)
	}

	// Parse sorting and paging params
	if r.target != "" {
		q.target = r.target
		q.page = DefaultPage
		q.limit = DefaultLimit
		q.order = r.order
		if r.order != model.OrderAscending {
			q.order = DefaultOrder
		}
		if r.page != "" {
			q.page, err = strconv.Atoi(r.page)
			if err != nil || q.page <= 0 {
				return q, errors.New(ErrInvalidPage)
			}
		}
		if r.limit != "" {
			q.limit, err = strconv.Atoi(r.limit)
			if err != nil || q.limit <= 0 {
				return q, errors.New(ErrInvalidLimit)
			}
		}
	}

	return q, nil
}
