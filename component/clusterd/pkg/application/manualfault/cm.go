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

// Package manualfault process manual separate npu info
package manualfault

import (
	"context"
	"strconv"
	"time"

	v1 "k8s.io/api/core/v1"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	"clusterd/pkg/application/faultmanager"
	"clusterd/pkg/application/publicfault"
	"clusterd/pkg/application/silentfault"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/conf"
	"clusterd/pkg/domain/manualfault"
)

func init() {
	publicfault.RegisterSilentFaultHandler(handleSilentFaultEvent)
}

// ProcessManuSep process manually separate npu info
func ProcessManuSep(ctx context.Context) {
	const updateCmInterval = 15 * time.Second
	ticker := time.NewTicker(updateCmInterval)
	defer ticker.Stop()

	for {
		select {
		case _, ok := <-ctx.Done():
			if !ok {
				hwlog.RunLog.Info("catch stop signal channel closed")
			}
			hwlog.RunLog.Infof("received stop signal: %v", ctx.Err())
			return
		case <-ticker.C:
			if manualfault.FaultCmInfo.Len() == 0 && silentfault.SilentFaultIsEmpty() {
				manualfault.DeleteManualCm()
				continue
			}
			cm, err := manualfault.TryGetManualCm()
			if err != nil {
				hwlog.RunLog.Errorf("get cm <%s/%s> failed, skip this round to avoid overwriting user deletion: %v",
					api.ClusterNS, constant.ManualDevInfoCmName, err)
				continue
			}
			checkManualDiffAndDelete(cm)
			releaseManualFault()
			releaseSilentFault(cm)
			updateManualCm(cm)
		}
	}
}

// checkManualDiffAndDelete check diff from cache and cm. delete the dev deleted from cm form the cache synchronously
func checkManualDiffAndDelete(cm *v1.ConfigMap) {
	manualDeleted := getManualDeletedDev(cm)
	for nodeName, info := range manualDeleted {
		for _, devId := range info {
			hwlog.RunLog.Errorf("node: %s, dev: %s is manually delete from cm, so delete from cache synchronously",
				nodeName, devId)
			manualfault.Counter.ClearDevFaults(nodeName, devId)
			manualfault.FaultCmInfo.DeleteSeparateDev(nodeName, devId)
		}
	}
}

// getManualDeletedDev delete manually separate npu in cm, data resource is cache
func getManualDeletedDev(cm *v1.ConfigMap) map[string][]string {
	lastSep := manualfault.GetSepNPUByLastCmInfo()
	currentSep := manualfault.GetSepNPUByCurrentCmInfo(cm)
	return utils.GetItemInANotInB(lastSep, currentSep)
}

func releaseManualFault() {
	if !conf.IsReleaseEnable() {
		return
	}

	nodeInfo, err := manualfault.FaultCmInfo.DeepCopy()
	if err != nil {
		hwlog.RunLog.Errorf("deep copy fault cm info failed, error: %v", err)
		return
	}
	for nodeName, info := range nodeInfo {
		for dev, devInfo := range info.Detail {
			doReleaseManualFault(nodeName, dev, devInfo)
		}
	}
}

func doReleaseManualFault(nodeName, dev string, devInfo []manualfault.DevCmInfo) {
	for _, cmInfo := range devInfo {
		if time.Now().UnixMilli()-cmInfo.LastSeparateTime >= conf.GetReleaseDuration() {
			hwlog.RunLog.Infof("node: %s, dev: %s, code: %s has been reached release time, released it",
				nodeName, dev, cmInfo.FaultCode)
			manualfault.Counter.ClearDevFault(nodeName, dev, cmInfo.FaultCode)
			manualfault.FaultCmInfo.DeleteDevCode(nodeName, dev, cmInfo.FaultCode)
			continue
		}
	}
}

// updateManualCm assembles the manual isolation and silent fault caches and writes them merged into clusterd-manual-info-cm.
// The assembly (including source awareness) lives in the application layer; domain's UpdateOrCreateManualCm only performs the raw write.
func updateManualCm(cm *v1.ConfigMap) {
	current, err := manualfault.FaultCmInfo.DeepCopy()
	if err != nil {
		hwlog.RunLog.Errorf("deep copy fault cm info failed, error: %v", err)
		return
	}
	manualfault.MergeNodeCmInfoMaps(current, silentfault.GetSilentFaultCmInfoForMerge())

	// Carry the resourceVersion read at the start of this tick for optimistic concurrency: if the cm
	// changed after that read (e.g. the user deleted all cards), the update conflicts and is skipped
	// instead of overwriting the deletion; the next tick re-reads and releases correctly.
	resourceVersion := ""
	if cm != nil {
		resourceVersion = cm.ResourceVersion
	}
	manualfault.UpdateOrCreateManualCm(current, resourceVersion)
}

