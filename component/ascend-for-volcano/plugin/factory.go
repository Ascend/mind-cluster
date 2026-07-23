/*
Copyright(C)2020-2025. Huawei Technologies Co.,Ltd. All rights reserved.

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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog"
	"volcano.sh/apis/pkg/apis/scheduling"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/conf"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/cache"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/k8s"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/config"
)

// InitNPUSession init npu plugin and nodes.
func (sHandle *ScheduleHandler) InitNPUSession(ssn *framework.Session) error {
	if sHandle == nil || ssn == nil {
		klog.V(util.LogDebugLev).Infof("InitNPUSession failed: %s.", util.ArgumentError)
		return errors.New(util.ArgumentError)
	}
	klog.V(util.LogDebugLev).Infof("enter %s InitNPUSession.", PluginName)
	defer klog.V(util.LogDebugLev).Infof("leave %s InitNPUSession.", PluginName)
	sHandle.CheckResult = CheckResult{
		ValidResult:         map[api.JobID]*api.ValidateResult{},
		EnqueueError:        map[api.JobID]error{},
		BatchOrderError:     map[api.JobID]error{},
		NodePredicateErrors: &NodePredicateError{NodeError: map[api.JobID]map[string]sets.String{}},
	}
	sHandle.PredicatedNodes = make(map[api.JobID]sets.String)
	sHandle.InitVolcanoFrameFromSsn(ssn)
	sHandle.initCmInformer()
	sHandle.InitNodesFromSsn(ssn)
	sHandle.InitJobsFromSsn(ssn)
	sHandle.initJobScheduleInfoRecorder()

	sHandle.InitTorNodeInfo(ssn)
	sHandle.initJobsPlugin()
	sHandle.initCache()
	sHandle.initAffinityCache()
	sHandle.ClusterCache.AffinityCache = sHandle.AffinityCache
	sHandle.startFaultHandler(ssn)
	sHandle.preStartPlugin(ssn)
	return nil
}

// InitJobsFromSsn init all jobs in ssn.
func (sHandle *ScheduleHandler) InitJobsFromSsn(ssn *framework.Session) {
	if sHandle == nil || ssn == nil {
		klog.V(util.LogInfoLev).Infof("InitJobsFromSsn failed: %s.", util.ArgumentError)
		return
	}
	newJobs := make(map[api.JobID]SchedulerJob, util.MapInitNum)
	for jobID, jobInfo := range ssn.Jobs {
		// get ownerInfo, deployment job need
		ownerInfo, err := getOwnerInfo(jobInfo, sHandle.FrameAttr)
		if err != nil {
			klog.V(util.LogWarningLev).Infof("%s getOwnerInfo failed: %s.", jobInfo.Name, util.SafePrint(err))
			continue
		}
		sJob := SchedulerJob{
			Owner:             ownerInfo,
			SuperPods:         sHandle.Jobs[jobID].SuperPods,
			JobReadyTag:       util.PtrInit(true),
			UnscheduledReason: newUnscheduledReason(),
			A5Fields:          A5Fields{},
		}
		if err := sJob.init(jobInfo, sHandle); err != nil {
			klog.V(util.LogWarningLev).Infof("%s InitJobsFromSsn failed: %s.", jobInfo.Name, util.SafePrint(err))
			continue
		}
		// here we should sync a5 fields for use later in scheduling
		sJob.updateSchedulerForA5Fields(sHandle.Jobs[jobID].A5Fields)
		newJobs[jobID] = sJob
	}
	sHandle.Jobs = newJobs
}

// initJobScheduleInfoRecorder update job schedule info recorder.
func (sHandle *ScheduleHandler) initJobScheduleInfoRecorder() {
	tmpRecorder := NewJobScheduleInfoRecorder()
	for jobID, sJob := range sHandle.Jobs {
		// mark the job which server list has been recorded in logs
		if _, ok := sHandle.ServerListRecordFlag[jobID]; ok && sJob.Status == util.PodGroupRunning {
			tmpRecorder.ServerListRecordFlag[jobID] = struct{}{}
		}
		// mark the job which reset configmap has been set
		if _, ok := sHandle.ResetCMSetFlag[jobID]; ok && sJob.SchedulingTaskNum == 0 {
			tmpRecorder.ResetCMSetFlag[jobID] = struct{}{}
		}
		// default value is last session scheduled info that job is in job scheduling or pod scheduling
		tmpRecorder.PodScheduleFlag[jobID] = sHandle.PodScheduleFlag[jobID]
		// if job is need scheduled in this scheduling session, record job is job scheduling or pod scheduling
		// if job is no need scheduled, use last session recorder.
		if sJob.isPodScheduling() {
			tmpRecorder.PodScheduleFlag[jobID] = sJob.SchedulingTaskNum != len(sJob.Tasks)
		}
		// record job last session pending message, for onsessionclose to compare pending message is change
		tmpRecorder.PendingMessage[jobID] = sHandle.PendingMessage[jobID]
	}
	sHandle.JobScheduleInfoRecorder = tmpRecorder

}

func getOwnerInfo(jobInfo *api.JobInfo, vf VolcanoFrame) (OwnerInfo, error) {
	if jobInfo == nil {
		return OwnerInfo{}, errors.New(util.ArgumentError)
	}
	owner := getPodGroupOwnerRef(jobInfo.PodGroup.PodGroup)
	if owner.Kind != ReplicaSetType {
		return OwnerInfo{OwnerReference: owner}, nil
	}
	rs, err := getReplicaSet(vf, jobInfo.Namespace, owner.Name)
	if err != nil {
		return OwnerInfo{}, err
	}
	return OwnerInfo{OwnerReference: owner, Replicas: rs.Spec.Replicas, Annotations: rs.Annotations}, nil
}

func getReplicaSet(vf VolcanoFrame, namespace, name string) (*appsv1.ReplicaSet, error) {
	var rs *appsv1.ReplicaSet
	var ok bool
	key := namespace + "/" + name
	obj, exist, err := vf.informerFactory.Apps().V1().ReplicaSets().Informer().GetIndexer().GetByKey(key)
	if err != nil || !exist {
		klog.V(util.LogWarningLev).Infof("Get rs from indexer failed err: %s, exist: %v.", util.SafePrint(err), exist)
		rs, err = vf.KubeClient.AppsV1().ReplicaSets(namespace).Get(context.TODO(), name,
			metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
	} else {
		rs, ok = obj.(*appsv1.ReplicaSet)
		if !ok {
			return nil, errors.New("the object is not a replicaset")
		}
	}
	return rs, nil
}

func getPodGroupOwnerRef(pg scheduling.PodGroup) metav1.OwnerReference {
	for _, ref := range pg.OwnerReferences {
		if *ref.Controller == true {
			return ref
		}
	}
	return metav1.OwnerReference{}
}

// getJobTemplate get template of all possible segmentation jobs
func (sHandle *ScheduleHandler) getJobTemplate() map[string]map[string]util.VResource {
	jobTemplate := map[string]map[string]util.VResource{
		util.Ascend310P: {
			VNPUTempVir01:        {Aicore: 1, Aicpu: 1, DVPP: AscendDVPPEnabledNull},
			VNPUTempVir02:        {Aicore: util.NPUIndex2, Aicpu: util.NPUIndex2, DVPP: AscendDVPPEnabledNull},
			VNPUTempVir02C1:      {Aicore: util.NPUIndex2, Aicpu: 1, DVPP: AscendDVPPEnabledNull},
			VNPUTempVir04:        {Aicore: util.NPUIndex4, Aicpu: util.NPUIndex4, DVPP: AscendDVPPEnabledNull},
			VNPUTempVir04C3:      {Aicore: util.NPUIndex4, Aicpu: util.NPUIndex3, DVPP: AscendDVPPEnabledNull},
			VNPUTempVir04C3NDVPP: {Aicore: util.NPUIndex4, Aicpu: util.NPUIndex3, DVPP: AscendDVPPEnabledOff},
			VNPUTempVir04C4cDVPP: {Aicore: util.NPUIndex4, Aicpu: util.NPUIndex4, DVPP: AscendDVPPEnabledOn},
		},
		util.Ascend910: {
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
		ChipTypeB2: {
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
		ChipTypeB41: {
			VNPUTempVir05: {Aicore: util.NPUIndex5, Aicpu: util.NPUIndex1, DVPP: AscendDVPPEnabledNull},
			VNPUTempVir10: {Aicore: util.NPUIndex10, Aicpu: util.NPUIndex3, DVPP: AscendDVPPEnabledNull},
		},
		ServerTypeA3X20: {
			VNPUTempVir05: {Aicore: util.NPUIndex5, Aicpu: util.NPUIndex1, DVPP: AscendDVPPEnabledNull},
			VNPUTempVir10: {Aicore: util.NPUIndex10, Aicpu: util.NPUIndex3, DVPP: AscendDVPPEnabledNull},
		},
		ServerTypeA3X24: {
			VNPUTempVir06: {Aicore: util.NPUIndex6, Aicpu: util.NPUIndex1, DVPP: AscendDVPPEnabledNull},
			VNPUTempVir12: {Aicore: util.NPUIndex12, Aicpu: util.NPUIndex3, DVPP: AscendDVPPEnabledNull},
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
	configs := getConfigurationByKey(initConfsFromSsn(ssn.Configurations))
	sHandle.FrameAttr.UID = ssn.UID
	sHandle.FrameAttr.KubeClient = ssn.KubeClient()
	sHandle.FrameAttr.informerFactory = ssn.InformerFactory()
	sHandle.FrameAttr.VJobTemplate = sHandle.getJobTemplate()
	sHandle.initDynamicParameters(configs)
	sHandle.initStaticParameters(configs)
}

// initStaticParameters
func (sHandle *ScheduleHandler) initStaticParameters(configs map[string]string) {
	sHandle.FrameAttr.OnceInit.Do(func() {
		sHandle.FrameAttr.NslbVersion = getNslbVersion(configs)
		sHandle.FrameAttr.SharedTorNum = getShardTorNum(configs)
		sHandle.FrameAttr.UseClusterD = getUseClusterDConfig(configs)
		sHandle.FrameAttr.ForceEnqueue = getForceEnqueueConfig(configs)
		sHandle.FrameAttr.SelfMaintainAvailCard = getSelfMaintainAvailCard(configs)
		klog.V(util.LogInfoLev).Info("param nslbVersion, sharedTorNum, useClusterInfoManager and self-maintain-mount-card " +
			"init success. can not change them and it will not be changed during normal operation of the volcano")
		klog.V(util.LogInfoLev).Infof("init static parameters, nslbversion is <%v>, SharedTorNum <%v>, UseClusterD"+
			" is <%v>", sHandle.FrameAttr.NslbVersion, sHandle.FrameAttr.SharedTorNum, sHandle.FrameAttr.UseClusterD)
	})
}

// initDynamicParameters
func (sHandle *ScheduleHandler) initDynamicParameters(configs map[string]string) {
	if sHandle == nil || configs == nil {
		klog.V(util.LogInfoLev).Infof("InitCache failed: %s.", util.ArgumentError)
		return
	}
	sHandle.FrameAttr.SuperPodSize, sHandle.FrameAttr.SuperPodSizeFromConf = getSizeOfSuperPod(configs)
	sHandle.FrameAttr.ReservePodSize = getReserveNodes(configs, sHandle.FrameAttr.SuperPodSize)
	sHandle.FrameAttr.GraceDeleteTime = getGraceDeleteTime(configs)
	sHandle.FrameAttr.PresetVirtualDevice = getPresetVirtualDeviceConfig(configs)
	sHandle.FrameAttr.ResourceLevelsInfo = initResourceLevels(configs)
	sHandle.FrameAttr.PreferPreviousNode = getPreferPreviousNodeConfig(configs)

}

func initResourceLevels(configs map[string]string) map[string][]util.ResourceTreeLevel {
	levels, err := getConfigLevels(configs)
	if err != nil {
		levels = map[string][]util.ResourceTreeLevel{}
		klog.V(util.LogInfoLev).Infof("init resource levels failed: %v, set resource-level-config to empty", err)
	}
	klog.V(util.LogInfoLev).Infof("init resourceLevels success, effected levels: %v", levels)
	return levels
}

func getConfigLevels(configurations map[string]string) (map[string][]util.ResourceTreeLevel, error) {
	confStr, ok := configurations[configResourceLevelConfig]
	if !ok {
		return nil, errors.New("resource-level-config doesn't exist")
	}

	var levelConfigs map[string]map[string]util.ResourceTreeLevel
	if err := json.Unmarshal([]byte(confStr), &levelConfigs); err != nil {
		return nil, fmt.Errorf("ummarshal config failed, %v", err)
	}

	resourceTreeLevels := make(map[string][]util.ResourceTreeLevel, len(levelConfigs))
	for topoKey, levelConfig := range levelConfigs {
		if len(levelConfigs) == 0 {
			klog.V(util.LogErrorLev).Infof("get %v resource level config failed, config is empty", topoKey)
			continue
		}
		level, err := getConfigLevel(levelConfig)
		if err != nil {
			klog.V(util.LogErrorLev).Infof("get %v resource level config failed, %v", topoKey, err)
			continue
		}
		resourceTreeLevels[topoKey] = level
	}
	if len(resourceTreeLevels) == 0 {
		return nil, errors.New("no valid resource config exist")
	}
	return resourceTreeLevels, nil
}

func getConfigLevel(levelConfig map[string]util.ResourceTreeLevel) ([]util.ResourceTreeLevel, error) {
	singleTopoTree := []util.ResourceTreeLevel{{Type: util.LevelTypeTree, Label: util.TopoTreeLabel}}
	for i := len(levelConfig); i >= util.Level1Number; i-- {
		treeLevel, ok := levelConfig[util.TopoLevelPrefix+strconv.Itoa(i)]
		if !ok {
			return nil, errors.New("topo-level config is not continuous")
		}
		level := util.ResourceTreeLevel{Label: treeLevel.Label, Type: util.LevelTypeMiddle}
		if i == util.Level1Number {
			level.ReservedNode = treeLevel.ReservedNode
		}
		singleTopoTree = append(singleTopoTree, level)
	}
	singleTopoTree = append(singleTopoTree, util.ResourceTreeLevel{Type: util.LevelTypeNode})

	return singleTopoTree, nil
}

// initConfsFromSsn init confs from session
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

// initJobsPlugin init job by plugins.
func (sHandle *ScheduleHandler) initJobsPlugin() {
	for _, vcJob := range sHandle.Jobs {
		if vcJob.policyHandler == nil {
			if vcJob.ReqNPUNum > 0 {
				klog.V(util.LogWarningLev).Infof("initJobsPlugin %s's plugin not register.", vcJob.Name)
			}
			continue
		}
		if err := vcJob.policyHandler.InitMyJobPlugin(vcJob.SchedulerJobAttr, sHandle.ScheduleEnv); err != nil {
			klog.V(util.LogErrorLev).Infof("initJobsPlugin %s init myJobPlugin err %v.", vcJob.Name, err)
			continue
		}
	}
}

// initCache init ScheduleHandler's cache.
func (sHandle *ScheduleHandler) initCache() {
	data := make(map[string]map[string]string, util.MapInitNum)
	data[util.RePropertyCacheName] = make(map[string]string, util.MapInitNum)
	data[util.JobRecovery] = make(map[string]string, util.MapInitNum)
	sHandle.OutputCache = ScheduleCache{
		Names:      make(map[string]string, util.MapInitNum),
		Namespaces: make(map[string]string, util.MapInitNum),
		Data:       data}
}

// initAffinityCache initializes the pod-to-node affinity cache.
// The cache is created only once and survives across scheduler sessions.
//
// On first creation (cold start after scheduler restart), the cache is seeded
// from the current pod→node assignments of all active jobs. Subsequent sessions
// only refresh timestamps and evict expired entries — cache content is maintained
// by NPUAllocateFunc/DeallocateFunc event-driven writes.
func (sHandle *ScheduleHandler) initAffinityCache() {
	if !sHandle.FrameAttr.PreferPreviousNode {
		return
	}

	if sHandle.AffinityCache == nil {
		sHandle.AffinityCache = cache.NewPodNodeAffinityCache()
		// One-time cold start: seed cache from currently running pods.
		for _, job := range sHandle.Jobs {
			if job.Owner.UID == "" {
				continue
			}
			for _, task := range job.Tasks {
				if task.NodeName == "" {
					continue
				}
				rankIndex := getRankFromTask(task)
				if rankIndex == "" {
					continue
				}
				sHandle.AffinityCache.RecordAssignment(job.Owner.UID, rankIndex, task.NodeName)
			}
		}
		klog.V(util.LogInfoLev).Info("affinity cache: initialized new in-memory cache from running pods")
	} else {
		// Refresh timestamps for active owners so TTL counts from last seen session.
		for _, job := range sHandle.Jobs {
			if job.Owner.UID != "" {
				sHandle.AffinityCache.RefreshOwner(job.Owner.UID)
			}
		}
		// Evict owners not seen within TTL (PG deleted long ago)
		sHandle.AffinityCache.EvictExpired(cache.DefaultTTL)
	}

	// Pre-load PrefNodeMap for each active job so scoring reads from this
	// snapshot without querying the cache again during session scheduling.
	for jobID, job := range sHandle.Jobs {
		if job.Owner.UID != "" {
			job.PrefNodeMap = sHandle.AffinityCache.GetPreferredNodeMap(job.Owner.UID)
			sHandle.Jobs[jobID] = job
		}
	}
}

// getRankFromTask extracts the rank index string from a task, preferring the
// hccl/rankIndex annotation, falling back to task.Index.
func getRankFromTask(task util.NPUTask) string {
	if task.Annotation != nil {
		if rank, ok := task.Annotation[PodRankIndexKey]; ok && rank != "" {
			return rank
		}
	}
	if task.Index >= 0 {
		return strconv.Itoa(task.Index)
	}
	return ""
}

// preStartPlugin preStart plugin action.
func (sHandle *ScheduleHandler) preStartPlugin(ssn *framework.Session) {
	for _, job := range sHandle.Jobs {
		// policyHandler of non-NPU job not be inited
		if job.policyHandler == nil {
			continue
		}
		if err := job.policyHandler.PreStartAction(ssn); err != nil {
			if strings.Contains(err.Error(), util.ArgumentError) {
				continue
			}
			klog.V(util.LogWarningLev).Infof("PreStartPlugin %s %s.", job.Name, err)
		}
	}
}

func (sHandle *ScheduleHandler) saveCacheToCm() {
	for spName, cmName := range sHandle.ScheduleEnv.OutputCache.Names {
		nameSpace, okSp := sHandle.ScheduleEnv.OutputCache.Namespaces[spName]
		data, okData := sHandle.ScheduleEnv.OutputCache.Data[spName]
		if !okSp || !okData {
			klog.V(util.LogErrorLev).Infof("SaveCacheToCm %s no namespace or Data in cache.", spName)
			continue
		}

		data, err := k8s.UpdateConfigmapIncrementally(sHandle.FrameAttr.KubeClient, nameSpace, cmName, data)
		if err != nil {
			klog.V(util.LogInfoLev).Infof("get old %s configmap failed: %v, write new data into cm", spName, err)
		}
		var tmpCM = &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      cmName,
				Namespace: nameSpace,
			},
			Data: data,
		}
		if err := k8s.CreateOrUpdateConfigMap(sHandle.FrameAttr.KubeClient, tmpCM, cmName, nameSpace); err != nil {
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
	if sHandle.FaultHandle != nil {
		if err := sHandle.FaultHandle.PreStopAction(&sHandle.ScheduleEnv); err != nil {
			klog.V(util.LogErrorLev).Infof("PreStopPlugin  %s.", util.SafePrint(err))
		}
	}

	sHandle.saveCacheToCm()

	if sHandle.Tors == nil || sHandle.Tors.GetNSLBVersion() == defaultNSLBVersion {
		return
	}
	err := sHandle.cacheToShareCM()
	if err != nil {
		klog.V(util.LogErrorLev).Infof("cacheToShareCM error: %v", err)
	}
}

// initCmInformer init cm informer, support cluster info manager and device plugin
func (sHandle *ScheduleHandler) initCmInformer() {
	if sHandle.FrameAttr.KubeClient == nil {
		klog.V(util.LogErrorLev).Info("kube client in session is nil")
		return
	}
	k8s.InitCmInformer(sHandle.FrameAttr.KubeClient, sHandle.FrameAttr.UseClusterD)
}

// startFaultHandler initialize re-scheduler
func (sHandle *ScheduleHandler) startFaultHandler(ssn *framework.Session) {
	if sHandle.FaultHandle == nil {
		return
	}
	if preErr := sHandle.FaultHandle.Execute(&sHandle.ScheduleEnv, ssn); preErr != nil {
		klog.V(util.LogWarningLev).Infof("PreStartAction failed by %s", preErr)
		return
	}
}

// cacheToShareCM cache tors info to configmap
func (sHandle *ScheduleHandler) cacheToShareCM() error {
	data := make(map[string]string, 1)
	toShareMap := sHandleTorsToTorShareMap(sHandle)
	dataByte, err := json.Marshal(toShareMap)
	if err != nil {
		return fmt.Errorf("marshal tor configmap data error %v", err)
	}
	data[GlobalTorInfoKey] = string(dataByte[:])
	putCM := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: TorShareCMName,
		Namespace: cmNameSpace}, Data: data}
	if err := k8s.CreateOrUpdateConfigMap(sHandle.FrameAttr.KubeClient, putCM, TorShareCMName,
		cmNameSpace); err != nil {
		klog.V(util.LogInfoLev).Infof("cacheToShareCM CreateOrUpdateConfigMap error: %s", util.SafePrint(err))
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
			for jobName := range server.Jobs {
				jobList = append(jobList, jobName)
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

// TaskOrderFn Sort the selected tasks.
// Non-fault (healthy) pods are ordered before fault pods, so healthy pods
// claim their original nodes first. Within each group, tasks are ordered by rank.
func (sHandle *ScheduleHandler) TaskOrderFn(InterfaceA interface{}, InterfaceB interface{}) int {
	taskInfoA, ok := InterfaceA.(*api.TaskInfo)
	if !ok {
		klog.V(util.LogDebugLev).Info("TaskOrderFn failed, object is not a TaskInfo")
		return taskOrderSamePriority
	}
	taskInfoB, ok := InterfaceB.(*api.TaskInfo)
	if !ok {
		klog.V(util.LogDebugLev).Info("TaskOrderFn failed, object is not a TaskInfo")
		return taskOrderSamePriority
	}

	job, ok := sHandle.Jobs[taskInfoA.Job]
	if !ok {
		klog.V(util.LogDebugLev).Infof("TaskOrderFn (%s/%s): job is not exist", taskInfoA.Namespace, taskInfoA.Name)
		return taskOrderSamePriority
	}

	rRankId := sHandle.resolveRankIndex(taskInfoB, job)
	lRankId := sHandle.resolveRankIndex(taskInfoA, job)

	podGroupEnable, exist := job.Label[PodGroupScheduleKey]
	if exist && podGroupEnable == PodGroupScheduleValue {
		lRank, lErr := strconv.Atoi(lRankId)
		rRank, rErr := strconv.Atoi(rRankId)
		if lErr != nil || rErr != nil {
			return taskOrderSamePriority
		}
		if lRank < rRank {
			return taskOrderHighPriority
		}
		return taskOrderLowPriority

	}
	if sHandle.FrameAttr.PreferPreviousNode {
		lNode := sHandle.AffinityCache.GetPreferredNode(job.Owner.UID, lRankId)
		rNode := sHandle.AffinityCache.GetPreferredNode(job.Owner.UID, rRankId)
		if lNode != rNode {
			if lNode != "" {
				return taskOrderHighPriority
			}
			return taskOrderLowPriority
		}
	}
	return taskOrderSamePriority
}

// isFaultPod returns true if the task's rank is recorded as a fault task in
// the fault job cache. Uses the fault snapshot rather than real-time node
// health, so tasks whose nodes recovered are still recognized as fault pods.
func (sHandle *ScheduleHandler) isFaultPod(task *api.TaskInfo, job SchedulerJob) bool {
	if sHandle.FaultHandle == nil {
		return false
	}
	rankIndex := sHandle.resolveRankIndex(task, job)
	if rankIndex == "" {
		return false
	}
	return sHandle.FaultHandle.IsFaultTaskByRank(task.Job, rankIndex)
}

// resolveRankIndex returns the rank for a task, trying the pod annotation
// first, then falling back to the task's index in the job.  Deployment-type
// tasks may not have the annotation set yet at scoring time.
func (sHandle *ScheduleHandler) resolveRankIndex(task *api.TaskInfo, vcJob SchedulerJob) string {
	if rankIndex, ok := task.Pod.Annotations[PodRankIndexKey]; ok && rankIndex != "" {
		return rankIndex
	}
	if nTask, ok := vcJob.Tasks[task.UID]; ok {
		return strconv.Itoa(nTask.Index)
	}
	return ""
}

func (sHandle *ScheduleHandler) obtainTaskRankId(task *api.TaskInfo) (int, error) {
	if task == nil {
		klog.V(util.LogDebugLev).Infof("obtainTaskRankId failed: %s.", util.ArgumentError)
		return 0, errors.New(util.ArgumentError)
	}
	rankIndex, ok := task.Pod.Annotations[PodRankIndexKey]
	if !ok {
		klog.V(util.LogDebugLev).Infof("obtainTaskRankId task(%s/%s): rankIndex not exist",
			task.Namespace, task.Name)
		return 0, errors.New(util.RankIdNotExistError)
	}
	rankId, err := strconv.Atoi(rankIndex)
	if err != nil {
		klog.V(util.LogDebugLev).Infof("obtainTaskRankId task(%s/%s): rankIndex(%s) is not int",
			task.Namespace, task.Name, rankIndex)
		return 0, errors.New(util.ArgumentError)
	}
	return rankId, nil
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

	if !util.IsNPUTask(task) {
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
	if !vcJob.isNPUJob() {
		klog.V(util.LogDebugLev).Infof("BatchNodeOrderFn vc-job:%#v is not npu job.", vcJob)
		return nil, nil
	}
	// 1. Policy-level mandatory scoring (super pod, multi-level, normal topology constraints)
	errGet := vcJob.policyHandler.ScoreBestNPUNodes(task, nodes, scoreMap)

	// 2. Add preference for the pod's previous node (tiebreaker, +AffScore4)
	sHandle.addPreferPreviousNodeScore(task, scoreMap, vcJob)

	// 3. Fault handling score reduction (sub-healthy -1, fault node -64)
	if sHandle.FaultHandle != nil {
		sHandle.FaultHandle.ScoreBestNPUNodes(task, scoreMap)
	}

	for nodeName := range scoreMap {
		scoreMap[nodeName] *= scoreWeight
	}
	if errGet != nil {
		// get suitable node failed
		klog.V(util.LogErrorLev).Infof("batchNodeOrderFn task[%s] failed by err:[%s].", task.Name, util.SafePrint(errGet))
		return scoreMap, errGet
	}
	klog.V(util.LogDebugLev).Infof("batchNodeOrderFn Get task:%s for NPU %+v.", task.Name, scoreMap)

	return scoreMap, nil
}

// addPreferPreviousNodeScore boosts a node in the score map based on
// pod-to-node affinity from prior scheduling sessions.
//
// Score map nodes are divided into three categories:
//   - selfNode:  the current pod's previous node (from PrefNodeMap[myRank])
//   - peerNodes: nodes previously used by other pods of the same job
//   - otherNodes: all remaining nodes
//
// Scoring rules:
//
//	Fault pod:
//	  1. otherNodes → boost the best-scoring otherNode
//	  2. selfNode   → fallback: boost selfNode if it is still in scoreMap
//
//	Non-fault pod:
//	  1. selfNode   → boost selfNode if in scoreMap
//	  2. otherNodes → boost the best-scoring otherNode
//	  3. peerNodes  → boost the best-scoring peerNode (lowest priority)
func (sHandle *ScheduleHandler) addPreferPreviousNodeScore(
	task *api.TaskInfo, scoreMap map[string]float64, vcJob SchedulerJob) {

	if vcJob.IsSuperPodJob() || vcJob.IsMultiLevelJob() || vcJob.IsJobHasTorAffinityLabel() {
		klog.V(util.LogDebugLev).Infof("addPreferPreviousNodeScore: skip, job %s is super-pod or multi-level or nslb",
			vcJob.Name)
		return
	}

	if !sHandle.FrameAttr.PreferPreviousNode {
		klog.V(util.LogDebugLev).Info("addPreferPreviousNodeScore: skip, prefer-previous-node is disabled")
		return
	}
	if task == nil || len(scoreMap) == 0 {
		klog.V(util.LogDebugLev).Infof("addPreferPreviousNodeScore: skip, task=%v scoreMapLen=%d",
			task != nil, len(scoreMap))
		return
	}

	rankIndex := sHandle.resolveRankIndex(task, vcJob)
	if rankIndex == "" {
		klog.V(util.LogDebugLev).Infof("addPreferPreviousNodeScore: skip, task %s has no rank index",
			task.Name)
		return
	}

	prefMap := vcJob.PrefNodeMap
	if prefMap == nil || len(prefMap) == 0 {
		klog.V(util.LogDebugLev).Infof("addPreferPreviousNodeScore: skip, no PrefNodeMap for "+
			"owner=%s", vcJob.Owner.UID)
		return
	}

	myRank, err := strconv.Atoi(rankIndex)
	if err != nil {
		klog.V(util.LogWarningLev).Infof("convert task %s rank %s failed", task.Name, rankIndex)
		return
	}

	sHandle.applyPreviousNodePreference(task, scoreMap, prefMap, myRank, vcJob)
}

// applyPreviousNodePreference categorizes scoreMap nodes and applies a scoring
// bonus according to the fault/non-fault priority rules.
func (sHandle *ScheduleHandler) applyPreviousNodePreference(
	task *api.TaskInfo, scoreMap map[string]float64, prefMap map[int]string,
	myRank int, vcJob SchedulerJob) {

	cat := categorizeNodes(scoreMap, prefMap, myRank)
	bonus := cat.maxScore + defaultPreferPreviousScore

	if sHandle.isFaultPod(task, vcJob) {
		sHandle.boostFaultPod(task, scoreMap, cat, bonus, myRank)
		return
	}
	sHandle.boostNonFaultPod(task, scoreMap, cat, bonus, myRank)
}

// nodeCategories holds the three-class partition of scoreMap nodes plus the
// global max score across all entries.
type nodeCategories struct {
	selfNode   string
	peerNodes  map[string]float64
	otherNodes map[string]float64
	maxScore   float64
}

// categorizeNodes partitions scoreMap into selfNode / peerNodes / otherNodes
// and returns the global max score.
func categorizeNodes(scoreMap map[string]float64, prefMap map[int]string, myRank int) nodeCategories {
	cat := nodeCategories{selfNode: prefMap[myRank]}

	peerSet := make(map[string]struct{})
	for rank, node := range prefMap {
		if rank != myRank {
			peerSet[node] = struct{}{}
		}
	}

	cat.peerNodes = make(map[string]float64)
	cat.otherNodes = make(map[string]float64)

	for nodeName, score := range scoreMap {
		if score > cat.maxScore {
			cat.maxScore = score
		}
		if nodeName == cat.selfNode {
			continue
		}
		if _, isPeer := peerSet[nodeName]; isPeer {
			cat.peerNodes[nodeName] = score
		} else {
			cat.otherNodes[nodeName] = score
		}
	}
	return cat
}

// boostFaultPod applies the fault-pod scoring rules: prefer otherNodes first,
// fall back to selfNode if no other nodes are available.
func (sHandle *ScheduleHandler) boostFaultPod(
	task *api.TaskInfo, scoreMap map[string]float64,
	cat nodeCategories, bonus float64, myRank int) {

	if best := nodeWithMaxScore(cat.otherNodes); best != "" {
		scoreMap[best] = bonus
		klog.V(util.LogInfoLev).Infof("addPreferPreviousNodeScore: fault pod task=%s rank=%d "+
			"boosted otherNode=%s score=%.0f", task.Name, myRank, best, bonus)
		return
	}
	if cat.selfNode != "" {
		if _, exists := scoreMap[cat.selfNode]; exists {
			scoreMap[cat.selfNode] = bonus
			klog.V(util.LogInfoLev).Infof("addPreferPreviousNodeScore: fault pod task=%s rank=%d "+
				"fallback selfNode=%s score=%.0f", task.Name, myRank, cat.selfNode, bonus)
		}
	}
}

// boostNonFaultPod applies the non-fault-pod scoring rules: selfNode first,
// then otherNodes, then peerNodes (lowest priority).
func (sHandle *ScheduleHandler) boostNonFaultPod(
	task *api.TaskInfo, scoreMap map[string]float64,
	cat nodeCategories, bonus float64, myRank int) {

	if cat.selfNode != "" {
		if _, exists := scoreMap[cat.selfNode]; exists {
			scoreMap[cat.selfNode] = bonus
			klog.V(util.LogInfoLev).Infof("addPreferPreviousNodeScore: task=%s rank=%d "+
				"boosted selfNode=%s score=%.0f", task.Name, myRank, cat.selfNode, bonus)
			return
		}
	}
	if best := nodeWithMaxScore(cat.otherNodes); best != "" {
		scoreMap[best] = bonus
		klog.V(util.LogInfoLev).Infof("addPreferPreviousNodeScore: task=%s rank=%d "+
			"boosted otherNode=%s score=%.0f", task.Name, myRank, best, bonus)
		return
	}
	if best := nodeWithMaxScore(cat.peerNodes); best != "" {
		scoreMap[best] = bonus
		klog.V(util.LogInfoLev).Infof("addPreferPreviousNodeScore: task=%s rank=%d "+
			"boosted peerNode=%s score=%.0f", task.Name, myRank, best, bonus)
	}
}

// nodeWithMaxScore returns the node name with the highest score in the given
// map, or empty string if the map is empty.
func nodeWithMaxScore(nodes map[string]float64) string {
	var best string
	var max float64
	for name, score := range nodes {
		if best == "" || score > max {
			best = name
			max = score
		}
	}
	return best
}

// getConfigurationByKey called by GetConfigFromSchedulerConfigMap
func getConfigurationByKey(configurations []config.Configuration) map[string]string {
	for _, cf := range configurations {
		if cf.Name == util.CMInitParamKey {
			return cf.Arguments
		}
	}
	return map[string]string{}
}

// getSizeOfSuperPod get size of super pod
func getSizeOfSuperPod(configurations map[string]string) (int, int) {
	superPodSize := getSuperPodInfoFromConfig(sizeOfSuperPodKey, configurations)
	// we need to cache the original value from configuration
	superPodSizeFromConfig := superPodSize
	if superPodSize == 0 {
		klog.V(util.LogWarningLev).Infof(" super-pod-size configuration should be a number bigger than 0, "+
			"set default super-pod-size: %d", defaultSuperPodSize)
		superPodSize = defaultSuperPodSize
	}
	return superPodSize, superPodSizeFromConfig
}

// getReserveNodes get reserve nodes
func getReserveNodes(configurations map[string]string, superPodSize int) int {
	reserve := getSuperPodInfoFromConfig(reserveNodesKey, configurations)
	if reserve == 0 {
		klog.V(util.LogWarningLev).Infof("reserve-nodes less than or equal 0, "+
			"set as default: %d", defaultReserveNodes)
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
	return reserve
}

func getSuperPodInfoFromConfig(key string, configurations map[string]string) int {
	if len(configurations) == 0 {
		klog.V(util.LogWarningLev).Info("volcano scheduler config init-params map is nil")
		return 0
	}
	value, ok := configurations[key]
	if !ok {
		klog.V(util.LogWarningLev).Infof("%s configuration not exist", key)
		return 0
	}

	res, err := strconv.Atoi(value)
	if err != nil {
		klog.V(util.LogWarningLev).Infof("cannot convert %s configuration, err: %v", key, err)
		return 0
	}
	if res < 0 {
		klog.V(util.LogWarningLev).Infof(" %s configuration should not be negative number", key)
		return 0
	}
	return res
}

// checkGraceDeleteTimeValid used by GetGraceDeleteTime for validity checking
func checkGraceDeleteTimeValid(overTime int64) bool {
	if overTime < minGraceOverTime || overTime > maxGraceOverTime {
		klog.V(util.LogWarningLev).Infof("GraceOverTime value should be range [2, 3600], configured is [%d], "+
			"GraceOverTime will not be changed", overTime)
		return false
	}
	// use user's configuration to set grace over time
	klog.V(util.LogInfoLev).Infof("set GraceOverTime to new value [%d].", overTime)
	return true
}

// getGraceDeleteTime get grace delete time
func getGraceDeleteTime(conf map[string]string) int64 {
	klog.V(util.LogDebugLev).Info("enter GetGraceDeleteTime ...")
	defer klog.V(util.LogDebugLev).Info("leave GetGraceDeleteTime ...")
	if len(conf) == 0 {
		klog.V(util.LogWarningLev).Infof("GetGraceDeleteTime failed: %s, no conf", util.ArgumentError)
		return DefaultGraceOverTime
	}
	// get grace over time by user configuration
	overTimeStr, ok := conf[GraceOverTimeKey]
	if !ok {
		klog.V(util.LogWarningLev).Info("set GraceOverTime failed and will not be changed, " +
			"key grace-over-time doesn't exists.")
		return DefaultGraceOverTime
	}
	overTime, err := strconv.ParseInt(overTimeStr, util.Base10, util.BitSize64)
	if err != nil {
		klog.V(util.LogWarningLev).Infof("set GraceOverTime failed and will not be changed, "+
			"grace-over-time is invalid [%s].", util.SafePrint(overTimeStr))
		return DefaultGraceOverTime
	}
	// check time validity
	if !checkGraceDeleteTimeValid(overTime) {
		return DefaultGraceOverTime
	}
	return overTime
}

// getUseClusterDConfig check use cluster info manager by config, default true
func getUseClusterDConfig(conf map[string]string) bool {
	useClusterInfoManager, ok := conf[util.UseClusterInfoManager]
	if !ok {
		klog.V(util.LogWarningLev).Info("CheckUseCIMByConfig doesn't exist useClusterInfoManager.")
		return true
	}
	return useClusterInfoManager == "true"
}

func getForceEnqueueConfig(conf map[string]string) bool {
	forceEnqueue, ok := conf[util.ForceEnqueue]
	if !ok {
		klog.V(util.LogWarningLev).Info("forceEnqueue doesn't exist in config, set as true.")
		return true
	}
	return forceEnqueue == "true"
}

// getSelfMaintainAvailCard check volcano self maintain available card by config, default true
func getSelfMaintainAvailCard(conf map[string]string) bool {
	selfMaintainAvailCard, ok := conf[util.SelfMaintainAvailCard]
	if !ok {
		klog.V(util.LogWarningLev).Info("CheckUseCIMByConfig doesn't exist self-maintain-available-card.")
		return true
	}
	return selfMaintainAvailCard == "true"
}

// getPresetVirtualDeviceConfig get VNPU segmentEnable by init plugin parameters, return true if static
func getPresetVirtualDeviceConfig(conf map[string]string) bool {
	// get segmentEnable by user configuration
	segmentEnable, ok := conf[util.SegmentEnable]
	if !ok {
		klog.V(util.LogWarningLev).Info("checkVNPUSegmentEnable doesn't exist presetVirtualDevice.")
		return false
	}
	return segmentEnable == "true"
}

// getPreferPreviousNodeConfig reads prefer-previous-node feature flag from config.
func getPreferPreviousNodeConfig(conf map[string]string) bool {
	enabled, ok := conf[preferPreviousNodeKey]
	if !ok {
		return false
	}
	return enabled == "true"
}

// getShardTorNum get shared tor num from configmap
func getShardTorNum(conf map[string]string) int {
	str := conf[keyOfSharedTorNum]
	sharedTorNum, err := strconv.Atoi(str)
	if err != nil {
		klog.V(util.LogWarningLev).Infof("getSharedTorNum %s.", err)
		return shareTorNum2
	}
	if sharedTorNum != shareTorNum1 && sharedTorNum != shareTorNum2 {
		klog.V(util.LogWarningLev).Info("sharedTorNum is illegal. use default config")
		return shareTorNum2
	}
	return sharedTorNum
}

// getNslbVersion get nslb version from config
func getNslbVersion(conf map[string]string) string {
	nslbVersion := conf[keyOfNSLBVersion]
	if nslbVersion != defaultNSLBVersion && nslbVersion != NSLB2Version {
		klog.V(util.LogWarningLev).Info("nslbVersion is illegal. use default config")
		return defaultNSLBVersion
	}
	return nslbVersion
}
