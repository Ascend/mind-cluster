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

// Package device a series of device function
package device

import (
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"Ascend-device-plugin/pkg/common"
)

// TestAscendToolsMethodSetRackID test set rack id
func TestAscendToolsMethodSetRackID(t *testing.T) {
	convey.Convey("test AscendTools method SetRackID", t, func() {
		convey.Convey("01-should success when set rack id", func() {
			tool := mockAscendTools()
			theRackID := int32(1)
			tool.SetRackID(theRackID)
			convey.So(tool.GetRackID(), convey.ShouldEqual, theRackID)
		})
	})
}

func TestAscendToolsMethodWriteNodeDeviceInfoDataA5(t *testing.T) {
	convey.Convey("test AscendTools method writeNodeDeviceInfoDataA5", t, func() {
		convey.Convey("01-should return false when write cm failed", func() {
			tool := mockAscendTools()
			patch := gomonkey.ApplyMethodReturn(tool.client, "WriteDeviceInfoDataIntoCMCacheA5",
				errors.New("write cm failed"))
			defer patch.Reset()
			ret, _ := tool.writeNodeDeviceInfoDataA5(map[string]string{}, "", common.SwitchFaultInfo{})
			convey.So(ret, convey.ShouldBeFalse)
		})
		convey.Convey("02-should return true when write cm success", func() {
			tool := mockAscendTools()
			patch := gomonkey.ApplyMethodReturn(tool.client, "WriteDeviceInfoDataIntoCMCacheA5", nil)
			defer patch.Reset()
			ret, _ := tool.writeNodeDeviceInfoDataA5(map[string]string{}, "", common.SwitchFaultInfo{})
			convey.So(ret, convey.ShouldBeTrue)
		})
	})
}
