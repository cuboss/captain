package openapi

import (
	"github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	"github.com/go-openapi/spec"
)

func AddToContainer(container *restful.Container) error {
	config := restfulspec.Config{
		WebServices:                   container.RegisteredWebServices(),
		APIPath:                       "/open-api",
		PostBuildSwaggerObjectHandler: enrichSwaggerObject,
	}

	container.Add(restfulspec.NewOpenAPIService(config))
	return nil

}

func enrichSwaggerObject(swo *spec.Swagger) {
	swo.Info = &spec.Info{
		InfoProps: spec.InfoProps{
			Title: "Captain Server",
		},
	}
}
