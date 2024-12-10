package common

import (
	"github.com/smartystreets/goconvey/convey"
	"testing"
)

// node fault name str
const (
	nodeHealthy    = "Healthy"
	nodeUnHealthy  = "UnHealthy"
	mockDeviceType = "cpu"
	mockDeviceID   = 0
)

func TestDeepEqualFaultDevInfo(t *testing.T) {
	convey.Convey("test DeepEqualFaultDevInfo", t, func() {
		convey.Convey("two nil FaultDevInfo should be deep equal", func() {
			res := DeepEqualFaultDevInfo(nil, nil)
			convey.So(res, convey.ShouldEqual, true)
		})

		convey.Convey("nit FaultDevInfo should not equal to FaultDevInfo which is not nil", func() {
			res := DeepEqualFaultDevInfo(nil, &FaultDevInfo{})
			convey.So(res, convey.ShouldEqual, false)
		})

		convey.Convey("two FaultDevInfo with different node status should not be deep equal", func() {
			faultDevInfo1 := &FaultDevInfo{
				NodeStatus: nodeUnHealthy,
			}
			faultDevInfo2 := &FaultDevInfo{
				NodeStatus: nodeHealthy,
			}
			res := DeepEqualFaultDevInfo(faultDevInfo1, faultDevInfo2)
			convey.So(res, convey.ShouldEqual, false)
		})

		convey.Convey("two FaultDevInfo with different device type should not be deep equal", func() {
			faultDevInfo1 := &FaultDevInfo{
				NodeStatus: nodeUnHealthy,
				FaultDevList: []*FaultDev{
					{
						DeviceType: mockDeviceType,
						DeviceId:   mockDeviceID,
					},
				},
			}
			faultDevInfo2 := &FaultDevInfo{
				NodeStatus: nodeUnHealthy,
				FaultDevList: []*FaultDev{
					{
						DeviceType: "memory",
						DeviceId:   mockDeviceID,
					},
				},
			}
			res := DeepEqualFaultDevInfo(faultDevInfo1, faultDevInfo2)
			convey.So(res, convey.ShouldEqual, false)
		})

		convey.Convey("two FaultDevInfo with different fault level should not be deep equal", func() {
			faultDevInfo1 := &FaultDevInfo{
				NodeStatus: nodeUnHealthy,
				FaultDevList: []*FaultDev{
					{
						DeviceType: mockDeviceType,
						DeviceId:   mockDeviceID,
						FaultLevel: NotHandleFault,
					},
				},
			}
			faultDevInfo2 := &FaultDevInfo{
				NodeStatus: nodeUnHealthy,
				FaultDevList: []*FaultDev{
					{
						DeviceType: mockDeviceType,
						DeviceId:   mockDeviceID,
						FaultLevel: PreSeparateFault,
					},
				},
			}
			res := DeepEqualFaultDevInfo(faultDevInfo1, faultDevInfo2)
			convey.So(res, convey.ShouldEqual, false)
		})
		convey.Convey("two FaultDevInfo with different fault level should not be deep equal", func() {
			faultDevInfo1 := &FaultDevInfo{
				NodeStatus: nodeUnHealthy,
				FaultDevList: []*FaultDev{
					{
						DeviceType: mockDeviceType,
						DeviceId:   mockDeviceID,
						FaultLevel: PreSeparateFault,
						FaultCode:  []string{"0000001D"},
					},
				},
			}
			faultDevInfo2 := &FaultDevInfo{
				NodeStatus: nodeUnHealthy,
				FaultDevList: []*FaultDev{
					{
						DeviceType: mockDeviceType,
						DeviceId:   mockDeviceID,
						FaultLevel: PreSeparateFault,
						FaultCode:  []string{"2800001F"},
					},
				},
			}
			res := DeepEqualFaultDevInfo(faultDevInfo1, faultDevInfo2)
			convey.So(res, convey.ShouldEqual, false)
		})

		convey.Convey("two FaultDevInfo with same attribute should be deep equal", func() {
			faultDevInfo1 := &FaultDevInfo{
				NodeStatus: nodeUnHealthy,
				FaultDevList: []*FaultDev{
					{
						DeviceType: mockDeviceType,
						DeviceId:   0,
						FaultLevel: PreSeparateFault,
						FaultCode:  []string{"0000001D"},
					},
				},
			}
			faultDevInfo2 := &FaultDevInfo{
				NodeStatus: nodeUnHealthy,
				FaultDevList: []*FaultDev{
					{
						DeviceType: mockDeviceType,
						DeviceId:   0,
						FaultLevel: PreSeparateFault,
						FaultCode:  []string{"0000001D"},
					},
				},
			}
			res := DeepEqualFaultDevInfo(faultDevInfo1, faultDevInfo2)
			convey.So(res, convey.ShouldEqual, true)
		})
	})
}
