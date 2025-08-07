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

package funchandler

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"ascend-faultdiag-online/pkg/core/model"
)

func TestGenerateExecuteFunc(t *testing.T) {
	convey.Convey("Test GenerateExecuteFunc", t, func() {
		convey.Convey("test normal case", func() {
			mockExecFunc := func(model.Input) int {
				return 0
			}
			execFunc := GenerateExecuteFunc(mockExecFunc, "testApp")
			ret, err := execFunc(model.Input{})
			convey.So(ret, convey.ShouldEqual, 0)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("test error case", func() {
			mockExecFunc := func(model.Input) int {
				return 1
			}
			execFunc := GenerateExecuteFunc(mockExecFunc, "testApp")
			ret, err := execFunc(model.Input{})
			convey.So(ret, convey.ShouldEqual, -1)
			convey.So(err.Error(), convey.ShouldEqual, "call [testApp] func [Execute] failed, return code: [1]")
		})
	})
}
