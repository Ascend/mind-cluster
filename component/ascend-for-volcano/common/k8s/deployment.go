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

	"k8s.io/api/apps/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
)

// GetDeployment Get deployment from k8s.
func GetDeployment(kubeClient kubernetes.Interface, namespace, depName string) (*v1.Deployment, error) {
	dep, err := kubeClient.AppsV1().Deployments(namespace).Get(context.TODO(), depName, v12.GetOptions{})
	if err != nil {
		klog.V(util.LogInfoLev).Infof("namespace %s deployment %s not in kubernetes, err: %s", namespace, depName, err)
	}
	return dep, err
}

// ClusterDDeploymentIsExist ClusterD deployment is exist
func ClusterDDeploymentIsExist(kubeClient kubernetes.Interface) bool {
	dep, err := GetDeployment(kubeClient, util.MindXDlNameSpace, util.ClusterD)
	return err == nil && dep != nil
}
