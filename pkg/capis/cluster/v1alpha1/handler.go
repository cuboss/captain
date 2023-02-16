package v1alpha1

import (
	"captain/apis/cluster/v1alpha1"
	"captain/pkg/utils/clusterclient"
	"context"
	"errors"
	"strings"

	"github.com/emicklei/go-restful"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	namespace              = "captain-system"
	serviceAccountName     = "captain-admin"
	clusterRoleBindName    = serviceAccountName
	AnnoServiceAccountName = "kubernetes.io/service-account.name"

	clusterRoleName     = "cluster-admin"
	clusterRoleAPIGroup = "rbac.authorization.k8s.io"
)

type Handler struct {
	clusterclient.ClusterClients
}

func NewHandler(clients clusterclient.ClusterClients) *Handler {
	return &Handler{ClusterClients: clients}
}

func (h *Handler) ClusterAdminToken(request *restful.Request, response *restful.Response) {
	clustername := request.PathParameter("name")
	cluster, err := h.GetByClusterName(clustername)
	if err != nil {
		response.WriteAsJson(map[string]interface{}{
			"code":  404,
			"error": err.Error(),
		})
		return
	}
	region := cluster.Annotations[v1alpha1.ClusterRegion]
	if len(region) > 0 {
		clustername = strings.TrimPrefix(clustername, region+"-")
	}
	dryRun := request.QueryParameter("dryRun")
	token, err := h.getToken(region, clustername, dryRun)
	if err != nil {
		response.WriteAsJson(map[string]interface{}{
			"code":  400,
			"error": err.Error(),
		})
		return
	}
	response.WriteAsJson(map[string]string{
		"token": token,
	})
}

func (h *Handler) getToken(region, cluster, dryRun string) (string, error) {
	// 获取k8s client
	cli, err := h.ClusterClients.GetClientSet(region, cluster)
	if err != nil {
		return "", nil
	}
	// 尝试获取 token secret
	secret, err := getTokenSecret(cli)
	if err != nil {
		return "", err
	}
	if secret == nil && dryRun != "true" {
		// 不存在token secret，进行创建
		err = createAdminToken(cli)
		if err != nil {
			return "", err
		}
		secret, err = getTokenSecret(cli)
		if err != nil {
			return "", err
		}
	}
	if secret == nil {
		return "", errors.New("secret is nil")
	}
	return string(secret.Data["token"]), nil
}

func getTokenSecret(cli *kubernetes.Clientset) (*v1.Secret, error) {
	// 尝试获取 token secret
	secretList, err := cli.CoreV1().Secrets(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	var secret *v1.Secret
	if secretList != nil && len(secretList.Items) > 0 {
		for ix := range secretList.Items {
			secret = &secretList.Items[ix]
			if secret.Annotations[AnnoServiceAccountName] == serviceAccountName {
				return secret, nil
			}
		}
	}
	return nil, nil
}

func createAdminToken(cli *kubernetes.Clientset) error {
	_, err := cli.CoreV1().ServiceAccounts(namespace).Create(context.Background(), &v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceAccountName,
			Namespace: namespace,
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	_, err = cli.RbacV1().ClusterRoleBindings().Create(context.Background(), &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterRoleBindName,
		},
		Subjects: []rbacv1.Subject{rbacv1.Subject{
			Kind:      "ServiceAccount",
			Name:      serviceAccountName,
			Namespace: namespace,
		}},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     clusterRoleName,
			APIGroup: clusterRoleAPIGroup,
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}
