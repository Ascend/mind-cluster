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

// Package topology for generating topology of Rack
package topology

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"Ascend-device-plugin/pkg/common"
	"Ascend-device-plugin/pkg/device"
	"Ascend-device-plugin/pkg/server"
	"ascend-common/api"
	"ascend-common/api/slownet"
	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
)

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwlog.InitRunLogger(&hwLogConfig, context.Background())
	common.ParamOption.PresetVDevice = true
}

// TestWriteTopology test case for WriteTopology
func TestWriteTopology(t *testing.T) {
	var topoFilePath string
	dir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error:", err)
		topoFilePath = "/tmp/topology.json"
	} else {
		topoFilePath = filepath.Join(dir, "topology.json")
	}

	defer func() {
		_, pathErr := os.Stat(topoFilePath)
		if os.IsNotExist(pathErr) {
			return
		}
		err := os.Remove(topoFilePath)
		if err != nil {
			fmt.Println("Error:", err)
		}
	}()

	doTestWriteTopology1(t, topoFilePath)
}

func doTestWriteTopology1(t *testing.T, topoFilePath string) {
	convey.Convey("test function GetTopoFileAndWrite superPodType is invalid", t, func() {
		superPodType = 5
		var count int32
		mock1 := gomonkey.ApplyFunc(ToFile,
			func(_, _ string) error {
				atomic.StoreInt32(&count, 1)
				return fmt.Errorf("err")
			})
		defer mock1.Reset()
		GetTopoFileAndWrite(topoFilePath)
		convey.So(atomic.LoadInt32(&count), convey.ShouldEqual, 0)
	})

	convey.Convey("test function  GetTopoFileAndWrite ToFile failed", t, func() {
		superPodType = 4
		var count int32
		mock1 := gomonkey.ApplyFunc(ToFile,
			func(_, _ string) error {
				atomic.StoreInt32(&count, 1)
				return fmt.Errorf("err")
			})
		defer mock1.Reset()
		GetTopoFileAndWrite(topoFilePath)
		convey.So(atomic.LoadInt32(&count), convey.ShouldEqual, 1)
	})

	convey.Convey("test function  GetTopoFileAndWrite ToFile success", t, func() {
		superPodType = 4
		var count int32
		mock1 := gomonkey.ApplyFunc(ToFile,
			func(_, _ string) error {
				atomic.StoreInt32(&count, 1)
				return nil
			})
		defer mock1.Reset()
		GetTopoFileAndWrite(topoFilePath)
		convey.So(atomic.LoadInt32(&count), convey.ShouldEqual, 1)
	})
}

func TestCheckConfigReady(t *testing.T) {
	TestCheckConfigReadyPart1(t)
	TestCheckConfigReadyPart2(t)
	TestCheckConfigReadyPart3(t)
	TestCheckConfigReadyPart4(t)
}

func TestCheckConfigReadyPart1(t *testing.T) {
	convey.Convey("test checkConfigReady case 1 GetSuperPodInfoFilePath err", t, func() {
		mock1 := gomonkey.ApplyFunc(slownet.GetSuperPodInfoFilePath, func(_, _ string) (string, error) {
			return "", errors.New("fake error")
		})
		defer mock1.Reset()
		filePath, ok := checkConfigReady("1")
		convey.So(filePath, convey.ShouldEqual, "")
		convey.So(ok, convey.ShouldEqual, false)
	})
}

func TestCheckConfigReadyPart2(t *testing.T) {
	convey.Convey("test checkConfigReady case 2 fileParentDir not exist", t, func() {
		mock1 := gomonkey.ApplyFunc(slownet.GetSuperPodInfoFilePath, func(_, _ string) (string, error) {
			return "", nil
		})
		defer mock1.Reset()
		mock2 := gomonkey.ApplyFunc(utils.IsLexist,
			func(_ string) bool {
				return false
			})
		defer mock2.Reset()
		_, ok := checkConfigReady("1")
		convey.So(ok, convey.ShouldEqual, false)
	})
}

func TestCheckConfigReadyPart3(t *testing.T) {
	convey.Convey("test checkConfigReady case 3 GetRackTopologyFilePath not exist", t, func() {
		mock1 := gomonkey.ApplyFunc(slownet.GetSuperPodInfoFilePath, func(_, _ string) (string, error) {
			return "", nil
		})
		defer mock1.Reset()
		mock2 := gomonkey.ApplyFunc(utils.IsLexist,
			func(_ string) bool {
				return true
			})
		defer mock2.Reset()
		mock3 := gomonkey.ApplyFunc(slownet.GetRackTopologyFilePath,
			func(_, _, _ int32) (string, error) {
				return "", errors.New("fake error")
			})
		defer mock3.Reset()
		_, ok := checkConfigReady("1")
		convey.So(ok, convey.ShouldEqual, false)
	})
}

