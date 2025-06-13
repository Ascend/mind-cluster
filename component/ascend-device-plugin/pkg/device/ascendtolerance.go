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

// Package device a series of device function
package device

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"Ascend-device-plugin/pkg/common"
	"Ascend-device-plugin/pkg/kubeclient"
	"ascend-common/common-utils/hwlog"
)

var processPolicyTable = map[string]int{
	common.EmptyError:          common.EmptyErrorLevel,
	common.IgnoreError:         common.IgnoreErrorLevel,
	common.RestartRequestError: common.RestartRequestErrorLevel,
	common.RestartError:        common.RestartErrorLevel,
	common.FreeResetError:      common.FreeResetErrorLevel,
	common.ResetError:          common.ResetErrorLevel,
	common.IsolateError:        common.IsolateErrorLevel,
}

// HotResetManager hot reset manager
type HotResetManager interface {
	GetResetDevNumOnce() (int, error)
	GetDevIdList(string) []int32
	GetTaskDevFaultInfoList(string) ([]*common.TaskDevInfo, error)
	GetTaskPod(string) (v1.Pod, error)
	GetAllTaskDevFaultInfoList() map[string][]*common.TaskDevInfo
	GetDevProcessPolicy(string) string
	GetTaskProcessPolicy(string) (string, int, error)
	GetDevListInReset() map[int32]struct{}
	GetDevListByPolicyLevel([]*common.TaskDevInfo, int) (map[int32]struct{}, error)
	GetNeedResetDevMap([]*common.TaskDevInfo) (map[int32]int32, error)
	GetGlobalDevFaultInfo(logicID int32) (*common.DevFaultInfo, error)
	GetTaskResetInfo([]*common.TaskDevInfo, string, string, string) (*common.TaskResetInfo, error)
	GetTaskFaultRankInfo([]*common.TaskDevInfo) (*common.TaskFaultInfo, error)
	GetFaultDev2PodMap() (map[int32]v1.Pod, error)
	GetTaskNameByPod(pod v1.Pod) string
	GenerateTaskDevFaultInfoList(devIdList []int32, rankIndex string) ([]*common.TaskDevInfo, error)
	UpdateFaultDev2PodMap([]int32, v1.Pod) error
	UpdateGlobalDevFaultInfoCache([]*common.NpuDevice, []int32) error
	UpdateTaskDevListCache(map[string][]int32) error
	UpdateTaskDevFaultInfoCache(map[string][]*common.TaskDevInfo) error
	UpdateTaskPodCache(map[string]v1.Pod) error
	UpdateFreeTask(map[string]struct{}, map[string][]int32)
	SetTaskInReset(string) error
	SetDevInReset(int32) error
	SetAllDevInReset(info *common.TaskResetInfo) error
	UnSetTaskInReset(string) error
	UnSetDevInReset(int32) error
	UnSetAllDevInReset(*common.TaskResetInfo) error
	IsCurNodeTaskInReset(string) bool
	IsExistFaultyDevInTask(string) bool
	DeepCopyDevInfo(*common.TaskDevInfo) *common.TaskDevInfo
	DeepCopyDevFaultInfoList([]*common.TaskDevInfo) []*common.TaskDevInfo
	SyncResetCM(context.Context, *kubeclient.ClientK8s)
	GetCMFromCache(string) (*v1.ConfigMap, error)
}

// HotResetTools hot reset tool
type HotResetTools struct {
	resetDevNumOnce     int
	allTaskDevList      map[string][]int32
	allTaskDevFaultInfo map[string][]*common.TaskDevInfo
	globalDevFaultInfo  map[int32]*common.DevFaultInfo
	taskPod             map[string]v1.Pod
	faultDev2PodMap     map[int32]v1.Pod
	resetTask           map[string]struct{}
	resetDev            map[int32]struct{}
	queue               workqueue.RateLimitingInterface
	podIndexer          cache.Indexer
	cmIndexer           cache.Indexer
	jobs                map[string]string
	noResetCmPodKeys    map[string]struct{}
}

// NewHotResetManager create HotResetManager and init data
func NewHotResetManager(devUsage string, deviceNum int, boardId uint32) HotResetManager {
	resetDevNumOnce := getResetDevNumOnce(devUsage, deviceNum, boardId)
	if resetDevNumOnce == 0 {
		return nil
	}
	return &HotResetTools{
		resetDevNumOnce:  resetDevNumOnce,
		resetTask:        map[string]struct{}{},
		resetDev:         map[int32]struct{}{},
		faultDev2PodMap:  map[int32]v1.Pod{},
		jobs:             map[string]string{},
		noResetCmPodKeys: map[string]struct{}{},
	}
}

// getResetDevNumOnce get reset device num at a time.
// 910 and 910A2 device reset by ring, 910A3 reset all devices on the node
func getResetDevNumOnce(devUsage string, deviceNum int, boardId uint32) int {
	var resetDevNumOnce int
	switch common.ParamOption.RealCardType {
	case common.Ascend910:
		resetDevNumOnce = common.Ascend910RingsNum
	case common.Ascend910B:
		if devUsage == common.Infer {
			if boardId == common.A300IA2BoardId || boardId == common.A800IA2NoneHccsBoardId || boardId == common.
				A800IA2NoneHccsBoardIdOld {
				return common.Ascend910BRingsNumInfer
			}
			resetDevNumOnce = common.Ascend910BRingsNumTrain
		}

		if devUsage == common.Train {
			resetDevNumOnce = common.Ascend910BRingsNumTrain
		}
	case common.Ascend910A3:
		// 900A3 device, deviceNum is 16; 9000A3 device, deviceNum is 8
		resetDevNumOnce = deviceNum
	default:
		hwlog.RunLog.Error("only 910 device support grace tolerance")
	}
	return resetDevNumOnce
}

