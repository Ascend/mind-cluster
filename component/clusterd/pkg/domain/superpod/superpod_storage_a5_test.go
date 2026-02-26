/*
Copyright(C)2025. Huawei Technologies Co.,Ltd. All rights reserved.

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

/*
Package superpod a series of cluster device info storage function.
*/
package superpod

import (
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"ascend-common/api"
)

func TestDeepCopyNpuInfo(t *testing.T) {
	convey.Convey("Test deepCopyNpuInfo", t, func() {
		convey.Convey("When input is nil", func() {
			result := deepCopyNpuInfo(nil)
			convey.So(result, convey.ShouldBeNil)
		})

		convey.Convey("When input has basic fields", func() {
			original := &api.NpuInfo{
				PhyId:     "1",
				LevelList: []api.LevelElement{},
			}

			copy := deepCopyNpuInfo(original)
			convey.So(copy, convey.ShouldNotBeNil)
			convey.So(copy.PhyId, convey.ShouldEqual, original.PhyId)
			convey.So(copy.LevelList, convey.ShouldResemble, original.LevelList)
		})
	})
}

func TestDeepCopyRackInfo(t *testing.T) {
	convey.Convey("Test deepCopyRackInfo", t, func() {
		convey.Convey("When input is nil", func() {
			result := deepCopyRackInfo(nil)
			convey.So(result, convey.ShouldBeNil)
		})
		convey.Convey("When input has basic fields with empty ServerMap", func() {
			original := &api.RackInfo{
				RackID:    "rack-1",
				ServerMap: make(map[string]*api.ServerInfo),
			}
			copy := deepCopyRackInfo(original)
			convey.So(copy, convey.ShouldNotBeNil)
			convey.So(copy.RackID, convey.ShouldEqual, original.RackID)
			convey.So(copy.ServerMap, convey.ShouldNotBeNil)
			convey.So(copy.ServerMap, convey.ShouldBeEmpty)
			original.RackID = "modified-rack"
			convey.So(copy.RackID, convey.ShouldEqual, "rack-1")
		})
	})
}

func TestGetFinalDelSuperPodID(t *testing.T) {
	t.Run("Empty map", func(t *testing.T) {
		superPodDeleteFinalManager = Manager{
			snMap: make(map[string]*api.SuperPodDevice, initSuperPodNum),
		}
		result := GetFinalDelSuperPodID()
		if len(result) != 0 {
			t.Errorf("Expected empty map, got %v", result)
		}
	})
}

