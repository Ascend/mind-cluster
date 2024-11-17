/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.

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
Package plugin is using for HuaWei Ascend pin affinity schedule.
*/
package plugin

import (
	"errors"
	"reflect"
	"strings"

	"gopkg.in/yaml.v2"
	v12 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
	"k8s.io/kube-scheduler/extender/v1"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/conf"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/config"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

// RegisterNPUScheduler register the plugin,like factory.
func (sHandle *ScheduleHandler) RegisterNPUScheduler(name string, pc NPUBuilder) {
	if sHandle == nil || pc == nil {
		klog.V(util.LogInfoLev).Infof("RegisterNPUScheduler : %s.", objectNilError)
		return
	}
	if _, ok := sHandle.NPUPlugins[name]; ok {
		klog.V(util.LogInfoLev).Infof("NPU Scheduler[%s] has been registered before.", name)
		return
	}

	sHandle.NPUPlugins[name] = pc
	klog.V(util.LogInfoLev).Infof("NPU Scheduler[%s] registered.", name)
}

// UnRegisterNPUScheduler unRegister the plugin
func (sHandle *ScheduleHandler) UnRegisterNPUScheduler(name string) error {
	if sHandle == nil {
		return errors.New(util.ArgumentError)
	}
	if _, ok := sHandle.NPUPlugins[name]; ok {
		sHandle.NPUPlugins[name] = nil
		delete(sHandle.NPUPlugins, name)
		klog.V(util.LogErrorLev).Infof("NPU Scheduler[%s] delete.", name)
	}
	klog.V(util.LogDebugLev).Infof("NPU Scheduler[%s] unRegistered.", name)
	return nil
}

// IsPluginRegistered Determine if the plug-in is registered.
func (sHandle *ScheduleHandler) IsPluginRegistered(name string) bool {
	if sHandle == nil {
		klog.V(util.LogErrorLev).Infof("IsPluginRegistered %s", objectNilError)
		return false
	}
	pNames := strings.Split(name, "-")
	if len(pNames) == 0 {
		klog.V(util.LogErrorLev).Infof("IsPluginRegistered %s %#v", name, pNames)
		return false
	}
	if len(pNames) > 1 {
		// vnpu support
		pNames[0] = pNames[0] + "-"
	}
	for k := range sHandle.NPUPlugins {
		if k == pNames[0] {
			return true
		}
	}
	klog.V(util.LogErrorLev).Infof("IsPluginRegistered %s not in NPUPlugins %+v", name, sHandle.NPUPlugins)
	return false
}

// checkSession check the ssn's parameters
func (sHandle *ScheduleHandler) checkSession(ssn *framework.Session) error {
	if sHandle == nil || ssn == nil {
		klog.V(util.LogInfoLev).Infof("%s nil session hence doing nothing.", PluginName)
		return errors.New("nil ssn")
	}
	return nil
}

// InitJobsFromSsn init all jobs in ssn.
func (sHandle *ScheduleHandler) InitJobsFromSsn(ssn *framework.Session) {
	if sHandle == nil || ssn == nil {
		klog.V(util.LogInfoLev).Infof("InitJobsFromSsn failed: %s.", util.ArgumentError)
		return
	}
	sHandle.Jobs = make(map[api.JobID]SchedulerJob, util.MapInitNum)
	tmpJobServerInfos := make(map[api.JobID]struct{})
	for jobID, jobInfo := range ssn.Jobs {
		// mark the job which server list has been recorded in logs
		if _, ok := sHandle.JobSeverInfos[jobID]; ok {
			tmpJobServerInfos[jobID] = struct{}{}
		}

		sJob := SchedulerJob{}
		if err := sJob.Init(jobInfo, sHandle); err != nil {
			if err.Error() != notNPUJobError {
				klog.V(util.LogErrorLev).Infof("%s InitJobsFromSsn failed: %s.", jobInfo.Name, util.SafePrint(err))
			}
			continue
		}
		sHandle.Jobs[jobID] = sJob
	}
	sHandle.JobSeverInfos = tmpJobServerInfos
	return
}

// AddDefaultSchedulerSelectorConfig Merge default and customer custom tags.
func (vf *VolcanoFrame) AddDefaultSchedulerSelectorConfig() {
	if vf == nil {
		klog.V(util.LogInfoLev).Infof("AddDefaultSchedulerSelectorConfig failed: %s.", util.ArgumentError)
		return
	}
	defaultSchedulerConfig := make(map[string]string, util.MapInitNum)
	defaultSchedulerConfig[util.ArchSelector] = util.HuaweiArchArm + "|" + util.HuaweiArchX86
	defaultSchedulerConfig[util.Accelerator] = util.Accelerator910Value + "|" + util.Accelerator310Value
	defaultSchedulerConfig[util.AcceleratorType] = util.CardAcceleratorType + "|" + util.ModuleAcceleratorType +
		"|" + util.ChipAcceleratorType + "|" + util.Card910bx2AcceleratorType + "|" + util.Card910bx2InferAcceleratorType +
		"|" + util.Module910bx16AcceleratorType + "|" + util.Module910bx8AcceleratorType

	defaultCfg := config.Configuration{Name: util.CMSelectorKey, Arguments: defaultSchedulerConfig}

	if len(vf.Confs) == 0 {
		vf.Confs = []config.Configuration{defaultCfg}
		return
	}

	var selectorConf config.Configuration
	var index int
	for idx, selectors := range vf.Confs {
		if selectors.Name == util.CMSelectorKey {
			selectorConf = selectors
			index = idx
			break
		}
	}

	if len(selectorConf.Arguments) == 0 {
		vf.Confs[index] = defaultCfg
		return
	}

	for k, v := range defaultSchedulerConfig {
		confs, ok := selectorConf.Arguments[k]
		if ok {
			selectorConf.Arguments[k] = addConf(confs, v)
			continue
		}
		selectorConf.Arguments[k] = v
	}
	vf.Confs[index] = selectorConf
}

// CheckVNPUSegmentEnableByConfig Check VNPU segmentEnable by init plugin parameters, return true if static
func (vf *VolcanoFrame) CheckVNPUSegmentEnableByConfig() bool {
	if vf == nil {
		klog.V(util.LogDebugLev).Infof("CheckVNPUSegmentEnableByConfig failed: %s.", util.ArgumentError)
		return false
	}
	configuration, err := util.GetConfigFromSchedulerConfigMap(util.CMInitParamKey, vf.Confs)
	if err != nil {
		klog.V(util.LogDebugLev).Info("cannot get configuration, segmentEnable.")
		return false
	}
	// get segmentEnable by user configuration
	segmentEnable, ok := configuration.Arguments[util.SegmentEnable]
	if !ok {
		klog.V(util.LogDebugLev).Info("checkVNPUSegmentEnable doesn't exist presetVirtualDevice.")
		return false
	}
	if segmentEnable == "true" {
		return true
	}
	return false
}

func addConf(configs, value string) string {
	for _, cfg := range strings.Split(value, "|") {
		if !isSelectorContains(configs, cfg) {
			configs += "|" + cfg
		}
	}
	return configs
}

// GetJobTemplate get template of all possible segmentation jobs
func (sHandle *ScheduleHandler) GetJobTemplate() map[string]map[string]util.VResource {
	jobTemplate := map[string]map[string]util.VResource{
		Ascend310P: {
			VNPUTempVir01:        {Aicore: 1, Aicpu: 1, DVPP: AscendDVPPEnabledNull},
			VNPUTempVir02:        {Aicore: util.NPUIndex2, Aicpu: util.NPUIndex2, DVPP: AscendDVPPEnabledNull},
			VNPUTempVir02C1:      {Aicore: util.NPUIndex2, Aicpu: 1, DVPP: AscendDVPPEnabledNull},
			VNPUTempVir04:        {Aicore: util.NPUIndex4, Aicpu: util.NPUIndex4, DVPP: AscendDVPPEnabledNull},
			VNPUTempVir04C3:      {Aicore: util.NPUIndex4, Aicpu: util.NPUIndex3, DVPP: AscendDVPPEnabledNull},
			VNPUTempVir04C3NDVPP: {Aicore: util.NPUIndex4, Aicpu: util.NPUIndex3, DVPP: AscendDVPPEnabledOff},
			VNPUTempVir04C4cDVPP: {Aicore: util.NPUIndex4, Aicpu: util.NPUIndex4, DVPP: AscendDVPPEnabledOn},
		},
		Ascend910: {
			VNPUTempVir02: {Aicore: util.NPUIndex2, Aicpu: 1, DVPP: AscendDVPPEnabledNull},
			VNPUTempVir04: {Aicore: util.NPUIndex4, Aicpu: 1, DVPP: AscendDVPPEnabledNull},
			VNPUTempVir08: {Aicore: util.NPUIndex8, Aicpu: util.NPUIndex3, DVPP: AscendDVPPEnabledNull},
			VNPUTempVir16: {Aicore: util.NPUIndex16, Aicpu: util.NPUIndex7, DVPP: AscendDVPPEnabledNull},
		},
	}
	return jobTemplate
}

// InitVolcanoFrameFromSsn init frame parameter from ssn.
func (sHandle *ScheduleHandler) InitVolcanoFrameFromSsn(ssn *framework.Session) {
	if sHandle == nil || ssn == nil {
		klog.V(util.LogErrorLev).Infof("InitVolcanoFrameFromSsn failed: %s.", util.ArgumentError)
		return
	}
	sHandle.FrameAttr = VolcanoFrame{
		UID:          ssn.UID,
		Confs:        initConfsFromSsn(ssn.Configurations),
		KubeClient:   ssn.KubeClient(),
		VJobTemplate: sHandle.GetJobTemplate(),
	}

	sHandle.FrameAttr.AddDefaultSchedulerSelectorConfig()
}

func initConfsFromSsn(confs []conf.Configuration) []config.Configuration {
	var out []byte
	var err error
	newConfs := make([]config.Configuration, len(confs))
	for idx, cfg := range confs {
		newCfg := &config.Configuration{}
		out, err = yaml.Marshal(cfg)
		if err != nil {
			klog.V(util.LogInfoLev).Infof("Marshal configuration failed: %s.", err)
			continue
		}
		if err = yaml.Unmarshal(out, newCfg); err != nil {
			klog.V(util.LogInfoLev).Infof("Unmarshal configuration failed: %s.", err)
			continue
		}
		newConfs[idx] = *newCfg
	}
	return newConfs
}

// InitJobsPlugin init job by plugins.
func (sHandle *ScheduleHandler) InitJobsPlugin() {
	if sHandle == nil {
		klog.V(util.LogErrorLev).Infof("InitJobsPlugin failed: %s.", util.ArgumentError)
		return
	}
	for _, vcJob := range sHandle.Jobs {
		if vcJob.handler == nil {
			klog.V(util.LogErrorLev).Infof("InitJobsPlugin %s's plugin not register.", vcJob.Name)
			continue
		}
		if err := vcJob.handler.InitMyJobPlugin(vcJob.SchedulerJobAttr, sHandle.ScheduleEnv); err != nil {
			return
		}
	}
}

// InitDeleteJobInfos init empty deleted jobinfos.
func (sHandle *ScheduleHandler) InitDeleteJobInfos() {
	if sHandle == nil {
		klog.V(util.LogErrorLev).Infof("InitDeleteJobInfos failed: %s.", util.ArgumentError)
		return
	}
	sHandle.DeleteJobInfos = map[api.JobID]*api.JobInfo{}
}

// InitCache init ScheduleHandler's cache.
func (sHandle *ScheduleHandler) InitCache() {
	if sHandle == nil {
		klog.V(util.LogInfoLev).Infof("InitCache failed: %s.", util.ArgumentError)
		return
	}
	data := make(map[string]map[string]string, util.MapInitNum)
	data[util.RePropertyCacheName] = make(map[string]string, util.MapInitNum)
	data[util.JobRecovery] = make(map[string]string, util.MapInitNum)
	sHandle.Cache = ScheduleCache{
		Names:           make(map[string]string, util.MapInitNum),
		Namespaces:      make(map[string]string, util.MapInitNum),
		FaultConfigMaps: map[api.JobID]*FaultRankIdData{},
		Data:            data}
}

// PreStartPlugin preStart plugin action.
func (sHandle *ScheduleHandler) PreStartPlugin(ssn *framework.Session) {
	if sHandle == nil || ssn == nil {
		klog.V(util.LogInfoLev).Infof("PreStartPlugin failed: %s.", util.ArgumentError)
		return
	}
	for _, job := range sHandle.Jobs {
		if err := job.handler.PreStartAction(sHandle.BaseHandle.GetReHandle(), ssn); err != nil {
			if strings.Contains(err.Error(), util.ArgumentError) {
				continue
			}
			klog.V(util.LogErrorLev).Infof("PreStartPlugin %s %s.", job.Name, err)
		}
	}
}

func (sHandle *ScheduleHandler) saveCacheToCm() {
	for spName, cmName := range sHandle.ScheduleEnv.Cache.Names {
		nameSpace, okSp := sHandle.ScheduleEnv.Cache.Namespaces[spName]
		data, okData := sHandle.ScheduleEnv.Cache.Data[spName]
		if !okSp || !okData {
			klog.V(util.LogErrorLev).Infof("SaveCacheToCm %s no namespace or Data in cache.", spName)
			continue
		}
		data, err := util.UpdateConfigmapIncrementally(sHandle.FrameAttr.KubeClient, nameSpace, cmName, data)
		if err != nil {
			klog.V(util.LogInfoLev).Infof("get old %s configmap failed: %v, write new data into cm", spName, err)
		}
		var tmpCM = &v12.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      cmName,
				Namespace: nameSpace,
			},
			Data: data,
		}
		if err := util.CreateOrUpdateConfigMap(sHandle.FrameAttr.KubeClient, tmpCM, cmName, nameSpace); err != nil {
			klog.V(util.LogErrorLev).Infof("CreateOrUpdateConfigMap : %s.", util.SafePrint(err))
		}
	}

	for _, faultConfig := range sHandle.ScheduleEnv.Cache.FaultConfigMaps {
		data, err := util.UpdateConfigmapIncrementally(sHandle.FrameAttr.KubeClient, faultConfig.Namespace,
			faultConfig.Name, faultConfig.Data)
		if err != nil {
			klog.V(util.LogInfoLev).Infof("get old %s configmap failed: %v", faultConfig.Name, err)
			continue
		}
		var tmpCM = &v12.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      faultConfig.Name,
				Namespace: faultConfig.Namespace,
			},
			Data: data,
		}
		if err = util.CreateOrUpdateConfigMap(sHandle.FrameAttr.KubeClient, tmpCM, faultConfig.Name,
			faultConfig.Namespace); err != nil {
			klog.V(util.LogErrorLev).Infof("CreateOrUpdateConfigMap : %s.", util.SafePrint(err))
		}
	}
}

