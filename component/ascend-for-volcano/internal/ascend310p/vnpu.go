/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package ascend310p is using for HuaWei 310P Ascend pin affinity schedule.
*/
package ascend310p

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

func (tp *ascend310P) GetVNPUTemplate() {
	if tp == nil {
		klog.V(util.LogDebugLev).Infof("GetVNPUTemplate failed:%s", util.ArgumentError)
		return
	}
	tp.vHandle.VT = vnpu.VTemplate{
		Data: tp.FrameAttr.VJobTemplate[plugin.Ascend310P],
	}
}

func (tp *ascend310P) GetPresetVirtualDevices() {
	if tp == nil {
		klog.V(util.LogDebugLev).Infof("GetPresetVirtualDevices failed:%s", util.ArgumentError)
		return
	}
	tp.vHandle.StaticByConf = tp.FrameAttr.CheckVNPUSegmentEnableByConfig()
}

func (tp *ascend310P) InitVNPU() {
	if tp == nil {
		klog.V(util.LogDebugLev).Infof("InitVNPU failed:%s", util.ArgumentError)
		return
	}
	tp.vHandle = &vnpu.VirtualNPU{
		DynamicVNPU: vnpu.DynamicVNPU{
			DowngradeCache: make(map[string][]string, util.MapInitNum),
		},
	}
}

func (tp *ascend310P) checkStVJobReq() error {
	if !tp.vHandle.StaticByConf {
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

func (tp *ascend310P) validStVNPUJob() *api.ValidateResult {
	if reqErr := tp.checkStVJobReq(); reqErr != nil {
		return &api.ValidateResult{Pass: false, Reason: reqErr.Error(), Message: reqErr.Error()}
	}
	return nil
}

func (tp *ascend310P) checkDyVJobReq() error {
	if !tp.IsVJob() {
		return fmt.Errorf("%s not VirtualNPU job", tp.Name)
	}
	if tp.vHandle.StaticByConf {
		return fmt.Errorf("volcano configuration %s true, only support static vnpu", util.SegmentEnable)
	}
	for _, vT := range tp.Tasks {
		if vT.ReqNPUNum == 1 || vT.ReqNPUNum == util.NPUIndex2 || vT.ReqNPUNum == util.NPUIndex4 ||
			vT.ReqNPUNum%util.NPUIndex8 == 0 {
			continue
		}
		return fmt.Errorf("%s req err %d", vT.Name, vT.ReqNPUNum)
	}
	return nil
}

func (tp *ascend310P) validDyVNPUTaskDVPPLabel(vT util.NPUTask) error {
	if !vT.IsVNPUTask() {
		return errors.New("not vNPU task")
	}

	dvppValue := GetVNPUTaskDVPP(vT)

	if vT.ReqNPUNum > 0 && vT.ReqNPUNum%util.NPUIndex8 == 0 {
		return nil
	}
	switch vT.ReqNPUNum {
	case 1, util.NPUIndex2:
		if dvppValue != plugin.AscendDVPPEnabledNull {
			return fmt.Errorf("%s req %d ai-core, but dvpp label is:%s", vT.Name, vT.ReqNPUNum, dvppValue)
		}
	case util.NPUIndex4:
		return nil
	default:
		return fmt.Errorf("err require number:%d", vT.ReqNPUNum)
	}
	return nil
}

// 310P vNPU tasks, must abide by the following conventions:
// 1.vir01: no vpu-dvpp, if has error; vpu-level ignore.
// 2.vir02: no vpu-dvpp, if has error; vpu-level ignore.
// 3.vir04: ignore vpu-level and vpu-dvpp.
// 4.every task must be vNPU Task.
func (tp *ascend310P) validDyVNPUJobLabel() error {
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

func (tp *ascend310P) validDyVNPUJob() *api.ValidateResult {
	if tp.Status == util.PodGroupRunning {
		klog.V(util.LogDebugLev).Infof("%s %s's pg is running", PluginName, tp.ComJob.Name)
		return nil
	}
	if reqErr := tp.checkDyVJobReq(); reqErr != nil {
		return &api.ValidateResult{Pass: false, Reason: reqErr.Error(), Message: reqErr.Error()}
	}
	if labelErr := tp.validDyVNPUJobLabel(); labelErr != nil {
		return &api.ValidateResult{Pass: false, Reason: labelErr.Error(), Message: labelErr.Error()}
	}
	return nil
}

func (tp *ascend310P) getAllDyJobs() map[api.JobID]plugin.SchedulerJob {
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

func (tp *ascend310P) getRestartDyTasksFromJobs(vJobs map[api.JobID]plugin.SchedulerJob,
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

func (tp *ascend310P) getAllNeedRestartDyTasks(ssn *framework.Session) []util.NPUTask {
	vJobs := tp.getAllDyJobs()
	if len(vJobs) == 0 {
		return nil
	}
	return tp.getRestartDyTasksFromJobs(vJobs, ssn)
}

func (tp *ascend310P) deleteDyCutErrTasks(ssn *framework.Session) error {
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
				klog.V(util.LogErrorLev).Infof("GetVTaskUseTemplate %s %s.", vT.Name, getErr)
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
func (tp *ascend310P) initConCache(ssn *framework.Session) error {
	if tp.vHandle == nil {
		return fmt.Errorf("initConCache : %s's vHandle not init", tp.GetPluginName())
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
	tp.vHandle.DynamicVNPU.ConCache = nodes
	return nil
}

func (tp *ascend310P) preStartDyVNPU(ssn *framework.Session) error {
	var reErrors []error

	reErrors = append(reErrors, tp.initConCache(ssn))
	reErrors = append(reErrors, tp.deleteDyCutErrTasks(ssn))

	return util.ConvertErrSliceToError(reErrors)
}

func (tp *ascend310P) preStartVNPU(ssn *framework.Session) error {
	tp.GetVNPUTemplate()
	tp.GetPresetVirtualDevices()
	tp.vHandle.DowngradeCache = make(map[string][]string, util.MapInitNum)
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