func TestSaveRackInfo(t *testing.T) {
	convey.Convey("Test saveRackInfo", t, func() {
		convey.Convey("When superPod is not A5 version", func() {
			superPod := &api.SuperPodDevice{Version: "non-A5"}
			node := &api.NodeDevice{}
			err := saveRackInfo(superPod, node)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("When adding new rack to empty superPod", func() {
			superPod := &api.SuperPodDevice{
				Version:    api.VersionNPU,
				RackMap:    make(map[string]*api.RackInfo),
				SuperPodID: "pod-1",
			}
			node := &api.NodeDevice{
				RackID:   "rack-1",
				NodeName: "node-1",
				ServerID: "server-1",
				NpuInfoMap: map[string]*api.NpuInfo{
					"npu-1": {PhyId: "1"},
				},
			}
			err := saveRackInfo(superPod, node)
			convey.So(err, convey.ShouldBeNil)
			convey.So(superPod.RackMap, convey.ShouldContainKey, "rack-1")
			convey.So(superPod.RackMap["rack-1"].ServerMap, convey.ShouldContainKey, "server-1")
			convey.So(superPod.RackMap["rack-1"].ServerMap["server-1"].NpuMap, convey.ShouldContainKey, "npu-1")
		})
	})
	convey.Convey("When adding node to existing rack", t, testGetRackInfo2)
	convey.Convey("When exceeding max rack number", t, testGetRackInfo1)
}

func testGetRackInfo1() {
	superPod := &api.SuperPodDevice{
		Version:    "npu",
		RackMap:    make(map[string]*api.RackInfo),
		SuperPodID: "pod-1",
	}
	for i := 0; i < maxRackNumPerSuperPod; i++ {
		superPod.RackMap[fmt.Sprintf("rack-%d", i)] = &api.RackInfo{}
	}
	node := &api.NodeDevice{
		RackID:   "new-rack",
		NodeName: "node-1",
	}
	err := saveRackInfo(superPod, node)
	convey.So(err, convey.ShouldNotBeNil)
	convey.So(err.Error(), convey.ShouldContainSubstring, "rackMap length exceeds the limit")
}

func testGetRackInfo2() {
	superPod := &api.SuperPodDevice{
		Version: api.VersionNPU,
		RackMap: map[string]*api.RackInfo{
			"rack-1": {
				RackID:    "rack-1",
				ServerMap: make(map[string]*api.ServerInfo),
			},
		},
		SuperPodID: "pod-1",
	}
	node := &api.NodeDevice{
		RackID:   "rack-1",
		NodeName: "node-1",
		ServerID: "server-1",
		NpuInfoMap: map[string]*api.NpuInfo{
			"npu-1": {PhyId: "1"},
		},
	}
	err := saveRackInfo(superPod, node)
	convey.So(err, convey.ShouldBeNil)
	convey.So(superPod.RackMap["rack-1"].ServerMap, convey.ShouldContainKey, "server-1")
}

func TestCanAddNodeToSuperPod(t *testing.T) {
	convey.Convey("Test canAddNodeToSuperPod", t, func() {
		convey.Convey("When versions don't match", func() {
			superPod := &api.SuperPodDevice{
				Version:    api.VersionNPU,
				SuperPodID: "pod-1",
			}
			node := &api.NodeDevice{
				ServerType: "A6",
				NodeName:   "node-1",
			}
			result := canAddNodeToSuperPod(superPod, node)
			convey.So(result, convey.ShouldBeFalse)
		})
		convey.Convey("When A5 version and exceeds max node number", func() {
			superPod := &api.SuperPodDevice{
				Version:       api.VersionNPU,
				SuperPodID:    "pod-1",
				NodeDeviceMap: make(map[string]*api.NodeDevice),
			}
			for i := 0; i <= maxNodeNumPerSuperPodA5; i++ {
				superPod.NodeDeviceMap[fmt.Sprintf("node-%d", i)] = &api.NodeDevice{}
			}
			node := &api.NodeDevice{
				NodeName: "new-node",
			}
			result := canAddNodeToSuperPod(superPod, node)
			convey.So(result, convey.ShouldBeFalse)
		})
	})
	convey.Convey("Test CanAddNodeToSuperPod", t, testCanAddNodeToSuperPod1)
	convey.Convey("Test CanAddNodeToSuperPod2", t, testCanAddNodeToSuperPod2)
}

func testCanAddNodeToSuperPod1() {
	superPod := &api.SuperPodDevice{
		Version:       api.VersionNPU,
		NodeDeviceMap: make(map[string]*api.NodeDevice),
	}
	node := &api.NodeDevice{
		ServerType: api.VersionNPU,
		NodeName:   "node-1",
	}
	result := canAddNodeToSuperPod(superPod, node)
	convey.So(result, convey.ShouldBeTrue)
}

func testCanAddNodeToSuperPod2() {
	superPod := &api.SuperPodDevice{
		Version:       "A6",
		SuperPodID:    "pod-1",
		NodeDeviceMap: make(map[string]*api.NodeDevice),
	}
	for i := 0; i <= maxNodeNumPerSuperPod; i++ {
		superPod.NodeDeviceMap[fmt.Sprintf("node-%d", i)] = &api.NodeDevice{}
	}
	node := &api.NodeDevice{
		NodeName: "new-node",
	}
	result := canAddNodeToSuperPod(superPod, node)
	convey.So(result, convey.ShouldBeFalse)
}

func TestDeleteNodeInRackMap(t *testing.T) {
	convey.Convey("TestDeleteNodeInRackMap case 1", t, func() {
		superPodManager.snMap["0"] = nil
		defer func() { superPodManager.snMap = make(map[string]*api.SuperPodDevice, initSuperPodNum) }()
		DeleteNodeInRackMap("1", nil)
		convey.So(len(superPodManager.snMap), convey.ShouldEqual, 1)
	})

	convey.Convey("TestDeleteNodeInRackMap case 2", t, func() {
		superPodManager.snMap["0"] = nil
		defer func() { superPodManager.snMap = make(map[string]*api.SuperPodDevice, initSuperPodNum) }()
		DeleteNodeInRackMap("0", nil)
		convey.So(len(superPodManager.snMap), convey.ShouldEqual, 1)
	})

	convey.Convey("TestDeleteNodeInRackMap case 3", t, func() {
		superPod := &api.SuperPodDevice{RackMap: make(map[string]*api.RackInfo)}
		superPodManager.snMap = map[string]*api.SuperPodDevice{"0": superPod}
		defer func() { superPodManager.snMap = make(map[string]*api.SuperPodDevice, initSuperPodNum) }()
		nodeDevice := &api.NodeDevice{RackID: "1", ServerType: "1"}
		DeleteNodeInRackMap("0", nodeDevice)
		convey.So(len(superPodManager.snMap), convey.ShouldEqual, 1)
	})

	convey.Convey("TestDeleteNodeInRackMap case 4", t, func() {
		rackInfo := &api.RackInfo{
			ServerMap: map[string]*api.ServerInfo{"1": nil},
		}
		rackMap := map[string]*api.RackInfo{"1": rackInfo}
		superPod := &api.SuperPodDevice{RackMap: rackMap}
		superPodManager.snMap = map[string]*api.SuperPodDevice{"0": superPod}
		defer func() { superPodManager.snMap = make(map[string]*api.SuperPodDevice, initSuperPodNum) }()
		nodeDevice := &api.NodeDevice{RackID: "1", ServerType: "1"}
		DeleteNodeInRackMap("0", nodeDevice)
		convey.So(len(superPodManager.snMap), convey.ShouldEqual, 1)
	})
}
