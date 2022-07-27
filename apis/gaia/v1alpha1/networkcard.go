package v1alpha1

//NetworkCardConf 定义网卡
type NetworkCardConf struct {
	Network string `json:"network"`          //网络名称
	NodeIP  string `json:"nodeIP,omitempty"` //IP
}
