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

// Package algo is a DT collection for func in algo
package algo

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/core/context"
	"ascend-faultdiag-online/pkg/core/model/enum"
	"ascend-faultdiag-online/pkg/model/slownode"
	"ascend-faultdiag-online/pkg/service/servicefunc/slownode/slownodejob"
)

const logLineLength = 256

func init() {
	config := hwlog.LogConfig{
		OnlyToStdout:  true,
		MaxLineLength: logLineLength,
	}
	err := hwlog.InitRunLogger(&config, nil)
	if err != nil {
		fmt.Println(err)
	}
}

func TestAlgo(t *testing.T) {
	convey.Convey("Test Algo", t, func() {
		ctx := &slownodejob.JobContext{
			Job: &slownode.Job{},
		}
		testRequestWithSuccess(ctx)
		testReqeuestWithError(ctx)
		testStartAndStop(ctx)
	})
}

func testRequestWithSuccess(ctx *slownodejob.JobContext) {
	convey.Convey("test request success", func() {
		mock := gomonkey.ApplyMethod(
			reflect.TypeOf(context.FdCtx),
			"Request",
			func(*context.FaultDiagContext, string, string) (string, error) {
				return `{"status":"success","msg":"","data":{}}`, nil
			})
		defer mock.Reset()
		err := NewController(ctx).request(enum.Start)
		convey.So(err, convey.ShouldBeNil)

		// test node deployment
		ctx.Deployment = enum.Node
		err = NewController(ctx).request(enum.Start)
		convey.So(err, convey.ShouldBeNil)
	})
}

func testReqeuestWithError(ctx *slownodejob.JobContext) {
	convey.Convey("test request marshal failed", func() {
		// marshal failed
		mock := gomonkey.ApplyFunc(json.Marshal, func(any) ([]byte, error) {
			return nil, errors.New("mock marshal error")
		})
		defer mock.Reset()
		err := NewController(ctx).request(enum.Start)
		convey.So(err.Error(), convey.ShouldEqual, "mock marshal error")
	})

	convey.Convey("test request FdCtx.Request failed", func() {
		// marshal failed
		mock := gomonkey.ApplyMethod(
			reflect.TypeOf(context.FdCtx),
			"Request",
			func(*context.FaultDiagContext, string, string) (string, error) {
				return "", errors.New("mock request error")
			})
		defer mock.Reset()
		err := NewController(ctx).request(enum.Start)
		convey.So(err.Error(), convey.ShouldEqual, "mock request error")
	})

	convey.Convey("test request unmarshal failed", func() {
		// marshal failed
		mock := gomonkey.ApplyMethod(
			reflect.TypeOf(context.FdCtx),
			"Request",
			func(*context.FaultDiagContext, string, string) (string, error) {
				return "}", nil
			})
		defer mock.Reset()
		err := NewController(ctx).request(enum.Start)
		convey.So(err.Error(), convey.ShouldEqual, "invalid character '}' looking for beginning of value")
	})

	convey.Convey("test request failed", func() {
		// marshal failed
		mock := gomonkey.ApplyMethod(
			reflect.TypeOf(context.FdCtx),
			"Request",
			func(*context.FaultDiagContext, string, string) (string, error) {
				return `{"status":"error","msg":"request failed","data":{}}`, nil
			})
		defer mock.Reset()
		err := NewController(ctx).request(enum.Start)
		convey.So(err.Error(), convey.ShouldEqual, "request failed")
	})
}

func testStartAndStop(ctx *slownodejob.JobContext) {
	convey.Convey("test start and stop success", func() {
		mock := gomonkey.ApplyPrivateMethod(
			reflect.TypeOf(&Controller{}),
			"request",
			func(*Controller, enum.Command) error {
				return nil
			})
		defer mock.Reset()
		NewController(ctx).Start()
		NewController(ctx).Stop()
	})

	convey.Convey("test start and stop failed", func() {
		mock := gomonkey.ApplyPrivateMethod(
			reflect.TypeOf(&Controller{}),
			"request",
			func(*Controller, enum.Command) error {
				return errors.New("mock request error")
			})
		defer mock.Reset()
		NewController(ctx).Start()
		NewController(ctx).Stop()
	})
}
