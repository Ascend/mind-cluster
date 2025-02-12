/*
Copyright(C)2020-2023. Huawei Technologies Co.,Ltd. All rights reserved.

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
Package plugin is using for HuaWei Ascend pin affinity schedule frame.
*/
package plugin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/k8s"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
)

// Determine if the selectors are exactly equal.
func isSelectorContains(defValue, jobValue string) bool {
	for _, v := range strings.Split(defValue, "|") {
		if strings.EqualFold(v, jobValue) {
			return true
		}
	}

	return false
}

// Determine if the two string has same element.
func isEachStringContainsSameElement(first, second, seq string) bool {
	if first == second {
		return true
	}
	fList := strings.Split(first, seq)
	sList := strings.Split(second, seq)
	for _, vFirst := range fList {
		for _, vSecond := range sList {
			if strings.EqualFold(vFirst, vSecond) {
				return true
			}
		}
	}
	return false
}

// GetTaskSelectors get task's selector.
func GetTaskSelectors(task *api.TaskInfo) map[string]string {
	if task == nil {
		klog.V(util.LogErrorLev).Infof("GetTaskSelectors task nil.")
		return nil
	}
	return task.Pod.Spec.NodeSelector
}

// GetTaskLabels get task's Labels.
func GetTaskLabels(task *api.TaskInfo) map[string]string {
	if task == nil {
		klog.V(util.LogErrorLev).Infof("GetTaskLabels task nil.")
		return nil
	}
	return task.Pod.Labels
}

// GetJobSelectorFromVcJob get job selector.
func GetJobSelectorFromVcJob(job *api.JobInfo) map[string]string {
	if job == nil {
		klog.V(util.LogErrorLev).Infof("GetJobSelectorFromVcJob job nil.")
		return nil
	}
	var jobLabel = make(map[string]string, util.MapInitNum)
	for _, task := range job.Tasks {
		taskSelectors := task.Pod.Spec.NodeSelector
		for k, v := range taskSelectors {
			label, ok := jobLabel[k]
			if !ok {
				// no task selector
				jobLabel[k] = v
				continue
			}
			if isSelectorContains(label, v) {
				// has task selector
				continue
			}
			// use '|' to join tasks
			jobLabel[k] = label + "|" + v
		}
	}
	return jobLabel
}

// GetJobLabelFromVcJob get job's label, not task's.
func GetJobLabelFromVcJob(job *api.JobInfo) map[string]string {
	if job == nil {
		klog.V(util.LogErrorLev).Infof("GetJobLabelFromVcJob job nil.")
		return nil
	}
	resLabel := make(map[string]string, util.MapInitNum)
	if job.PodGroup != nil {
		for labelKey, labelValue := range job.PodGroup.Labels {
			resLabel[labelKey] = labelValue
		}
	}

	for _, task := range job.Tasks {
		taskSelector := GetTaskLabels(task)
		for k, v := range taskSelector {
			label, ok := resLabel[k]
			if !ok {
				// no task selector
				resLabel[k] = v
				continue
			}
			if isSelectorContains(label, v) {
				// has task selector
				continue
			}
			// use '|' to join tasks
			resLabel[k] = label + "|" + v
		}
	}
	return resLabel
}

// GetVCJobReqNPUTypeFromJobInfo get job request resource, only NPU.
func GetVCJobReqNPUTypeFromJobInfo(vcJob *api.JobInfo) (string, int, error) {
	if vcJob == nil || vcJob.TotalRequest == nil {
		klog.V(util.LogInfoLev).Infof("GetVCJobReqNPUTypeFromJobInfo nil job's parameter.")
		return "", 0.0, errors.New("nil parameter")
	}

	vcMinResource := getVcjobMinResource(vcJob)
	for k, v := range vcMinResource.ScalarResources {
		// must contain "huawei.com/"
		if strings.Contains(string(k), util.HwPreName) {
			return string(k), int(v / util.NPUHexKilo), nil
		}
	}
	klog.V(util.LogDebugLev).Infof("GetVCJobReqNPUTypeFromJobInfo %+v.", vcMinResource.ScalarResources)
	return "", 0.0, errors.New("nil NPU")
}

func getVcjobMinResource(job *api.JobInfo) *api.Resource {
	if job.PodGroup.Spec.MinResources == nil {
		return api.EmptyResource()
	}
	return api.NewResource(*job.PodGroup.Spec.MinResources)
}

// GetVCTaskReqNPUTypeFromTaskInfo get task request resource, only NPU.
func GetVCTaskReqNPUTypeFromTaskInfo(vcTask *api.TaskInfo) (string, int) {
	if vcTask == nil || vcTask.Resreq == nil {
		klog.V(util.LogInfoLev).Infof("GetVCTaskReqNPUTypeFromTaskInfo nil job's parameter.")
		return "", 0
	}
	for k, v := range vcTask.Resreq.ScalarResources {
		// must contain "huawei.com/"
		if strings.Contains(string(k), util.HwPreName) {
			return string(k), int(v / util.NPUHexKilo)
		}
		continue
	}
	klog.V(util.LogInfoLev).Infof("GetVCTaskReqNPUTypeFromTaskInfo %+v.", vcTask.Resreq.ScalarResources)
	return "", 0
}

// GetJobNPUTasks get NPUTask from jobInfo.
func GetJobNPUTasks(vcJob *api.JobInfo) map[api.TaskID]util.NPUTask {
	if vcJob == nil {
		return nil
	}
	if len(vcJob.Tasks) == 0 {
		klog.V(util.LogDebugLev).Infof("GetJobNPUTasks %s not init has no task.", vcJob.Name)
		return nil
	}
	resultMap := make(map[api.TaskID]util.NPUTask, util.MapInitNum)
	for taskID, taskInf := range vcJob.Tasks {
		initVcJobHcclIndex(taskInf)
		name, num := GetVCTaskReqNPUTypeFromTaskInfo(taskInf)
		resultMap[taskID] = util.NPUTask{
			Name:       taskInf.Name,
			NameSpace:  taskInf.Namespace,
			ReqNPUName: name,
			ReqNPUNum:  num,
			Selector:   GetTaskSelectors(taskInf),
			Label:      GetTaskLabels(taskInf),
			VTask:      &util.VTask{},
			NodeName:   taskInf.NodeName,
			Annotation: taskInf.Pod.Annotations,
			PodStatus:  taskInf.Pod.Status.Phase,
		}
	}
	return resultMap
}

// initSelfPluginByJobInfo init job's handler, the deal plugin.
func (sJob *SchedulerJob) initSelfPluginByJobInfo(sHandle *ScheduleHandler) {
	if sJob == nil {
		return
	}

	pluginName := sJob.getPluginNameByReq()
	if pluginName == "" {
		return
	}

	plugin, ok := sHandle.NPUPlugins[pluginName]
	if !ok {
		return
	}

	sJob.handler = plugin(pluginName)
}

// IsJobInitial Determine if the task is ready.
func initVcJobHcclIndex(taskInf *api.TaskInfo) {
	if taskInf.Pod.Annotations == nil {
		taskInf.Pod.Annotations = make(map[string]string)
	}
	if _, ok := taskInf.Pod.Annotations[podRankIndex]; ok {
		return
	}
	for _, c := range taskInf.Pod.Spec.Containers {
		for _, env := range c.Env {
			if env.Name == vcTaskIndex {
				taskInf.Pod.Annotations[podRankIndex] = env.Value
				return
			}
		}
	}
}

// IsJobInitial Determine if the task is ready.
func IsJobInitial(job *api.JobInfo) bool {
	return job.ValidTaskNum() >= job.MinAvailable && getJobTerminatingPodNum(job) == 0
}

func getJobTerminatingPodNum(job *api.JobInfo) int {
	tNum := 0
	for _, task := range job.Tasks {
		if task.Pod != nil && task.Pod.DeletionTimestamp != nil {
			tNum++
		}
	}
	return tNum
}

// IsJobRestarted used for rescheduling, judge if job restarted
func IsJobRestarted(job *api.JobInfo) bool {
	return IsJobInitial(job) && job.PodGroup.Status.Phase == util.PodGroupRunning
}

// Init the SchedulerJob's init.
func (sJob *SchedulerJob) Init(vcJob *api.JobInfo, sHandle *ScheduleHandler) error {
	if sJob == nil || vcJob == nil {
		klog.V(util.LogErrorLev).Infof("SchedulerJob_Init: parameter is nil.")
		return errors.New("parameter is nil")
	}
	if initErr := sJob.initByJobInfo(vcJob); initErr != nil {
		klog.V(util.LogDebugLev).Infof("%s initByJobInfo %s", vcJob.UID, initErr)
		return initErr
	}

	if !sJob.isJobSupportByPlugin(sHandle) {
		klog.V(util.LogDebugLev).Infof("%s IsJobSupportByPlugin not has suitable plugin.", sJob.Name)
		return fmt.Errorf("%s's plugin not regist", sJob.Name)
	}

	sJob.initSelfPluginByJobInfo(sHandle)
	return nil
}

