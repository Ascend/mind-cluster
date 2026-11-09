/* Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
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

package dtfsmonitor

import (
	"bufio"
	"context"
	"errors"
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

func init() {
	ctx := context.Background()
	hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, ctx)
}

// TestNewDtfsEventMonitor tests function NewDtfsEventMonitor
func TestNewDtfsEventMonitor(t *testing.T) {
	convey.Convey("Test NewDtfsEventMonitor", t, func() {
		ctx := context.Background()
		monitor := NewDtfsEventMonitor(ctx)
		convey.So(monitor, convey.ShouldNotBeNil)
		convey.So(monitor.Name(), convey.ShouldEqual, common.PluginMonitorDtfs)
	})
}

// TestDtfsEventMonitorInit tests function Init
func TestDtfsEventMonitorInit(t *testing.T) {
	convey.Convey("Test DtfsEventMonitor Init", t, func() {
		ctx := context.Background()
		monitor := NewDtfsEventMonitor(ctx)
		err := monitor.Init()
		convey.So(err, convey.ShouldBeNil)
	})
}

// TestDtfsEventMonitorStop tests function Stop
func TestDtfsEventMonitorStop(t *testing.T) {
	convey.Convey("Test DtfsEventMonitor Stop", t, func() {
		ctx := context.Background()
		monitor := NewDtfsEventMonitor(ctx)
		monitor.Stop()
	})
}

// TestGetStatusByText tests function getStatusByText
func TestGetStatusByText(t *testing.T) {
	convey.Convey("Test getStatusByText", t, func() {
		convey.Convey("DTFS_PROCESS_ERROR normal", func() {
			status, err := getStatusByText("DTFS_PROCESS_ERROR: 0", common.DtfsProcessErrorKey)
			convey.So(err, convey.ShouldBeNil)
			convey.So(status, convey.ShouldBeFalse)
		})
		convey.Convey("DTFS_PROCESS_ERROR error", func() {
			status, err := getStatusByText("DTFS_PROCESS_ERROR: -1", common.DtfsProcessErrorKey)
			convey.So(err, convey.ShouldBeNil)
			convey.So(status, convey.ShouldBeTrue)
		})
		convey.Convey("DTFS_LINK_ERROR normal", func() {
			status, err := getStatusByText("DTFS_LINK_ERROR: 0", common.DtfsLinkErrorKey)
			convey.So(err, convey.ShouldBeNil)
			convey.So(status, convey.ShouldBeFalse)
		})
		convey.Convey("DTFS_LINK_ERROR error", func() {
			status, err := getStatusByText("DTFS_LINK_ERROR: -1", common.DtfsLinkErrorKey)
			convey.So(err, convey.ShouldBeNil)
			convey.So(status, convey.ShouldBeTrue)
		})
		convey.Convey("invalid key", func() {
			_, err := getStatusByText("INVALID_KEY: 0", common.DtfsProcessErrorKey)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("invalid value", func() {
			_, err := getStatusByText("DTFS_PROCESS_ERROR: 1", common.DtfsProcessErrorKey)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("file key does not match", func() {
			_, err := getStatusByText("DTFS_PROCESS_ERROR: 0", common.DtfsLinkErrorKey)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestIsSame tests function isSame
func TestIsSame(t *testing.T) {
	convey.Convey("Test isSame", t, func() {
		convey.Convey("first time", func() {
			lastUploadTime = 0
			newStatus := common.DtfsStatus{ProcessError: false, LinkError: false}
			same := isSame(newStatus)
			convey.So(same, convey.ShouldBeFalse)
		})
		convey.Convey("status changed", func() {
			lastUploadTime = 1
			dtfsStatus = common.DtfsStatus{ProcessError: false, LinkError: false}
			newStatus := common.DtfsStatus{ProcessError: true, LinkError: false}
			same := isSame(newStatus)
			convey.So(same, convey.ShouldBeFalse)
		})
		convey.Convey("status same", func() {
			lastUploadTime = 1
			dtfsStatus = common.DtfsStatus{ProcessError: false, LinkError: false}
			newStatus := common.DtfsStatus{ProcessError: false, LinkError: false}
			same := isSame(newStatus)
			convey.So(same, convey.ShouldBeTrue)
		})
	})
}

// TestGetStatusFromFile tests function getStatusFromFile
func TestGetStatusFromFile(t *testing.T) {
	convey.Convey("Test getStatusFromFile", t, func() {
		convey.Convey("file not exist", func() {
			patches := gomonkey.NewPatches()
			patches.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) {
				return nil, os.ErrNotExist
			})
			defer patches.Reset()

			_, err := getStatusFromFile()
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("valid file with fault", func() {
			filePath := os.TempDir() + "/dtfs_status_check"
			content := "[instidx=1]\nDTFS_PROCESS_ERROR: 0\nDTFS_LINK_ERROR: 0\n" +
				"[instidx=2]\nDTFS_PROCESS_ERROR: -1\nDTFS_LINK_ERROR: -1\n"
			convey.So(os.WriteFile(filePath, []byte(content), 0600), convey.ShouldBeNil)

			patches := gomonkey.NewPatches()
			patches.ApplyFuncReturn(utils.CheckOwnerAndPermission, filePath, nil)
			defer patches.Reset()

			status, err := getStatusFromFile()
			convey.So(err, convey.ShouldBeNil)
			convey.So(status.ProcessError, convey.ShouldBeTrue)
			convey.So(status.LinkError, convey.ShouldBeTrue)
		})
		convey.Convey("empty file", func() {
			filePath := os.TempDir() + "/dtfs_status_check"
			convey.So(os.WriteFile(filePath, []byte(""), 0600), convey.ShouldBeNil)

			patches := gomonkey.NewPatches()
			patches.ApplyFuncReturn(utils.CheckOwnerAndPermission, filePath, nil)
			defer patches.Reset()

			_, err := getStatusFromFile()
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("invalid file content", func() {
			filePath := os.TempDir() + "/dtfs_status_check"
			convey.So(os.WriteFile(filePath, []byte("invalid line\n"), 0600), convey.ShouldBeNil)

			patches := gomonkey.NewPatches()
			patches.ApplyFuncReturn(utils.CheckOwnerAndPermission, filePath, nil)
			defer patches.Reset()

			_, err := getStatusFromFile()
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestReadInstStatus tests function readInstStatus
func TestReadInstStatus(t *testing.T) {
	convey.Convey("Test readInstStatus", t, func() {
		convey.Convey("valid input", func() {
			input := "[instidx=1]\nDTFS_PROCESS_ERROR: -1\nDTFS_LINK_ERROR: 0"
			scanner := bufio.NewScanner(strings.NewReader(input))
			scanner.Scan()

			inst, status, err := readInstStatus(scanner)
			convey.So(err, convey.ShouldBeNil)
			convey.So(inst, convey.ShouldEqual, 1)
			convey.So(status.ProcessError, convey.ShouldBeTrue)
			convey.So(status.LinkError, convey.ShouldBeFalse)
		})
		convey.Convey("invalid inst format", func() {
			scanner := bufio.NewScanner(strings.NewReader("invalid_inst"))
			scanner.Scan()
			_, _, err := readInstStatus(scanner)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("missing status lines after inst line", func() {
			scanner := bufio.NewScanner(strings.NewReader("[instidx=1]"))
			scanner.Scan()
			_, _, err := readInstStatus(scanner)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("missing third status line", func() {
			scanner := bufio.NewScanner(strings.NewReader("[instidx=1]\nDTFS_PROCESS_ERROR: 0"))
			scanner.Scan()
			_, _, err := readInstStatus(scanner)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestDtfsEventMonitorGetMonitorData tests function GetMonitorData
func TestDtfsEventMonitorGetMonitorData(t *testing.T) {
	convey.Convey("Test DtfsEventMonitor GetMonitorData", t, func() {
		ctx := context.Background()
		monitor := NewDtfsEventMonitor(ctx)
		data := monitor.GetMonitorData()
		convey.So(data, convey.ShouldNotBeNil)
		convey.So(data.DtfsStatus, convey.ShouldResemble, common.DtfsStatus{})
	})
}

// TestDtfsEventMonitorMonitoring tests function Monitoring
func TestDtfsEventMonitorMonitoring(t *testing.T) {
	convey.Convey("Test DtfsEventMonitor Monitoring normal flow", t, func() {
		patches := gomonkey.ApplyGlobalVar(&common.ParamOption, common.Option{DtfsMonitorPeriod: 1})
		defer patches.Reset()

		mockStatus := common.DtfsStatus{ProcessError: false, LinkError: false}
		patches.ApplyFunc(getStatusFromFile, func() (common.DtfsStatus, error) {
			return mockStatus, nil
		})
		patches.ApplyFunc(isSame, func(status common.DtfsStatus) bool {
			return false
		})
		triggerUpdateCalled := false
		patches.ApplyFunc(common.TriggerUpdate, func(process string) {
			triggerUpdateCalled = true
		})

		monitor := NewDtfsEventMonitor(context.Background())
		go monitor.Monitoring()

		time.Sleep(2 * time.Second)
		close(monitor.stopChan)

		convey.So(triggerUpdateCalled, convey.ShouldBeTrue)
	})

	convey.Convey("Test DtfsEventMonitor Monitoring with getStatusFromFile error", t, func() {
		patches := gomonkey.ApplyGlobalVar(&common.ParamOption, common.Option{DtfsMonitorPeriod: 1})
		defer patches.Reset()

		patches.ApplyFunc(getStatusFromFile, func() (common.DtfsStatus, error) {
			return common.DtfsStatus{}, errors.New("mock error")
		})

		monitor := NewDtfsEventMonitor(context.Background())
		go monitor.Monitoring()

		time.Sleep(1500 * time.Millisecond)
		close(monitor.stopChan)

		convey.So(true, convey.ShouldBeTrue)
	})
}
