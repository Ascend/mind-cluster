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

// Package register is a DT collection for func in funcmap
package register

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"ascend-faultdiag-online/pkg/core/context"
)

func TestRegisterFunctionMap(t *testing.T) {
	convey.Convey("Test RegisterFuncMap", t, func() {
		convey.Convey("test normal case", func() {
			fdCtx := &context.FaultDiagContext{}
			functionMap(fdCtx)
			convey.So(fdCtx.FuncHandler, convey.ShouldNotBeNil)
		})

		convey.Convey("test error case", func() {
			functionMap(nil)
		})
	})
}
