// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package job a series of job function
package job

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"huawei.com/npu-exporter/v6/common-utils/hwlog"
	apiCoreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"clusterd/pkg/common/util"
)

var _ PodWorker = &Worker{}

// PodWorker The main function of PodWorker is to get the information of NPU from the generated POD
type PodWorker interface {
	doPodWork(pod *apiCoreV1.Pod, podInfo *podIdentifier)
	Stat(stopTime time.Duration)
	handlePodAddUpdateEvent(podInfo *podIdentifier, pod *apiCoreV1.Pod) error
	handlePodDelEvent(podInfo *podIdentifier) error
	constructionFinished() bool
	endConstruction(podInfo *podIdentifier) error
	modifyStat(diff int32)
	CloseStat()
	handler(pod *apiCoreV1.Pod, podInfo *podIdentifier) error
	UpdateCMWhenJobEnd(podKeyInfo *podIdentifier) error
	HandleJobStatus(jobPhase string) string
	UpdateJobNodeHealthyStatus(nodeName string, nodeHealth bool)
	UpdateJobDeviceHealthyStatus(nodeName string, networkUnhealthyCards, unHealthyCards string)
	GetJobHealth() (bool, []string)
	PGRunning() bool
	GetBaseInfo() Info
	GetDeviceNumPerNode() int
}

// NewJobWorker Generates a PodWorker that handles the Job
func NewJobWorker(agent *Agent, job Info, ranktable RankTabler, replicasTotal int32) *Worker {
	jobWorker := &Worker{
		WorkerInfo: WorkerInfo{
			vcClient:   agent.vcClient,
			clientSet:  agent.KubeClientSet,
			podIndexer: agent.podsIndexer,
			statSwitch: make(chan struct{}),
			CMName:     fmt.Sprintf("%s-%s", ConfigmapPrefix, job.Name),
			CMData:     ranktable, statStopped: false,
			cachedPodNum:      0,
			jobReplicasTotal:  replicasTotal,
			succeedPodNum:     0,
			podSchedulerCache: []string{},
			JobType:           job.JobType,
		},
		Info: job}
	return jobWorker
}

func (b *Worker) doPodWork(pod *apiCoreV1.Pod, podInfo *podIdentifier) {
	err := b.doPreCheck(pod, podInfo)
	if err != nil {
		return
	}
	// start to sync current pod
	if err = b.handler(pod, podInfo); err != nil {
		hwlog.RunLog.Errorf("error syncing '%s': %v", podInfo, err)
	}
}

func (b *Worker) doPreCheck(pod *apiCoreV1.Pod, podInfo *podIdentifier) error {
	// scenario check A: For an identical job, create it immediately after deletion
	// check basis: job uid + creationTimestamp
	if !isReferenceJobSameWithWorker(pod, podInfo.jobName, b.Uid) {
		if pod.CreationTimestamp.Before(&b.CreationTimestamp) {
			// old pod + new worker
			hwlog.RunLog.Errorf("syncing '%s' terminated: corresponding job worker is no "+
				"longer exist (basis: job uid + creationTimestamp)", podInfo)
			return fmt.Errorf("pod %s does not exist", podInfo)
		}
		// new pod + old worker
		hwlog.RunLog.Errorf("syncing '%s' delayed: corresponding job worker is "+
			"uninitialized (basis: job uid + creationTimestamp)", podInfo)
		return fmt.Errorf("pod %s does not initialize", podInfo)
	}
	// scenario check C: if current pod use chip, the device info may not be ready
	// check basis: limits + annotations
	if (podInfo.eventType == EventAdd || podInfo.eventType == EventUpdate) && !isPodAnnotationsReady(pod,
		podInfo.podInfo2String()) && containerUsedChip(pod) {
		hwlog.RunLog.Errorf("syncing '%s' terminated: pod annotation is not ready", podInfo)
		return fmt.Errorf("pod %s does not initialize", podInfo)
	}
	if b.CMData.GetStatus() == ConfigmapCompleted {
		hwlog.RunLog.Infof("syncing '%s' terminated: corresponding rank table is completed", podInfo)
		return fmt.Errorf("pod %s has completed ranktable", podInfo)
	}
	return nil
}

