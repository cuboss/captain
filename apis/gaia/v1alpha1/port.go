package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
)

//PortConf 定义的存储
type PortConf struct {
	Name       string      `json:"name,omitempty"`       //端口绑定的名称
	Protocol   v1.Protocol `json:"protocol,omitempty"`   //协议的类型
	TargetPort string      `json:"targetPort,omitempty"` //对应容器的containerPort
	Port       string      `json:"port,omitempty"`       //宿主机一侧的端口
	NodePort   string      `json:"nodePort,omitempty"`   //通过每个 Node 上的 IP 和静态端口（NodePort）暴露服务
}
