package crd_controller

/*
request body:
{
    "name": "testcluster",
    "account_id": "1234567",
    "k8s_ver": "1.20",
    "runtime": {
        "name": "docker",
        "version": "1.19.03"
    },
    "region_id": "123",
    "guest_node_runtime": "kata",
    "vpc": "111111",
    "subnetwork": "2222222",
    "network_plugin": "Calico",
    "pod_cidr": "192.168.0.0/24",
    "service_cidr": "192.168.1.0/24",
    "kubeproxy_model": "ipvs",
    "dns_forward": "100.10.10.1",
    "nodeport_scope": [
        {
            "begin": 30000,
            "end": 32000
        }
    ],
    "nodes": [
        {
            "type": "master",
            "name": "master01",
            "number": 1,
            "disk_uuid_map": {
                "0": {
                    "data_disk_type": "efficiency",
                    "data_disk_size": 10,
                    "volume_uuid": "xxx-yyy"
                },
                "1": {
                    "data_disk_type": "efficiency",
                    "data_disk_size": 10,
                    "volume_uuid": "xxxx-yyyy"
                },
                "2": {
                    "data_disk_type": "efficiency",
                    "data_disk_size": 10,
                    "volume_uuid": "xxxdd-yyyyy"
                }
            },
            "node_ip": "192.168.0.1",
            "system_disk": {
                "disk_capacity": 50,
                "disk_type": "rbd"
            },
            "data_disk": [
                {
                    "disk_capacity": 10,
                    "disk_type": "rbd"
                },
                {
                    "disk_capacity": 10,
                    "disk_type": "rbd"
                },
                {
                    "disk_capacity": 10,
                    "disk_type": "rbd"
                }
            ],
            "res": {
                "resource": {
                    "cpu": 1,
                    "mem": 2
                }
            }
        }
    ],
    "ca_auth_ips": [
        "172.18.1.1",
        "172.18.1.3"
    ],
    "node_ips": [
        "1092.168.0.1"
    ],
    "user_vpc": {
        "vpc_id": "111111111",
        "vpc_name": "vpc_name",
        "network_uuid": "network_uuid",
        "subnetwork_id": "subnetwork_id",
        "subnetwork_name": "subnetwork_name",
        "subnetwork_uuid": "subnetwork_uuid",
        "subnetwork_cidr": "172.16.1.0/24"
    }
}
*/
func newCluster() *CkeClusterVo {
	runtime := RunTime{
		Name: "docker",
		Version: "1.19.03",
	}
	var cluster = &CkeClusterVo{
		Name: "test_cluster",
		AccountID: "account_id",
		K8sVer: "1.20",
		Runtime:  runtime, // 容器运行时
		RegionId: "region_123",
		GuestNodeRuntime: "kata", // 节点运行时
		VPC: "vpc_id",
		Subnetwork: "subnetwork_id",
		NetworkPlugin: "Calico",
		PodCIDR: "192.168.0.0/24",
		ServiceCIDR: "192.168.1.0/24",
		KubeProxyModel: "ipvs",
		DnsForward: "100.10.10.1",
		Nodes: []CkeClusterNodeVo{},
		CaAuthIps: []string{"172.18.1.1", "172.18.1.3"},
	}
	return cluster
}

func newUserVpc() *UserVpc {
	userVpc := &UserVpc{
		VpcId: "111111111",
		VpcName: "vpc_name",
		NetworkUuid: "network_uuid",
		SubnetworkId: "subnetwork_id",
		SubnetworkName: "subnetwork_name",
		SubnetworkUuid: "subnetwork_uuid",
		SubnetworkCidr: "172.16.1.0/24",
	}
	return userVpc
}

// fixme： 此方式需要重新构造aipserver, 不合理，post的方式可以创建，参数如上注释
//func TestCreateCluster(t *testing.T) {
//	cluster := newCluster()
//	userVpc := newUserVpc()
//	ctx := context.Context(nil)
//	ips := []string{}
//	gaiaCluster, err := CreateGaiaCluster(ctx, *cluster, *userVpc, ips)
//	if err != nil {
//		t.Errorf("Error of CreateGaiaCluster: %v", err)
//	}
//	klog.Infof("gaiaCluster: %v", gaiaCluster)
//}

