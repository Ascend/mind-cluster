/* Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.
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

package common

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

// TestDeepCopyHccsBandwidthInfo TestDeepCopySlice
func TestDeepCopyHccsBandwidthInfo(t *testing.T) {

	convey.Convey("should copy a new []int", t, func() {
		slice := []int{1, 2}
		newSlice := deepCopySlice(slice)
		convey.So(&newSlice, convey.ShouldNotEqual, &slice)
	})

	convey.Convey("should copy a new []int32", t, func() {
		slice := []uint32{1, 2}

		newSlice := deepCopySlice(slice)
		convey.So(&newSlice, convey.ShouldNotEqual, &slice)
	})

	convey.Convey("should copy a new []float64", t, func() {
		slice := []float64{1, 2}
		newSlice := deepCopySlice(slice)
		convey.So(&newSlice, convey.ShouldNotEqual, &slice)
	})
}
