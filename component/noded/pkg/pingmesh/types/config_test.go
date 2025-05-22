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

// Package types for
package types

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestDeepClone(t *testing.T) {
	convey.Convey("Testing HccspingMeshPolicy DeepClone", t, func() {
		convey.Convey("01.deep copy empty success", func() {
			current := &HccspingMeshPolicy{}
			res := current.DeepCopy()
			convey.So(res.UID, convey.ShouldEqual, current.UID)
		})

		convey.Convey("02.deep copy big data success", func() {
			current := &HccspingMeshPolicy{
				Config: &HccspingMeshConfig{
					Activate:     ActivateOn,
					TaskInterval: 1,
				},
				Address: map[string]SuperDeviceIDs{
					"1": SuperDeviceIDs{},
				},
				DestAddr: map[string]DestinationAddress{
					"1": DestinationAddress{},
				},
				UID: "testId",
			}
			res := current.DeepCopy()
			convey.So(res.UID, convey.ShouldEqual, current.UID)
			convey.So(len(res.Address), convey.ShouldEqual, len(current.Address))
			convey.So(len(res.DestAddr), convey.ShouldEqual, len(current.DestAddr))
		})
	})
}
