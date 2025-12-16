// Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.

// Package pod a series of pod util function
package pod

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"volcano.sh/apis/pkg/apis/scheduling/v1beta1"

	"ascend-common/api"
	"clusterd/pkg/common/constant"
)

func makeAddr(addrType, addr string) api.RankAddrItem {
	return api.RankAddrItem{AddrType: addrType, Addr: addr}
}

func TestAdjustScaleOutTypeForStacking(t *testing.T) {
	convey.Convey("Test adjustScaleOutTypeForStacking", t, func() {
		convey.Convey("worker pod should keep UBOE", func() {
			pods := map[string]v1.Pod{
				"p1": {
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{"replica-type": "worker"},
					},
				},
			}
			got := adjustScaleOutTypeForStacking("UBOE", pods)
			convey.So(got, convey.ShouldEqual, "UBOE")
		})

		convey.Convey("master pod + stacking should force ROCE", func() {
			pods := map[string]v1.Pod{
				"p2": {
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{"replica-type": "master"},
					},
					Spec: v1.PodSpec{
						NodeSelector: map[string]string{api.AcceleratorTypeKey: api.Ascend800ia5Stacking},
					},
				},
			}
			got := adjustScaleOutTypeForStacking("UBOE", pods)
			convey.So(got, convey.ShouldEqual, "ROCE")
		})
	})
}

func TestGetScaleOutType(t *testing.T) {
	convey.Convey("Test getScaleOutType", t, func() {
		pg := v1beta1.PodGroup{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{},
			},
		}
		pods := map[string]v1.Pod{}

		convey.Convey("label not set should return adjusted default", func() {
			got, err := getScaleOutType(pg, pods)
			convey.So(err, convey.ShouldBeNil)
			convey.So(got, convey.ShouldNotBeNil)
		})

		convey.Convey("label=roce should normalize to ROCE", func() {
			pg.ObjectMeta.Labels[api.ScaleOutType] = "roce"
			got, err := getScaleOutType(pg, pods)
			convey.So(err, convey.ShouldBeNil)
			convey.So(got, convey.ShouldEqual, "ROCE")
		})

		convey.Convey("label=invalid should return error", func() {
			pg.ObjectMeta.Labels[api.ScaleOutType] = "invalid"
			got, err := getScaleOutType(pg, pods)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(got, convey.ShouldEqual, "")
		})
	})
}

func TestSetScaleOutNetwork(t *testing.T) {
	convey.Convey("Test setScaleOutNetwork", t, func() {
		dev := constant.Device{DeviceID: "d1"}
		serverDev := &constant.Device{}

		convey.Convey("empty LevelList should skip", func() {
			setScaleOutNetwork(dev, "ROCE", serverDev)
			convey.So(serverDev.ScaleOutNetwork, convey.ShouldBeNil)
		})

		convey.Convey("non-empty LevelList should set ScaleOutNetwork", func() {
			dev.LevelList = []api.RankLevel{
				{
					Level: api.Level3,
					Info: map[string]api.LevelElement{
						"ROCE": {
							RankAddrList: []api.RankAddrItem{
								makeAddr("IPV4", "192.168.1.1"),
							},
						},
					},
				},
			}
			setScaleOutNetwork(dev, "ROCE", serverDev)
			convey.So(serverDev.ScaleOutNetwork, convey.ShouldNotBeNil)
			convey.So(serverDev.ScaleOutNetwork.Addrs, convey.ShouldContain, "192.168.1.1")
		})
	})
}

func TestCollectValidAddrs(t *testing.T) {
	convey.Convey("Test collectValidAddrs", t, func() {
		addrs := []api.RankAddrItem{
			{AddrType: "IPV4", Addr: ""},
			{AddrType: "IPV4", Addr: "192.168.2.2"},
		}
		result := collectValidAddrs(api.Level2, "UBOE", addrs)
		convey.So(len(result), convey.ShouldEqual, 1)
		convey.So(result[0].Addr, convey.ShouldEqual, "192.168.2.2")
	})
}

func TestIsValidNetType(t *testing.T) {
	const InvalidLevel = 99
	convey.Convey("Test isValidNetType", t, func() {
		convey.So(isValidNetType(api.Level2, "UBOE"), convey.ShouldBeTrue)
		convey.So(isValidNetType(api.Level2, "ROCE"), convey.ShouldBeFalse)
		convey.So(isValidNetType(api.Level3, "ROCE"), convey.ShouldBeTrue)
		convey.So(isValidNetType(InvalidLevel, "ROCE"), convey.ShouldBeFalse)
	})
}

func TestSelectScaleOutNetwork(t *testing.T) {
	convey.Convey("Test selectScaleOutNetwork", t, func() {
		serverDev := &constant.Device{}
		portMap := map[string][]api.RankAddrItem{
			"ROCE": {makeAddr("IPV4", "192.168.3.3")},
			"UBOE": {makeAddr("IPV4", "192.168.4.4")},
			"UBG":  {makeAddr("IPV4", "192.168.5.5")},
		}

		convey.Convey("empty scaleOutType should pick default", func() {
			selectScaleOutNetwork(portMap, "", serverDev)
			convey.So(serverDev.ScaleOutNetwork, convey.ShouldNotBeNil)
		})

		convey.Convey("ROCE should set ROCE network", func() {
			selectScaleOutNetwork(portMap, "ROCE", serverDev)
			convey.So(serverDev.ScaleOutNetwork.Addrs, convey.ShouldContain, "192.168.3.3")
		})

		convey.Convey("UBOE should set UBOE network", func() {
			selectScaleOutNetwork(portMap, "UBOE", serverDev)
			convey.So(serverDev.ScaleOutNetwork.Addrs, convey.ShouldContain, "192.168.4.4")
		})

		convey.Convey("invalid type should not set", func() {
			selectScaleOutNetwork(portMap, "INVALID", serverDev)
			convey.So(serverDev.ScaleOutNetwork, convey.ShouldBeNil)
		})
	})
}

func TestHandleScaleOutNetworkInfo(t *testing.T) {
	convey.Convey("Test handleScaleOutNetworkInfo", t, func() {
		serverDev := &constant.Device{}

		convey.Convey("empty ports should skip", func() {
			handleScaleOutNetworkInfo(serverDev, []api.RankAddrItem{})
			convey.So(serverDev.ScaleOutNetwork, convey.ShouldBeNil)
		})

		convey.Convey("IPV4 ports should set AddrType IPV4", func() {
			handleScaleOutNetworkInfo(serverDev, []api.RankAddrItem{makeAddr("IPV4", "192.168.6.6")})
			convey.So(serverDev.ScaleOutNetwork, convey.ShouldNotBeNil)
			convey.So(serverDev.ScaleOutNetwork.AddrType, convey.ShouldEqual, api.AddrTypeIPV4)
			convey.So(serverDev.ScaleOutNetwork.Addrs, convey.ShouldContain, "192.168.6.6")
		})

		convey.Convey("EID ports should set AddrType EID", func() {
			handleScaleOutNetworkInfo(serverDev, []api.RankAddrItem{makeAddr("EID", "eid-addr")})
			convey.So(serverDev.ScaleOutNetwork, convey.ShouldNotBeNil)
			convey.So(serverDev.ScaleOutNetwork.AddrType, convey.ShouldEqual, api.AddrTypeEID)
			convey.So(serverDev.ScaleOutNetwork.Addrs, convey.ShouldContain, "eid-addr")
		})
	})
}
