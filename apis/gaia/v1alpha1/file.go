package v1alpha1

//FileConf 定义的存储
type FileConf struct {
	Path string `json:"path"`           //在容器(service/process)中挂载的路径
	Data string `json:"data"`           //存储如果是文件则记录文件的内容
	Mode uint32 `json:"node,omitempty"` //存储的权限
}
