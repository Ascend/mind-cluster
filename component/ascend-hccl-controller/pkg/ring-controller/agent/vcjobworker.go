/* Copyright(C) 2020-2023. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package agent for logic
package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"huawei.com/npu-exporter/v5/common-utils/hwlog"
	apiCoreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/util/retry"

	"hccl-controller/pkg/ring-controller/common"
	ranktablev1 "hccl-controller/pkg/ring-controller/ranktable/v1"
)

const maxRankIndex = 10000

// Worker :The main function of Worker is to get the information of NPU from the generated POD,
// and then assemble it into a complete HCCL.JSON file.
type Worker interface {
	doWork(pod *apiCoreV1.Pod, podInfo *podIdentifier) (bool, bool)
	Statistic(stopTime time.Duration)
	WorkerCommon
}

// NewVCJobWorker : Generates a Worker that handles the VCJob type
func NewVCJobWorker(agent *BusinessAgent, job JobInfo, ranktable ranktablev1.RankTabler,
	replicasTotal int32, version int) *VCJobWorker {
	jobWorker := &VCJobWorker{
		WorkerInfo: WorkerInfo{
			kubeclientset:     agent.KubeClientSet,
			podsIndexer:       agent.PodsIndexer,
			informerFactory:   agent.informerFactory,
			recorder:          agent.recorder,
			dryRun:            agent.dryRun,
			statisticSwitch:   make(chan struct{}),
			configmapName:     fmt.Sprintf("%s-%s", ConfigmapPrefix, job.JobName),
			configmapData:     ranktable,
			statisticStopped:  false,
			cachedPodNum:      0,
			taskReplicasTotal: replicasTotal,
			cachedPods:        &sync.Map{},
			cachedIndex:       newCachedIndex(int(replicasTotal)),
			resourceVersion:   version,
		},
		JobInfo: job,
	}
	return jobWorker
}

func newCachedIndex(n int) *sync.Map {
	m := &sync.Map{}
	for i := 0; i < n; i++ {
		m.Store(strconv.Itoa(i), false)
	}
	return m
}

func (b *VCJobWorker) doWork(pod *apiCoreV1.Pod, podInfo *podIdentifier) (bool, bool) {
	hwlog.RunLog.Debugf("syncing %s", podInfo)
	if b.dryRun {
		return true, true
	}

	defer func() {
		if b.configmapData.GetStatus() == ConfigmapCompleted {
			return
		}
		if b.cacheRankTable() {
			if err := b.endRankTableConstruction(pod.Namespace); err != nil {
				hwlog.RunLog.Errorf("error end rank table construction: %s", err)
			}
		}
	}()

	forgetQueue, finished, err := b.doPreCheck(pod, podInfo)
	if err != nil {
		hwlog.RunLog.Debugf("error do pre check: %s", err)
		return forgetQueue, finished
	}

	if _, ok := b.cachedPods.Load(pod.UID); ok {
		return true, true
	}
	b.cachedPods.Store(pod.UID, true)
	b.modifyStatistics(1)
	return true, true
}

func (b *VCJobWorker) cacheRankTable() bool {
	if !b.tableConstructionFinished() {
		hwlog.RunLog.Debugf("job %s/%s rank table construction not finished", b.JobNamespace, b.JobName)
		return false
	}
	pods, err := b.getPodsFromCache()
	if err != nil {
		hwlog.RunLog.Errorf("error listing pods: %s", err.Error())
		return false
	}
	if !b.checkJobIsRunning(pods) {
		hwlog.RunLog.Debugf("job %s/%s not running", b.JobNamespace, b.JobName)
		return false
	}

	if err = b.cacheReadyPods(pods); err != nil {
		hwlog.RunLog.Errorf("error cache ready pods: %s", err.Error())
		return false
	}

	return true
}

func (b *VCJobWorker) getPodsFromCache() ([]*apiCoreV1.Pod, error) {
	return b.informerFactory.Core().V1().Pods().Lister().List(labels.SelectorFromSet(
		map[string]string{
			VolcanoJobNameKey:      b.JobInfo.JobName,
			VolcanoJobNamespaceKey: b.JobInfo.JobNamespace,
		}))
}

func (b *VCJobWorker) checkJobIsRunning(pods []*apiCoreV1.Pod) bool {
	readyPods := int32(0)
	for _, p := range pods {
		if p.GetDeletionTimestamp() == nil {
			readyPods++
		}
	}
	if readyPods != b.taskReplicasTotal {
		hwlog.RunLog.Infof("job %s/%s ready pods: %d, total pods: %d", b.JobNamespace, b.JobName, readyPods,
			b.taskReplicasTotal)
		return false
	}
	return true
}

func (b *VCJobWorker) cachePod(wg *sync.WaitGroup, pod *apiCoreV1.Pod, errs *sync.Map) {
	defer wg.Done()
	deviceInfo := pod.Annotations[PodDeviceKey]

	var instance ranktablev1.Instance
	if err := json.Unmarshal([]byte(deviceInfo), &instance); err != nil {
		errs.Store(pod.Name, fmt.Errorf("parse annotation of pod %s/%s error: %#v", pod.Namespace, pod.Name,
			err))
		return
	}
	if !ranktablev1.CheckDeviceInfo(&instance) {
		errs.Store(pod.Name, fmt.Errorf("deviceInfo failed the validation"))
		return
	}

	rankIndexStr, err := b.getOrSetPodIndex(pod)
	if err != nil {
		errs.Store(pod.Name, fmt.Errorf("error get or set pod index: %s", err))
		return
	}

	if err = b.configmapData.CachePodInfo(pod, instance, rankIndexStr); err != nil {
		errs.Store(pod.Name, fmt.Errorf("error cache pod info: %s", err))
	}
}

func (b *VCJobWorker) cacheReadyPods(pods []*apiCoreV1.Pod) error {
	errs := sync.Map{}

	wg := &sync.WaitGroup{}
	for _, p := range pods {
		wg.Add(1)
		go b.cachePod(wg, p, &errs)
	}
	wg.Wait()

	var err error
	errs.Range(func(key, value interface{}) bool {
		if value != nil {
			errVal, ok := value.(error)
			if !ok {
				hwlog.RunLog.Error("failed to convert value")
				return false
			}
			err = errVal
			return false
		}
		return true
	})

	return err
}

func (b *VCJobWorker) doPreCheck(pod *apiCoreV1.Pod, podInfo *podIdentifier) (bool, bool, error) {
	// scenario check A: For an identical job, create it immediately after deletion
	// check basis: job uid + creationTimestamp
	if !isReferenceJobSameWithBsnsWorker(pod, podInfo.jobName, b.JobUID) {
		if pod.CreationTimestamp.Before(&b.JobCreationTimestamp) {
			// old pod + new worker
			hwlog.RunLog.Debugf("syncing '%s' terminated: corresponding job worker is no "+
				"longer exist (basis: job uid + creationTimestamp)", podInfo)
			return true, false, errors.New("")
		}
		// new pod + old worker
		hwlog.RunLog.Infof("syncing '%s' delayed: corresponding job worker is "+
			"uninitialized (basis: job uid + creationTimestamp)", podInfo)
		return false, false, errors.New("")
	}
	// scenario check B: job set restart policy, delete pod
	// check basis: job version
	val, exists := pod.Annotations[PodJobVersion]
	if !exists {
		return true, true, fmt.Errorf("the key of " + PodJobVersion + " does not exist")
	}
	version64, err := strconv.ParseInt(val, common.Decimal, common.BitSize32)
	if err != nil {
		return true, true, fmt.Errorf("syncing '%s' failed, parse pod annotation error: %v", podInfo, err)
	}
	// job restart action will increase job version number
	if version64 < int64(b.JobVersion) {
		return true, true, fmt.Errorf("syncing '%s' terminated: corresponding job worker "+
			"is no longer exist (basis: job version number)", podInfo)
	}

	// check whether pod has used npu
	if used := containerUsedChip(pod); !used {
		return true, true, fmt.Errorf("pod %s doesn't use npu, so no longer dealing with it", podInfo)
	}
	// scenario check C: if current pod use chip, its' device info may not be ready
	// check basis: limits + annotations
	if (podInfo.eventType == EventAdd || podInfo.eventType == EventUpdate) && (!isPodAnnotationsReady(pod,
		podInfo.String()) || pod.Status.PodIP == "") {
		return false, false, fmt.Errorf("pod %s doesn't have device info, so no longer dealing with it", podInfo)
	}

	return true, true, nil
}

// Statistic : Determine whether CM has been built, process the build completion or change the goroutine exit signal.
// No need to add lock here, deviation from true value is acceptable
func (b *VCJobWorker) Statistic(stopTime time.Duration) {
	for {
		select {
		case c, ok := <-b.statisticSwitch:
			if !ok {
				hwlog.RunLog.Error(c)
			}
			return
		default:
			if b.taskReplicasTotal == b.cachedPodNum {
				hwlog.RunLog.Infof("rank table build progress for %s/%s is completed",
					b.JobNamespace, b.JobName)
				b.CloseStatistic()
				return
			}
			hwlog.RunLog.Infof("rank table build progress for %s/%s: pods need to be cached = %d,"+
				"pods already cached = %d", b.JobNamespace, b.JobName, b.taskReplicasTotal, b.cachedPodNum)
			time.Sleep(stopTime)
		}
	}
}

// WorkerCommon : The common methods of Worker, these methods have a certain degree of fixedness,
// if the new Worker type does not apply to these methods, they can be overwritten.
type WorkerCommon interface {
	handleAddUpdateEvent(podInfo *podIdentifier, pod *apiCoreV1.Pod) error
	handleDeleteEvent(podInfo *podIdentifier) error
	tableConstructionFinished() bool
	endRankTableConstruction(string) error
	modifyStatistics(diff int32)
	// CloseStatistic : to close statisticSwitch chan
	CloseStatistic()
	syncHandler(pod *apiCoreV1.Pod, podInfo *podIdentifier) error
}

func (b *WorkerInfo) syncHandler(pod *apiCoreV1.Pod, podInfo *podIdentifier) error {
	hwlog.RunLog.Infof("syncHandler start, current pod is %s", podInfo)

	// if use 0 chip, end pod sync
	if b.taskReplicasTotal == 0 && b.tableConstructionFinished() {
		hwlog.RunLog.Infof("job %s/%s doesn't use d chip, rank table construction is finished",
			podInfo.namespace, podInfo.jobName)
		if err := b.endRankTableConstruction(pod.Namespace); err != nil {
			return err
		}
		hwlog.RunLog.Infof("rank table for job %s/%s has finished construction", podInfo.namespace, podInfo.jobName)
		return nil //  need return directly
	}

	// dryRun is for empty running and will not be committed
	if b.dryRun {
		hwlog.RunLog.Infof("I'am handling %s", podInfo)
		return nil
	}

	if podInfo.eventType == EventAdd || podInfo.eventType == EventUpdate {
		return b.handleAddUpdateEvent(podInfo, pod)
	}
	hwlog.RunLog.Infof("undefined condition, pod: %s", podInfo)
	return nil
}

func (b *WorkerInfo) tableConstructionFinished() bool {
	b.statisticMu.Lock()
	defer b.statisticMu.Unlock()

	return b.cachedPodNum == b.taskReplicasTotal
}

func (b *WorkerInfo) handleAddUpdateEvent(podInfo *podIdentifier, pod *apiCoreV1.Pod) error {
	hwlog.RunLog.Debugf("current addUpdate pod is %s", podInfo)
	// because this annotation is already used to filter pods in previous step scenario check C
	// it can be used to identify if pod use chip here
	deviceInfo, exist := pod.Annotations[PodDeviceKey]
	if !exist {
		return errors.New("the key of " + PodDeviceKey + " does not exist ")
	}
	var instance ranktablev1.Instance
	if err := json.Unmarshal([]byte(deviceInfo), &instance); err != nil {
		return fmt.Errorf("parse annotation of pod %s/%s error: %#v", pod.Namespace, pod.Name, err)
	}
	if !ranktablev1.CheckDeviceInfo(&instance) {
		return errors.New("deviceInfo failed the validation")
	}

	hwlog.RunLog.Infof("deviceId: (%#v)", deviceInfo)

	b.cmMu.Lock()
	defer b.cmMu.Unlock()
	var rankIndexStr string
	// Get rankIndex from pod, use rankIndex if rankIndex exists in pod, use memory if it doesn't.
	rankIndexStr, rankExist := pod.Annotations[PodRankIndexKey]
	if rankExist {
		return fmt.Errorf("pod %s/%s already has rankIndex: %s", pod.Namespace, pod.Name, rankIndexStr)
	}
	rankIndexStr = strconv.Itoa(int(b.rankIndex))

	// Cache device info from the pod
	err := b.configmapData.CachePodInfo(pod, instance, rankIndexStr)
	if err != nil {
		return err
	}

	err = b.updatePod(pod, func(newPod *apiCoreV1.Pod) {
		newPod.Annotations[PodRankIndexKey] = rankIndexStr
	})
	if err != nil {
		return err
	}
	b.rankIndex++

	// Cache pod num plus one
	b.modifyStatistics(1)
	hwlog.RunLog.Infof("rank table build progress for %s/%s: pods need to be cached = %d, "+
		"pods already cached = %d", podInfo.namespace, podInfo.jobName, b.taskReplicasTotal, b.cachedPodNum)
	// update configmap if finishing caching all pods' info
	errs := updateWithFinish(b, podInfo.namespace)
	if errs != nil {
		return errs
	}

	return nil
}

func validate(rank int64) error {
	if rank < 0 || rank > maxRankIndex {
		return fmt.Errorf("rank index from pod is error")
	}
	return nil
}

func (b *WorkerInfo) getOrSetPodIndex(pod *apiCoreV1.Pod) (string, error) {
	var rankIndexStr string

	rankIndexStr, rankExist := pod.Annotations[PodRankIndexKey]

	if rankExist {
		hwlog.RunLog.Infof("pod(%s/%s) already has rankIndex: %s", pod.Namespace, pod.Name, rankIndexStr)
	} else {
		for _, env := range pod.Spec.Containers[0].Env {
			if env.Name == vcPodIndexKey {
				rankIndexStr = env.Value
			}
		}
		if rankIndexStr == "" {
			return "", errors.New("index env not found in pod")
		}
		err := b.updatePod(pod, func(newPod *apiCoreV1.Pod) {
			newPod.Annotations[PodRankIndexKey] = rankIndexStr
		})
		if err != nil {
			return "", err
		}
		hwlog.RunLog.Infof("set pod(%s/%s) rankIndex: %s", pod.Namespace, pod.Name, rankIndexStr)
	}
	b.cachedIndex.Store(rankIndexStr, true)
	return rankIndexStr, nil
}

func (b *WorkerInfo) updatePod(pod *apiCoreV1.Pod, updateFunc func(*apiCoreV1.Pod)) error {
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		newPod, err := b.kubeclientset.CoreV1().Pods(pod.Namespace).Get(context.TODO(), pod.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		updateFunc(newPod)
		_, err = b.kubeclientset.CoreV1().Pods(pod.Namespace).Update(context.TODO(), newPod, metav1.UpdateOptions{})
		return err
	})
}

func (b *WorkerInfo) handleDeleteEvent(podInfo *podIdentifier) error {
	hwlog.RunLog.Infof("current handleDeleteEvent pod is %s", podInfo)
	b.cmMu.Lock()
	defer b.cmMu.Unlock()

	status := b.configmapData.GetStatus()

	err := b.configmapData.RemovePodInfo(podInfo.namespace, podInfo.uid)
	if err != nil {
		hwlog.RunLog.Warn(err)
	}

	hwlog.RunLog.Infof("start to remove data of pod %s/%s", podInfo.namespace, podInfo.name)

	if status == ConfigmapCompleted {
		b.configmapData.SetStatus(ConfigmapInitializing)
		hwlog.RunLog.Infof("pod(%s/%s) is delete, start to update configmap(%s) to initializing", podInfo.namespace,
			podInfo.name, b.configmapName)
		err = updateConfigMap(b, podInfo.namespace)
		if err != nil {
			b.configmapData.SetStatus(ConfigmapCompleted)
			return err
		}
	}

	rankIndex := podInfo.rankIndex
	if rankIndex != "" {
		_, ok := b.cachedIndex.Load(rankIndex)
		if !ok {
			return fmt.Errorf("cannot find pod(%v) rank index %s", podInfo, rankIndex)
		}
		b.cachedIndex.Store(rankIndex, false)
	}
	hwlog.RunLog.Infof("data of pod %s/%s is removed", podInfo.namespace, podInfo.name)
	if b.cachedPods != nil {
		b.cachedPods.Delete(podInfo.uid)
	}
	b.configmapData.DeletePod(podInfo.uid)
	b.modifyStatistics(-1)
	return nil
}

func (b *WorkerInfo) endRankTableConstruction(namespace string) error {
	b.configmapData.SetStatus(ConfigmapCompleted)
	b.configmapData.BeforeUpdate()
	b.resourceVersion++
	hwlog.RunLog.Infof("job is ready, start to update configmap(%s/%s) to completed", namespace, b.configmapName)
	if err := updateConfigMap(b, namespace); err != nil {
		hwlog.RunLog.Error("update configmap failed")
		b.resourceVersion--
		return err
	}

	return nil
}

// modifyStatistics statistic about how many pods have already cached
func (b *WorkerInfo) modifyStatistics(diff int32) {
	atomic.AddInt32(&b.cachedPodNum, diff)
}

// CloseStatistic : to close statisticSwitch chan
func (b *WorkerInfo) CloseStatistic() {
	if !b.statisticStopped {
		close(b.statisticSwitch)
		b.statisticStopped = true
	}
}

func updateWithFinish(b *WorkerInfo, namespace string) error {
	if b.tableConstructionFinished() {
		if err := b.endRankTableConstruction(namespace); err != nil {
			return err
		}
	}
	return nil
}

func getWorkName(labels map[string]string) string {
	if label, ok := labels[VolcanoJobNameKey]; ok {
		return label
	}
	if label, ok := labels[DeploymentNameKey]; ok {
		return label
	}
	return ""
}

func updateConfigMap(w *WorkerInfo, namespace string) error {
	cm, err := w.kubeclientset.CoreV1().ConfigMaps(namespace).Get(context.TODO(),
		w.configmapName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("get configmap error: %v", err)
	}
	oldCM, ok := cm.Data[ConfigmapKey]
	if !ok {
		err = fmt.Errorf("old cm ranktable not exists")
		hwlog.RunLog.Debug(err)
		return err
	}
	hwlog.RunLog.Debugf("old cm ranktable %#v", oldCM)
	label910, exist := (*cm).Labels[Key910]
	if !exist || !(label910 == Val910B || label910 == Val910) {
		return fmt.Errorf("invalid configmap label: %s", label910)
	}
	dataByteArray, err := json.Marshal(w.configmapData)
	if err != nil {
		return fmt.Errorf("marshal configmap data error: %v", err)
	}
	cm.Data[ConfigmapKey] = string(dataByteArray[:])
	cm.Data[ConfigmapVersion] = strconv.Itoa(w.resourceVersion)

	if _, err = w.kubeclientset.CoreV1().ConfigMaps(namespace).Update(context.TODO(), cm,
		metav1.UpdateOptions{}); err != nil {
		return fmt.Errorf("failed to update ConfigMap for Job %v", err)
	}
	hwlog.RunLog.Debugf("new cm ranktable %s, version: %d", cm.Data[ConfigmapKey], w.resourceVersion)
	return nil
}
