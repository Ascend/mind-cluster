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

// Package manualfault counter for hardware frequency fault
package manualfault

import (
	"sync"
	"time"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/conf"
)

// Counter tn instance of FaultCounter
var Counter FaultCounter

// FaultCounter fault counter for hardware frequency fault
type FaultCounter struct {
	// key: node, value: dev fault info
	faults map[string]map[string]devFault
	mutex  sync.RWMutex
}

type devFault struct {
	// key: fault code, value: fault times
	fault map[string][]int64
}

// FaultInfo frequency fault info
type FaultInfo struct {
	NodeName    string
	DevName     string
	FaultCode   string
	ReceiveTime int64 // unit: millisecond
}

func init() {
	InitCounter()
}

// InitCounter init Counter
func InitCounter() {
	Counter = FaultCounter{
		faults: make(map[string]map[string]devFault),
		mutex:  sync.RWMutex{},
	}
}

func constructFaultInfo(fault *Fault) FaultInfo {
	return FaultInfo{
		NodeName:    fault.NodeName,
		DevName:     fault.DevName,
		FaultCode:   fault.Code,
		ReceiveTime: fault.ReceiveTime,
	}
}

// AddFault add fault to cache
func (c *FaultCounter) AddFault(input *Fault) {
	if input == nil {
		return
	}
	fault := constructFaultInfo(input)
	c.mutex.Lock()
	defer func() {
		c.mutex.Unlock()
		c.printCountInfo()
	}()
	nodeInfo, ok := c.faults[fault.NodeName]
	if !ok {
		if c.isReachFrequency([]int64{fault.ReceiveTime}) {
			c.dealFrequencyFault(fault)
			return
		}
		c.faults[fault.NodeName] = map[string]devFault{
			fault.DevName: {fault: map[string][]int64{fault.FaultCode: {fault.ReceiveTime}}},
		}
		return
	}
	devInfo, ok := nodeInfo[fault.DevName]
	if !ok {
		if c.isReachFrequency([]int64{fault.ReceiveTime}) {
			c.dealFrequencyFault(fault)
			return
		}
		nodeInfo[fault.DevName] = devFault{
			fault: map[string][]int64{fault.FaultCode: {fault.ReceiveTime}},
		}
		return
	}
	times, ok := devInfo.fault[fault.FaultCode]
	if !ok {
		if c.isReachFrequency([]int64{fault.ReceiveTime}) {
			c.dealFrequencyFault(fault)
			return
		}
		devInfo.fault[fault.FaultCode] = []int64{fault.ReceiveTime}
		return
	}
	times = append(times, fault.ReceiveTime)

	times = c.deleteExpiredFaultTime(times)
	if c.isReachFrequency(times) {
		c.dealFrequencyFault(fault)
		return
	}
	devInfo.fault[fault.FaultCode] = times
}

func (c *FaultCounter) dealFrequencyFault(fault FaultInfo) {
	hwlog.RunLog.Errorf("node: %s, dev: %s, code: %s, reach frequency threshold, set to manually separate",
		fault.NodeName, fault.DevName, fault.FaultCode)
	c.clearDevFault(fault.NodeName, fault.DevName, fault.FaultCode)
	FaultCmInfo.AddSeparateDev(fault)
}

func (c *FaultCounter) printCountInfo() {
	if len(c.faults) == 0 {
		hwlog.RunLog.Info("faults count is empty")
		return
	}
	hwlog.RunLog.Info("begin record faults count")
	for node, nodeInfo := range c.faults {
		for dev, devInfo := range nodeInfo {
			for code, times := range devInfo.fault {
				hwlog.RunLog.Infof("node: %s, dev: %s, code: %s, times: %d", node, dev, code, len(times))
			}
		}
	}
	hwlog.RunLog.Info("record faults count end")
}

// ClearDevFault safe clear dev fault
func (c *FaultCounter) ClearDevFault(node, devId, code string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.clearDevFault(node, devId, code)
}

// clearDevFault clear dev fault without lock
func (c *FaultCounter) clearDevFault(node, devId, code string) {
	dev, ok := c.faults[node]
	if !ok {
		return
	}
	devInfo, ok := dev[devId]
	if !ok {
		return
	}
	delete(devInfo.fault, code)
	if len(devInfo.fault) == 0 {
		delete(c.faults[node], devId)
	}
	if len(c.faults[node]) == 0 {
		delete(c.faults, node)
	}
}

// only retain faults in last MaxFaultWindowHours without lock
func (c *FaultCounter) deleteExpiredFaultTime(times []int64) []int64 {
	var filtered []int64
	threshold := int64(conf.MaxFaultWindowHours * constant.HoursToMilliseconds)
	for _, t := range times {
		if time.Now().UnixMilli()-t <= threshold {
			filtered = append(filtered, t)
		}
	}
	return filtered
}

func (c *FaultCounter) isReachFrequency(faultTimes []int64) bool {
	if len(faultTimes) < conf.GetSeparateThreshold() {
		return false
	}
	if conf.GetSeparateThreshold() == conf.MinFaultThreshold {
		return true
	}

	recentTimes := faultTimes[len(faultTimes)-conf.GetSeparateThreshold():]
	if len(recentTimes) < conf.GetSeparateThreshold() {
		return false
	}
	if recentTimes[conf.GetSeparateThreshold()-1]-recentTimes[0] < conf.GetSeparateWindow() {
		return true
	}
	return false
}
