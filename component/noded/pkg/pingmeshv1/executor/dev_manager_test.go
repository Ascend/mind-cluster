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
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/devmanager"
	"ascend-common/devmanager/common"
	"nodeD/pkg/device"
	"nodeD/pkg/pingmeshv1/types"
	_ "nodeD/pkg/testtool"
)

func patchGetDeviceManager(m *devmanager.DeviceManager) *gomonkey.Patches {
	return gomonkey.ApplyFunc(device.GetDeviceManager, func() *devmanager.DeviceManager {
		return m
	})
}

func patchGetChipBaseInfos(chips []*common.ChipBaseInfo, err error) *gomonkey.Patches {
	return gomonkey.ApplyMethod(new(devmanager.DeviceManager), "GetChipBaseInfos",
		func(_ *devmanager.DeviceManager) ([]*common.ChipBaseInfo, error) {
			return chips, err
		})
}

func patchDcGetHccsPingMeshState(state int, err error) *gomonkey.Patches {
	return gomonkey.ApplyMethod(new(devmanager.DeviceManager), "DcGetHccsPingMeshState", func(
		*devmanager.DeviceManager, int32, int32, int, uint) (int, error) {
		return state, err
	})
}

func patchGetSuperPodInfo(spInfo common.CgoSuperPodInfo, err error) *gomonkey.Patches {
	return gomonkey.ApplyMethod(new(devmanager.DeviceManager), "GetSuperPodInfo",
		func(*devmanager.DeviceManager, int32) (common.CgoSuperPodInfo, error) {
			return spInfo, err
		})
}

func TestNew(t *testing.T) {
	convey.Convey("Testing New", t, func() {
		convey.Convey("01-GetDeviceManager failed should return error", func() {
			patch := patchGetDeviceManager(nil)
			defer patch.Reset()
			_, err := New()
			convey.So(err, convey.ShouldNotBeNil)
		})
		patch := patchGetDeviceManager(&devmanager.DeviceManager{})
		defer patch.Reset()
		convey.Convey("02-GetChipBaseInfos failed should return error", func() {
			patch1 := patchGetChipBaseInfos(nil, errors.New("getChipBaseInfos failed"))
			defer patch1.Reset()
			_, err := New()
			convey.So(err, convey.ShouldNotBeNil)
		})
		patch1 := patchGetChipBaseInfos([]*common.ChipBaseInfo{{}}, nil)
		defer patch1.Reset()
		convey.Convey("03-DcGetHccsPingMeshState failed should return error", func() {
			patch2 := patchDcGetHccsPingMeshState(0, errors.New("gcGetHccsPingMeshState failed, error code: -99998"))
			defer patch2.Reset()
			_, err := New()
			convey.So(err, convey.ShouldNotBeNil)
		})
		patch2 := patchDcGetHccsPingMeshState(0, nil)
		defer patch2.Reset()
		convey.Convey("04-GetSuperPodInfo failed should return error", func() {
			patch3 := patchGetSuperPodInfo(common.CgoSuperPodInfo{}, errors.New("getSuperPodInfo failed"))
			defer patch3.Reset()
			_, err := New()
			convey.So(err, convey.ShouldNotBeNil)
		})
		patch3 := patchGetSuperPodInfo(common.CgoSuperPodInfo{}, nil)
		defer patch3.Reset()
		convey.Convey("05-GetSuperPodInfo success should return success", func() {
			_, err := New()
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestStart(t *testing.T) {
	convey.Convey("Testing Start", t, func() {
		executor := &DevManager{
			devManager:    &devmanager.DeviceManagerMock{},
			wg:            &sync.WaitGroup{},
			commandChan:   make(chan *types.HccspingMeshPolicy, 1),
			currentPolicy: nil,
			chips:         map[string]*common.ChipBaseInfo{"0": {}},
			SuperPodId:    1,
		}
		var expected atomic.Int32
		executor.SetResultHandler(func(result *types.HccspingMeshResult) {
			expected.Add(1)
		})
		convey.Convey("01-activate is off will do nothing", func() {
			stopChan := make(chan struct{})
			go executor.Start(stopChan)
			executor.UpdateConfig(&types.HccspingMeshPolicy{
				Config: &types.HccspingMeshConfig{
					Activate: "off",
				},
				UID: "",
			})
			time.Sleep(1 * time.Second)
			convey.So(expected.Load(), convey.ShouldEqual, 0)
			close(stopChan)
		})
		convey.Convey("02-activate is on will change expected", func() {
			stopChan := make(chan struct{})
			go executor.Start(stopChan)
			executor.UpdateConfig(&types.HccspingMeshPolicy{
				Config: &types.HccspingMeshConfig{
					Activate:     "on",
					TaskInterval: 1,
				},
				DestAddr: map[string]types.DestinationAddress{
					"0": {0: "127.0.0.1"},
				},
			})
			const sleepNum = 11
			time.Sleep(sleepNum * time.Second)
			convey.So(expected.Load(), convey.ShouldEqual, 1)
			close(stopChan)
		})
	})
}
