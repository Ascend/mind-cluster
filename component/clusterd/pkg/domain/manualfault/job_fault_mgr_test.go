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

// Package manualfault test of cache for hardware frequency fault with job
package manualfault

import (
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"volcano.sh/apis/pkg/apis/scheduling/v1beta1"

	"clusterd/pkg/domain/conf"
	"clusterd/pkg/domain/podgroup"
)

const slidingWindow = 3

func TestJobFaultMgrAdd(t *testing.T) {
	conf.SetManualSeparatePolicy(validPolicy)
	p1 := gomonkey.ApplyFuncReturn(podgroup.GetPodGroup, v1beta1.PodGroup{
		ObjectMeta: v1.ObjectMeta{
			Name: job1,
		},
	})
	defer p1.Reset()
	convey.Convey("test func AddFault, fault is nil", t, testNilFault)
	convey.Convey("test func AddFault, job is empty", t, testEmptyJob)
	convey.Convey("test func AddFault, add same fault dev, not software fault", t, testAddSameFaultDev)
	convey.Convey("test func AddFault, add diff fault dev, is software fault", t, testAddDiffFaultDev)
}

func testNilFault() {
	JobFaultMgr.AddFault(nil)
	convey.So(len(JobFaultMgr.jobFault), convey.ShouldEqual, len0)
}

func testEmptyJob() {
	InitJobFaultManager(slidingWindow)
	InitCounter()
	fault1 := getDemoFault1()
	fault1.JobId = ""
	convey.So(len(Counter.faults), convey.ShouldEqual, len0)
	JobFaultMgr.AddFault(fault1)
	convey.So(len(Counter.faults), convey.ShouldEqual, len0)
	convey.So(len(JobFaultMgr.jobFault), convey.ShouldEqual, len0)
}

func testAddSameFaultDev() {
	InitJobFaultManager(slidingWindow)
	InitCounter()
	// node1, dev1, code1, time1
	fault1 := getDemoFault1()
	fault1.ReceiveTime = time.Now().UnixMilli()
	// node1, dev1, code1, time2
	fault2 := getDemoFault2()
	fault2.ReceiveTime = fault1.ReceiveTime
	// node1, dev1, code2, time4
	fault4 := getDemoFault4()
	fault4.ReceiveTime = fault1.ReceiveTime
	JobFaultMgr.AddFault(fault1)
	JobFaultMgr.AddFault(fault2)
	JobFaultMgr.AddFault(fault4)
	convey.So(len(JobFaultMgr.jobFault), convey.ShouldEqual, len1)
	faults, ok := JobFaultMgr.jobFault[fault1.JobId]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(len(faults.faults), convey.ShouldEqual, len3) // 3 faults
	time.Sleep((slidingWindow + 1) * time.Second)           // deal sfw fault
	convey.So(len(JobFaultMgr.jobFault), convey.ShouldEqual, len1)
	faults, ok = JobFaultMgr.jobFault[fault1.JobId]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(len(faults.faults), convey.ShouldEqual, len0)
	// count
	info, ok := Counter.faults[fault1.NodeName]
	convey.So(ok, convey.ShouldBeTrue)
	devInfo, ok := info[fault1.DevName]
	convey.So(ok, convey.ShouldBeTrue)
	times, ok := devInfo.fault[fault1.Code]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(len(times), convey.ShouldEqual, len2)

	times, ok = devInfo.fault[fault4.Code]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(len(times), convey.ShouldEqual, len1)

	time.Sleep((slidingWindow + 1) * time.Second)
	_, ok = JobFaultMgr.jobFault[fault1.JobId]
	convey.So(ok, convey.ShouldBeFalse)
}

func testAddDiffFaultDev() {
	const interval = 2
	InitJobFaultManager(slidingWindow)
	InitCounter()
	// node1, dev1, code1, time1
	fault1 := getDemoFault1()
	fault1.ReceiveTime = time.Now().UnixMilli()
	// node1, dev1, code1, time2
	fault2 := getDemoFault2()
	fault2.ReceiveTime = fault1.ReceiveTime
	// node1, dev2, code1, time5
	fault5 := getDemoFault5()
	fault5.ReceiveTime = fault1.ReceiveTime
	JobFaultMgr.AddFault(fault1)
	JobFaultMgr.AddFault(fault2)
	JobFaultMgr.AddFault(fault5)
	convey.So(len(JobFaultMgr.jobFault), convey.ShouldEqual, len1)
	time.Sleep((slidingWindow + interval) * time.Second)
	_, ok := JobFaultMgr.jobFault[fault1.JobId]
	convey.So(ok, convey.ShouldBeFalse)
	// don't count
	convey.So(len(Counter.faults), convey.ShouldEqual, len0)
}

func TestJobFaultMgrGet(t *testing.T) {
	conf.SetManualSeparatePolicy(validPolicy)
	p1 := gomonkey.ApplyFuncReturn(podgroup.GetPodGroup, v1beta1.PodGroup{
		ObjectMeta: v1.ObjectMeta{
			Name: job1,
		},
	})
	defer p1.Reset()
	convey.Convey("test func GetFaultsByJobId", t, func() {
		fault1 := getDemoFault1()
		fault2 := getDemoFault2()
		JobFaultMgr.AddFault(fault1)
		JobFaultMgr.AddFault(fault2)
		jobFaults := JobFaultMgr.GetFaultsByJobId(fault1.JobId)
		convey.So(len(jobFaults), convey.ShouldEqual, len2)
	})
}
