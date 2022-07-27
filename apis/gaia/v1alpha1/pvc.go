package v1alpha1

type Pvc struct {
	Name string `json:"name"`          //pvc名称
	Kind string `json:"kind"`          //存储类型
}
