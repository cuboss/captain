package v1alpha1

type Labels struct {
	Key string `json:"key"`          //pvc名称
	Value string `json:"value"`          //存储类型
}
