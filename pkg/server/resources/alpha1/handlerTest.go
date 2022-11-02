package alpha1

import (
	"fmt"
	"testing"

	"captain/pkg/bussiness/kube-resources/alpha1/resource"
	"captain/pkg/informers"
	"captain/pkg/server/config"
	"captain/pkg/test/fake"
	"captain/pkg/unify/query"
	"captain/pkg/unify/response"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	// metav1 "k8s.io/api/apps/v1"

	"github.com/google/go-cmp/cmp"
)

var (
	// request = restful.NewRequest()

	//default: 5
	KubeQps = 10
	//default: 10
	KubeBurst = 20

	KubeConfigPath = "~/.kube/config"

	deployments = []interface{}{
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "nginx",
				Namespace: "default",
				Labels:    map[string]string{"app": "nginx"},
			},
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"app": "nginx"},
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{"app": "nginx"},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:            "nginx",
								Image:           "nginx:latest",
								ImagePullPolicy: corev1.PullIfNotPresent,
							},
						},
					},
				},
			},
		},
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "busybox",
				Namespace: "default",
				Labels:    map[string]string{"app": "busybox"},
			},
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"app": "busybox"},
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{"app": "busybox"},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:            "busybox",
								Image:           "busybox:latest",
								ImagePullPolicy: corev1.PullIfNotPresent,
							},
						},
					},
				},
			},
		},
	}
)

func prepare() (informers.CapInformerFactory, error) {
	cli, err := fake.NewClientSet()
	if err != nil {
		return nil, err
	}

	informerFac := informers.NewInformerFactories(cli, nil)

	kubeInformer := informerFac.KubernetesSharedInformerFactory()

	for _, deploy := range deployments {
		err := kubeInformer.Apps().V1().Deployments().Informer().GetIndexer().Add(deploy)
		if err != nil {
			fmt.Printf("adding msg: %s", err.Error())
			return nil, err
		}
	}

	return informerFac, nil
}

func TestHandleListResources(t *testing.T) {
	tests := []struct {
		description   string
		namespace     string
		resource      string
		query         *query.QueryInfo
		expectedError error
		expected      *response.ListResult
	}{
		{
			description: "list namespaces",
			namespace:   "default",
			resource:    "deployment",
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
				Items:       []interface{}{deployments[1], deployments[0]},
				Total:       2,
				PageSize:    10,
				TotalPages:  1,
				CurrentPage: 1,
			},
		},
	}

	//resource provider: k8s client -> Capinformer -> kubernetessharedinformerfactory
	factory, err := prepare()
	if err != nil {
		t.Fatalf(err.Error())
	}

	handler := New(resource.NewResourceProcessor(factory, nil, config.New()))

	for _, test := range tests {
		res, err := handler.resourceProviderAlpha1.List("", "", test.resource, test.namespace, test.query)
		if err != nil {
			t.Errorf("failed with %s", err.Error())
		}

		if diff := cmp.Diff(res, test.expected); diff != "" {
			t.Errorf("%T differ (-got, +want): %s", test.expected, diff)
		}
	}
}