// Stat Determine whether CM has been built, process the build completion or change the goroutine exit signal.
// No need to add lock here, deviation from true value is acceptable
func (b *Worker) Stat(stopTime time.Duration) {
	for {
		select {
		case c, ok := <-b.statSwitch:
			if !ok {
				hwlog.RunLog.Infof("statSwitch error : %v", c)
			}
			return
		default:
			if b.jobReplicasTotal == b.cachedPodNum {
				hwlog.RunLog.Infof("rank table build progress for %s/%s is completed",
					b.Namespace, b.Name)
				b.CloseStat()
				return
			}
			hwlog.RunLog.Infof("rank table build progress for %s/%s: pods need to be cached = %d,"+
				"pods already cached = %d", b.Namespace, b.Name, b.jobReplicasTotal, b.cachedPodNum)
			time.Sleep(stopTime)
		}
	}
}

func (b *WorkerInfo) handler(pod *apiCoreV1.Pod, podInfo *podIdentifier) error {
	hwlog.RunLog.Debugf("handler start, current pod is %s", podInfo)

	// if pod use 0 chip, end pod sync
	if b.jobReplicasTotal == 0 && b.constructionFinished() {
		hwlog.RunLog.Infof("job %s/%s doesn't use d chip, rank table construction is finished", podInfo.namespace,
			podInfo.jobName)
		if err := b.endConstruction(podInfo); err != nil {
			return err
		}
		hwlog.RunLog.Infof("rank table for job %s/%s has finished construction", podInfo.namespace, podInfo.jobName)
		// need return directly
		return nil
	}

	// dryRun is for empty running and will not be committed
	if b.dryRun {
		hwlog.RunLog.Debugf("dryRun handling: %s", podInfo)
		return nil
	}

	if podInfo.eventType == EventAdd || podInfo.eventType == EventUpdate {
		hwlog.RunLog.Debugf("current addUpdate pod is %s", podInfo)
		return b.handlePodAddUpdateEvent(podInfo, pod)
	}
	hwlog.RunLog.Infof("undefined condition, pod: %s", podInfo)
	return nil
}

func (b *WorkerInfo) constructionFinished() bool {
	b.statMu.Lock()
	defer b.statMu.Unlock()

	return b.cachedPodNum == b.jobReplicasTotal
}

// handlePodWithoutChip handle pod with no chip
func (b *WorkerInfo) handlePodWithoutChip(podInfo *podIdentifier, pod *apiCoreV1.Pod) {
	for _, cachedSchedulerUID := range b.podSchedulerCache {
		if cachedSchedulerUID == string(pod.UID) {
			hwlog.RunLog.Warnf("pod %s/%s without npu is already cached", pod.Namespace, pod.Name)
			return
		}
	}
	b.podSchedulerCache = append(b.podSchedulerCache, string(pod.UID))
	b.modifyStat(1)
	hwlog.RunLog.Debugf("pod %s does not use npu, pod cached num %d, job replicas total %d",
		podInfo, b.cachedPodNum, b.jobReplicasTotal)
	if err := b.updateWithFinish(podInfo); err != nil {
		hwlog.RunLog.Errorf("pod %s ranktable error: %v", podInfo, err)
	}
}

// handlePodAddUpdateEvent handle pod with add or update event
func (b *WorkerInfo) handlePodAddUpdateEvent(podInfo *podIdentifier, pod *apiCoreV1.Pod) error {
	// check whether pod has used npu
	if !containerUsedChip(pod) {
		b.handlePodWithoutChip(podInfo, pod)
		return nil
	}
	deviceInfo, exist := pod.Annotations[PodDeviceKey]
	if !exist {
		return fmt.Errorf("the key of " + PodDeviceKey + " does not exist ")
	}
	var instance Instance
	if err := json.Unmarshal([]byte(deviceInfo), &instance); err != nil {
		return fmt.Errorf("parse annotation of pod %s/%s error: %#v", pod.Namespace, pod.Name, err)
	}
	podLabel, podLabelExist := pod.Labels[PodLabelKey]
	if podLabelExist {
		ModelFramework = podLabel
	}
	b.CmMutex.Lock()
	defer b.CmMutex.Unlock()
	tmpRankIndex := b.rankIndex
	rankIndexStr, rankExist := pod.Annotations[PodRankIndexKey]
	if rankExist {
		rank, err := strconv.ParseInt(rankIndexStr, Decimal, BitSize32)
		if err != nil {
			return err
		}
		if err = validateRank(rank); err != nil {
			return err
		}
		b.rankIndex = int(rank)
	}
	err := b.CMData.CachePodInfo(pod, instance, &b.rankIndex)
	if rankExist {
		b.rankIndex = tmpRankIndex
	}
	if err != nil {
		return err
	}
	b.modifyStat(1)
	hwlog.RunLog.Infof("rank table build progress for %s/%s: pods need to be cached = %d, "+
		"pods already cached = %d", podInfo.namespace, podInfo.jobName, b.jobReplicasTotal, b.cachedPodNum)
	if err = b.updateWithFinish(podInfo); err != nil {
		return err
	}
	b.setSharedTorIp(pod)
	return nil
}

