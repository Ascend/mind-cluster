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

// Pakcage cluster is a DT collection for func in job_summary
package cluster

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/core/model/enum"
	"ascend-faultdiag-online/pkg/model"
	"ascend-faultdiag-online/pkg/model/slownode"
	"ascend-faultdiag-online/pkg/service/servicefunc/slownode/common"
	"ascend-faultdiag-online/pkg/service/servicefunc/slownode/slownodejob"
)

func init() {
	config := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	err := hwlog.InitRunLogger(&config, nil)
	if err != nil {
		fmt.Println(err)
	}
}

var (
	testJobName   = "testJobName"
	testNamespace = "testNamespace"
	testJobId     = "testJobId"
)

func ctxGenerator() *slownodejob.JobContext {
	ctx := &slownodejob.JobContext{
		Job: &slownode.Job{},
	}
	ctx.Job.JobName = testJobName
	ctx.Job.Namespace = testNamespace
	return ctx
}

func jobSummaryGenerator(t *testing.T) *model.JobSummary {
	var jobSummary = &model.JobSummary{
		JobName:   testJobName,
		Namespace: testNamespace,
		JobId:     testJobId,
	}
	var data = `{
		"server_list": [
			{
				"pod_id": "123",
				"server_id": "127.0.0.1",
				"server_sn": "321123",
				"device": [
					{
						"rank_id": "1"
					},
										{
						"rank_id": "2"
					}
				]
			}
		]}`
	err := json.Unmarshal([]byte(data), &jobSummary.HcclJson)
	assert.Nil(t, err)
	return jobSummary
}

func TestJobSummaryProcessor(t *testing.T) {
	slownodejob.GetJobCtxMap().Clear()
	var jobSummary = jobSummaryGenerator(t)
	var ctx = ctxGenerator()
	defer slownodejob.GetJobCtxMap().Clear()
	convey.Convey("test jobSummaryProcessor", t, func() {
		testJobSummaryProcessorCase1(ctx, jobSummary)
		testJobSummaryProcessorCase2(ctx, jobSummary)
	})
}

func testJobSummaryProcessorCase1(ctx *slownodejob.JobContext, jobSummary *model.JobSummary) {
	convey.Convey("test no ctx found", func() {
		output := captureOutput(func() {
			jobSummaryProcessor(jobSummary)
		})
		convey.So(output, convey.ShouldEqual, "")
	})
	convey.Convey("test found ctx with different jobId", func() {
		slownodejob.GetJobCtxMap().Insert(ctx.Job.KeyGenerator(), ctx)
		// job summary status is empty, do the default case
		jobSummaryProcessor(jobSummary)
		convey.So(ctx.Job.JobId, convey.ShouldEqual, testJobId)
		convey.So(ctx.TrainingJobStatus, convey.ShouldEqual, "")
	})
}

func testJobSummaryProcessorCase2(ctx *slownodejob.JobContext, jobSummary *model.JobSummary) {
	convey.Convey("test found ctx with job status is add", func() {
		patch := gomonkey.ApplyFunc(jobStatusProcessor, func(*slownodejob.JobContext, *model.JobSummary) {
			fmt.Println("mock the jobStatusProcessor")
		})
		defer patch.Reset()
		slownodejob.GetJobCtxMap().Clear()
		slownodejob.GetJobCtxMap().Insert(ctx.Job.KeyGenerator(), ctx)
		jobSummary.Operator = add
		output := captureOutput(func() {
			jobSummaryProcessor(jobSummary)
		})

		convey.So(ctx.Job.JobId, convey.ShouldEqual, testJobId)
		convey.So(ctx.TrainingJobStatus, convey.ShouldEqual, "")
		convey.So(output, convey.ShouldContainSubstring, "mock the jobStatusProcessor")
	})
	convey.Convey("test found ctx with job status is delete", func() {
		patch := gomonkey.ApplyPrivateMethod(
			reflect.TypeOf(&jobProcessor{}), "delete", func(*jobProcessor) { fmt.Println("mock delete") },
		)
		defer patch.Reset()
		slownodejob.GetJobCtxMap().Clear()
		slownodejob.GetJobCtxMap().Insert(ctx.Job.KeyGenerator(), ctx)
		jobSummary.Operator = del
		output := captureOutput(func() {
			jobSummaryProcessor(jobSummary)
		})
		convey.So(ctx.Job.JobId, convey.ShouldEqual, testJobId)
		convey.So(ctx.TrainingJobStatus, convey.ShouldEqual, "")
		convey.So(output, convey.ShouldContainSubstring, "mock delete")
	})
}

