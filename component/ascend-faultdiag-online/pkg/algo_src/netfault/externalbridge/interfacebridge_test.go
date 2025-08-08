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

// Package externalbridge for node and cluster level detection interact interface
package externalbridge

import (
	"context"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/core/model"
	"ascend-faultdiag-online/pkg/core/model/enum"
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

// TestSwitchCommand
func TestSwitchCommand(t *testing.T) {
	input := &model.Input{}
	convey.Convey("TestSwitchCommand", t, func() {
		convey.Convey("func is nil", func() {
			input.Command = enum.Register
			input.Func = nil
			_, flag := switchCommand(input)
			convey.So(flag, convey.ShouldBeFalse)
		})
		convey.Convey("start api", func() {
			input.Command = enum.Start
			_, flag := switchCommand(input)
			convey.So(flag, convey.ShouldBeTrue)
		})
		convey.Convey("stop api", func() {
			input.Command = enum.Stop
			_, flag := switchCommand(input)
			convey.So(flag, convey.ShouldBeTrue)
		})
		convey.Convey("reload api", func() {
			input.Command = enum.Reload
			_, flag := switchCommand(input)
			convey.So(flag, convey.ShouldBeTrue)
		})
		convey.Convey("invalid api", func() {
			input.Command = "invalid"
			_, flag := switchCommand(input)
			convey.So(flag, convey.ShouldBeFalse)
		})
	})
}

// TestCheckInputInvalid test for func checkInputInvalid
func TestCheckInputInvalid(t *testing.T) {
	input := &model.Input{}
	convey.Convey("Test checkInputInvalid", t, func() {
		convey.Convey("should return false when command is invalid", func() {
			input.Command = "invalid"
			ret := checkInputInvalid(input)
			convey.So(ret, convey.ShouldBeFalse)
		})

		convey.Convey("should return false when register function is nil", func() {
			input.Command = enum.Register
			input.Func = nil
			ret := checkInputInvalid(input)
			convey.So(ret, convey.ShouldBeFalse)
		})

		convey.Convey("should true when register function is valid", func() {
			input.Command = enum.Register
			input.Func = func(string) {}
			ret := checkInputInvalid(input)
			convey.So(ret, convey.ShouldBeTrue)
		})
	})
}

// TestExecuteFailCases test for func Execute
func TestExecuteFailCases(t *testing.T) {
	input := &model.Input{}
	convey.Convey("Test Execute Fail Cases", t, func() {
		convey.Convey("should return -1 when command is empty", func() {
			ret := Execute(input)
			convey.So(ret, convey.ShouldEqual, -1)
		})

		convey.Convey("should return -1 when command is invalid", func() {
			input.Command = "invalid"
			ret := Execute(input)
			convey.So(ret, convey.ShouldEqual, -1)
		})

		convey.Convey("should return -1 when command is registerCallBack with no func", func() {
			input.Command = enum.Register
			input.Func = nil
			patch := gomonkey.ApplyFunc(checkInputInvalid, func(input *model.Input) bool {
				return true
			})
			defer patch.Reset()
			ret := Execute(input)
			convey.So(ret, convey.ShouldEqual, -1)
		})
	})
}