func (sJob *SchedulerJob) recordTorJobServerList(sHandle *ScheduleHandler) {
	if sJob == nil || sHandle == nil || sHandle.Tors == nil || !sJob.IsTorAffinityJob() {
		return
	}
	if _, found := sHandle.JobSeverInfos[sJob.Name]; found {
		return
	}
	torShareMap := sHandleTorsToTorShareMap(sHandle)
	jobLog := make(map[string]TorShare)
	var nodeJobs []NodeJobInfo
	for ip, v := range torShareMap {
		nodeJobs = []NodeJobInfo{}
		for _, nodeJob := range v.NodeJobs {
			if isContain(sJob.ReferenceName, nodeJob.JobName) {
				nodeJobs = append(nodeJobs, nodeJob)
			}
		}
		if len(nodeJobs) > 0 {
			jobLog[ip] = TorShare{
				IsHealthy:   v.IsHealthy,
				IsSharedTor: v.IsSharedTor,
				NodeJobs:    nodeJobs,
			}
		}
	}
	dataByte, err := json.Marshal(jobLog)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("failed to convert jobLog to dataByte %v", err)
		return
	}
	sHandle.JobSeverInfos[sJob.Name] = struct{}{}
	klog.V(util.LogWarningLev).Infof("record job %s , global tors info  %s", sJob.ReferenceName, string(dataByte))
}

func (sJob *SchedulerJob) updateResetConfigMap(sHandle *ScheduleHandler) {
	if sJob == nil || sHandle == nil {
		return
	}
	if k, ok := sJob.Label[util.SinglePodTag]; !ok || k != util.EnableFunc {
		return
	}
	if _, found := sHandle.JobDeleteFlag[sJob.Name]; found {
		return
	}
	if k, ok := sJob.Label[util.ProcessRecoverEnable]; ok && k == util.EnableFunc {
		return
	}
	cm, err := k8s.GetConfigMapWithRetry(sHandle.FrameAttr.KubeClient, sJob.NameSpace,
		ResetInfoCMNamePrefix+sJob.ReferenceName)
	if err != nil {
		klog.V(util.LogWarningLev).Infof("get reset cm err by:%s", err)
		return
	}
	cmData, ok := cm.Data[ResetInfoCMDataKey]
	if !ok {
		klog.V(util.LogWarningLev).Infof("get reset cm err by %s is not exist", ResetInfoCMDataKey)
		return
	}
	resetCm := TaskResetInfo{}
	err = json.Unmarshal([]byte(cmData), &resetCm)
	if err != nil {
		klog.V(util.LogWarningLev).Infof("get reset cm unmarshal err:%s", err)
		return
	}
	if upErr := updateResetCm(sJob, sHandle.FrameAttr.KubeClient, resetCm,
		sHandle.JobSinglePodFlag[sJob.Name]); upErr != nil {
		klog.V(util.LogWarningLev).Infof("update cm err:%s", upErr)
		return
	}
	sHandle.JobDeleteFlag[sJob.Name] = struct{}{}
}

func updateResetCm(sJob *SchedulerJob, k8sClient kubernetes.Interface, resetCm TaskResetInfo, isSinglePod bool) error {
	resetCm.RankList = []*TaskDevInfo{}
	resetCm.UpdateTime = time.Now().Unix()
	resetCm.RetryTime++
	if !isSinglePod {
		resetCm.RetryTime = 0
	}
	checkCode := util.MakeDataHash(resetCm)
	str, err := json.Marshal(resetCm)
	if err != nil {
		klog.V(util.LogWarningLev).Infof("get reset cm marshal err:%s", err)
		return err
	}
	upCm := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ResetInfoCMNamePrefix + sJob.ReferenceName,
			Namespace: sJob.NameSpace,
			Labels:    map[string]string{"reset": "true"},
		},
		Data: map[string]string{
			util.CmCheckCode:   checkCode,
			ResetInfoCMDataKey: string(str),
			ResetInfoTypeKey:   PodRescheduleRestartType,
		},
	}
	_, err = k8sClient.CoreV1().ConfigMaps(sJob.NameSpace).
		Update(context.TODO(), upCm, metav1.UpdateOptions{})
	if err != nil {
		klog.V(util.LogWarningLev).Infof("set update reset cm err:%s", err)
		return err
	}
	klog.V(util.LogDebugLev).Infof("set update reset cm<%s/%s> success, data: %v", upCm.Namespace, upCm.Name,
		upCm.Data)
	return nil
}

// setJobType get job type, used in vJob temporary.
func (sJob *SchedulerJob) initVTasks(vcJob *api.JobInfo) {
	for tID, t := range vcJob.Tasks {
		tmpTask, ok := sJob.SchedulerJobAttr.NPUJob.Tasks[tID]
		if !ok {
			klog.V(util.LogDebugLev).Infof("%s not in frame tasks.", tID)
			continue
		}
		if initErr := tmpTask.InitVTask(t); initErr != nil {
			klog.V(util.LogErrorLev).Infof("Init vTask %s %s.", tID, initErr)
			continue
		}
		sJob.SchedulerJobAttr.NPUJob.Tasks[tID] = tmpTask
	}
}

// IsTorAffinityJob check job is tor affinity job
func (sJob *SchedulerJob) IsTorAffinityJob() bool {
	if sJob == nil {
		return false
	}
	if k, ok := sJob.Label[TorAffinityKey]; ok && (k == LargeModelTag || k == NormalSchema) {
		return true
	}
	return false
}

// initNPUJob get job type, used in vJob temporary.
func (sJob *SchedulerJob) initNPUJob(vcJob *api.JobInfo) {
	sJob.SetJobType()
	sJob.SetJobStatusByInf(vcJob)
	sJob.initVTasks(vcJob)
	return
}

func (sJob *SchedulerJob) initByJobInfo(vcJob *api.JobInfo) error {
	name, num, err := GetVCJobReqNPUTypeFromJobInfo(vcJob)
	if err != nil {
		return err
	}
	sJob.JobReadyTag = true
	sJob.HealthTorRankIndex = map[string]string{}
	sJob.TorBlackMaps = map[string]struct{}{}
	sJob.UnschedulableReason = UnschedulableReason{
		Reason: map[string]map[string]struct{}{},
		Mutex:  &sync.Mutex{},
	}
	sJob.SchedulerJobAttr.ComJob = util.ComJob{
		Name: vcJob.UID, NameSpace: vcJob.Namespace,
		ReferenceName: util.ReferenceNameOfJob(vcJob),
		Selector:      GetJobSelectorFromVcJob(vcJob),
		Label:         GetJobLabelFromVcJob(vcJob),
		Annotation:    vcJob.PodGroup.Annotations,
	}
	if sJob.Owner.Kind == ReplicaSetType {
		num *= int(*sJob.Owner.Replicas)
		sJob.SchedulerJobAttr.ComJob.Annotation = sJob.Owner.Annotations
	}
	subHealthyStrategy, exist := sJob.Label[util.SubHealthyStrategyLabel]
	if !exist || !util.CheckStrInSlice(subHealthyStrategy,
		[]string{util.SubHealthyIgnore, util.SubHealthyGraceExit, util.SubHealthyForceExit}) {
		subHealthyStrategy = util.SubHealthyIgnore
		klog.V(util.LogDebugLev).Infof("job=%s get label error, use default strategy=%s",
			sJob.Name, subHealthyStrategy)
	}
	sJob.SubHealthyStrategy = subHealthyStrategy
	spBlock := 0
	spBlockStr, ok := sJob.Annotation[util.SuperPodAnnoKey]
	if ok {
		if spBlock, err = strconv.Atoi(spBlockStr); err != nil {
			klog.V(util.LogErrorLev).Infof("get job %s spBlock %s failed %v", vcJob.UID, spBlockStr, err)
		}
	}
	sJob.SchedulerJobAttr.NPUJob = nil
	sJob.handler = nil
	sJob.SchedulerJobAttr.NPUJob = &util.NPUJob{ReqNPUName: name, ReqNPUNum: num, Tasks: GetJobNPUTasks(vcJob),
		VJob: &util.VJob{}, SpBlockNPUNum: spBlock}
	sJob.NPUTaskNum = sJob.GetNPUTaskNumInJob()
	sJob.initNPUJob(vcJob)
	if vcJob.MinAvailable != int32(len(vcJob.Tasks)) {
		sJob.SchedulingTaskNum = defaultSchedulingTaskNum
		return nil
	}
	sJob.SchedulingTaskNum = sJob.GetSchedulingTaskNum()
	return nil
}

