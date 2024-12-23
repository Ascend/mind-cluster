/* Copyright(C) 2024. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package common for common function
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
	mockFaultCode  = "0000001D"
)

func TestDeepEqualFaultDevInfo01(t *testing.T) {
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
	})
}

func TestDeepEqualFaultDevInfo02(t *testing.T) {
	convey.Convey("test DeepEqualFaultDevInfo", t, func() {
		faultDevInfo := &FaultDevInfo{
			NodeStatus: nodeUnHealthy,
			FaultDevList: []*FaultDev{
				{
					DeviceType: mockDeviceType,
					DeviceId:   mockDeviceID,
				},
			},
		}
		convey.Convey("two FaultDevInfo with different fault level should not be deep equal", func() {
			faultDevInfo1 := *faultDevInfo
			faultDevInfo1.FaultDevList[0].FaultLevel = NotHandleFault
			faultDevInfo2 := *faultDevInfo
			faultDevInfo2.FaultDevList[0].FaultLevel = PreSeparateFault
			convey.So(DeepEqualFaultDevInfo(&faultDevInfo1, &faultDevInfo2), convey.ShouldEqual, false)
		})
		convey.Convey("two FaultDevInfo with different fault code should not be deep equal", func() {
			faultDevInfo1 := *faultDevInfo
			faultDevInfo1.FaultDevList[0].FaultLevel = PreSeparateFault
			faultDevInfo1.FaultDevList[0].FaultCode = []string{mockFaultCode}
			faultDevInfo2 := *faultDevInfo
			faultDevInfo2.FaultDevList[0].FaultLevel = PreSeparateFault
			faultDevInfo2.FaultDevList[0].FaultCode = []string{"2800001F"}
			convey.So(DeepEqualFaultDevInfo(&faultDevInfo1, &faultDevInfo2), convey.ShouldEqual, false)
		})
		convey.Convey("two FaultDevInfo with same attribute should be deep equal", func() {
			faultDevInfo1 := *faultDevInfo
			faultDevInfo1.FaultDevList[0].FaultLevel = PreSeparateFault
			faultDevInfo1.FaultDevList[0].FaultCode = []string{mockFaultCode}
			faultDevInfo2 := faultDevInfo1
			res := DeepEqualFaultDevInfo(&faultDevInfo1, &faultDevInfo2)
			convey.So(res, convey.ShouldEqual, true)
		})
	})
}
