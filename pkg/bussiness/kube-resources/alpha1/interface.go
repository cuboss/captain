package alpha1

import (
	"captain/pkg/unify/query"
	"captain/pkg/unify/response"
	"math"
	"sort"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type KubeResProvider interface {
	// Get retrieves a single object by its namespace and name
	Get(namespace, name string) (runtime.Object, error)

	// List retrieves a collection of objects matches given query
	List(namespace string, query *query.QueryInfo) (*response.ListResult, error)
}

type MultiClusterKubeResProvider interface {
	// Get retrieves a single object by its namespace and name
	Get(region, cluster, namespace, name string) (runtime.Object, error)

	// List retrieves a collection of objects matches given query
	List(region, cluster, namespace string, query *query.QueryInfo) (*response.ListResult, error)
}

// CompareFunc return true is left great than right
type CompareFunc func(runtime.Object, runtime.Object, query.Field) bool

type FilterFunc func(runtime.Object, query.Filter) bool

type TransformFunc func(runtime.Object) runtime.Object

func DefaultList(objects []runtime.Object, q *query.QueryInfo, compareFunc CompareFunc, filterFunc FilterFunc, transferFuncs ...TransformFunc) *response.ListResult {

	var filtered []runtime.Object
	for _, obj := range objects {
		//is targeted by such filter key/values
		targeted := true
		for k, v := range q.Filters {
			if !filterFunc(obj, query.Filter{Field: k, Value: v}) {
				targeted = false
			}
		}

		if targeted {
			for _, transform := range transferFuncs {
				obj = transform(obj)
			}
			filtered = append(filtered, obj)
		}
	}

	//sort by some field
	sort.Slice(filtered, func(i, j int) bool {
		//Ascending
		if q.Ascending {
			return compareFunc(filtered[i], filtered[j], q.SortBy)
		}
		return !compareFunc(filtered[i], filtered[j], q.SortBy)
	})

	//summarize

	total := len(filtered)

	begin, end := q.Pagination.GetValidPagination(total)

	return &response.ListResult{
		Total:       total,
		CurrentPage: q.Pagination.Page,
		PageSize:    q.Pagination.PageSize,
		TotalPages:  int(math.Ceil(float64(total) / float64(q.Pagination.PageSize))),
		Items:       objects2Interfaces(filtered[begin:end]),
	}
}

func objects2Interfaces(objs []runtime.Object) []interface{} {
	res := make([]interface{}, 0)
	for _, obj := range objs {
		res = append(res, obj)
	}
	return res
}

// DefaultObjectMetaCompare ... return true if left gt right
func DefaultObjectMetaCompare(left, right metav1.ObjectMeta, sortBy query.Field) bool {

	switch sortBy {
	default:
		fallthrough
	case query.FieldName:
		return strings.Compare(left.GetName(), right.GetName()) < 0
		//	?sortBy=creationTimestamp
	case query.FieldCreateTime:
		fallthrough
	case query.FieldCreationTimeStamp:
		if left.CreationTimestamp.Equal(&right.CreationTimestamp) {
			return strings.Compare(left.GetName(), right.GetName()) < 0
		}
		return left.CreationTimestamp.After(right.CreationTimestamp.Time)
	}
}

// Default metadata filter
func DefaultObjectMetaFilter(item metav1.ObjectMeta, filter query.Filter) bool {
	switch filter.Field {
	case query.FieldNames:
		for _, name := range strings.Split(string(filter.Value), ",") {
			if item.Name == name {
				return true
			}
		}
		return false
	// /namespaces?page=1&limit=10&name=default
	case query.FieldName:
		return strings.Contains(item.Name, string(filter.Value))
		// /namespaces?page=1&limit=10&uid=a8a8d6cf-f6a5-4fea-9c1b-e57610115706
	case query.FieldUID:
		return strings.Compare(string(item.UID), string(filter.Value)) == 0
		// /deployments?page=1&limit=10&namespace=kubesphere-system
	case query.FieldNamespace:
		return strings.Compare(item.Namespace, string(filter.Value)) == 0
		// /namespaces?page=1&limit=10&ownerReference=a8a8d6cf-f6a5-4fea-9c1b-e57610115706
	case query.FieldOwnerReference:
		for _, ownerReference := range item.OwnerReferences {
			if strings.Compare(string(ownerReference.UID), string(filter.Value)) == 0 {
				return true
			}
		}
		return false
		// /namespaces?page=1&limit=10&ownerKind=Workspace
	case query.FieldOwnerKind:
		for _, ownerReference := range item.OwnerReferences {
			if strings.Compare(ownerReference.Kind, string(filter.Value)) == 0 {
				return true
			}
		}
		return false
		// /namespaces?page=1&limit=10&annotation=openpitrix_runtime
	case query.FieldAnnotation:
		return labelMatch(item.Annotations, string(filter.Value))
		// /namespaces?page=1&limit=10&label=kubesphere.io/workspace:system-workspace
	case query.FieldLabel:
		return labelMatch(item.Labels, string(filter.Value))
	default:
		return false
	}
}

func labelMatch(labels map[string]string, filter string) bool {
	fields := strings.SplitN(filter, "=", 2)
	var key, value string
	var opposite bool
	if len(fields) == 2 {
		key = fields[0]
		if strings.HasSuffix(key, "!") {
			key = strings.TrimSuffix(key, "!")
			opposite = true
		}
		value = fields[1]
	} else {
		key = fields[0]
		value = "*"
	}
	for k, v := range labels {
		if opposite {
			if (k == key) && v != value {
				return true
			}
		} else {
			if (k == key) && (value == "*" || v == value) {
				return true
			}
		}
	}
	return false
}
