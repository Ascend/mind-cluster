/* Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, convey.Software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

// Package cluster a DT collection for slownode cluster feature func
package cluster

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/apimachinery/pkg/watch"

	"ascend-faultdiag-online/pkg/core/config"
	"ascend-faultdiag-online/pkg/core/context"
	"ascend-faultdiag-online/pkg/model/slownode"
	"ascend-faultdiag-online/pkg/service/servicefunc/slownode/slownodejob"
	"ascend-faultdiag-online/pkg/utils/constants"
)

func TestWaitNodeReport(t *testing.T) {
	patches := gomonkey.NewPatches()
	defer patches.Reset()
	timeout := 1

	// Mock context
	context.FdCtx = &context.FaultDiagContext{}
	context.FdCtx.Config = &config.FaultDiagConfig{
		Cluster: config.Cluster{
			NodeReportTimeout: timeout,
		},
	}
	convey.Convey("Test waitNodeReport", t, func() {
		ctx := ctxGenerator()
		ctx.StopChan = make(chan struct{})
		ctx.NodeReportSignal = make(chan struct{})
		j := &jobProcessor{
			ctx: ctx,
			job: &slownode.Job{},
		}
		convey.Convey("When node reports before timeout", func() {
			j.waitNodeReport()
			ctx.NodeReportSignal <- struct{}{}
		})
		convey.Convey("When timeout occurs", func() {
			mockStop := gomonkey.ApplyPrivateMethod(
				reflect.TypeOf(&jobProcessor{}),
				"stop",
				func(*jobProcessor) {
					fmt.Println("stop called")
				},
			)
			defer mockStop.Reset()
			ctx.StopChan = make(chan struct{})
			ctx.NodeReportSignal = make(chan struct{})
			j.waitNodeReport()
			time.Sleep(time.Second)
		})
		convey.Convey("When job is stopped", func() {
			ctx.StopChan = make(chan struct{})
			ctx.NodeReportSignal = make(chan struct{})
			j.waitNodeReport()
			close(j.ctx.StopChan)
			time.Sleep(time.Millisecond)
		})
	})
}

func TestJobProcessor(t *testing.T) {
	patches := gomonkey.NewPatches()
	defer patches.Reset()
	job := &slownode.Job{}
	convey.Convey("Test JobProcessor", t, func() {
		testJobProcessorWithCase1(job)
		testJobProcessorWithCase2(job)
	})
}

func testJobProcessorWithCase1(job *slownode.Job) {
	convey.Convey("When operator is Added", func() {
		job.JobName = testJobName
		mockAdd := gomonkey.ApplyPrivateMethod(&jobProcessor{}, "add", func(*jobProcessor) {
			fmt.Println("mock start")
		})
		defer mockAdd.Reset()
		output := captureOutput(func() {
			JobProcessor(nil, job, watch.Added)
		})
		convey.So(output, convey.ShouldContainSubstring, "mock start")
	})

	convey.Convey("When operator is Modified", func() {
		mockUpdate := gomonkey.ApplyPrivateMethod(&jobProcessor{}, "update", func(*jobProcessor) {
			fmt.Println("mock update")
		})
		defer mockUpdate.Reset()
		output := captureOutput(func() {
			JobProcessor(nil, job, watch.Modified)
		})
		convey.So(output, convey.ShouldContainSubstring, "mock update")
	})

	convey.Convey("When operator is Deleted", func() {
		mockDelete := gomonkey.ApplyPrivateMethod(&jobProcessor{}, "delete", func(*jobProcessor) {
			fmt.Println("mock delete")
		})
		defer mockDelete.Reset()
		output := captureOutput(func() {
			JobProcessor(nil, job, watch.Deleted)
		})
		convey.So(output, convey.ShouldContainSubstring, "mock delete")
	})
}

func testJobProcessorWithCase2(job *slownode.Job) {
	convey.Convey("When job name is empty", func() {
		job.JobName = ""
		output := captureOutput(func() {
			JobProcessor(nil, job, watch.Added)
		})
		convey.So(output, convey.ShouldEqual, "")
	})
	convey.Convey("When operator is unknown", func() {
		job.JobName = testJobName
		output := captureOutput(func() {
			JobProcessor(nil, job, watch.Bookmark)
		})
		convey.So(output, convey.ShouldEqual, "")
	})
}

func TestJobRestartProcessor(t *testing.T) {
	var ip = "127.0.0.1"
	patches := gomonkey.ApplyPrivateMethod(reflect.TypeOf(&jobProcessor{}), "stop", func(*jobProcessor) {
		fmt.Println("mock stop success")
	})
	patches.ApplyPrivateMethod(reflect.TypeOf(&jobProcessor{}), "start", func(*jobProcessor) {
		fmt.Println("mock start success")
	})
	defer patches.Reset()
	convey.Convey("test JobRestartProcessor", t, func() {
		convey.Convey("test no job found", func() {
			output := captureOutput(func() {
				JobRestartProcessor(&ip, &ip, watch.Modified)
			})
			convey.So(output, convey.ShouldEqual, "")
		})
		convey.Convey("test delete eventType", func() {
			output := captureOutput(func() {
				JobRestartProcessor(&ip, &ip, watch.Deleted)
			})
			convey.So(output, convey.ShouldEqual, "")
		})
		convey.Convey("test found job", func() {
			ctx := &slownodejob.JobContext{Job: &slownode.Job{}}
			ctx.Job.Servers = []slownode.Server{{Ip: ip}}
			slownodejob.GetJobCtxMap().Clear()
			slownodejob.GetJobCtxMap().Insert("testKey", ctx)
			JobRestartProcessor(&ip, &ip, watch.Modified)
			time.Sleep(constants.RestartInterval * time.Millisecond)
		})
	})
}
