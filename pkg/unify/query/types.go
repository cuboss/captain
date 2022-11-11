package query

import (
	"fmt"
	"math"
	"strconv"

	"captain/pkg/utils/base"

	"github.com/emicklei/go-restful"
	"k8s.io/apimachinery/pkg/labels"
)

type Field string
type Value string

const (
	FieldName                = "name"
	FieldNames               = "names"
	FieldUID                 = "uid"
	FieldCreationTimeStamp   = "creationTimestamp"
	FieldCreateTime          = "createTime"
	FieldLastUpdateTimestamp = "lastUpdateTimestamp"
	FieldUpdateTime          = "updateTime"
	FieldStartTime           = "startTime"
	FieldLabel               = "label"
	FieldAnnotation          = "annotation"
	FieldNamespace           = "namespace"
	FieldStatus              = "status"
	FieldOwnerReference      = "ownerReference"
	FieldOwner               = "owner"
	FieldOwnerKind           = "ownerKind"
	FieldOwnerName           = "ownerName"
	FieldRole                = "role"
	FieldUserfacing          = "userfacing"
	FieldType                = "type"
)

const (
	ParameterName          = "name"
	ParameterLabelSelector = "labelSelector"
	ParameterFieldSelector = "fieldSelector"
	ParameterPage          = "page"
	// ParameterLimit         = "limit"
	ParameterPageSize  = "pageSize"
	ParameterOrderBy   = "sortBy"
	ParameterAscending = "ascending"
)

// Query represents api search terms
type QueryInfo struct {
	Pagination *Pagination

	// sort result in which field, default to FieldCreationTimeStamp
	SortBy Field

	// sort result in ascending or descending order, default to descending
	Ascending bool

	//
	Filters map[Field]Value

	LabelSelector string
}

func (q *QueryInfo) String() string {
	return fmt.Sprintf("Incoming Query info %T \n Pagination: { Page: %d, PageSize: %d} \n SortBy: %s, Ascending: %v, \n Filters: %+v \n ",
		q, q.Pagination.Page, q.Pagination.PageSize, q.SortBy, q.Ascending, q.Filters)
}

type Pagination struct {
	// items per page
	PageSize int

	// pageID in request query, started with 1
	Page int
}

type Filter struct {
	Field Field
	Value Value
}

func New() *QueryInfo {
	return &QueryInfo{
		Pagination: DefaultPagination,
		SortBy:     "",
		Ascending:  false,
		Filters:    map[Field]Value{},
	}
}

func (q *QueryInfo) GetSelector() labels.Selector {
	selector, err := labels.Parse(q.LabelSelector)
	if err != nil {
		return labels.Everything()
	}
	return selector
}

var DefaultPagination = newPagination(1, 10)

func newPagination(page, pageSize int) *Pagination {
	// handling invalid number
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	return &Pagination{
		Page:     page,
		PageSize: pageSize,
	}
}

func (p *Pagination) GetValidPagination(total int) (startIndex, endIndex int) {

	// out of range
	if p.Page <= 0 || p.PageSize <= 0 || p.Page > int(math.Ceil(float64(total)/float64(p.PageSize))) {
		return 0, 0
	}

	if p.Page == 1 {
		startIndex = 0
	} else {
		startIndex = (p.Page - 1) * p.PageSize
	}
	endIndex = startIndex + p.PageSize

	if endIndex > total {
		endIndex = total
	}

	return startIndex, endIndex
}

func ParseQueryParameter(request *restful.Request) *QueryInfo {
	query := New()

	pageSize, err := strconv.Atoi(request.QueryParameter(ParameterPageSize))
	// equivalent to undefined, use the default value
	if err != nil {
		pageSize = 10
	}
	page, err := strconv.Atoi(request.QueryParameter(ParameterPage))
	// equivalent to undefined, use the default value
	if err != nil {
		page = 1
	}

	query.Pagination = newPagination(page, pageSize)

	query.SortBy = Field(defaultString(request.QueryParameter(ParameterOrderBy), FieldCreationTimeStamp))

	ascending, err := strconv.ParseBool(defaultString(request.QueryParameter(ParameterAscending), "false"))
	if err != nil {
		query.Ascending = false
	} else {
		query.Ascending = ascending
	}

	query.LabelSelector = request.QueryParameter(ParameterLabelSelector)

	for key, values := range request.Request.URL.Query() {
		if !base.HasString([]string{ParameterPage, ParameterPageSize, ParameterOrderBy, ParameterAscending, ParameterLabelSelector}, key) {
			// support multiple query condition
			for _, value := range values {
				query.Filters[Field(key)] = Value(value)
			}
		}
	}

	return query
}

func defaultString(value, defaultValue string) string {
	if len(value) == 0 {
		return defaultValue
	}
	return value
}
