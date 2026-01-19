/* Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package node is a DT collection for func in job_processor
package node

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-faultdiag-online/pkg/model/slownode"
	"ascend-faultdiag-online/pkg/service/servicefunc/slownode/slownodejob"
)

func TestAdd(t *testing.T) {
	var job = &slownode.Job{}
	job.JobId = "testJobId"
	job.JobName = "testJobName"
	job.Namespace = "testNamespace"
	slownodejob.GetJobCtxMap().Clear()
	defer slownodejob.GetJobCtxMap().Clear()
	var jp = jobProcessor{
		ctx:          &slownodejob.JobContext{Job: job},
		job:          job,
		nodeIp:       "127.0.0.1",
		available:    true,
		availableIps: []string{"127.0.0.1", "127.0.0.2"},
	}
	convey.Convey("test add job", t, func() {
		convey.Convey("test not available", func() {
			jp.available = false
			jp.nodeIp = "127.0.0.3"
			jp.add()
		})
		// reset jp
		jp.available = true
		jp.nodeIp = "127.0.0.1"
		convey.Convey("test get jobId", func() {
			slownodejob.GetJobCtxMap().Insert(jp.job.JobId, jp.ctx)
			jp.add()
		})
		convey.Convey("test success case", func() {
			patch := gomonkey.ApplyPrivateMethod(reflect.TypeOf(&jp), "start", func(*jobProcessor) {
				fmt.Println("mock job processor start")
			})
			slownodejob.GetJobCtxMap().Delete(jp.job.JobId)
			output := captureOutput(func() {
				jp.add()
			})
			convey.So(output, convey.ShouldContainSubstring, "mock job processor start")
			patch.Reset()
		})
	})
}

func TestUpdate(t *testing.T) {
	var job = &slownode.Job{}
	job.JobId = "testJobId"
	job.JobName = "testJobName"
	job.Namespace = "testNamespace"
	defer slownodejob.GetJobCtxMap().Clear()
	var jp = &jobProcessor{}
	var reset = func(jp *jobProcessor) {
		slownodejob.GetJobCtxMap().Clear()
		jp.ctx = &slownodejob.JobContext{Job: job}
		jp.job = job
		jp.nodeIp = "127.0.0.1"
		jp.available = true
		jp.availableIps = []string{"127.0.0.1", "127.0.0.2"}
	}
	reset(jp)
	convey.Convey("test update job", t, func() {
		testUpdateWithFoundCtx(jp, job)
		testUpdateWithNoCtx(jp, reset)
	})
}

func testUpdateWithFoundCtx(jp *jobProcessor, job *slownode.Job) {
	convey.Convey("test get ctx with error", func() {
		slownodejob.GetJobCtxMap().Insert(job.JobId, jp.ctx)
		// not available -> call delete
		jp.available = false
		jp.nodeIp = "127.0.0.3"
		patch := gomonkey.ApplyPrivateMethod(reflect.TypeOf(jp), "delete", func(*jobProcessor) {
			fmt.Println("mock job processor delete")
		})
		defer patch.Reset()
		output := captureOutput(func() {
			jp.update()
		})
		convey.So(output, convey.ShouldContainSubstring, "mock job processor delete")
		// available, rankIds changed, stop and start
		jp.available = true
		jp.nodeIp = "127.0.0.1"
		patch.ApplyPrivateMethod(reflect.TypeOf(jp), "stop", func(*jobProcessor) {
			fmt.Println("mock job processor stop")
		})
		patch.ApplyPrivateMethod(reflect.TypeOf(jp), "start", func(*jobProcessor) {
			fmt.Println("mock job processor start")
		})
		newJob := &slownode.Job{}
		newJob.JobId = "testJobId"
		newJob.Servers = []slownode.Server{
			{
				Sn:      "1",
				Ip:      "127.0.1.1",
				RankIds: []string{"1", "2"},
			},
			{
				Sn:      "2",
				Ip:      "127.0.1.2",
				RankIds: []string{"1", "2", "3", "4"},
			},
		}
		jp.job = newJob
		jp.nodeIp = "127.0.1.1"
		jp.available = true
		jp.availableIps = []string{"127.0.1.1", "127.0.1.2"}
		output = captureOutput(func() {
			jp.update()
		})
		convey.So(reflect.DeepEqual(jp.ctx.Job.Servers, jp.job.Servers), convey.ShouldBeTrue)
		convey.So(output, convey.ShouldContainSubstring, "mock job processor stop")
		convey.So(output, convey.ShouldContainSubstring, "mock job processor start")
	})
}

func testUpdateWithNoCtx(jp *jobProcessor, reset func(jp *jobProcessor)) {
	convey.Convey("test no ctx in ctxMap", func() {
		reset(jp)
		// call add
		patch := gomonkey.ApplyPrivateMethod(reflect.TypeOf(jp), "add", func(*jobProcessor) {
			fmt.Println("mock job processor add")
		})
		defer patch.Reset()
		output := captureOutput(func() {
			jp.update()
		})
		convey.So(output, convey.ShouldContainSubstring, "mock job processor add")
	})
}

func TestDelete(t *testing.T) {
	slownodejob.GetJobCtxMap().Clear()
	defer slownodejob.GetJobCtxMap().Clear()
	jp := &jobProcessor{}
	ctx := &slownodejob.JobContext{}
	jp.job = &slownode.Job{}
	jp.job.JobId = "testJobId"
	convey.Convey("test delete job", t, func() {
		// no ctx found
		jp.delete()
		// found ctx
		slownodejob.GetJobCtxMap().Insert(jp.job.JobId, ctx)
		patches := gomonkey.ApplyPrivateMethod(reflect.TypeOf(jp), "stop", func(*jobProcessor) {
			fmt.Println("mock job processor stop")
		})
		patches.ApplyMethod(reflect.TypeOf(ctx), "RemoveAllCM", func(*slownodejob.JobContext) {
			fmt.Println("mock job context RemoveAllCM")
		})
		defer patches.Reset()
		output := captureOutput(func() {
			jp.delete()
		})
		convey.So(output, convey.ShouldContainSubstring, "mock job processor stop")
		convey.So(output, convey.ShouldContainSubstring, "mock job context RemoveAllCM")
		_, ok := slownodejob.GetJobCtxMap().Get(jp.job.JobId)
		convey.So(ok, convey.ShouldBeFalse)
	})
}
