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
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"Ascend-device-plugin/pkg/device/deviceswitch"
	"ascend-common/devmanager"
)

// TestInitDevManager for test initDevManager
func TestInitDevManager(t *testing.T) {
	convey.Convey("test initDevManager", t, func() {
		convey.Convey("test init dcmi manager failed, should return err", func() {
			devM, _, err := initDevManager()
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(devM, convey.ShouldBeNil)
		})
		convey.Convey("test init lqdcmi manager failed, "+
			"err should be nil and lq manager should be nil", func() {
			devM := &devmanager.DeviceManager{DevType: "test1"}
			p1 := gomonkey.ApplyFuncReturn(devmanager.AutoInit, devM, nil)
			defer p1.Reset()
			devM, lqM, err := initDevManager()
			convey.So(err, convey.ShouldBeNil)
			convey.So(devM, convey.ShouldNotBeNil)
			convey.So(lqM, convey.ShouldBeNil)
		})
		convey.Convey("test init manager success,"+
			" err should be nil and lq manager should not be nil", func() {
			devM := &devmanager.DeviceManager{DevType: "test1"}
			p1 := gomonkey.ApplyFuncReturn(devmanager.AutoInit, devM, nil)
			p1.ApplyMethodReturn(deviceswitch.NewSwitchDevManager(), "InitSwitchDev", nil)
			defer p1.Reset()
			devM, lqM, err := initDevManager()
			convey.So(err, convey.ShouldBeNil)
			convey.So(devM, convey.ShouldNotBeNil)
			convey.So(lqM, convey.ShouldNotBeNil)
		})
	})
}