// BeforeCloseHandler do the action before ssn close.
func (sHandle *ScheduleHandler) BeforeCloseHandler() {
	if sHandle == nil {
		klog.V(util.LogInfoLev).Infof("BeforeCloseHandler failed: %s.", util.ArgumentError)
		return
	}
	for _, job := range sHandle.Jobs {
		if job.Status == util.PodGroupRunning {
			job.recordTorAffinityJobServerList(sHandle)
		}
	}
	if sHandle.BaseHandle != nil {
		if err := sHandle.BaseHandle.PreStopAction(&sHandle.ScheduleEnv); err != nil {
			klog.V(util.LogErrorLev).Infof("PreStopPlugin  %s.", util.SafePrint(err))
		}
	}
	sHandle.saveCacheToCm()
}

// InitNPUSession init npu plugin and nodes.
func (sHandle *ScheduleHandler) InitNPUSession(ssn *framework.Session) error {
	if sHandle == nil || ssn == nil {
		klog.V(util.LogDebugLev).Infof("InitNPUSession failed: %s.", util.ArgumentError)
		return errors.New(util.ArgumentError)
	}
	klog.V(util.LogDebugLev).Infof("enter %s InitNPUSession.", PluginName)
	defer klog.V(util.LogDebugLev).Infof("leave %s InitNPUSession.", PluginName)

	if err := sHandle.checkSession(ssn); err != nil {
		klog.V(util.LogErrorLev).Infof("%s checkSession : %s.", PluginName, err)
		return err
	}

	sHandle.InitVolcanoFrameFromSsn(ssn)
	sHandle.initCmInformer(ssn)
	sHandle.InitDeleteJobInfos()
	sHandle.InitNodesFromSsn(ssn)
	sHandle.InitJobsFromSsn(ssn)

	sHandle.InitTorNodeInfo(ssn)
	sHandle.InitJobsPlugin()
	sHandle.InitCache()
	sHandle.InitReschedulerFromSsn(ssn)
	if sHandle.BaseHandle != nil {
		sHandle.PreStartPlugin(ssn)
	}
	return nil
}

