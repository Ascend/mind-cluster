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

func TestClusterInfos_GetCluster(t *testing.T) {
	convey.Convey("TestClusterInfos_GetCluster", t, func() {
		clusterInfos := &ClusterInfos{
			Clusters: map[string]*ClusterInfo{
				"cluster1": {HeartBeat: time.Now(), RWMutex: sync.RWMutex{}},
			},
			RWMutex: sync.RWMutex{},
		}
		convey.Convey("should return cluster for existing name", func() {
			cluster, err := clusterInfos.GetCluster("cluster1")
			convey.So(err, convey.ShouldBeNil)
			convey.So(cluster, convey.ShouldNotBeNil)
		})
	})
}

func TestClusterInfos_DeepCopy(t *testing.T) {
	convey.Convey("TestClusterInfos_DeepCopy", t, func() {
		clusterInfos := &ClusterInfos{
			Clusters: map[string]*ClusterInfo{
				"cluster1": {HeartBeat: time.Now(), RWMutex: sync.RWMutex{}},
				"cluster2": nil,
			},
			AllStatus: map[string]string{"cluster1": "running"},
			RWMutex:   sync.RWMutex{},
		}
		clone := clusterInfos.DeepCopy()
		convey.So(clone.AllStatus["cluster1"], convey.ShouldEqual, "running")
		convey.So(clone.Clusters["cluster2"], convey.ShouldBeNil)
	})
}

func TestClusterInfo_SetCommandVal(t *testing.T) {
	convey.Convey("TestClusterInfo_SetCommandVal", t, func() {
		clusterInfo := &ClusterInfo{
			Command: map[string]string{},
			RWMutex: sync.RWMutex{},
		}
		clusterInfo.SetCommandVal("key1", "val1")
		convey.So(clusterInfo.Command["key1"], convey.ShouldEqual, "val1")
	})
}

func TestClusterInfo_DeepCopy(t *testing.T) {
	convey.Convey("TestClusterInfo_DeepCopy", t, func() {
		convey.Convey("should deep copy with pos", func() {
			clusterInfo := &ClusterInfo{
				Command:   map[string]string{"cmd": "val"},
				FaultInfo: map[string]string{"f1": "v1"},
				Business:  []int32{1, 2, 3},
				HeartBeat: time.Now(),
				Pos:       &common.Position{Role: "cluster"},
				RWMutex:   sync.RWMutex{},
			}
			clone := clusterInfo.DeepCopy()
			convey.So(clone.Command["cmd"], convey.ShouldEqual, "val")
			convey.So(clone.Pos.Role, convey.ShouldEqual, "cluster")
		})
	})
}
