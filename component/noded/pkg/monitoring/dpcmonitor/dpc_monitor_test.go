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

// Package dpcmonitor for monitor the fault by dpc on the server
package dpcmonitor

import (
	"bufio"
	"context"
	"fmt"

	"os"
	"strings"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	"nodeD/pkg/common"
)

const (
	int6 = 6
)

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	err := hwlog.InitRunLogger(&hwLogConfig, context.Background())
	if err != nil {
		return
	}
}

func TestNewDpcEventMonitor(t *testing.T) {
	convey.Convey("Test NewDpcEventMonitor", t, func() {
		ctx := context.Background()
		monitor := NewDpcEventMonitor(ctx)
		convey.So(monitor, convey.ShouldNotBeNil)
		convey.So(monitor.stopChan, convey.ShouldNotBeNil)
	})
}

func TestDpcEventMonitor_Init(t *testing.T) {
	convey.Convey("Test Init", t, func() {
		monitor := NewDpcEventMonitor(context.Background())
		convey.So(monitor.Init(), convey.ShouldBeNil)
	})
}

func TestDpcEventMonitor_Stop(t *testing.T) {
	convey.Convey("Test Stop", t, func() {
		monitor := NewDpcEventMonitor(context.Background())
		monitor.Stop()
		_, ok := <-monitor.stopChan
		convey.So(ok, convey.ShouldBeTrue)
	})
}

func TestDpcEventMonitor_Name(t *testing.T) {
	convey.Convey("Test Name", t, func() {
		monitor := NewDpcEventMonitor(context.Background())
		convey.So(monitor.Name(), convey.ShouldEqual, common.PluginMonitorDpc)
	})
}

func TestDpcEventMonitor_GetMonitorData(t *testing.T) {
	convey.Convey("Test GetMonitorData", t, func() {
		monitor := NewDpcEventMonitor(context.Background())
		data := monitor.GetMonitorData()
		convey.So(data, convey.ShouldNotBeNil)
		convey.So(data.DpcStatusMap, convey.ShouldEqual, dpcMap)
	})
}

func TestSetNewDpcMapTime(t *testing.T) {
	convey.Convey("Test setNewDpcMapTime", t, func() {
		now := time.Now().UnixMilli()
		dpcMap = map[int]common.DpcStatus{1: {ProcessError: true, MemoryError: false, ProcessErrorTime: now}}
		newMap := map[int]common.DpcStatus{
			1: {ProcessError: true, MemoryError: true, ProcessErrorTime: now},
			2: {ProcessError: false, MemoryError: false, ProcessErrorTime: now},
		}

		result := setNewDpcMapTime(newMap)
		convey.Convey("should keep old time for unchanged status", func() {
			convey.So(result[1].ProcessErrorTime, convey.ShouldEqual, now)
		})
		convey.Convey("should update time for changed status", func() {
			convey.So(result[1].MemoryErrorTime, convey.ShouldBeGreaterThanOrEqualTo, now)
		})
		convey.Convey("should set current time for new entry", func() {
			convey.So(result[2].ProcessErrorTime, convey.ShouldBeGreaterThanOrEqualTo, now)
		})
	})
}

func TestIsSame(t *testing.T) {
	dpcMap = map[int]common.DpcStatus{1: {ProcessError: true, MemoryError: false}}
	convey.Convey("Test isSame", t, func() {
		map1 := map[int]common.DpcStatus{1: {ProcessError: true, MemoryError: false}}
		map2 := map[int]common.DpcStatus{1: {ProcessError: false, MemoryError: true}}
		map3 := map[int]common.DpcStatus{1: {ProcessError: true, MemoryError: false},
			2: {ProcessError: false, MemoryError: true}}
		convey.Convey("should return false for lastUpdateTime is 0", func() {
			lastUploadTime = 0
			convey.So(isSame(map1), convey.ShouldBeFalse)
		})
		convey.Convey("should return true for same maps", func() {
			lastUploadTime = time.Now().UnixMilli()
			convey.So(isSame(map1), convey.ShouldBeTrue)
		})
		convey.Convey("should return false for different length", func() {
			convey.So(isSame(map2), convey.ShouldBeFalse)
		})
		convey.Convey("should return false for different status", func() {
			convey.So(isSame(map3), convey.ShouldBeFalse)
		})
	})
}

