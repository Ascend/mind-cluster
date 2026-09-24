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
	"encoding/json"
	"strconv"
	"strings"

	"ascend-common/api/annotation"
	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/conf"
	"clusterd/pkg/domain/manualfault"
	domainpublicfault "clusterd/pkg/domain/publicfault"
	"clusterd/pkg/interface/kube"
)

// LoadSilentFaultCmInfo restores on restart: rebuilds the silent fault result cache from the restored
// public fault cache (statistic-fault-info). The device type is taken from clusterd-manual-info-cm
// (the already-persisted silent device name) first, then falls back to the node annotation
// huawei.com/npu.base-device-infos, so it does not depend on DeviceCenter being ready on startup.
func LoadSilentFaultCmInfo() {
	if !conf.GetSilentFaultEnabled() {
		hwlog.RunLog.Info("skip load silent fault from cache: silent fault switch is off")
		return
	}

	manualDevType := loadSilentDevTypeFromManualCm()

	pubFaults, _ := domainpublicfault.PubFaultCache.GetPubFaultsForCM()
	for nodeName, nodeFaults := range pubFaults {
		for _, nf := range nodeFaults {
			if nf.FaultLevel != constant.SilentFault {
				continue
			}
			devType := manualDevType[nodeName]
			if devType == "" {
				devType = resolveDevTypeFromAnnotation(nodeName)
			}
			devNames := buildDevNames(devType, nf.FaultDevIds)
			hwlog.RunLog.Infof("restore silent fault from cache: node %s, resource %s, faultId %s, devices %v",
				nodeName, nf.FaultResource, nf.FaultId, devNames)
			SilentFaultCmInfo.Restore(nodeName, nf.FaultId, nf.FaultResource, devNames,
				nf.FaultTime*constant.SecondsToMilliseconds)
		}
	}
}

// loadSilentDevTypeFromManualCm returns the device type of each silent-isolated node read from
// clusterd-manual-info-cm. Only one silent device name is needed per node to derive the type prefix.
func loadSilentDevTypeFromManualCm() map[string]string {
	res := map[string]string{}
	cm, err := manualfault.TryGetManualCm()
	if err != nil || cm == nil {
		return res
	}
	cmInfo, err := manualfault.ParseManualCm(cm)
	if err != nil {
		hwlog.RunLog.Errorf("parse manual cm failed when load silent dev type, error: %v", err)
		return res
	}
	for node, info := range cmInfo {
		if devType := silentDevTypeFromDetail(info); devType != "" {
			res[node] = devType
		}
	}
	return res
}

// silentDevTypeFromDetail derives the device type from the first silent-fault-level device name in the node detail.
func silentDevTypeFromDetail(info manualfault.NodeCmInfo) string {
	for dev, details := range info.Detail {
		for _, d := range details {
			if d.FaultLevel == constant.SilentFault {
				if devType := deviceTypeFromName(dev); devType != "" {
					return devType
				}
			}
		}
	}
	return ""
}

// resolveDevTypeFromAnnotation derives the device type from the node annotation
// huawei.com/npu.base-device-infos, whose value is a JSON map keyed by "<type>-<id>".
func resolveDevTypeFromAnnotation(node string) string {
	n := kube.GetNode(node)
	if n == nil || n.Annotations == nil {
		return ""
	}
	val := n.Annotations[annotation.NPUBaseDevInfosAnnotation]
	if val == "" {
		return ""
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal([]byte(val), &m); err != nil {
		hwlog.RunLog.Warnf("unmarshal annotation %s failed: %v", annotation.NPUBaseDevInfosAnnotation, err)
		return ""
	}
	for devName := range m {
		return deviceTypeFromName(devName)
	}
	return ""
}

// deviceTypeFromName extracts the "<type>" prefix from "<type>-<id>"; returns empty when there is no prefix.
func deviceTypeFromName(name string) string {
	if idx := strings.Index(name, constant.Minus); idx > 0 {
		return name[:idx]
	}
	return ""
}

// buildDevNames builds device names "<type>-<id>" from the device type and logical ids. When the
// device type is empty (still unavailable), it degrades to bare ids.
func buildDevNames(devType string, ids []int32) []string {
	names := make([]string, 0, len(ids))
	for _, id := range ids {
		if devType == "" {
			names = append(names, strconv.Itoa(int(id)))
			continue
		}
		names = append(names, devType+constant.Minus+strconv.Itoa(int(id)))
	}
	return names
}
