/* Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package device a series of device function
package device

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"Ascend-device-plugin/pkg/kubeclient"
)

const (
	testCardId, testDeviceId = 0, 0
	id1, id2, id3            = 1, 2, 3
	zeroVal                  = 0
)

var resettoolTestErr = errors.New("test")

func init() {
	mgr = &ResetInfoMgr{resetInfo: &ResetInfo{}}
}

// TestWriteResetInfo tests WriteResetInfo
func TestWriteResetInfo(t *testing.T) {
	flag := false
	patch := gomonkey.ApplyFunc(writeNodeAnnotation, func(resetStr string) {
		flag = true
	})
	defer patch.Reset()
	convey.Convey("test WriteResetInfo", t, func() {
		convey.Convey("01-write cache success, flag should be true", func() {
			patch1 := gomonkey.ApplyFuncReturn(json.Marshal, []byte{}, nil)
			defer patch1.Reset()
			WriteResetInfo(ResetInfo{}, WMOverwrite, true)
			convey.So(flag, convey.ShouldBeTrue)
		})
		convey.Convey("02-json marshal failed, flag should be false", func() {
			patch1 := gomonkey.ApplyFuncReturn(json.Marshal, []byte{}, resettoolTestErr)
			defer patch1.Reset()
			flag = false
			WriteResetInfo(ResetInfo{}, WMOverwrite, true)
			convey.So(flag, convey.ShouldBeFalse)
		})
	})
}

// TestIsDevBusy test the function IsDevBusy
func TestIsDevBusy(t *testing.T) {
	AddBusyDev(testCardId, testDeviceId)
	ret := IsDevBusy(testCardId, testDeviceId)
	if !ret {
		t.Errorf("expected true, got false")
	}
}

// TestFreeBusyDev test the function FreeBusyDev
func TestFreeBusyDev(t *testing.T) {
	AddBusyDev(testCardId, testDeviceId)
	FreeBusyDev(testCardId, testDeviceId)
	if IsDevBusy(testCardId, testDeviceId) {
		t.Errorf("expected false, got true")
	}

	// free twice
	FreeBusyDev(testCardId, testDeviceId)
	if IsDevBusy(testCardId, testDeviceId) {
		t.Errorf("expected false, got true")
	}
}

// TestGetResetCnt test the function GetResetCnt
func TestGetResetCnt(t *testing.T) {
	const testNum = 1
	convey.Convey("test GetResetCnt", t, func() {
		convey.Convey("01-not exist, should return 0", func() {
			SetResetCnt(testCardId, testDeviceId, zeroVal)
			cnt := GetResetCnt(testCardId, testDeviceId)
			convey.So(cnt, convey.ShouldEqual, zeroVal)
		})
		convey.Convey("02-set to 1, should return 1", func() {
			SetResetCnt(testCardId, testDeviceId, testNum)
			cnt := GetResetCnt(testCardId, testDeviceId)
			convey.So(cnt, convey.ShouldEqual, testNum)
		})
	})
}

// TestAddResetCnt test the function AddResetCnt
func TestAddResetCnt(t *testing.T) {
	SetResetCnt(testCardId, testDeviceId, zeroVal)
	AddResetCnt(testCardId, testDeviceId)
	ret := GetResetCnt(testCardId, testDeviceId)
	const expectVal = 1
	if ret != expectVal {
		t.Errorf("expect %v, got %v", expectVal, ret)
	}
}

// TestSetResetCnt test the function SetResetCnt
func TestSetResetCnt(t *testing.T) {
	const testVal = 1
	SetResetCnt(testCardId, testDeviceId, testVal)
	ret := GetResetCnt(testCardId, testDeviceId)
	if ret != testVal {
		t.Errorf("expect %v, got %v", testVal, ret)
	}
}

// TestMergeAndDeduplicate test the function mergeAndDeduplicate
func TestMergeAndDeduplicate(t *testing.T) {
	convey.Convey("test mergeAndDeduplicate", t, func() {
		arr1 := []ResetDevice{
			{PhyID: id1},
			{PhyID: id2},
		}
		arr2 := []ResetDevice{
			{PhyID: id2},
			{PhyID: id3},
		}
		expect := []ResetDevice{
			{PhyID: id1},
			{PhyID: id2},
			{PhyID: id3},
		}
		ret := mergeAndDeduplicate(arr1, arr2)
		convey.So(ret, convey.ShouldResemble, expect)
	})
}

// TestReadResetNodeAnnotation tests ReadResetNodeAnnotation
func TestReadResetNodeAnnotation(t *testing.T) {
	const id0 = 0
	patch := gomonkey.ApplyFunc(writeNodeAnnotation, func(resetStr string) {
		return
	})
	defer patch.Reset()
	WriteResetInfo(ResetInfo{
		ManualResetDevs: []ResetDevice{
			{
				PhyID: id0,
			},
		},
	}, WMOverwrite, true)
	info := ReadResetInfo()
	if len(info.ManualResetDevs) == 0 {
		t.Errorf("devs length expected greater than zero, but got zero")
	}
}

// TestWriteNodeAnnotation tests writeNodeAnnotation
func TestWriteNodeAnnotation(t *testing.T) {
	flag := false
	patch := gomonkey.ApplyMethod(&kubeclient.ClientK8s{}, "AddAnnotation",
		func(_ *kubeclient.ClientK8s, key, value string) error {
			flag = true
			return resettoolTestErr
		})
	defer patch.Reset()
	convey.Convey("test writeNodeAnnotation", t, func() {
		convey.Convey("enter AddAnnotation, flag should be true", func() {
			writeNodeAnnotation("")
			convey.So(flag, convey.ShouldBeTrue)
		})
	})
}

// TestMergeFailDevs tests mergeFailDevs
func TestMergeFailDevs(t *testing.T) {

	convey.Convey("test mergeFailDevs", t, func() {
		curDevs := []ResetDevice{
			{PhyID: id1},
		}
		newDevs := []ResetDevice{
			{PhyID: id2},
		}
		convey.Convey("01-overwrite mode, should return newDevs", func() {
			devs := mergeFailDevs(curDevs, newDevs, WMOverwrite)
			convey.So(devs, convey.ShouldResemble, newDevs)
		})
		convey.Convey("02-append mode, should merge devs", func() {
			devs := mergeFailDevs(curDevs, newDevs, WMAppend)
			mergeDevs := append(curDevs, newDevs...)
			convey.So(devs, convey.ShouldResemble, mergeDevs)
		})
	})
}

// TestReadAnnotation tests readAnnotation
func TestReadAnnotation(t *testing.T) {
	convey.Convey("test readAnnotation", t, func() {
		convey.Convey("01-key not exists, should return empty info", func() {
			annotations := map[string]string{}
			key := "test"
			info := readAnnotation(annotations, key)
			convey.So(info.ThirdPartyResetDevs, convey.ShouldBeEmpty)
		})
		convey.Convey("02-unmarshal error, should return empty info", func() {
			annotations := map[string]string{"test": "test"}
			key := "test"
			patch1 := gomonkey.ApplyFuncReturn(json.Unmarshal, resettoolTestErr)
			defer patch1.Reset()
			info := readAnnotation(annotations, key)
			convey.So(info.ThirdPartyResetDevs, convey.ShouldBeEmpty)
		})
		convey.Convey("03-success, should return info", func() {
			res := &ResetInfo{
				ManualResetDevs: []ResetDevice{
					{PhyID: id1},
				},
			}
			patch := gomonkey.ApplyFunc(json.Unmarshal, func(data []byte, v any) error {
				v = res
				return nil
			})
			defer patch.Reset()
			annotations := map[string]string{"test": "test"}
			key := "test"
			info := readAnnotation(annotations, key)
			convey.So(*info, convey.ShouldResemble, *res)
		})
	})
}

// TestDeduplicate test function deduplicate
func TestDeduplicate(t *testing.T) {
	convey.Convey("test deduplicate", t, func() {
		devs := []ResetDevice{
			{}, {},
		}
		target := []ResetDevice{
			{},
		}
		ret := deduplicate(devs)
		convey.So(ret, convey.ShouldResemble, target)
	})
}