// UpdateJobPendingMessage update job pending message
func (sJob *SchedulerJob) UpdateJobPendingMessage(message, nodeName string) {
	if _, ok := sJob.Reason[message]; !ok {
		sJob.Reason[message] = make(map[string]struct{})
	}
	sJob.Reason[message][nodeName] = struct{}{}
}

// IsNPUJob check SchedulerJob is npu job
func (sJob SchedulerJob) IsNPUJob() bool {
	return sJob.handler != nil
}

// ValidJobFn valid job.
func (sJob SchedulerJob) ValidJobFn() *api.ValidateResult {
	if sJob.Owner.Kind == ReplicaSetType {
		if len(sJob.Tasks) < int(*sJob.Owner.Replicas) {
			return &api.ValidateResult{
				Message: fmt.Sprintf("job %s task num %d less than replicas %d", sJob.Name, len(sJob.Tasks), *sJob.Owner.Replicas),
				Reason:  "job is not ready",
				Pass:    false,
			}
		}
		i := 0
		for id := range sJob.Tasks {
			task := sJob.Tasks[id]
			task.Index = i
			sJob.Tasks[id] = task
			i++
		}
	}

	if result := sJob.handler.ValidNPUJob(); result != nil {
		klog.V(util.LogErrorLev).Infof("%s validNPUJob failed:%s.", PluginName, result.Message)
		return result
	}

	klog.V(util.LogInfoLev).Infof("%s valid ok.", sJob.Name)
	return nil
}

// IsJobSinglePodDelete valid job.
func (sJob SchedulerJob) isJobSinglePodRunAsNormal() bool {
	if sJob.SchedulingTaskNum == len(sJob.Tasks) {
		return true
	}
	return false
}

// PreCheckNodePredicate PreCheck Predicate nodes.
func (sJob SchedulerJob) preCheckNodePredicate(taskInfo *api.TaskInfo, vcNode NPUNode) error {
	if nodedEnable := vcNode.Label[util.NodeDEnableKey]; nodedEnable == util.NodeDEnableOnValue {
		nodeHealthyStatusByNodeD := vcNode.Annotation[util.NodedNodeHealtyStatuskey]
		if nodeHealthyStatusByNodeD == util.PreSeparateFaultCode || nodeHealthyStatusByNodeD == util.NodeUnHealthyByNodeD {
			klog.V(util.LogDebugLev).Infof("NodePredicate %s failed, cause node is %s.", vcNode.Name,
				nodeHealthyStatusByNodeD)
			return fmt.Errorf("node is %s, due to nodeD reported node status", nodeHealthyStatusByNodeD)
		}
	}
	klog.V(util.LogDebugLev).Infof("sub-healthy strategy=%s", sJob.SubHealthyStrategy)
	nodeHealthyStatusBySwitch := vcNode.Annotation[util.SwitchNodeHealtyStatuskey]
	if nodeHealthyStatusBySwitch == util.NodeUnHealthyByNodeD {
		klog.V(util.LogDebugLev).Infof("NodePredicate %s failed, cause node is %s by reported switch info.", vcNode.Name,
			nodeHealthyStatusBySwitch)
		return fmt.Errorf("node is %s, due to switch reported node status", nodeHealthyStatusBySwitch)
	}

	if err := vcNode.CheckNPUResourceStable(sJob); err != nil {
		return err
	}
	if err := sJob.CheckNodeNum(taskInfo, vcNode); err != nil {
		return err
	}
	return nil
}

// CheckTorJobSinglePodDeleteV1 valid node.
func (sJob SchedulerJob) CheckTorJobSinglePodDeleteV1(sHandler *ScheduleHandler,
	taskInfo *api.TaskInfo, vcNode NPUNode) error {
	if sJob.isFillJob() {
		return fmt.Errorf("check node err by: large model job can not over tor")
	}
	nodeName, ok := sJob.Annotation[taskInfo.Name]
	if !ok {
		klog.V(util.LogWarningLev).Infof("Cannot get task used fault node name")
		return nil
	}
	faultServer, isTorNode := sHandler.Tors.serverMaps[nodeName]
	if !isTorNode {
		return fmt.Errorf("cannot get task used fault node name")
	}

	server, isTorNode := sHandler.Tors.serverMaps[vcNode.Name]
	if !isTorNode {
		return fmt.Errorf("node is not in tor node list by not get server")
	}

	torIp, getTorIp := sHandler.Tors.torIpMap[vcNode.Name]
	if !getTorIp {
		return fmt.Errorf("node is not in tor node list by not get tor ip")
	}

	tor, isTor := sHandler.Tors.torMaps[torIp]
	if !isTor {
		return fmt.Errorf("node is not in tor node list by not get tor")
	}

	if faultServer.SliceId != server.SliceId || tor.HasAcrossJob(false, sJob.Name) {
		return fmt.Errorf("node sliceId is not meet task require")
	}
	return nil
}

func (sJob SchedulerJob) isPodScheduling() bool {
	return sJob.SchedulingTaskNum != defaultSchedulingTaskNum && sJob.SchedulingTaskNum != 0
}

// CheckTorJobSinglePodDeleteV2 valid node.
func (sJob SchedulerJob) CheckTorJobSinglePodDeleteV2(sHandler *ScheduleHandler, vcNode NPUNode) error {
	if _, ok := sJob.TorBlackMaps[sHandler.Tors.torIpMap[vcNode.Name]]; ok {
		return fmt.Errorf("tor check failed node by nslb2.0")
	}
	return nil
}

// PreCheckTorEnv precheck the env of cluster is ready for tor affinity job
func (sJob SchedulerJob) PreCheckTorEnv(sHandler *ScheduleHandler, nodeMaps map[string]*api.NodeInfo) error {
	if sHandler == nil || sHandler.Tors == nil || sHandler.Tors.Tors == nil {
		return fmt.Errorf("validJobFn [%s] failed:%s", sJob.Name, objectNilError)
	}
	if len(nodeMaps) < sJob.NPUTaskNum {
		return fmt.Errorf("tor check failed not enough resource by "+
			"node num %d is not meet job require %d", len(nodeMaps), sJob.NPUTaskNum)
	}
	return nil
}

func (sJob SchedulerJob) getUsedTorInfos(sHandler *ScheduleHandler) usedTorInfos {
	usedTorInfo := usedTorInfos{
		sharedTorNum:   noneSharedTor,
		isSingleTorJob: false,
		usedTors:       map[string]*Tor{},
	}
	shareTorMap := map[string]struct{}{}
	for _, task := range sJob.Tasks {
		if task.PodStatus != v1.PodRunning {
			continue
		}
		if task.Annotation[isSharedTor] == freeTorAnno {
			usedTorInfo.isSingleTorJob = true
		}
		torIp, ok := sHandler.Tors.torIpMap[task.NodeName]
		if !ok {
			klog.V(util.LogWarningLev).Infof("cannot find tor ip by task %s node name, skip", task.Name)
			continue
		}
		tor := sHandler.Tors.torMaps[torIp]
		usedTorInfo.usedTors[tor.IP] = tor
		if tor.IsSharedTor == sharedTor {
			shareTorMap[tor.IP] = struct{}{}
		}
	}
	usedTorInfo.sharedTorNum = sHandler.getSharedTorNum() - len(shareTorMap)
	return usedTorInfo
}

func (sJob *SchedulerJob) initJobBlackTorMaps(torMaps map[string]*Tor, usedTorInfo usedTorInfos) {
	for _, tor := range torMaps {
		if _, ok := usedTorInfo.usedTors[tor.IP]; ok {
			continue
		}
		if usedTorInfo.isSingleTorJob {
			sJob.TorBlackMaps[tor.IP] = struct{}{}
		}
		if tor.IsSharedTor == exclusiveTor || (tor.IsSharedTor == sharedTor && usedTorInfo.sharedTorNum <= 0) {
			sJob.TorBlackMaps[tor.IP] = struct{}{}
		}
	}
}

func (sJob SchedulerJob) isFillJob() bool {
	return sJob.Label[TorAffinityKey] == LargeModelTag && sJob.NPUTaskNum < fillJobMaxNPUTaskNum
}

func (sJob SchedulerJob) isNormalJob() bool {
	return sJob.Label[TorAffinityKey] == NormalSchema
}

