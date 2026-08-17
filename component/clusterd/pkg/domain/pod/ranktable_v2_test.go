// Copyright (c) Huawei Technologies Co., Ltd. 2026-2026. All rights reserved.

// Package pod tests for rank table v2.0 generation
package pod

import (
	"encoding/json"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"ascend-common/api"
	"clusterd/pkg/common/constant"
)

// fakeDeviceWithLevelList builds a device whose LevelList covers level0..level3.
// Info keys are UPPERCASE, matching device-plugin output.
func fakeDeviceWithLevelList(deviceID string) constant.Device {
	return constant.Device{
		DeviceID: deviceID,
		DeviceIP: "192.168.0.1",
		LevelList: []api.RankLevel{
			{Level: 0, Info: map[string]api.LevelElement{"UB": {NetLayer: 0, NetInstanceID: "L0"}}},
			{Level: 1, Info: map[string]api.LevelElement{"UB": {NetLayer: 1, NetInstanceID: "L1"}}},
			{Level: 2, Info: map[string]api.LevelElement{
				"UBOE": {NetLayer: 2, NetInstanceID: "L2-UBOE"},
				"UBG":  {NetLayer: 2, NetInstanceID: "L2-UBG"},
			}},
			{Level: 3, Info: map[string]api.LevelElement{"ROCE": {NetLayer: 3, NetInstanceID: "L3-ROCE"}}},
		},
	}
}

// fakePodWithDevice builds a pod carrying the given devices in the NPU device annotation.
func fakePodWithDevice(name, rankIndex string, devices []constant.Device) *v1.Pod {
	podDev := constant.PodDevice{PodName: name, Devices: devices}
	b, _ := json.Marshal(podDev)
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Annotations: map[string]string{
				api.PodNPUDeviceAnno: string(b),
				api.PodRankIndexAnno: rankIndex,
			},
		},
	}
}

func TestShouldInclude(t *testing.T) {
	convey.Convey("test shouldInclude", t, func() {
		convey.Convey("level2 is kept for UBOE and UBG, skipped for ROCE and UB", func() {
			convey.So(shouldInclude(level2, constant.PortAddrTypeUBoE, ""), convey.ShouldBeTrue)
			convey.So(shouldInclude(level2, constant.PortAddrTypeUBG, ""), convey.ShouldBeTrue)
			convey.So(shouldInclude(level2, constant.PortAddrTypeRoCE, ""), convey.ShouldBeFalse)
			convey.So(shouldInclude(level2, "UB", ""), convey.ShouldBeFalse)
		})
		convey.Convey("level3 is kept for ROCE, skipped for UBOE", func() {
			convey.So(shouldInclude(level3, constant.PortAddrTypeRoCE, ""), convey.ShouldBeTrue)
			convey.So(shouldInclude(level3, constant.PortAddrTypeUBoE, ""), convey.ShouldBeFalse)
		})
		convey.Convey("level0 and level1 are always kept", func() {
			convey.So(shouldInclude(level0, "", ""), convey.ShouldBeTrue)
			convey.So(shouldInclude(level0, constant.PortAddrTypeRoCE, ""), convey.ShouldBeTrue)
			convey.So(shouldInclude(level1, "", ""), convey.ShouldBeTrue)
			convey.So(shouldInclude(level1, constant.PortAddrTypeUBG, ""), convey.ShouldBeTrue)
		})
		convey.Convey("empty portAddrType falls back to customScaleOutType", func() {
			convey.So(shouldInclude(level2, "", constant.ScaleOutTypeUBoE), convey.ShouldBeTrue)
			convey.So(shouldInclude(level3, "", constant.ScaleOutTypeRoCE), convey.ShouldBeTrue)
			convey.So(shouldInclude(level3, "", constant.ScaleOutTypeUBoE), convey.ShouldBeFalse)
		})
	})
}

