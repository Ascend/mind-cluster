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
	"time"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/conf"
	"clusterd/pkg/domain/silentfault"
)

// validatePending stage-1 periodic validation goroutine (fixed 60s): performs O-second closure validation and dispatches to stage 2
func validatePending() {
	if !conf.GetSilentFaultEnabled() {
		return
	}
	now := time.Now().Unix()
	hwWindowSeconds := conf.GetHwWindowSeconds()
	ms := int64(constant.SecondsToMilliseconds)

	for _, ev := range pendingCache.TakeDue(now) {
		// stale record: beyond the detect window + 1h retention, the [ts-O, ts+O] hardware fault window cannot be fully validated, so drop it as fallback
		if now-ev.Timestamp > conf.GetWindowSeconds()+int64(time.Hour/time.Second) {
			hwlog.RunLog.Warnf("drop stale reschedule event: job %s node %s ts %d",
				ev.JobID, ev.FailNode, ev.Timestamp)
			continue
		}
		// check whether a hardware fault exists within O seconds before/after (rule 3); the hardware fault log stores milliseconds, so convert the window
		if faultHardwareLog.HasFaultIn(ev.FailNode, (ev.Timestamp-hwWindowSeconds)*ms, (ev.Timestamp+hwWindowSeconds)*ms) {
			hwlog.RunLog.Infof("skip silent fault record: job %s node %s ts %d has hardware fault in window [%d, %d]",
				ev.JobID, ev.FailNode, ev.Timestamp, (ev.Timestamp-hwWindowSeconds)*ms, (ev.Timestamp+hwWindowSeconds)*ms)
			continue
		}
		// no fault -> dispatch to the first-error node recording module (stage 2)
		for _, node := range ev.TaskNodes {
			firstFaultMgr.AddEvent(node, &silentfault.FirstFaultEvent{
				JobID:     ev.JobID,
				IsFirst:   node == ev.FailNode,
				Timestamp: ev.Timestamp,
			})
		}
	}
}
