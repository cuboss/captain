package persistentvolumeclaim

import (
	"captain/pkg/bussiness/kube-resources/alpha1"
	"captain/pkg/unify/query"
	"captain/pkg/unify/response"
	snapshotinformers "github.com/kubernetes-csi/external-snapshotter/client/v4/informers/externalversions"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	"strconv"
	"strings"
)

const (
	storageClassName = "storageClassName"

	annotationInUse              = "captain.io/in-use"
	annotationAllowSnapshot      = "captain.io/allow-snapshot"
	annotationStorageProvisioner = "volume.beta.kubernetes.io/storage-provisioner"
)

type persistentvolumeclaimProvider struct {
	sharedInformers   informers.SharedInformerFactory
	snapshotInformers snapshotinformers.SharedInformerFactory
}

func New(informer informers.SharedInformerFactory, snapshotInformer snapshotinformers.SharedInformerFactory) persistentvolumeclaimProvider {
	return persistentvolumeclaimProvider{sharedInformers: informer}
}

func (p persistentvolumeclaimProvider) Get(namespace, name string) (runtime.Object, error) {
	pvc, err := p.sharedInformers.Core().V1().PersistentVolumeClaims().Lister().PersistentVolumeClaims(namespace).Get(name)
	if err != nil {
		return pvc, err
	}
	// we should never mutate the shared objects from informers
	pvc = pvc.DeepCopy()
	p.annotatePVC(pvc)
	return pvc, nil
}

func (p persistentvolumeclaimProvider) List(namespace string, query *query.QueryInfo) (*response.ListResult, error) {
	all, err := p.sharedInformers.Core().V1().PersistentVolumeClaims().Lister().PersistentVolumeClaims(namespace).List(query.GetSelector())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, pvc := range all {
		pvc = pvc.DeepCopy()
		p.annotatePVC(pvc)
		result = append(result, pvc)
	}
	return alpha1.DefaultList(result, query, compareFunc, filter), nil
}

func filter(object runtime.Object, filter query.Filter) bool {
	pvc, ok := object.(*v1.PersistentVolumeClaim)
	if !ok {
		return false
	}

	switch filter.Field {
	case query.FieldStatus:
		return strings.EqualFold(string(pvc.Status.Phase), string(filter.Value))
	case storageClassName:
		return pvc.Spec.StorageClassName != nil && *pvc.Spec.StorageClassName == string(filter.Value)
	default:
		return alpha1.DefaultObjectMetaFilter(pvc.ObjectMeta, filter)
	}
}

func compareFunc(left, right runtime.Object, field query.Field) bool {

	leftClaim, ok := left.(*v1.PersistentVolumeClaim)
	if !ok {
		return false
	}
	rightClaim, ok := right.(*v1.PersistentVolumeClaim)
	if !ok {
		return false
	}
	return alpha1.DefaultObjectMetaCompare(leftClaim.ObjectMeta, rightClaim.ObjectMeta, field)
}

func (p *persistentvolumeclaimProvider) annotatePVC(pvc *v1.PersistentVolumeClaim) {
	inUse := p.countPods(pvc.Name, pvc.Namespace)
	isSnapshotAllow := p.isSnapshotAllowed(pvc.GetAnnotations()[annotationStorageProvisioner])
	if pvc.Annotations == nil {
		pvc.Annotations = make(map[string]string)
	}
	pvc.Annotations[annotationInUse] = strconv.FormatBool(inUse)
	pvc.Annotations[annotationAllowSnapshot] = strconv.FormatBool(isSnapshotAllow)
}

func (p *persistentvolumeclaimProvider) countPods(name, namespace string) bool {
	pods, err := p.sharedInformers.Core().V1().Pods().Lister().Pods(namespace).List(labels.Everything())
	if err != nil {
		return false
	}
	for _, pod := range pods {
		for _, pvc := range pod.Spec.Volumes {
			if pvc.PersistentVolumeClaim != nil && pvc.PersistentVolumeClaim.ClaimName == name {
				return true
			}
		}
	}
	return false
}

func (p *persistentvolumeclaimProvider) isSnapshotAllowed(provisioner string) bool {
	if len(provisioner) == 0 {
		return false
	}
	volumeSnapshotClasses, err := p.snapshotInformers.Snapshot().V1().VolumeSnapshotClasses().Lister().List(labels.Everything())
	if err != nil {
		return false
	}
	for _, volumeSnapshotClass := range volumeSnapshotClasses {
		if volumeSnapshotClass.Driver == provisioner {
			return true
		}
	}
	return false
}