func (sHandle *ScheduleHandler) initCmInformer(ssn *framework.Session) {
	sHandle.Do(func() {
		informerFactory := informers.NewSharedInformerFactory(ssn.KubeClient(), 0)
		cmInformer := informerFactory.Core().V1().ConfigMaps().Informer()
		cmInformer.AddEventHandler(cache.FilteringResourceEventHandler{
			FilterFunc: util.InformerConfigmapFilter,
			Handler: cache.ResourceEventHandlerFuncs{
				AddFunc: func(obj interface{}) {
					sHandle.UpdateConfigMap(obj, util.AddOperator)
				},
				UpdateFunc: func(oldObj, newObj interface{}) {
					sHandle.UpdateConfigMap(newObj, util.UpdateOperator)
				},
				DeleteFunc: func(obj interface{}) {
					sHandle.UpdateConfigMap(obj, util.DeleteOperator)
				},
			},
		})
		informerFactory.Start(wait.NeverStop)
	})
}

func (sHandle *ScheduleHandler) InitReschedulerFromSsn(ssn *framework.Session) {
	var i interface{}
	if sHandle.BaseHandle == nil {
		return
	}
	if err := sHandle.BaseHandle.InitMyJobPlugin(util.SchedulerJobAttr{}, sHandle.ScheduleEnv); err != nil {
		klog.V(util.LogWarningLev).Infof("InitBasePlugin failed by %s", err)
		return
	}
	if preErr := sHandle.BaseHandle.PreStartAction(i, ssn); preErr != nil {
		klog.V(util.LogWarningLev).Infof("PreStartAction failed by %s", preErr)
		return
	}
}

