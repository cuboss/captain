package v1alpha1

//VolumeConf 定义的目录挂载
type VolumeConf struct {
	Name                 string `json:"name"`
	Type                 string `json:"type"`
	HostPath             string `json:"hostPath"`
	Path                 string `json:"path"`
	Data                 string `json:"data,omitempty"` //存储如果是文件则记录文件的内容
	Mode                 uint32 `json:"mode,omitempty"` //存储的权限
	MountPropagationMode string `json:"mountPropagationMode,omitempty"`
}
