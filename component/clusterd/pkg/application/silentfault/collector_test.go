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
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	v1 "k8s.io/api/core/v1"

	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/conf"
	"clusterd/pkg/domain/job"
	"clusterd/pkg/domain/silentfault"
	"clusterd/pkg/interface/kube"
)

func TestJobCardsAtLeast(t *testing.T) {
	convey.Convey("job cards at least threshold", t, func() {
		ji := constant.JobInfo{PreServerList: []constant.ServerHccl{
			{ServerName: "n1", DeviceList: []constant.Device{{}, {}, {}}},
		}}
		convey.So(jobCardsAtLeast(ji, 2), convey.ShouldBeTrue)
		convey.So(jobCardsAtLeast(ji, 3), convey.ShouldBeTrue)
		convey.So(jobCardsAtLeast(ji, 4), convey.ShouldBeFalse)
	})
}

func TestIsWholeNodeJob(t *testing.T) {
	convey.Convey("is whole node job", t, func() {
		p := gomonkey.ApplyFunc(kube.GetNode, func(string) *v1.Node { return newNodeWithCards("Ascend910", 2) })
		defer p.Reset()

		ji := constant.JobInfo{ResourceType: "Ascend910", PreServerList: []constant.ServerHccl{
			{ServerName: "n1", PodName: "p1", DeviceList: []constant.Device{{}, {}}},
		}}
		convey.So(isWholeNodeJob(ji, "p1"), convey.ShouldBeTrue)

		ji2 := constant.JobInfo{ResourceType: "Ascend910", PreServerList: []constant.ServerHccl{
			{ServerName: "n1", PodName: "p1", DeviceList: []constant.Device{{}}},
		}}
		convey.So(isWholeNodeJob(ji2, "p1"), convey.ShouldBeFalse)

		// empty failed pod name is treated as invalid
		convey.So(isWholeNodeJob(ji, ""), convey.ShouldBeFalse)

		// failed pod is a non-NPU pod co-located with a whole-node NPU pod on the same node
		ji3 := constant.JobInfo{ResourceType: "Ascend910", PreServerList: []constant.ServerHccl{
			{ServerName: "n1", PodName: "p1", DeviceList: []constant.Device{{}, {}}},
		}}
		convey.So(isWholeNodeJob(ji3, "coordinator"), convey.ShouldBeFalse)

		// every NPU pod of the job must be whole-node, not only the failed one
		ji4 := constant.JobInfo{ResourceType: "Ascend910", PreServerList: []constant.ServerHccl{
			{ServerName: "n1", PodName: "p1", DeviceList: []constant.Device{{}, {}}},
			{ServerName: "n2", PodName: "p2", DeviceList: []constant.Device{{}}},
		}}
		convey.So(isWholeNodeJob(ji4, "p1"), convey.ShouldBeFalse)
	})
}

func TestCollectRescheduleTs(t *testing.T) {
	convey.Convey("collect reschedule ts", t, func() {
		records := []RescheduleRecord{{RescheduleTimeStamp: 10}, {RescheduleTimeStamp: 20}}
		ts := collectRescheduleTs(records)
		convey.So(len(ts), convey.ShouldEqual, 2)
		_, ok := ts[10]
		convey.So(ok, convey.ShouldBeTrue)
	})
}

func TestOnReschedule(t *testing.T) {
	convey.Convey("on reschedule", t, func() {
		pendingCache.ResetEvents()
		pendingCache.ResetProcessed()
		// batch reschedule skip
		OnReschedule("j1", RescheduleRecord{ReasonOfTask: []RescheduleTask{{}, {}}})
		convey.So(len(pendingCache.TakeDue(0)), convey.ShouldEqual, 0)
		// non pod-failed skip
		OnReschedule("j1", RescheduleRecord{ReasonOfTask: []RescheduleTask{{RescheduleReason: "other"}}})
		convey.So(len(pendingCache.TakeDue(0)), convey.ShouldEqual, 0)

		p := gomonkey.ApplyFunc(buildPendingEvent, func(string, string, string, int64) *silentfault.PendingEvent {
			return &silentfault.PendingEvent{JobID: "j1", FailNode: "n1", Timestamp: 5}
		})
		defer p.Reset()
		OnReschedule("j1", RescheduleRecord{
			RescheduleTimeStamp: 5,
			ReasonOfTask:        []RescheduleTask{{RescheduleReason: constant.PodFailedReason, NodeName: "n1"}},
		})
		convey.So(len(pendingCache.TakeDue(100)), convey.ShouldEqual, 1)
	})
}

