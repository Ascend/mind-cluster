/* Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
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

package utils

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestGenerateUUIDv4(t *testing.T) {
	convey.Convey("TestGenerateUUIDv4", t, func() {
		id, err := generateUUIDv4()
		convey.So(err, convey.ShouldBeNil)
		convey.So(len(id), convey.ShouldEqual, 36)
		convey.So(id[8], convey.ShouldEqual, '-')
		convey.So(id[13], convey.ShouldEqual, '-')
		convey.So(id[18], convey.ShouldEqual, '-')
		convey.So(id[23], convey.ShouldEqual, '-')
		convey.So(id[14], convey.ShouldEqual, '4')
		convey.So(string(id[19]), convey.ShouldBeIn, "8", "9", "a", "b")

		id2, err2 := generateUUIDv4()
		convey.So(err2, convey.ShouldBeNil)
		convey.So(id2, convey.ShouldNotEqual, id)
	})
}
