package v1alpha1

import (
	"flag"
	"os"
	"strings"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

// GroupName is the group name used in this package
const (
	GroupName = "gaia.welkin"
	Version   = "v1alpha1"
)

// GaiaResName() name of resource for gaia crd framework
var (
	gaiaResName      func() string
	gaiaResourceName string
	crdName          string
)

func init() {
	//crd的缺省值必须是空，因为runc wrapper是靠env传crd的，如果有缺省值,wrapper永远读不到env的crd
	flag.StringVar(&crdName, "crd", "", "CRD name")
	gaiaResName = func() string {
		if gaiaResourceName == "" {
			InitScheme()
		}
		return gaiaResourceName
	}
}

func GetGaiaResName() string {
	return gaiaResName()
}

func GetClusterKind() string {
	return gaiaResName() + "Cluster"
}

func GetClusterListKind() string {
	return GetClusterKind() + "List"
}

func GetClusterResName() string {
	return strings.ToLower(GetClusterKind()) + "s"
}

func GetClusterCRDName() string {
	return GetClusterResName() + "." + GroupName
}

func GetSetKind() string {
	return gaiaResName() + "Set"
}

func GetSetListKind() string {
	return GetSetKind() + "List"
}

func GetSetResName() string {
	return strings.ToLower(GetSetKind()) + "s"
}

func GetSetCRDName() string {
	return GetSetResName() + "." + GroupName
}

func GetNodeKind() string {
	return gaiaResName() + "Node"
}

func GetNodeKindList() string {
	return GetNodeKind() + "List"
}

func GetNodeResName() string {
	return strings.ToLower(GetNodeKind()) + "s"
}

func GetNodeCRDName() string {
	return GetNodeResName() + "." + GroupName
}

var SchemeGroupVersion = schema.GroupVersion{Group: GroupName, Version: Version}

func Kind(kind string) schema.GroupKind {
	return SchemeGroupVersion.WithKind(kind).GroupKind()
}

func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

var (
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes) // runtime.SchemeBuilder
	AddToScheme   = SchemeBuilder.AddToScheme
)

func addKnownTypes(scheme *runtime.Scheme) error {

	scheme.AddKnownTypeWithName(schema.GroupVersionKind{
		Group:   GroupName,
		Version: Version,
		Kind:    GetClusterKind(),
	}, &GaiaCluster{})

	scheme.AddKnownTypeWithName(schema.GroupVersionKind{
		Group:   GroupName,
		Version: Version,
		Kind:    GetClusterListKind(),
	}, &GaiaClusterList{})

	scheme.AddKnownTypeWithName(schema.GroupVersionKind{
		Group:   GroupName,
		Version: Version,
		Kind:    GetSetKind(),
	}, &GaiaSet{})

	scheme.AddKnownTypeWithName(schema.GroupVersionKind{
		Group:   GroupName,
		Version: Version,
		Kind:    GetSetListKind(),
	}, &GaiaSetList{})

	scheme.AddKnownTypeWithName(schema.GroupVersionKind{
		Group:   GroupName,
		Version: Version,
		Kind:    GetNodeKind(),
	}, &GaiaNode{})

	scheme.AddKnownTypeWithName(schema.GroupVersionKind{
		Group:   GroupName,
		Version: Version,
		Kind:    GetNodeKindList(),
	}, &GaiaNodeList{})

	// scheme.AddKnownTypes(SchemeGroupVersion,
	// 	&GuestCluster{},
	// 	&GuestClusterList{},
	// )
	meta_v1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: Version}
var Scheme = runtime.NewScheme()
var Codecs = serializer.NewCodecFactory(Scheme)

func InitScheme() *runtime.Scheme {
	initCRDName()
	schemeBuilder := runtime.NewSchemeBuilder(addKnownTypes)
	v1.AddToGroupVersion(Scheme, GroupVersion)
	schemeBuilder.AddToScheme(Scheme)
	return Scheme
}

func initCRDName() {
	var crdstr string
	if len(crdName) <= 0 {
		crdstr = os.Getenv("GAIA_CRD_NAME")
		if len(crdstr) <= 0 {
			// 默认为gaia
			crdstr = "Gaia"
		}
	} else {
		crdstr = crdName
	}
	gaiaResourceName = strFirstToUpper(crdstr)
}

func strFirstToUpper(str string) string {
	str = strings.ToLower(str)
	if len(str) < 1 {
		return ""
	}
	first := strings.ToUpper(str[0:1])

	return first + str[1:]
}
