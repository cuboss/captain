package crd_controller

import (
	"captain/apis/gaia/v1alpha1"
	"captain/pkg/simple/client/k8s"
	"context"
	"encoding/json"
	"fmt"
	apiv1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/klog"
	"strconv"
	"strings"
	"time"
)

// Kind ... type of k8s resource kind(string)
type Kind string

// ResourceKind ... the plural form of Kind(string)
type ResourceKind string

const (
	// KindGaiaCluster ... the Kind of GaiaCluster
	KindGaiaCluster Kind = "GaiaCluster"
	// ResourceGaiaCluster ... the ResourceKind of GaiaCluster
	ResourceGaiaCluster ResourceKind = "gaiaclusters"
	// KindGaiaNode ... the Kind of GaiaNode
	KindGaiaNode Kind = "GaiaNode"
	// ResourceGaiaNode ... the ResourceKind of GaiaNode
	ResourceGaiaNode ResourceKind = "gaianodes"

	AgileNodeNameChar = "abcdefghijklmnopqrstuvwxyz0123456789"
	AgliNodeNameSize = 7
	DataDiskTypeEfficiency = "efficiency"
	DataDiskTypeSSD = "ssd"
)

var KubernetesClient k8s.Client

func GetKubernetesClient(client k8s.Client) k8s.Client {
	KubernetesClient = client
	return KubernetesClient
}

type RunTime struct {
	Name    string `json:"name"`
	Version string                    `json:"version"`
}

type NodeportScope struct {
	Begin int64 `json:"begin"`
	End   int64 `json:"end"`
}

type CkeClusterVo struct {
	Name             string                    `json:"name"`
	AccountID        string                    `json:"account_id"`
	K8sVer           string                    `json:"k8s_ver"` // 1.18 1.17
	Runtime          RunTime                   `json:"runtime"` // 容器运行时
	RegionId         string                    `json:"region_id"`
	GuestNodeRuntime v1alpha1.GuestNodeRuntime `json:"guest_node_runtime"` // 节点运行时
	VPC              string                    `json:"vpc_id"`
	Subnetwork       string                    `json:"subnetwork_id"`
	NetworkPlugin    string                    `json:"network_plugin"`
	PodCIDR          string                    `json:"pod_cidr"`
	ServiceCIDR      string                    `json:"service_cidr"`
	KubeProxyModel   string                    `json:"kubeproxy_model"`
	DnsForward       string                    `json:"dns_forward"`
	Nodes            []CkeClusterNodeVo        `json:"nodes"`
	NodeportScope    []NodeportScope           `json:"nodeport_scope"`
	CaAuthIps        []string                  `json:"ca_auth_ips"`
	UserVpc          UserVpc                   `json:"user_vpc"`
	NodeIps          []string                  `json:"node_ips"`
}

type UserVpc struct {
	VpcId          string `json:"vpc_id"`
	VpcName        string `json:"vpc_name"`
	NetworkUuid    string `json:"network_uuid"`
	SubnetworkId   string `json:"subnetwork_id"`
	SubnetworkName string `json:"subnetwork_name"`
	SubnetworkUuid string `json:"subnetwork_uuid"`
	SubnetworkCidr string `json:"subnetwork_cidr"`
}

type Disk struct {
	DiskType     string `json:"disk_type"` // rbd - efficiency | rbd-ssd - ssd
	DiskCapacity int64  `json:"disk_capacity"`  // eg. 50G
	DiskId       string `json:"disk_id"`  // gaia不涉及
}

//数据盘
type DataDisk struct {
	DataDiskType string `json:"data_disk_type"`//数据盘相关
	DataDiskSize int    `json:"data_disk_size"`
	VolumeUUID   string `json:"volume_uuid"`
}

type CkeNodeResource struct {
	CPU float64 `json:"cpu"`
	Mem float64 `json:"mem"`
}

type CkeClusterNodeResource struct {
	TypeName   string          `json:"type_name"`   // gaia不涉及
	TypeId     string          `json:"type_id"`     // gaia不涉及
	ResourceId string          `json:"resource_id"` // gaia不涉及
	Resource   CkeNodeResource `json:"resource"`
}