// SyncResetCM sync reset-cm event
func (hrt *HotResetTools) SyncResetCM(ctx context.Context, client *kubeclient.ClientK8s) {
	cmFactory := informers.NewSharedInformerFactoryWithOptions(client.Clientset, 0,
		informers.WithTweakListOptions(func(options *metav1.ListOptions) {
			options.LabelSelector = labels.SelectorFromSet(labels.Set{"reset": "true"}).String()
		}),
	)
	cmInformer := cmFactory.Core().V1().ConfigMaps().Informer()
	cmInformer.AddEventHandler(client.ResourceEventHandler(kubeclient.CMResource, checkConfigMap))
	go cmInformer.Run(ctx.Done())

	hrt.queue = client.Queue
	hrt.podIndexer = client.PodInformer.GetIndexer()
	hrt.cmIndexer = cmInformer.GetIndexer()

	cache.WaitForCacheSync(ctx.Done(), cmInformer.HasSynced, client.PodInformer.HasSynced)

	go hrt.run()

	go func() {
		<-ctx.Done()
		hrt.queue.ShutDown()
	}()
}

func (hrt *HotResetTools) run() {
	hwlog.RunLog.Info("starting handle reset-cm event")
	for hrt.processNextWorkItem() {
	}
}

func (hrt *HotResetTools) processNextWorkItem() bool {
	hwlog.RunLog.Debugf("queue length: %d", hrt.queue.Len())
	obj, shutdown := hrt.queue.Get()
	if shutdown {
		hwlog.RunLog.Error("shutdown, stop processing work queue")
		return false
	}
	defer hrt.queue.Done(obj)
	_, ok := obj.(kubeclient.Event)
	if !ok {
		hrt.queue.Forget(obj)
		return true
	}
	hrt.handleEvent(obj)
	return true
}

func (hrt *HotResetTools) handleEvent(obj interface{}) {
	switch obj.(kubeclient.Event).Resource {
	case kubeclient.PodResource:
		hrt.handlePodEvent(obj)
	case kubeclient.CMResource:
		hrt.handleConfigMapEvent(obj)
	default:
		hrt.queue.Forget(obj)
		hwlog.RunLog.Errorf("unsupported resource: %s", obj.(kubeclient.Event).Resource)
	}
}

func (hrt *HotResetTools) handlePodEvent(obj interface{}) {
	switch obj.(kubeclient.Event).Type {
	case kubeclient.EventTypeAdd:
		hrt.handlePodAddEvent(obj)
	case kubeclient.EventTypeDelete:
		hrt.handlePodDeleteEvent(obj)
	default:
		hrt.queue.Forget(obj)
		hwlog.RunLog.Debugf("hotReset scene not watch %s event(%s)", obj.(kubeclient.Event).Resource,
			obj.(kubeclient.Event).Type)
	}
}

func (hrt *HotResetTools) handlePodAddEvent(obj interface{}) {
	event, ok := obj.(kubeclient.Event)
	if !ok {
		hwlog.RunLog.Error("get kubeclient event error")
		return
	}
	hwlog.RunLog.Debugf("handle pod(%s) %s event", event.Key, event.Type)
	pod, err := hrt.getPodFromCache(event.Key)
	if err != nil {
		hwlog.RunLog.Warn(err)
		if hrt.queue.NumRequeues(obj) < common.MaxPodEventRetryTimes {
			hrt.queue.AddRateLimited(obj)
		} else {
			hrt.queue.Forget(obj)
		}
		return
	}
	jobName := common.GetJobNameOfPod(pod)
	if jobName == "" {
		hwlog.RunLog.Errorf("get job name of pod(%s) failed", event.Key)
		hrt.queue.Forget(obj)
		return
	}
	hrt.jobs[event.Key] = jobName
	hrt.writeCmToFileWhilePodAdd(pod, event)
	hrt.queue.Forget(obj)
}

