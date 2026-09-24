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

package silentfault

import (
	"strings"
	"sync"

	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/conf"
)

// FaultRecorderProcessor hardware fault timeline collector; the same instance is registered at the tail of the Device/Node/Switch processing chains
var FaultRecorderProcessor = newFaultRecorderProcessor()

type faultRecorderProcessor struct {
	nodeDeviceCmMap map[string]*constant.AdvanceDeviceFaultCm // previous-round device snapshot
	mutex           sync.RWMutex
}

func newFaultRecorderProcessor() *faultRecorderProcessor {
	return &faultRecorderProcessor{
		nodeDeviceCmMap: make(map[string]*constant.AdvanceDeviceFaultCm),
		mutex:           sync.RWMutex{},
	}
}

// Process dispatches by payload type and records the hardware fault timeline
func (r *faultRecorderProcessor) Process(info any) any {
	if !conf.GetSilentFaultEnabled() {
		return info
	}
	switch content := info.(type) {
	case constant.OneConfigmapContent[*constant.AdvanceDeviceFaultCm]:
		r.recordDeviceFaults(content.AllConfigmap)
	case constant.OneConfigmapContent[*constant.NodeInfo]:
		r.recordNodeFaults(content.AllConfigmap)
	case constant.OneConfigmapContent[*constant.SwitchInfo]:
		r.recordSwitchFaults(content.AllConfigmap)
	}
	return info
}

// recordDeviceFaults incrementally records chip-level and injected public faults (called at the tail of the DeviceCenter chain)
func (r *faultRecorderProcessor) recordDeviceFaults(devCms map[string]*constant.AdvanceDeviceFaultCm) {
	r.mutex.Lock()
	oldMap := r.nodeDeviceCmMap
	r.nodeDeviceCmMap = devCms
	r.mutex.Unlock()

	for nodeName, devCm := range devCms {
		fallback := devCm.UpdateTime
		old, ok := oldMap[nodeName]
		if !ok {
			r.recordAll(nodeName, devCm.FaultDeviceList, fallback)
			continue
		}
		if old.IsSame(devCm) {
			continue
		}
		r.recordIncrement(nodeName, old.FaultDeviceList, devCm.FaultDeviceList, fallback)
	}
}

// recordAll performs the initial full recording of fault times
func (r *faultRecorderProcessor) recordAll(node string, newFaults map[string][]constant.DeviceFault, fallback int64) {
	for _, f := range flattenFaults(newFaults) {
		if isSilentSelfFault(f) {
			continue
		}
		for _, t := range faultTimes(f, fallback) {
			faultHardwareLog.Append(node, t)
		}
	}
}

// recordIncrement appends newly appeared fault times to the hardware fault log (duplicate times are skipped; fallback is used when the time field is missing)
func (r *faultRecorderProcessor) recordIncrement(node string,
	oldFaults, newFaults map[string][]constant.DeviceFault, fallback int64) {
	oldTimes := collectFaultTimes(oldFaults, fallback)
	for _, f := range flattenFaults(newFaults) {
		if isSilentSelfFault(f) {
			continue
		}
		for _, t := range faultTimes(f, fallback) {
			if !oldTimes[t] {
				faultHardwareLog.Append(node, t)
			}
		}
	}
}

// recordNodeFaults appends node-level faults (heartbeat timeout, etc.) to the hardware fault log using NodeInfo.UpdateTime
func (r *faultRecorderProcessor) recordNodeFaults(nodeCms map[string]*constant.NodeInfo) {
	for nodeName, nodeCm := range nodeCms {
		if len(nodeCm.FaultDevList) == 0 {
			continue
		}
		faultHardwareLog.Append(nodeName, nodeCm.UpdateTime)
	}
}

// recordSwitchFaults appends switch-level faults (RoCE network/link faults, etc.) to the hardware fault log
func (r *faultRecorderProcessor) recordSwitchFaults(switchCms map[string]*constant.SwitchInfo) {
	for cmName, switchCm := range switchCms {
		if len(switchCm.FaultTimeAndLevelMap) == 0 {
			continue
		}
		node := strings.TrimPrefix(cmName, constant.SwitchInfoPrefix)
		for _, tl := range switchCm.FaultTimeAndLevelMap {
			t := tl.FaultTime
			if t == 0 {
				t = tl.FaultReceivedTime
			}
			if t == 0 {
				t = switchCm.UpdateTime
			}
			if t > 0 {
				faultHardwareLog.Append(node, t)
			}
		}
	}
}

// isSilentSelfFault reports whether the fault is a silent-fault self-injected entry, which is not used as hardware fault evidence
func isSilentSelfFault(f constant.DeviceFault) bool {
	return f.FaultLevel == constant.SilentFault
}

// faultTimes returns the fault's occurrence times normalized to milliseconds by iterating
// FaultTimeAndLevelMap and skipping empty fault-code keys. When no valid time exists, fallback is returned.
func faultTimes(f constant.DeviceFault, fallback int64) []int64 {
	times := make([]int64, 0, len(f.FaultTimeAndLevelMap))
	for code, tl := range f.FaultTimeAndLevelMap {
		if code == "" {
			continue
		}
		if t := normalizeFaultTime(f.FaultType, tl); t > 0 {
			times = append(times, t)
		}
	}
	if len(times) == 0 && fallback > 0 {
		times = append(times, fallback)
	}
	return times
}

// normalizeFaultTime prefers FaultTime, converting PublicFault seconds to milliseconds;
// it falls back to FaultReceivedTime when FaultTime is 0.
func normalizeFaultTime(faultType string, tl constant.FaultTimeAndLevel) int64 {
	if tl.FaultTime > 0 {
		if faultType == constant.PublicFaultType {
			return tl.FaultTime * constant.SecondsToMilliseconds
		}
		return tl.FaultTime
	}
	return tl.FaultReceivedTime
}

func collectFaultTimes(faults map[string][]constant.DeviceFault, fallback int64) map[int64]bool {
	res := make(map[int64]bool)
	for _, f := range flattenFaults(faults) {
		if isSilentSelfFault(f) {
			continue
		}
		for _, t := range faultTimes(f, fallback) {
			res[t] = true
		}
	}
	return res
}

func flattenFaults(faults map[string][]constant.DeviceFault) []constant.DeviceFault {
	var res []constant.DeviceFault
	for _, list := range faults {
		res = append(res, list...)
	}
	return res
}
