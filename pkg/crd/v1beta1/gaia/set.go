package gaia

import (
	gv1 "captain/apis/gaia/v1alpha1"
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)


type GaiaSetGetter interface {
	GaiaSet(namespace string) GaiaSetClient
}

//GaiaSetClient 操作GaiaSet
type GaiaSetClient interface {
	List(ctx context.Context, opts metav1.ListOptions) (*gv1.GaiaSetList, error)
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*gv1.GaiaSet, error)
	Create(ctx context.Context, clusterSet *gv1.GaiaSet, opts metav1.CreateOptions) (*gv1.GaiaSet, error)
	Update(ctx context.Context, clusterSet *gv1.GaiaSet, opts metav1.UpdateOptions) (*gv1.GaiaSet, error)
	UpdateStatus(ctx context.Context, clusterSet *gv1.GaiaSet, opts metav1.UpdateOptions) (*gv1.GaiaSet, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
}

type ClusterSetClient struct {
	Client *GaiaClient
	Ns     string
}

func (cc *ClusterSetClient) List(ctx context.Context, opts metav1.ListOptions) (*gv1.GaiaSetList, error) {
	result := gv1.GaiaSetList{}
	err := cc.Client.restClient.
		Get().
		Namespace(cc.Ns).
		Resource(gv1.GetSetResName()).
		VersionedParams(&opts, cc.Client.paramCodec).
		Do(ctx).
		Into(&result)

	return &result, err
}

func (cc *ClusterSetClient) Get(ctx context.Context, name string, opts metav1.GetOptions) (*gv1.GaiaSet, error) {
	result := gv1.GaiaSet{}
	err := cc.Client.restClient.
		Get().
		Namespace(cc.Ns).
		Resource(gv1.GetSetResName()).
		Name(name).
		VersionedParams(&opts, cc.Client.paramCodec).
		Do(ctx).
		Into(&result)

	return &result, err
}

func (cc *ClusterSetClient) Create(ctx context.Context, clusterSet *gv1.GaiaSet, opts metav1.CreateOptions) (*gv1.GaiaSet, error) {
	result := gv1.GaiaSet{}
	err := cc.Client.restClient.
		Post().
		Namespace(cc.Ns).
		Resource(gv1.GetSetResName()).
		Body(clusterSet).
		VersionedParams(&opts, cc.Client.paramCodec).
		Do(ctx).
		Into(&result)

	return &result, err
}

func (cc *ClusterSetClient) Update(ctx context.Context, clusterSet *gv1.GaiaSet, opts metav1.UpdateOptions) (*gv1.GaiaSet, error) {
	result := gv1.GaiaSet{}
	err := cc.Client.restClient.
		Put().
		Namespace(cc.Ns).
		Resource(gv1.GetSetResName()).
		Name(clusterSet.Name).
		VersionedParams(&opts, cc.Client.paramCodec).
		Body(clusterSet).
		Do(ctx).
		Into(&result)

	return &result, err
}

func (cc *ClusterSetClient) UpdateStatus(ctx context.Context, clusterSet *gv1.GaiaSet, opts metav1.UpdateOptions) (*gv1.GaiaSet, error) {
	result := &gv1.GaiaSet{}
	err := cc.Client.restClient.Put().
		Namespace(cc.Ns).
		Resource(gv1.GetSetResName()).
		Name(clusterSet.Name).
		SubResource("status").
		VersionedParams(&opts, cc.Client.paramCodec).
		Body(clusterSet).
		Do(ctx).
		Into(result)
	return result, err
}

func (cc *ClusterSetClient) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	return cc.Client.restClient.Delete().
		Namespace(cc.Ns).
		Resource(gv1.GetSetResName()).
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

func (cc *ClusterSetClient) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return cc.Client.restClient.
		Get().
		Namespace(cc.Ns).
		Resource(gv1.GetSetResName()).
		VersionedParams(&opts, cc.Client.paramCodec).
		Watch(ctx)
}