type Process struct {
	Cmd       string              `json:"cmd"`
	Name      string              `json:"name"`
	Envs      []string            `json:"envs"`
	Args      []string            `json:"args"`
	Expect    string              `json:"expect"`
	LogPath   string              `json:"logPath"`
	StartTime time.Time           `json:"start_time"`
	Status    v1alpha1.SvcState   `json:"status"`
}

type CkeClusterNodeVo struct {
	Type          string                 `json:"type"`
	Name          string                 `json:"name"`
	Number        int                    `json:"number"`
	NodeIP        string                 `json:"node_ip"`
	SystemDisk    Disk                   `json:"system_disk"`
	DataDisk      []Disk                 `json:"data_disk"`
	Resource      CkeClusterNodeResource `json:"res"`
	Processes     []Process              `json:"process"`
	DiskUuidMap   map[int]*DataDisk      `json:"disk_uuid_map"`
}

type WooshnetPorts struct {
	NetworkId string    `json:"network_id"`
	FixedIps  []FixedIp `json:"fixed_ips"`
}

type FixedIp struct {
	SubnetId  string `json:"subnet_id"`
	IpAddress string `json:"ip_address"`
}

//BuildGaiaCluster ...
func BuildGaiaCluster(cluster CkeClusterVo, accountID string) *v1alpha1.GaiaCluster {
	var gaiaCluster v1alpha1.GaiaCluster
	gaiaCluster.APIVersion = "gaia.welkin/v1alpha1"
	gaiaCluster.Kind = string(KindGaiaCluster)
	gaiaCluster.Name = cluster.Name
	gaiaCluster.Spec = buildGaiaClusterSpec(cluster)
	gaiaCluster.Labels = make(map[string]string)
	gaiaCluster.Labels["productType"] = "CKE"
	gaiaCluster.Labels["accountID"] = accountID
	gaiaCluster.Labels["clusterName"] = cluster.Name
	klog.Infof("BuildGaiaCluster: %v", gaiaCluster)
	return &gaiaCluster
}

func buildGaiaClusterSpec(cluster CkeClusterVo) v1alpha1.GaiaClusterSpec {
	var gaiaClusterSpec v1alpha1.GaiaClusterSpec
	userVpc := cluster.UserVpc
	gaiaClusterSpec.Runtime = cluster.GuestNodeRuntime
	// 1.18-kata-docker ｜ 1.18-kata-containerd
	gaiaClusterSpec.Template = fmt.Sprintf("%s-%s-%s", cluster.K8sVer, "kata", strings.ToLower(string(cluster.Runtime.Name)))
	gaiaClusterSpec.VPC = userVpc.VpcName
	gaiaClusterSpec.ClusterIPRange = cluster.ServiceCIDR
	gaiaClusterSpec.ClusterPodCidr = cluster.PodCIDR
	gaiaClusterSpec.CaAuthIps = cluster.CaAuthIps
	gaiaClusterSpec.DnsForward = cluster.DnsForward

	// 固定 DeploymentFeature
	deploymentFeature := v1alpha1.DeployFeature{
		DifferentHost:     false,
		PersistentStorage: true,
	}
	gaiaClusterSpec.DeploymentFeature = deploymentFeature

	return gaiaClusterSpec
}

func ConvertResourceCpuFloatToK8s(value float64) resource.Quantity {
	resourceStr := fmt.Sprintf("%d%s", int(value*1000), "m")
	resourceQuantity, _ := resource.ParseQuantity(resourceStr)
	return resourceQuantity
}

func ConvertResourceMemFloatToK8s(value float64) resource.Quantity {
	resourceStr := strconv.Itoa(int(value)) + "Gi"
	resourceQuantity, _ := resource.ParseQuantity(resourceStr)
	return resourceQuantity

}

