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

package silentfault

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"clusterd/pkg/common/constant"
)

func TestGetSilentFaultCmInfoForMerge(t *testing.T) {
	convey.Convey("get silent fault cm info for merge", t, func() {
		SilentFaultCmInfo.Reset()
		SilentFaultCmInfo.Upsert("n1", "silent-fault-n1", "clusterd", []string{"Ascend910-1"})
		merged := GetSilentFaultCmInfoForMerge()
		convey.So(len(merged), convey.ShouldEqual, 1)
		convey.So(merged["n1"].Detail["Ascend910-1"][0].FaultLevel, convey.ShouldEqual, constant.SilentFault)
		SilentFaultCmInfo.Reset()
	})
}

func TestResetCacheAndIsEmpty(t *testing.T) {
	convey.Convey("reset cache and is empty", t, func() {
		SilentFaultCmInfo.Upsert("n1", "f1", "clusterd", []string{"Ascend910-0"})
		convey.So(SilentFaultIsEmpty(), convey.ShouldBeFalse)
		ResetCache()
		convey.So(SilentFaultIsEmpty(), convey.ShouldBeTrue)
	})
}
