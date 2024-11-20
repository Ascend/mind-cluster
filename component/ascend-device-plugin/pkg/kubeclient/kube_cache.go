/* Copyright(C) 2023. Huawei Technologies Co.,Ltd. All rights reserved.
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
	"errors"
	"hash/fnv"
	"sync"
	"time"

	"huawei.com/npu-exporter/v5/common-utils/hwlog"
	"k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"

	"Ascend-device-plugin/pkg/common"
)

var (
	podCache            = map[types.UID]*podInfo{}
	lock                = sync.Mutex{}
	nodeServerIp        string
	nodeDeviceInfoCache *common.NodeDeviceInfoCache
)

type podInfo struct {
	*v1.Pod
	updateTime time.Time
}

const podDeleteOperator = "delete"
const podAddOperator = "add"
const podUpdateOperator = "update"

const (
	timeIntervalForCheckPod = 10 * time.Minute
	periodicForStartCheck   = 600
	podCacheTimeout         = time.Hour
)

// PodInformerInspector check pod in cache
func (ki *ClientK8s) PodInformerInspector(ctx context.Context) {
	if ki == nil {
		return
	}
	hashVal := fnv.New32()
	if _, err := hashVal.Write([]byte(ki.NodeName)); err != nil {
		hwlog.RunLog.Errorf("failed to write nodeName to hash, err: %v", err)
		return
	}

	val := hashVal.Sum32() % periodicForStartCheck
	hwlog.RunLog.Infof("after %d second, pod informer inspector will start", val)
	time.Sleep(time.Duration(val) * time.Second)
	wait.Until(func() { ki.checkPodInCache(ctx) }, timeIntervalForCheckPod, ctx.Done())
}

func (ki *ClientK8s) checkPodInCache(ctx context.Context) {
	lock.Lock()
	defer lock.Unlock()
	needDelete := make([]types.UID, 0)
	needRefresh := make([]types.UID, 0)
	for uid, pi := range podCache {
		hwlog.RunLog.Debugf("check pod(%s/%s) in cache, updateTime: %v, now: %v", pi.Namespace, pi.Name,
			pi.updateTime.Format(time.DateTime), time.Now().Format(time.DateTime))
		if time.Since(pi.updateTime) < podCacheTimeout {
			continue
		}
		pod, err := ki.getPod(ctx, pi.Namespace, pi.Name)
		if err != nil {
			if k8serr.IsNotFound(err) {
				hwlog.RunLog.Infof("delete pod(%s/%s) from cache", pi.Namespace, pi.Name)
				needDelete = append(needDelete, uid)
				continue
			}
			hwlog.RunLog.Errorf("failed to get pod %s/%s, err: %v", pi.Pod.Namespace, pi.Pod.Name, err)
			continue
		}
		if pod.Spec.NodeName != ki.NodeName || pod.UID != uid {
			hwlog.RunLog.Infof("delete pod(%s/%s) from cache", pod.Namespace, pod.Name)
			needDelete = append(needDelete, uid)
			continue
		}
		needRefresh = append(needRefresh, uid)
	}
	for _, uid := range needDelete {
		delete(podCache, uid)
	}
	for _, uid := range needRefresh {
		podCache[uid].updateTime = time.Now()
	}
}

func (ki *ClientK8s) getPod(ctx context.Context, namespace, name string) (*v1.Pod, error) {
	return ki.Clientset.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
}

// UpdatePodList update pod list by informer
func UpdatePodList(oldObj, newObj interface{}, operator string) {
	newPod, ok := newObj.(*v1.Pod)
	if !ok {
		return
	}
	lock.Lock()
	defer lock.Unlock()
	switch operator {
	case podAddOperator, podUpdateOperator:
		hwlog.RunLog.Infof("pod(%s/%s) is %s to cache", newPod.Namespace, newPod.Name, operator)
		podCache[newPod.UID] = &podInfo{
			Pod:        newPod,
			updateTime: time.Now(),
		}
	case podDeleteOperator:
		hwlog.RunLog.Infof("pod(%s/%s) is deleted from cache", newPod.Namespace, newPod.Name)
		delete(podCache, newPod.UID)
	default:
		hwlog.RunLog.Errorf("operator is undefined, find operater: %s", operator)
	}
}

func (ki *ClientK8s) refreshPodList() {
	newV1PodList, err := ki.GetAllPodList()
	if err != nil {
		hwlog.RunLog.Errorf("get pod list from api-server failed: %v", err)
		return
	}
	newPodCache := map[types.UID]*podInfo{}
	for _, pod := range newV1PodList.Items {
		newPodCache[pod.UID] = &podInfo{
			Pod:        &pod,
			updateTime: time.Now(),
		}
	}
	lock.Lock()
	podCache = newPodCache
	lock.Unlock()
	ki.IsApiErr = false
	hwlog.RunLog.Info("get new pod list success")
}

// GetAllPodListCache get pod list by field selector with cache,
func (ki *ClientK8s) GetAllPodListCache() []v1.Pod {
	if ki == nil {
		return []v1.Pod{}
	}
	if ki.IsApiErr {
		ki.refreshPodList()
	}

	lock.Lock()
	defer lock.Unlock()
	pods := make([]v1.Pod, 0, len(podCache))
	for _, pi := range podCache {
		pods = append(pods, *pi.Pod)
	}
	return pods
}

// GetActivePodListCache is to get active pod list with cache
func (ki *ClientK8s) GetActivePodListCache() []v1.Pod {
	if ki == nil {
		return []v1.Pod{}
	}
	if ki.IsApiErr {
		ki.refreshPodList()
	}

	lock.Lock()
	defer lock.Unlock()
	newPodList := make([]v1.Pod, 0, common.GeneralMapSize)
	for _, pi := range podCache {
		if err := common.CheckPodNameAndSpace(pi.GetName(), common.PodNameMaxLength); err != nil {
			hwlog.RunLog.Warnf("pod name syntax illegal, err: %v", err)
			continue
		}
		if err := common.CheckPodNameAndSpace(pi.GetNamespace(), common.PodNameSpaceMaxLength); err != nil {
			hwlog.RunLog.Warnf("pod namespace syntax illegal, err: %v", err)
			continue
		}
		if pi.Status.Phase == v1.PodFailed || pi.Status.Phase == v1.PodSucceeded {
			continue
		}
		newPodList = append(newPodList, *pi.Pod)
	}
	return newPodList
}

// GetPodCache get pod by namespace and name with cache
func (ki *ClientK8s) GetPodCache(namespace, name string) v1.Pod {
	if ki == nil {
		return v1.Pod{}
	}
	lock.Lock()
	defer lock.Unlock()
	for _, p := range podCache {
		if p.Namespace == namespace && p.Name == name {
			return *p.Pod
		}
	}
	return v1.Pod{}
}

// GetNodeServerIDCache Get Node Server ID with cache
func (ki *ClientK8s) GetNodeServerIDCache() (string, error) {
	if ki == nil {
		return "", errors.New("client is nil")
	}
	if nodeServerIp != "" {
		return nodeServerIp, nil
	}
	serverID, err := ki.GetNodeServerID()
	if err != nil {
		return "", err
	}
	nodeServerIp = serverID
	return serverID, nil
}

// GetDeviceInfoCMCache get device info configMap with cache
func (ki *ClientK8s) GetDeviceInfoCMCache() *common.NodeDeviceInfoCache {
	if ki == nil {
		return nil
	}
	return nodeDeviceInfoCache
}

// WriteDeviceInfoDataIntoCMCache write deviceinfo into config map with cache
func (ki *ClientK8s) WriteDeviceInfoDataIntoCMCache(deviceInfo map[string]string, manuallySeparateNPU string) error {
	if ki == nil {
		return nil
	}
	newNodeDeviceInfoCache, err := ki.WriteDeviceInfoDataIntoCM(deviceInfo, manuallySeparateNPU)
	if err != nil {
		return err
	}
	nodeDeviceInfoCache = newNodeDeviceInfoCache
	return nil
}

// SetNodeDeviceInfoCache set device info cache
func (ki *ClientK8s) SetNodeDeviceInfoCache(deviceInfoCache *common.NodeDeviceInfoCache) {
	if ki == nil {
		return
	}
	nodeDeviceInfoCache = deviceInfoCache
}