func TestBuildPendingEvent(t *testing.T) {
	policy := conf.SilentFaultPolicy{}
	policy.Detect.MinTaskCards = 0
	conf.SetSilentFaultPolicy(policy)

	convey.Convey("build pending event job not found", t, func() {
		p := gomonkey.ApplyFunc(job.GetJobCache, func(string) (constant.JobInfo, bool) {
			return constant.JobInfo{}, false
		})
		defer p.Reset()
		convey.So(buildPendingEvent("j1", "", "n1", 1), convey.ShouldBeNil)
	})

	convey.Convey("build pending event not whole node", t, func() {
		p := gomonkey.ApplyFunc(job.GetJobCache, func(string) (constant.JobInfo, bool) {
			ji := constant.JobInfo{ResourceType: "Ascend910", PreServerList: []constant.ServerHccl{
				{ServerName: "n1", PodName: "p1", DeviceList: []constant.Device{{}}},
			}}
			return ji, true
		})
		defer p.Reset()
		p2 := gomonkey.ApplyFunc(kube.GetNode, func(string) *v1.Node { return newNodeWithCards("Ascend910", 2) })
		defer p2.Reset()
		convey.So(buildPendingEvent("j1", "p1", "n1", 1), convey.ShouldBeNil)
	})

	convey.Convey("build pending event success", t, func() {
		p := gomonkey.ApplyFunc(job.GetJobCache, func(string) (constant.JobInfo, bool) {
			ji := constant.JobInfo{ResourceType: "Ascend910", PreServerList: []constant.ServerHccl{
				{ServerName: "n1", PodName: "p1", DeviceList: []constant.Device{{}}},
			}}
			return ji, true
		})
		defer p.Reset()
		p2 := gomonkey.ApplyFunc(kube.GetNode, func(string) *v1.Node { return newNodeWithCards("Ascend910", 1) })
		defer p2.Reset()
		ev := buildPendingEvent("j1", "p1", "n1", 1)
		convey.So(ev, convey.ShouldNotBeNil)
		convey.So(ev.FailNode, convey.ShouldEqual, "n1")
		convey.So(len(ev.TaskNodes), convey.ShouldEqual, 1)
	})
}

func TestCmHandler(t *testing.T) {
	convey.Convey("cm handler disabled returns early", t, func() {
		conf.SetSilentFaultPolicy(conf.SilentFaultPolicy{Enabled: false})
		pendingCache.ResetEvents()
		pendingCache.ResetProcessed()
		CmHandler(nil, nil, constant.AddOperator)
		convey.So(len(pendingCache.TakeDue(0)), convey.ShouldEqual, 0)
	})

	convey.Convey("cm handler deleted clears processed, keeps pending events", t, func() {
		enableSilentFault()
		pendingCache.Add(&silentfault.PendingEvent{JobID: "j1", Timestamp: 1})
		pendingCache.MarkProcessed("j1", 1)
		CmHandler(nil, nil, constant.DeleteOperator)
		convey.So(len(pendingCache.TakeDue(100)), convey.ShouldEqual, 1)
		convey.So(pendingCache.HasProcessed("j1", 1), convey.ShouldBeFalse)
	})

	convey.Convey("cm handler process reschedule records", t, func() {
		enableSilentFault()
		pendingCache.ResetEvents()
		pendingCache.ResetProcessed()
		data := `{"j1":{"jobUID":"j1","rescheduleRecords":[{"rescheduleTimeStamp":5,"reasonOfTask":[{"rescheduleReason":"pod-failed","nodeName":"n1"}]}]}}`
		cm := &v1.ConfigMap{Data: map[string]string{constant.RescheduleReasonCmKey: data}}
		p := gomonkey.ApplyFunc(buildPendingEvent, func(string, string, string, int64) *silentfault.PendingEvent {
			return &silentfault.PendingEvent{JobID: "j1", FailNode: "n1", Timestamp: 5}
		})
		defer p.Reset()
		CmHandler(nil, cm, constant.AddOperator)
		convey.So(len(pendingCache.TakeDue(100)), convey.ShouldEqual, 1)
	})
}

func TestCmHandlerEdgeCases(t *testing.T) {
	convey.Convey("cm handler invalid op returns", t, func() {
		enableSilentFault()
		pendingCache.ResetEvents()
		pendingCache.ResetProcessed()
		CmHandler(nil, nil, "unknown")
		convey.So(len(pendingCache.TakeDue(0)), convey.ShouldEqual, 0)
	})
	convey.Convey("cm handler nil cm returns", t, func() {
		enableSilentFault()
		pendingCache.ResetEvents()
		pendingCache.ResetProcessed()
		CmHandler(nil, nil, constant.AddOperator)
		convey.So(len(pendingCache.TakeDue(0)), convey.ShouldEqual, 0)
	})
	convey.Convey("cm handler empty data returns", t, func() {
		enableSilentFault()
		pendingCache.ResetEvents()
		pendingCache.ResetProcessed()
		CmHandler(nil, &v1.ConfigMap{Data: map[string]string{constant.RescheduleReasonCmKey: ""}}, constant.AddOperator)
		convey.So(len(pendingCache.TakeDue(0)), convey.ShouldEqual, 0)
	})
	convey.Convey("cm handler unmarshal error returns", t, func() {
		enableSilentFault()
		pendingCache.ResetEvents()
		pendingCache.ResetProcessed()
		CmHandler(nil, &v1.ConfigMap{Data: map[string]string{constant.RescheduleReasonCmKey: "not-json"}}, constant.AddOperator)
		convey.So(len(pendingCache.TakeDue(0)), convey.ShouldEqual, 0)
	})
}
