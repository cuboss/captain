package alpha1

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/server/options"
)

type Interface interface {
	Get(namespace, name string) runtime.Object
	List(namespace string, ops options.CoreAPIOptions) []runtime.Object
}
