/*
Copyright 2019 The KubeSphere Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package crd

import (
	"captain/apis/gaia/v1alpha1"
	"captain/pkg/controller/crd_controller"
	"context"
	"k8s.io/client-go/kubernetes"
)

const orphanFinalizer = "orphan.finalizers.kubesphere.io"

type Interface interface {
	CreateGaiaClusterTemplate(ctx context.Context, cluster crd_controller.CkeClusterVo) (*v1alpha1.GaiaCluster, error)
}

type CrdOperator struct {
	k8sclient      kubernetes.Interface
}

func NewCrdOperator(k8sclient kubernetes.Interface) Interface {
	return &CrdOperator{
		k8sclient:      k8sclient,
	}
}

func (co *CrdOperator) CreateGaiaClusterTemplate(ctx context.Context, cluster crd_controller.CkeClusterVo) (*v1alpha1.GaiaCluster, error) {
	return crd_controller.CreateGaiaCluster(ctx, cluster)
}
