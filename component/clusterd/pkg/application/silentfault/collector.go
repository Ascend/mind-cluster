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

	v1 "k8s.io/api/core/v1"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/conf"
	"clusterd/pkg/domain/job"
	"clusterd/pkg/domain/silentfault"
)

// rescheduleReason set of reschedule reasons for a single job
type rescheduleReason struct {
	JobID                string             `json:"jobID"`
	JobUID               string             `json:"jobUID"`
	TotalRescheduleTimes int                `json:"totalRescheduleTimes"`
	RescheduleRecords    []RescheduleRecord `json:"rescheduleRecords"`
}

// RescheduleRecord one reschedule record
type RescheduleRecord struct {
	RescheduleTimeStamp int64            `json:"rescheduleTimeStamp"`
	ReasonOfTask        []RescheduleTask `json:"reasonOfTask"`
}

// RescheduleTask task that triggered the reschedule
type RescheduleTask struct {
	RescheduleReason string `json:"rescheduleReason"`
	PodName          string `json:"podName"`
	NodeName         string `json:"nodeName"`
}

// CmHandler informer callback that carries the full recent-reschedule-records
func CmHandler(oldCm, newCm *v1.ConfigMap, op string) {
	if !conf.GetSilentFaultEnabled() {
		return
	}
	if op == constant.DeleteOperator {
		hwlog.RunLog.Infof("reschedule reason cm deleted, clear processed dedup set, keep pending events")
		pendingCache.ResetProcessed()
		return
	}
	if op != constant.AddOperator && op != constant.UpdateOperator {
		return
	}
	if newCm == nil {
		return
	}
	data, ok := newCm.Data[constant.RescheduleReasonCmKey]
	if !ok || data == "" {
		return
	}
	var reasons map[string]rescheduleReason
	if err := json.Unmarshal([]byte(data), &reasons); err != nil {
		hwlog.RunLog.Warnf("unmarshal reschedule reason cm failed: %v", err)
		return
	}

	curJobs := make(map[string]struct{}, len(reasons))
	for _, reason := range reasons {
		curJobs[reason.JobUID] = struct{}{}
	}
	pendingCache.PruneOrphanProcessed(curJobs)

	for _, reason := range reasons {
		jobID := reason.JobUID
		pendingCache.PruneProcessed(jobID, collectRescheduleTs(reason.RescheduleRecords))
		for _, rr := range reason.RescheduleRecords {
			if pendingCache.HasProcessed(jobID, rr.RescheduleTimeStamp) {
				continue
			}
			pendingCache.MarkProcessed(jobID, rr.RescheduleTimeStamp)
			OnReschedule(jobID, rr)
		}
	}
}

// OnReschedule stage-1 collection: preprocess each new record
func OnReschedule(jobID string, rr RescheduleRecord) {
	if len(rr.ReasonOfTask) != 1 {
		hwlog.RunLog.Infof("skip silent fault record: job %s reschedule task count %d not equal 1",
			jobID, len(rr.ReasonOfTask))
		return
	}
	task := rr.ReasonOfTask[0]
	if task.RescheduleReason != constant.PodFailedReason {
		hwlog.RunLog.Infof("skip silent fault record: job %s reschedule reason %s not pod-failed",
			jobID, task.RescheduleReason)
		return
	}
	ev := buildPendingEvent(jobID, task.PodName, task.NodeName, rr.RescheduleTimeStamp)
	if ev == nil {
		hwlog.RunLog.Infof("skip silent fault record: job %s pod %s node %s build pending event failed",
			jobID, task.PodName, task.NodeName)
		return
	}
	pendingCache.Add(ev)
}

// buildPendingEvent builds a pending event (with the job node snapshot); returns nil when rule 1 is not satisfied
func buildPendingEvent(jobID, failPod, failNode string, ts int64) *silentfault.PendingEvent {
	jobInfo, ok := job.GetJobCache(jobID)
	if !ok {
		hwlog.RunLog.Infof("skip silent fault record: job %s not found in cache", jobID)
		return nil
	}
	if !jobCardsAtLeast(jobInfo, conf.GetMinTaskCards()) {
		hwlog.RunLog.Infof("skip silent fault record: job %s node %s timestamp %v not satisfy rule1 (cards below threshold)",
			jobID, failNode, ts)
		return nil
	}
	if !isWholeNodeJob(jobInfo, failPod) {
		hwlog.RunLog.Infof("skip silent fault record: job %s pod %s node %s timestamp %v not satisfy whole-node job",
			jobID, failPod, failNode, ts)
		return nil
	}
	nodes := make([]string, 0, len(jobInfo.PreServerList))
	for _, srv := range jobInfo.PreServerList {
		nodes = append(nodes, srv.ServerName)
	}
	if len(nodes) == 0 {
		hwlog.RunLog.Infof("skip silent fault record: job %s node %s has no task nodes", jobID, failNode)
		return nil
	}
	return &silentfault.PendingEvent{JobID: jobID, FailNode: failNode, TaskNodes: nodes, Timestamp: ts}
}

// jobCardsAtLeast reports whether the job's total card count is at least the threshold (rule 1).
// It reads PreServerList (the last completed rank-table snapshot) instead of the live JobRankTable,
// which fluctuates during pod rescheduling and would wrongly drop events between two failed pods.
func jobCardsAtLeast(jobInfo constant.JobInfo, minCards int) bool {
	cards := 0
	for _, srv := range jobInfo.PreServerList {
		cards += len(srv.DeviceList)
	}
	return cards >= minCards
}

// isWholeNodeJob reports whether the job is a whole-node NPU job: every NPU pod of the job
// occupies all cards of its own node, and the failed pod is one of those NPU pods. Pods that
// applied for no NPU are ignored; when the failed pod is empty or is not an NPU pod, the record
// is invalid. It reads PreServerList for the same reason as jobCardsAtLeast: the live rank table
// is unstable during rescheduling.
func isWholeNodeJob(jobInfo constant.JobInfo, failPod string) bool {
	if failPod == "" {
		return false
	}
	failedIsNPUPod := false
	for _, srv := range jobInfo.PreServerList {
		if len(srv.DeviceList) == 0 {
			continue
		}
		if len(srv.DeviceList) != nodeTotalCards(srv.ServerName, jobInfo.ResourceType) {
			return false
		}
		if srv.PodName == failPod {
			failedIsNPUPod = true
		}
	}
	return failedIsNPUPod
}

// collectRescheduleTs collects all reschedule timestamps of the job in the current CM
func collectRescheduleTs(records []RescheduleRecord) map[int64]struct{} {
	res := make(map[int64]struct{}, len(records))
	for _, rr := range records {
		res[rr.RescheduleTimeStamp] = struct{}{}
	}
	return res
}