// AddConfigMap add deviceInfo to cache
func (sHandle *ScheduleHandler) AddConfigMap(obj interface{}) {
	if sHandle == nil {
		klog.V(util.LogDebugLev).Infof("AddConfigMap failed: %s.", util.ArgumentError)
		return
	}
	klog.V(util.LogDebugLev).Infof("Add DeviceInfo to cache")
	sHandle.createOrUpdateDeviceInfo(obj)

}

func (sHandle *ScheduleHandler) UpdateConfigMap(obj interface{}, operator string) {
	if sHandle == nil {
		klog.V(util.LogDebugLev).Infof("UpdateConfigMap failed: %s.", util.ArgumentError)
		return
	}
	klog.V(util.LogDebugLev).Infof("Update DeviceInfo to cache")

	cm, ok := obj.(*v12.ConfigMap)
	if !ok {
		klog.V(util.LogErrorLev).Infof(convertFailedErrorPattern, obj)
		return
	}
	if util.CheckConfigMapIsDeviceInfo(cm) {
		if operator == util.AddOperator || operator == util.UpdateOperator {
			sHandle.createOrUpdateDeviceInfo(cm)
		} else if operator == util.DeleteOperator {
			klog.V(util.LogDebugLev).Infof("Del DeviceInfo to cache")
			nodeName := strings.TrimPrefix(cm.Name, util.DevInfoPreName)
			sHandle.DeviceInfos.Lock()
			delete(sHandle.DeviceInfos.Devices, nodeName)
			sHandle.DeviceInfos.Unlock()
		}
	}
	if util.CheckConfigMapIsNodeInfo(cm) {
		if operator == util.AddOperator || operator == util.UpdateOperator {
			sHandle.createOrUpdateNodeInfo(cm)
		} else if operator == util.DeleteOperator {
			klog.V(util.LogDebugLev).Infof("Del NodeInfo to cache")
			nodeName := strings.TrimPrefix(cm.Name, util.NodeDCmInfoNamePrefix)
			sHandle.NodeInfosFromCm.Lock()
			delete(sHandle.NodeInfosFromCm.Nodes, nodeName)
			sHandle.NodeInfosFromCm.Unlock()
		}

	}
}

