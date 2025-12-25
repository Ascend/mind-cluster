/*
Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.

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

package v1

import (
	"fmt"
	"reflect"

	commonv1 "github.com/kubeflow/common/pkg/apis/common/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"ascend-common/common-utils/hwlog"
	"rl-operator/pkg/common"
)

type PodInfo struct {
	rtype   commonv1.ReplicaType
	index   int
	spec    *commonv1.ReplicaSpec
	svcName string
}

func (r *RayWorkerReconciler) newPodInfo(
	job client.Object, rtype commonv1.ReplicaType, spec *commonv1.ReplicaSpec) (*PodInfo, error) {
	svcName, hasName := job.GetLabels()[common.ServiceNameLabelKey]
	if !hasName {
		return nil, fmt.Errorf("service label of %s<%s> not found in job, maybe RayHead service not started",
			reflect.TypeOf(job).Name(), job.GetName())
	}
	return &PodInfo{
		spec:    spec,
		svcName: svcName,
		rtype:   rtype,
	}, nil
}

func (pi *PodInfo) DeepCopy() *PodInfo {
	return &PodInfo{
		rtype:   pi.rtype,
		spec:    pi.spec,
		index:   pi.index,
		svcName: pi.svcName,
	}
}

func setHeadPodCmdArgs(podTemplate *corev1.PodTemplateSpec, labels map[string]string) {
	paramMap := common.GenHeadParamMap(labels)
	command := common.GenerateRayStartCommand(common.RayHeadType, paramMap)
	common.AddCommandToSpec(podTemplate, command)
	hwlog.RunLog.Infof("set head pod command<%v>", podTemplate.Spec.Containers[0].Command)
}

func setWorkerPodCmdArgs(podTemplate *corev1.PodTemplateSpec, pi *PodInfo, labels map[string]string) {
	paramMap := common.GenWorkerParamMap(labels, pi.svcName)
	command := common.GenerateRayStartCommand(string(pi.rtype), paramMap)
	common.AddCommandToSpec(podTemplate, command)
}
