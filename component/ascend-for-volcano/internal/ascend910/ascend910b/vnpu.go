/*
Copyright(C)2024. Huawei Technologies Co.,Ltd. All rights reserved.

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
Package ascend910b is using for HuaWei pin affinity schedule.
*/
package ascend910b

import (
	"errors"
	"fmt"
	"strings"

	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/vnpu"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

// GetVNPUTemplate get vnpu template
func (tp *Base910b) GetVNPUTemplate() {
	if tp == nil {
		klog.V(util.LogDebugLev).Infof("GetVNPUTemplate failed:%s", util.ArgumentError)
		return
	}
	temp := tp.getEnvTemplate()
	if temp == "" {
		return
	}
	tp.VHandle.VT = vnpu.VTemplate{
		Data: tp.FrameAttr.VJobTemplate[temp],
		Temp: temp,
	}
}

func (tp *Base910b) getEnvTemplate() string {
	if tp == nil {
		klog.V(util.LogDebugLev).Infof("GetVNPUTemplate failed:%s", util.ArgumentError)
		return ""
	}
	for _, node := range tp.Nodes {
		if node.ChipType != "" {
			return node.ChipType
		}
	}
	return ""
}

// GetPresetVirtualDevices get preset virtual devices
func (tp *Base910b) GetPresetVirtualDevices() {
	if tp == nil {
		klog.V(util.LogDebugLev).Infof("GetPresetVirtualDevices failed:%s", util.ArgumentError)
		return
	}
	tp.VHandle.StaticByConf = tp.FrameAttr.CheckVNPUSegmentEnableByConfig()
}

// InitVNPU init map of vnpu tmp
func (tp *Base910b) InitVNPU() {
	if tp == nil {
		klog.V(util.LogDebugLev).Infof("InitVNPU failed:%s", util.ArgumentError)
		return
	}
	tp.VHandle = &vnpu.VirtualNPU{
		DynamicVNPU: vnpu.DynamicVNPU{
			DowngradeCache: make(map[string][]string, util.MapInitNum),
		},
	}
}

func (tp *Base910b) checkStVJobReq() error {
	if !tp.VHandle.StaticByConf {
		return fmt.Errorf("volcano configuration %s false, only support dynamic vnpu", util.SegmentEnable)
	}
	for _, vT := range tp.Tasks {
		if !strings.Contains(vT.ReqNPUName, tp.GetPluginName()) {
			return fmt.Errorf("%s req %s not in template", vT.Name, vT.ReqNPUName)
		}
		if vT.ReqNPUNum != 1 {
			return fmt.Errorf("%s req %d not 1", vT.Name, vT.ReqNPUNum)
		}
	}
	return nil
}

func (tp *Base910b) validStVNPUJob() *api.ValidateResult {
	if reqErr := tp.checkStVJobReq(); reqErr != nil {
		return &api.ValidateResult{Pass: false, Reason: reqErr.Error(), Message: reqErr.Error()}
	}
	return nil
}

func (tp *Base910b) checkDyVJobReq() error {
	if !tp.IsVJob() {
		return fmt.Errorf("%s not VirtualNPU job", tp.Name)
	}
	if tp.VHandle.StaticByConf {
		return fmt.Errorf("volcano configuration %s true, only support static vnpu", util.SegmentEnable)
	}
	for _, vT := range tp.Tasks {
		if tp.checkDyVJobReqByTemp(vT) {
			continue
		}
		return fmt.Errorf("%s req err %d", vT.Name, vT.ReqNPUNum)
	}
	return nil
}

func (tp *Base910b) checkDyVJobReqByTemp(vT util.NPUTask) bool {
	switch tp.VHandle.VT.Temp {
	case plugin.ChipTypeB1:
		return checkB1DyJobRequire(vT)
	case plugin.ChipTypeB2C, plugin.ChipTypeB2:
		return checkB2AndB2CDyJobRequire(vT)
	case plugin.ChipTypeB3:
		return checkB3DyJobRequire(vT)
	case plugin.ChipTypeB4:
		return checkB4DyJobRequire(vT)
	default:
		return true
	}
}

func checkB1DyJobRequire(vT util.NPUTask) bool {
	return vT.ReqNPUNum == util.CoreNum6 || vT.ReqNPUNum == util.CoreNum12 || vT.ReqNPUNum%util.CoreNum25 == 0
}

func checkB2AndB2CDyJobRequire(vT util.NPUTask) bool {
	return vT.ReqNPUNum == util.CoreNum6 || vT.ReqNPUNum == util.CoreNum12 || vT.ReqNPUNum%util.CoreNum24 == 0
}

func checkB3DyJobRequire(vT util.NPUTask) bool {
	return vT.ReqNPUNum == util.CoreNum5 || vT.ReqNPUNum == util.CoreNum10 || vT.ReqNPUNum%util.CoreNum20 == 0
}

func checkB4DyJobRequire(vT util.NPUTask) bool {
	return vT.ReqNPUNum == util.CoreNum5 || vT.ReqNPUNum == util.CoreNum10 || vT.ReqNPUNum%util.CoreNum20 == 0
}

func (tp *Base910b) validDyVNPUTaskDVPPLabel(vT util.NPUTask) error {
	if !vT.IsVNPUTask() {
		return errors.New("not vNPU task")
	}

	dvppValue := GetVNPUTaskDVPP(vT)
	cpuLevel := GetVNPUTaskCpuLevel(vT)
	if !(tp.VHandle.VT.Temp == plugin.ChipTypeB4 && vT.ReqNPUNum == util.NPUIndex10) &&
		cpuLevel != plugin.AscendVNPULevelLow {
		return fmt.Errorf("%s req %d ai-core and npu is %s, but cpu level is:%s", vT.Name, vT.ReqNPUNum,
			tp.VHandle.VT.Temp, cpuLevel)
	}
	if !(tp.VHandle.VT.Temp == plugin.ChipTypeB4 && vT.ReqNPUNum == util.NPUIndex10) &&
		dvppValue != plugin.AscendDVPPEnabledNull {
		return fmt.Errorf("%s req %d ai-core and npu is %s, but dvpp label is:%s", vT.Name, vT.ReqNPUNum,
			tp.VHandle.VT.Temp, dvppValue)
	}
	return nil
}

func (tp *Base910b) validDyVNPUJobLabel() error {
	if !tp.IsVJob() {
		return fmt.Errorf("%s not VirtualNPU job", tp.Name)
	}
	for _, vT := range tp.Tasks {
		if tErr := tp.validDyVNPUTaskDVPPLabel(vT); tErr != nil {
			return tErr
		}
	}
	return nil
}

// ValidDyVNPUJob valid dynamic cut job
func (tp *Base910b) ValidDyVNPUJob() *api.ValidateResult {
	if tp.Status == util.PodGroupRunning {
		klog.V(util.LogDebugLev).Infof("%s's pg is running", tp.ComJob.Name)
		return nil
	}
	// 2.check ring-controller.atlas
	if vErr := tp.CheckJobForm(); vErr != nil {
		klog.V(util.LogErrorLev).Infof("checkJobForm: %s.", vErr)
		return &api.ValidateResult{Pass: false, Reason: vErr.Error(), Message: vErr.Error()}
	}
	if reqErr := tp.checkDyVJobReq(); reqErr != nil {
		return &api.ValidateResult{Pass: false, Reason: reqErr.Error(), Message: reqErr.Error()}
	}
	if labelErr := tp.validDyVNPUJobLabel(); labelErr != nil {
		return &api.ValidateResult{Pass: false, Reason: labelErr.Error(), Message: labelErr.Error()}
	}
	return nil
}

func (tp *Base910b) getAllDyJobs() map[api.JobID]plugin.SchedulerJob {
	jobMap := make(map[api.JobID]plugin.SchedulerJob, util.MapInitNum)
	for jobID, vJob := range tp.Jobs {
		if vJob.VJob == nil {
			continue
		}
		if vJob.VJob.Type == util.JobTypeDyCut {
			jobMap[jobID] = vJob
		}
	}
	return jobMap
}

func getFailedDyTasksFromJobs(vJobs map[api.JobID]plugin.SchedulerJob) map[api.TaskID]util.NPUTask {
	vTasks := make(map[api.TaskID]util.NPUTask, util.MapInitNum)
	for _, vJob := range vJobs {
		for tID, vTask := range vJob.Tasks {
			if vTask.Status == util.TaskStatusAllocate || vTask.Status == util.TaskStatusFailed {
				vTasks[tID] = vTask
			}
		}
	}
	return vTasks
}

func getDyFailedNamespaces(vT map[api.TaskID]util.NPUTask) map[string]struct{} {
	nsMap := make(map[string]struct{}, util.MapInitNum)
	for _, nT := range vT {
		nsMap[nT.NameSpace] = struct{}{}
	}
	return nsMap
}

func getAllDyFailedTasks(ssn *framework.Session, nsMap map[string]struct{}) []api.TaskID {
	var tIDs []api.TaskID
	for ns := range nsMap {
		tmp := vnpu.GetSegmentFailureTaskIDs(ssn, ns)
		if len(tmp) == 0 {
			continue
		}
		tIDs = append(tIDs, tmp...)
	}
	return tIDs
}

func getDyFailedTaskIDsInFaileds(allIDS []api.TaskID, vT map[api.TaskID]util.NPUTask) []api.TaskID {
	var tIDs []api.TaskID
	for _, tID := range allIDS {
		if _, ok := vT[tID]; !ok {
			klog.V(util.LogErrorLev).Infof("getDyFailedTaskIDsInFaileds taskID(%s) not in tasks.", tID)
			continue
		}
		tIDs = append(tIDs, tID)
	}
	return tIDs
}

func getDyFailedTasksFromFailed(ssn *framework.Session, vT map[api.TaskID]util.NPUTask) []api.TaskID {
	if len(vT) == 0 {
		return nil
	}
	nsMap := getDyFailedNamespaces(vT)

	allIDS := getAllDyFailedTasks(ssn, nsMap)

	return getDyFailedTaskIDsInFaileds(allIDS, vT)
}

func (tp *Base910b) getRestartDyTasksFromJobs(vJobs map[api.JobID]plugin.SchedulerJob,
	ssn *framework.Session) []util.NPUTask {
	vTasks := getFailedDyTasksFromJobs(vJobs)
	fTIDs := getDyFailedTasksFromFailed(ssn, vTasks)
	if len(fTIDs) == 0 {
		return nil
	}
	var nSlice []util.NPUTask
	for _, tID := range fTIDs {
		vT, ok := vTasks[tID]
		if !ok {
			klog.V(util.LogErrorLev).Infof("getRestartDyTasksFromJobs taskID(%s) not found.", tID)
			continue
		}
		nSlice = append(nSlice, vT)
	}
	return nSlice
}

func (tp *Base910b) getAllNeedRestartDyTasks(ssn *framework.Session) []util.NPUTask {
	vJobs := tp.getAllDyJobs()
	if len(vJobs) == 0 {
		return nil
	}
	return tp.getRestartDyTasksFromJobs(vJobs, ssn)
}

func (tp *Base910b) deleteDyCutErrTasks(ssn *framework.Session) error {
	nTasks := tp.getAllNeedRestartDyTasks(ssn)
	if len(nTasks) == 0 {
		return nil
	}
	for _, nT := range nTasks {
		if nT.VTask == nil {
			klog.V(util.LogErrorLev).Infof("deleteDyCutErrTasks vTask %s is nil.", nT.Name)
			continue
		}
		if delErr := nT.ForceDeletePodByTaskInf(ssn, vnpu.DyCutFailedError, nT.VTask.Allocated.NodeName); delErr != nil {
			klog.V(util.LogErrorLev).Infof("ForceDeletePodByTaskInf %s: %s.", nT.Name, delErr)
		}
	}
	return nil
}

func initDyCutConCacheByJobInfo(nodes map[string]map[string]map[api.TaskID]struct{}, jobInf *api.JobInfo,
	vJob plugin.SchedulerJob) error {
	if jobInf == nil {
		return fmt.Errorf("initDyCutConCacheByJobInfo :%s", util.ArgumentError)
	}
	for taskID, vT := range vJob.Tasks {
		if !vT.IsNPUTask() {
			continue
		}
		if vT.Status == util.TaskStatusAllocate {
			taskInfo, taskOK := jobInf.Tasks[taskID]
			if !taskOK {
				klog.V(util.LogErrorLev).Infof("initConCache %s not in job.", vT.Name)
				continue
			}
			template, getErr := util.GetVTaskUseTemplate(taskInfo)
			if getErr != nil {
				klog.V(util.LogDebugLev).Infof("GetVTaskUseTemplate %s %s.", vT.Name, getErr)
				continue
			}
			initConcacheByTemplate(nodes, vT, template, taskID)
		}
	}
	return nil
}

func initConcacheByTemplate(nodes map[string]map[string]map[api.TaskID]struct{}, vT util.NPUTask,
	template string, taskID api.TaskID) {
	if nodes == nil {
		return
	}
	if vT.Allocated.NodeName != "" {
		templates, nodeOk := nodes[vT.Allocated.NodeName]
		if !nodeOk {
			templates = make(map[string]map[api.TaskID]struct{}, util.MapInitNum)
		}
		tasks, ok := templates[template]
		if !ok {
			tasks = make(map[api.TaskID]struct{}, util.MapInitNum)
		}
		tasks[taskID] = struct{}{}
		templates[template] = tasks
		nodes[vT.Allocated.NodeName] = templates
	}
}

// ConCache format nodeName: templateName:taskUID
func (tp *Base910b) initConCache(ssn *framework.Session) error {
	if tp.VHandle == nil {
		return fmt.Errorf("initConCache : %s's VHandle not init", tp.GetPluginName())
	}

	nodes := make(map[string]map[string]map[api.TaskID]struct{}, util.MapInitNum)
	for jobID, vJob := range tp.Jobs {
		jobInf, jobOk := ssn.Jobs[jobID]
		if !jobOk {
			klog.V(util.LogErrorLev).Infof("initConCache %s not in ssn.", jobID)
			continue
		}
		if initErr := initDyCutConCacheByJobInfo(nodes, jobInf, vJob); initErr != nil {
			continue
		}
	}
	tp.VHandle.DynamicVNPU.ConCache = nodes
	return nil
}

func (tp *Base910b) preStartDyVNPU(ssn *framework.Session) error {
	var reErrors []error

	reErrors = append(reErrors, tp.initConCache(ssn))
	reErrors = append(reErrors, tp.deleteDyCutErrTasks(ssn))

	return util.ConvertErrSliceToError(reErrors)
}

// PreStartVNPU do something before schedule for vnpu
func (tp *Base910b) PreStartVNPU(ssn *framework.Session) error {
	tp.GetVNPUTemplate()
	tp.GetPresetVirtualDevices()
	tp.VHandle.DowngradeCache = make(map[string][]string, util.MapInitNum)
	return tp.preStartDyVNPU(ssn)
}

// GetVNPUTaskDVPP dvpp default is null
func GetVNPUTaskDVPP(asTask util.NPUTask) string {
	value, ok := asTask.Label[plugin.AscendVNPUDVPP]
	if !ok {
		value = plugin.AscendDVPPEnabledNull
	}
	return value
}

// GetVNPUTaskCpuLevel cpu default is null
func GetVNPUTaskCpuLevel(asTask util.NPUTask) string {
	value, ok := asTask.Label[plugin.AscendVNPULevel]
	if !ok {
		value = plugin.AscendVNPULevelLow
	}
	return value
}
