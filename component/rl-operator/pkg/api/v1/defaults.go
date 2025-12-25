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
	"errors"

	commonv1 "github.com/kubeflow/common/pkg/apis/common/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"rl-operator/pkg/common"
)

type defaultFuncPair struct {
	obj         runtime.Object
	defaultFunc func(obj interface{})
}

// Int32 is a helper routine that allocates a new int32 value
// to store v and returns a pointer to it.
func Int32(v int32) *int32 {
	return &v
}

func RegisterDefaults(scheme *runtime.Scheme) error {
	if scheme == nil {
		return errors.New("schema is nil")
	}
	defaultFuncPairs := []defaultFuncPair{
		{obj: &VerlJob{}, defaultFunc: func(obj interface{}) { SetDefaultsVerlJob(obj.(*VerlJob)) }},
		{obj: &VerlJobList{}, defaultFunc: func(obj interface{}) { SetDefaultsVerlJobList(obj.(*VerlJobList)) }},
		{obj: &RayCluster{}, defaultFunc: func(obj interface{}) { SetDefaultsRayCluster(obj.(*RayCluster)) }},
		{obj: &RayClusterList{}, defaultFunc: func(obj interface{}) { SetDefaultsRayClusterList(obj.(*RayClusterList)) }},
	}
	for _, pair := range defaultFuncPairs {
		scheme.AddTypeDefaultingFunc(pair.obj, pair.defaultFunc)
	}
	return nil
}

func SetDefaultsVerlJob(in *VerlJob) {
	if in == nil {
		return
	}
	if in.Spec.RunPolicy.CleanPodPolicy == nil {
		running := commonv1.CleanPodPolicyNone
		in.Spec.RunPolicy.CleanPodPolicy = &running
	}
	setDefaultPort(in)
	setDefaultVerlLabels(in)
	for _, spec := range in.Spec.ReplicaSpecs {
		if spec == nil {
			continue
		}
		setDefaultReplicas(spec)
	}
}

func SetDefaultsVerlJobList(in *VerlJobList) {
	if in == nil {
		return
	}
	for i := range in.Items {
		item := &in.Items[i]
		SetDefaultsVerlJob(item)
	}
}

func SetDefaultsRayCluster(in *RayCluster) {
	if in == nil {
		return
	}
	setDefaultPort(in)
	// Set default cleanpod policy to Running.
	if in.Spec.RunPolicy.CleanPodPolicy == nil {
		running := commonv1.CleanPodPolicyNone
		in.Spec.RunPolicy.CleanPodPolicy = &running
	}
	for _, spec := range in.Spec.ReplicaSpecs {
		if spec == nil {
			continue
		}
		setDefaultReplicas(spec)
	}
}

func SetDefaultsRayClusterList(in *RayClusterList) {
	if in == nil {
		return
	}
	for i := range in.Items {
		item := &in.Items[i]
		SetDefaultsRayCluster(item)
	}
}

func setDefaultVerlLabels(in *VerlJob) {
	annos := in.GetAnnotations()
	if _, ok := annos[common.AutoSubmitLabelKey]; !ok {
		annos[common.AutoSubmitLabelKey] = "false"
	}
	if _, ok := annos[common.VerlExecLabelKey]; !ok {
		annos[common.VerlExecLabelKey] = "verl.trainer.main_ppo"
	}
	if _, ok := annos[common.VerlConfigLabelKey]; !ok {
		annos[common.VerlConfigLabelKey] = "config/ppo_trainer.yaml"
	}
	if _, ok := annos[common.CheckRayStatusLabelKey]; !ok {
		annos[common.VerlConfigLabelKey] = "false"
	}
}

func setDefaultPort(in client.Object) {
	labels := in.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}
	if _, ok := labels[common.RayGcsPortLabelKey]; !ok {
		labels[common.RayGcsPortLabelKey] = common.DefaultRayGcsPort
	}
	if _, ok := labels[common.RayClientPortLabelKey]; !ok {
		labels[common.RayClientPortLabelKey] = common.DefaultRayClientPort
	}
	if _, ok := labels[common.RayDashboardPortLabelKey]; !ok {
		labels[common.RayDashboardPortLabelKey] = common.DefaultRayDashboardPort
	}
	in.SetLabels(labels)
}

func setDefaultReplicas(spec *commonv1.ReplicaSpec) {
	if spec.Replicas == nil {
		spec.Replicas = Int32(1)
	}
	if spec.RestartPolicy == "" {
		spec.RestartPolicy = common.DefaultRestartPolicy
	}
}