func (hrt *HotResetTools) writeCmToFileWhilePodAdd(pod *v1.Pod, event kubeclient.Event) {
	if pod == nil {
		return
	}
	podNameSpace := pod.GetNamespace()
	jobName := common.GetJobNameOfPod(pod)
	dataTraceDir := fmt.Sprintf("%s/%s", common.DataTraceConfigDir, podNameSpace+"."+common.DataTraceCmPrefix+jobName)
	resetDir := common.GenResetDirName(podNameSpace, common.ResetInfoCMNamePrefix+jobName)
	if err := os.MkdirAll(dataTraceDir, common.DefaultPerm); err != nil {
		hwlog.RunLog.Warnf("failed to create data trace configmap dir for pod %s, err: %v", pod.Name, err)
	}
	if err := os.MkdirAll(resetDir, common.DefaultPerm); err != nil {
		hwlog.RunLog.Warnf("failed to create reset configmap dir for pod %s, err: %v", pod.Name, err)
	}

	dataTraceCm, err := hrt.GetCMFromCache(fmt.Sprintf(pod.GetNamespace() + "/" + common.DataTraceCmPrefix + jobName))
	if err != nil {
		hwlog.RunLog.Warnf("failed to get cm cache for pod %s", pod.Name)
	} else {
		if err := hrt.writeCMToFile(dataTraceCm); err != nil {
			hwlog.RunLog.Errorf("failed to write cm(%s) to file, err: %v", dataTraceCm.Name, err)
		}
	}

	resetCm, err := hrt.GetCMFromCache(fmt.Sprintf(pod.GetNamespace() + "/" + common.ResetInfoCMNamePrefix + jobName))
	if err != nil {
		_, ok := hrt.noResetCmPodKeys[event.Key]
		if !ok {
			hwlog.RunLog.Warn(err)
			hrt.noResetCmPodKeys[event.Key] = struct{}{}
		}
	} else {
		if err := hrt.writeCMToFile(resetCm); err != nil {
			hwlog.RunLog.Errorf("failed to write cm(%s) to file, err: %v", resetCm.Name, err)
		}
	}
}

func (hrt *HotResetTools) handlePodDeleteEvent(obj interface{}) {
	event, ok := obj.(kubeclient.Event)
	if !ok {
		hwlog.RunLog.Error("get kubeclient event error")
		return
	}
	hwlog.RunLog.Debugf("handle pod(%s) delete event", event.Key)
	if _, ok = hrt.noResetCmPodKeys[event.Key]; ok {
		delete(hrt.noResetCmPodKeys, event.Key)
	}
	jobName, ok := hrt.jobs[event.Key]
	if !ok {
		hwlog.RunLog.Errorf("job of pod(%s) not found in cache", event.Key)
		hrt.queue.Forget(obj)
		return
	}
	keySlice := strings.Split(event.Key, "/")
	if len(keySlice) != common.KeySliceLength {
		hwlog.RunLog.Errorf("pod(%s) is invalid", event.Key)
		hrt.queue.Forget(obj)
		return
	}
	namespace := keySlice[0]
	if rmErr := common.RemoveResetFileAndDir(namespace, common.ResetInfoCMNamePrefix+jobName); rmErr != nil {
		hwlog.RunLog.Errorf("failed to remove file: %v", rmErr)
	}
	if rmErr := common.RemoveDataTraceFileAndDir(namespace, jobName); rmErr != nil {
		hwlog.RunLog.Errorf("failed to remove file: %v", rmErr)
	}

	delete(hrt.jobs, event.Key)
	hrt.queue.Forget(obj)
}

func (hrt *HotResetTools) getPodFromCache(podKey string) (*v1.Pod, error) {
	item, exist, err := hrt.podIndexer.GetByKey(podKey)
	if err != nil || !exist {
		return nil, fmt.Errorf("get pod(%s) failed, err: %v, exist: %v", podKey, err, exist)
	}
	pod, ok := item.(*v1.Pod)
	if !ok {
		return nil, fmt.Errorf("convert pod(%s) failed", podKey)
	}
	return pod, nil
}

// GetCMFromCache get configmap from indexer cache
func (hrt *HotResetTools) GetCMFromCache(cmKey string) (*v1.ConfigMap, error) {
	item, exist, err := hrt.cmIndexer.GetByKey(cmKey)
	if err != nil || !exist {
		return nil, fmt.Errorf("get cm(%s) failed, err: %v, exist: %v", cmKey, err, exist)
	}
	cm, ok := item.(*v1.ConfigMap)
	if !ok {
		return nil, fmt.Errorf("convert pod(%s) failed", cmKey)
	}
	return cm, nil
}

func (hrt *HotResetTools) writeCMToFile(cm *v1.ConfigMap) error {
	if strings.HasPrefix(cm.Name, common.DataTraceCmPrefix) {
		dir := fmt.Sprintf("%s/%s", common.DataTraceConfigDir, cm.Namespace+"."+cm.Name)
		fileFullName := filepath.Join(dir, common.DataTraceCmProfilingSwitchKey)
		data, ok := cm.Data[common.DataTraceCmProfilingSwitchKey]
		if !ok {
			return fmt.Errorf("found cm %s, but without key %s",
				cm.Namespace+"."+cm.Name, common.DataTraceCmProfilingSwitchKey)
		}
		if err := common.WriteToFile(data, fileFullName); err != nil {
			return fmt.Errorf("failed to write file %s, err: %v", fileFullName, common.DataTraceCmProfilingSwitchKey)
		}
		hwlog.RunLog.Infof("suceessfully wrote file %s", fileFullName)
		return nil
	}
	if strings.HasPrefix(cm.Name, common.ResetInfoCMNamePrefix) {
		data, ok := cm.Data[common.ResetInfoCMDataKey]
		if !ok {
			return fmt.Errorf("cm(%s) data(%s) not exist", cm.Name, common.ResetInfoCMDataKey)
		}
		writeErr := common.WriteToFile(data, common.GenResetFileName(cm.Namespace, cm.Name))
		if writeErr != nil {
			return fmt.Errorf("failed to write data to file: %v", writeErr)
		}
		hwlog.RunLog.Debugf("write cm(%s) data(%s) to file success", cm.Name, data)

		restartType, ok := cm.Data[common.ResetInfoTypeKey]
		if ok {
			err := common.WriteToFile(restartType, common.GenResetTypeFileName(cm.Namespace, cm.Name))
			if err != nil {
				return fmt.Errorf("failed to write restartType to file: %v", err)
			}
			hwlog.RunLog.Debugf("write cm(%s) restartType(%s) to file success", cm.Name, restartType)
		}
	}
	return nil
}