func TestGetStatusByText(t *testing.T) {
	convey.Convey("Test getStatusByText", t, func() {
		convey.Convey("should handle DPC_INTERNAL_ERROR healthy", func() {
			status, err := getStatusByText("DPC_INTERNAL_ERROR: 0", dpcInternalErrorKey)
			convey.So(err, convey.ShouldBeNil)
			convey.So(status, convey.ShouldBeFalse)
		})
		convey.Convey("should handle DPC_INTERNAL_ERROR error", func() {
			status, err := getStatusByText("DPC_INTERNAL_ERROR: -12", dpcInternalErrorKey)
			convey.So(err, convey.ShouldBeNil)
			convey.So(status, convey.ShouldBeTrue)
		})
		convey.Convey("should handle invalid format", func() {
			_, err := getStatusByText("invalid text", dpcInternalErrorKey)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestReadInstStatus(t *testing.T) {
	convey.Convey("Test readInstStatus", t, func() {
		input := "[instidx=1]\nDPC_INTERNAL_ERROR: 0\nDPC_PROCESS_ERROR: -1"
		scanner := bufio.NewScanner(strings.NewReader(input))
		scanner.Scan()

		inst, status, err := readInstStatus(scanner)
		convey.Convey("should read inst correctly", func() {
			convey.So(err, convey.ShouldBeNil)
			convey.So(inst, convey.ShouldEqual, 1)
			convey.So(status.MemoryError, convey.ShouldBeFalse)
			convey.So(status.ProcessError, convey.ShouldBeTrue)
		})

		convey.Convey("should fail on invalid inst format", func() {
			badInput := "invalid_inst"
			scanner := bufio.NewScanner(strings.NewReader(badInput))
			scanner.Scan()
			_, _, err := readInstStatus(scanner)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

var _ = common.DpcStatus{}

func mockReadInstStatus(s *bufio.Scanner) (int, common.DpcStatus, error) {
	return 1, common.DpcStatus{}, nil
}

func testInvalidPath() {
	convey.Convey("Test invalid path", func() {
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		patches.ApplyFunc(utils.CheckPath, func(path string) (string, error) {
			return "", fmt.Errorf("invalid path")
		})

		_, err := getStatusFromFile()
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "the filePath is invalid")
	})
}

func testFileOpenFailure() {
	convey.Convey("Test file open failure", func() {
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		patches.ApplyFunc(utils.CheckPath, func(path string) (string, error) {
			return "/valid/path", nil
		})
		patches.ApplyFunc(os.Open, func(name string) (*os.File, error) {
			return nil, fmt.Errorf("open failed")
		})

		_, err := getStatusFromFile()
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "open file failed")
	})
}

func TestGetStatusFromFile(t *testing.T) {
	convey.Convey("Test getStatusFromFile", t, func() {
		testInvalidPath()
		testFileOpenFailure()
	})
}

func testReceiveStopSignal(monitor *DpcEventMonitor) {
	convey.Convey("Receive stop signal with open channel", func() {
		patches := gomonkey.NewPatches()
		defer patches.Reset()

		go func() {
			monitor.stopChan <- struct{}{}
		}()

		logInfoPatches := patches.ApplyFunc(hwlog.RunLog.Info, func(args ...interface{}) {})
		defer logInfoPatches.Reset()

		monitor.Monitoring()
	})
}

func testStopChannelClosed(monitor *DpcEventMonitor) {
	convey.Convey("Stop channel is closed", func() {
		patches := gomonkey.NewPatches()
		defer patches.Reset()

		close(monitor.stopChan)

		logErrorPatches := patches.ApplyFunc(hwlog.RunLog.Error, func(args ...interface{}) {})
		defer logErrorPatches.Reset()

		monitor.Monitoring()
	})
}

func testGetStatusFromFileError(monitor *DpcEventMonitor) {
	convey.Convey("getStatusFromFile returns error", func() {
		patches := gomonkey.NewPatches()
		defer patches.Reset()

		patches.ApplyFunc(getStatusFromFile, func() (map[int]common.DpcStatus, error) {
			return nil, fmt.Errorf("mock error")
		})

		go func() {
			monitor.Monitoring()
		}()

		time.Sleep(time.Second)
		close(monitor.stopChan)
	})
}

func testNormalFlow(monitor *DpcEventMonitor) {
	convey.Convey("Normal flow", func() {
		patches := gomonkey.NewPatches()
		defer patches.Reset()

		mockMap := map[int]common.DpcStatus{1: {}}
		patches.ApplyFunc(getStatusFromFile, func() (map[int]common.DpcStatus, error) {
			return mockMap, nil
		})

		patches.ApplyFunc(isSame, func(m map[int]common.DpcStatus) bool {
			return false
		})

		triggerUpdateCalled := false
		patches.ApplyFunc(common.TriggerUpdate, func(process string) {
			triggerUpdateCalled = true
		})

		go func() {
			monitor.Monitoring()
		}()

		time.Sleep(time.Second * int6)
		close(monitor.stopChan)

		convey.So(triggerUpdateCalled, convey.ShouldBeTrue)
	})
}

func TestDpcEventMonitor_Monitoring(t *testing.T) {
	convey.Convey("Test DpcEventMonitor.Monitoring", t, func() {
		monitor := &DpcEventMonitor{
			stopChan: make(chan struct{}),
		}

		testReceiveStopSignal(monitor)
		testStopChannelClosed(monitor)
		testGetStatusFromFileError(monitor)
		testNormalFlow(monitor)
	})
}