// DeleteConfigMap del deviceInfo in cache
func (sHandle *ScheduleHandler) DeleteConfigMap(obj interface{}) {
	if sHandle == nil {
		klog.V(util.LogDebugLev).Infof("DeleteConfigMap failed: %s.", util.ArgumentError)
		return
	}
	klog.V(util.LogDebugLev).Infof("Del DeviceInfo to cache")
	cm, ok := obj.(*v12.ConfigMap)
	if !ok {
		klog.V(util.LogErrorLev).Infof(convertFailedErrorPattern, obj)
		return
	}
	nodeName := strings.TrimPrefix(cm.Name, util.DevInfoPreName)
	sHandle.DeviceInfos.Lock()
	delete(sHandle.DeviceInfos.Devices, nodeName)
	sHandle.DeviceInfos.Unlock()
}

func (sHandle *ScheduleHandler) createOrUpdateDeviceInfo(obj interface{}) {
	cm, ok := obj.(*v12.ConfigMap)
	if !ok {
		klog.V(util.LogErrorLev).Infof(convertFailedErrorPattern, obj)
		return
	}
	deviceInfo, err := getNodeDeviceInfoFromCM(cm)
	if err != nil {
		klog.V(util.LogWarningLev).Infof("get device info failed:%s", err)
		return
	}
	nodeName := strings.TrimPrefix(cm.Name, util.DevInfoPreName)
	sHandle.DeviceInfos.Lock()
	sHandle.DeviceInfos.Devices[nodeName] = deviceInfo
	sHandle.DeviceInfos.Unlock()
}

