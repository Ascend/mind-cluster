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
	"encoding/json"
	"sort"
	"strconv"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"

	"ascend-common/api"
	"ascend-common/api/annotation"
	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/application/publicfault"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/conf"
	domainpublicfault "clusterd/pkg/domain/publicfault"
	"clusterd/pkg/interface/kube"
)

// detectFirstFault stage-2 periodic detection (fixed 60s): checks for C consecutive first errors in reverse order
func detectFirstFault() {
	if !conf.GetSilentFaultEnabled() {
		return
	}
	now := time.Now().Unix()
	c := conf.GetConsecutiveTimes()
	m := conf.GetWindowSeconds()

	for _, node := range firstFaultMgr.Nodes() {
		events := firstFaultMgr.GetEvents(node)
		run, hitJobs := 0, make([]string, 0, c)
		hit := false
		for i := len(events) - 1; i >= 0; i-- {
			if events[i].Timestamp < now-m {
				break
			}
			if !events[i].IsFirst {
				run, hitJobs = 0, hitJobs[:0]
				continue
			}
			run++
			hitJobs = append(hitJobs, events[i].JobID)
			if run >= c {
				hit = true
				break
			}
		}
		if hit {
			writeSilentFault(node, hitJobs)
			firstFaultMgr.ClearNode(node)
		}
	}
	firstFaultMgr.PruneExpiredAll(now, m+constant.SilentFaultFirstFaultExtraRetentionSec)
}

// writeSilentFault builds a standard occur message after a hit and sends it through the common fault entry (no cache is written directly)
func writeSilentFault(node string, hitJobs []string) {
	// switch off: send no message (the detection entry already short-circuits; this is a double guard)
	if !conf.GetSilentFaultEnabled() {
		return
	}
	level := domainpublicfault.GetFaultLevelByCode(constant.SilentFaultCode)
	if level == "" {
		hwlog.RunLog.Errorf("silent fault code %s is not configured in publicFaultConfiguration.json, "+
			"skip writing node %s", constant.SilentFaultCode, node)
		return
	}
	hwlog.RunLog.Infof("silent fault detected on node %s, hit jobs: %v", node, hitJobs)

	deviceIDs := nodeCardIDs(node)
	if len(deviceIDs) == 0 {
		hwlog.RunLog.Warnf("get device list of node %s failed, skip write silent fault", node)
		return
	}
	now := time.Now()
	faultID := constant.SilentFaultIdPrefix + node

	pubFaultInfo := &api.PubFaultInfo{
		Id:        constant.SilentFaultOccurMsgIdPrefix + node + constant.Minus + strconv.FormatInt(now.UnixMilli(), constant.FormatBase),
		TimeStamp: now.UnixMilli(),
		Version:   constant.PubFaultVersion,
		Resource:  constant.SilentFaultResource,
		Faults: []api.Fault{{
			FaultId:   faultID,
			FaultType: constant.FaultTypeNPU,
			FaultCode: constant.SilentFaultCode,
			FaultTime: now.UnixMilli(),
			Assertion: constant.AssertionOccur,
			Influence: []api.Influence{{
				NodeName:  node,
				DeviceIds: deviceIDs,
			}},
		}},
	}
	if err := publicfault.PubFaultCollector(pubFaultInfo); err != nil {
		hwlog.RunLog.Errorf("send silent fault occur message failed, node %s, error: %v", node, err)
	}
}

// nodeCardIDs returns the node's physical device IDs read from the node annotation
// huawei.com/npu.base-device-infos (falling back to the deprecated baseDeviceInfos), whose value
// is a JSON map keyed by "<type>-<id>". It returns the "<id>" parts sorted in ascending order.
func nodeCardIDs(node string) []int32 {
	n := kube.GetNode(node)
	if n == nil || n.Annotations == nil {
		return nil
	}
	val := n.Annotations[annotation.NPUBaseDevInfosAnnotation]
	if val == "" {
		val = n.Annotations[annotation.BaseDevInfoAnnoDeprecated]
	}
	if val == "" {
		return nil
	}
	var devMap map[string]json.RawMessage
	if err := json.Unmarshal([]byte(val), &devMap); err != nil {
		hwlog.RunLog.Warnf("unmarshal node %s annotation %s failed: %v",
			node, annotation.NPUBaseDevInfosAnnotation, err)
		return nil
	}
	ids := make([]int32, 0, len(devMap))
	for devName := range devMap {
		if id := deviceIDFromName(devName); id >= 0 {
			ids = append(ids, id)
		}
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

// deviceIDFromName extracts the numeric "<id>" suffix from "<type>-<id>"; returns -1 when absent.
func deviceIDFromName(name string) int32 {
	idx := strings.LastIndex(name, constant.Minus)
	if idx < 0 || idx == len(name)-1 {
		return -1
	}
	id, err := strconv.ParseInt(name[idx+1:], 10, 32)
	if err != nil {
		return -1
	}
	return int32(id)
}

// nodeTotalCards returns the node's total physical card count of the given resource type
// from the K8s node capacity (the total card count regardless of health; falls back to
// allocatable). The count is stable regardless of job scheduling.
func nodeTotalCards(node, resourceType string) int {
	if resourceType == "" {
		return 0
	}
	n := kube.GetNode(node)
	if n == nil || n.Status.Capacity == nil {
		return 0
	}
	resName := corev1.ResourceName(api.ResourceNamePrefix + resourceType)
	quant := n.Status.Capacity[resName]
	return int(quant.Value())
}