func (b *WorkerInfo) setSharedTorIp(pod *apiCoreV1.Pod) {
	if pod.Annotations == nil {
		return
	}
	k, ok := pod.Annotations[torTag]
	if !ok || k != sharedTor {
		return
	}
	sharedTorIp, ok := pod.Annotations[torIpTag]
	if !ok {
		return
	}
	b.SharedTorIp = append(b.SharedTorIp, sharedTorIp)
	return
}

func validateRank(rank int64) error {
	if rank < 0 || rank > maxRankIndex {
		return fmt.Errorf("rank index from pod is error")
	}
	return nil
}

// deletePodUIDFromList delete pod UID from cache
func (b *WorkerInfo) deletePodUIDFromList(podInfo *podIdentifier) {
	index := -1
	for i, podUID := range b.podSchedulerCache {
		if podUID == podInfo.UID {
			index = i
			break
		}
	}
	if index != -1 {
		b.podSchedulerCache = append(b.podSchedulerCache[:index], b.podSchedulerCache[index+1:]...)
		return
	}
}

// handlePodDelEvent handle pod delete event
func (b *WorkerInfo) handlePodDelEvent(podInfo *podIdentifier) error {
	hwlog.RunLog.Infof("current handlePodDelEvent pod is %s", podInfo)

	b.CmMutex.Lock()
	defer b.CmMutex.Unlock()
	b.CMData.SetStatus(ConfigmapInitializing)
	b.deletePodUIDFromList(podInfo)

	err := b.CMData.RemovePodInfo(podInfo.namespace, podInfo.name)
	if err != nil {
		hwlog.RunLog.Warnf("no device info found, might be a no chip pod: %v", err)
	}

	hwlog.RunLog.Infof("start to remove data of pod %s/%s", podInfo.namespace, podInfo.name)
	err = b.UpdateConfigMap(podInfo, StatusJobDelete)
	if err != nil {
		return err
	}
	b.modifyStat(-1)
	hwlog.RunLog.Infof("data of pod %s/%s is removed", podInfo.namespace, podInfo.name)

	return nil
}

// endConstruction rank table has done
func (b *WorkerInfo) endConstruction(podInfo *podIdentifier) error {
	b.CMData.SetStatus(ConfigmapCompleted)
	if err := b.UpdateConfigMap(podInfo, StatusJobRunning); err != nil {
		hwlog.RunLog.Errorf("update configmap failed")
		return err
	}

	return nil
}

// modifyStatistics statistic about how many pods have already cached
func (b *WorkerInfo) modifyStat(diff int32) {
	if b.cachedPodNum == 0 && diff < 0 {
		hwlog.RunLog.Warn("cached pod num cannot be less than 0")
		return
	}
	b.statMu.Lock()
	b.cachedPodNum += diff
	b.statMu.Unlock()
}

// CloseStat : to close statSwitch chan
func (b *WorkerInfo) CloseStat() {
	if !b.statStopped {
		close(b.statSwitch)
		b.statStopped = true
	}
}

func (b *WorkerInfo) updateWithFinish(podInfo *podIdentifier) error {
	if b.constructionFinished() {
		if err := b.endConstruction(podInfo); err != nil {
			return err
		}
	}
	return nil
}

