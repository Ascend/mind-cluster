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
Package executor is using for execute hccsping mesh
*/

package executor

import (
	"encoding/hex"
	"errors"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/devmanager"
	"ascend-common/devmanager/common"
	"nodeD/pkg/pingmesh/types"
)

const (
	testCardID0   = 0
	testCardID1   = 1
	testCardID2   = 2
	testDeviceID0 = 0
	testDeviceID1 = 1
	testDeviceID2 = 2
)

type TestCaseBuildUb struct {
	name     string
	grouped  map[string][]string
	interval int
	expected []common.UBPingMeshOperate
}

type TestCaseGroupPing struct {
	name     string
	items    []types.PingItem
	expected map[string][]string
}

func TestGroupPingItemsBySrcAddr(t *testing.T) {
	testCases := []TestCaseGroupPing{
		{
			name: "Normal case",
			items: []types.PingItem{
				{
					SrcType:      1,
					DstType:      1,
					PktSize:      1,
					SrcCardPhyId: 1,
					SrcAddr:      "ip1",
					DstAddr:      "ip2",
				},
			},
			expected: map[string][]string{
				"ip1": {"ip2"},
			},
		},
		{
			name:     "Empty case",
			items:    []types.PingItem{},
			expected: map[string][]string{},
		},
	}
	convey.Convey("TestGroupPingItemsBySrcAddr", t, func() {
		for _, tc := range testCases {
			convey.Convey(tc.name, func() {
				grouped := groupPingItemsBySrcAddr(tc.items)
				convey.So(reflect.DeepEqual(grouped, tc.expected), convey.ShouldBeTrue)
			})
		}
	})
}

func buildCaseBuildUb() []TestCaseBuildUb {
	testCases := []TestCaseBuildUb{
		{
			name: "Normal case",
			grouped: map[string][]string{
				"af33": {"cf01", "de44"},
			},
			interval: 1,
			expected: []common.UBPingMeshOperate{
				{
					SrcEID: common.Eid{Raw: [16]byte{175, 51}},
					DstEIDList: []common.Eid{
						{Raw: [16]byte{207, 1}},
						{Raw: [16]byte{222, 68}},
					},
					DstNum:       2,
					PktSize:      common.DefaultPktSize,
					PktSendNum:   common.DefaultPktSendNum,
					PktInterval:  common.DefaultPktInterval,
					Timeout:      common.DefaultTimeout,
					TaskInterval: 1,
					TaskID:       0,
				},
			},
		},
		{
			name:     "Empty groups",
			grouped:  map[string][]string{},
			interval: 1,
			expected: []common.UBPingMeshOperate{},
		},
	}
	return testCases
}

func TestBuildUbOperateList(t *testing.T) {
	testCases := buildCaseBuildUb()
	convey.Convey("TestBuildUbOperateList", t, func() {
		for _, tc := range testCases {
			convey.Convey(tc.name, func() {
				ubt := buildUbOperateList(tc.grouped, tc.interval)
				convey.So(reflect.DeepEqual(ubt, tc.expected), convey.ShouldBeTrue)
			})
		}
	})
	convey.Convey("TestBuildUbOperateList continue for decode string failed", t, func() {
		patch := gomonkey.ApplyFuncReturn(hex.DecodeString, []byte{}, errors.New("failed"))
		defer patch.Reset()
		for _, tc := range testCases {
			convey.Convey(tc.name, func() {
				ubt := buildUbOperateList(tc.grouped, tc.interval)
				convey.ShouldEqual(len(ubt), 0)
			})
		}
	})
}

func TestDevManagerStartUbPingMesh(t *testing.T) {
	convey.Convey("Test startUbPingMesh", t, func() {
		mockDevManager := &DevManager{
			devManager: &devmanager.DeviceManagerMock{},
			currentPolicy: &types.HccspingMeshPolicy{
				DestAddrMap: map[string][]types.PingItem{
					"physicID1": {
						{
							SrcType:      1,
							DstType:      1,
							PktSize:      1,
							SrcCardPhyId: 1,
							SrcAddr:      "ip1",
							DstAddr:      "ip2",
						},
					},
				},
				Config: &types.HccspingMeshConfig{
					TaskInterval: 100,
				},
			},
			chips: map[string]*common.ChipBaseInfo{
				"physicID1": {
					CardID:   1,
					DeviceID: 101,
					PhysicID: 1,
					LogicID:  1,
				},
			},
		}
		convey.Convey("01-should start ub ping mesh successfully", func() {
			count := 0
			patch := gomonkey.ApplyFunc(groupPingItemsBySrcAddr, func(items []types.PingItem) map[string][]string {
				count += len(items)
				return map[string][]string{}
			})
			defer patch.Reset()
			mockDevManager.startUbPingMesh()
			convey.So(count, convey.ShouldEqual, 1)
		})
	})
}

func TestDevManagerRestartUbPingMesh(t *testing.T) {
	convey.Convey("Test restartUbPingMesh", t, func() {
		var d *DevManager
		convey.Convey("01-should restart successfully when all conditions are met", func() {
			d = &DevManager{
				chips: map[string]*common.ChipBaseInfo{
					"physicID1": {CardID: 1, DeviceID: 1},
					"phtsicID2": {CardID: 2, DeviceID: 2},
				},
				currentPolicy: &types.HccspingMeshPolicy{
					DestAddrMap: map[string][]types.PingItem{
						"physicID1": {
							{SrcAddr: "src1", DstAddr: "dst1"},
							{SrcAddr: "src1", DstAddr: "dst2"},
						},
					},
					Config: &types.HccspingMeshConfig{
						TaskInterval: 100,
					},
				},
				devManager: &devmanager.DeviceManagerMock{},
			}
			count := 0
			patch := gomonkey.ApplyFunc(groupPingItemsBySrcAddr, func(items []types.PingItem) map[string][]string {
				count++
				return map[string][]string{}
			})
			defer patch.Reset()
			d.restartUbPingMesh(testCardID0, testDeviceID0)
			convey.So(count, convey.ShouldEqual, 0)
			d.restartUbPingMesh(testCardID1, testDeviceID1)
			convey.So(count, convey.ShouldEqual, 1)
			d.restartUbPingMesh(testCardID2, testDeviceID2)
			convey.So(count, convey.ShouldEqual, 1)
		})
	})
}
