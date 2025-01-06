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

// Package dcmi is used to work with Ascend devices
package dcmi

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
)

func TestGetChipInfo(t *testing.T) {
	convey.Convey("Test get chip function", t, func() {
		var cardId int32 = 0
		var devId int32 = 0
		w := NpuWorker{}
		convey.Convey("01-not valid id, should return error", func() {
			gomonkey.ApplyFunc(isValidCardIDAndDeviceID, func(a, b int32) bool {
				return false
			})
			_, err := w.GetChipInfo(cardId, devId)
			convey.ShouldBeError(err)
		})
		convey.Convey("02-get no chip info, should return nil", func() {
			gomonkey.ApplyFunc(isValidCardIDAndDeviceID, func(a, b int32) bool {
				return true
			})
			chip, _ := w.GetChipInfo(cardId, devId)
			convey.ShouldBeNil(chip)
		})
	})
}
