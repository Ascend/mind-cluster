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
	"fmt"
	"hash/fnv"
	"math"
	"strconv"
	"sync"
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"

	"Ascend-device-plugin/pkg/common"
	"ascend-common/common-utils/hwlog"
)

var (
	podCache            = map[types.UID]*podInfo{}
	lock                = sync.Mutex{}
	nodeServerIp        string
	serverUsageLabel    string
	nodeDeviceInfoCache *common.NodeDeviceInfoCache
)

type podInfo struct {
	*v1.Pod
	updateTime time.Time
}

const (
	timeIntervalForCheckPod = 10 * time.Minute
	periodicForStartCheck   = 600
	podCacheTimeout         = time.Hour
)

// PodInformerInspector check pod in cache
func (ki *ClientK8s) PodInformerInspector(ctx context.Context) {
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
			if errors.IsNotFound(err) {
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
func (ki *ClientK8s) UpdatePodList(newObj interface{}, operator EventType) {
	newPod, ok := newObj.(*v1.Pod)
	if !ok {
		return
	}
	lock.Lock()
	defer lock.Unlock()
	obj, exist, err := ki.PodInformer.GetIndexer().GetByKey(newPod.Namespace + "/" + newPod.Name)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get pod %s/%s from indexer, err: %v", newPod.Namespace, newPod.Name, err)
		return
	}
	if !exist {
		hwlog.RunLog.Infof("pod(%s/%s) is not exist in indexer", newPod.Namespace, newPod.Name)
		delete(podCache, newPod.UID)
		common.TriggerUpdate(fmt.Sprintf("pod %v", string(operator)))
		return
	}
	pod, ok := obj.(*v1.Pod)
	if !ok {
		return
	}
	ki.updatePodPredicateTime(pod)
	podCache[pod.UID] = &podInfo{
		Pod:        pod,
		updateTime: time.Now(),
	}

	common.TriggerUpdate(fmt.Sprintf("pod %v", string(operator)))
}

func (ki *ClientK8s) updatePodPredicateTime(newPod *v1.Pod) {
	cachedPodInfo, exists := podCache[newPod.UID]
	if !exists {
		hwlog.RunLog.Debugf("pod(%s/%s) is not exist in cache", newPod.Namespace, newPod.Name)
		return
	}
	if cachedPodInfo.Pod.Annotations == nil {
		hwlog.RunLog.Debugf("pod(%s/%s) annotation is nil", newPod.Namespace, newPod.Name)
		return
	}
	cachedPredicateTime := cachedPodInfo.Pod.Annotations[common.PodPredicateTime]
	if cachedPredicateTime != strconv.FormatUint(math.MaxUint64, common.BaseDec) {
		hwlog.RunLog.Debugf("pod(%s/%s) predicate-time is not maxUint64", newPod.Namespace, newPod.Name)
		return
	}
	needsUpdate := false
	if newPod.Annotations == nil {
		hwlog.RunLog.Warnf("pod(%s/%s) annotation is nil", newPod.Namespace, newPod.Name)
		newPod.Annotations = make(map[string]string)
		needsUpdate = true
	}
	if newPod.Annotations[common.PodPredicateTime] != cachedPredicateTime {
		hwlog.RunLog.Warnf("pod(%s/%s) predicate-time is not equal to cache", newPod.Namespace, newPod.Name)
		needsUpdate = true
	}
	if needsUpdate {
		oldValueForLog := newPod.Annotations[common.PodPredicateTime]
		if oldValueForLog == "" {
			oldValueForLog = "<nil>"
		}
		hwlog.RunLog.Infof("Correcting pod %s/%s predicate-time from %s to %s",
			newPod.Name, newPod.Namespace, oldValueForLog, cachedPredicateTime)
		newPod.Annotations[common.PodPredicateTime] = cachedPredicateTime
		annotation := map[string]string{common.PodPredicateTime: strconv.FormatUint(math.MaxUint64, common.BaseDec)}
		go ki.TryUpdatePodCacheAnnotation(newPod, annotation)
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
		// attention: using 'for range' for value slice, the pointer addr of the slice element is the same
		func(pod v1.Pod) {
			newPodCache[pod.UID] = &podInfo{
				Pod:        &pod,
				updateTime: time.Now(),
			}
		}(pod)
	}
	lock.Lock()
	podCache = newPodCache
	lock.Unlock()
	ki.IsApiErr = false
	hwlog.RunLog.Info("get new pod list success")
}

// GetAllPodListCache get pod list by field selector with cache,
func (ki *ClientK8s) GetAllPodListCache() []v1.Pod {
	if ki.IsApiErr {
		ki.refreshPodList()
	}
	pods := make([]v1.Pod, 0, len(podCache))
	lock.Lock()
	defer lock.Unlock()

	for _, pi := range podCache {
		pods = append(pods, *pi.Pod)
	}
	return pods
}

// GetActivePodListCache is to get active pod list with cache
func (ki *ClientK8s) GetActivePodListCache() []v1.Pod {
	if ki.IsApiErr {
		ki.refreshPodList()
	}
	newPodList := make([]v1.Pod, 0, common.GeneralMapSize)
	lock.Lock()
	defer lock.Unlock()
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

// GetNodeIpCache Get Node Server ID with cache
func (ki *ClientK8s) GetNodeIpCache() (string, error) {
	if nodeServerIp != "" {
		return nodeServerIp, nil
	}
	nodeIp, err := ki.GetNodeIp()
	if err != nil {
		return "", err
	}
	nodeServerIp = nodeIp
	return nodeIp, nil
}

// GetServerUsageLabelCache get node label:server-usage, and cache it in memory, if label updated, restart is required
// if server-usage label is set return the value of label
// if server-usage label is not set return 'unknown'
func (ki *ClientK8s) GetServerUsageLabelCache() (string, error) {
	if serverUsageLabel != "" {
		hwlog.RunLog.Debugf("get node server usage label from cache,label:%s", serverUsageLabel)
		return serverUsageLabel, nil
	}
	node, err := ki.GetNode()
	if err != nil {
		return "", err
	}
	label, ok := node.Labels[common.ServerUsageLabelKey]
	if !ok {
		serverUsageLabel = "unknown"
		hwlog.RunLog.Errorf("failed to get server-usage label")
		return "unknown", nil
	}
	serverUsageLabel = label
	hwlog.RunLog.Debugf("update node server usage label ,label:%s", serverUsageLabel)
	return serverUsageLabel, nil
}

// GetDeviceInfoCMCache get device info configMap with cache
func (ki *ClientK8s) GetDeviceInfoCMCache() *common.NodeDeviceInfoCache {
	return nodeDeviceInfoCache
}

// WriteDeviceInfoDataIntoCMCache write deviceinfo into config map with cache
func (ki *ClientK8s) WriteDeviceInfoDataIntoCMCache(deviceInfo map[string]string, manuallySeparateNPU string,
	switchInfo common.SwitchFaultInfo, superPodID, serverIndex int32) error {
	newNodeDeviceInfoCache, err := ki.WriteDeviceInfoDataIntoCM(deviceInfo, manuallySeparateNPU, switchInfo,
		superPodID, serverIndex)
	if err != nil {
		return err
	}

	nodeDeviceInfoCache = newNodeDeviceInfoCache
	return nil
}

// SetNodeDeviceInfoCache set device info cache
func (ki *ClientK8s) SetNodeDeviceInfoCache(deviceInfoCache *common.NodeDeviceInfoCache) {
	nodeDeviceInfoCache = deviceInfoCache
}