func (sHandler *ScheduleHandler) createOrUpdateNodeInfo(cm *v12.ConfigMap) {
	nodeInfo, err := getNodeInfoFromCM(cm)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("get node info from configmap %s/%s failed, err: %s", cm.Namespace, cm.Name, err)
		return
	}
	nodeName := strings.TrimPrefix(cm.Name, util.NodeDCmInfoNamePrefix)
	sHandler.NodeInfosFromCm.Lock()
	sHandler.NodeInfosFromCm.Nodes[nodeName] = nodeInfo.NodeInfo
	sHandler.NodeInfosFromCm.Unlock()
}

// GetNPUScheduler get the NPU scheduler by name
func (sHandle *ScheduleHandler) GetNPUScheduler(name string) (ISchedulerPlugin, bool) {
	if sHandle == nil {
		klog.V(util.LogInfoLev).Infof("GetNPUScheduler failed: %s.", util.ArgumentError)
		return nil, false
	}
	pb, found := sHandle.NPUPlugins[name]
	if found && pb != nil {
		return pb(name), found
	}

	return nil, found
}

// BatchNodeOrderFn Score the selected nodes.
func (sHandle *ScheduleHandler) BatchNodeOrderFn(task *api.TaskInfo, nodes []*api.NodeInfo) (map[string]float64, error) {
	if sHandle == nil || task == nil || len(nodes) == 0 {
		klog.V(util.LogDebugLev).Infof("BatchNodeOrderFn failed: %s.", util.ArgumentError)
		return nil, errors.New(util.ArgumentError)
	}
	klog.V(util.LogDebugLev).Infof("Enter batchNodeOrderFn")
	defer klog.V(util.LogDebugLev).Infof("leaving batchNodeOrderFn")

	if !IsNPUTask(task) {
		return nil, nil
	}
	if len(sHandle.Nodes) == 0 {
		klog.V(util.LogDebugLev).Infof("%s batchNodeOrderFn %s.", PluginName, util.ArgumentError)
		return nil, nil
	}
	// init score-map
	var interPodAffinityScore v1.HostPriorityList
	scoreMap := initScoreMap(nodes, interPodAffinityScore)
	vcJob, ok := sHandle.Jobs[task.Job]
	if !ok {
		klog.V(util.LogDebugLev).Infof("BatchNodeOrderFn %s not req npu.", task.Name)
		return scoreMap, nil
	}

	k, ok := vcJob.Label[TorAffinityKey]
	if ok && (k == LargeModelTag || k == NormalSchema) {
		klog.V(util.LogDebugLev).Infof("validNPUJob job is now use tor affinity")
		return sHandle.SetTorAffinityJobNodesScore(task, nodes, vcJob, k, scoreMap)
	}

	// 2.Get the best node and top by A,B,C,D rules and require numbers.
	errGet := vcJob.handler.ScoreBestNPUNodes(task, nodes, scoreMap)
	for nodeName := range scoreMap {
		scoreMap[nodeName] *= scoreWeight
	}
	if errGet != nil {
		// get suitable node failed
		klog.V(util.LogErrorLev).Infof("batchNodeOrderFn task[%s] failed[%s].", task.Name, util.SafePrint(errGet))
		return scoreMap, nil
	}
	klog.V(util.LogInfoLev).Infof("batchNodeOrderFn Get %s for NPU %+v.", task.Name, scoreMap)

	return scoreMap, nil
}