func (hrt *HotResetTools) handleConfigMapEvent(obj interface{}) {
	switch obj.(kubeclient.Event).Type {
	case kubeclient.EventTypeAdd:
		hrt.handleCMUpdateEvent(obj)
	case kubeclient.EventTypeUpdate:
		hrt.handleCMUpdateEvent(obj)
	case kubeclient.EventTypeDelete:
		hrt.handleCMDeleteEvent(obj)
	default:
		hrt.queue.Forget(obj)
		hwlog.RunLog.Debugf("hotReset scene not watch %s event(%s)", obj.(kubeclient.Event).Resource,
			obj.(kubeclient.Event).Type)
	}
}

func (hrt *HotResetTools) handleCMUpdateEvent(obj interface{}) {
	event, ok := obj.(kubeclient.Event)
	if !ok {
		hwlog.RunLog.Error("get kube client event failed")
		return
	}
	hwlog.RunLog.Infof("handle cm(%s) update event", event.Key)
	cm, err := hrt.GetCMFromCache(event.Key)
	if err != nil {
		hwlog.RunLog.Errorf("get cm(%s) failed, err: %v", event.Key, err)
		hrt.queue.Forget(obj)
		return
	}
	if strings.HasPrefix(cm.Name, common.DataTraceCmPrefix) {
		dir := fmt.Sprintf("%s/%s", common.DataTraceConfigDir, cm.Namespace+"."+cm.Name)
		fileFullName := filepath.Join(dir, common.DataTraceCmProfilingSwitchKey)
		// if file is not created yet, will not deal with it, only when pod is sighted by this node
		// the file will be created in pod informer handler
		// then if the configmap is updated, this handler will update file
		if _, checkErr := os.Stat(dir); checkErr != nil {
			hwlog.RunLog.Debugf("check file(%s) failed, err: %v", dir, checkErr)
			hrt.queue.Forget(obj)
			return
		}
		if err = hrt.writeCmToFileSystem(cm, common.DataTraceCmProfilingSwitchKey, fileFullName, obj); err != nil {
			hwlog.RunLog.Error(err)
		}
		return
	}
	if strings.HasPrefix(cm.Name, common.ResetInfoCMNamePrefix) {
		dir := common.GenResetDirName(cm.Namespace, cm.Name)
		if _, checkErr := os.Stat(dir); checkErr != nil {
			hwlog.RunLog.Debugf("check file(%s) failed, err: %v", dir, checkErr)
			hrt.queue.Forget(obj)
			return
		}
		if err = hrt.writeCMToFile(cm); err != nil {
			hwlog.RunLog.Errorf("failed to write cm(%s) to file, err: %v", cm.Name, err)
			hrt.queue.AddRateLimited(obj)
			return
		}
	}
	hrt.queue.Forget(obj)
}

func (hrt *HotResetTools) handleCMDeleteEvent(obj interface{}) {
	event, ok := obj.(kubeclient.Event)
	if !ok {
		hwlog.RunLog.Errorf("get kube-client event failed")
		return
	}
	hwlog.RunLog.Debugf("handle cm(%s) delete event", event.Key)
	keySlice := strings.Split(event.Key, "/")
	if len(keySlice) != common.KeySliceLength {
		hrt.queue.Forget(obj)
		return
	}
	namespace, name := keySlice[0], keySlice[1]
	if strings.HasPrefix(name, common.ResetInfoCMNamePrefix) {
		file := common.GenResetFileName(namespace, name)
		rmErr := os.Remove(file)
		if rmErr != nil && !os.IsNotExist(rmErr) {
			hwlog.RunLog.Errorf("failed to remove file(%s): %v", file, rmErr)
		}
		typeFile := common.GenResetTypeFileName(namespace, name)
		rmErr = os.Remove(typeFile)
		if rmErr != nil && !os.IsNotExist(rmErr) {
			hwlog.RunLog.Errorf("failed to remove file(%s): %v", typeFile, rmErr)
		}
	}
	if strings.HasPrefix(name, common.DataTraceCmPrefix) {
		jobName := strings.TrimPrefix(name, common.DataTraceCmPrefix)
		dataTraceFileName := fmt.Sprintf("%s/%s/%s", common.DataTraceConfigDir,
			namespace+"."+common.DataTraceCmPrefix+jobName, common.DataTraceCmProfilingSwitchKey)
		hwlog.RunLog.Infof("will delete data trace file: %s", dataTraceFileName)
		if rmErr := os.Remove(dataTraceFileName); rmErr != nil && !os.IsNotExist(rmErr) {
			hwlog.RunLog.Errorf("failed to remove file(%s): %v", dataTraceFileName, rmErr)
		}
	}
	hrt.queue.Forget(obj)
}