func BuildNode(node CkeClusterNodeVo, vpc string, subnetwork string, ip string, name string, runtime v1alpha1.GuestNodeRuntime, pvcs []v1alpha1.Pvc) v1alpha1.GaiaNodeSpec {
	gaiaNode := v1alpha1.GaiaNodeSpec{}
	gaiaNode.Name = name
	gaiaNode.Type = node.Type
	gaiaNode.Runtime = runtime
	// gaiaNode.ExpectHost = node.
	// 构建虚拟IP [根据ID 获取 VPC信息以及 可用IP]
	fixedIp := &FixedIp{
		SubnetId:  subnetwork, // 替换成真实的子网ID
		IpAddress: ip,
	}

	var fixedIps []FixedIp

	fixedIps = append(fixedIps, *fixedIp)

	wooshnetPort := &WooshnetPorts{
		NetworkId: vpc,
		FixedIps:  fixedIps,
	}

	var wooshnetPorts []WooshnetPorts

	wooshnetPorts = append(wooshnetPorts, *wooshnetPort)

	gaiaNode.Annotations = make(map[string]string)

	gaiaNode.Annotations["v1.multus-cni.io/default-network"] = "wooshnet"

	networkJson, _ := json.Marshal(wooshnetPorts)
	networkStr := string(networkJson)

	gaiaNode.Annotations["wooshnet/ports"] = networkStr

	// build networkcard
	networkCard := make([]v1alpha1.NetworkCardConf, 1)

	networkCard[0] = v1alpha1.NetworkCardConf{
		Network: vpc,
		NodeIP:  ip,
	}
	gaiaNode.NetworkCards = networkCard

	res := apiv1.ResourceList{}
	res[apiv1.ResourceCPU] = ConvertResourceCpuFloatToK8s(node.Resource.Resource.CPU)
	res[apiv1.ResourceMemory] = ConvertResourceMemFloatToK8s(node.Resource.Resource.Mem)
	gaiaNode.Resource.Requests = res
	gaiaNode.Resource.Limits = res

	if len(pvcs) > 0 {
		gaiaNode.Pvc = append(gaiaNode.Pvc, pvcs...)
	}

	labels := v1alpha1.Labels{
		Key: "demo",
		Value: "demo",
	}
	gaiaNode.Labels = append(gaiaNode.Labels, labels)
	toleration := apiv1.Toleration{
		Key: "cucloud.cn/infra.k8s",
		Operator: "Equal",
		Effect: "NoSchedule",
	}
	gaiaNode.Tolerations = append(gaiaNode.Tolerations, toleration)

	gaiaNode.Services = make(map[string]v1alpha1.Service)
	for _, process := range node.Processes {
		gaiaNode.Services[process.Name] = v1alpha1.Service{
			Srv: v1alpha1.Process{
				Args: process.Args,
				Envs: process.Envs,
			},
		}
	}
	return gaiaNode
}

func buildDiskData(disk Disk) *DataDisk {
	var volumeName string
	if disk.DiskType == "rbd" {
		volumeName = DataDiskTypeEfficiency
	} else {
		volumeName = DataDiskTypeSSD
	}
	dcsDisk := &DataDisk{
		DataDiskType: volumeName, //数据盘相关
		DataDiskSize: int(disk.DiskCapacity),
	}
	return dcsDisk
}

// RandAgleNodeName 敏捷版集群 node 随机命名
func RandAgleNodeName(size int) string {
	name := []byte(AgileNodeNameChar)
	var result []byte
	rand.Seed(time.Now().UnixNano()+ int64(rand.Intn(100)))
	for i := 0; i < size; i++ {
		result = append(result, name[rand.Intn(len(name))])
	}
	return string(result)
}

func convertDiskToDataDisk(disk Disk) DataDisk {
	var dataDisk DataDisk
	dataDisk.DataDiskSize = int(disk.DiskCapacity)
	if disk.DiskType == "rbd" {
		dataDisk.DataDiskType = DataDiskTypeEfficiency
	} else {
		dataDisk.DataDiskType = DataDiskTypeSSD
	}
	return dataDisk
}

