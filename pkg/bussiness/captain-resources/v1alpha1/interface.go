package v1alpha1

import (
	"captain/pkg/unify/query"
	"captain/pkg/unify/response"

	"k8s.io/apimachinery/pkg/runtime"
)

type CaptainResProvider interface {
	// Get retrieves a single object by its namespace and name
	Get(namespace, name string) (runtime.Object, error)

	// List retrieves a collection of objects matches given query
	List(namespace string, query *query.QueryInfo) (*response.ListResult, error)

	// Create object by its namespace and obj
	Create(namespace string, obj runtime.Object) (runtime.Object, error)

	// Delete a object by its namespace and name
	Delete(namespace, name string) error

	// Update a object by its namespace and name
	Update(namespace, name string, obj runtime.Object) (runtime.Object, error)
}
