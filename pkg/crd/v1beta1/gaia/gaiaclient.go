package gaia

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
	"k8s.io/klog"

	gaiav1alpha1 "captain/apis/gaia/v1alpha1"
)

type GaiaCrdSet struct {
	GaiaCluster *ClusterClient
	GaiaNode *NodeClient
	GaiaSet *ClusterSetClient
}

type GaiaClient struct {
	scheme     *runtime.Scheme
	codecs     serializer.CodecFactory
	paramCodec runtime.ParameterCodec
	restClient *rest.RESTClient
	config     *rest.Config
}

type GaiaInterface interface {
	GaiaNodes(namespace string) GaiaNodeClient
	GaiaSets(namespace string) GaiaSetClient
	GaiaClusters(namespace string) GaiaClusterClient
}

func NewGaiaCrd(c *rest.Config) (*GaiaCrdSet, error) {
	cluster, node, set, err := newGaiaClientForConfig(c)
	if err != nil {
		klog.Errorf("Could not initiate gaia client, %s", err)
		return nil, err
	}
	return &GaiaCrdSet{
		GaiaCluster:cluster,
		GaiaNode:node,
		GaiaSet:set,
	}, nil
}

func newGaiaClientForConfig(c *rest.Config) (*ClusterClient, *NodeClient, *ClusterSetClient, error) {
	gaiaClient := &GaiaClient{
		scheme:     gaiav1alpha1.Scheme,
		codecs:     gaiav1alpha1.Codecs,
		paramCodec: runtime.NewParameterCodec(gaiav1alpha1.Scheme),
	}
	config := *c
	config.GroupVersion = &gaiav1alpha1.GroupVersion
	config.APIPath = "/apis"
	config.NegotiatedSerializer = gaiav1alpha1.Codecs.WithoutConversion()
	config.UserAgent = rest.DefaultKubernetesUserAgent()
	client, err := rest.RESTClientFor(&config)
	if err != nil {
		klog.Errorf("Could not initiate gaia client, %s", err)
		return nil, nil, nil, err
	}
	clusterClient := &ClusterClient{
		Client: gaiaClient,
	}
	nodeClient := &NodeClient{
		Client: gaiaClient,
	}

	clusterSetClient := &ClusterSetClient{
		Client: gaiaClient,
	}
	gaiaClient.restClient = client
	gaiaClient.config = &config
	return clusterClient, nodeClient, clusterSetClient, nil
}

// ClusterSets 获取ClusterSet Client
func (c *GaiaCrdSet) GaiaClusters(namespace string) GaiaClusterClient {
	return &ClusterClient{
		Client: c.GaiaCluster.Client,
		Ns:     namespace,
	}
}

// Nodes 获取Node Client
func (c *GaiaCrdSet) GaiaNodes(namespace string) GaiaNodeClient {
	return &NodeClient{
		Client: c.GaiaNode.Client,
		Ns:     namespace,
	}
}

// ClusterSets 获取ClusterSet Client
func (c *GaiaCrdSet) GaiaSets(namespace string) GaiaSetClient {
	return &ClusterSetClient{
		Client: c.GaiaSet.Client,
		Ns:     namespace,
	}
}