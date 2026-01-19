/*
Copyright(C)2025. Huawei Technologies Co.,Ltd. All rights reserved.

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

// Package cluster is a DT collection for func in slownode_cluster
package cluster

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"testing"
	"unsafe"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-faultdiag-online/pkg/model/slownode"
	"ascend-faultdiag-online/pkg/service/servicefunc/slownode/algo"
	"ascend-faultdiag-online/pkg/service/servicefunc/slownode/slownodejob"
	"ascend-faultdiag-online/pkg/utils/grpc"
	"ascend-faultdiag-online/pkg/utils/grpc/pubfault"
)

func setUnexportedFiled(obj any, filedName string, value any) {
	v := reflect.ValueOf(obj).Elem()
	f := v.FieldByName(filedName)
	if !f.CanSet() {
		f = reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem()
	}
	f.Set(reflect.ValueOf(value))
}

func captureOutput(f func()) string {
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		return ""
	}
	os.Stdout = w
	os.Stderr = w

	f()

	err = w.Close()
	if err != nil {
		return ""
	}
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	if err != nil {
		return ""
	}
	return buf.String()
}

func TestAlgoCallbackProcessor(t *testing.T) {
	message := `{"slownode_default-test-pytorch-2pod-16npu":{"127.0.0.1":{"isSlow":0,"degradationLevel":"0.0%", ` +
		`"jobName":"default-test-pytorch-2pod-16npu","nodeRank":"127.0.0.1","slowCalculateRanks":null,` +
		`"slowCommunicationDomains":null,"slowSendRanks":null,"slowHostNodes":null,"slowIORanks":null,` +
		`"jobId":"test_jobId"}}}`
	ctx := &slownodejob.JobContext{Job: &slownode.Job{}}
	ctx.Job.JobId = "test_jobId"
	convey.Convey("test AlgoCallbackProcessor", t, func() {
		convey.Convey("unmarshal failed", func() {
			AlgoCallbackProcessor("}")
		})
		convey.Convey("ConvertMaptoStruct failed", func() {
			AlgoCallbackProcessor("{}")
		})
		convey.Convey("no ctx or ctx is not running", func() {
			AlgoCallbackProcessor(message)
			slownodejob.GetJobCtxMap().Insert("default/default-test-pytorch-2pod-16npu", ctx)
			AlgoCallbackProcessor(message)
		})
		convey.Convey("normal", func() {
			mock := gomonkey.ApplyPrivateMethod(
				reflect.TypeOf(&degradationProcessor{}),
				"handle",
				func(degradationProcessor, *slownodejob.JobContext, *slownode.ClusterAlgoResult) {
					fmt.Println("call degradationProcessor handle")
				},
			)
			defer mock.Reset()
			slownodejob.GetJobCtxMap().Insert("default/default-test-pytorch-2pod-16npu", ctx)
			setUnexportedFiled(ctx, "isRunning", true)
			output := captureOutput(func() {
				AlgoCallbackProcessor(message)
			})
			convey.So(output, convey.ShouldContainSubstring, "call degradationProcessor handle")
		})
	})
}

func TestParallelGroupInfoCallbackProcessor(t *testing.T) {
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyMethod(reflect.TypeOf(&algo.Controller{}), "Start", func(*algo.Controller) {
		fmt.Println("algo Start called")
	})

	convey.Convey("test ParallelGroupInfoCallbackProcessor", t, func() {
		testUnmarshalData()
		testFailedJob()
		testNormal()
	})
}

func testUnmarshalData() {
	message := `{"jobId":"not_existed_job_id","jobName":"test-job-name"`
	ParallelGroupInfoCallbackProcessor(message)
}

func testFailedJob() {
	message := `{"jobId":"not_existed_job_id","jobName":"test-job-name"}`
	ParallelGroupInfoCallbackProcessor(message)
	// has data but not running
	ctx := &slownodejob.JobContext{
		Job: &slownode.Job{},
	}
	ctx.Job.JobId = "existed_group_info_job_id"
	slownodejob.GetJobCtxMap().Insert("existed_group_info_job_id", ctx)
	defer slownodejob.GetJobCtxMap().Delete("existed_group_info_job_id")
	message = `{"jobId":"existed_group_info_job_id","jobName":"test-job-name"}`
	ParallelGroupInfoCallbackProcessor(message)
}

func testNormal() {
	message := `{"jobId":"existed_group_info_job_id","jobName":"test-job-name","isFinished":false}`
	ctx := &slownodejob.JobContext{
		Job: &slownode.Job{},
	}
	ctx.Job.JobId = "existed_group_info_job_id"
	setUnexportedFiled(ctx, "isRunning", true)
	slownodejob.GetJobCtxMap().Insert("existed_group_info_job_id", ctx)
	// isFinished is false, ignore it
	ParallelGroupInfoCallbackProcessor(message)
	output := captureOutput(func() {
		ParallelGroupInfoCallbackProcessor(message)
	})
	convey.So(output, convey.ShouldEqual, "")
	message = `{"jobId":"existed_group_info_job_id","jobName":"test-job-name","isFinished":true}`
	// isFinished is true, step unmatches, ignore it
	setUnexportedFiled(ctx, "step", slownodejob.ClusterStep1)
	ParallelGroupInfoCallbackProcessor(message)
	output = captureOutput(func() {
		ParallelGroupInfoCallbackProcessor(message)
	})
	convey.So(output, convey.ShouldEqual, "")
	// isFinished is true, step matches, call StartSlowNodeAlgo
	setUnexportedFiled(ctx, "step", slownodejob.ClusterStep2)
	output = captureOutput(func() {
		ParallelGroupInfoCallbackProcessor(message)
	})
	convey.So(output, convey.ShouldContainSubstring, "algo Start called")

}

func mockFunc(ctx *slownodejob.JobContext) *gomonkey.Patches {
	patches := gomonkey.NewPatches()
	patches.ApplyMethod(reflect.TypeOf(ctx), "StartHeavyProfiling", func(_ *slownodejob.JobContext) {
		setUnexportedFiled(ctx, "isStartedHeavyProfiling", true)
	})
	patches.ApplyMethod(reflect.TypeOf(ctx), "StopHeavyProfiling", func(_ *slownodejob.JobContext) {
		setUnexportedFiled(ctx, "isStartedHeavyProfiling", false)
		ctx.AlgoRes = make([]*slownode.ClusterAlgoResult, 0)
	})
	return patches
}

// close the inline opt by: go env -w GOFLAGS="-gcflags=-l"
func TestDegradationProcessor1(t *testing.T) {
	ctx := ctxGenerator()
	patches := mockFunc(ctx)
	defer patches.Reset()
	patches.ApplyPrivateMethod(
		reflect.TypeOf(&degradationProcessor{}),
		"reportSlowNode",
		func(degradationProcessor, *slownodejob.JobContext, *slownode.ClusterAlgoResult) {})
	convey.Convey("TestDegradationProcessor1", t, func() {
		testNoSlow(ctx)
		testIsSlowCase(ctx)
		testRecovery(ctx)
		testStartHeavyProfiling(ctx)
	})
}

func testNoSlow(ctx *slownodejob.JobContext) {
	result := &slownode.ClusterAlgoResult{}
	result.IsSlow = 0
	processor := newDegradationProcessor(ctx, result)
	processor.handle()
	convey.So(ctx.IsDegradation, convey.ShouldBeFalse)
	ctx.IsDegradation = true
	processor.handle()
	convey.So(ctx.IsDegradation, convey.ShouldBeFalse)
}

func testIsSlowCase(ctx *slownodejob.JobContext) {
	var num2 = 2
	var callReportSlowNodeCount int
	patch := gomonkey.ApplyPrivateMethod(
		reflect.TypeOf(&degradationProcessor{}),
		"reportSlowNode",
		func(degradationProcessor, *slownodejob.JobContext, *slownode.ClusterAlgoResult) {
			callReportSlowNodeCount++
		},
	)
	defer patch.Reset()
	result := &slownode.ClusterAlgoResult{}
	result.IsSlow = 1
	processor := newDegradationProcessor(ctx, result)
	processor.handle()
	// start heavy profiling
	convey.So(len(ctx.AlgoRes), convey.ShouldEqual, 1)
	convey.So(ctx.IsStartedHeavyProfiling(), convey.ShouldBeTrue)
	// one more time, no need to start heavy profiling again
	processor.handle()
	convey.So(len(ctx.AlgoRes), convey.ShouldEqual, num2)
	// 3 times degradation, stop heavy profiling
	for i := 0; i < 3; i++ {
		processor.handle()
	}
	convey.So(len(ctx.AlgoRes), convey.ShouldEqual, 0)
	convey.So(ctx.IsStartedHeavyProfiling(), convey.ShouldBeFalse)
	convey.So(ctx.IsDegradation, convey.ShouldBeTrue)
	convey.So(callReportSlowNodeCount, convey.ShouldEqual, 0)
	// rank ids exists, call reportSlowNode
	result.SlowCalculateRanks = []int{0}
	processor.handle()
	convey.So(ctx.IsDegradation, convey.ShouldBeTrue)
	convey.So(callReportSlowNodeCount, convey.ShouldEqual, 1)
}

func testRecovery(ctx *slownodejob.JobContext) {
	var count = 4
	result := &slownode.ClusterAlgoResult{}
	result.IsSlow = 0
	processor := newDegradationProcessor(ctx, result)
	// recovery, stop heavy profiling
	processor.handle()
	convey.So(ctx.IsStartedHeavyProfiling(), convey.ShouldBeFalse)
	convey.So(ctx.IsDegradation, convey.ShouldBeFalse)
	// no slow in the 5 times degradation
	result.IsSlow = 1
	for i := 0; i < count; i++ {
		processor.handle()
	}
	convey.So(len(ctx.AlgoRes), convey.ShouldEqual, count)
	convey.So(ctx.IsDegradation, convey.ShouldBeFalse)
	convey.So(ctx.IsStartedHeavyProfiling(), convey.ShouldBeTrue)
	// IsSlow is 0, stop heavy profiling
	result.IsSlow = 0
	processor.handle()
	convey.So(ctx.IsStartedHeavyProfiling(), convey.ShouldBeFalse)
	convey.So(ctx.IsDegradation, convey.ShouldBeFalse)
}

func testStartHeavyProfiling(ctx *slownodejob.JobContext) {
	result := &slownode.ClusterAlgoResult{}
	result.IsSlow = 0
	processor := newDegradationProcessor(ctx, result)
	// clear
	processor.handle()
	// is slow, start heavy profiling
	result.IsSlow = 1
	processor.handle()
	convey.So(ctx.IsStartedHeavyProfiling(), convey.ShouldBeTrue)
	// 5 times isSlow, Degradation, stop heavy profiling
	for i := 0; i < 4; i++ {
		processor.handle()
	}
	convey.So(ctx.IsStartedHeavyProfiling(), convey.ShouldBeFalse)
	convey.So(ctx.IsDegradation, convey.ShouldBeTrue)
	// 100 times isSlow, ignore it
	for i := 0; i < 100; i++ {
		processor.handle()
	}
	convey.So(ctx.IsStartedHeavyProfiling(), convey.ShouldBeFalse)
	convey.So(ctx.IsDegradation, convey.ShouldBeTrue)
}

func TestDegradationProcessor2(t *testing.T) {
	// no rank ids, no call reportSlowNode
	// got rankd ids during heavy profiling, call reportSlowNode
	// no more need to call reportSlowNode after called once
	ctx := ctxGenerator()
	patches := mockFunc(ctx)
	defer patches.Reset()
	var callReportSlowNodeCount int
	patches.ApplyPrivateMethod(
		reflect.TypeOf(&degradationProcessor{}),
		"reportSlowNode",
		func(degradationProcessor, *slownodejob.JobContext, *slownode.ClusterAlgoResult) {
			callReportSlowNodeCount++
		})
	convey.Convey("test got SlowCalculateRanks during heavy profiling", t, func() {
		result := &slownode.ClusterAlgoResult{}
		result.IsSlow = 1
		processor := newDegradationProcessor(ctx, result)
		// start heavy profiling
		processor.handle()
		convey.So(ctx.IsStartedHeavyProfiling(), convey.ShouldBeTrue)
		convey.So(callReportSlowNodeCount, convey.ShouldEqual, 0)
		// 4 times degradation, stop heavy profiling
		result.SlowCalculateRanks = []int{0}
		for i := 0; i < 4; i++ {
			processor.handle()
		}
		convey.So(ctx.IsStartedHeavyProfiling(), convey.ShouldBeFalse)
		convey.So(ctx.IsDegradation, convey.ShouldBeTrue)
		convey.So(callReportSlowNodeCount, convey.ShouldEqual, 1)
		// 100 times degradation, no need to call reportSlowNode again
		for i := 0; i < 100; i++ {
			processor.handle()
		}
		convey.So(callReportSlowNodeCount, convey.ShouldEqual, 1)
	})
}

func TestDegradationProcessor3(t *testing.T) {
	ctx := ctxGenerator()
	patches := mockFunc(ctx)
	defer patches.Reset()
	var callReportSlowNodeCount int
	patches.ApplyPrivateMethod(
		reflect.TypeOf(&degradationProcessor{}),
		"reportSlowNode",
		func(degradationProcessor, *slownodejob.JobContext, *slownode.ClusterAlgoResult) {
			callReportSlowNodeCount++
		})
	convey.Convey("test got SlowCalculateRanks after heavy profiling", t, func() {
		callReportSlowNodeCount = 0
		result := &slownode.ClusterAlgoResult{}
		result.IsSlow = 1
		processor := newDegradationProcessor(ctx, result)
		// start heavy profiling
		processor.handle()
		convey.So(ctx.IsStartedHeavyProfiling(), convey.ShouldBeTrue)
		convey.So(callReportSlowNodeCount, convey.ShouldEqual, 0)
		// 4 times degradation, stop heavy profiling
		for i := 0; i < 4; i++ {
			processor.handle()
		}
		convey.So(ctx.IsStartedHeavyProfiling(), convey.ShouldBeFalse)
		convey.So(ctx.IsDegradation, convey.ShouldBeTrue)
		convey.So(callReportSlowNodeCount, convey.ShouldEqual, 0)
		// got rank ids, call reportSlowNode
		result.SlowCalculateRanks = []int{0}
		processor.handle()
		convey.So(callReportSlowNodeCount, convey.ShouldEqual, 1)
		// 100 times degradation, no need to call reportSlowNode again
		for i := 0; i < 100; i++ {
			processor.handle()
		}
		convey.So(callReportSlowNodeCount, convey.ShouldEqual, 1)
	})
}

func TestReportSlowNodet(t *testing.T) {
	ctx := &slownodejob.JobContext{Job: &slownode.Job{}}
	ctx.Job.JobId = "test_job_id"
	ctx.Job.JobName = "test_job_name"
	result := &slownode.ClusterAlgoResult{}
	processor := newDegradationProcessor(ctx, result)
	convey.Convey("test reportSlowNode", t, func() {
		convey.Convey("get grpc client failed", func() {
			mock := gomonkey.ApplyFunc(grpc.GetClient, func() (*grpc.Client, error) {
				fmt.Println("mock get client failed")
				return nil, errors.New("mock get client failed")
			})
			defer mock.Reset()
			output := captureOutput(func() {
				processor.reportSlowNode()
			})
			convey.So(output, convey.ShouldContainSubstring, "mock get client failed")
		})
		patch := gomonkey.ApplyFunc(grpc.GetClient, func() (*grpc.Client, error) {
			return &grpc.Client{}, nil
		})
		defer patch.Reset()
		convey.Convey("test report failed", func() {
			mock := gomonkey.ApplyMethod(
				reflect.TypeOf(&grpc.Client{}),
				"ReportFault",
				func(*grpc.Client, []*pubfault.Fault) error {
					fmt.Println("mock ReportFault failed")
					return errors.New("mock ReportFault failed")
				})
			defer mock.Reset()
			output := captureOutput(func() {
				processor.reportSlowNode()
			})
			convey.So(output, convey.ShouldContainSubstring, "mock ReportFault failed")
		})
		patch.ApplyMethod(
			reflect.TypeOf(&grpc.Client{}),
			"ReportFault",
			func(*grpc.Client, []*pubfault.Fault) error {
				return nil
			})
		processor.reportSlowNode()
	})
}