// writeCmToFileSystem write the cm data cm.data[key] into filePath
func (hrt *HotResetTools) writeCmToFileSystem(cm *v1.ConfigMap, key, filePath string, obj interface{}) error {
	data, ok := cm.Data[key]
	if !ok {
		hwlog.RunLog.Warnf("found cm %s, but without key %s", cm.Namespace+"."+cm.Name, key)
		hrt.queue.Forget(obj)
		return fmt.Errorf("found cm %s, but without key %s", cm.Namespace+"."+cm.Name, key)
	}
	if err := common.WriteToFile(data, filePath); err != nil {
		hwlog.RunLog.Errorf("failed to write file: %s for cm: %s, err: %v", filePath,
			cm.Namespace+"."+cm.Name, err)
		hrt.queue.AddRateLimited(obj)
		return fmt.Errorf("failed to write file: %s for cm: %s, err: %v", filePath, cm.Namespace+"."+cm.Name, err)
	}
	hrt.queue.Forget(obj)
	return nil
}

func checkConfigMap(obj interface{}) bool {
	cm, ok := obj.(*v1.ConfigMap)
	if !ok {
		hwlog.RunLog.Debugf("Cannot convert to ConfigMap:%#v", obj)
		return false
	}
	return strings.HasPrefix(cm.Name, common.ResetInfoCMNamePrefix) || strings.HasPrefix(cm.Name, "data-trace-")
}

// GetResetDevNumOnce get reset device num at a time
func (hrt *HotResetTools) GetResetDevNumOnce() (int, error) {
	// not initialized or not 910 device, the value will be zero
	if hrt.resetDevNumOnce == 0 {
		return 0, errors.New("reset device num at a time is zero")
	}
	return hrt.resetDevNumOnce, nil
}

// GetTaskDevFaultInfoList return task device fault info list
func (hrt *HotResetTools) GetTaskDevFaultInfoList(taskName string) ([]*common.TaskDevInfo, error) {
	taskDevFaultInfoList, ok := hrt.allTaskDevFaultInfo[taskName]
	if !ok {
		return nil, fmt.Errorf("task %s is not in task device fault info list cache", taskName)
	}
	return taskDevFaultInfoList, nil
}

// GetTaskPod return task pod
func (hrt *HotResetTools) GetTaskPod(taskName string) (v1.Pod, error) {
	pod, ok := hrt.taskPod[taskName]
	if !ok {
		return v1.Pod{}, fmt.Errorf("task %s is not in task pod cache", taskName)
	}
	return pod, nil
}

// GetAllTaskDevFaultInfoList return all task device fault info list
func (hrt *HotResetTools) GetAllTaskDevFaultInfoList() map[string][]*common.TaskDevInfo {
	return hrt.allTaskDevFaultInfo
}

// GetDevListInReset return the logic id list of device in reset
func (hrt *HotResetTools) GetDevListInReset() map[int32]struct{} {
	return hrt.resetDev
}

// GetGlobalDevFaultInfo return global device fault info from cache using input logic id
func (hrt *HotResetTools) GetGlobalDevFaultInfo(logicID int32) (*common.DevFaultInfo, error) {
	globalDevFaultInfo, ok := hrt.globalDevFaultInfo[logicID]
	if !ok {
		return nil, fmt.Errorf("device %d is not in global device fault info list cache", logicID)
	}
	return globalDevFaultInfo, nil
}

// GetDevProcessPolicy return the policy of device with fault
func (hrt *HotResetTools) GetDevProcessPolicy(faultType string) string {
	switch faultType {
	case common.NormalNPU, common.NotHandleFault, common.SubHealthFault:
		return common.EmptyError
	case common.RestartRequest:
		return common.RestartRequestError
	case common.RestartBusiness:
		return common.RestartError
	case common.FreeRestartNPU:
		return common.FreeResetError
	case common.RestartNPU:
		return common.ResetError
	default:
		return common.IsolateError
	}
}

// GetTaskProcessPolicy return a task process policy
func (hrt *HotResetTools) GetTaskProcessPolicy(taskName string) (string, int, error) {
	devFaultInfoList, ok := hrt.allTaskDevFaultInfo[taskName]
	if !ok {
		return "", -1, fmt.Errorf("this task is not in the cache")
	}
	var processPolicy string
	var processPolicyLevel int
	for _, devFaultInfo := range devFaultInfoList {
		devPolicyLevel, ok := processPolicyTable[devFaultInfo.Policy]
		if !ok {
			return "", -1, fmt.Errorf("invalid policy of device fault info in task %s", taskName)
		}
		if devPolicyLevel > processPolicyLevel {
			processPolicy = devFaultInfo.Policy
			processPolicyLevel = devPolicyLevel
		}
	}
	return processPolicy, processPolicyLevel, nil
}

// GetDevIdList convert device str to device logic id list
func (hrt *HotResetTools) GetDevIdList(devStr string) []int32 {
	var phyIDs []int32
	for _, deviceName := range strings.Split(devStr, common.CommaSepDev) {
		phyID, _, err := common.GetDeviceID(deviceName, common.CommaSepDev)
		if err != nil {
			hwlog.RunLog.Errorf("get phyID failed, err: %v", err)
			return nil
		}
		phyIDs = append(phyIDs, int32(phyID))
	}
	return phyIDs
}

