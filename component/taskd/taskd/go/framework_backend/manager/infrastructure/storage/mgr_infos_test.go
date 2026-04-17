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

package storage

import (
	"sync"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestMgrInfo_SetStatusVal(t *testing.T) {
	convey.Convey("TestMgrInfo_SetStatusVal", t, func() {
		mgrInfo := &MgrInfo{
			Status:  map[string]string{},
			RWMutex: sync.RWMutex{},
		}
		mgrInfo.SetStatusVal("key1", "val1")
		convey.So(mgrInfo.Status["key1"], convey.ShouldEqual, "val1")
	})
}

func TestMgrInfo_GetStatusVal(t *testing.T) {
	convey.Convey("TestMgrInfo_GetStatusVal", t, func() {
		mgrInfo := &MgrInfo{
			Status:  map[string]string{"key1": "val1"},
			RWMutex: sync.RWMutex{},
		}
		convey.Convey("should return value for existing key", func() {
			val, ok := mgrInfo.GetStatusVal("key1")
			convey.So(ok, convey.ShouldBeTrue)
			convey.So(val, convey.ShouldEqual, "val1")
		})
	})
}

func TestMgrInfo_DeepCopy(t *testing.T) {
	convey.Convey("TestMgrInfo_DeepCopy", t, func() {
		mgrInfo := &MgrInfo{
			Status:  map[string]string{"key1": "val1"},
			RWMutex: sync.RWMutex{},
		}
		clone := mgrInfo.DeepCopy()
		convey.So(clone.Status["key1"], convey.ShouldEqual, "val1")
	})
}
