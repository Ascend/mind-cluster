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

package controllerflags

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestGetState(t *testing.T) {
	convey.Convey("test get state", t, func() {
		convey.Convey("get state", func() {
			state := IsControllerExited.GetState()
			if state == true {
				convey.So(state, convey.ShouldBeTrue)
			} else {
				convey.So(state, convey.ShouldBeFalse)
			}
		})
	})
}

func TestSetState(t *testing.T) {
	convey.Convey("test set state", t, func() {
		convey.Convey("set state", func() {
			state := IsControllerExited.GetState()
			IsControllerExited.SetState(state)
			if state == true {
				convey.So(state, convey.ShouldBeTrue)
			} else {
				convey.So(state, convey.ShouldBeFalse)
			}
		})
	})
}
