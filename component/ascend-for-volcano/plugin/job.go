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
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
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
	for labelKey, labelValue := range job.PodGroup.Labels {
		resLabel[labelKey] = labelValue
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

	for k, v := range getVcjobMinResource(vcJob).ScalarResources {
		// must contain "huawei.com/"
		if strings.Contains(string(k), util.HwPreName) {
			return string(k), int(v / util.NPUHexKilo), nil
		}
	}
	klog.V(util.LogDebugLev).Infof("GetVCJobReqNPUTypeFromJobInfo %+v.", getVcjobMinResource(vcJob).ScalarResources)
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
func IsJobInitial(job *api.JobInfo) bool {
	return job.ValidTaskNum() >= job.MinAvailable
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

func (sJob *SchedulerJob) recordTorAffinityJobServerList(sHandle *ScheduleHandler) {
	if sJob == nil || sHandle == nil || !sJob.IsTorAffinityJob() {
		return
	}

	// the job that has been recorded in the logs should not be recorded again
	if _, found := sHandle.JobSeverInfos[sJob.Name]; found {
		return
	}

	servers := map[string]string{}
	for _, task := range sJob.Tasks {
		if task.NodeName == "" {
			return
		}
		servers[sHandle.Tors.serverIps[task.NodeName]] += task.NodeName + " "
	}
	str, err := json.Marshal(servers)
	if err != nil {
		return
	}
	sHandle.JobSeverInfos[sJob.Name] = struct{}{}
	klog.V(util.LogWarningLev).Infof("record job %s serverList is %s", sJob.Name, string(str))
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
	sJob.JobReadyTag = true
	sJob.HealthTorRankIndex = map[string]string{}
	sJob.SchedulerJobAttr.ComJob = util.ComJob{
		Name: vcJob.UID, NameSpace: vcJob.Namespace,
		ReferenceName: util.ReferenceNameOfJob(vcJob),
		Selector:      GetJobSelectorFromVcJob(vcJob),
		Label:         GetJobLabelFromVcJob(vcJob),
		Annotation:    vcJob.PodGroup.Annotations,
	}
	sJob.SchedulerJobAttr.NPUJob = nil
	sJob.handler = nil
	name, num, err := GetVCJobReqNPUTypeFromJobInfo(vcJob)
	if err != nil {
		return err
	}
	sJob.SchedulerJobAttr.NPUJob = &util.NPUJob{ReqNPUName: name, ReqNPUNum: num, Tasks: GetJobNPUTasks(vcJob),
		VJob: &util.VJob{}}
	sJob.NPUTaskNum = sJob.GetNPUTaskNumInJob()
	sJob.initNPUJob(vcJob)
	return nil
}

// IsNPUJob check SchedulerJob is npu job
func (sJob SchedulerJob) IsNPUJob() bool {
	return sJob.handler != nil
}

// ValidJobSelector validate the job selector.
func (sJob SchedulerJob) ValidJobSelector(vcFrame VolcanoFrame) error {
	if len(sJob.Selector) == 0 || len(vcFrame.Confs) == 0 || len(vcFrame.Confs[0].Arguments) == 0 {
		msg := fmt.Errorf("%s or vcFrame's selectors nil", sJob.Name)
		klog.V(util.LogErrorLev).Infof("%s.", msg.Error())
		return msg
	}

	// check the job selector
	if !util.IsSelectorMeetJob(sJob.Selector, vcFrame.Confs[0].Arguments) {
		meetErr := fmt.Errorf("job(%s) selector:%#v not meet scheduler conf:%#v", sJob.Name, sJob.Selector,
			vcFrame.Confs[0].Arguments)
		klog.V(util.LogErrorLev).Infof(meetErr.Error())
		return meetErr
	}
	return nil
}

func (sJob SchedulerJob) preCheckJob(vcFrame VolcanoFrame) error {
	return sJob.ValidJobSelector(vcFrame)
}

// ValidJobFn valid job.
func (sJob SchedulerJob) ValidJobFn(vcFrame VolcanoFrame) *api.ValidateResult {
	if errPreCheck := sJob.preCheckJob(vcFrame); errPreCheck != nil {
		klog.V(util.LogErrorLev).Infof("%s %s, err: %s.", PluginName, sJob.Name, util.SafePrint(errPreCheck))

		msg := "Job selector error"
		return &api.ValidateResult{
			Pass:    false,
			Reason:  msg,
			Message: fmt.Sprintf("%s", errPreCheck),
		}
	}
	if result := sJob.handler.ValidNPUJob(); result != nil {
		klog.V(util.LogErrorLev).Infof("%s validNPUJob failed:%s.", PluginName, result.Message)
		return result
	}

	klog.V(util.LogInfoLev).Infof("%s valid ok.", sJob.Name)
	return nil
}

func (sJob SchedulerJob) ValidTorInfo(sHandler *ScheduleHandler) error {
	if sHandler == nil || sHandler.Tors == nil || sHandler.Tors.Tors == nil {
		return fmt.Errorf("validJobFn [%s] failed:%s", sJob.Name, objectNilError)
	}
	return nil
}

func CheckNetSliceIsMeetJobRequire(sJob SchedulerJob, sHandler *ScheduleHandler, nodes []*api.NodeInfo) error {
	if sJob.ServerList != nil {
		return nil
	}
	if err := sJob.ValidTorInfo(sHandler); err != nil {
		return err
	}
	if len(nodes) < sJob.NPUTaskNum {
		return fmt.Errorf("tor check failed not enough resource by "+
			"node num %d is not meet job require %d", len(nodes), sJob.NPUTaskNum)
	}
	sJob.GetEnableServerList(nodes, sHandler)
	fullTorNum := sJob.GetFullTorNumFromTorInfo(sHandler)
	n := sJob.NPUTaskNum - len(sJob.HealthTorRankIndex)
	sort.Sort(TorLs(sHandler.Tors.Tors))
	netSliceNum := sHandler.Tors.TorCount
	if sJob.NPUTaskNum < netSliceNum {
		if err := sJob.SetFillJobServerList(sHandler, sHandler.Tors.Tors, n); err == nil ||
			sJob.Label[TorAffinityKey] == LargeModelTag {
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

func (sJob *SchedulerJob) SetJobServerCacheTosHandler(sHandler *ScheduleHandler, pyTor []*Tor, taskRow, taskColumn int) {
	if sJob == nil || sHandler == nil || len(pyTor) == 0 {
		klog.V(util.LogDebugLev).Infof("SetJobServerCacheTosHandler failed:%s", util.ArgumentError)
		return
	}
	if taskRow >= len(pyTor) {
		klog.V(util.LogDebugLev).Infof("invalid taskRow: %d, pyTor length: %d", taskRow, len(pyTor))
		return
	}
	tmpTors := pyTor[:taskRow]
	tmpTor := &Tor{}
	tmpTor.Servers = append(tmpTor.Servers, pyTor[taskRow].Servers[:taskColumn+1]...)
	tmpTors = append(tmpTors, tmpTor)
	sJob.ServerList = make([]*Tor, taskRow+1)
	copy(sJob.ServerList, tmpTors)
	sHandler.Jobs[sJob.Name] = *sJob
}

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

func (sJob *SchedulerJob) GetFullTorNumFromTorInfo(sHandler *ScheduleHandler) int {
	var fullTorNum int
	for _, tor := range sHandler.Tors.Tors {
		count := 0
		for _, l := range tor.Servers {
			if l.CurrentJob == sJob.Name {
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

func (sJob SchedulerJob) GetPhyTosList(sHandler *ScheduleHandler, logicList [][]*Server) ([]*Tor, int) {
	var tors []*Tor
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

func (sJob SchedulerJob) SetFillJobServerList(sHandler *ScheduleHandler, Tors []*Tor, taskNum int) error {
	var count int
	for i := len(Tors) - 1; i >= 0; i-- {
		if Tors[i].FreeServerCount >= taskNum {
			tmpTor := &Tor{}
			for _, k := range Tors[i].Servers {
				if k.CurrentJob == sJob.Name {
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
	}
	return fmt.Errorf("tor check failed not enough resource for fill job")
}

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
			if server.CurrentJob == sJob.Name {
				tmpTor.Servers = append(tmpTor.Servers, server)
				count++
			}
			if count == taskNum {
				sJob.ServerList = append(sJob.ServerList, tmpTor)
				if len(sJob.ServerList) > 1 {
					sJob.MarkMulJobServerList()
				}
				return
			}
		}
		sJob.ServerList = append(sJob.ServerList, tmpTor)
	}
	if len(sJob.ServerList) > 1 {
		sJob.MarkMulJobServerList()
	}
}

func (sJob SchedulerJob) GetLogicTorList(sHandler *ScheduleHandler, netSliceNum int) [][]*Server {
	if netSliceNum > util.MaxSliceNum {
		klog.V(util.LogDebugLev).Infof("GetLogicTorList failed:%s", util.ArgumentError)
		return nil
	}
	logicTorList := make([][]*Server, netSliceNum)
	for _, tor := range sHandler.Tors.Tors {
		for i, server := range tor.Servers {
			if server.CurrentJob == sJob.Name {
				if i >= len(logicTorList) {
					klog.V(util.LogDebugLev).Infof("invalid i: %d, logicTorList length: %d", i, len(logicTorList))
				}
				logicTorList[i] = append(logicTorList[i], server)
			}
		}
	}
	return logicTorList

}

func (sJob *SchedulerJob) GetEnableServerList(nodes []*api.NodeInfo, sHandler *ScheduleHandler) {
	if sHandler == nil {
		return
	}
	if sHandler.Tors == nil {
		return
	}
	sJob.SelectServers = ""
	sJob.getNormalTorListBeforeRestart(sHandler.Tors.TorCount, nodes)
	for _, node := range nodes {
		for _, tor := range sHandler.Tors.Tors {
			if tor.HasAcrossJob() && sJob.NPUTaskNum >= sHandler.Tors.TorCount {
				continue
			}
			for _, server := range tor.Servers {
				if server.Name == node.Name && sJob.HealthTorRankIndex[node.Name] == "" {
					server.CurrentJob = sJob.Name
					sJob.SelectServers += node.Name + " "
				}
			}
		}
	}
}

func (sJob *SchedulerJob) getNormalTorListBeforeRestart(torCount int, nodes []*api.NodeInfo) {
	if torCount == 0 {
		klog.V(util.LogInfoLev).Infof("getNormalTorListBeforeRestart torCount is zero number")
		return
	}
	nrt, ok := sJob.Annotation[JobDeleteFlag]
	if !ok {
		return
	}
	var rts []AllocNodeRankOccurrence
	if err := json.Unmarshal([]byte(nrt), &rts); err != nil {
		klog.V(util.LogInfoLev).Infof("Unmarshal job %s AllocNodeRankOccurrence failed:%s", sJob.Name,
			util.SafePrint(err))
		return
	}
	tmpNodeMaps := util.ChangeNodesToNodeMaps(nodes)
	var faultIndexes []int
	for _, rt := range rts {
		if rt.IsFault || tmpNodeMaps[rt.NodeName] == nil {
			i, err := strconv.Atoi(rt.RankIndex)
			if err != nil {
				klog.V(util.LogInfoLev).Infof("getNormalTorListBeforeRestart change RankIndex to int failed")
				return
			}
			faultIndex := i / torCount * torCount
			faultIndexes = append(faultIndexes, faultIndex)
		}
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

func (sJob SchedulerJob) SortJobServerListBySliceId() []*Tor {
	for _, tor := range sJob.ServerList {
		sort.Sort(JobServers(tor.Servers))
	}
	return sJob.ServerList
}

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

type JobServers []*Server

func (s JobServers) Len() int {
	return len(s)
}

func (s JobServers) Less(i, j int) bool {
	if i > s.Len() || j > s.Len() {
		return false
	}
	count1 := s[i].SliceId
	count2 := s[j].SliceId
	return count1 < count2
}

func (s JobServers) Swap(i, j int) {
	if i > s.Len() || j > s.Len() {
		return
	}
	s[i], s[j] = s[j], s[i]
}

type TorLs []*Tor

func (tp TorLs) Len() int {
	return len(tp)
}

func (tp TorLs) Less(i, j int) bool {
	if i > tp.Len() || j > tp.Len() {
		return false
	}
	count1 := tp[i].FreeServerCount
	count2 := tp[j].FreeServerCount
	return count1 > count2
}

func (tp TorLs) Swap(i, j int) {
	if i > tp.Len() || j > tp.Len() {
		return
	}
	tp[i], tp[j] = tp[j], tp[i]
}

type LogicTorList [][]*Server

func (tp LogicTorList) Len() int {
	return len(tp)
}

func (tp LogicTorList) Less(i, j int) bool {
	if i > tp.Len() || j > tp.Len() {
		return false
	}
	count1 := len(tp[i])
	count2 := len(tp[j])
	return count1 > count2
}

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
	default:
		return fmt.Errorf("assert reason(%T) failed", reason)
	}
	// for write pending reason into vcjob
	sHandle.updatePodGroupPendingReason(vcJob, reasonTmp)
	return nil
}

func (sHandle *ScheduleHandler) updatePodGroupPendingReason(job *api.JobInfo, reason string) {
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

// JobValid the job valid, used by volcano frame.
func (sHandle *ScheduleHandler) JobValid(obj interface{}) *api.ValidateResult {
	klog.V(util.LogInfoLev).Infof("enter job valid")
	defer klog.V(util.LogInfoLev).Infof("leave job valid")

	if sHandle == nil {
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

	result := vcJob.ValidJobFn(sHandle.FrameAttr)
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
		return fmt.Errorf("no %s in SchedulerJob", taskInfo.Name)
	}
	nodeNPUNum, ok := vcNode.Idle[v1.ResourceName(vcTask.ReqNPUName)]
	if !ok {
		return fmt.Errorf("%s not have %s", vcNode.Name, vcTask.ReqNPUName)
	}
	if int(nodeNPUNum/util.NPUHexKilo) < vcTask.ReqNPUNum {
		return fmt.Errorf("%s not meet %s's %s:%d",
			vcNode.Name, vcTask.Name, vcTask.ReqNPUName, vcTask.ReqNPUNum)
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
		case util.JobKind910Value:
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

// GetReqCardNameFromRingController Get request card name from RingController.
func (sJob SchedulerJob) GetReqCardNameFromRingController() string {
	ringType, ok := sJob.Label[util.JobKindKey]
	if !ok {
		return ""
	}
	ringTypeSplit := strings.Split(ringType, "-")
	if len(ringTypeSplit) < util.NPUIndex2 {
		return ""
	}
	return util.NPUCardPreName + ringTypeSplit[util.NPUIndex1]
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