func ConvertResourceStorageToK8s(value float64) resource.Quantity {
	resourceStr := strconv.Itoa(int(value)) + "Gi"
	resourceQuantity, _ := resource.ParseQuantity(resourceStr)
	return resourceQuantity
}

func CreatePv(disk DataDisk, name, volumeMode string, kubernetesClient k8s.Client, c context.Context) (*apiv1.PersistentVolume, error) {
	volumeMounts := apiv1.PersistentVolume{}
	volumeMounts.APIVersion = "v1"
	volumeMounts.Kind = "PersistentVolume"
	var mode apiv1.PersistentVolumeMode
	mode = v1.PersistentVolumeMode(volumeMode)
	volumeMounts.Spec.VolumeMode = &mode
	volumeMounts.Spec.AccessModes = append(volumeMounts.Spec.AccessModes, apiv1.ReadWriteOnce)
	volumeMounts.Spec.StorageClassName = "chinaunicom.cinder.sc"
	res := apiv1.ResourceList{}
	res[apiv1.ResourceStorage] = ConvertResourceStorageToK8s(float64(disk.DataDiskSize))
	volumeMounts.Spec.Capacity = res
	var csi apiv1.CSIPersistentVolumeSource
	csi.Driver = "chinaunicom.cinder.csi"
	fmt.Printf("csi.VolumeHandle: %s", disk.VolumeUUID)
	csi.VolumeHandle = disk.VolumeUUID
	volumeMounts.Spec.CSI = &csi
	volumeMounts.Name = name
	pv, err := kubernetesClient.Kubernetes().CoreV1().PersistentVolumes().Create(c, &volumeMounts, metav1.CreateOptions{})
	if err != nil {
		klog.Errorf("guestClusterService | CreatePv : Create PersistentVolume failed %s", err.Error())
	}
	return pv, err
}

func CreatePvc(disk DataDisk, name, volumeMode string, pvNameSpace string, kubernetesClient k8s.Client, c context.Context) (*apiv1.PersistentVolumeClaim, error) {
	volumeClaim := apiv1.PersistentVolumeClaim{}
	volumeClaim.APIVersion = "v1"
	volumeClaim.Kind = "PersistentVolumeClaim"
	var mode apiv1.PersistentVolumeMode
	mode = v1.PersistentVolumeMode(volumeMode)
	volumeClaim.Spec.VolumeMode = &mode
	volumeClaim.Spec.AccessModes = append(volumeClaim.Spec.AccessModes, apiv1.ReadWriteOnce)
	var storageClassName string
	storageClassName = "chinaunicom.cinder.sc"
	volumeClaim.Spec.StorageClassName = &storageClassName
	res := apiv1.ResourceList{}
	res[apiv1.ResourceStorage] = ConvertResourceStorageToK8s(float64(disk.DataDiskSize))
	requestResource := apiv1.ResourceRequirements{}
	requestResource.Limits = res
	requestResource.Requests = res
	volumeClaim.Spec.Resources = requestResource
	klog.Infof("volume name : %s", name)
	volumeClaim.Spec.VolumeName = name
	volumeClaim.Name = name
	klog.Infof("volumeClaim : %v", volumeClaim)
	pvc, err := kubernetesClient.Kubernetes().CoreV1().PersistentVolumeClaims(pvNameSpace).Create(c, &volumeClaim, metav1.CreateOptions{})
	if err != nil {
		klog.Infof("create pvc failed : err %s", err.Error())
	}
	return pvc, err
}