func TestGetElement(t *testing.T) {
	convey.Convey("test getElement", t, func() {
		dev := fakeDeviceWithLevelList("0")
		convey.Convey("level0 with empty portAddrType returns the first Info entry", func() {
			elem, ok := getElement(dev, level0, "")
			convey.So(ok, convey.ShouldBeTrue)
			convey.So(elem.NetInstanceID, convey.ShouldEqual, "L0")
		})
		convey.Convey("level2 with UBOE returns the UBOE element", func() {
			elem, ok := getElement(dev, level2, constant.PortAddrTypeUBoE)
			convey.So(ok, convey.ShouldBeTrue)
			convey.So(elem.NetInstanceID, convey.ShouldEqual, "L2-UBOE")
		})
		convey.Convey("level2 with ROCE returns not ok (no ROCE key at level2)", func() {
			_, ok := getElement(dev, level2, constant.PortAddrTypeRoCE)
			convey.So(ok, convey.ShouldBeFalse)
		})
		convey.Convey("level3 with ROCE returns the ROCE element", func() {
			elem, ok := getElement(dev, level3, constant.PortAddrTypeRoCE)
			convey.So(ok, convey.ShouldBeTrue)
			convey.So(elem.NetInstanceID, convey.ShouldEqual, "L3-ROCE")
		})
		convey.Convey("a level not present in the device returns not ok", func() {
			onlyL0 := constant.Device{
				DeviceID: "0",
				LevelList: []api.RankLevel{
					{Level: 0, Info: map[string]api.LevelElement{"UB": {NetLayer: 0, NetInstanceID: "L0"}}},
				},
			}
			_, ok := getElement(onlyL0, level3, constant.PortAddrTypeRoCE)
			convey.So(ok, convey.ShouldBeFalse)
		})
		convey.Convey("level0/1 with empty Info map returns (zero, false)", func() {
			emptyInfoDev := constant.Device{
				DeviceID: "0",
				LevelList: []api.RankLevel{
					{Level: 0, Info: map[string]api.LevelElement{}},
					{Level: 1, Info: map[string]api.LevelElement{}},
				},
			}
			elem0, ok0 := getElement(emptyInfoDev, level0, "")
			convey.So(ok0, convey.ShouldBeFalse)
			convey.So(elem0, convey.ShouldResemble, api.LevelElement{})
			elem1, ok1 := getElement(emptyInfoDev, level1, "")
			convey.So(ok1, convey.ShouldBeFalse)
			convey.So(elem1, convey.ShouldResemble, api.LevelElement{})
		})
	})
}

