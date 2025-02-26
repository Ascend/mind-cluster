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
	"k8s.io/api/core/v1"

	"Ascend-device-plugin/pkg/kubeclient"
)

var resettoolTestErr = errors.New("test")

// TestResetToolInstance tests ResetToolInstance
func TestResetToolInstance(t *testing.T) {
	convey.Convey("test ResetToolInstance", t, func() {
		patch := gomonkey.ApplyMethodReturn(&kubeclient.ClientK8s{}, "GetNode",
			&v1.Node{}, resettoolTestErr)
		defer patch.Reset()
		convey.Convey("01-get node error, should return empty object", func() {
			resetTool := GetResetInfoMgr(&kubeclient.ClientK8s{})
			convey.So(resetTool.resetInfo.ManualResetDevs, convey.ShouldBeEmpty)
		})
	})
}

// TestWriteResetInfo tests WriteResetInfo
func TestWriteResetInfo(t *testing.T) {
	resetTool := &ResetInfoMgr{
		resetInfo: &ResetInfo{},
	}
	flag := false
	patch := gomonkey.ApplyPrivateMethod(&ResetInfoMgr{}, "writeNodeAnnotation", func() {
		flag = true
	})
	defer patch.Reset()
	convey.Convey("test WriteResetInfo", t, func() {
		convey.Convey("01-write cache success, flag should be true", func() {
			patch1 := gomonkey.ApplyFuncReturn(json.Marshal, []byte{}, nil)
			defer patch1.Reset()
			resetTool.WriteResetInfo(ResetInfo{}, WMOverwrite)
			convey.So(flag, convey.ShouldBeTrue)
		})
		convey.Convey("02-json marshal failed, flag should be false", func() {
			patch1 := gomonkey.ApplyFuncReturn(json.Marshal, []byte{}, resettoolTestErr)
			defer patch1.Reset()
			flag = false
			resetTool.WriteResetInfo(ResetInfo{}, WMOverwrite)
			convey.So(flag, convey.ShouldBeFalse)
		})
	})
}

// TestReadResetNodeAnnotation tests ReadResetNodeAnnotation
func TestReadResetNodeAnnotation(t *testing.T) {
	const id0 = 0
	resetTool := &ResetInfoMgr{
		resetInfo: &ResetInfo{
			ManualResetDevs: []ResetFailDevice{
				{
					PhyID: id0,
				},
			},
		},
	}
	info := resetTool.ReadResetInfo()
	if len(info.ManualResetDevs) == 0 {
		t.Errorf("devs length expected greater than zero, but got zero")
	}
}

// TestWriteNodeAnnotation tests writeNodeAnnotation
func TestWriteNodeAnnotation(t *testing.T) {
	resetTool := &ResetInfoMgr{
		resetInfo: &ResetInfo{},
	}
	flag := false
	patch := gomonkey.ApplyMethod(&kubeclient.ClientK8s{}, "AddAnnotation",
		func(_ *kubeclient.ClientK8s, key, value string) error {
			flag = true
			return resettoolTestErr
		})
	defer patch.Reset()
	convey.Convey("test writeNodeAnnotation", t, func() {
		convey.Convey("enter AddAnnotation, flag should be true", func() {
			resetTool.writeNodeAnnotation("")
			convey.So(flag, convey.ShouldBeTrue)
		})
	})
}

// TestMergeFailDevs tests mergeFailDevs
func TestMergeFailDevs(t *testing.T) {
	const id1, id2 = 1, 2
	convey.Convey("test mergeFailDevs", t, func() {
		curDevs := []ResetFailDevice{
			{PhyID: id1},
		}
		newDevs := []ResetFailDevice{
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
	const id1 = 1
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
				ManualResetDevs: []ResetFailDevice{
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
