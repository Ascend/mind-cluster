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

// Package devicefactory a series of driver manager init function
package devicefactory

import (
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"Ascend-device-plugin/pkg/device/deviceswitch"
	"ascend-common/api"
	"ascend-common/common-utils/utils"
	"ascend-common/devmanager"
)

// TestInitDevManager for test initDevManager
func TestInitDevManager(t *testing.T) {
	convey.Convey("test initDevManager", t, func() {
		convey.Convey("test init dcmi manager failed, should return err", func() {
			devM, err := initDevManager()
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(devM, convey.ShouldBeNil)
		})
	})
}

// TestInitSwitchDevManager for test initSwitchDevManager
func TestInitSwitchDevManager(t *testing.T) {
	convey.Convey("test initSwitchDevManager", t, func() {
		convey.Convey("910A5 should skip without loading the switch library", func() {
			libPathCalled := false
			p1 := gomonkey.ApplyFunc(utils.GetDriverLibPath, func(string) (string, error) {
				libPathCalled = true
				return "", nil
			})
			defer p1.Reset()
			devM := &devmanager.DeviceManager{DevType: api.Ascend910A5}
			convey.So(initSwitchDevManager(devM), convey.ShouldBeNil)
			convey.So(libPathCalled, convey.ShouldBeFalse)
		})
		convey.Convey("non-A3 device should still try to load the switch library", func() {
			p1 := gomonkey.ApplyFuncReturn(utils.GetDriverLibPath, "", errors.New("lib not found"))
			defer p1.Reset()
			devM := &devmanager.DeviceManager{DevType: api.Ascend910B}
			convey.So(initSwitchDevManager(devM), convey.ShouldBeNil)
		})
		convey.Convey("910A3 without switch library should return nil manager", func() {
			p1 := gomonkey.ApplyFuncReturn(utils.GetDriverLibPath, "", errors.New("lib not found"))
			defer p1.Reset()
			devM := &devmanager.DeviceManager{DevType: api.Ascend910A3}
			convey.So(initSwitchDevManager(devM), convey.ShouldBeNil)
		})
		convey.Convey("910A3 with switch library inited should return manager", func() {
			p1 := gomonkey.ApplyMethodReturn(deviceswitch.NewSwitchDevManager(), "InitSwitchDev", nil)
			defer p1.Reset()
			devM := &devmanager.DeviceManager{DevType: api.Ascend910A3}
			convey.So(initSwitchDevManager(devM), convey.ShouldNotBeNil)
		})
	})
}
