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
Package filewriter is using for pingmeshv1 result write to file
*/

package filewriter

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/hwlog"
	"ascend-common/devmanager/common"
	"nodeD/pkg/pingmeshv1/types"
	_ "nodeD/pkg/testtool"
)

func patchNewCustomLogger(log *hwlog.CustomLogger, err error) *gomonkey.Patches {
	return gomonkey.ApplyFunc(hwlog.NewCustomLogger, func(*hwlog.LogConfig,
		context.Context) (*hwlog.CustomLogger, error) {
		return log, err
	})
}

func TestNew(t *testing.T) {
	convey.Convey("TestNew", t, func() {
		convey.Convey("01-nil config should return nil", func() {
			fm := New(nil)
			convey.So(fm, convey.ShouldBeNil)
		})
		convey.Convey("02-config with empty path should return nil", func() {
			fm := New(&Config{Path: ""})
			convey.So(fm, convey.ShouldBeNil)
		})
		convey.Convey("03-new custom logger failed should return error", func() {
			patch := patchNewCustomLogger(nil, errors.New("new custom logger failed"))
			defer patch.Reset()
			fm := New(&Config{Path: "test"})
			convey.So(fm, convey.ShouldBeNil)
		})
		convey.Convey("04-new custom logger success should return manager", func() {
			patch := patchNewCustomLogger(&hwlog.CustomLogger{}, nil)
			defer patch.Reset()
			fm := New(&Config{Path: "test"})
			convey.So(fm, convey.ShouldNotBeNil)
		})
	})
}

func patchGetEnv(node string) *gomonkey.Patches {
	return gomonkey.ApplyFunc(os.Getenv, func(key string) string {
		return node
	})
}

func TestHandlePingMeshInfo(t *testing.T) {
	convey.Convey("TestHandlePingMeshInfo", t, func() {
		m := mockManager()
		convey.Convey("01-nil hccsping mesh result will return error", func() {
			err := m.HandlePingMeshInfo(nil)
			convey.So(err, convey.ShouldNotBeNil)
		})
		patch := patchGetEnv("node1")
		defer patch.Reset()
		convey.Convey("02-result without localhost will do nothing", func() {
			err := m.HandlePingMeshInfo(&types.HccspingMeshResult{
				Policy: &types.HccspingMeshPolicy{
					DestAddr: map[string]types.DestinationAddress{"node2": {}},
				},
				Results: map[string]map[uint]*common.HccspingMeshInfo{"1": {0: {}}},
			})
			convey.So(err, convey.ShouldBeNil)
		})
		patch1 := patchGetEnv("node2")
		defer patch1.Reset()
		convey.Convey("03-result without card will do nothing", func() {
			err := m.HandlePingMeshInfo(&types.HccspingMeshResult{
				Policy: &types.HccspingMeshPolicy{
					Address: map[string]types.SuperDeviceIDs{"node2": {}},
				},
				Results: map[string]map[uint]*common.HccspingMeshInfo{"1": {}},
			})
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("04-result without card will do nothing", func() {
			err := m.HandlePingMeshInfo(&types.HccspingMeshResult{
				Policy: &types.HccspingMeshPolicy{
					Address: map[string]types.SuperDeviceIDs{"node2": {"1": "4259841"}},
				},
				Results: map[string]map[uint]*common.HccspingMeshInfo{"1": {0: mockHccspingMeshInfo()}},
			})
			const perm = 0666
			f, err := os.OpenFile("test", os.O_RDONLY, perm)
			defer f.Close()
			convey.So(err, convey.ShouldBeNil)
			err = os.Remove("test")
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func mockManager() *manager {
	const defaultMaxAge = 7
	return New(&Config{Path: "test", MaxAge: defaultMaxAge})
}

func mockHccspingMeshInfo() *common.HccspingMeshInfo {
	return &common.HccspingMeshInfo{
		DstAddr:      []string{"111"},
		SucPktNum:    []uint{1},
		FailPktNum:   []uint{1},
		MaxTime:      []int{1},
		MinTime:      []int{1},
		AvgTime:      []int{1},
		TP95Time:     []int{1},
		ReplyStatNum: []int{1},
		PingTotalNum: []int{1},
		DestNum:      1,
	}
}