func TestCheckConfigReadyPart4(t *testing.T) {
	convey.Convey("test checkConfigReady case 4 normal", t, func() {
		mock1 := gomonkey.ApplyFunc(slownet.GetSuperPodInfoFilePath, func(_, _ string) (string, error) {
			return "", nil
		})
		defer mock1.Reset()
		mock2 := gomonkey.ApplyFunc(utils.IsLexist,
			func(_ string) bool {
				return true
			})
		defer mock2.Reset()
		mock3 := gomonkey.ApplyFunc(slownet.GetRackTopologyFilePath,
			func(_, _, _ int32) (string, error) {
				return "", nil
			})
		defer mock3.Reset()
		mock4 := gomonkey.ApplyFunc(os.Chmod,
			func(_ string, _ fs.FileMode) error {
				return nil
			})
		defer mock4.Reset()
		_, ok := checkConfigReady("1")
		convey.So(ok, convey.ShouldEqual, true)
	})
}

func TestRasTopoWriteTask(t *testing.T) {
	convey.Convey("Test RasTopoWriteTask", t, func() {
		common.ParamOption.RealCardType = common.Ascend910A5
		var count int32
		ctx, cancel := context.WithCancel(context.Background())
		hdm := &server.HwDevManager{}

		patches := gomonkey.NewPatches()
		defer patches.Reset()

		patches.ApplyFunc(slownet.GetRasNetRootPath, func() (string, error) { return "/mock/path", nil })
		patches.ApplyMethod(&server.HwDevManager{}, "GetSuperPodID", func(_ *server.HwDevManager) int32 { return 0 })
		patches.ApplyMethod(&server.HwDevManager{}, "GetSuperPodType", func(_ *server.HwDevManager) int8 { return 0 })
		patches.ApplyMethod(&server.HwDevManager{}, "GetRackID", func(_ *server.HwDevManager) int32 { return 0 })
		manager := device.NewHwAscend910Manager()
		manager.SetServerIndex(1)
		patches.ApplyMethod(&server.HwDevManager{}, "GetDevManager", func(_ *server.HwDevManager) device.DevManager { return manager })
		patches.ApplyFunc(checkConfigReady, func(podID string) (string, bool) {
			atomic.StoreInt32(&count, 1)
			return "", false
		})

		go RasTopoWriteTask(ctx, hdm)
		time.Sleep(common.TopologyRefreshTime * time.Millisecond)

		convey.So(atomic.LoadInt32(&count), convey.ShouldEqual, 1)
		cancel()
		time.Sleep(common.TopologyRefreshTime * time.Millisecond)
	})
}

func TestRasTopoWriteTask2(t *testing.T) {
	convey.Convey("Test RasTopoWriteTask2", t, func() {
		common.ParamOption.RealCardType = api.Ascend310P
		var count int32
		ctx, cancel := context.WithCancel(context.Background())
		hdm := &server.HwDevManager{}

		patches := gomonkey.NewPatches()
		defer patches.Reset()

		patches.ApplyFunc(slownet.GetRasNetRootPath, func() (string, error) { return "/mock/path", nil })
		patches.ApplyMethod(&server.HwDevManager{}, "GetSuperPodID", func(_ *server.HwDevManager) int32 { return 0 })
		patches.ApplyMethod(&server.HwDevManager{}, "GetSuperPodType", func(_ *server.HwDevManager) int8 { return 0 })
		patches.ApplyMethod(&server.HwDevManager{}, "GetRackID", func(_ *server.HwDevManager) int32 { return 0 })
		manager := device.NewHwAscend910Manager()
		manager.SetServerIndex(1)
		patches.ApplyMethod(&server.HwDevManager{}, "GetDevManager", func(_ *server.HwDevManager) device.DevManager { return manager })
		patches.ApplyFunc(checkConfigReady, func(podID string) (string, bool) {
			atomic.StoreInt32(&count, 1)
			return "", false
		})

		go RasTopoWriteTask(ctx, hdm)
		time.Sleep(common.TopologyRefreshTime * time.Millisecond)

		convey.So(atomic.LoadInt32(&count), convey.ShouldEqual, 0)
		cancel()
		time.Sleep(common.TopologyRefreshTime * time.Millisecond)
	})
}