func (sJob SchedulerJob) isLargeModelJob() bool {
	return sJob.Label[TorAffinityKey] == LargeModelTag && sJob.NPUTaskNum >= fillJobMaxNPUTaskNum
}

// SetJobServerList check the single layer tor whether meet the job require and set the job server list
func SetJobServerList(sJob SchedulerJob, sHandler *ScheduleHandler,
	nodeMaps map[string]*api.NodeInfo) error {
	if sJob.ServerList != nil {
		return nil
	}
	if err := sJob.PreCheckTorEnv(sHandler, nodeMaps); err != nil {
		return err
	}
	if sJob.NPUTaskNum > sHandler.Tors.TorCount {
		return fmt.Errorf("job's task number is bigger than torCount")
	}
	sJob.GetEnableServerList(nodeMaps, sHandler)
	_ = sJob.GetFullTorNumFromTorInfo(sHandler)
	n := sJob.NPUTaskNum - len(sJob.HealthTorRankIndex)
	sort.Sort(TorLs(sHandler.Tors.Tors))
	return sJob.SetFillJobServerList(sHandler, sHandler.Tors.Tors, n)
}

// CheckNetSliceIsMeetJobRequire check the net slice is meet the job require in nslb 1.0 and set the job server list
func CheckNetSliceIsMeetJobRequire(sJob SchedulerJob, sHandler *ScheduleHandler,
	nodeMaps map[string]*api.NodeInfo) error {
	if sJob.ServerList != nil {
		return nil
	}
	if err := sJob.PreCheckTorEnv(sHandler, nodeMaps); err != nil {
		return err
	}
	sJob.GetEnableServerList(nodeMaps, sHandler)
	fullTorNum := sJob.GetFullTorNumFromTorInfo(sHandler)
	n := sJob.NPUTaskNum - len(sJob.HealthTorRankIndex)
	sort.Sort(TorLs(sHandler.Tors.Tors))
	netSliceNum := sHandler.Tors.TorCount
	if sJob.NPUTaskNum < netSliceNum {
		if err := sJob.SetFillJobServerList(sHandler, sHandler.Tors.Tors, n); err == nil || sJob.isFillJob() {
			return err
		}
	}
	taskRow, taskColumn := getTaskRowAndTaskColumn(n, netSliceNum)
	if taskRow == -1 {
		return fmt.Errorf("taskRow and taskColumn is illegal")
	}
	if taskRow+1 <= fullTorNum {
		sJob.SetJobServerCacheTosHandler(sHandler, sHandler.Tors.Tors, taskRow, taskColumn)
		sJob.MarkMulJobServerList()
		return nil
	}

	logicList := sJob.GetLogicTorList(sHandler, netSliceNum)
	if logicList == nil {
		return fmt.Errorf("tor check failed logicTorList is nil")
	}
	sort.Sort(LogicTorList(logicList))

	// judge the node num in slice x is enough for job request
	// for example: a job has 22 npu task , netSliceNum is 4. taskRow = 4, taskColumn = 1
	// slice 1 must have 5 nodes
	if len(logicList[taskColumn]) < taskRow+1 {
		return fmt.Errorf("tor check failed not enough resource by netslice <%d> server num <%d> is not enough "+
			"for job require %d", getNetSliceId(logicList[taskColumn]), len(logicList[taskColumn]), taskRow+1)
	}

	if taskRow > 0 && len(logicList[netSliceNum-1]) < taskRow {
		return fmt.Errorf("tor check failed not enough resource by "+
			"logicTor full tor num <%d> is not enough for job require <%d>", len(logicList[netSliceNum-1]), taskRow)
	}

	pyTor, fullTorNum := sJob.GetPhyTosList(sHandler, logicList)

	if taskRow < 1 && taskColumn != netSliceNum-1 {
		err := sJob.SetFillJobServerList(sHandler, pyTor, n)
		sJob.MarkMulJobServerList()
		return err
	}

	sJob.SetJobServerCacheTosHandler(sHandler, pyTor, taskRow, taskColumn)
	sJob.MarkMulJobServerList()
	return nil
}

// setJobAvailableNodes obtain the node that a job can be scheduled in nslb 2.0
func setJobAvailableNodes(sJob *SchedulerJob, sHandler *ScheduleHandler, nodeMaps map[string]*api.NodeInfo) error {
	if err := sJob.PreCheckTorEnv(sHandler, nodeMaps); err != nil {
		return err
	}
	nTaskNum := sJob.NPUTaskNum
	netSliceNum := sHandler.Tors.TorCount
	if nTaskNum < netSliceNum {
		if err := sJob.setNotOverTorCountJobServerList(sHandler, nTaskNum); err == nil || sJob.isFillJob() {
			return err
		}
	}
	if nTaskNum > GetLargeModelMaxServerNum(sHandler.Tors.Tors, sHandler.getSharedTorNum()) {
		if sJob.Label[TorAffinityKey] == LargeModelTag {
			return errors.New("set tor failed by not enough nodes")
		}
		return sJob.setNormalJobServerListV2(sHandler.Tors.Tors, nTaskNum)
	}

	n, tor := sJob.setExclusiveTorAndGetSharedTorNum(GetNotShareAndFreeTorServer(sHandler.Tors.Tors, descOrder), nTaskNum)
	if n == 0 {
		return nil
	}
	sharedTors := GetSharedTorServer(sHandler.Tors.Tors, ascOrder)
	if tor == nil || n <= GetMaxSharedTorServerNum(sharedTors, sHandler.getSharedTorNum()) {
		sJob.setBestSharedTorServer(sharedTors, sHandler.getSharedTorNum(), n)
		return nil
	}
	sJob.setOneTorServer(tor, sharedTor, n)
	return nil
}

// SetJobServerCacheTosHandler set job server list and update the job in sHandler
func (sJob *SchedulerJob) SetJobServerCacheTosHandler(sHandler *ScheduleHandler,
	pyTor []*Tor, taskRow, taskColumn int) {
	if sJob == nil || sHandler == nil || len(pyTor) == 0 {
		klog.V(util.LogDebugLev).Infof("SetJobServerCacheTosHandler failed:%s", util.ArgumentError)
		return
	}
	if taskRow >= len(pyTor) {
		klog.V(util.LogDebugLev).Infof("invalid taskRow: %d, pyTor length: %d", taskRow, len(pyTor))
		return
	}
	tmpTors := copyTorList(pyTor[:taskRow])
	tmpTor := &Tor{}
	tmpTor.Servers = append(tmpTor.Servers, pyTor[taskRow].Servers[:taskColumn+1]...)
	tmpTors = append(tmpTors, tmpTor)
	sJob.ServerList = tmpTors
	sHandler.Jobs[sJob.Name] = *sJob
}

// MarkMulJobServerList mark the job if the server job used is over 1 tor
func (sJob *SchedulerJob) MarkMulJobServerList() {
	if sJob.ServerList == nil {
		return
	}
	for _, tor := range sJob.ServerList {
		if tor.Servers == nil {
			continue
		}
		for _, server := range tor.Servers {
			server.IsUsedByMulJob = true
		}
	}
}

func getTaskRowAndTaskColumn(nTaskNum int, netSliceNum int) (int, int) {
	if netSliceNum == 0 {
		return -1, -1
	}
	taskRow := nTaskNum / netSliceNum
	if nTaskNum%netSliceNum == 0 {
		taskRow = nTaskNum/netSliceNum - 1
	}
	taskColumn := (nTaskNum%netSliceNum + netSliceNum - 1) % netSliceNum
	return taskRow, taskColumn
}

// GetFullTorNumFromTorInfo get the num of full tor
func (sJob *SchedulerJob) GetFullTorNumFromTorInfo(sHandler *ScheduleHandler) int {
	var fullTorNum int
	for _, tor := range sHandler.Tors.Tors {
		count := 0
		for _, l := range tor.Servers {
			if l.CurrentJob != nil && *l.CurrentJob == sJob.Name {
				count++
			}
		}
		if count == sHandler.Tors.TorCount {
			fullTorNum++
		}
		tor.FreeServerCount = count
	}
	return fullTorNum
}