// UpdateConfigMap updates the job summary configmap
func (b *WorkerInfo) UpdateConfigMap(podInfo *podIdentifier, jobStatus string) error {
	cm, err := b.clientSet.CoreV1().ConfigMaps(podInfo.namespace).Get(context.TODO(),
		b.CMName, metav1.GetOptions{})
	if err != nil {
		hwlog.RunLog.Errorf("get configmap namespace %s name %s failed, error: %v", podInfo.namespace, b.CMName, err)
		return fmt.Errorf("get configmap error: %v", err)
	}
	_, ok := cm.Data[ConfigmapKey]
	if !ok {
		hwlog.RunLog.Errorf("old cm ranktable not exists %v", err)
		return fmt.Errorf("old cm ranktable not exists")
	}
	if jobStatus == StatusJobRunning {
		cm.Data[JobStatus] = StatusJobRunning
		cm.Data[DeleteTime] = "0"
		cm.Data[ConfigmapOperator] = OperatorAdd
	} else {
		if cm.Data[JobStatus] == StatusJobFail {
			return nil
		}
		if cm.Data[JobStatus] != StatusJobSucceed {
			hwlog.RunLog.Infof("pod %s:%s deleted, job status update to failed", podInfo.namespace, podInfo.name)
			cm.Data[JobStatus] = StatusJobFail
		} else {
			hwlog.RunLog.Infof("pod %s:%s deleted, job status is complete", podInfo.namespace, podInfo.name)
		}
	}
	label910, exist := (*cm).Labels[Key910]
	if !exist || !(label910 == Val910B || label910 == Val910) {
		return fmt.Errorf("invalid configmap label: %s", label910)
	}
	dataByteArray, err := json.Marshal(b.CMData)
	if err != nil {
		return fmt.Errorf("marshal configmap data error: %v", err)
	}
	cm.Data[ConfigmapKey] = string(dataByteArray[:])
	cm.Data[FrameWork] = ModelFramework
	if ModelFramework == ptFramework {
		b.updateSharedTorInfo(podInfo, cm)
	}
	if upErr := b.updateJobHccLJson(*cm); upErr != nil {
		return upErr
	}
	b.rankIndex = b.CMData.GetPodNum()
	hwlog.RunLog.Debugf("new cm ranktable %s", cm.Data[ConfigmapKey])
	return nil
}

func (b *WorkerInfo) updateJobHccLJson(cm apiCoreV1.ConfigMap) error {
	tmpCm := cm
	hcclJsons := b.CMData.GetHccLJsonSlice()
	for i := 0; i < len(hcclJsons); i++ {
		tmpCm.Data[cmCutNumKey] = strconv.Itoa(len(hcclJsons))
		tmpCm.Data[cmIndex] = strconv.Itoa(i)
		if i != 0 {
			tmpCm.Name = cm.Name + "-" + strconv.Itoa(i)
		}
		tmpCm.Data[ConfigmapKey] = hcclJsons[i]
		if ucErr := util.CreateOrUpdateCm(b.clientSet, &tmpCm); ucErr != nil {
			return fmt.Errorf("failed to update ConfigMap for Job %v", ucErr)
		}
	}
	return nil
}

func (b *WorkerInfo) updateSharedTorInfo(podInfo *podIdentifier, cm *apiCoreV1.ConfigMap) {
	sharedTorInfo, msErr := json.Marshal(util.RemoveSliceDuplicateElement(b.SharedTorIp))
	if msErr != nil {
		hwlog.RunLog.Warnf("pod %s:%s set shared tor ip failed by %v", podInfo.namespace, podInfo.name, msErr)
	} else {
		cm.Data[torIpTag] = string(sharedTorInfo)
	}
	cm.Data[masterAddrKey] = b.initMasterAddrByJobType(podInfo)
	return
}

func (b *WorkerInfo) initMasterAddrByJobType(podInfo *podIdentifier) string {
	if b.JobType == vcJobKind {
		return b.CMData.GetFirstServerIp()
	}
	return util.GetServiceIpWithRetry(b.clientSet, podInfo.namespace, podInfo.jobName+acJobMasterSuffix)
}

// UpdateCMWhenJobEnd is to update configmap from pod stop job
func (b *WorkerInfo) UpdateCMWhenJobEnd(podKeyInfo *podIdentifier) error {
	cm, err := b.clientSet.CoreV1().ConfigMaps(podKeyInfo.namespace).Get(context.TODO(),
		b.CMName, metav1.GetOptions{})
	if err != nil {
		hwlog.RunLog.Errorf("get configmap namespace %s name %s failed, error: %v",
			podKeyInfo.namespace, b.CMName, err)
		return fmt.Errorf("get configmap error: %v", err)
	}
	if cm.Data[JobStatus] == StatusJobFail {
		return nil
	}
	pod, err := b.clientSet.CoreV1().Pods(podKeyInfo.namespace).Get(context.TODO(), podKeyInfo.name, metav1.GetOptions{})
	if err != nil {
		hwlog.RunLog.Errorf("get pod namespace %s name %s failed, error: %v",
			podKeyInfo.namespace, b.CMName, err)
		return fmt.Errorf("get pod error: %v", err)
	}
	podPhase := string(pod.Status.Phase)
	if podPhase == PhaseJobRunning || podPhase == PhaseJobPending {
		hwlog.RunLog.Debugf("current pod status is %s", string(pod.Status.Phase))
		return nil
	}
	curJobStatus := b.HandleJobStatus(podPhase)
	switch curJobStatus {
	case PhaseJobRunning:
		return nil
	case StatusJobFail:
		hwlog.RunLog.Infof("job status update to %s", StatusJobFail)
		cm.Data[JobStatus] = StatusJobFail
		cm.Data[DeleteTime] = getUnixTime2String()
		cm.Data[ConfigmapOperator] = OperatorDelete
	default:
		hwlog.RunLog.Infof("job status update to %s", StatusJobSucceed)
		cm.Data[JobStatus] = StatusJobSucceed
	}
	if _, err = b.clientSet.CoreV1().ConfigMaps(podKeyInfo.namespace).Update(context.TODO(), cm,
		metav1.UpdateOptions{}); err != nil {
		hwlog.RunLog.Errorf("update ConfigMap for Job failed, err %v", err)
		return fmt.Errorf("failed to update ConfigMap for Job %v", err)
	}
	return util.GetAndUpdateCmByTotalNum(cm.Data[cmCutNumKey], cm.Name, cm.Namespace,
		map[string]string{JobStatus: cm.Data[JobStatus]}, b.clientSet)
}

