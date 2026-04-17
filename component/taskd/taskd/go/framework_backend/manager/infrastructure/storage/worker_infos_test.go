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
	"time"

	"github.com/smartystreets/goconvey/convey"

	"taskd/toolkit_backend/net/common"
)

func TestWorkerInfos_DeepCopy(t *testing.T) {
	convey.Convey("TestWorkerInfos_DeepCopy", t, func() {
		workerInfos := &WorkerInfos{
			Workers: map[string]*WorkerInfo{
				"worker1": {GlobalRank: "0", RWMutex: sync.RWMutex{}},
				"worker2": nil,
			},
			AllStatus: map[string]string{"worker1": "running"},
			RWMutex:   sync.RWMutex{},
		}
		clone := workerInfos.DeepCopy()
		convey.So(clone.AllStatus["worker1"], convey.ShouldEqual, "running")
		convey.So(clone.Workers["worker2"], convey.ShouldBeNil)
	})
}

func TestWorkerInfo_SetStatusVal(t *testing.T) {
	convey.Convey("TestWorkerInfo_SetStatusVal", t, func() {
		workerInfo := &WorkerInfo{
			Status:  map[string]string{},
			RWMutex: sync.RWMutex{},
		}
		workerInfo.SetStatusVal("key1", "val1")
		convey.So(workerInfo.Status["key1"], convey.ShouldEqual, "val1")
	})
}

func TestWorkerInfo_DeepCopy(t *testing.T) {
	convey.Convey("TestWorkerInfo_DeepCopy", t, func() {
		convey.Convey("should deep copy with pos", func() {
			workerInfo := &WorkerInfo{
				Config:     map[string]string{"c1": "v1"},
				Actions:    map[string]string{"a1": "v1"},
				FaultInfo:  map[string]string{"f1": "v1"},
				Status:     map[string]string{"s1": "v1"},
				GlobalRank: "0",
				HeartBeat:  time.Now(),
				Pos:        &common.Position{Role: "worker"},
				RWMutex:    sync.RWMutex{},
			}
			clone := workerInfo.DeepCopy()
			convey.So(clone.GlobalRank, convey.ShouldEqual, "0")
			convey.So(clone.Pos.Role, convey.ShouldEqual, "worker")
		})
	})
}