// LoadManualCmInfo load manually separate npu info from configmap
func LoadManualCmInfo() {
	cm, err := manualfault.TryGetManualCm()
	if err != nil {
		hwlog.RunLog.Errorf("load cm <%s/%s> err: %v", api.ClusterNS, constant.ManualDevInfoCmName, err)
		return
	}
	if cm == nil {
		hwlog.RunLog.Infof("manually separate npu cm <%s/%s> is not found", api.ClusterNS, constant.ManualDevInfoCmName)
		return
	}
	cmInfo, err := manualfault.ParseManualCm(cm)
	if err != nil {
		hwlog.RunLog.Errorf("parse separate npu cm failed, error: %v", err)
		return
	}
	// load only manual isolation entries and drop silent fault entries to avoid wrong injection by ManualFaultProcessor
	cmInfo = filterSilentFaultNodes(cmInfo)
	manualfault.FaultCmInfo.SetNodeInfo(cmInfo)
	hwlog.RunLog.Info("save manually separate npu info to cache success")
}

// filterSilentFaultNodes deletes entries whose FaultLevel is SilentFault, keeping only manual isolation entries
func filterSilentFaultNodes(cmInfo map[string]manualfault.NodeCmInfo) map[string]manualfault.NodeCmInfo {
	for node, info := range cmInfo {
		for dev, details := range info.Detail {
			kept := details[:0]
			for _, d := range details {
				if d.FaultLevel == constant.SilentFault {
					continue
				}
				kept = append(kept, d)
			}
			if len(kept) == 0 {
				delete(info.Detail, dev)
				info.Total = utils.Remove(info.Total, dev)
				continue
			}
			info.Detail[dev] = kept
		}
		if len(info.Detail) == 0 {
			delete(cmInfo, node)
		} else {
			cmInfo[node] = info
		}
	}
	return cmInfo
}

// handleSilentFaultEvent receives silent-fault-level messages routed by PubFaultCollector and updates SilentFaultCmInfo
func handleSilentFaultEvent(event publicfault.SilentFaultEvent) {
	if !conf.GetSilentFaultEnabled() {
		return // double guard short-circuit
	}
	faultKey := event.Resource + event.FaultId
	switch event.Assertion {
	case constant.AssertionOccur, constant.AssertionOnce:
		devNames := resolveDevNames(event.NodeName, event.DevIds)
		hwlog.RunLog.Infof("handle silent fault occur event: node %s, faultKey %s, assertion %s, devices %v",
			event.NodeName, faultKey, event.Assertion, devNames)
		silentfault.SilentFaultCmInfo.Upsert(event.NodeName, event.FaultId, event.Resource, devNames)
	case constant.AssertionRecover:
		hwlog.RunLog.Infof("handle silent fault recover event: node %s, faultKey %s", event.NodeName, faultKey)
		silentfault.SilentFaultCmInfo.RemoveSource(event.NodeName, faultKey)
	}
}

// resolveDevNames converts the device logical IDs in the message back to device names
// ("<DeviceType>-<id>"). The device type comes from the node's latest device CM; when it is
// unavailable (e.g. DeviceCenter not ready on restart), the name degrades to the bare id
// ("<id>") so the persisted silent fault entry keeps a stable, parseable identifier.
func resolveDevNames(node string, devIds []int32) []string {
	devType := getDeviceType(node)
	names := make([]string, 0, len(devIds))
	for _, id := range devIds {
		if devType == "" {
			names = append(names, strconv.Itoa(int(id)))
			continue
		}
		names = append(names, devType+constant.Minus+strconv.Itoa(int(id)))
	}
	return names
}

// getDeviceType returns the node's device type; returns empty when the node has no device info.
func getDeviceType(node string) string {
	devCm, ok := faultmanager.QueryDeviceInfoToReport()[node]
	if !ok || devCm == nil {
		return ""
	}
	return devCm.DeviceType
}

