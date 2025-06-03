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
package manager

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/hwlog"
	"nodeD/pkg/common"
)

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwlog.InitRunLogger(&hwLogConfig, context.Background())
}

func mockWrongFaultManager() FaultManager {
	return &FaultTools{}
}

// TestSetFaultDevInfo test the function of set fault dev info
func TestSetFaultDevInfo(t *testing.T) {
	convey.Convey("test set fault dev info", t, func() {
		convey.Convey("fault manager set fault dev info", func() {
			faultManager := NewFaultManager()
			faultDevInfo := &common.FaultDevInfo{
				FaultDevList: []*common.FaultDev{&common.FaultDev{
					DeviceType: "CPU",
					DeviceId:   0,
					FaultCode:  []string{"01010001"},
					FaultLevel: common.NotHandleFault,
				}},
				NodeStatus: common.NodeHealthy,
			}
			faultManager.SetFaultDevInfo(faultDevInfo)
			FaultDevInfoEqual(faultManager.GetFaultDevInfo(), faultDevInfo)
		})
		convey.Convey("wrong fault manager set fault dev info", func() {
			faultManager := mockWrongFaultManager()
			faultDevInfo := &common.FaultDevInfo{
				FaultDevList: []*common.FaultDev{&common.FaultDev{
					DeviceType: "CPU",
					DeviceId:   0,
					FaultCode:  []string{"01010001"},
					FaultLevel: common.NotHandleFault,
				}},
				NodeStatus: common.NodeHealthy,
			}
			faultManager.SetFaultDevInfo(faultDevInfo)
			convey.So(faultManager.GetFaultDevList(), convey.ShouldBeNil)
			convey.So(faultManager.GetNodeStatus(), convey.ShouldEqual, "")
		})
	})
}

// TestSetFaultDevList test the function of set fault dev list
func TestSetFaultDevList(t *testing.T) {
	convey.Convey("test set fault dev list", t, func() {
		convey.Convey("fault manager set fault dev list", func() {
			faultManager := NewFaultManager()
			faultDevList := []*common.FaultDev{&common.FaultDev{
				DeviceType: "CPU",
				DeviceId:   0,
				FaultCode:  []string{"01010001"},
				FaultLevel: common.NotHandleFault,
			},
			}
			faultManager.SetFaultDevList(faultDevList)
			FaultDevListEqual(faultManager.GetFaultDevList(), faultDevList)
		})
		convey.Convey("wrong fault manager set fault dev List", func() {
			faultManager := mockWrongFaultManager()
			faultDevList := []*common.FaultDev{&common.FaultDev{
				DeviceType: "CPU",
				DeviceId:   0,
				FaultCode:  []string{"01010001"},
				FaultLevel: common.NotHandleFault,
			},
			}
			faultManager.SetFaultDevList(faultDevList)
			convey.So(faultManager.GetFaultDevList(), convey.ShouldBeNil)
		})
	})
}

// TestSetNodeStatus test the function of set node status
func TestSetNodeStatus(t *testing.T) {
	convey.Convey("test set node status", t, func() {
		convey.Convey("fault manager set node status", func() {
			faultManager := NewFaultManager()
			faultManager.SetNodeStatus(common.NodeUnHealthy)
			convey.So(faultManager.GetNodeStatus(), convey.ShouldEqual, common.NodeUnHealthy)
		})
	})
}

// FaultDevInfoEqual judge if fault dev info is equal
func FaultDevInfoEqual(oldFaultDevInfo, newFaultDevInfo *common.FaultDevInfo) {
	FaultDevListEqual(oldFaultDevInfo.FaultDevList, newFaultDevInfo.FaultDevList)
	convey.So(oldFaultDevInfo.NodeStatus, convey.ShouldEqual, newFaultDevInfo.NodeStatus)
}

// FaultDevListEqual judge if fault dev list is equal
func FaultDevListEqual(oldFaultDevList, newFaultDevList []*common.FaultDev) {
	convey.So(len(oldFaultDevList), convey.ShouldEqual, len(newFaultDevList))
	for i, faultDev := range newFaultDevList {
		if i >= len(oldFaultDevList) {
			return
		}
		FaultDevEqual(oldFaultDevList[i], faultDev)
	}
}

// FaultDevEqual judge if fault dev is equal
func FaultDevEqual(oldFaultDev, newFaultDev *common.FaultDev) {
	convey.So(oldFaultDev.DeviceType, convey.ShouldEqual, newFaultDev.DeviceType)
	convey.So(oldFaultDev.DeviceId, convey.ShouldEqual, newFaultDev.DeviceId)
	convey.So(oldFaultDev.FaultLevel, convey.ShouldEqual, newFaultDev.FaultLevel)
	SliceStrEqual(oldFaultDev.FaultCode, newFaultDev.FaultCode)
}
