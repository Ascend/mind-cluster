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

// Package node is a DT collection for func in callback_processor
package node

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"testing"
	"unsafe"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	corev1 "k8s.io/api/core/v1"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/model/slownode"
	"ascend-faultdiag-online/pkg/service/servicefunc/slownode/algo"
	"ascend-faultdiag-online/pkg/service/servicefunc/slownode/slownodejob"
	"ascend-faultdiag-online/pkg/utils"
	"ascend-faultdiag-online/pkg/utils/fileutils"
	"ascend-faultdiag-online/pkg/utils/k8s"
)

const logLineLength = 256

func init() {
	config := hwlog.LogConfig{
		OnlyToStdout:  true,
		MaxLineLength: logLineLength,
	}
	err := hwlog.InitRunLogger(&config, context.TODO())
	if err != nil {
		fmt.Println(err)
	}
}

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
	var message = `{"slownode_testJobName":{"127.0.0.1":{"isSlow":1,"degradationLevel":"10.0%","jobId":"testJobId"` +
		`,"jobName":"testJobName","nodeRank":"127.0.0.1","slowCalculateRanks":null,"slowCommunicationDomains":null,` +
		`"slowSendRanks":null,"slowHostNodes":null,"slowIORanks":null}}}`
	var ctx = &slownodejob.JobContext{Job: &slownode.Job{}}
	ctx.Job.JobId = "testJobId"
	defer slownodejob.GetJobCtxMap().Delete("testJobId")
	convey.Convey("test AlgoCallbackProcessor", t, func() {
		testAlgoCallbackProcessorWithError(ctx, message)
	})
}

func testAlgoCallbackProcessorWithError(ctx *slownodejob.JobContext, message string) {
	convey.Convey("test unmarshal failed", func() {
		AlgoCallbackProcessor("}")
	})
	convey.Convey("test convert failed", func() {
		AlgoCallbackProcessor("{}")
	})
	convey.Convey("test job is not exist or not running", func() {
		// jos is not existed
		AlgoCallbackProcessor(message)
		// job is not running
		slownodejob.GetJobCtxMap().Insert("testJobId", ctx)
		AlgoCallbackProcessor(message)
	})
	convey.Convey("test normal case", func() {
		slownodejob.GetJobCtxMap().Insert("testJobId", ctx)
		setUnexportedFiled(ctx, "isRunning", true)
		AlgoCallbackProcessor(message)
	})
}

func TestDataParseCallbackProcessor(t *testing.T) {
	var message = `{"jobId":"testJobId","jobName":"testJobName","stepCount":20,"isFinished":true,"rankIds":["1","2"]}`
	var ctx = &slownodejob.JobContext{Job: &slownode.Job{}}
	ctx.Job.JobId = "testJobId"
	convey.Convey("Test DataParseCallbackProcessor", t, func() {
		testDataParseCallbackProcessorWithError(ctx, message)
		testDataParseCallbackProcessorWithSuccess(ctx, message)
	})
}

func testDataParseCallbackProcessorWithError(ctx *slownodejob.JobContext, message string) {
	convey.Convey("test json unmarshal failed", func() {
		DataParseCallbackProcessor("}")
	})

	convey.Convey("test ctx failed or not running", func() {
		DataParseCallbackProcessor(message)
		slownodejob.GetJobCtxMap().Insert("testJobId", ctx)
		// ctx is existed but not running
		setUnexportedFiled(ctx, "isRunning", false)
		DataParseCallbackProcessor(message)
		// ctx is running but wrong data of callback data
		setUnexportedFiled(ctx, "isRunning", true)
		// callback data is not finished
		message = `{"jobId":"testJobId","jobName":"testJobName","stepCount":20,"isFinished":false,` +
			`"rankIds":["1","2"]}`
		DataParseCallbackProcessor(message)
		// stepCount is less than 20
		message = `{"jobId":"testJobId","jobName":"testJobName","stepCount":10,"isFinished":true,` +
			`"rankIds":["1","2"]}`
		DataParseCallbackProcessor(message)
		// ctx step is not NodeStep1
		message = `{"jobId":"testJobId","jobName":"testJobName","stepCount":20,"isFinished":true,` +
			`"rankIds":["1","2"]}`
		setUnexportedFiled(ctx, "step", slownodejob.NodeStep2)
		DataParseCallbackProcessor(message)
	})

	convey.Convey("test data reporting failed case", func() {
		patch := gomonkey.ApplyFunc(dataProfilingReport, func(*slownodejob.JobContext) error {
			fmt.Println("mock dataProfilingReport failed")
			return errors.New("mock dataProfilingReport failed")
		})
		defer patch.Reset()
		setUnexportedFiled(ctx, "step", slownodejob.NodeStep1)
		output := captureOutput(func() {
			DataParseCallbackProcessor(message)
		})
		convey.So(output, convey.ShouldContainSubstring, "mock dataProfilingReport failed")
	})
}