// GetDevListByPolicyLevel return the dev list by policy level
func (hrt *HotResetTools) GetDevListByPolicyLevel(devFaultInfoList []*common.TaskDevInfo,
	policyLevel int) (map[int32]struct{}, error) {
	devList := make(map[int32]struct{})
	for _, devFaultInfo := range devFaultInfoList {
		policyType, ok := processPolicyTable[devFaultInfo.Policy]
		if !ok {
			err := fmt.Errorf("invalid policy str of device %d", devFaultInfo.LogicId)
			hwlog.RunLog.Error(err)
			return nil, err
		}
		if policyType >= policyLevel {
			if _, ok := devList[devFaultInfo.LogicId]; !ok {
				devList[devFaultInfo.LogicId] = struct{}{}
			}
		}
	}
	return devList, nil
}

// GetNeedResetDevMap return device logic id list to be reset
func (hrt *HotResetTools) GetNeedResetDevMap(devFaultInfoList []*common.TaskDevInfo) (map[int32]int32, error) {
	needResetDevMap := make(map[int32]int32)
	resetDevNumOnce, err := hrt.GetResetDevNumOnce()
	if err != nil {
		hwlog.RunLog.Error(err)
		return nil, err
	}
	for _, devFaultInfo := range devFaultInfoList {
		policyType, ok := processPolicyTable[devFaultInfo.Policy]
		if !ok {
			err := fmt.Errorf("invalid policy str of device %d", devFaultInfo.LogicId)
			hwlog.RunLog.Error(err)
			return nil, err
		}
		if policyType == common.RestartErrorLevel || policyType == common.ResetErrorLevel ||
			policyType == common.RestartRequestErrorLevel {
			resetIndex := devFaultInfo.LogicId / int32(resetDevNumOnce)
			if _, ok := needResetDevMap[devFaultInfo.LogicId]; !ok {
				needResetDevMap[resetIndex*int32(resetDevNumOnce)] = devFaultInfo.LogicId
			}
		}
	}
	return needResetDevMap, nil
}

// GetTaskResetInfo return the detail reset info of task to process
func (hrt *HotResetTools) GetTaskResetInfo(devFaultInfoList []*common.TaskDevInfo, policy, initPolicy,
	status string) (*common.TaskResetInfo, error) {
	faultRing := make(map[int]struct{}, common.RingSum)
	var rankList = make([]*common.TaskDevInfo, 0)
	resetDevNumOnce, err := hrt.GetResetDevNumOnce()
	if err != nil {
		hwlog.RunLog.Error(err)
		return nil, err
	}
	for _, devFaultInfo := range devFaultInfoList {
		policyType, ok := processPolicyTable[devFaultInfo.Policy]
		if !ok {
			err := fmt.Errorf("invalid policy str of device %d", devFaultInfo.LogicId)
			hwlog.RunLog.Error(err)
			return nil, err
		}
		if policyType != common.RestartErrorLevel && policyType != common.ResetErrorLevel &&
			policyType != common.RestartRequestErrorLevel {
			continue
		}
		ringStartIndex := int(devFaultInfo.LogicId) / resetDevNumOnce
		faultRing[ringStartIndex] = struct{}{}
	}
	for _, devInfo := range devFaultInfoList {
		ringIndex := int(devInfo.LogicId) / resetDevNumOnce
		if _, ok := faultRing[ringIndex]; !ok {
			continue
		}
		newDevInfo := hrt.DeepCopyDevInfo(devInfo)
		newDevInfo.Policy = policy
		newDevInfo.InitialPolicy = initPolicy
		newDevInfo.Status = status
		rankList = append(rankList, newDevInfo)
	}
	return &common.TaskResetInfo{
		RankList: rankList,
	}, nil
}

// GetTaskFaultRankInfo return the fault rank info of task to update fault cm
func (hrt *HotResetTools) GetTaskFaultRankInfo(devFaultInfoList []*common.TaskDevInfo) (*common.TaskFaultInfo, error) {
	taskFaultInfo := &common.TaskFaultInfo{
		FaultRank: make([]int, 0),
	}
	faultRing := make(map[int]struct{}, common.RingSum)
	resetDevNumOnce, err := hrt.GetResetDevNumOnce()
	if err != nil {
		hwlog.RunLog.Error(err)
		return nil, err
	}
	for _, devFaultInfo := range devFaultInfoList {
		policy := processPolicyTable[devFaultInfo.Policy]
		if policy != common.RestartErrorLevel && policy != common.ResetErrorLevel &&
			policy != common.RestartRequestErrorLevel {
			continue
		}
		ringStartIndex := int(devFaultInfo.LogicId) / resetDevNumOnce
		faultRing[ringStartIndex] = struct{}{}
	}
	for _, devInfo := range devFaultInfoList {
		ringIndex := int(devInfo.LogicId) / resetDevNumOnce
		if _, ok := faultRing[ringIndex]; !ok {
			continue
		}
		taskFaultInfo.FaultRank = append(taskFaultInfo.FaultRank, devInfo.RankId)
	}
	return taskFaultInfo, nil
}

