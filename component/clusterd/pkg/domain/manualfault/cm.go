/* Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package manualfault manual separate npu info cache
package manualfault

import (
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/interface/kube"
)

const (
	manualDevInfoCmName = "clusterd-manual-info-cm"
)

// LastCmInfo last cm info
var LastCmInfo map[string]nodeCmInfo

// FaultCmInfo an instance of manual fault cm cache
var FaultCmInfo Cache

// Cache cache for manual fault cm
type Cache struct {
	manualInfo map[string]nodeCmInfo
	mutex      sync.RWMutex
}

// nodeCmInfo total and detail info for dev is consistent
type nodeCmInfo struct {
	Total []string
	// key: dev name, value: dev fault
	Detail map[string][]devCmInfo
}

type devCmInfo struct {
	FaultCode        string
	FaultLevel       string
	LastSeparateTime int64 // unit: millisecond
}

func init() {
	InitFaultCmInfo()
}

// InitFaultCmInfo init FaultCmInfo
func InitFaultCmInfo() {
	FaultCmInfo = Cache{
		manualInfo: make(map[string]nodeCmInfo),
		mutex:      sync.RWMutex{},
	}
	LastCmInfo = make(map[string]nodeCmInfo)
}

// SetNodeInfo set node info
func (c *Cache) SetNodeInfo(nodeInfo map[string]nodeCmInfo) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if len(nodeInfo) == 0 {
		c.manualInfo = make(map[string]nodeCmInfo)
		return
	}
	c.manualInfo = nodeInfo
}

func (c *Cache) Len() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return len(c.manualInfo)
}

// DeepCopy deep copy node info
func (c *Cache) DeepCopy() (map[string]nodeCmInfo, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	result := new(map[string]nodeCmInfo)
	if err := util.DeepCopy(result, c.manualInfo); err != nil {
		return nil, err
	}
	return *result, nil
}

// AddSeparateDev add manually separate npu info to cache
func (c *Cache) AddSeparateDev(faultInfo FaultInfo) {
	devInfo := devCmInfo{
		FaultCode:        faultInfo.FaultCode,
		FaultLevel:       constant.ManuallySeparateNPU,
		LastSeparateTime: faultInfo.ReceiveTime,
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	var addMsg = fmt.Sprintf("node: %s, dev: %s, code: %s is not found in manual fault cache, add",
		faultInfo.NodeName, faultInfo.DevName, faultInfo.FaultCode)
	var updateMsg = fmt.Sprintf("node: %s, dev: %s, code: %s is found in manual fault cache, update last separate time",
		faultInfo.NodeName, faultInfo.DevName, faultInfo.FaultCode)

	info, ok := c.manualInfo[faultInfo.NodeName]
	if !ok {
		hwlog.RunLog.Infof(addMsg)
		c.manualInfo[faultInfo.NodeName] = nodeCmInfo{
			Total:  []string{faultInfo.DevName},
			Detail: map[string][]devCmInfo{faultInfo.DevName: {devInfo}},
		}
		return
	}

	if !utils.Contains(info.Total, faultInfo.DevName) {
		hwlog.RunLog.Infof(addMsg)
		info.Total = append(info.Total, faultInfo.DevName)
		info.Detail[faultInfo.DevName] = []devCmInfo{devInfo}
		c.manualInfo[faultInfo.NodeName] = info
		return
	}
	var found bool
	for idx, fault := range info.Detail[faultInfo.DevName] {
		if fault.FaultCode == faultInfo.FaultCode {
			found = true
			hwlog.RunLog.Infof(updateMsg)
			fault.LastSeparateTime = faultInfo.ReceiveTime
			// need to write it back
			info.Detail[faultInfo.DevName][idx] = fault
			break
		}
	}
	if !found {
		hwlog.RunLog.Infof(addMsg)
		info.Detail[faultInfo.DevName] = append(info.Detail[faultInfo.DevName], devInfo)
		c.manualInfo[faultInfo.NodeName] = info
	}
	return
}

// HasDevManualSep check if dev has manually separated npu
func (c *Cache) HasDevManualSep(nodeName, devName string) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	info, ok := c.manualInfo[nodeName]
	if !ok {
		return false
	}
	return utils.Contains(info.Total, devName)
}

// DeleteSeparateDev delete manually separate dev info from cache
func (c *Cache) DeleteSeparateDev(nodeName, devId string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	info, ok := c.manualInfo[nodeName]
	if !ok {
		return
	}
	info.Total = utils.Remove(info.Total, devId)
	if len(info.Total) == 0 {
		delete(c.manualInfo, nodeName)
		return
	}
	delete(info.Detail, devId)
	c.manualInfo[nodeName] = info
}

// DeleteDevCode delete manually separate dev info by fault code from cache
func (c *Cache) DeleteDevCode(nodeName, devId, faultCode string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	info, ok := c.manualInfo[nodeName]
	if !ok {
		return
	}
	devList, ok := info.Detail[devId]
	if ok {
		for idx, fault := range devList {
			if fault.FaultCode == faultCode {
				info.Detail[devId] = append(info.Detail[devId][:idx], info.Detail[devId][idx+1:]...)
				break
			}
		}
	}

	if len(info.Detail[devId]) == 0 {
		delete(info.Detail, devId)
		if len(info.Detail) == 0 {
			delete(c.manualInfo, nodeName)
			return
		}
		info.Total = utils.Remove(info.Total, devId)
	}
	c.manualInfo[nodeName] = info
}

// GetSepNPUByCurrentCmInfo get manually separate npu info from cm
func GetSepNPUByCurrentCmInfo(cm *v1.ConfigMap) map[string][]string {
	if cm == nil {
		// allows users to manually delete the cm
		return nil
	}
	cmInfo, err := ParseManualCm(cm)
	if err != nil {
		hwlog.RunLog.Errorf("parse manual cm failed, error: %v", err)
		return nil
	}
	return GetSeparateNPU(cmInfo)
}

func GetSepNPUByLastCmInfo() map[string][]string {
	return GetSeparateNPU(LastCmInfo)
}

// GetSeparateNPU get manually separate npu info from node info
func GetSeparateNPU(nodeInfo map[string]nodeCmInfo) map[string][]string {
	if len(nodeInfo) == 0 {
		return nil
	}
	var separateNPU = make(map[string][]string)
	for nodeName, info := range nodeInfo {
		for _, dev := range info.Total {
			separateNPU[nodeName] = append(separateNPU[nodeName], dev)
		}
	}
	return separateNPU
}

// ParseManualCm parse manually separate npu info from configmap
func ParseManualCm(cm *v1.ConfigMap) (map[string]nodeCmInfo, error) {
	nodeInfo := nodeCmInfo{}
	manualCmInfo := map[string]nodeCmInfo{}
	if cm.Data == nil {
		return nil, fmt.Errorf("cm has no data")
	}
	for node, info := range cm.Data {
		if err := json.Unmarshal([]byte(info), &nodeInfo); err != nil {
			return nil, fmt.Errorf("unmarshal configmap <%s/%s> key %s failed: %v", cm.Namespace, cm.Name, node, err)
		}
		manualCmInfo[node] = nodeInfo
	}
	return manualCmInfo, nil
}

// DeleteManualCm delete manually separate npu info configmap
func DeleteManualCm() {
	if err := kube.DeleteConfigMap(manualDevInfoCmName, api.ClusterNS); err != nil {
		// delete non-existent cm will be failed, need filter
		if !errors.IsNotFound(err) {
			hwlog.RunLog.Errorf("manually separate npu info is nill, delete configmap failed, error: %v", err)
			return
		}
		hwlog.RunLog.Debugf("cm<%s/%s> not found, ignore", api.ClusterNS, manualDevInfoCmName)
	}
	LastCmInfo = make(map[string]nodeCmInfo)
}

// TryGetManualCm try get manually info configmap
func TryGetManualCm() (*v1.ConfigMap, error) {
	const retryTime = 3
	const retryInterval = 500 * time.Millisecond
	var cm *v1.ConfigMap
	var err error
	for i := 0; i < retryTime; i++ {
		cm, err = GetManualCm()
		if err != nil {
			hwlog.RunLog.Errorf("get cm <%s/%s> info failed, error: %v, retry", api.ClusterNS, manualDevInfoCmName, err)
			time.Sleep(retryInterval)
			continue
		}
		break
	}
	return cm, err
}

// GetManualCm get manually separate npu info configmap
func GetManualCm() (*v1.ConfigMap, error) {
	cm, err := kube.GetConfigMap(manualDevInfoCmName, api.ClusterNS)
	if err != nil {
		// get non-existent cm will be failed, need filter
		if !errors.IsNotFound(err) {
			hwlog.RunLog.Errorf("manually separate npu info is nil, get cm <%s/%s> failed, err: %v", api.ClusterNS, manualDevInfoCmName, err)
			return nil, err
		}
		hwlog.RunLog.Debugf("cm <%s/%s> not found, ignore", api.ClusterNS, manualDevInfoCmName)
		return nil, nil
	}
	return cm, nil
}

// UpdateOrCreateManualCm update manually separate npu info configmap. if currentCmInfo is empty, delete cm
func UpdateOrCreateManualCm() {
	currentCmInfo, err := FaultCmInfo.DeepCopy()
	if err != nil {
		hwlog.RunLog.Errorf("deep copy fault cm info failed, error: %v", err)
		return
	}

	if len(currentCmInfo) == 0 {
		DeleteManualCm()
		return
	}

	for _, info := range currentCmInfo {
		sort.Strings(info.Total)
	}
	data := convertNodeInfoToCmData(currentCmInfo)
	if err := kube.UpdateOrCreateConfigMap(manualDevInfoCmName, api.ClusterNS, data, nil); err != nil {
		hwlog.RunLog.Errorf("manually separate npu info is nil, update configmap err: %v", err)
		return
	}

	LastCmInfo = currentCmInfo
}

func convertNodeInfoToCmData(cmInfo map[string]nodeCmInfo) map[string]string {
	cmData := make(map[string]string)
	for node, info := range cmInfo {
		data, err := json.Marshal(info)
		if err != nil {
			hwlog.RunLog.Errorf("marshal node info to json string failed, error: %v", err)
			continue
		}
		cmData[node] = string(data)
	}
	return cmData
}
