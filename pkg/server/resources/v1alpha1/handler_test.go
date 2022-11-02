package v1alpha1

import (
	"fmt"
	"testing"

	"captain/apis/cluster/v1alpha1"
	"captain/pkg/bussiness/captain-resources/v1alpha1/resource"
	"captain/pkg/crd"
	"captain/pkg/informers"
	"captain/pkg/unify/query"
	"captain/pkg/unify/response"

	"captain/pkg/client/clientset/versioned/fake"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	clusters = []interface{}{
		&v1alpha1.Cluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "cluster-demo1",
				Namespace: "test1",
			},
			Spec: v1alpha1.ClusterSpec{
				JoinFederation: false,
				Enable:         true,
				Provider:       "CKE",
				Connection: v1alpha1.Connection{
					Type:       v1alpha1.ConnectionTypeDirect,
					KubeConfig: []byte("test kubeconfig context"),
				},
			},
		},

		&v1alpha1.Cluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "cluster-demo2",
				Namespace: "test2",
			},
			Spec: v1alpha1.ClusterSpec{
				JoinFederation: false,
				Enable:         true,
				Provider:       "CKE",
				Connection: v1alpha1.Connection{
					Type:       v1alpha1.ConnectionTypeDirect,
					KubeConfig: []byte("test kubeconfig context"),
				},
			},
		},
	}
)

func prepare() (informers.CapInformerFactory, crd.CrdInterface, error) {
	cli := fake.NewSimpleClientset()
	crdInterface := crd.New(nil, cli)
	informerFac := informers.NewInformerFactories(nil, crdInterface)
	captainInformer := informerFac.CaptainSharedInformerFactory()

	for _, cluster := range clusters {
		err := captainInformer.Cluster().V1alpha1().Clusters().Informer().GetIndexer().Add(cluster)
		if err != nil {
			fmt.Printf("adding msg: %s", err.Error())
			return nil, nil, err
		}
	}

	return informerFac, crdInterface, nil
}

func TestHandleListResources(t *testing.T) {
	tests := []struct {
		description   string
		resource      string
		query         *query.QueryInfo
		expectedError error
		expected      *response.ListResult
	}{
		{
			description: "list clusters",
			resource:    "clusters",
			query: &query.QueryInfo{
				Pagination: &query.Pagination{
					PageSize: 10,
					Page:     1,
				},
				SortBy:    "name",
				Ascending: false,
				Filters:   nil,
			},
			expectedError: nil,
			expected: &response.ListResult{
				Items:       []interface{}{clusters[1], clusters[0]},
				Total:       2,
				PageSize:    10,
				TotalPages:  1,
				CurrentPage: 1,
			},
		},
	}

	//resource provider: k8s client -> Capinformer -> kubernetessharedinformerfactory
	factory, cli, err := prepare()
	if err != nil {
		t.Fatalf(err.Error())
	}

	handler := New(resource.NewResourceProcessor(factory, cli, nil))

	for _, test := range tests {
		res, err := handler.resourceProvider.List(test.resource, "", test.query)
		if err != nil {
			t.Errorf("failed with %s", err.Error())
		}

		if diff := cmp.Diff(res, test.expected); diff != "" {
			t.Errorf("%T differ (-got, +want): %s", test.expected, diff)
		}
	}
}

func TestHandleGetResources(t *testing.T) {

	tests := []struct {
		description   string
		name          string
		resource      string
		expectedError error
		expected      runtime.Object
	}{
		{
			description:   "get clusters",
			resource:      "clusters",
			name:          "test1",
			expectedError: nil,
			expected:      clusters[0].(*v1alpha1.Cluster),
		},
		{
			description:   "get clusters",
			resource:      "clusters",
			name:          "test2",
			expectedError: nil,
			expected:      clusters[1].(*v1alpha1.Cluster),
		},
	}

	//resource provider: k8s client -> Capinformer -> kubernetessharedinformerfactory
	factory, cli, err := prepare()
	if err != nil {
		t.Fatalf(err.Error())
	}

	handler := New(resource.NewResourceProcessor(factory, cli, nil))
	for _, test := range tests {
		handler.resourceProvider.Get(test.resource, "", "")
	}

}
