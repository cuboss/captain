package capis

import "github.com/emicklei/go-restful"

var RegionScopeService = &restful.WebService{}

func init() {

	RegionScopeService = RegionScopeService.Path("/regions/{region}/clusters/{cluster}/capis/").
		Param(RegionScopeService.PathParameter("region", "region id of cluster")).
		Param(RegionScopeService.PathParameter("cluster", "name of cluster")).
		Produces(restful.MIME_JSON)

}