func TestRasTopoWriteTask3(t *testing.T) {
	convey.Convey("Test RasTopoWriteTask", t, func() {
		common.ParamOption.RealCardType = common.Ascend910A5
		var count int32
		ctx, cancel := context.WithCancel(context.Background())
		hdm := &server.HwDevManager{}

		patches := gomonkey.NewPatches()
		defer patches.Reset()

		patches.ApplyFunc(slownet.GetRasNetRootPath, func() (string, error) { return "/mock/path", fmt.Errorf("err") })
		patches.ApplyMethod(&server.HwDevManager{}, "GetSuperPodID", func(_ *server.HwDevManager) int32 { return 0 })
		patches.ApplyMethod(&server.HwDevManager{}, "GetSuperPodType", func(_ *server.HwDevManager) int8 { return 0 })
		patches.ApplyMethod(&server.HwDevManager{}, "GetRackID", func(_ *server.HwDevManager) int32 { return 0 })
		manager := device.NewHwAscend910Manager()
		manager.SetServerIndex(1)
		patches.ApplyMethod(&server.HwDevManager{}, "GetDevManager", func(_ *server.HwDevManager) device.DevManager { return manager })
		patches.ApplyFunc(checkConfigReady, func(podID string) (string, bool) {
			atomic.StoreInt32(&count, 1)
			return "", false
		})

		go RasTopoWriteTask(ctx, hdm)
		time.Sleep(common.TopologyRefreshTime * time.Millisecond)

		convey.So(atomic.LoadInt32(&count), convey.ShouldEqual, 0)
		cancel()
		time.Sleep(common.TopologyRefreshTime * time.Millisecond)
	})
}

func TestRasTopoWriteTask5(t *testing.T) {
	convey.Convey("Test RasTopoWriteTask", t, func() {
		common.ParamOption.RealCardType = common.Ascend910A5
		var count int32
		ctx, cancel := context.WithCancel(context.Background())
		var hdm server.HwDevManager

		patches := gomonkey.NewPatches()
		defer patches.Reset()

		patches.ApplyFunc(slownet.GetRasNetRootPath, func() (string, error) { return "/mock/path", fmt.Errorf("err") })
		patches.ApplyMethod(&server.HwDevManager{}, "GetSuperPodID", func(_ *server.HwDevManager) int32 { return 0 })
		patches.ApplyMethod(&server.HwDevManager{}, "GetSuperPodType", func(_ *server.HwDevManager) int8 { return 0 })
		patches.ApplyMethod(&server.HwDevManager{}, "GetRackID", func(_ *server.HwDevManager) int32 { return 0 })
		manager := device.NewHwAscend910Manager()
		manager.SetServerIndex(1)
		patches.ApplyMethod(&server.HwDevManager{}, "GetDevManager", func(_ *server.HwDevManager) device.DevManager { return manager })
		patches.ApplyFunc(checkConfigReady, func(podID string) (string, bool) {
			atomic.StoreInt32(&count, 1)
			return "", false
		})

		go RasTopoWriteTask(ctx, &hdm)
		time.Sleep(common.TopologyRefreshTime * time.Millisecond)

		convey.So(atomic.LoadInt32(&count), convey.ShouldEqual, 0)
		cancel()
		time.Sleep(common.TopologyRefreshTime * time.Millisecond)
	})
}

func TestRasTopoWriteTask4(t *testing.T) {
	convey.Convey("Test RasTopoWriteTask4", t, func() {
		common.ParamOption.RealCardType = common.Ascend910A5
		var tag1 int32
		var tag2 int32
		ctx, cancel := context.WithCancel(context.Background())
		hdm := &server.HwDevManager{}

		patches := gomonkey.NewPatches()
		defer patches.Reset()

		patches.ApplyFunc(slownet.GetRasNetRootPath, func() (string, error) { return "/mock/path", nil })
		patches.ApplyMethod(&server.HwDevManager{}, "GetSuperPodID", func(_ *server.HwDevManager) int32 { return 0 })
		patches.ApplyMethod(&server.HwDevManager{}, "GetSuperPodType", func(_ *server.HwDevManager) int8 { return 0 })
		patches.ApplyMethod(&server.HwDevManager{}, "GetRackID", func(_ *server.HwDevManager) int32 { return 0 })
		manager := device.NewHwAscend910Manager()
		manager.SetServerIndex(1)
		patches.ApplyMethod(&server.HwDevManager{}, "GetDevManager", func(_ *server.HwDevManager) device.DevManager { return manager })
		patches.ApplyFunc(checkConfigReady, func(podID string) (string, bool) {
			atomic.StoreInt32(&tag1, 1)
			return "", true
		})
		patches.ApplyFunc(GetTopoFileAndWrite, func(topoJsonFile string) {
			atomic.StoreInt32(&tag2, 1)
		})

		go RasTopoWriteTask(ctx, hdm)
		time.Sleep(common.TopologyRefreshTime * time.Millisecond)

		convey.So(atomic.LoadInt32(&tag1), convey.ShouldEqual, 1)
		convey.So(atomic.LoadInt32(&tag2), convey.ShouldEqual, 1)
		cancel()
		time.Sleep(common.TopologyRefreshTime * time.Millisecond)
	})
}
