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

// Package devicefactory a series of entry test function
package devicefactory

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"Ascend-device-plugin/pkg/common"
	"Ascend-device-plugin/pkg/device/deviceswitch"
	"Ascend-device-plugin/pkg/server"
	"ascend-common/devmanager"
)

// TestInitFunction for test InitFunction
func TestInitFunction(t *testing.T) {
	convey.Convey("test InitFunction", t, func() {
		convey.Convey("test initDevManager failed, should return err", func() {
			_, err := InitFunction()
			convey.So(err, convey.ShouldNotBeNil)
		})
		mockDevManager := &devmanager.DeviceManager{}
		convey.Convey("test initDevManager success but NewHwDevManager failed, should return err", func() {
			p1 := gomonkey.ApplyFuncReturn(initDevManager, mockDevManager, nil, nil)
			defer p1.Reset()
			_, err := InitFunction()
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("test initDevManager success and NewHwDevManager success, err should be nil", func() {
			p1 := gomonkey.ApplyFuncReturn(initDevManager, mockDevManager, nil, nil)
			p1.ApplyFuncReturn(server.NewHwDevManager, &server.HwDevManager{})
			defer p1.Reset()
			_, err := InitFunction()
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("test switchDevM is not nil, EnableSwitchFault should be true", func() {
			p1 := gomonkey.ApplyFuncReturn(initDevManager, mockDevManager, &deviceswitch.SwitchDevManager{}, nil)
			p1.ApplyFuncReturn(server.NewHwDevManager, &server.HwDevManager{})
			defer p1.Reset()
			_, _ = InitFunction()
			convey.So(common.ParamOption.EnableSwitchFault, convey.ShouldBeTrue)
		})
	})
}
