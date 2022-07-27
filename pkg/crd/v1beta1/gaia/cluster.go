package gaia

import (
	"context"

	gv1 "captain/apis/gaia/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

//参考 https://github.com/kubernetes/client-go/blob/master/kubernetes/typed/core/v1/pod.go

type GaiaClusterGetter interface {
	GaiaCluster(namespace string) GaiaClusterClient
}

type GaiaClusterClient interface {
	List(ctx context.Context, opts metav1.ListOptions) (*gv1.GaiaClusterList, error)
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*gv1.GaiaCluster, error)
	Create(ctx context.Context, cluster *gv1.GaiaCluster, opts metav1.CreateOptions) (*gv1.GaiaCluster, error)
	Update(ctx context.Context, cluster *gv1.GaiaCluster, opts metav1.UpdateOptions) (*gv1.GaiaCluster, error)
	UpdateStatus(ctx context.Context, cluster *gv1.GaiaCluster, opts metav1.UpdateOptions) (*gv1.GaiaCluster, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
}

type ClusterClient struct {
	Client *GaiaClient
	Ns     string
}

func (cc *ClusterClient) List(ctx context.Context, opts metav1.ListOptions) (*gv1.GaiaClusterList, error) {
	result := gv1.GaiaClusterList{}
	err := cc.Client.restClient.
		Get().
		Namespace(cc.Ns).
		Resource(gv1.GetClusterResName()).
		VersionedParams(&opts, cc.Client.paramCodec).
		Do(ctx).
		Into(&result)

	return &result, err
}

func (cc *ClusterClient) Get(ctx context.Context, name string, opts metav1.GetOptions) (*gv1.GaiaCluster, error) {
	result := gv1.GaiaCluster{}
	err := cc.Client.restClient.
		Get().
		Namespace(cc.Ns).
		Resource(gv1.GetClusterResName()).
		Name(name).
		VersionedParams(&opts, cc.Client.paramCodec).
		Do(ctx).
		Into(&result)

	return &result, err
}

func (cc *ClusterClient) Create(ctx context.Context, cluster *gv1.GaiaCluster, opts metav1.CreateOptions) (*gv1.GaiaCluster, error) {
	result := gv1.GaiaCluster{}
	err := cc.Client.restClient.
		Post().
		Namespace(cc.Ns).
		Resource(gv1.GetClusterResName()).
		Body(cluster).
		VersionedParams(&opts, cc.Client.paramCodec).
		Do(ctx).
		Into(&result)

	return &result, err
}

func (cc *ClusterClient) Update(ctx context.Context, cluster *gv1.GaiaCluster, opts metav1.UpdateOptions) (*gv1.GaiaCluster, error) {
	result := gv1.GaiaCluster{}
	err := cc.Client.restClient.
		Put().
		Namespace(cc.Ns).
		Resource(gv1.GetClusterResName()).
		Name(cluster.Name).
		VersionedParams(&opts, cc.Client.paramCodec).
		Body(cluster).
		Do(ctx).
		Into(&result)

	return &result, err
}

func (cc *ClusterClient) UpdateStatus(ctx context.Context, cluster *gv1.GaiaCluster, opts metav1.UpdateOptions) (*gv1.GaiaCluster, error) {
	result := &gv1.GaiaCluster{}
	err := cc.Client.restClient.Put().
		Namespace(cc.Ns).
		Resource(gv1.GetClusterResName()).
		Name(cluster.Name).
		SubResource("status").
		VersionedParams(&opts, cc.Client.paramCodec).
		Body(cluster).
		Do(ctx).
		Into(result)
	return result, err
}

func (cc *ClusterClient) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	return cc.Client.restClient.Delete().
		Namespace(cc.Ns).
		Resource(gv1.GetClusterResName()).
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

func (cc *ClusterClient) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return cc.Client.restClient.
		Get().
		Namespace(cc.Ns).
		Resource(gv1.GetClusterResName()).
		VersionedParams(&opts, cc.Client.paramCodec).
		Watch(ctx)
}
