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

// Package ascendfaultdiagonline is DT collection for func in app.go
package ascendfaultdiagonline

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/core/context"
	"ascend-faultdiag-online/pkg/core/model/enum"
	"ascend-faultdiag-online/pkg/register"
	"ascend-faultdiag-online/pkg/service/servicefunc/slownode"
	"ascend-faultdiag-online/pkg/utils"
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

func TestStartFDOnline(t *testing.T) {
	patches := gomonkey.ApplyFunc(
		context.NewFaultDiagContext,
		func(configPath string) (*context.FaultDiagContext, error) {
			return &context.FaultDiagContext{}, nil
		})
	defer patches.Reset()
	patches.ApplyFunc(register.Setup, func(*context.FaultDiagContext) {})
	patches.ApplyMethod(reflect.TypeOf(&context.FaultDiagContext{}), "StartService", func(*context.FaultDiagContext) {})
	patches.ApplyFunc(utils.WriteStartInfo, func() {})
	convey.Convey("test StartFDOnline", t, func() {
		// wrong app name
		apps := []string{"test"}
		StartFDOnline("", apps, "test")

		// right name
		patches.ApplyFunc(slownode.StartSlowNode, func(enum.DeployMode) {})
		apps = append(apps, enum.SlowNode)
		StartFDOnline("", apps, "test")
	})
}
