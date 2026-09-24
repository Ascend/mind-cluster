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

// Package silentfault silent fault detection application logic
package silentfault

import (
	"context"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/hardwarefault"
	"clusterd/pkg/domain/manualfault"
	"clusterd/pkg/domain/silentfault"
)

var (
	pendingCache     = silentfault.NewPendingCache()
	firstFaultMgr    = silentfault.NewFirstFaultMgr()
	faultHardwareLog = hardwarefault.NewFaultHardwareLog()
	// SilentFaultCmInfo silent fault result cache (message-driven, updated by the manualfault-side handler)
	SilentFaultCmInfo = silentfault.NewSilentFaultCmInfo()
)

// ResetCache clears the silent-fault-side detection caches and the result cache (used for cleanup when the switch is off)
func ResetCache() {
	hwlog.RunLog.Infof("reset silent fault caches: pendingCache + firstFaultMgr + silentFaultCmInfo")
	pendingCache.ResetEvents()
	pendingCache.ResetProcessed()
	firstFaultMgr.Reset()
	SilentFaultCmInfo.Reset()
}

// ClearDetectionCache clears only the detection-side caches (pending events + node first-fault events)
// when the silent fault config changed, so already-collected events are re-collected/re-validated
// against the new config instead of being judged with the old config. The result cache
// (SilentFaultCmInfo) is intentionally kept: already-determined silent faults remain.
func ClearDetectionCache() {
	hwlog.RunLog.Infof("clear silent fault detection caches on config change: pendingCache + firstFaultMgr")
	pendingCache.ResetEvents()
	pendingCache.ResetProcessed()
	firstFaultMgr.Reset()
}

// StartFaultHardwareLogCleanup starts the daily cleanup of the shared hardware fault log in its
// own goroutine. The cleanup logic lives in the public hardwarefault package.
func StartFaultHardwareLogCleanup(ctx context.Context) {
	faultHardwareLog.RunDailyCleanup(ctx)
}

// SilentFaultIsEmpty reports whether the silent fault result cache is empty
func SilentFaultIsEmpty() bool {
	return SilentFaultCmInfo.Len() == 0
}

// GetSilentFaultCmInfoForMerge returns a manualfault.NodeCmInfo snapshot of the silent fault result cache for merging into the CM
func GetSilentFaultCmInfoForMerge() map[string]manualfault.NodeCmInfo {
	res := make(map[string]manualfault.NodeCmInfo)
	for node, info := range SilentFaultCmInfo.GetAll() {
		detail := make(map[string][]manualfault.DevCmInfo, len(info.DevList))
		for _, dev := range info.DevList {
			detail[dev] = []manualfault.DevCmInfo{{
				FaultCode:        info.FaultCode,
				FaultLevel:       constant.SilentFault,
				LastSeparateTime: info.LastSeparateTime,
			}}
		}
		res[node] = manualfault.NodeCmInfo{
			Total:  append([]string{}, info.DevList...),
			Detail: detail,
		}
	}
	return res
}