func (sHandle *ScheduleHandler) SetTorAffinityJobNodesScore(task *api.TaskInfo, nodes []*api.NodeInfo,
	vcJob SchedulerJob, label string, scoreMap map[string]float64) (map[string]float64, error) {
	if sHandle == nil || task == nil || len(nodes) == 0 || len(scoreMap) == 0 || !vcJob.JobReadyTag {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogDebugLev).Infof("ScoreBestNPUNodes %s.", err)
		return scoreMap, nil
	}

	result := CheckNetSliceIsMeetJobRequire(vcJob, sHandle, nodes)
	vcJob = sHandle.Jobs[task.Job]
	if result != nil {
		klog.V(util.LogErrorLev).Infof("check job %s tor affinity failed: %s,"+
			"used servers is %s", vcJob.Name, result, vcJob.SelectServers)
		switch label {
		case LargeModelTag:
			vcJob.JobReadyTag = false
		case NormalSchema:
			vcJob.SetNormalJobServerList(sHandle)
		default:
			return scoreMap, nil
		}
		sHandle.Jobs[task.Job] = vcJob
	}
	if errGet := sHandle.ScoreBestNPUNodes(task, nodes, scoreMap); errGet != nil {
		// get suitable node failed
		klog.V(util.LogDebugLev).Infof("batchNodeOrderFn task[%s] failed[%s].", task.Name, util.SafePrint(errGet))
	}
	klog.V(util.LogDebugLev).Infof("batchNodeOrderFn Get %s for NPU %+v.", task.Name, scoreMap)
	return scoreMap, result
}

func (sHandle *ScheduleHandler) ScoreBestNPUNodes(task *api.TaskInfo, nodes []*api.NodeInfo, sMap map[string]float64) error {
	if sHandle == nil || task == nil || len(nodes) == 0 || len(sMap) == 0 {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogErrorLev).Infof("ScoreBestNPUNodes %s.", err)
		return err
	}
	vcjob, ok := sHandle.ScheduleEnv.Jobs[task.Job]
	if !ok {
		return errors.New(util.ArgumentError)
	}
	vcjob.ServerList = vcjob.SortJobServerListBySliceId()
	for nodeName, index := range vcjob.HealthTorRankIndex {
		if index == task.Pod.Annotations[podRankIndex] {
			sMap[nodeName] = maxTorAffinityNodeScore
			return nil
		}
	}

	if len(vcjob.FaultRankIndex) == 0 {
		vcjob.SetJobRankIndex()
	} else {
		vcjob.SetFaultJobRankIndex()
	}

	for _, sl := range vcjob.ServerList {
		if reflect.ValueOf(sl).IsNil() {
			continue
		}
		for _, server := range sl.Servers {
			if reflect.ValueOf(server).IsNil() {
				continue
			}
			for _, node := range nodes {
				if server.Name != node.Name || server.NodeRank != task.Pod.Annotations[podRankIndex] {
					continue
				}
				sMap[server.Name] = maxTorAffinityNodeScore
				break
			}
		}
	}

	klog.V(util.LogInfoLev).Infof("ScoreBestNPUNodes task<%s> sMap<%v>", task.Name, sMap)
	return nil
}
