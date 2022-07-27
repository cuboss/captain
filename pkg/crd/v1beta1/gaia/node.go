package gaia

import (
	gv1 "captain/apis/gaia/v1alpha1"
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
)

type GaiaNodeGetter interface {
	GaiaNode(namespace string) GaiaNodeClient
}


//GaiaNodeClient 操作GaiaNode
type GaiaNodeClient interface {
	List(ctx context.Context, opts metav1.ListOptions) (*gv1.GaiaNodeList, error)
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*gv1.GaiaNode, error)
	Create(ctx context.Context, node *gv1.GaiaNode, opts metav1.CreateOptions) (*gv1.GaiaNode, error)
	Update(ctx context.Context, node *gv1.GaiaNode, opts metav1.UpdateOptions) (*gv1.GaiaNode, error)
	UpdateStatus(ctx context.Context, node *gv1.GaiaNode, opts metav1.UpdateOptions) (*gv1.GaiaNode, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, subresources ...string) (*gv1.GaiaNode, error)
}

type NodeClient struct {
	Client *GaiaClient
	Ns     string
}

func (nc *NodeClient) List(ctx context.Context, opts metav1.ListOptions) (*gv1.GaiaNodeList, error) {
	result := gv1.GaiaNodeList{}
	err := nc.Client.restClient.
		Get().
		Namespace(nc.Ns).
		Resource(gv1.GetNodeResName()).
		VersionedParams(&opts, nc.Client.paramCodec).
		Do(ctx).
		Into(&result)

	return &result, err
}

func (nc *NodeClient) Get(ctx context.Context, name string, opts metav1.GetOptions) (*gv1.GaiaNode, error) {
	result := gv1.GaiaNode{}
	err := nc.Client.restClient.
		Get().
		Namespace(nc.Ns).
		Resource(gv1.GetNodeResName()).
		Name(name).
		VersionedParams(&opts, nc.Client.paramCodec).
		Do(ctx).
		Into(&result)

	return &result, err
}

func (nc *NodeClient) Create(ctx context.Context, node *gv1.GaiaNode, opts metav1.CreateOptions) (*gv1.GaiaNode, error) {
	result := gv1.GaiaNode{}
	err := nc.Client.restClient.
		Post().
		Namespace(nc.Ns).
		Resource(gv1.GetNodeResName()).
		Body(node).
		VersionedParams(&opts, nc.Client.paramCodec).
		Do(ctx).
		Into(&result)

	return &result, err
}

func (nc *NodeClient) Update(ctx context.Context, node *gv1.GaiaNode, opts metav1.UpdateOptions) (*gv1.GaiaNode, error) {
	result := gv1.GaiaNode{}
	err := nc.Client.restClient.
		Put().
		Namespace(nc.Ns).
		Resource(gv1.GetNodeResName()).
		Name(node.Name).
		VersionedParams(&opts, nc.Client.paramCodec).
		Body(node).
		Do(ctx).
		Into(&result)

	return &result, err
}

func (nc *NodeClient) UpdateStatus(ctx context.Context, node *gv1.GaiaNode, opts metav1.UpdateOptions) (*gv1.GaiaNode, error) {
	result := &gv1.GaiaNode{}
	err := nc.Client.restClient.Put().
		Namespace(nc.Ns).
		Resource(gv1.GetNodeResName()).
		Name(node.Name).
		SubResource("status").
		VersionedParams(&opts, nc.Client.paramCodec).
		Body(node).
		Do(ctx).
		Into(result)
	return result, err
}

func (nc *NodeClient) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	return nc.Client.restClient.Delete().
		Namespace(nc.Ns).
		Resource(gv1.GetNodeResName()).
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

func (nc *NodeClient) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return nc.Client.restClient.
		Get().
		Namespace(nc.Ns).
		Resource(gv1.GetNodeResName()).
		VersionedParams(&opts, nc.Client.paramCodec).
		Watch(ctx)
}

func (nc *NodeClient) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, subresources ...string) (*gv1.GaiaNode, error) {
	result := &gv1.GaiaNode{}
	err := nc.Client.restClient.Patch(pt).
		Namespace(nc.Ns).
		Resource(gv1.GetNodeResName()).
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do(ctx).
		Into(result)
	return result, err
}