func TestServersGenerator(t *testing.T) {
	convey.Convey("test serversGenerator", t, func() {
		var hcclJson = model.HcclJson{}
		servers := serversGenerator(hcclJson)
		convey.So(servers, convey.ShouldBeEmpty)
		var data = `{
		"server_list": [
			{
				"server_id": "127.0.0.1",
				"server_sn": "321123",
				"device": [
					{
						"rank_id": "1"
					},
										{
						"rank_id": "2"
					}
				]
			}
		]}`
		err := json.Unmarshal([]byte(data), &hcclJson)
		convey.So(err, convey.ShouldBeNil)
		var expect = []slownode.Server{
			{
				Sn:      "321123",
				Ip:      "127.0.0.1",
				RankIds: []string{"1", "2"},
			},
		}
		servers = serversGenerator(hcclJson)
		convey.So(reflect.DeepEqual(servers, expect), convey.ShouldBeTrue)
	})
}

func TestJobStatusProcessor(t *testing.T) {
	slownodejob.GetJobCtxMap().Clear()
	// prepare the jobSummary
	var jobSummary = jobSummaryGenerator(t)
	var ctx = ctxGenerator()
	defer slownodejob.GetJobCtxMap().Clear()
	convey.Convey("Test jobStatusProcessor", t, func() {
		testJobStatusProcessorWithOtherJobStatus(ctx, jobSummary)
		testJobStatusProcessorWithJobStatusIsRunning(ctx, jobSummary)
	})
}

func testJobStatusProcessorWithJobStatusIsRunning(ctx *slownodejob.JobContext, jobSummary *model.JobSummary) {
	convey.Convey("test job status is runing", func() {
		jobSummary.JobStatus = enum.IsRunning
		convey.Convey("test ctx is not running", func() {
			// start job
			patch := gomonkey.ApplyPrivateMethod(
				reflect.TypeOf(&jobProcessor{}), "start", func(*jobProcessor) { fmt.Println("mock start") },
			)
			defer patch.Reset()
			output := captureOutput(func() {
				jobStatusProcessor(ctx, jobSummary)
			})
			convey.So(output, convey.ShouldContainSubstring, "mock start")
		})
		convey.Convey("test ctx is running and not reschedule", func() {
			// servers are equal -> empty output
			output := captureOutput(func() {
				jobStatusProcessor(ctx, jobSummary)
			})
			convey.So(output, convey.ShouldBeEmpty)
		})
		convey.Convey("test ctx is running and reschedule", func() {
			setUnexportedFiled(ctx, "isRunning", true)
			// servers are not equal -> stop and start
			patch := gomonkey.ApplyPrivateMethod(
				reflect.TypeOf(&jobProcessor{}), "stop", func(*jobProcessor) { fmt.Println("mock stop") },
			)
			patch.ApplyPrivateMethod(
				reflect.TypeOf(&jobProcessor{}), "start", func(*jobProcessor) { fmt.Println("mock start") },
			)
			patch.ApplyFunc(common.AreServersEqual, func(a, b []slownode.Server) bool { return false })
			defer patch.Reset()
			output := captureOutput(func() {
				jobStatusProcessor(ctx, jobSummary)
			})
			convey.So(output, convey.ShouldContainSubstring, "mock stop")
			convey.So(output, convey.ShouldContainSubstring, "mock start")
		})
	})
}

func testJobStatusProcessorWithOtherJobStatus(ctx *slownodejob.JobContext, jobSummary *model.JobSummary) {
	convey.Convey("test update server and jobStatus is empty", func() {
		jobStatusProcessor(ctx, jobSummary)
		servers := serversGenerator(jobSummary.HcclJson)
		convey.So(reflect.DeepEqual(ctx.Job.Servers, servers), convey.ShouldBeTrue)
	})
	convey.Convey("test job status is pending", func() {
		patch := gomonkey.ApplyPrivateMethod(
			reflect.TypeOf(&jobProcessor{}), "stop", func(*jobProcessor) { fmt.Println("mock stop") },
		)
		defer patch.Reset()
		jobSummary.JobStatus = enum.IsPending
		output := captureOutput(func() {
			jobStatusProcessor(ctx, jobSummary)
		})
		convey.So(output, convey.ShouldContainSubstring, "mock stop")
	})
	convey.Convey("test job status is failed", func() {
		patch := gomonkey.ApplyPrivateMethod(
			reflect.TypeOf(&jobProcessor{}), "stop", func(*jobProcessor) { fmt.Println("mock stop") },
		)
		defer patch.Reset()
		jobSummary.JobStatus = enum.IsFailed
		output := captureOutput(func() {
			jobStatusProcessor(ctx, jobSummary)
		})
		convey.So(output, convey.ShouldContainSubstring, "mock stop")
	})
	convey.Convey("test job status is complete", func() {
		patch := gomonkey.ApplyPrivateMethod(
			reflect.TypeOf(&jobProcessor{}), "delete", func(*jobProcessor) { fmt.Println("mock delete") },
		)
		defer patch.Reset()
		jobSummary.JobStatus = enum.IsCompleted
		output := captureOutput(func() {
			jobStatusProcessor(ctx, jobSummary)
		})
		convey.So(output, convey.ShouldContainSubstring, "mock delete")
	})
}