// GetFaultDev2PodMap return map which contains fault device and pod
func (hrt *HotResetTools) GetFaultDev2PodMap() (map[int32]v1.Pod, error) {
	if hrt.faultDev2PodMap == nil {
		return nil, fmt.Errorf("no valid faultDev2PodMap here")
	}
	return hrt.faultDev2PodMap, nil
}

// GetTaskNameByPod get task name which written by volcano or operator
func (hrt *HotResetTools) GetTaskNameByPod(pod v1.Pod) string {
	return common.GetJobNameOfPod(&pod)
}

// GenerateTaskDevFaultInfoList generate device fault info list in a task by device logic id list and rank index
func (hrt *HotResetTools) GenerateTaskDevFaultInfoList(devIdList []int32,
	rankIndex string) ([]*common.TaskDevInfo, error) {
	sort.Slice(devIdList, func(i, j int) bool {
		return devIdList[i] < devIdList[j]
	})
	rankStart, err := strconv.Atoi(rankIndex)
	if err != nil {
		hwlog.RunLog.Errorf("failed to convert rank index to int, err: %v", err)
		return nil, err
	}
	devNum := len(devIdList)
	taskDevInfoList := make([]*common.TaskDevInfo, 0, len(devIdList))
	for _, devId := range devIdList {
		var rankId int
		switch rankIndex {
		case common.InferRankIndex:
			rankId = rankStart
		default:
			rankId = rankStart*devNum + len(taskDevInfoList)
		}
		faultInfo, ok := hrt.globalDevFaultInfo[devId]
		if !ok {
			return nil, fmt.Errorf("device %d is not in global cache", devId)
		}
		taskDevInfo := &common.TaskDevInfo{
			RankId:       rankId,
			DevFaultInfo: *faultInfo,
		}
		taskDevInfoList = append(taskDevInfoList, taskDevInfo)
	}
	return taskDevInfoList, nil
}

// UpdateFaultDev2PodMap updates the mapping between the unhealthy device and pod
func (hrt *HotResetTools) UpdateFaultDev2PodMap(devList []int32, pod v1.Pod) error {
	if hrt.faultDev2PodMap == nil {
		return fmt.Errorf("no valid faultDev2PodMap here")
	}
	for _, device := range devList {
		// save when device is unhealthy
		if hrt.globalDevFaultInfo[device].Policy != common.EmptyError &&
			hrt.globalDevFaultInfo[device].Policy != common.IgnoreError {
			hrt.faultDev2PodMap[device] = pod
			continue
		}

		// do not delete cache after receive recover event because device is resetting now
		if _, ok := hrt.resetDev[device]; ok {
			continue
		}

		// delete when device is healthy
		if _, ok := hrt.faultDev2PodMap[device]; ok {
			delete(hrt.faultDev2PodMap, device)
		}
	}
	return nil
}

// UpdateGlobalDevFaultInfoCache update global device fault info cache
func (hrt *HotResetTools) UpdateGlobalDevFaultInfoCache(devDeviceList []*common.NpuDevice, isoDevList []int32) error {
	if len(devDeviceList) == 0 {
		return fmt.Errorf("npu device list is nil")
	}
	hrt.globalDevFaultInfo = make(map[int32]*common.DevFaultInfo, len(devDeviceList))
	for _, device := range devDeviceList {
		hrt.globalDevFaultInfo[device.LogicID] = &common.DevFaultInfo{}
		hrt.globalDevFaultInfo[device.LogicID].LogicId = device.LogicID
		hrt.globalDevFaultInfo[device.LogicID].ErrorCode = device.FaultCodes
		if common.IntInList(device.LogicID, isoDevList) {
			hrt.globalDevFaultInfo[device.LogicID].Policy = common.IsolateError
		} else {
			hrt.globalDevFaultInfo[device.LogicID].Policy =
				hrt.GetDevProcessPolicy(common.GetFaultType(device.FaultCodes, device.LogicID))
		}
	}
	return nil
}

// UpdateTaskDevListCache update all task device list cache
func (hrt *HotResetTools) UpdateTaskDevListCache(taskDevList map[string][]int32) error {
	if taskDevList == nil {
		return fmt.Errorf("task device list is nil")
	}
	hrt.allTaskDevList = taskDevList
	return nil
}

// UpdateTaskDevFaultInfoCache update all task device fault info cache
func (hrt *HotResetTools) UpdateTaskDevFaultInfoCache(taskDevFaultInfo map[string][]*common.TaskDevInfo) error {
	if taskDevFaultInfo == nil {
		return fmt.Errorf("taskDevFaultInfo is nil")
	}
	hrt.allTaskDevFaultInfo = taskDevFaultInfo
	return nil
}

// UpdateTaskPodCache update all task pod cache
func (hrt *HotResetTools) UpdateTaskPodCache(taskPod map[string]v1.Pod) error {
	if taskPod == nil {
		return fmt.Errorf("taskPod is nil")
	}
	hrt.taskPod = taskPod
	return nil
}

