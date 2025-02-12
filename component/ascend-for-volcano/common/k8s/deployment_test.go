/*
Copyright(C)2020-2025. Huawei Technologies Co.,Ltd. All rights reserved.

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

/*
Package k8s is using for the k8s operation.
*/
package k8s

import (
	"context"
	"reflect"
	"testing"

	"k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
)

const (
	deployName      = "test"
	deployNamespace = "default"
)

func TestGetDeployment(t *testing.T) {
	tests := []struct {
		name       string
		kubeClient kubernetes.Interface
		mockDeploy *v1.Deployment
		namespace  string
		depName    string
		want       *v1.Deployment
		wantErr    bool
	}{
		{
			name:       "TestGetDeployment normal case",
			kubeClient: fake.NewSimpleClientset(),
			namespace:  deployNamespace,
			depName:    deployName,
			mockDeploy: &v1.Deployment{
				ObjectMeta: metav1.ObjectMeta{Name: deployName, Namespace: deployNamespace}},
			want: &v1.Deployment{
				ObjectMeta: metav1.ObjectMeta{Name: deployName, Namespace: deployNamespace}},
			wantErr: false,
		},
		{
			name:       "TestGetDeployment not found",
			kubeClient: fake.NewSimpleClientset(),
			namespace:  deployNamespace,
			depName:    deployName,
			mockDeploy: &v1.Deployment{
				ObjectMeta: metav1.ObjectMeta{Name: deployName + "1", Namespace: deployNamespace}},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.kubeClient.AppsV1().Deployments(tt.namespace).Create(context.Background(),
				tt.mockDeploy, metav1.CreateOptions{})
			got, err := GetDeployment(tt.kubeClient, tt.namespace, tt.depName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDeployment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetDeployment() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClusterDDeploymentIsExist(t *testing.T) {
	tests := []struct {
		name       string
		kubeClient kubernetes.Interface
		mockDeploy *v1.Deployment
		want       bool
	}{
		{
			name:       "01-TestClusterDDeploymentIsExist normal case",
			kubeClient: fake.NewSimpleClientset(),
			mockDeploy: &v1.Deployment{
				ObjectMeta: metav1.ObjectMeta{Name: util.ClusterD, Namespace: util.MindXDlNameSpace}},
			want: true,
		},
		{
			name:       "02-TestClusterDDeploymentIsExist not found clusterd deployment",
			kubeClient: fake.NewSimpleClientset(),
			mockDeploy: &v1.Deployment{
				ObjectMeta: metav1.ObjectMeta{Name: deployName, Namespace: deployNamespace}},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.kubeClient.AppsV1().Deployments(util.MindXDlNameSpace).Create(context.Background(),
				tt.mockDeploy, metav1.CreateOptions{})
			if got := ClusterDDeploymentIsExist(tt.kubeClient); got != tt.want {
				t.Errorf("ClusterDDeploymentIsExist() = %v, want %v", got, tt.want)
			}
		})
	}
}