// releaseSilentFault silent fault release detection (runs in the manualfault 15s ticker, not reusing manual's releaseManualFault).
// The current cm is read once at the tick start and passed in, so the silent-release diff always runs on the
// same snapshot as the subsequent cm rewrite, preventing an uncorrected silent cache from overwriting user deletion.
func releaseSilentFault(cm *v1.ConfigMap) {
	if !conf.GetSilentFaultEnabled() {
		return
	}
	now := time.Now().UnixMilli()

	// 1) auto release: LastSeparateTime + fault_free_seconds expired
	for _, node := range silentfault.SilentFaultCmInfo.Expired(now, conf.GetSilentReleaseSeconds()) {
		sendSilentFaultRecover(node, "auto release")
	}

	// 2) manual removal: diff clusterd-manual-info-cm to detect silent fault entries deleted by the user
	for _, node := range diffManuallyDeletedSilent(cm) {
		sendSilentFaultRecover(node, "manually deleted")
	}
}

// sendSilentFaultRecover sends a recover message for each active source of the node (through the common fault entry)
func sendSilentFaultRecover(node, reason string) {
	hwlog.RunLog.Infof("silent fault on node %s removed, reason: %s", node, reason)
	info, ok := silentfault.SilentFaultCmInfo.Get(node)
	if !ok {
		return
	}
	now := time.Now()
	for _, src := range info.Sources {
		pubFaultInfo := buildSilentFaultRecoverMsg(src.Resource, src.FaultId, node, now)
		if err := publicfault.PubFaultCollector(pubFaultInfo); err != nil {
			hwlog.RunLog.Errorf("send silent fault recover message failed, node %s, error: %v", node, err)
		}
	}
}

// buildSilentFaultRecoverMsg builds a standard silent fault recover message (assertion=recover, fields satisfy checker constraints)
func buildSilentFaultRecoverMsg(resource, faultId, node string, now time.Time) *api.PubFaultInfo {
	return &api.PubFaultInfo{
		Id:        constant.SilentFaultRecoverMsgIdPrefix + node + constant.Minus + strconv.FormatInt(now.UnixMilli(), constant.FormatBase),
		TimeStamp: now.UnixMilli(),
		Version:   constant.PubFaultVersion,
		Resource:  resource,
		Faults: []api.Fault{{
			FaultId:   faultId,
			FaultType: constant.FaultTypeNPU,
			FaultCode: constant.SilentFaultCode,
			FaultTime: now.UnixMilli(),
			Assertion: constant.AssertionRecover,
			Influence: []api.Influence{{NodeName: node, DeviceIds: []int32{0}}},
		}},
	}
}

// diffManuallyDeletedSilent detects silent fault entries deleted by the user in clusterd-manual-info-cm.
// Silent fault isolation is node-level, so only deleting ALL cards of a node counts as a manual release;
// deleting a subset keeps the node isolated and lets updateManualCm write the full node back.
// It diffs the last-written cm snapshot (manualfault.LastCmInfo) against the current cm, so a silent
// fault that has not been written yet is not mistaken for a manual deletion.
func diffManuallyDeletedSilent(cm *v1.ConfigMap) []string {
	lastCm := manualfault.LastCmInfo
	lastNodes := silentNodesIn(lastCm)
	if len(lastNodes) == 0 {
		return nil
	}
	curInfo := map[string]manualfault.NodeCmInfo{}
	if cm != nil {
		cmInfo, err := manualfault.ParseManualCm(cm)
		if err != nil {
			hwlog.RunLog.Errorf("parse manual cm failed when diff silent fault, error: %v", err)
			return nil
		}
		curInfo = cmInfo
	}
	var deleted []string
	for _, node := range lastNodes {
		lastTotal := len(lastCm[node].Total)
		curTotal := 0
		if info, ok := curInfo[node]; ok {
			curTotal = len(info.Total)
		}
		if curTotal == 0 {
			// all cards removed from Total: manual release
			hwlog.RunLog.Infof("silent fault node %s: all %d cards removed from cm Total, treat as manual release",
				node, lastTotal)
			deleted = append(deleted, node)
		} else if curTotal < lastTotal {
			// only a subset removed: keep isolation, updateManualCm rewrites the full node
			hwlog.RunLog.Infof("silent fault node %s: %d of %d cards removed from cm Total, keep isolation and rewrite full node",
				node, lastTotal-curTotal, lastTotal)
		}
	}
	return deleted
}

// silentNodesIn returns the nodes that had silent fault isolation when the given cm was written
func silentNodesIn(cmInfo map[string]manualfault.NodeCmInfo) []string {
	nodes := make([]string, 0)
	for node, info := range cmInfo {
		if hasSilentEntry(info) {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

// hasSilentEntry reports whether the node has any silent-fault-level entry, used to identify silent nodes
func hasSilentEntry(info manualfault.NodeCmInfo) bool {
	for _, details := range info.Detail {
		for _, d := range details {
			if d.FaultLevel == constant.SilentFault {
				return true
			}
		}
	}
	return false
}
