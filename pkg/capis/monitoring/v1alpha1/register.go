package v1alpha1

import (
	"captain/pkg/server/runtime"

	"github.com/emicklei/go-restful"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	groupName = "monitoring.captain.io"
	respOK    = "ok"
)

var GroupVersion = schema.GroupVersion{Group: groupName, Version: "v1alpha1"}

func AddToContainer(c *restful.Container) error {
	ws := runtime.NewWebService(GroupVersion)

	c.Add(ws)
	return nil
}
