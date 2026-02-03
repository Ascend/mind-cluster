/* Copyright(C) 2022. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package kubeclient a series of k8s function
package kubeclient

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"

	"Ascend-device-plugin/pkg/common"
	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
)

const domainForKubeletConnectErr = "kubeletConnect"

// Similar to the K8s metadata structure
type metaData struct {
	Annotation map[string]string `json:"annotations"`
}

type podMetaData map[string]metaData

var tryUpdatePodWaitTime = common.UpdatePodWaitTime * time.Millisecond

// TryUpdatePodAnnotation is to try updating pod annotation
func (ki *ClientK8s) TryUpdatePodAnnotation(pod *v1.Pod, annotation map[string]string) error {
	if pod == nil {
		return fmt.Errorf("param pod is nil")
	}
	if annotation == nil {
		return fmt.Errorf("invalid annotation")
	}

	newPodMetaData := podMetaData{common.MetaData: metaData{Annotation: annotation}}
	podUpdateMetaData, err := json.Marshal(newPodMetaData)
	if err != nil {
		hwlog.RunLog.Errorf("failed to marshal the node status data, error is %v", err)
		return err
	}

	for i := 0; i < common.RetryUpdateCount; i++ {
		if _, err = ki.PatchPod(pod, podUpdateMetaData); err == nil {
			return nil
		}

		// There is no need to retry if the pod does not exist
		if errors.IsNotFound(err) {
			return err
		}

		hwlog.RunLog.Warnf("patch pod annotation failed: %v, try again", err)
		time.Sleep(tryUpdatePodWaitTime)
	}

	return fmt.Errorf("patch pod annotation failed, exceeded max number of retries")
}

// TryUpdatePodCacheAnnotation is to try updating pod annotation in both api server and cache
func (ki *ClientK8s) TryUpdatePodCacheAnnotation(pod *v1.Pod, annotation map[string]string) error {
	if pod == nil {
		return fmt.Errorf("param pod is nil")
	}
	if err := ki.TryUpdatePodAnnotation(pod, annotation); err != nil {
		hwlog.RunLog.Errorf("update pod annotation in api server failed, err: %v", err)
		return err
	}
	// update cache
	lock.Lock()
	defer lock.Unlock()
	for i, podInCache := range podCache {
		if podInCache.Namespace == pod.Namespace && podInCache.Name == pod.Name {
			for k, v := range annotation {
				podCache[i].Annotations[k] = v
			}
			hwlog.RunLog.Debugf("update annotation in pod cache success, name: %s, namespace: %s", pod.Name, pod.Namespace)
			return nil
		}
	}
	hwlog.RunLog.Warnf("no pod found in cache when update annotation, name: %s, namespace: %s", pod.Name, pod.Namespace)
	return nil
}

func (ki *ClientK8s) createOrUpdateDeviceCM(cm *v1.ConfigMap) error {
	// use update first
	if _, err := ki.UpdateConfigMap(cm); errors.IsNotFound(err) {
		if _, err := ki.CreateConfigMap(cm); err != nil {
			return fmt.Errorf("unable to create configmap, %v", err)
		}
		return nil
	} else {
		return err
	}
}

func getDeviceInfoManuallySeparateNPUData(deviceInfo *v1.ConfigMap) (string, error) {
	data, ok := deviceInfo.Data[common.DeviceInfoCMManuallySeparateNPUKey]
	if !ok {
		return "", fmt.Errorf("%s not exist, from %s", common.DeviceInfoCMManuallySeparateNPUKey, deviceInfo.Name)
	}

	return data, nil
}

func (ki *ClientK8s) GetManuallySeparateNPUFromDeviceInfo(deviceInfo *v1.ConfigMap) []common.PhyId {
	manuallySeparateNPUData, err := getDeviceInfoManuallySeparateNPUData(deviceInfo)
	phyIDs := make([]common.PhyId, 0)
	if err != nil {
		hwlog.RunLog.Warnf("failed to get manually seperate NPU data, error: %v", err)
		return phyIDs
	}

	deviceRunMode, err := common.GetDeviceRunMode()
	if err != nil {
		hwlog.RunLog.Warnf("failed to get device run mode, error: %v", err)
		return phyIDs
	}

	manuallySeparateNPUs := strings.Split(manuallySeparateNPUData, ",")
	if len(manuallySeparateNPUs) == 1 && manuallySeparateNPUs[0] == "" {
		hwlog.RunLog.Debug("manually seperate NPU cache is empty, skip the lookup phase")
		return phyIDs
	}

	for _, manuallySeparateNPU := range manuallySeparateNPUs {
		deviceNameCheck := common.CheckDeviceName(manuallySeparateNPU, deviceRunMode)
		if !deviceNameCheck {
			hwlog.RunLog.Warnf("in %v run mode, device name %s is illegal, it will be ignored",
				deviceRunMode, manuallySeparateNPU)
			continue
		}
		manuallySeparateNPUStrs := strings.Split(manuallySeparateNPU, "-")
		if len(manuallySeparateNPUStrs) <= 1 {
			hwlog.RunLog.Warnf("manually seperate NPU split slice length(%d) less than 2",
				len(manuallySeparateNPUStrs))
			continue
		}
		phyIDStr := manuallySeparateNPUStrs[1]
		phyID, err := strconv.Atoi(phyIDStr)
		if err != nil {
			hwlog.RunLog.Warnf("failed to convert %v string type to int type, error: %v", phyIDStr, err)
			return phyIDs
		}

		phyIDs = append(phyIDs, common.PhyId(phyID))
	}
	return phyIDs
}

// GetUpgradeFaultReasonFromDeviceInfo returns the UpgradeFaultReason from device info
func (ki *ClientK8s) GetUpgradeFaultReasonFromDeviceInfo(
	deviceInfo *v1.ConfigMap) (common.UpgradeFaultReasonMap[common.PhyId], error) {
	reasonStr, ok := deviceInfo.Data[common.UpgradeFaultReasonKey]
	if !ok {
		err := fmt.Errorf("GetUpgradeFaultReasonFromDeviceInfo failed")
		return nil, err
	}
	reasonCm, err := common.StringToReasonCm(reasonStr)
	if err != nil {
		return nil, fmt.Errorf("GetUpgradeFaultReasonFromDeviceInfo failed err: %v", err)
	}
	return reasonCm, nil
}

// WriteDeviceInfoDataIntoCM write deviceinfo into config map
func (ki *ClientK8s) WriteDeviceInfoDataIntoCM(nodeDeviceData *common.NodeDeviceInfoCache, manuallySeparateNPU string,
	switchInfo common.SwitchFaultInfo, dpuInfo common.DpuInfo, reasonCm string) (*common.NodeDeviceInfoCache, error) {
	nodeDeviceData.CheckCode = common.MakeDataHash(nodeDeviceData.DeviceInfo)
	var data, switchData, dpuData []byte
	dpuOpen := !reflect.DeepEqual(dpuInfo, common.DpuInfo{})
	if data = common.MarshalData(nodeDeviceData); len(data) == 0 {
		return nil, fmt.Errorf("marshal nodeDeviceData failed")
	}
	if switchData = common.MarshalData(switchInfo); len(switchData) == 0 {
		return nil, fmt.Errorf("marshal switchDeviceData failed")
	}
	deviceInfoCM := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ki.DeviceInfoName,
			Namespace: api.KubeNS,
			Labels:    map[string]string{api.CIMCMLabelKey: common.CmConsumerValue},
		},
	}
	switch common.ParamOption.RealCardType {
	case api.Ascend910A5:
		deviceInfoCM.Data = map[string]string{
			api.DeviceInfoCMDataKey:                   string(data),
			api.SwitchInfoCMDataKey:                   string(switchData),
			common.UpgradeFaultReasonKey:              reasonCm,
			common.DeviceInfoCMManuallySeparateNPUKey: manuallySeparateNPU,
			common.DescriptionKey:                     common.DescriptionValue}
		if dpuOpen {
			if dpuData = common.MarshalData(dpuInfo); len(dpuData) == 0 {
				return nil, fmt.Errorf("marshal DpuDeviceData failed")
			}
			deviceInfoCM.Data[api.DpuInfoCMDataKey] = string(dpuData)
		}
	case api.Ascend910A3:
		deviceInfoCM.Data = map[string]string{
			api.DeviceInfoCMDataKey:                   string(data),
			api.SwitchInfoCMDataKey:                   string(switchData),
			common.DeviceInfoCMManuallySeparateNPUKey: manuallySeparateNPU,
			common.UpgradeFaultReasonKey:              reasonCm,
			common.DescriptionKey:                     common.DescriptionValue}
	default:
		deviceInfoCM.Data = map[string]string{
			api.DeviceInfoCMDataKey:                   string(data),
			common.DeviceInfoCMManuallySeparateNPUKey: manuallySeparateNPU,
			common.UpgradeFaultReasonKey:              reasonCm,
			common.DescriptionKey:                     common.DescriptionValue}
	}

	hwlog.RunLog.Debugf("write device info cache into cm: %s/%s.", deviceInfoCM.Namespace, deviceInfoCM.Name)
	if err := ki.createOrUpdateDeviceCM(deviceInfoCM); err != nil {
		return nil, err
	}
	return nodeDeviceData, nil
}

// WriteResetInfoDataIntoCM write reset info into config map
func (ki *ClientK8s) WriteResetInfoDataIntoCM(taskName string, namespace string,
	taskInfo *common.TaskResetInfo, needAddRetry bool) (*v1.ConfigMap, error) {
	oldCM, err := ki.GetConfigMap(common.ResetInfoCMNamePrefix+taskName, namespace)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get reset cm of task %s, err: %v", taskName, err)
		return nil, err
	}
	oldResetInfoData, ok := oldCM.Data[common.ResetInfoCMDataKey]
	if !ok {
		return nil, fmt.Errorf("invalid reset info data")
	}
	if strings.Contains(oldResetInfoData, common.IsolateError) && len(taskInfo.RankList) != 0 {
		return nil, fmt.Errorf("task should be rescheduled")
	}
	var oldTaskInfo common.TaskResetInfo
	err = json.Unmarshal([]byte(oldResetInfoData), &oldTaskInfo)
	if err != nil {
		hwlog.RunLog.Errorf("failed to unmarshal reset info data, err: %v", err)
		return nil, fmt.Errorf("failed to unmarshal reset info data, err: %v", err)
	}
	retryTime := oldTaskInfo.RetryTime
	if needAddRetry {
		retryTime = retryTime + 1
	}
	newTaskInfo := setNewTaskInfoWithHexString(taskInfo)
	newTaskInfo.UpdateTime = time.Now().Unix()
	newTaskInfo.RetryTime = retryTime
	checkCode := common.MakeDataHash(newTaskInfo)
	var data []byte
	if data = common.MarshalData(newTaskInfo); len(data) == 0 {
		return nil, fmt.Errorf("marshal task reset data failed")
	}
	resetInfoCM := &v1.ConfigMap{
		TypeMeta:   oldCM.TypeMeta,
		ObjectMeta: oldCM.ObjectMeta,
		Data: map[string]string{
			common.ResetInfoCMDataKey:      string(data),
			common.ResetInfoCMCheckCodeKey: checkCode,
		},
	}
	oldRestartType, ok := oldCM.Data[common.ResetInfoTypeKey]
	if ok {
		resetInfoCM.Data[common.ResetInfoTypeKey] = oldRestartType
	}
	if needAddRetry {
		resetInfoCM.Data[common.ResetInfoTypeKey] = common.HotResetRestartType
	}

	hwlog.RunLog.Debugf("write reset info cache into cm: %s/%s.", resetInfoCM.Namespace, resetInfoCM.Name)
	return ki.UpdateConfigMap(resetInfoCM)
}

func setNewTaskInfoWithHexString(taskInfo *common.TaskResetInfo) *common.TaskResetInfo {
	var newTaskInfo common.TaskResetInfo
	for _, deviceInfo := range taskInfo.RankList {
		newDeviceInfo := *deviceInfo
		newDeviceInfo.ErrorCodeHex = strings.ToUpper(common.Int64Tool.ToHexString(newDeviceInfo.ErrorCode))
		newDeviceInfo.ErrorCode = []int64{}
		newTaskInfo.RankList = append(newTaskInfo.RankList, &newDeviceInfo)
	}
	if newTaskInfo.RankList == nil {
		newTaskInfo.RankList = make([]*common.TaskDevInfo, 0)
	}
	return &newTaskInfo
}

// WriteFaultInfoDataIntoCM write fault info into config map
func (ki *ClientK8s) WriteFaultInfoDataIntoCM(taskName string, namespace string,
	faultInfo *common.TaskFaultInfo) (*v1.ConfigMap, error) {
	oldCM, err := ki.GetConfigMap(common.FaultInfoCMNamePrefix+taskName, namespace)
	if err != nil {
		if errors.IsNotFound(err) {
			hwlog.RunLog.Infof("fault config map in task %s is not found", taskName)
			return nil, nil
		}
		hwlog.RunLog.Errorf("failed to get fault cm of task %s, err: %v", taskName, err)
		return nil, err
	}
	taskFaultInfo := &common.TaskFaultInfoCache{
		FaultInfo: faultInfo,
	}
	taskFaultInfo.FaultInfo.UpdateTime = time.Now().Unix()
	checkCode := common.MakeDataHash(taskFaultInfo.FaultInfo)
	var data []byte
	if data = common.MarshalData(taskFaultInfo.FaultInfo); len(data) == 0 {
		return nil, fmt.Errorf("marshal task reset data failed")
	}
	faultInfoCM := &v1.ConfigMap{
		TypeMeta:   oldCM.TypeMeta,
		ObjectMeta: oldCM.ObjectMeta,
		Data: map[string]string{
			common.FaultInfoCMDataKey:      string(data),
			common.FaultInfoCMCheckCodeKey: checkCode,
		},
	}

	hwlog.RunLog.Debugf("write fault info cache into cm: %s/%s.", faultInfoCM.Namespace, faultInfoCM.Name)
	return ki.UpdateConfigMap(faultInfoCM)
}

// AnnotationReset reset annotation and device info
func (ki *ClientK8s) AnnotationReset() error {
	curNode, err := ki.GetNode()
	if err != nil {
		hwlog.RunLog.Errorf("failed to get node, nodeName: %s, err: %v", ki.NodeName, err)
		return err
	}
	if curNode == nil {
		hwlog.RunLog.Error("invalid node")
		return fmt.Errorf("invalid node")
	}
	newNode := curNode.DeepCopy()
	ki.resetNodeAnnotations(newNode)
	for i := 0; i < common.RetryUpdateCount; i++ {
		if _, _, err = ki.PatchNodeState(curNode, newNode); err == nil {
			hwlog.RunLog.Infof("reset annotation success")
			return nil
		}
		hwlog.RunLog.Errorf("failed to patch volcano npu resource, times:%d", i+1)
		time.Sleep(time.Second)
		continue
	}
	hwlog.RunLog.Errorf("failed to patch volcano npu resource: %v", err)
	return err
}

// GetPodsUsedNpuByCommon get npu by status
func (ki *ClientK8s) GetPodsUsedNpuByCommon() sets.String {
	podList := ki.GetActivePodListCache()
	var useNpu = make([]string, 0)
	for _, pod := range podList {
		tmpNpu, ok := pod.Annotations[api.PodAnnotationAscendReal]
		if !ok || len(tmpNpu) == 0 || len(tmpNpu) > common.PodAnnotationMaxLength {
			continue
		}
		tmpNpuList := strings.Split(tmpNpu, common.CommaSepDev)
		if len(tmpNpuList) == 0 || len(tmpNpuList) > common.MaxDevicesNum {
			hwlog.RunLog.Warnf("invalid annotation, len is %d", len(tmpNpu))
			continue
		}
		useNpu = append(useNpu, tmpNpuList...)
		hwlog.RunLog.Debugf("pod Name: %s, getNPUByStatus vol : %#v", pod.Name, tmpNpu)
	}
	hwlog.RunLog.Debugf("get pods by cache from api-server, used NPU: %v", useNpu)
	return sets.NewString(useNpu...)
}

// GetPodsUsedNPUByKlt returns NPUs used by Pods
func (ki *ClientK8s) GetPodsUsedNPUByKlt() sets.String {
	podList, err := ki.getPodsByKltPort()
	if err != nil {
		hwlog.RunLog.ErrorfWithLimit(domainForKubeletConnectErr, os.Getenv(KubeletPortEnv),
			"get pods used NPU failed: %v", err)
		return ki.GetPodsUsedNpuByCommon()
	}
	usedNPU := make([]string, 0)
	for _, pod := range podList.Items {
		if err := common.CheckPodNameAndSpace(pod.GetName(), common.PodNameMaxLength); err != nil {
			hwlog.RunLog.Warnf("pod name syntax illegal, err: %v", err)
			continue
		}
		if err := common.CheckPodNameAndSpace(pod.GetNamespace(), common.PodNameSpaceMaxLength); err != nil {
			hwlog.RunLog.Warnf("pod namespace syntax illegal, err: %v", err)
			continue
		}
		if pod.Status.Phase == v1.PodFailed || pod.Status.Phase == v1.PodSucceeded {
			continue
		}
		realAllocTag := fmt.Sprintf("%s", api.PodAnnotationAscendReal)
		tmpNPU, ok := pod.Annotations[realAllocTag]
		if !ok || len(tmpNPU) == 0 || len(tmpNPU) > common.PodAnnotationMaxLength {
			continue
		}
		tmpNPUList := strings.Split(tmpNPU, common.CommaSepDev)
		if len(tmpNPUList) == 0 || len(tmpNPUList) > common.MaxDevicesNum {
			hwlog.RunLog.Warnf("invalid annotation, len is %d", len(tmpNPUList))
			continue
		}
		usedNPU = append(usedNPU, tmpNPUList...)
		hwlog.RunLog.Debugf("pod Name: %s, get real allocate npu by pod, tmpNPU: %v", pod.GetName(), tmpNPU)
	}
	hwlog.RunLog.Debugf("get pods by klt port, used NPU: %v", usedNPU)
	return sets.NewString(usedNPU...)
}

// GetNodeIp Get Node IP
func (ki *ClientK8s) GetNodeIp() (string, error) {
	node, err := ki.GetNode()
	if err != nil {
		return "", err
	}
	if len(node.Status.Addresses) > common.MaxPodLimit {
		hwlog.RunLog.Error("the number of node status in exceeds the upper limit")
		return "", fmt.Errorf("the number of node status in exceeds the upper limit")
	}
	var nodeIp string
	for _, addresses := range node.Status.Addresses {
		if addresses.Type == v1.NodeInternalIP && net.ParseIP(addresses.Address) != nil {
			nodeIp = addresses.Address
			break
		}
	}
	return nodeIp, nil
}