func createPvPvc(ctx context.Context, kubernetesClient k8s.Client, dataDisks []Disk, dbDisk []*DataDisk, cluster CkeClusterVo, nodeName string, nodeType string, diskUuidMap map[int]*DataDisk) ([]*DataDisk, []v1alpha1.Pvc, error) {
	volumeIndex := 0
	var pvcs []v1alpha1.Pvc
	for dcsDiskIndex, dataDisk := range dataDisks {
		volumeIndex++
		disk := convertDiskToDataDisk(dataDisk)
		dcsDisk := diskUuidMap[dcsDiskIndex]
		var volumeMode string
		if nodeType == "master" {
			if volumeIndex < len(dataDisks) {
				volumeMode = "Block"
			} else {
				volumeMode = "Filesystem"
			}
		} else {
			volumeMode = "Block"
		}

		dbDisk = append(dbDisk, dcsDisk)
		disk.VolumeUUID = dcsDisk.VolumeUUID
		pvName := fmt.Sprintf("%s-%s-%d", cluster.Name, nodeName, volumeIndex)
		pv, err := CreatePv(disk, pvName, volumeMode, kubernetesClient, ctx)
		if err != nil {
			klog.Errorf("pv create failed :%s %v", err.Error(), pv)
		}
		pvNameSpace := cluster.AccountID
		pvc, err := CreatePvc(disk, pvName, volumeMode, pvNameSpace, kubernetesClient, ctx)
		if err != nil {
			klog.Errorf("pvc create failed :%s %v", err.Error(), pvc)
		}
		gaiaPvc := v1alpha1.Pvc{
			Name: pvName,
			Kind: strings.ToLower(volumeMode),
		}
		pvcs = append(pvcs, gaiaPvc)
	}
	return dbDisk, pvcs, nil
}

// CreateGaiaCluster ... create gaia cluster
func CreateGaiaCluster(ctx context.Context, cluster CkeClusterVo) (*v1alpha1.GaiaCluster, error) {
	ips := cluster.NodeIps
	gaiaCluster := &v1alpha1.GaiaCluster{}
	userVpc := cluster.UserVpc
	var gaiaNodeSpecs []v1alpha1.GaiaNodeSpec
	namespace := cluster.AccountID
	newNamespace := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	_, err := KubernetesClient.Kubernetes().CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		klog.Infof("gaia namespace: %s not exists, need to create", namespace)
		_, errCreate := KubernetesClient.Kubernetes().CoreV1().Namespaces().Create(ctx, newNamespace, metav1.CreateOptions{})
		if errCreate != nil {
			klog.Errorf("Create gaia namespace Error: %s", errCreate.Error())
			return gaiaCluster, errCreate
		}
	} else if err != nil {
		klog.Errorf("Get gaia namespace Error: %s", err.Error())
		return gaiaCluster, err
	}
	var index = 0
	for _, clusterNode := range cluster.Nodes {
		// 获取硬盘信息: UUID 规格
		// build node yaml 提前构建 append yaml
		// 创建NODE数据
		// 插入所有NODE数据
		nodeVolumesMap := clusterNode.DiskUuidMap
		for i := 0; i < clusterNode.Number; i++ {
			index++
			nodeName := fmt.Sprintf("%s-%s", clusterNode.Type, RandAgleNodeName(AgliNodeNameSize))
			var disks []*DataDisk
			dataDisks := clusterNode.DataDisk
			systemDisk := buildDiskData(clusterNode.SystemDisk)
			disks = append(disks, systemDisk)
			disks, pvcs, err := createPvPvc(ctx, KubernetesClient, dataDisks, disks, cluster, nodeName, clusterNode.Type, nodeVolumesMap)
			if err != nil {
				klog.Errorf("createPvPvc failed: %s", err.Error())
				return gaiaCluster, err
			}

			nodeSpec := BuildNode(clusterNode, userVpc.NetworkUuid, userVpc.SubnetworkUuid, ips[index-1], nodeName, cluster.GuestNodeRuntime, pvcs)
			gaiaNodeSpecs = append(gaiaNodeSpecs, nodeSpec)
		}
	}

	gaiaCluster = BuildGaiaCluster(cluster, cluster.AccountID)
	gaiaCluster.Spec.Nodes = gaiaNodeSpecs
	// todo： 创建 apiserver 和 dashboard 的svc
	gaiaCluster, err = KubernetesClient.Crd().V1beta1().GaiaCluster(namespace).Create(ctx, gaiaCluster, metav1.CreateOptions{})
	if err != nil {
		klog.Errorf("create gaia cluster err: %s", err.Error())
		return nil, err
	}
	return gaiaCluster, nil
}
