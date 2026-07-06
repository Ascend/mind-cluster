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

// Package kubeclient a series of k8s function
package kubeclient

import (
	"context"
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	"Ascend-device-plugin/pkg/common"
	"ascend-common/common-utils/hwlog"
)

// PodManager defines pod operations. Read operations can be served by kubelet or apiserver.
type PodManager interface {
	GetPod(ctx context.Context, pod *v1.Pod) (*v1.Pod, error)
	PatchPod(pod *v1.Pod, data []byte) (*v1.Pod, error)
	GetActivePodList() ([]v1.Pod, error)
	GetAllPodList() (*v1.PodList, error)
	GetAllPodListCache() []v1.Pod
	GetActivePodListCache() []v1.Pod
	InitPodInformer()
	PodInformerInspector(ctx context.Context)
	UpdatePodList(newObj interface{}, operator EventType)
	FlushPodCacheNextQuerying()
}

// Apiserver implements pod operations through Kubernetes apiserver.
type Apiserver struct {
	client *ClientK8s
}

// Kubelet implements pod read operations through kubelet local /pods endpoint.
type Kubelet struct {
	client *ClientK8s
}

func (ki *ClientK8s) initPodManagers() {
	if ki.apiserver == nil {
		ki.apiserver = &Apiserver{client: ki}
	}
	if ki.kubelet == nil {
		ki.kubelet = &Kubelet{client: ki}
	}
}

func (ki *ClientK8s) getChannel() PodManager {
	ki.initPodManagers()
	if common.ParamOption.GetPodFromKubelet {
		return ki.kubelet
	}
	return ki.apiserver
}

func (a *Apiserver) GetPod(ctx context.Context, pod *v1.Pod) (*v1.Pod, error) {
	return a.client.getPodFromApiserver(ctx, pod)
}

func (a *Apiserver) PatchPod(pod *v1.Pod, data []byte) (*v1.Pod, error) {
	return a.client.patchPodToApiserver(pod, data)
}

func (a *Apiserver) GetActivePodList() ([]v1.Pod, error) {
	return a.client.getActivePodListFromApiserver()
}

func (a *Apiserver) GetAllPodList() (*v1.PodList, error) {
	return a.client.getAllPodListFromApiserver()
}

func (a *Apiserver) GetAllPodListCache() []v1.Pod {
	return a.client.getAllPodListCache()
}

func (a *Apiserver) GetActivePodListCache() []v1.Pod {
	return a.client.getActivePodListCache()
}

func (a *Apiserver) InitPodInformer() {
	a.client.initPodInformer()
}

func (a *Apiserver) PodInformerInspector(ctx context.Context) {
	a.client.podInformerInspector(ctx)
}

func (a *Apiserver) UpdatePodList(newObj interface{}, operator EventType) {
	a.client.updatePodList(newObj, operator)
}

func (a *Apiserver) FlushPodCacheNextQuerying() {
	a.client.IsApiErr = true
}

func (k *Kubelet) GetPod(ctx context.Context, pod *v1.Pod) (*v1.Pod, error) {
	if pod == nil {
		return nil, fmt.Errorf("param pod is nil")
	}
	podList, err := k.GetAllPodList()
	if err != nil {
		return nil, err
	}
	for index := range podList.Items {
		item := &podList.Items[index]
		if item.Namespace == pod.Namespace && item.Name == pod.Name {
			return item, nil
		}
	}
	return nil, errors.NewNotFound(v1.Resource("pods"), pod.Name)
}

func (k *Kubelet) PatchPod(pod *v1.Pod, data []byte) (*v1.Pod, error) {
	// Kubelet /pods is read-only. Keep pod annotation writes on apiserver for compatibility.
	return k.client.patchPodToApiserver(pod, data)
}

func (k *Kubelet) GetActivePodList() ([]v1.Pod, error) {
	podList, err := k.GetAllPodList()
	if err != nil {
		return nil, err
	}
	return checkPodList(filterActivePods(podList))
}

func (k *Kubelet) GetAllPodList() (*v1.PodList, error) {
	podList, err := k.client.getPodsByKltPort()
	if err != nil {
		return nil, err
	}
	if podList == nil {
		return nil, fmt.Errorf("pod list is invalid")
	}
	items := make([]v1.Pod, 0, len(podList.Items))
	for _, pod := range podList.Items {
		if pod.Spec.NodeName != "" && pod.Spec.NodeName != k.client.NodeName {
			continue
		}
		items = append(items, pod)
	}
	filtered := &v1.PodList{
		TypeMeta: podList.TypeMeta,
		ListMeta: podList.ListMeta,
		Items:    items,
	}
	if len(filtered.Items) >= common.MaxPodLimit {
		hwlog.RunLog.Errorf("The number of pods exceeds the upper limit, count: %d", len(filtered.Items))
		return nil, fmt.Errorf("pod list count invalid")
	}
	return filtered, nil
}

func (k *Kubelet) GetAllPodListCache() []v1.Pod {
	return k.client.getAllPodListCache()
}

func (k *Kubelet) GetActivePodListCache() []v1.Pod {
	return k.client.getActivePodListCache()
}

func (k *Kubelet) InitPodInformer() {
	hwlog.RunLog.Info("get pod from kubelet, skip pod informer")
}

func (k *Kubelet) PodInformerInspector(ctx context.Context) {
	k.client.podInformerInspector(ctx)
}

func (k *Kubelet) UpdatePodList(newObj interface{}, operator EventType) {
	k.client.updatePodList(newObj, operator)
}

func (k *Kubelet) FlushPodCacheNextQuerying() {
	k.client.refreshPodList()
}

func filterActivePods(podList *v1.PodList) *v1.PodList {
	if podList == nil {
		return nil
	}
	activePods := make([]v1.Pod, 0, len(podList.Items))
	for _, pod := range podList.Items {
		if pod.Status.Phase == v1.PodFailed || pod.Status.Phase == v1.PodSucceeded {
			continue
		}
		activePods = append(activePods, pod)
	}
	return &v1.PodList{
		TypeMeta: podList.TypeMeta,
		ListMeta: podList.ListMeta,
		Items:    activePods,
	}
}

func refreshPodCacheFromList(podList *v1.PodList) {
	newPodCache := map[types.UID]*podInfo{}
	if podList != nil {
		for _, pod := range podList.Items {
			func(pod v1.Pod) {
				newPodCache[pod.UID] = &podInfo{
					Pod:        &pod,
					updateTime: time.Now(),
				}
			}(pod)
		}
	}
	lock.Lock()
	podCache = newPodCache
	lock.Unlock()
}