// UpdateFreeTask unset task in reset task after delete task
func (hrt *HotResetTools) UpdateFreeTask(taskListUsedDevice map[string]struct{}, newTaskDevList map[string][]int32) {
	for taskName := range hrt.resetTask {
		if _, ok := taskListUsedDevice[taskName]; !ok || hrt.isTaskDevListChange(taskName, newTaskDevList) {
			delete(hrt.resetTask, taskName)
			hwlog.RunLog.Infof("success to delete task reset cache for %s, is in used list: %v, "+
				"reset tasks is %v", taskName, ok, hrt.resetTask)
		}
	}
}

func (hrt *HotResetTools) isTaskDevListChange(taskName string, newTaskDevList map[string][]int32) bool {
	if _, ok := hrt.allTaskDevList[taskName]; !ok {
		return false
	}
	if _, ok := newTaskDevList[taskName]; !ok {
		return false
	}
	return common.Int32Join(hrt.allTaskDevList[taskName], common.UnderLine) !=
		common.Int32Join(newTaskDevList[taskName], common.UnderLine)
}

// IsCurNodeTaskInReset check whether the current task is being reset on the current node
func (hrt *HotResetTools) IsCurNodeTaskInReset(taskName string) bool {
	if _, ok := hrt.resetTask[taskName]; !ok {
		return false
	}
	return true
}

// IsExistFaultyDevInTask check if any fault device exist on current task
func (hrt *HotResetTools) IsExistFaultyDevInTask(taskName string) bool {
	if _, ok := hrt.allTaskDevList[taskName]; !ok {
		hwlog.RunLog.Warnf("task: %s is not exist in cache", taskName)
		return false
	}
	for _, pod := range hrt.faultDev2PodMap {
		taskNameInPod, ok := pod.Annotations[common.ResetTaskNameKey]
		if !ok {
			taskNameInPod, ok = pod.Labels[common.ResetTaskNameKeyInLabel]
			if !ok {
				hwlog.RunLog.Error("failed to get task name by task key in IsExistFaultyDevInTask")
				return false
			}
		}
		if taskNameInPod == taskName {
			hwlog.RunLog.Infof("faulty device exists in task: %s", taskName)
			return true
		}
	}

	return false
}

// SetTaskInReset set a task to the reset state
func (hrt *HotResetTools) SetTaskInReset(taskName string) error {
	if _, ok := hrt.resetTask[taskName]; ok {
		return fmt.Errorf("task %s is resetting", taskName)
	}
	hrt.resetTask[taskName] = struct{}{}
	hwlog.RunLog.Infof("set task %s to reset state, reset tasks is %v", taskName, hrt.resetTask)
	return nil
}

// SetDevInReset set a device to the reset state
func (hrt *HotResetTools) SetDevInReset(devId int32) error {
	if _, ok := hrt.resetDev[devId]; ok {
		return fmt.Errorf("dev %d is resetting", devId)
	}
	hrt.resetDev[devId] = struct{}{}
	return nil
}

// SetAllDevInReset set all device in a task to the reset state
func (hrt *HotResetTools) SetAllDevInReset(resetInfo *common.TaskResetInfo) error {
	for _, devInfo := range resetInfo.RankList {
		if err := hrt.SetDevInReset(devInfo.LogicId); err != nil {
			return err
		}
	}
	return nil
}

// UnSetDevInReset unset a device in a task to leave the reset state
func (hrt *HotResetTools) UnSetDevInReset(devId int32) error {
	if _, ok := hrt.resetDev[devId]; !ok {
		return fmt.Errorf("device %d is not resetting", devId)
	}
	delete(hrt.resetDev, devId)
	return nil
}

// UnSetAllDevInReset unset all device in a task to leave the reset state
func (hrt *HotResetTools) UnSetAllDevInReset(resetInfo *common.TaskResetInfo) error {
	for _, devInfo := range resetInfo.RankList {
		if err := hrt.UnSetDevInReset(devInfo.LogicId); err != nil {
			return err
		}
	}
	return nil
}

// UnSetTaskInReset unset a task to leave the reset state
func (hrt *HotResetTools) UnSetTaskInReset(taskName string) error {
	if _, ok := hrt.resetTask[taskName]; !ok {
		return fmt.Errorf("task %s is not in reset task cache", taskName)
	}
	delete(hrt.resetTask, taskName)
	hwlog.RunLog.Infof("success to delete task reset cache for %s, reset tasks is %v", taskName, hrt.resetTask)
	return nil
}

// DeepCopyDevInfo copy device info deeply
func (hrt *HotResetTools) DeepCopyDevInfo(devInfo *common.TaskDevInfo) *common.TaskDevInfo {
	return &common.TaskDevInfo{
		RankId:       devInfo.RankId,
		DevFaultInfo: devInfo.DevFaultInfo,
	}
}

// DeepCopyDevFaultInfoList copy device fault info list deeply
func (hrt *HotResetTools) DeepCopyDevFaultInfoList(devFaultInfoList []*common.TaskDevInfo) []*common.TaskDevInfo {
	var newDevFaultInfoList []*common.TaskDevInfo
	for _, devFaultInfo := range devFaultInfoList {
		newDevFaultInfoList = append(newDevFaultInfoList, &common.TaskDevInfo{
			RankId:       devFaultInfo.RankId,
			DevFaultInfo: devFaultInfo.DevFaultInfo,
		})
	}
	return newDevFaultInfoList
}
