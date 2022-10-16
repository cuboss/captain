package persistentvolumeclaim

import (
	"context"
	"strconv"

	volumesnapshotv1 "github.com/kubernetes-csi/external-snapshotter/client/v4/apis/volumesnapshot/v1"
	versioned "github.com/kubernetes-csi/external-snapshotter/client/v4/clientset/versioned"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"

	"captain/pkg/bussiness/kube-resources/alpha1"
	"captain/pkg/unify/query"
	"captain/pkg/unify/response"
	"captain/pkg/utils/clusterclient"
)

type mcPersistentVolumeClaimProvider struct {
	clusterclient.ClusterClients
}

func NewMCResProvider(clients clusterclient.ClusterClients) mcPersistentVolumeClaimProvider {
	return mcPersistentVolumeClaimProvider{ClusterClients: clients}
}

func (pd mcPersistentVolumeClaimProvider) Get(region, cluster, namespace, name string) (runtime.Object, error) {
	cli, err := pd.GetClientSet(region, cluster)
	if err != nil {
		return nil, err
	}

	pvc, err := cli.CoreV1().PersistentVolumeClaims(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// we should never mutate the shared objects from informers
	pvc = pvc.DeepCopy()

	helper := &pvcHelper{Clientset: cli}
	helper.annotatePVC(pvc)

	return pvc, nil
}

func (pd mcPersistentVolumeClaimProvider) List(region, cluster, namespace string, query *query.QueryInfo) (*response.ListResult, error) {
	cli, err := pd.GetClientSet(region, cluster)
	if err != nil {
		return nil, err
	}
	list, err := cli.CoreV1().PersistentVolumeClaims(namespace).List(context.Background(), metav1.ListOptions{LabelSelector: query.LabelSelector})
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	if list != nil && list.Items != nil {
		for i := 0; i < len(list.Items); i++ {
			pvc := &list.Items[i]
			pvc = pvc.DeepCopy()
			helper := &pvcHelper{Clientset: cli}
			helper.annotatePVC(pvc)
			result = append(result, &list.Items[i])
		}
	}

	return alpha1.DefaultList(result, query, compareFunc, filter), nil
}

type pvcHelper struct {
	*kubernetes.Clientset
	pods      *v1.PodList
	snapshots *volumesnapshotv1.VolumeSnapshotClassList
}

func (h *pvcHelper) annotatePVC(pvc *v1.PersistentVolumeClaim) {
	inUse := h.countPods(pvc.Name, pvc.Namespace)
	isSnapshotAllow := h.isSnapshotAllowed(pvc.GetAnnotations()[annotationStorageProvisioner])
	if pvc.Annotations == nil {
		pvc.Annotations = make(map[string]string)
	}
	pvc.Annotations[annotationInUse] = strconv.FormatBool(inUse)
	pvc.Annotations[annotationAllowSnapshot] = strconv.FormatBool(isSnapshotAllow)
}

func (h *pvcHelper) countPods(name, namespace string) bool {
	var pods *v1.PodList
	if h.pods != nil {
		pods = h.pods
	} else {
		list, err := h.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return false
		}
		h.pods = list
	}

	if pods != nil && pods.Items != nil {
		for i := 0; i < len(pods.Items); i++ {
			pod := pods.Items[i]
			for _, pvc := range pod.Spec.Volumes {
				if pvc.PersistentVolumeClaim != nil && pvc.PersistentVolumeClaim.ClaimName == name {
					return true
				}
			}
		}

	}
	return false
}

func (h *pvcHelper) isSnapshotAllowed(provisioner string) bool {
	if len(provisioner) == 0 {
		return false
	}

	var snapshots *volumesnapshotv1.VolumeSnapshotClassList
	if h.snapshots != nil {
		snapshots = h.snapshots
	} else {
		ssCli := versioned.New(h.RESTClient())
		snapshots, err := ssCli.SnapshotV1().VolumeSnapshotClasses().List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return false
		}
		h.snapshots = snapshots
	}

	if snapshots != nil && snapshots.Items != nil {
		for i := 0; i < len(snapshots.Items); i++ {
			volumeSnapshotClass := snapshots.Items[i]
			if volumeSnapshotClass.Driver == provisioner {
				return true
			}
		}
	}
	return false
}
