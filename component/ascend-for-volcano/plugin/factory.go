/*
Copyright(C)2020-2024. Huawei Technologies Co.,Ltd. All rights reserved.

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
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"gopkg.in/yaml.v2"
	v12 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
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
	oldJobs := sHandle.Jobs
	sHandle.Jobs = make(map[api.JobID]SchedulerJob, util.MapInitNum)
	tmpJobServerInfos := make(map[api.JobID]struct{})
	tmpJobDeleteFlags := make(map[api.JobID]struct{})
	tmpJobSinglePodFlag := make(map[api.JobID]bool)
	tmpJobPendingMessage := make(map[api.JobID]map[string]map[string]struct{})
	for jobID, jobInfo := range ssn.Jobs {
		sJob := SchedulerJob{}
		if err := sJob.Init(jobInfo, sHandle); err != nil {
			klog.V(util.LogDebugLev).Infof("%s InitJobsFromSsn failed: %s.", jobInfo.Name, util.SafePrint(err))
			continue
		}
		if oldJob, ok := oldJobs[jobID]; ok {
			sJob.SuperPods = oldJob.SuperPods
		}
		sHandle.Jobs[jobID] = sJob
		// mark the job which server list has been recorded in logs
		if _, ok := sHandle.JobSeverInfos[jobID]; ok && sJob.Status == util.PodGroupRunning {
			tmpJobServerInfos[jobID] = struct{}{}
		}
		// mark the job which reset configmap has been set
		if _, ok := sHandle.JobDeleteFlag[jobID]; ok && sJob.SchedulingTaskNum == 0 {
			tmpJobDeleteFlags[jobID] = struct{}{}
		}
		tmpJobSinglePodFlag[jobID] = sHandle.JobSinglePodFlag[jobID]
		if sJob.isPodScheduling() {
			tmpJobSinglePodFlag[jobID] = sJob.SchedulingTaskNum != len(sJob.Tasks)
		}
		tmpJobPendingMessage[jobID] = sHandle.JobPendingMessage[jobID]
	}
	sHandle.JobSeverInfos = tmpJobServerInfos
	sHandle.JobDeleteFlag = tmpJobDeleteFlags
	sHandle.JobSinglePodFlag = tmpJobSinglePodFlag
	sHandle.JobPendingMessage = tmpJobPendingMessage
	return
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
		ChipTypeB1: {
			VNPUTempVir06: {Aicore: util.NPUIndex6, Aicpu: util.NPUIndex1, DVPP: AscendDVPPEnabledNull},
			VNPUTempVir03: {Aicore: util.NPUIndex3, Aicpu: util.NPUIndex1, DVPP: AscendDVPPEnabledNull},
			VNPUTempVir12: {Aicore: util.NPUIndex12, Aicpu: util.NPUIndex3, DVPP: AscendDVPPEnabledNull},
		},
		ChipTypeB2C: {
			VNPUTempVir06: {Aicore: util.NPUIndex6, Aicpu: util.NPUIndex1, DVPP: AscendDVPPEnabledNull},
			VNPUTempVir03: {Aicore: util.NPUIndex3, Aicpu: util.NPUIndex1, DVPP: AscendDVPPEnabledNull},
			VNPUTempVir12: {Aicore: util.NPUIndex12, Aicpu: util.NPUIndex3, DVPP: AscendDVPPEnabledNull},
		},
		ChipTypeB3: {
			VNPUTempVir05: {Aicore: util.NPUIndex5, Aicpu: util.NPUIndex1, DVPP: AscendDVPPEnabledNull},
			VNPUTempVir10: {Aicore: util.NPUIndex10, Aicpu: util.NPUIndex3, DVPP: AscendDVPPEnabledNull},
		},
		ChipTypeB4: {
			VNPUB4TempVir05:     {Aicore: util.NPUIndex5, Aicpu: util.NPUIndex1, DVPP: AscendDVPPEnabledNull},
			VNPUB4TempVir10C3NM: {Aicore: util.NPUIndex10, Aicpu: util.NPUIndex3, DVPP: AscendDVPPEnabledOff},
			VNPUB4TempVir10C4M:  {Aicore: util.NPUIndex10, Aicpu: util.NPUIndex4, DVPP: AscendDVPPEnabledOn},
			VNPUB4TempVir10:     {Aicore: util.NPUIndex10, Aicpu: util.NPUIndex3, DVPP: AscendDVPPEnabledNull},
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

	configs := initConfsFromSsn(ssn.Configurations)
	superPodSize, err := util.GetSizeOfSuperPod(configs)
	if err != nil {
		klog.V(util.LogWarningLev).Infof("GetSizeOfSuperPod failed: %s, set default super-pod-size: %d", err,
			defaultSuperPodSize)
		superPodSize = defaultSuperPodSize
	}

	if superPodSize == 0 {
		klog.V(util.LogWarningLev).Infof(" super-pod-size configuration should be a number bigger than 0, "+
			"set default super-pod-size: %d", defaultSuperPodSize)
		superPodSize = defaultSuperPodSize
	}

	reserve, err := util.GetReserveNodes(configs)
	if err != nil {
		klog.V(util.LogWarningLev).Infof("GetReserveNodes failed: %s, set default reserve-nodes: %d", err,
			defaultReserveNodes)
		reserve = defaultReserveNodes
	}
	if reserve >= superPodSize {
		validRes := 0
		if superPodSize > defaultReserveNodes {
			validRes = defaultReserveNodes
		}
		klog.V(util.LogWarningLev).Infof("reserve-nodes(%d) is larger than super-pod-size(%d), set reserve-nodes: %d",
			reserve, superPodSize, validRes)
		reserve = validRes
	}

	sHandle.FrameAttr = VolcanoFrame{
		UID:            ssn.UID,
		Confs:          configs,
		KubeClient:     ssn.KubeClient(),
		VJobTemplate:   sHandle.GetJobTemplate(),
		SuperPodSize:   superPodSize,
		ReservePodSize: reserve,
	}
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
		if job.SchedulingTaskNum == 0 {
			job.recordTorJobServerList(sHandle)
			job.updateResetConfigMap(sHandle)
		}
	}
	if sHandle.BaseHandle != nil {
		if err := sHandle.BaseHandle.PreStopAction(&sHandle.ScheduleEnv); err != nil {
			klog.V(util.LogErrorLev).Infof("PreStopPlugin  %s.", util.SafePrint(err))
		}
	}

	sHandle.saveCacheToCm()
	if sHandle.Tors == nil || sHandle.getNSLBVsersion() == defaultNSLBVersion {
		return
	}
	err := sHandle.CacheToShareCM()
	if err != nil {
		klog.V(util.LogErrorLev).Infof("CacheToShareCM error: %v", err)
	}
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

	if sHandle.Tors == nil || sHandle.getNSLBVsersion() == defaultNSLBVersion {
		return nil
	}
	klog.V(util.LogInfoLev).Infof("InitNSLB2.0")
	sHandle.InitNSLB2(ssn)
	return nil
}

// initCmInformer init cm informer, support cluster info manager and device plugin
func (sHandle *ScheduleHandler) initCmInformer(ssn *framework.Session) {
	sHandle.Do(func() {
		if sHandle.FrameAttr.CheckUseCIMByConfig() {
			sHandle.initClusterCmInformer(ssn)
			if !util.ClusterDDeploymentIsExist(ssn.KubeClient()) {
				klog.V(util.LogErrorLev).Info("ClusterD deployment is not exist， please apply ClusterD")
			}
			return
		}
		sHandle.initDeviceAndNodeDCmInformer(ssn)
	})
}

func (sHandle *ScheduleHandler) initClusterCmInformer(ssn *framework.Session) {
	informerFactory := informers.NewSharedInformerFactoryWithOptions(ssn.KubeClient(), 0,
		informers.WithNamespace(util.MindXDlNameSpace),
		informers.WithTweakListOptions(func(options *metav1.ListOptions) {
			options.LabelSelector = util.CmConsumer + "=" + util.CmConsumerValue
		}))
	cmInformer := informerFactory.Core().V1().ConfigMaps().Informer()
	cmInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			sHandle.updateConfigMapCluster(obj, util.AddOperator)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			sHandle.updateConfigMapCluster(newObj, util.UpdateOperator)
		},
		DeleteFunc: func(obj interface{}) {
			sHandle.updateConfigMapCluster(obj, util.DeleteOperator)
		},
	})
	informerFactory.Start(wait.NeverStop)
	informerFactory.WaitForCacheSync(wait.NeverStop)
}

func (sHandle *ScheduleHandler) initDeviceAndNodeDCmInformer(ssn *framework.Session) {
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
	informerFactory.WaitForCacheSync(wait.NeverStop)
}

func (sHandle *ScheduleHandler) updateConfigMapCluster(obj interface{}, operator string) {
	if sHandle == nil {
		klog.V(util.LogDebugLev).Infof("updateConfigMapCluster failed: %s.", util.ArgumentError)
		return
	}
	klog.V(util.LogDebugLev).Infof("update cluster configMap to cache")

	cm, ok := obj.(*v12.ConfigMap)
	if !ok {
		klog.V(util.LogErrorLev).Infof("cannot convert to ConfigMap: %#v", obj)
		return
	}
	sHandle.dealClusterDeviceInfo(cm, operator)
	sHandle.dealClusterNodeInfo(cm, operator)
	sHandle.dealClusterSwitchInfo(cm, operator)
}

func (sHandle *ScheduleHandler) dealClusterDeviceInfo(cm *v12.ConfigMap, operator string) {
	if !strings.HasPrefix(cm.Name, util.ClusterDeviceInfo) {
		return
	}
	deviceInfoMap, err := getDeviceClusterInfoFromCM(cm)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("get device info failed :%#v", err)
		return
	}
	sHandle.DeviceInfos.Lock()
	for deviceCmName, deviceInfo := range deviceInfoMap {
		nodeName := strings.TrimPrefix(deviceCmName, util.DevInfoPreName)
		if operator == util.AddOperator || operator == util.UpdateOperator {
			sHandle.DeviceInfos.Devices[nodeName] = NodeDeviceInfoWithID{
				NodeDeviceInfo: deviceInfo.NodeDeviceInfo,
				SuperPodID:     deviceInfo.SuperPodID,
			}
		} else if operator == util.DeleteOperator {
			delete(sHandle.DeviceInfos.Devices, nodeName)
		}
	}
	sHandle.DeviceInfos.Unlock()
}

func (sHandle *ScheduleHandler) dealClusterNodeInfo(cm *v12.ConfigMap, operator string) {
	if !strings.HasPrefix(cm.Name, util.ClusterNodeInfo) {
		return
	}
	nodeInfoMap, err := getNodeClusterInfoFromCM(cm)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("get node info failed :%#v", err)
		return
	}
	sHandle.NodeInfosFromCm.Lock()
	for nodeCmName, nodeInfo := range nodeInfoMap {
		nodeName := strings.TrimPrefix(nodeCmName, util.NodeDCmInfoNamePrefix)
		if operator == util.AddOperator || operator == util.UpdateOperator {
			sHandle.NodeInfosFromCm.Nodes[nodeName] = nodeInfo
		} else if operator == util.DeleteOperator {
			delete(sHandle.NodeInfosFromCm.Nodes, nodeName)
		}
	}
	sHandle.NodeInfosFromCm.Unlock()
}

func (sHandle *ScheduleHandler) dealClusterSwitchInfo(cm *v12.ConfigMap, operator string) {
	if !strings.HasPrefix(cm.Name, util.ClusterSwitchInfo) {
		return
	}
	switchInfoMap, err := getSwitchClusterInfoFromCM(cm)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("get switch info failed :%v", err)
		return
	}
	sHandle.SwitchInfosFromCm.Lock()
	for switchCmName, switchInfo := range switchInfoMap {
		nodeName := strings.TrimPrefix(switchCmName, util.SwitchCmInfoNamePrefix)
		if operator == util.AddOperator || operator == util.UpdateOperator {
			sHandle.SwitchInfosFromCm.Switches[nodeName] = switchInfo
		} else if operator == util.DeleteOperator {
			delete(sHandle.SwitchInfosFromCm.Switches, nodeName)
		}
	}
	sHandle.SwitchInfosFromCm.Unlock()
}

// CheckUseCIMByConfig check use cluster info manager by config, default true
func (vf *VolcanoFrame) CheckUseCIMByConfig() bool {
	if vf == nil {
		klog.V(util.LogDebugLev).Infof("CheckUseCIMByConfig failed: %s. use default true", util.ArgumentError)
		return true
	}
	configuration, err := util.GetConfigFromSchedulerConfigMap(util.CMInitParamKey, vf.Confs)
	if err != nil {
		klog.V(util.LogDebugLev).Info("cannot get configuration, segmentEnable.")
		return true
	}
	// get segmentEnable by user configuration
	useClusterInfoManager, ok := configuration.Arguments[util.UseClusterInfoManager]
	if !ok {
		klog.V(util.LogDebugLev).Info("CheckUseCIMByConfig doesn't exist useClusterInfoManager.")
		return true
	}
	return useClusterInfoManager == "true"
}

// InitReschedulerFromSsn initialize re-scheduler
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

// InitNSLB2 Init NSLB 2.0
func (sHandle *ScheduleHandler) InitNSLB2(ssn *framework.Session) {
	tmpJobMaps := make(map[api.JobID]SchedulerJob)
	for _, vcJob := range sHandle.Jobs {
		if vcJob.SchedulingTaskNum == len(vcJob.Tasks) || !vcJob.IsTorAffinityJob() {
			tmpJobMaps[vcJob.Name] = vcJob
			continue
		}
		vcJob.initJobBlackTorMaps(sHandle.Tors.torMaps, vcJob.getUsedTorInfos(sHandle))
		vcJob.Annotation = ssn.Jobs[vcJob.Name].PodGroup.Annotations
		tmpJobMaps[vcJob.Name] = vcJob
	}
	sHandle.Jobs = tmpJobMaps
}

// CacheToShareCM cache tors info to configmap
func (sHandle *ScheduleHandler) CacheToShareCM() error {
	data := make(map[string]string, 1)
	toShareMap := sHandleTorsToTorShareMap(sHandle)
	dataByte, err := json.Marshal(toShareMap)
	if err != nil {
		return fmt.Errorf("marshal tor configmap data error %v", err)
	}
	data[GlobalTorInfoKey] = string(dataByte[:])
	putCM := &v12.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: TorShareCMName,
		Namespace: cmNameSpace}, Data: data}
	if err := util.CreateOrUpdateConfigMap(sHandle.FrameAttr.KubeClient, putCM, TorShareCMName,
		cmNameSpace); err != nil {
		klog.V(util.LogInfoLev).Infof("CacheToShareCM CreateOrUpdateConfigMap error: %s", util.SafePrint(err))
	}
	return nil
}

func sHandleTorsToTorShareMap(sHandle *ScheduleHandler) map[string]TorShare {
	torShareMap := make(map[string]TorShare)
	if sHandle.Tors == nil || sHandle.Tors.Tors == nil {
		return torShareMap
	}
	var nodeJobs []NodeJobInfo
	var jobList []string
	var nodeJob NodeJobInfo
	for _, tor := range sHandle.Tors.Tors {
		nodeJobs = []NodeJobInfo{}
		for _, server := range tor.Servers {
			jobList = []string{}
			for _, job := range server.Jobs {
				jobList = append(jobList, job.ReferenceName)
			}
			nodeJob = NodeJobInfo{
				NodeIp:   server.IP,
				NodeName: server.Name,
				JobName:  jobList,
			}
			nodeJobs = append(nodeJobs, nodeJob)
		}
		torShareMap[tor.IP] = TorShare{
			IsHealthy:   tor.IsHealthy,
			IsSharedTor: tor.IsSharedTor,
			NodeJobs:    nodeJobs,
		}
	}
	return torShareMap
}

func isContain(target string, strArray []string) bool {
	for _, each := range strArray {
		if each == target {
			return true
		}
	}
	return false
}

// UpdateConfigMap update deviceInfo in cache
func (sHandle *ScheduleHandler) UpdateConfigMap(obj interface{}, operator string) {
	if sHandle == nil {
		klog.V(util.LogDebugLev).Infof("UpdateConfigMap failed: %s.", util.ArgumentError)
		return
	}
	klog.V(util.LogDebugLev).Infof("Update DeviceInfo to cache")

	cm, ok := obj.(*v12.ConfigMap)
	if !ok {
		klog.V(util.LogErrorLev).Infof("Cannot convert to ConfigMap:%#v", obj)
		return
	}
	if util.CheckConfigMapIsDeviceInfo(cm) {
		if operator == util.AddOperator || operator == util.UpdateOperator {
			sHandle.createOrUpdateDeviceInfo(cm)
			sHandle.createOrUpdateSwitchInfo(cm)
		} else if operator == util.DeleteOperator {
			klog.V(util.LogDebugLev).Infof("Del DeviceInfo from cache")
			nodeName := strings.TrimPrefix(cm.Name, util.DevInfoPreName)
			sHandle.DeviceInfos.Lock()
			delete(sHandle.DeviceInfos.Devices, nodeName)
			sHandle.DeviceInfos.Unlock()

			sHandle.SwitchInfosFromCm.Lock()
			delete(sHandle.SwitchInfosFromCm.Switches, nodeName)
			sHandle.SwitchInfosFromCm.Unlock()
		}
	}
	if util.CheckConfigMapIsNodeInfo(cm) {
		if operator == util.AddOperator || operator == util.UpdateOperator {
			sHandle.createOrUpdateNodeInfo(cm)
		} else if operator == util.DeleteOperator {
			klog.V(util.LogDebugLev).Infof("Del NodeInfo from cache")
			nodeName := strings.TrimPrefix(cm.Name, util.NodeDCmInfoNamePrefix)
			sHandle.NodeInfosFromCm.Lock()
			delete(sHandle.NodeInfosFromCm.Nodes, nodeName)
			sHandle.NodeInfosFromCm.Unlock()
		}
	}
}

func (sHandle *ScheduleHandler) createOrUpdateDeviceInfo(cm *v12.ConfigMap) {
	devInfo, err := getNodeDeviceInfoFromCM(cm)
	if err != nil {
		klog.V(util.LogWarningLev).Infof("get device info failed:%s", err)
		return
	}

	nodeName := strings.TrimPrefix(cm.Name, util.DevInfoPreName)
	sHandle.DeviceInfos.Lock()
	sHandle.DeviceInfos.Devices[nodeName] = NodeDeviceInfoWithID{
		NodeDeviceInfo: devInfo.DeviceInfo,
		SuperPodID:     devInfo.SuperPodID,
	}
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

func (sHandler *ScheduleHandler) createOrUpdateSwitchInfo(cm *v12.ConfigMap) {
	if cm == nil {
		return
	}
	switchInfo := SwitchFaultInfo{}
	data, ok := cm.Data[util.SwitchInfoCmKey]
	if !ok {
		return
	}
	unmarshalErr := json.Unmarshal([]byte(data), &switchInfo)
	if unmarshalErr != nil {
		klog.V(util.LogInfoLev).Infof("unmarshal switchInfo info failed, err: %s.", util.SafePrint(unmarshalErr))
		return
	}
	nodeName := strings.TrimPrefix(cm.Name, util.DevInfoPreName)
	sHandler.SwitchInfosFromCm.Lock()
	sHandler.SwitchInfosFromCm.Switches[nodeName] = switchInfo
	sHandler.SwitchInfosFromCm.Unlock()
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
func (sHandle *ScheduleHandler) BatchNodeOrderFn(task *api.TaskInfo,
	nodes []*api.NodeInfo) (map[string]float64, error) {
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
	scoreMap := initScoreMap(nodes)
	vcJob, ok := sHandle.Jobs[task.Job]
	if !ok {
		klog.V(util.LogDebugLev).Infof("BatchNodeOrderFn %s not req npu.", task.Name)
		return scoreMap, nil
	}

	if vcJob.IsTorAffinityJob() {
		nodeMaps := util.ChangeNodesToNodeMaps(nodes)
		klog.V(util.LogDebugLev).Infof("validNPUJob job is now use tor affinity")
		if sHandle.Tors.torLevel == SingleLayer {
			return sHandle.SetSingleLayerTorAffinityJobNodesScore(task, nodeMaps, vcJob, scoreMap)
		}
		if sHandle.getNSLBVsersion() == defaultNSLBVersion && vcJob.isJobSinglePodRunAsNormal() {
			return sHandle.SetTorAffinityJobNodesScore(task, nodeMaps, vcJob, vcJob.Label[TorAffinityKey], scoreMap)
		}
		if sHandle.getNSLBVsersion() == NSLB2Version {
			return sHandle.SetTorAffinityJobNodesScoreV2(task, nodeMaps, vcJob, scoreMap)
		}
	}

	// 2.Get the best node and top by A,B,C,D rules and require numbers.
	errGet := vcJob.handler.ScoreBestNPUNodes(task, nodes, scoreMap)
	for nodeName := range scoreMap {
		scoreMap[nodeName] *= scoreWeight
	}
	if errGet != nil {
		// get suitable node failed
		klog.V(util.LogErrorLev).Infof("batchNodeOrderFn task[%s] failed by err:[%s].", task.Name, util.SafePrint(errGet))
		return scoreMap, nil
	}
	klog.V(util.LogInfoLev).Infof("batchNodeOrderFn Get task:%s for NPU %+v.", task.Name, scoreMap)

	return scoreMap, nil
}

// SetSingleLayerTorAffinityJobNodesScore single layer switch networking rule
func (sHandle *ScheduleHandler) SetSingleLayerTorAffinityJobNodesScore(task *api.TaskInfo,
	nodeMaps map[string]*api.NodeInfo, vcJob SchedulerJob, scoreMap map[string]float64) (map[string]float64, error) {
	if sHandle == nil || task == nil || len(nodeMaps) == 0 || len(scoreMap) == 0 || !vcJob.JobReadyTag {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogDebugLev).Infof("ScoreBestNPUNodes %s.", err)
		return scoreMap, nil
	}
	result := SetJobServerList(vcJob, sHandle, nodeMaps)
	vcJob = sHandle.Jobs[task.Job]
	if result != nil {
		klog.V(util.LogErrorLev).Infof("check job %s tor affinity failed: %s,"+
			"used servers is %s", vcJob.Name, result, vcJob.SelectServers)
		vcJob.JobReadyTag = false
		sHandle.Jobs[task.Job] = vcJob
	}
	if errGet := sHandle.scoreBestNPUNodes(task, nodeMaps, scoreMap); errGet != nil {
		// get suitable node failed
		klog.V(util.LogDebugLev).Infof("batchNodeOrderFn task[%s] is failed[%s].", task.Name, util.SafePrint(errGet))
	}
	klog.V(util.LogDebugLev).Infof("batchNodeOrderFn set %s for NPU %+v.", task.Name, scoreMap)
	return scoreMap, result
}

// SetTorAffinityJobNodesScore nslb 1.0 rule
func (sHandle *ScheduleHandler) SetTorAffinityJobNodesScore(task *api.TaskInfo, nodeMaps map[string]*api.NodeInfo,
	vcJob SchedulerJob, label string, scoreMap map[string]float64) (map[string]float64, error) {
	if sHandle == nil || task == nil || len(nodeMaps) == 0 || len(scoreMap) == 0 || !vcJob.JobReadyTag {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogDebugLev).Infof("ScoreBestNPUNodes %s.", err)
		return scoreMap, nil
	}

	result := CheckNetSliceIsMeetJobRequire(vcJob, sHandle, nodeMaps)
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
	if errGet := sHandle.scoreBestNPUNodes(task, nodeMaps, scoreMap); errGet != nil {
		// get suitable node failed
		klog.V(util.LogDebugLev).Infof("batchNodeOrderFn task[%s] is failed[%s].", task.Name, util.SafePrint(errGet))
	}
	klog.V(util.LogDebugLev).Infof("batchNodeOrderFn set %s for NPU %+v.", task.Name, scoreMap)
	return scoreMap, result
}

// SetTorAffinityJobNodesScoreV2 nslb 2.0 rule
func (sHandle *ScheduleHandler) SetTorAffinityJobNodesScoreV2(task *api.TaskInfo, nodeMaps map[string]*api.NodeInfo,
	vcJob SchedulerJob, scoreMap map[string]float64) (map[string]float64, error) {
	if sHandle == nil || task == nil || len(nodeMaps) == 0 || len(scoreMap) == 0 || !vcJob.JobReadyTag {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogDebugLev).Infof("ScoreBestNPUNodes err: %s.", err)
		return scoreMap, nil
	}
	defer func() {
		if errGet := sHandle.scoreBestNPUNodes(task, nodeMaps, scoreMap); errGet != nil {
			// get suitable node failed
			klog.V(util.LogDebugLev).Infof("batchNodeOrderFn task[%s] failed[%s].", task.Name, util.SafePrint(errGet))
		}
	}()
	if vcJob.ServerList != nil {
		return scoreMap, nil
	}
	vcJob.MarkTorListByJob(nodeMaps, sHandle)
	if ri := vcJob.getJobsRestartedInfo(); ri != nil {
		if err := vcJob.setJobNodesAfterRestarted(sHandle, ri, vcJob.SchedulingTaskNum); err == nil {
			vcJob.initJobNodeRankByFaultRank(ri, nodeMaps)
			sHandle.Jobs[task.Job] = vcJob
			return scoreMap, nil
		}
		if vcJob.IsJobSinglePodDelete() {
			vcJob.JobReadyTag = false
			sHandle.Jobs[task.Job] = vcJob
			return scoreMap, nil
		}
		klog.V(util.LogWarningLev).Infof("the job is not meet the rescheduling logic and will be scheduled normally")
	}

	tmpTors := deepCopyTorList(sHandle.Tors.Tors)
	result := setJobAvailableNodes(&vcJob, sHandle, nodeMaps)
	if result != nil {
		// recovery the Tors in global
		sHandle.recoveryGlobalTor(tmpTors)
		klog.V(util.LogErrorLev).Infof("check job %s tor affinity failed: %s", vcJob.Name, result)
		vcJob.JobReadyTag = false
	}
	sHandle.Jobs[task.Job] = vcJob

	klog.V(util.LogDebugLev).Infof("batchNodeOrderFn Get %s for NPU %+v.", task.Name, scoreMap)
	return scoreMap, result
}

func (sHandle *ScheduleHandler) recoveryGlobalTor(tors []*Tor) {
	sHandle.Tors.Tors = tors
	sHandle.Tors.initTorMaps()
}

func (sHandle *ScheduleHandler) scoreBestNPUNodes(task *api.TaskInfo, nodeMaps map[string]*api.NodeInfo,
	sMap map[string]float64) error {

	vcjob, ok := sHandle.ScheduleEnv.Jobs[task.Job]
	if !ok {
		return errors.New(util.ArgumentError)
	}

	if vcjob.setBestNodeFromRankIndex(task, sMap) {
		return nil
	}

	vcjob.setJobFaultRankIndex()

	for _, sl := range vcjob.ServerList {
		for _, server := range sl.Servers {
			if vcjob.HealthTorRankIndex[server.Name] != "" {
				continue
			}
			setNodeScoreByTorAttr(sMap, server.Name, sl)
			if _, exist := nodeMaps[server.Name]; exist && server.NodeRank == task.Pod.Annotations[podRankIndex] {
				sMap[server.Name] = maxTorAffinityNodeScore
				return nil
			}
		}
	}
	klog.V(util.LogInfoLev).Infof("ScoreBestNPUNodes task<%s> sMap<%v>", task.Name, sMap)
	return nil
}

func (sHandle *ScheduleHandler) getSharedTorNum() int {
	if sHandle == nil || sHandle.Tors == nil {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogDebugLev).Infof("getSharedTorNum %s.", err)
		return util.ErrorInt
	}
	return sHandle.Tors.sharedTorNum
}

func (sHandle *ScheduleHandler) getNSLBVsersion() string {
	if sHandle == nil || sHandle.Tors == nil {
		err := errors.New(util.ArgumentError)
		klog.V(util.LogDebugLev).Infof("getNSLBVsersion %s.", err)
		return ""
	}
	return sHandle.Tors.nslbVersion
}

func setNodeScoreByTorAttr(sMap map[string]float64, nodeName string, sl *Tor) {
	if sMap == nil {
		return
	}
	sMap[nodeName] = halfTorAffinityNodeScore
	if sl.IsSharedTor == sharedTor {
		sMap[nodeName] = sharedTorAffinityNodeScore
	}
}
