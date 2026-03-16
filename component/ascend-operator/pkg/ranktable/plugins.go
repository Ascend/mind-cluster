/*
Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.

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

package ranktable

import (
	"fmt"
	"strconv"
	"sync"

	commonv1 "github.com/kubeflow/common/pkg/apis/common/v1"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"volcano.sh/apis/pkg/apis/batch/v1alpha1"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	mindxdlv1 "ascend-operator/pkg/api/v1"
	"ascend-operator/pkg/ranktable/common"
)

var (
	pluginRegistry = &registry{
		plugins: make(map[string]RankTablePlugin),
	}
	cachedPlugins []RankTablePlugin
	pluginsCached bool
)

type registry struct {
	sync.RWMutex
	plugins map[string]RankTablePlugin
}

func registerPlugin(p RankTablePlugin) {
	pluginRegistry.Lock()
	defer pluginRegistry.Unlock()
	pluginRegistry.plugins[p.Name()] = p
}

func getAllPlugins() []RankTablePlugin {
	if pluginsCached {
		return cachedPlugins
	}

	pluginRegistry.RLock()
	defer pluginRegistry.RUnlock()

	cachedPlugins = make([]RankTablePlugin, 0, len(pluginRegistry.plugins))
	for _, p := range pluginRegistry.plugins {
		cachedPlugins = append(cachedPlugins, p)
	}
	pluginsCached = true
	return cachedPlugins
}

// RankTablePlugin the interface of rankTable plugin to classify methods
type RankTablePlugin interface {
	Name() string
	CanHandle(obj metav1.Object) bool
	ExtractObjToAscendJob(obj metav1.Object) (*mindxdlv1.AscendJob, error)
	ShouldGenerateRankTable(obj metav1.Object) bool
}

func init() {
	registerPlugin(&deploymentPlugin{})
	registerPlugin(&statefulSetPlugin{})
	registerPlugin(&vcJobPlugin{})
}

type deploymentPlugin struct{}

// Name the name of deployment plugin
func (p *deploymentPlugin) Name() string {
	return mindxdlv1.DeploymentPlugin
}

// CanHandle to make sure the deployment can be handled
func (p *deploymentPlugin) CanHandle(obj metav1.Object) bool {
	_, ok := obj.(*appv1.Deployment)
	return ok
}

// ShouldGenerateRankTable to make sure the deployment can be used to generate ranktable file
func (p *deploymentPlugin) ShouldGenerateRankTable(obj metav1.Object) bool {
	deploy, ok := obj.(*appv1.Deployment)
	if !ok {
		hwlog.RunLog.Errorf("Object to Deployment failed")
		return false
	}

	if _, ok := deploy.Labels[api.AtlasTaskLabel]; ok {
		return true
	}

	return hasRankTableMount(&deploy.Spec.Template)
}

// ExtractObjToAscendJob return acjob from deployment
func (p *deploymentPlugin) ExtractObjToAscendJob(obj metav1.Object) (*mindxdlv1.AscendJob, error) {
	deploy, ok := obj.(*appv1.Deployment)
	if !ok {
		return nil, fmt.Errorf("object is not Deployment")
	}
	repSpec := map[commonv1.ReplicaType]*commonv1.ReplicaSpec{
		mindxdlv1.DeploymentKindName: {
			Template: deploy.Spec.Template,
			Replicas: deploy.Spec.Replicas,
		},
	}
	objectMeta := deploy.ObjectMeta
	for key, value := range deploy.Spec.Template.Annotations {
		if oldValue, ok := objectMeta.Annotations[key]; ok && oldValue != value {
			hwlog.RunLog.Warnf("%s annotation %s value %s change to %s", deploy.Name, key, oldValue, value)
		}
		if objectMeta.Annotations == nil {
			objectMeta.Annotations = make(map[string]string)
		}
		objectMeta.Annotations[key] = value
	}

	return &mindxdlv1.AscendJob{
		TypeMeta:   deploy.TypeMeta,
		ObjectMeta: objectMeta,
		Spec: mindxdlv1.AscendJobSpec{
			ReplicaSpecs: repSpec,
		},
	}, nil
}

type statefulSetPlugin struct{}

// Name the name of the statefulSet plugin
func (p *statefulSetPlugin) Name() string {
	return mindxdlv1.StatefulSetPlugin
}

// CanHandle to make sure the statefulSet can be handled
func (p *statefulSetPlugin) CanHandle(obj metav1.Object) bool {
	_, ok := obj.(*appv1.StatefulSet)
	return ok
}

// ShouldGenerateRankTable to make sure the statefulSet can be used to generate ranktable file
func (p *statefulSetPlugin) ShouldGenerateRankTable(obj metav1.Object) bool {
	statefulSet, ok := obj.(*appv1.StatefulSet)
	if !ok {
		hwlog.RunLog.Errorf("Object to StatefulSet failed")
		return false
	}
	if _, ok := statefulSet.Labels[api.AtlasTaskLabel]; ok {
		return true
	}

	return hasRankTableMount(&statefulSet.Spec.Template)
}

// ExtractObjToAscendJob return acjob from statefulSet
func (p *statefulSetPlugin) ExtractObjToAscendJob(obj metav1.Object) (*mindxdlv1.AscendJob, error) {
	statefulSet, ok := obj.(*appv1.StatefulSet)
	if !ok {
		return nil, fmt.Errorf("object is not StatefulSet")
	}

	repSpec := map[commonv1.ReplicaType]*commonv1.ReplicaSpec{
		mindxdlv1.StatefulSetKindName: {
			Template: statefulSet.Spec.Template,
			Replicas: statefulSet.Spec.Replicas,
		},
	}
	objectMeta := statefulSet.ObjectMeta
	for key, value := range statefulSet.Spec.Template.Annotations {
		if oldValue, ok := objectMeta.Annotations[key]; ok && oldValue != value {
			hwlog.RunLog.Warnf("%s annotation %s value %s change to %s", statefulSet.Name, key, oldValue, value)
		}
		if objectMeta.Annotations == nil {
			objectMeta.Annotations = make(map[string]string)
		}
		objectMeta.Annotations[key] = value
	}
	return &mindxdlv1.AscendJob{
		TypeMeta:   statefulSet.TypeMeta,
		ObjectMeta: objectMeta,
		Spec: mindxdlv1.AscendJobSpec{
			ReplicaSpecs: repSpec,
		}}, nil
}

type vcJobPlugin struct{}

// Name the name of vcjob plugin
func (p *vcJobPlugin) Name() string {
	return mindxdlv1.VcJobPlugin
}

// CanHandle to make sure the vcjob can be handled
func (p *vcJobPlugin) CanHandle(obj metav1.Object) bool {
	_, ok := obj.(*v1alpha1.Job)
	return ok
}

// ShouldGenerateRankTable to make sure the vcjob can be used to generate ranktable file
func (p *vcJobPlugin) ShouldGenerateRankTable(obj metav1.Object) bool {
	vcjob, ok := obj.(*v1alpha1.Job)
	if !ok {
		hwlog.RunLog.Errorf("Object to VcJob failed")
		return false
	}

	if _, ok := vcjob.Labels[api.AtlasTaskLabel]; ok {
		return true
	}

	return hasRankTableMountInVcJob(vcjob)
}

// ExtractObjToAscendJob return acjob from vcjob
func (p *vcJobPlugin) ExtractObjToAscendJob(obj metav1.Object) (*mindxdlv1.AscendJob, error) {
	vcjob, ok := obj.(*v1alpha1.Job)
	if !ok {
		return nil, fmt.Errorf("object is not VcJob")
	}

	repSpecs := map[commonv1.ReplicaType]*commonv1.ReplicaSpec{}
	for i, task := range vcjob.Spec.Tasks {
		repSpecs[commonv1.ReplicaType(mindxdlv1.VcJobKindName+strconv.Itoa(i))] = &commonv1.ReplicaSpec{
			Template: task.Template,
			Replicas: &task.Replicas,
		}
	}
	return &mindxdlv1.AscendJob{
		TypeMeta:   vcjob.TypeMeta,
		ObjectMeta: vcjob.ObjectMeta,
		Spec: mindxdlv1.AscendJobSpec{
			ReplicaSpecs: repSpecs,
		},
	}, nil
}

// FindPluginForObject return the ranktable plugin
func FindPluginForObject(obj metav1.Object) RankTablePlugin {
	for _, plugin := range getAllPlugins() {
		if plugin.CanHandle(obj) {
			return plugin
		}
	}
	return nil
}

// GetAscendJobFromObject extract config from object of crd resources and return ascendjob
func GetAscendJobFromObject(obj metav1.Object) (*mindxdlv1.AscendJob, error) {
	plugin := FindPluginForObject(obj)
	if plugin == nil {
		return nil, fmt.Errorf("no plugin found for object type %T", obj)
	}
	return plugin.ExtractObjToAscendJob(obj)
}

func hasRankTableMountInVcJob(job *v1alpha1.Job) bool {
	for _, task := range job.Spec.Tasks {
		if hasRankTableMount(&task.Template) {
			return true
		}
	}
	return false
}

func hasRankTableMount(template *corev1.PodTemplateSpec) bool {
	for _, volume := range template.Spec.Volumes {
		if volume.Name == common.RanktableStr {
			return true
		}
	}
	return false
}