// GetPhyTosList transpose the logic tor list
func (sJob SchedulerJob) GetPhyTosList(sHandler *ScheduleHandler, logicList [][]*Server) ([]*Tor, int) {
	tors := make([]*Tor, 0)
	var fullTor int
	for i := 0; i <= len(logicList[0]); i++ {
		tmpTor := &Tor{}
		for j := 0; j < sHandler.Tors.TorCount; j++ {
			if j >= len(logicList) {
				klog.V(util.LogDebugLev).Infof("invalid j: %d, logicList length: %d", j, len(logicList))
				return tors, fullTor
			}
			if len(logicList[j]) < i+1 {
				break
			}
			tmpTor.Servers = append(tmpTor.Servers, logicList[j][i])
			if j == sHandler.Tors.TorCount-1 {
				fullTor++
			}
		}
		tmpTor.FreeServerCount = len(tmpTor.Servers)
		tors = append(tors, tmpTor)
	}
	return tors, fullTor
}

// SetFillJobServerListV2 set the fill job server list in nslb 2.0
func (sJob *SchedulerJob) SetFillJobServerListV2(Tors []*Tor, taskNum int) error {
	var count int
	for i := 0; i < len(Tors); i++ {
		if Tors[i].FreeServerCount < taskNum {
			continue
		}
		tmpTor := &Tor{}
		for _, k := range Tors[i].Servers {
			if k.CurrentJob != nil && *k.CurrentJob == sJob.Name {
				count++
				tmpTor.Servers = append(tmpTor.Servers, k)
			}
			if count == taskNum {
				break
			}
		}
		sJob.ServerList = append(sJob.ServerList, tmpTor)
		return nil
	}
	return fmt.Errorf("tor check failed not enough resource for fill job")
}

// SetFillJobServerList set the fill job server list in nslb 1.0 and single layer switch networking rule
func (sJob SchedulerJob) SetFillJobServerList(sHandler *ScheduleHandler, Tors []*Tor, taskNum int) error {
	var count int
	for i := len(Tors) - 1; i >= 0; i-- {
		if Tors[i].FreeServerCount < taskNum {
			continue
		}
		tmpTor := &Tor{}
		for _, k := range Tors[i].Servers {
			if k.CurrentJob != nil && *k.CurrentJob == sJob.Name {
				count++
				tmpTor.Servers = append(tmpTor.Servers, k)
			}
			if count == taskNum {
				break
			}
		}
		sJob.ServerList = append(sJob.ServerList, tmpTor)
		sHandler.Jobs[sJob.Name] = sJob
		return nil
	}
	return fmt.Errorf("tor check failed not enough resource for job")
}

// setNotOverTorCountJobServerList set the job server List
// if the task num of the job is ont over the server num of a tor
func (sJob *SchedulerJob) setNotOverTorCountJobServerList(sHandler *ScheduleHandler, nTaskNum int) error {
	var err error
	if err = sJob.SetFillJobServerListV2(GetNotShareTorServer(sHandler.Tors.Tors, ascOrder), nTaskNum); err == nil {
		return nil
	}
	if err = sJob.SetFillJobServerListV2(GetUnhealthyTorServer(sHandler.Tors.Tors, ascOrder), nTaskNum); err == nil {
		return nil
	}
	if err = sJob.SetFillJobServerListV2(GetSharedTorServer(sHandler.Tors.Tors, ascOrder), nTaskNum); err == nil {
		return nil
	}
	if err = sJob.SetFillJobServerListV2(GetNotShareAndFreeTorServer(sHandler.Tors.Tors, ascOrder), nTaskNum); err == nil {
		return nil
	}
	return err
}

func (sJob *SchedulerJob) initJobNodeRankByFaultRank(ri []AllocNodeRankOccurrence, nodeMaps map[string]*api.NodeInfo) {
	for _, r := range ri {
		if r.IsFault || nodeMaps[r.NodeName] == nil {
			continue
		}
		sJob.HealthTorRankIndex[r.NodeName] = r.RankIndex
	}
}

// SetNormalJobServerList set the server list of normal job in nslb 1.0
func (sJob *SchedulerJob) SetNormalJobServerList(sHandler *ScheduleHandler) {
	if sJob == nil || sHandler == nil {
		klog.V(util.LogDebugLev).Infof("SetNormalJobServerList failed:%s", util.ArgumentError)
		return
	}
	sJob.ServerList = []*Tor{}
	var count int
	taskNum := sJob.NPUTaskNum
	for _, tor := range sHandler.Tors.Tors {
		tmpTor := &Tor{}
		tmpTor.IP = tor.IP
		tmpTor.Id = tor.Id
		for _, server := range tor.Servers {
			if server.CurrentJob != nil && *server.CurrentJob == sJob.Name {
				tmpTor.Servers = append(tmpTor.Servers, server)
				count++
			}
			if count != taskNum-len(sJob.HealthTorRankIndex) {
				continue
			}
			sJob.ServerList = append(sJob.ServerList, tmpTor)
			if len(sJob.ServerList) > 1 {
				sJob.MarkMulJobServerList()
			}
			return
		}
		sJob.ServerList = append(sJob.ServerList, tmpTor)
	}
	sJob.JobReadyTag = false
}

// setNormalJobServerListV2 set the normal job server list in nslb 2.0
func (sJob *SchedulerJob) setNormalJobServerListV2(tors []*Tor, nTaskNum int) error {
	sharedTorNum, _ := sJob.setExclusiveTorAndGetSharedTorNum(GetNotShareAndFreeTorServer(tors, descOrder), nTaskNum)
	normalTorNum, tor := sJob.setTorAndGetSharedTorNum(GetUnhealthyTorServer(tors, ascOrder), sharedTor, sharedTorNum)
	if normalTorNum == 0 {
		return nil
	}
	if tor != nil {
		sJob.setOneUnhealthySharedTor(tor, normalTorNum)
		return nil
	}
	n, tmpTor := sJob.setTorAndGetSharedTorNum(GetHealthyTorUsedByNormalJob(tors, ascOrder), sharedTor, normalTorNum)
	if n == 0 {
		return nil
	}
	if tmpTor != nil {
		sJob.setOneUnhealthySharedTor(tmpTor, n)
		return nil
	}
	return errors.New("not enough node for normal job to schedule")
}

// GetLogicTorList get logic tor list by global tor list
func (sJob SchedulerJob) GetLogicTorList(sHandler *ScheduleHandler, netSliceNum int) [][]*Server {
	if netSliceNum > util.MaxSliceNum {
		klog.V(util.LogDebugLev).Infof("GetLogicTorList failed:%s", util.ArgumentError)
		return nil
	}
	logicTorList := make([][]*Server, netSliceNum)
	for _, tor := range sHandler.Tors.Tors {
		for i, server := range tor.Servers {
			if server.CurrentJob == nil || *server.CurrentJob != sJob.Name {
				continue
			}
			if i >= len(logicTorList) {
				klog.V(util.LogDebugLev).Infof("invalid i: %d, logicTorList length: %d", i, len(logicTorList))
			}
			logicTorList[i] = append(logicTorList[i], server)
		}
	}

	return logicTorList

}

// GetEnableServerList get global tor list ,mark the server a job can be scheduled
func (sJob *SchedulerJob) GetEnableServerList(nodes map[string]*api.NodeInfo, sHandler *ScheduleHandler) {
	if sHandler == nil {
		return
	}
	if sHandler.Tors == nil {
		return
	}
	sJob.SelectServers = ""
	sJob.getNormalTorListBeforeRestart(sHandler.Tors.TorCount, nodes)
	sHandler.Jobs[sJob.Name] = *sJob
	delete(sHandler.JobSeverInfos, sJob.Name)

	for _, tor := range sHandler.Tors.Tors {
		count := 0
		tmpName := sJob.Name
		for _, server := range tor.Servers {
			if nodes[server.Name] != nil && sJob.HealthTorRankIndex[server.Name] == "" {
				count++
				server.CurrentJob = &tmpName
				sJob.SelectServers += server.Name + " "
			}
		}
		if tor.HasAcrossJob(false, sJob.Name) && sJob.NPUTaskNum > count {
			tmpName = ""
		}
	}
}

// MarkTorListByJob mark the global tor list by node list a job can be scheduled
func (sJob *SchedulerJob) MarkTorListByJob(nodes map[string]*api.NodeInfo, sHandler *ScheduleHandler) {
	if sHandler == nil || sHandler.Tors == nil {
		return
	}
	for _, tor := range sHandler.Tors.Tors {
		count := 0
		tmpName := sJob.Name
		for _, server := range tor.Servers {
			if nodes[server.Name] != nil {
				count++
				server.CurrentJob = &tmpName
				sJob.SelectServers += server.Name + " "
			}
		}
		tor.FreeServerCount = count
	}
}

