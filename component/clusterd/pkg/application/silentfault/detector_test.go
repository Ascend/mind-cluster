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
	"errors"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	corev1 "k8s.io/api/core/v1"

	"ascend-common/api"
	"clusterd/pkg/application/publicfault"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/conf"
	"clusterd/pkg/domain/job"
	domainpublicfault "clusterd/pkg/domain/publicfault"
	"clusterd/pkg/domain/silentfault"
	"clusterd/pkg/interface/kube"
)

func TestDetectFirstFault(t *testing.T) {
	convey.Convey("detect first fault hit", t, func() {
		firstFaultMgr.Reset()
		enableSilentFault()
		conf.SetSilentFaultPolicy(func() conf.SilentFaultPolicy {
			p := conf.SilentFaultPolicy{Enabled: true}
			p.Detect.ConsecutiveTimes = 2
			p.Detect.WindowSecond = 3600
			p.Detect.HardwareFaultWindowSecond = 30
			return p
		}())

		now := time.Now().Unix()
		firstFaultMgr.AddEvent("n1", &silentfault.FirstFaultEvent{JobID: "j1", IsFirst: true, Timestamp: now - 10})
		firstFaultMgr.AddEvent("n1", &silentfault.FirstFaultEvent{JobID: "j2", IsFirst: true, Timestamp: now - 5})

		p0 := gomonkey.ApplyFunc(job.GetJobCache, func(string) (constant.JobInfo, bool) {
			return constant.JobInfo{ResourceType: "Ascend910", PreServerList: []constant.ServerHccl{
				{ServerName: "n1", DeviceList: []constant.Device{{}}},
			}}, true
		})
		defer p0.Reset()
		p1 := gomonkey.ApplyFunc(kube.GetNode, func(string) *corev1.Node { return newNodeWithCards("Ascend910", 1) })
		defer p1.Reset()

		called := false
		p2 := gomonkey.ApplyFunc(writeSilentFault, func(string, []string) { called = true })
		defer p2.Reset()

		detectFirstFault()
		convey.So(called, convey.ShouldBeTrue)
		convey.So(len(firstFaultMgr.GetEvents("n1")), convey.ShouldEqual, 0)
	})
}

func TestDetectFirstFaultNotHit(t *testing.T) {
	convey.Convey("detect first fault not hit keeps node", t, func() {
		firstFaultMgr.Reset()
		conf.SetSilentFaultPolicy(func() conf.SilentFaultPolicy {
			p := conf.SilentFaultPolicy{Enabled: true}
			p.Detect.ConsecutiveTimes = 3
			p.Detect.WindowSecond = 3600
			return p
		}())
		now := time.Now().Unix()
		firstFaultMgr.AddEvent("n1", &silentfault.FirstFaultEvent{JobID: "j1", IsFirst: true, Timestamp: now - 10})

		called := false
		p := gomonkey.ApplyFunc(writeSilentFault, func(string, []string) { called = true })
		defer p.Reset()

		detectFirstFault()
		convey.So(called, convey.ShouldBeFalse)
	})
}

func TestWriteSilentFault(t *testing.T) {
	convey.Convey("write silent fault success", t, func() {
		enableSilentFault()
		p0 := gomonkey.ApplyFunc(job.GetJobCache, func(string) (constant.JobInfo, bool) {
			return constant.JobInfo{ResourceType: "Ascend910"}, true
		})
		defer p0.Reset()
		p1 := gomonkey.ApplyFunc(kube.GetNode, func(string) *corev1.Node { return newNodeWithBaseDevInfos("Ascend910-0", "Ascend910-1") })
		defer p1.Reset()
		p2 := gomonkey.ApplyFunc(domainpublicfault.GetFaultLevelByCode, func(string) string {
			return constant.SilentFault
		})
		defer p2.Reset()
		var got *api.PubFaultInfo
		p3 := gomonkey.ApplyFunc(publicfault.PubFaultCollector, func(info *api.PubFaultInfo) error {
			got = info
			return nil
		})
		defer p3.Reset()

		writeSilentFault("n1", []string{"j1"})
		convey.So(got, convey.ShouldNotBeNil)
		convey.So(got.Faults[0].FaultCode, convey.ShouldEqual, constant.SilentFaultCode)
		convey.So(got.Faults[0].Assertion, convey.ShouldEqual, constant.AssertionOccur)
		convey.So(got.Faults[0].Influence[0].DeviceIds, convey.ShouldResemble, []int32{0, 1})
	})
}

func TestWriteSilentFaultNoLevel(t *testing.T) {
	convey.Convey("write silent fault skip when level not configured", t, func() {
		enableSilentFault()
		p1 := gomonkey.ApplyFunc(domainpublicfault.GetFaultLevelByCode, func(string) string { return "" })
		defer p1.Reset()
		called := false
		p2 := gomonkey.ApplyFunc(publicfault.PubFaultCollector, func(*api.PubFaultInfo) error {
			called = true
			return nil
		})
		defer p2.Reset()
		writeSilentFault("n1", []string{"j1"})
		convey.So(called, convey.ShouldBeFalse)
	})
}

func TestWriteSilentFaultNoDevices(t *testing.T) {
	convey.Convey("write silent fault skip when no devices", t, func() {
		enableSilentFault()
		p0 := gomonkey.ApplyFunc(job.GetJobCache, func(string) (constant.JobInfo, bool) {
			return constant.JobInfo{ResourceType: "Ascend910"}, true
		})
		defer p0.Reset()
		p1 := gomonkey.ApplyFunc(kube.GetNode, func(string) *corev1.Node { return nil })
		defer p1.Reset()
		p2 := gomonkey.ApplyFunc(domainpublicfault.GetFaultLevelByCode, func(string) string { return constant.SilentFault })
		defer p2.Reset()
		called := false
		p3 := gomonkey.ApplyFunc(publicfault.PubFaultCollector, func(*api.PubFaultInfo) error { called = true; return nil })
		defer p3.Reset()
		writeSilentFault("n1", []string{"j1"})
		convey.So(called, convey.ShouldBeFalse)
	})
}

func TestWriteSilentFaultCollectorError(t *testing.T) {
	convey.Convey("write silent fault logs error when collector fails", t, func() {
		enableSilentFault()
		p0 := gomonkey.ApplyFunc(job.GetJobCache, func(string) (constant.JobInfo, bool) {
			return constant.JobInfo{ResourceType: "Ascend910"}, true
		})
		defer p0.Reset()
		p1 := gomonkey.ApplyFunc(kube.GetNode, func(string) *corev1.Node { return newNodeWithBaseDevInfos("Ascend910-0") })
		defer p1.Reset()
		p2 := gomonkey.ApplyFunc(domainpublicfault.GetFaultLevelByCode, func(string) string { return constant.SilentFault })
		defer p2.Reset()
		called := false
		p3 := gomonkey.ApplyFunc(publicfault.PubFaultCollector, func(*api.PubFaultInfo) error {
			called = true
			return errors.New("boom")
		})
		defer p3.Reset()
		writeSilentFault("n1", []string{"j1"})
		convey.So(called, convey.ShouldBeTrue)
	})
}