func TestGenRankList(t *testing.T) {
	convey.Convey("test genRankList", t, func() {
		convey.Convey("normal device builds rank with level0 and level1", func() {
			var rank constant.Rank
			err := genRankList(&rank, fakeDeviceWithLevelList("0"))
			convey.So(err, convey.ShouldBeNil)
			convey.So(rank.DeviceID, convey.ShouldEqual, 0)
			convey.So(rank.LocalID, convey.ShouldEqual, 0)
			convey.So(len(rank.LevelList), convey.ShouldEqual, 2)
			convey.So(rank.LevelList[0].NetLayer, convey.ShouldEqual, 0)
			convey.So(rank.LevelList[1].NetLayer, convey.ShouldEqual, 1)
		})
		convey.Convey("non-numeric device id returns error", func() {
			var rank constant.Rank
			err := genRankList(&rank, fakeDeviceWithLevelList("not-a-number"))
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("device missing level0/1 skips those levels without error", func() {
			onlyL3 := constant.Device{
				DeviceID: "0",
				LevelList: []api.RankLevel{
					{Level: 3, Info: map[string]api.LevelElement{"ROCE": {NetLayer: 3, NetInstanceID: "L3"}}},
				},
			}
			var rank constant.Rank
			err := genRankList(&rank, onlyL3)
			convey.So(err, convey.ShouldBeNil)
			convey.So(rank.DeviceID, convey.ShouldEqual, 0)
			convey.So(rank.LocalID, convey.ShouldEqual, 0)
			convey.So(len(rank.LevelList), convey.ShouldEqual, 0)
		})
	})
}

func TestConstructRankListV2(t *testing.T) {
	convey.Convey("test ConstructRankListV2", t, func() {
		convey.Convey("default network selection picks ROCE: level2 skipped, level3 kept", func() {
			rt := constant.RankTable{}
			pods := map[string]v1.Pod{
				"u1": *fakePodWithDevice("test-pod", "0", []constant.Device{fakeDeviceWithLevelList("0")}),
			}
			ConstructRankListV2(&rt, pods, 1, "")
			convey.So(rt.Version, convey.ShouldEqual, "2.0")
			convey.So(rt.RankCount, convey.ShouldEqual, 1)
			convey.So(len(rt.RankList), convey.ShouldEqual, 1)
			rank := rt.RankList[0]
			convey.So(rank.RankID, convey.ShouldEqual, 0)
			convey.So(rank.LocalID, convey.ShouldEqual, 0)
			convey.So(rank.DeviceID, convey.ShouldEqual, 0)
			convey.So(len(rank.LevelList), convey.ShouldEqual, 3)
			netLayers, instanceIDs := rankLevelLayersAndIDs(rank)
			convey.So(netLayers, convey.ShouldResemble, []int{0, 1, 3})
			convey.So(instanceIDs, convey.ShouldResemble, []string{"L0", "L1", "L3-ROCE"})
		})
		convey.Convey("custom UBOE keeps level2 and skips level3", func() {
			rt := constant.RankTable{}
			pods := map[string]v1.Pod{
				"u1": *fakePodWithDevice("test-pod", "0", []constant.Device{fakeDeviceWithLevelList("0")}),
			}
			ConstructRankListV2(&rt, pods, 1, constant.ScaleOutTypeUBoE)
			convey.So(rt.RankCount, convey.ShouldEqual, 1)
			convey.So(len(rt.RankList), convey.ShouldEqual, 1)
			rank := rt.RankList[0]
			convey.So(len(rank.LevelList), convey.ShouldEqual, 3)
			netLayers, instanceIDs := rankLevelLayersAndIDs(rank)
			convey.So(netLayers, convey.ShouldResemble, []int{0, 1, 2})
			convey.So(instanceIDs, convey.ShouldResemble, []string{"L0", "L1", "L2-UBOE"})
		})
		convey.Convey("two pods produce two ranks sorted ascending by RankID", func() {
			rt := constant.RankTable{}
			pods := map[string]v1.Pod{
				"u1": *fakePodWithDevice("pod-1", "1", []constant.Device{fakeDeviceWithLevelList("1")}),
				"u0": *fakePodWithDevice("pod-0", "0", []constant.Device{fakeDeviceWithLevelList("0")}),
			}
			ConstructRankListV2(&rt, pods, 2, "")
			convey.So(rt.RankCount, convey.ShouldEqual, 2)
			convey.So(len(rt.RankList), convey.ShouldEqual, 2)
			convey.So(rt.RankList[0].RankID, convey.ShouldEqual, 0)
			convey.So(rt.RankList[1].RankID, convey.ShouldEqual, 1)
			convey.So(rt.RankList[0].DeviceID, convey.ShouldEqual, 0)
			convey.So(rt.RankList[1].DeviceID, convey.ShouldEqual, 1)
		})
		convey.Convey("empty pods leave the rank table untouched", func() {
			rt := constant.RankTable{}
			ConstructRankListV2(&rt, map[string]v1.Pod{}, 1, "")
			convey.So(len(rt.RankList), convey.ShouldEqual, 0)
			convey.So(rt.RankCount, convey.ShouldEqual, 0)
			convey.So(rt.Version, convey.ShouldEqual, "")
		})
		convey.Convey("pod with rankIndex >= replicas or missing rank index is skipped", func() {
			rt := constant.RankTable{}
			noRankPod := fakePodWithDevice("pod-no-rank", "0", []constant.Device{fakeDeviceWithLevelList("0")})
			delete(noRankPod.Annotations, api.PodRankIndexAnno)
			pods := map[string]v1.Pod{
				"u0":   *fakePodWithDevice("pod-ok", "0", []constant.Device{fakeDeviceWithLevelList("0")}),
				"u5":   *fakePodWithDevice("pod-too-big", "5", []constant.Device{fakeDeviceWithLevelList("5")}),
				"u-no": *noRankPod,
			}
			ConstructRankListV2(&rt, pods, 2, "")
			convey.So(rt.RankCount, convey.ShouldEqual, 1)
			convey.So(len(rt.RankList), convey.ShouldEqual, 1)
			convey.So(rt.RankList[0].DeviceID, convey.ShouldEqual, 0)
		})
		convey.Convey("pod with empty devices is skipped", func() {
			rt := constant.RankTable{}
			pods := map[string]v1.Pod{
				"u0": *fakePodWithDevice("pod-empty", "0", []constant.Device{}),
			}
			ConstructRankListV2(&rt, pods, 1, "")
			convey.So(rt.RankCount, convey.ShouldEqual, 0)
			convey.So(len(rt.RankList), convey.ShouldEqual, 0)
		})
		convey.Convey("non-numeric device id is skipped but valid devices still produce ranks", func() {
			rt := constant.RankTable{}
			pods := map[string]v1.Pod{
				"u0": *fakePodWithDevice("pod-mixed", "0",
					[]constant.Device{fakeDeviceWithLevelList("0"), fakeDeviceWithLevelList("not-a-number")}),
			}
			ConstructRankListV2(&rt, pods, 1, "")
			convey.So(rt.RankCount, convey.ShouldEqual, 1)
			convey.So(len(rt.RankList), convey.ShouldEqual, 1)
			convey.So(rt.RankList[0].DeviceID, convey.ShouldEqual, 0)
			convey.So(rt.RankList[0].RankID, convey.ShouldEqual, 0)
		})
		convey.Convey("invalid customScaleOutType still generates level0/1 only ranks", func() {
			rt := constant.RankTable{}
			pods := map[string]v1.Pod{
				"u0": *fakePodWithDevice("pod-invalid-custom", "0",
					[]constant.Device{fakeDeviceWithLevelList("0")}),
			}
			ConstructRankListV2(&rt, pods, 1, "XXX")
			convey.So(rt.RankCount, convey.ShouldEqual, 1)
			convey.So(len(rt.RankList), convey.ShouldEqual, 1)
			rank := rt.RankList[0]
			convey.So(len(rank.LevelList), convey.ShouldEqual, 2)
			netLayers, _ := rankLevelLayersAndIDs(rank)
			convey.So(netLayers, convey.ShouldResemble, []int{0, 1})
		})
		convey.Convey("missing level3 element gets default LevelElement appended", func() {
			rt := constant.RankTable{}
			devB := constant.Device{
				DeviceID: "1",
				DeviceIP: "192.168.0.2",
				LevelList: []api.RankLevel{
					{Level: 0, Info: map[string]api.LevelElement{"UB": {NetLayer: 0, NetInstanceID: "L0"}}},
					{Level: 1, Info: map[string]api.LevelElement{"UB": {NetLayer: 1, NetInstanceID: "L1"}}},
					{Level: 3, Info: map[string]api.LevelElement{"UB": {NetLayer: 3, NetInstanceID: "L3-UB"}}},
				},
			}
			pods := map[string]v1.Pod{
				"u0": *fakePodWithDevice("pod-a", "0", []constant.Device{fakeDeviceWithLevelList("0")}),
				"u1": *fakePodWithDevice("pod-b", "1", []constant.Device{devB}),
			}
			ConstructRankListV2(&rt, pods, 2, "")
			convey.So(rt.RankCount, convey.ShouldEqual, 2)
			convey.So(len(rt.RankList), convey.ShouldEqual, 2)
			var rankB *constant.Rank
			for i := range rt.RankList {
				if rt.RankList[i].DeviceID == 1 {
					rankB = &rt.RankList[i]
				}
			}
			convey.So(rankB, convey.ShouldNotBeNil)
			netLayers, instanceIDs := rankLevelLayersAndIDs(*rankB)
			convey.So(netLayers, convey.ShouldResemble, []int{0, 1, 3})
			convey.So(instanceIDs, convey.ShouldResemble, []string{"L0", "L1", api.DefaultClusterName})
			convey.So(rankB.LevelList[2].NetInstanceID, convey.ShouldEqual, "CLUSTER1")
			convey.So(rankB.LevelList[2].NetType, convey.ShouldEqual, api.NetTypeCLOS)
		})
	})
}

// rankLevelLayersAndIDs extracts net layers and instance ids from a rank's LevelList.
func rankLevelLayersAndIDs(rank constant.Rank) ([]int, []string) {
	netLayers := make([]int, 0, len(rank.LevelList))
	instanceIDs := make([]string, 0, len(rank.LevelList))
	for _, elem := range rank.LevelList {
		netLayers = append(netLayers, elem.NetLayer)
		instanceIDs = append(instanceIDs, elem.NetInstanceID)
	}
	return netLayers, instanceIDs
}