func (sJob *SchedulerJob) getNormalTorListBeforeRestart(torCount int, nodes map[string]*api.NodeInfo) {
	if torCount == 0 || len(nodes) == 0 {
		klog.V(util.LogInfoLev).Infof("getNormalTorListBeforeRestart torCount is zero number")
		return
	}
	rts := sJob.getJobsRestartedInfo()
	if rts == nil {
		return
	}
	faultIndexes := setFaultTaskIndex(rts, torCount, nodes)
	if faultIndexes == nil {
		return
	}
	m := make(map[string]string)
	sJob.FaultRankIndex = initJobFaultIndexMaps(faultIndexes, torCount)
	for _, rt := range rts {
		i, err := strconv.Atoi(rt.RankIndex)
		if err != nil {
			klog.V(util.LogInfoLev).Infof("getNormalTorListBeforeRestart change RankIndex to int failed")
			return
		}
		if _, exist := sJob.FaultRankIndex[i]; exist {
			continue
		}
		m[rt.NodeName] = rt.RankIndex
	}
	sJob.HealthTorRankIndex = m
}

// SortJobServerListBySliceId sort JobServer list by SliceId
func (sJob SchedulerJob) SortJobServerListBySliceId() []*Tor {
	for _, tor := range sJob.ServerList {
		sort.Sort(JobServers(tor.Servers))
	}
	return sJob.ServerList
}

// SetJobRankIndex set rank index for job
func (sJob *SchedulerJob) SetJobRankIndex() {
	if sJob == nil {
		klog.V(util.LogDebugLev).Infof("SetJobRankIndex failed:%s", util.ArgumentError)
		return
	}
	var rankIndex int
	for _, tor := range sJob.ServerList {
		for _, server := range tor.Servers {
			if server.NodeRank != "" {
				return
			}
			server.NodeRank = strconv.Itoa(rankIndex)
			rankIndex++
		}
	}
}

// SetFaultJobRankIndex set rank index for fault job's fault task
func (sJob *SchedulerJob) SetFaultJobRankIndex() {
	if sJob == nil {
		klog.V(util.LogDebugLev).Infof("SetJobRankIndex failed:%s", util.ArgumentError)
		return
	}

	indexes := changeJobsIndexMapsIntoSlice(sJob.FaultRankIndex)
	if len(indexes) < sJob.NPUTaskNum-len(sJob.HealthTorRankIndex) {
		return
	}
	var i int
	for _, tor := range sJob.ServerList {
		for _, server := range tor.Servers {
			if server.NodeRank != "" {
				return
			}
			server.NodeRank = strconv.Itoa(indexes[i])
			i++
		}
	}
}

// JobServers job server
type JobServers []*Server

// Len get length
func (s JobServers) Len() int {
	return len(s)
}

// Less define rule
func (s JobServers) Less(i, j int) bool {
	if i > s.Len() || j > s.Len() {
		return false
	}
	count1 := s[i].SliceId
	count2 := s[j].SliceId
	return count1 < count2
}

// Swap swap element
func (s JobServers) Swap(i, j int) {
	if i > s.Len() || j > s.Len() {
		return
	}
	s[i], s[j] = s[j], s[i]
}

// TorLs tor list
type TorLs []*Tor

// Len get length
func (tp TorLs) Len() int {
	return len(tp)
}

// Less define rule
func (tp TorLs) Less(i, j int) bool {
	if i > tp.Len() || j > tp.Len() {
		return false
	}
	count1 := tp[i].FreeServerCount
	count2 := tp[j].FreeServerCount
	return count1 > count2
}

// Swap swap element
func (tp TorLs) Swap(i, j int) {
	if i > tp.Len() || j > tp.Len() {
		return
	}
	tp[i], tp[j] = tp[j], tp[i]
}

// LogicTorList logic tor list
type LogicTorList [][]*Server

// Len get length
func (tp LogicTorList) Len() int {
	return len(tp)
}

// Less define rule
func (tp LogicTorList) Less(i, j int) bool {
	if i > tp.Len() || j > tp.Len() {
		return false
	}
	count1 := len(tp[i])
	count2 := len(tp[j])
	return count1 > count2
}

// Swap swap element
func (tp LogicTorList) Swap(i, j int) {
	if i > tp.Len() || j > tp.Len() {
		return
	}
	tp[i], tp[j] = tp[j], tp[i]
}

func updatePodsPendingReason(job *api.JobInfo, tID api.TaskID, reason string) {
	if tID != "" {
		if t, ok := job.Tasks[tID]; ok {
			updatePodPendingReason(t, reason)
			return
		}
		return
	}

	for _, task := range job.Tasks {
		updatePodPendingReason(task, reason)
	}
}

// SetJobPendingReason set the pod and podGroup pending reason.
func (sHandle *ScheduleHandler) SetJobPendingReason(vcJob *api.JobInfo, reason interface{}) error {
	if sHandle == nil || vcJob == nil {
		klog.V(util.LogErrorLev).Infof("SetJobPendingReason not init jobs.")
		return errors.New(util.ArgumentError)
	}
	var reasonTmp string

	switch value := reason.(type) {
	case string:
		// job failed
		vcJob.JobFitErrors = value
		reasonTmp = value
		// for write pending reason into pod
		updatePodsPendingReason(vcJob, "", reasonTmp)
	case map[api.TaskID]*api.FitErrors:
		vcJob.NodesFitErrors = value
		for tID, nodeErrors := range value {
			// for write pending reason into pod
			updatePodsPendingReason(vcJob, tID, nodeErrors.Error())
			reasonTmp += nodeErrors.Error()
		}
		if sJob, ok := sHandle.Jobs[vcJob.UID]; ok {
			sHandle.RecordJobPendingMessage(sJob)
		}
	default:
		return fmt.Errorf("assert reason(%T) failed", reason)
	}
	// for write pending reason into vcjob
	sHandle.UpdatePodGroupPendingReason(vcJob, reasonTmp)
	return nil
}

// UpdatePodGroupPendingReason update pg
func (sHandle *ScheduleHandler) UpdatePodGroupPendingReason(job *api.JobInfo, reason string) {
	job.JobFitErrors = reason

	if len(job.PodGroup.Status.Conditions) == 0 {
		return
	}

	jc := job.PodGroup.Status.Conditions[0].DeepCopy()
	jc.Type = util.PodGroupUnschedulableType
	jc.Status = v1.ConditionTrue
	jc.LastTransitionTime = metav1.Now()
	jc.TransitionID = string(sHandle.FrameAttr.UID)
	jc.Reason = reason
	jc.Message = reason

	for k, value := range job.PodGroup.Status.Conditions {
		if strings.Contains(value.Message, reason) {
			job.PodGroup.Status.Conditions[k].LastTransitionTime = jc.LastTransitionTime
			job.PodGroup.Status.Conditions[k].TransitionID = jc.TransitionID
			return
		}
	}

	job.PodGroup.Status.Conditions = append(job.PodGroup.Status.Conditions, *jc)
}

// RecordJobPendingMessage record the job pending message to log
func (sHandle *ScheduleHandler) RecordJobPendingMessage(vcJob SchedulerJob) {
	if util.MakeDataHash(sHandle.JobPendingMessage[vcJob.Name]) == util.MakeDataHash(vcJob.Reason) {
		return
	}
	for reason, nodes := range vcJob.Reason {
		nodeNames := ""
		for nodeName := range nodes {
			nodeNames += nodeName + " "
		}
		klog.V(util.LogWarningLev).Infof("job %s schedule failed by:%s node list is %s",
			vcJob.Name, reason, nodeNames)
	}
	sHandle.JobPendingMessage[vcJob.Name] = vcJob.Reason
}

