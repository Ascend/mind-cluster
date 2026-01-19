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
			faultDevInfo1 := mockFaultDevInfo(nodeUnHealthy, nil)
			faultDevInfo2 := mockFaultDevInfo(nodeHealthy, nil)
			res := DeepEqualFaultDevInfo(faultDevInfo1, faultDevInfo2)
			convey.So(res, convey.ShouldEqual, false)
		})

		convey.Convey("two FaultDevInfo with different device type should not be deep equal", func() {
			faultDevInfo1 := mockFaultDevInfo(nodeUnHealthy, []*FaultDev{mockFaultDev(mockDeviceType, "",
				mockDeviceID, nil)})
			faultDevInfo2 := mockFaultDevInfo(nodeUnHealthy, []*FaultDev{mockFaultDev("memory", "",
				mockDeviceID, nil)})
			res := DeepEqualFaultDevInfo(faultDevInfo1, faultDevInfo2)
			convey.So(res, convey.ShouldEqual, false)
		})
	})
}

func TestDeepEqualFaultDevInfo02(t *testing.T) {
	convey.Convey("test DeepEqualFaultDevInfo", t, func() {
		convey.Convey("two FaultDevInfo with different fault level should not be deep equal", func() {
			faultDevInfo1 := mockFaultDevInfo(nodeUnHealthy, []*FaultDev{mockFaultDev(mockDeviceType, NotHandleFault,
				mockDeviceID, nil)})
			faultDevInfo2 := mockFaultDevInfo(nodeUnHealthy, []*FaultDev{mockFaultDev(mockDeviceType, PreSeparateFault,
				mockDeviceID, nil)})
			convey.So(DeepEqualFaultDevInfo(faultDevInfo1, faultDevInfo2), convey.ShouldEqual, false)
		})
		convey.Convey("two FaultDevInfo with different fault code should not be deep equal", func() {
			faultDevInfo1 := mockFaultDevInfo(nodeUnHealthy, []*FaultDev{mockFaultDev(mockDeviceType, NotHandleFault,
				mockDeviceID, []string{mockFaultCode})})
			faultDevInfo2 := mockFaultDevInfo(nodeUnHealthy, []*FaultDev{mockFaultDev(mockDeviceType, NotHandleFault,
				mockDeviceID, []string{"2800001F"})})
			convey.So(DeepEqualFaultDevInfo(faultDevInfo1, faultDevInfo2), convey.ShouldEqual, false)
		})
		convey.Convey("two FaultDevInfo with same attribute should be deep equal", func() {
			faultDevInfo1 := mockFaultDevInfo(nodeUnHealthy, []*FaultDev{mockFaultDev(mockDeviceType, NotHandleFault,
				mockDeviceID, []string{mockFaultCode})})
			faultDevInfo2 := faultDevInfo1
			res := DeepEqualFaultDevInfo(faultDevInfo1, faultDevInfo2)
			convey.So(res, convey.ShouldEqual, true)
		})
	})
}

func mockFaultDevInfo(status string, devs []*FaultDev) *FaultDevInfo {
	return &FaultDevInfo{
		FaultDevList: devs,
		NodeStatus:   status,
	}
}

func mockFaultDev(tp, level string, id int64, faultCode []string) *FaultDev {
	return &FaultDev{
		DeviceType: tp,
		DeviceId:   id,
		FaultLevel: level,
		FaultCode:  faultCode,
	}
}

func TestGenerateFaultID(t *testing.T) {
	convey.Convey("test GenerateFaultID", t, func() {
		convey.Convey("01-when node name and id is different, should return different id", func() {
			name1 := "test1"
			id1 := "1"
			name2 := "test2"
			id2 := "2"
			convey.So(GenerateFaultID(name1, id1), convey.ShouldNotEqual, GenerateFaultID(name2, id2))
		})
	})
}

func TestFaultDevListEqual(t *testing.T) {
	convey.Convey("test faultDevListEqual", t, func() {
		convey.Convey("case length is not equal", func() {
			src := []*FaultDev{{DeviceId: 0}}
			res := faultDevListEqual(src, nil)
			convey.So(res, convey.ShouldBeFalse)
		})
		convey.Convey("case map length is not equal", func() {
			src := []*FaultDev{{DeviceId: 0}, {DeviceId: 0}}
			dst := []*FaultDev{{DeviceId: 0}, {DeviceId: 1}}
			res := faultDevListEqual(src, dst)
			convey.So(res, convey.ShouldBeFalse)
		})
	})
}