func testDataParseCallbackProcessorWithSuccess(ctx *slownodejob.JobContext, message string) {
	convey.Convey("test success case", func() {

		patch := gomonkey.ApplyFunc(dataProfilingReport, func(*slownodejob.JobContext) error { return nil })
		patch.ApplyMethod(reflect.TypeOf(&algo.Controller{}), "Start", func(*algo.Controller) {
			fmt.Println("mock start algo")
		})
		defer patch.Reset()

		slownodejob.GetJobCtxMap().Insert("testJobId", ctx)
		setUnexportedFiled(ctx, "step", slownodejob.NodeStep1)
		setUnexportedFiled(ctx, "isRunning", true)
		output := captureOutput(func() {
			DataParseCallbackProcessor(message)
		})
		DataParseCallbackProcessor(message)
		convey.So(output, convey.ShouldContainSubstring, "mock start algo")
		convey.So(ctx.Step(), convey.ShouldEqual, slownodejob.NodeStep2)
		convey.So(reflect.DeepEqual(ctx.RealRankIds, []string{"1", "2"}), convey.ShouldBeTrue)
	})
}

func TestDataProfilingReport(t *testing.T) {
	ctx := &slownodejob.JobContext{
		Job: &slownode.Job{},
	}
	ctx.Job.JobId = "testJobId"
	ctx.Job.JobName = "testJobName"
	convey.Convey("Test DataProfilingReport", t, func() {
		convey.Convey("test parallelGroupInfoReader failed", func() {
			patch := gomonkey.ApplyFunc(parallelGroupInfoReader, func(string) (map[string]any, error) {
				return nil, errors.New("mock parallelGroupInfoReader failed")
			})
			defer patch.Reset()
			err := dataProfilingReport(ctx)
			convey.So(err.Error(), convey.ShouldContainSubstring, "mock parallelGroupInfoReader failed")
		})
		mock := gomonkey.ApplyFunc(parallelGroupInfoReader, func(string) (map[string]any, error) {
			return map[string]any{"group1": "node1", "group2": "node2"}, nil
		})
		defer mock.Reset()

		convey.Convey("test get node ip failed", func() {
			patch := gomonkey.ApplyFunc(utils.GetNodeIp, func() (string, error) {
				return "", errors.New("mock get node ip failed")
			})
			defer patch.Reset()
			err := dataProfilingReport(ctx)
			convey.So(err.Error(), convey.ShouldContainSubstring, "mock get node ip failed")
		})
		mock.ApplyFunc(utils.GetNodeIp, func() (string, error) { return "127.0.0.1", nil })
		mock.ApplyFunc(createOrUpdateCM, func(string, string, string, []byte) error { return nil })
		mock.ApplyFunc(k8s.GetClient, func() (*k8s.Client, error) { return &k8s.Client{}, nil })
		mock.ApplyMethod(reflect.TypeOf(&k8s.Client{}), "DeleteConfigMap",
			func(*k8s.Client, string, string) error {
				return nil
			})

		convey.Convey("test dataProfilingReport success", func() {
			err := dataProfilingReport(ctx)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestCallbackReporter(t *testing.T) {
	var data1 = &slownode.NodeAlgoResult{}
	var data2 = &slownode.NodeDataProfilingResult{}
	var ctx = &slownodejob.JobContext{
		Job: &slownode.Job{},
	}
	convey.Convey("test callbackReporter", t, func() {
		testCallbackReportError(data1, ctx)
		testCallbackReportSuccess(data1, data2, ctx)
	})
}

func testCallbackReportSuccess(
	data1 *slownode.NodeAlgoResult,
	data2 *slownode.NodeDataProfilingResult,
	ctx *slownodejob.JobContext) {
	convey.Convey("test report success", func() {
		patch := gomonkey.ApplyFunc(createOrUpdateCM, func(string, string, string, []byte) error {
			return nil
		})
		patch.ApplyFunc(k8s.GetClient, func() (*k8s.Client, error) {
			return &k8s.Client{}, nil
		})
		patch.ApplyMethod(reflect.TypeOf(&k8s.Client{}), "DeleteConfigMap",
			func(*k8s.Client, string, string) error {
				return nil
			})
		defer patch.Reset()
		err := callbackReport(data1, "", "", ctx)
		convey.So(err, convey.ShouldBeNil)
		err = callbackReport(data2, "", "", ctx)
		convey.So(err, convey.ShouldBeNil)
	})
}

func testCallbackReportError(data *slownode.NodeAlgoResult, ctx *slownodejob.JobContext) {
	convey.Convey("test json.MarshalIndent failed", func() {
		patch := gomonkey.ApplyFunc(json.MarshalIndent, func(any, string, string) ([]byte, error) {
			return nil, errors.New("json marshal indent faield")
		})
		defer patch.Reset()
		err := callbackReport(data, "", "", ctx)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "json marshal indent faield")
	})
	convey.Convey("test createOrUpdateCm failed", func() {
		patch := gomonkey.ApplyFunc(createOrUpdateCM, func(string, string, string, []byte) error {
			return errors.New("createOrUpdateCm failed")
		})
		defer patch.Reset()
		err := callbackReport(data, "", "", ctx)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "createOrUpdateCm failed")
	})
	mock := gomonkey.ApplyFunc(createOrUpdateCM, func(string, string, string, []byte) error {
		return nil
	})
	defer mock.Reset()
	convey.Convey("test get k8s client failed", func() {
		patch := gomonkey.ApplyFunc(k8s.GetClient, func() (*k8s.Client, error) {
			return nil, errors.New("get k8s client failed")
		})
		defer patch.Reset()
		err := callbackReport(data, "", "", ctx)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "get k8s client faile")
	})
	mock.ApplyFunc(k8s.GetClient, func() (*k8s.Client, error) {
		return &k8s.Client{}, nil
	})
	convey.Convey("test DeleteConfigMap failed", func() {
		patch := gomonkey.ApplyMethod(reflect.TypeOf(&k8s.Client{}), "DeleteConfigMap",
			func(*k8s.Client, string, string) error {
				return errors.New("deleteConfigMap failed")
			})
		defer patch.Reset()
		err := callbackReport(data, "", "", ctx)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "deleteConfigMap failed")
	})
}

func TestCreateOrUpdateCM(t *testing.T) {
	convey.Convey("test createOrUpdateCM", t, func() {
		convey.Convey("test create k8s client failed", func() {
			patch := gomonkey.ApplyFunc(k8s.GetClient, func() (*k8s.Client, error) {
				return nil, errors.New("get k8s client failed")
			})
			defer patch.Reset()
			err := createOrUpdateCM("", "", "", []byte{})
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldContainSubstring, "get k8s client failed")
		})
		mock := gomonkey.ApplyFunc(k8s.GetClient, func() (*k8s.Client, error) {
			return &k8s.Client{}, nil
		})
		defer mock.Reset()
		convey.Convey("test failed", func() {
			patch := gomonkey.ApplyMethod(reflect.TypeOf(&k8s.Client{}), "CreateOrUpdateConfigMap",
				func(*k8s.Client, *corev1.ConfigMap) error {
					return errors.New("createOrUpdateConfigMap failed")
				})
			defer patch.Reset()
			err := createOrUpdateCM("", "", "", []byte{})
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldContainSubstring, "createOrUpdateConfigMap failed")
		})
		convey.Convey("test success", func() {
			patch := gomonkey.ApplyMethod(reflect.TypeOf(&k8s.Client{}), "CreateOrUpdateConfigMap",
				func(*k8s.Client, *corev1.ConfigMap) error {
					return nil
				})
			defer patch.Reset()
			err := createOrUpdateCM("", "", "testKey", []byte("this is value"))
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestParallelGroupInfoReader(t *testing.T) {
	convey.Convey("test parallelGroupInfoReader", t, func() {
		convey.Convey("test read file failed", func() {
			patch := gomonkey.ApplyFuncReturn(fileutils.ReadLimitBytes, nil, errors.New("read file failed"))
			defer patch.Reset()
			data, err := parallelGroupInfoReader("testJobId")
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(data, convey.ShouldBeNil)
		})
		convey.Convey("test unmarshal failed", func() {
			patch := gomonkey.ApplyFuncReturn(fileutils.ReadLimitBytes, []byte("invalid json"), nil)
			defer patch.Reset()
			data, err := parallelGroupInfoReader("testJobId")
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(data, convey.ShouldBeNil)
		})
		convey.Convey("test success", func() {
			mockData := `{"group1": "node1", "group2": "node2"}`
			patch := gomonkey.ApplyFuncReturn(fileutils.ReadLimitBytes, []byte(mockData), nil)
			defer patch.Reset()
			data, err := parallelGroupInfoReader("testJobId")
			convey.So(err, convey.ShouldBeNil)
			convey.So(data, convey.ShouldResemble, map[string]any{"group1": "node1", "group2": "node2"})
		})
	})
}