// JobValid the job valid, used by volcano frame.
func (sHandle *ScheduleHandler) JobValid(obj interface{}) *api.ValidateResult {
	klog.V(util.LogInfoLev).Infof("enter job valid")
	defer klog.V(util.LogInfoLev).Infof("leave job valid")

	if sHandle == nil || *sHandle.IsFirstSession {
		return &api.ValidateResult{Pass: false, Reason: objectNilError,
			Message: fmt.Sprintf("validJobFn [%#v] failed:%s", obj, objectNilError)}
	}
	job, ok := obj.(*api.JobInfo)
	if !ok {
		reason := "job convert failed"
		klog.V(util.LogErrorLev).Infof("%s :%#v.", reason, obj)
		return &api.ValidateResult{Pass: false, Reason: reason,
			Message: fmt.Sprintf("validJobFn [%#v] failed:%s", obj, reason)}
	}
	if !IsJobInitial(job) {
		reason := "job is not ready"
		klog.V(util.LogErrorLev).Infof("%s job(%s) not ready:%s.", PluginName, job.Name,
			job.PodGroup.Status.Phase)
		return &api.ValidateResult{Pass: false, Reason: reason,
			Message: fmt.Sprintf("validJobFn [%#v] failed:%s", obj, reason)}
	}
	vcJob, ok := sHandle.Jobs[job.UID]
	if !ok {
		klog.V(util.LogDebugLev).Infof("%s %s not support or init", PluginName, job.Name)
		return nil
	}

	if vcJob.IsTorAffinityJob() {
		if sHandle.Tors == nil {
			reason := "job tor affinity check failed, cluster basic-tor-node-cm is not imported"
			klog.V(util.LogWarningLev).Infof(reason)
			return &api.ValidateResult{Pass: false, Reason: reason,
				Message: fmt.Sprintf("validJobFn [%#v] failed:%s", obj, reason)}
		}
	}
	if k, ok := vcJob.Label[TorAffinityKey]; ok && k != LargeModelTag && k != NormalSchema && k != NullTag {
		reason := fmt.Sprintf("job tor affinity label check failed,tor-affinity label value is %s", k)
		klog.V(util.LogWarningLev).Infof(reason)
		return &api.ValidateResult{Pass: false, Reason: reason,
			Message: fmt.Sprintf("validJobFn [%#v] failed:%s label is %s ", obj, reason, k)}
	}

	result := vcJob.ValidJobFn()
	if result != nil {
		if setErr := sHandle.SetJobPendingReason(job, result.Message); setErr != nil {
			klog.V(util.LogErrorLev).Infof("%s setJobFailed err: %s.", PluginName, util.SafePrint(setErr))
		}
		return result
	}
	return nil
}

// SetJobPendReasonByNodesCase In nodes select case, set node failed and add failed reason.
func (sHandle ScheduleHandler) SetJobPendReasonByNodesCase(job *api.JobInfo) {
	if int32(len(job.Tasks)-len(job.NodesFitErrors)) >= job.MinAvailable {
		klog.V(util.LogDebugLev).Infof("%s not block by nodes(tasks:%d -> jobMin:%d -> nodeErrs:%d).", job.Name,
			len(job.Tasks), job.MinAvailable, len(job.NodesFitErrors))
		return
	}
	if setErr := sHandle.SetJobPendingReason(job, job.NodesFitErrors); setErr != nil {
		klog.V(util.LogErrorLev).Infof("%s setJobFailed err:%s.", PluginName, setErr)
	}
}

// CheckNodeNum Check whether the number of cards on the node meets the task requirements.
func (sJob *SchedulerJob) CheckNodeNum(taskInfo *api.TaskInfo, vcNode NPUNode) error {
	if sJob == nil || taskInfo == nil {
		return errors.New(objectNilError)
	}
	vcTask, ok := sJob.NPUJob.Tasks[taskInfo.UID]
	if !ok {
		klog.V(util.LogErrorLev).Infof("CheckNodeNum %+v.", sJob.SchedulerJobAttr.NPUJob)
		return fmt.Errorf("no %s in SchedulerJob", taskInfo.UID)
	}
	nodeNPUNum, ok := vcNode.Idle[v1.ResourceName(vcTask.ReqNPUName)]
	if !ok {
		return fmt.Errorf("not have %s", vcTask.ReqNPUName)
	}
	if int(nodeNPUNum/util.NPUHexKilo) < vcTask.ReqNPUNum {
		return fmt.Errorf("node not meet task request %s:%d", vcTask.ReqNPUName, vcTask.ReqNPUNum)
	}
	return nil
}

func (sJob SchedulerJob) getPluginNameByReq() string {
	name := sJob.ReqNPUName
	// 1. dynamic vJobs
	if strings.Contains(name, "npu-core") {
		label, ok := sJob.Label[util.JobKindKey]
		if !ok {
			klog.V(util.LogErrorLev).Infof("%s no has %s label in dyCut mode.", sJob.Name, util.JobKindKey)
			return ""
		}
		switch label {
		case util.JobKind910Value, util.JobKind910BValue:
			name = util.NPU910CardName
		case util.JobKind310Value:
			name = util.NPU310CardName
		case util.JobKind310PValue:
			name = util.NPU310PCardName
		default:
			klog.V(util.LogErrorLev).Infof("%s unknown label: %s in dyCut mode.", sJob.Name, label)
			return ""
		}
	}
	// 2. static vJobs
	if strings.HasSuffix(name, "c") {
		nameSplit := strings.Split(name, "-")
		if len(nameSplit) < util.NPUIndex2 {
			return ""
		}
		return nameSplit[0]
	}
	return name
}

// isJobSupportByPlugin judge job whether has it's plugin.
func (sJob SchedulerJob) isJobSupportByPlugin(sHandle *ScheduleHandler) bool {
	name := sJob.getPluginNameByReq()
	if name == "" {
		return false
	}
	return sHandle.IsPluginRegistered(name)
}

// GetAnnoName get job AnnoName, include vNPU job.
func (sJob SchedulerJob) GetAnnoName() (string, error) {
	name := sJob.ReqNPUName
	if strings.Contains(name, "npu-core") {
		_, ok := sJob.Label[util.JobKindKey]
		if !ok {
			klog.V(util.LogErrorLev).Infof("%s no has %s label in dyCut mode.", sJob.Name, util.JobKindKey)
			return "", fmt.Errorf("no %s label in dyCut mode", util.JobKindKey)
		}
		return util.AscendNPUCore, nil
	}
	return sJob.handler.GetAnnoName(), nil
}

func (sJob SchedulerJob) getJobsRestartedInfo() []AllocNodeRankOccurrence {
	ri, ok := sJob.Annotation[JobDeleteFlag]
	if !ok {
		return nil
	}
	var rts []AllocNodeRankOccurrence
	if err := json.Unmarshal([]byte(ri), &rts); err != nil {
		klog.V(util.LogInfoLev).Infof("Unmarshal AllocNodeRankOccurrence failed:%s", util.SafePrint(err))
		return nil
	}
	return rts
}

// setExclusiveTorAndGetSharedTorNum set  exclusive Tor ,get  not scheduled pod num and last not filled tor
func (sJob *SchedulerJob) setExclusiveTorAndGetSharedTorNum(tors []*Tor, serverNum int) (int, *Tor) {
	return sJob.setTorAndGetSharedTorNum(tors, exclusiveTor, serverNum)
}

// setTorAndGetSharedTorNum set tor attr and return not scheduled pod num , last not filled tor
func (sJob *SchedulerJob) setTorAndGetSharedTorNum(tors []*Tor, isShared, serverNum int) (int, *Tor) {
	if len(tors) == 0 {
		return serverNum, nil
	}
	var isHealthyTor int
	for _, tor := range tors {
		if serverNum < tor.FreeServerCount {
			return serverNum, tor
		}
		isHealthyTor = healthyTor
		if isShared == sharedTor {
			isHealthyTor = unhealthyTor
		}
		// if isShared  represent sharedTor in this func
		tor.IsHealthy = isHealthyTor
		sJob.setOneTorServer(tor, isShared, tor.FreeServerCount)
		serverNum -= tor.FreeServerCount
	}
	return serverNum, nil
}

// setBestSharedTorServer get the best shared tor for job ,and set shared tor attr
func (sJob *SchedulerJob) setBestSharedTorServer(tors []*Tor, sharedTorNum, serverNum int) {
	if len(tors) == 0 || (sharedTorNum != oneTor && sharedTorNum != twoTor) {
		return
	}
	if sharedTorNum == 1 {
		sJob.setOneTorServer(getOneSharedTorServer(tors, serverNum), 1, serverNum)
		return
	}
	sJob.setTwoSharedTorServer(tors, serverNum)
}

// setOneTorServer set 1 shared tor attr
func (sJob *SchedulerJob) setOneTorServer(tor *Tor, isShared, serverNum int) {
	if tor == nil || serverNum <= 0 {
		return
	}
	t := initTempTor(tor, isShared, healthyTor)
	tor.IsSharedTor = isShared
	var n int
	for _, sr := range tor.Servers {
		if sr.CurrentJob != nil && *sr.CurrentJob == sJob.Name {
			t.Servers = append(t.Servers, sr)
			n++
		}
		if n == serverNum {
			sJob.ServerList = append(sJob.ServerList, t)
			return
		}
	}
}

// setTwoSharedTorServer set 2 shared tor attr
func (sJob *SchedulerJob) setTwoSharedTorServer(tors []*Tor, serverNum int) {
	if len(tors) == 0 {
		return
	}
	if len(tors) == 1 {
		sJob.setOneTorServer(tors[0], 1, serverNum)
		return
	}
	for i := 0; i < len(tors)-1; i++ {
		if (tors[i].FreeServerCount + tors[i+1].FreeServerCount) < serverNum {
			continue
		}
		sJob.setOneTorServer(tors[i], 1, tors[i].FreeServerCount)
		sJob.setOneTorServer(tors[i+1], 1, serverNum-tors[i].FreeServerCount)
	}
}