// HandleJobStatus change the job status to what we need
func (b *WorkerInfo) HandleJobStatus(podPhase string) string {
	switch podPhase {
	case PhaseJobSucceed:
		b.succeedPodNum += 1
		if b.succeedPodNum == b.cachedPodNum {
			b.succeedPodNum = 0
			return StatusJobSucceed
		}
		hwlog.RunLog.Debugf("found succeed pod number %d, cached pod number %d", b.succeedPodNum, b.cachedPodNum)
		return PhaseJobRunning
	default:
		b.succeedPodNum = 0
		return StatusJobFail
	}
}

// UpdateJobNodeHealthyStatus update job's node healthy status
func (b *WorkerInfo) UpdateJobNodeHealthyStatus(nodeName string, nodeHealth bool) {
	b.CmMutex.Lock()
	defer b.CmMutex.Unlock()
	b.CMData.SetJobNodeHealthy(nodeName, nodeHealth)
}

// UpdateJobDeviceHealthyStatus update job's device healthy status
func (b *WorkerInfo) UpdateJobDeviceHealthyStatus(nodeName string, networkUnhealthyCards, unHealthyCards string) {
	b.CmMutex.Lock()
	defer b.CmMutex.Unlock()
	b.CMData.SetJobDeviceHealthy(nodeName, networkUnhealthyCards, unHealthyCards)
}

// GetJobHealth get job's healthy status
func (b *WorkerInfo) GetJobHealth() (bool, []string) {
	b.CmMutex.Lock()
	defer b.CmMutex.Unlock()
	return b.CMData.GetJobHealthy()
}

// PGRunning return whether job is running
func (b *WorkerInfo) PGRunning() bool {
	return b.constructionFinished()
}

// GetBaseInfo return base info
func (b *Worker) GetBaseInfo() Info {
	return b.Info
}

// GetDeviceNumPerNode get job use device num per node
func (b *Worker) GetDeviceNumPerNode() int {
	return b.CMData.GetJobDeviceNumPerNode()
}

func isReferenceJobSameWithWorker(pod *apiCoreV1.Pod, jobName string, workerUID string) bool {
	sameWorker := false
	for _, owner := range pod.OwnerReferences {
		if owner.Name == jobName && string(owner.UID) == workerUID {
			sameWorker = true
			break
		}
	}
	return sameWorker
}

func isPodAnnotationsReady(pod *apiCoreV1.Pod, identifier string) bool {
	_, exist := pod.Annotations[PodDeviceKey]
	if !exist {
		hwlog.RunLog.Warnf("syncing '%s' delayed: device info is not ready", identifier)
		return false
	}
	return true
}

func containerUsedChip(pod *apiCoreV1.Pod) bool {
	for _, container := range pod.Spec.Containers {
		if GetNPUNum(container) > 0 {
			return true
		}
	}

	return false
}

// GetNPUNum get npu number
func GetNPUNum(c apiCoreV1.Container) int32 {
	for name, qtt := range c.Resources.Limits {
		if !strings.HasPrefix(string(name), A910ResourceName) {
			continue
		}
		if A800MaxChipNum < qtt.Value() || qtt.Value() < 0 {
			return InvalidNPUNum
		}
		return int32(qtt.Value())
	}
	return 0
}