func initJobFaultIndexMaps(indexes []int, torCount int) map[int]struct{} {
	tmpFaultRankIndex := make(map[int]struct{}, len(indexes)*torCount)
	for _, index := range indexes {
		for i := index; i < torCount+index; i++ {
			tmpFaultRankIndex[i] = struct{}{}
		}
	}
	return tmpFaultRankIndex
}

func changeJobsIndexMapsIntoSlice(indexes map[int]struct{}) []int {
	var tmpIndexes []int
	for i := range indexes {
		tmpIndexes = append(tmpIndexes, i)
	}
	sort.Ints(tmpIndexes)
	return tmpIndexes
}

func setFaultTaskIndex(rts []AllocNodeRankOccurrence, torCount int, nodes map[string]*api.NodeInfo) []int {
	var faultIndexes []int
	for _, rt := range rts {
		if rt.IsFault || nodes[rt.NodeName] == nil {
			i, err := strconv.Atoi(rt.RankIndex)
			if err != nil {
				klog.V(util.LogInfoLev).Infof("getNormalTorListBeforeRestart change RankIndex to int failed")
				return nil
			}
			faultIndex := i / torCount * torCount
			faultIndexes = append(faultIndexes, faultIndex)
		}
	}
	return faultIndexes
}

// divideTorInfoByJob divide global tors into used tors and not used tor,and get used server num in every used tor
func divideTorInfoByJob(torIpMap map[string]string, torMaps map[string]*Tor,
	ranks []AllocNodeRankOccurrence) jobTorInfos {
	torInfos := initJobTorInfos()
	usedTors := make(map[string]*Tor)
	usedAllTors := make(map[string]*Tor)
	for _, rank := range ranks {
		torIp := torIpMap[rank.NodeName]
		usedAllTors[torIp] = torMaps[torIp]
		if !rank.IsFault {
			torInfos.torNums[torIp]++
			usedTors[torIp] = torMaps[torIp]

		}
	}
	for _, tor := range torMaps {
		if _, ok := usedTors[tor.IP]; !ok {
			torInfos.otherTor = append(torInfos.otherTor, tor)
		}
	}
	torInfos.usedHealthyTor = changeTorMapsToSlice(usedTors)
	torInfos.usedAllTorNum = len(usedAllTors)
	return torInfos
}

func changeTorMapsToSlice(torMaps map[string]*Tor) []*Tor {
	var tors []*Tor
	for _, tor := range torMaps {
		tors = append(tors, tor)
	}
	return tors
}

// setJobNodesAfterRestarted set fault job can be scheduled nodes after reschedule
func (sJob *SchedulerJob) setJobNodesAfterRestarted(sHandler *ScheduleHandler,
	ranks []AllocNodeRankOccurrence, taskNum int) error {
	if sHandler == nil {
		return fmt.Errorf(util.ArgumentError)
	}
	// obtain the tors which job used and not used  before restart
	torsInfos := divideTorInfoByJob(sHandler.Tors.torIpMap, sHandler.Tors.torMaps, ranks)
	// obtain the num of shared tors that Job will use
	sharedNum := getSharedTorNumFromTor(torsInfos.usedHealthyTor)
	if sharedNum > sHandler.getSharedTorNum() {
		return fmt.Errorf("shared tor num is over global shared tor num")
	}
	addSharedTorNum := sHandler.getSharedTorNum() - sharedNum
	serNum := getJobFreeServerNum(sJob.Name, torsInfos.usedHealthyTor)
	// if the tor which job used before restart is enough for job, return
	if taskNum <= serNum {
		sJob.ServerList = copyTorList(torsInfos.usedHealthyTor)
		setUsedTorAttr(sJob.ServerList, torsInfos.torNums, addSharedTorNum)
		return nil
	}

	// if job used tor num is 1, and used tor node num is not enough for job. job end this rescheduling process
	if sJob.isFillJob() || torsInfos.usedAllTorNum == 1 {
		return fmt.Errorf("not enough tor for fill job restart")
	}

	// if otherTor max node is lower than notScheduleNum. job end this rescheduling process
	notScheduleNum := taskNum - serNum
	if GetLargeModelMaxServerNum(torsInfos.otherTor, addSharedTorNum) < notScheduleNum {
		return fmt.Errorf("not enough tor for job restart")
	}

	// add the used tor and mark the used tor attr. if tor is shared, skip it. if tor is freeTor, mark it exclusive
	sJob.ServerList = copyTorList(torsInfos.usedHealthyTor)
	setUsedTorAttr(sJob.ServerList, torsInfos.torNums, noneSharedTor)

	// add exclusive tor from the free tor in the other tor
	n, tor := sJob.setExclusiveTorAndGetSharedTorNum(
		GetNotShareAndFreeTorServer(torsInfos.otherTor, ascOrder), notScheduleNum)
	if n == 0 {
		return nil
	}

	enableSharedTor := GetSharedTorServer(torsInfos.otherTor, ascOrder)
	if tor == nil || n <= GetMaxSharedTorServerNum(enableSharedTor, addSharedTorNum) {
		sJob.setBestSharedTorServer(enableSharedTor, addSharedTorNum, n)
		return nil
	}

	if addSharedTorNum > 0 {
		sJob.setOneTorServer(tor, sharedTor, n)
		return nil
	}
	sJob.setOneTorServer(tor, exclusiveTor, n)
	return nil
}

func (sJob SchedulerJob) preCheckForTorHasAcrossJob(isNSLBv2 bool, jobName api.JobID) bool {
	if sJob.Name == jobName {
		return false
	}
	if sJob.Status != util.PodGroupRunning {
		return false
	}
	if !sJob.isLargeModelJob() && isNSLBv2 {
		return false
	}
	return true
}

func (sJob *SchedulerJob) setBestNodeFromRankIndex(task *api.TaskInfo, sMap map[string]float64) bool {
	sJob.ServerList = sJob.SortJobServerListBySliceId()
	for nodeName, index := range sJob.HealthTorRankIndex {
		if index == task.Pod.Annotations[podRankIndex] {
			sMap[nodeName] = maxTorAffinityNodeScore
			return true
		}
	}
	return false
}

func (sJob *SchedulerJob) setJobFaultRankIndex() {
	// if a job first in scheduling,set node rank for task scheduling
	if len(sJob.FaultRankIndex) == 0 {
		// if a job is nslb2.0 rescheduling job skip set node rank index
		if len(sJob.HealthTorRankIndex) > 0 {
			return
		}
		sJob.SetJobRankIndex()
		return
	}
	sJob.SetFaultJobRankIndex()
}

// setOneUnhealthySharedTor set 1  unhealthy shared tor attr
func (sJob *SchedulerJob) setOneUnhealthySharedTor(tor *Tor, serverNum int) {
	if tor == nil {
		return
	}
	t := initTempTor(tor, sharedTor, unhealthyTor)
	tor.IsSharedTor = sharedTor
	tor.IsHealthy = unhealthyTor
	var n int
	for _, sr := range tor.Servers {
		if sr.CurrentJob != nil && *sr.CurrentJob == sJob.Name {
			t.Servers = append(t.Servers, sr)
			n++
		}
		if n == serverNum {
			sJob.ServerList = append(sJob.ServerList, t)
			return
		}
	}
}

// getSharedTorNumFromTor get a tors`s shared tor num
func getSharedTorNumFromTor(tors []*Tor) int {
	var count int
	for _, tor := range tors {
		if tor.IsSharedTor == sharedTor {
			count++
		}
	}
	return count
}

// setUsedTorAttr set the tor attr used before job restarted
func setUsedTorAttr(tors []*Tor, torNums map[string]int, addSharedTorNum int) {
	for _, tor := range tors {
		if tor.IsSharedTor != freeTor {
			continue
		}
		if tor.FreeServerCount > torNums[tor.IP]+1 && addSharedTorNum != 0 {
			tor.IsSharedTor = sharedTor
			addSharedTorNum--
			continue
		}
		tor.IsSharedTor = exclusiveTor
	}
}

// GetJobInfoAllocatedTaskNum get job allocated task num
func GetJobInfoAllocatedTaskNum(jobInfo *api.JobInfo) int32 {
	allocated := int32(0)
	if jobInfo == nil {
		return allocated
	}
	for _, task := range jobInfo.Tasks {
		if task.NodeName != "" {
			allocated++
		}
	}
	return allocated
}
