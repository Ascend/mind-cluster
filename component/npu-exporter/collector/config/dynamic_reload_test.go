/* Copyright(C) 2026. Huawei Technologies Co., Ltd. All rights reserved.
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

// Package config for general collector
package config

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/fsnotify/fsnotify"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"

	"ascend-common/common-utils/hwlog"
	"huawei.com/npu-exporter/v6/utils/logger"
)

func init() {
	logger.HwLogConfig = &hwlog.LogConfig{
		OnlyToStdout: true,
	}
	logger.InitLogger("Prometheus")
}

const (
	testConfigDir   = "/user/mind-cluster/npu-exporter-config"
	testFileName    = "metricConfiguration.json"
	testPluginFile  = "pluginConfiguration.json"
	testDataSymlink = "..data"
)

// isRelevantEventCase is one table entry for testing isRelevantEvent.
type isRelevantEventCase struct {
	name      string
	eventName string
	eventOp   fsnotify.Op
	expect    bool
}

func TestIsRelevantEvent(t *testing.T) {
	cases := []isRelevantEventCase{
		{name: "should return true when metric config file is written",
			eventName: testFileName, eventOp: fsnotify.Write, expect: true},
		{name: "should return true when metric config file is created",
			eventName: testFileName, eventOp: fsnotify.Create, expect: true},
		{name: "should return true when metric config file is renamed",
			eventName: testFileName, eventOp: fsnotify.Rename, expect: true},
		{name: "should return true when plugin config file is written",
			eventName: testPluginFile, eventOp: fsnotify.Write, expect: true},
		{name: "should return true when plugin config file is created",
			eventName: testPluginFile, eventOp: fsnotify.Create, expect: true},
		{name: "should return false when metric config file is removed",
			eventName: testFileName, eventOp: fsnotify.Remove, expect: false},
		{name: "should return false when metric config file is chmod",
			eventName: testFileName, eventOp: fsnotify.Chmod, expect: false},
	}
	runIsRelevantEventCases(t, cases)
}

func TestIsRelevantEventForK8sConfigMap(t *testing.T) {
	cases := []isRelevantEventCase{
		{name: "should return true when ..data symlink is renamed",
			eventName: testDataSymlink, eventOp: fsnotify.Rename, expect: true},
		{name: "should return true when ..data symlink is created",
			eventName: testDataSymlink, eventOp: fsnotify.Create, expect: true},
		{name: "should return false when ..data symlink is written",
			eventName: testDataSymlink, eventOp: fsnotify.Write, expect: false},
		{name: "should return false when ..data symlink is removed",
			eventName: testDataSymlink, eventOp: fsnotify.Remove, expect: false},
		{name: "should return false when ..data symlink is chmod",
			eventName: testDataSymlink, eventOp: fsnotify.Chmod, expect: false},
		{name: "should return false when timestamp directory is created",
			eventName: "..2026_06_09_08_30_00", eventOp: fsnotify.Create, expect: false},
		{name: "should return false when timestamp directory with double dot is created",
			eventName: "..2026_06_11_16_35_44.3381415300", eventOp: fsnotify.Create, expect: false},
	}
	runIsRelevantEventCases(t, cases)
}

func TestIsRelevantEventForIrrelevantEvents(t *testing.T) {
	cases := []isRelevantEventCase{
		{name: "should return false when event is in different directory",
			eventName: "/other/path/" + testFileName, eventOp: fsnotify.Write, expect: false},
		{name: "should return false when file name is unknown",
			eventName: "unknown.txt", eventOp: fsnotify.Write, expect: false},
	}
	runIsRelevantEventCases(t, cases)
}

// runIsRelevantEventCases executes table-driven test cases for isRelevantEvent.
func runIsRelevantEventCases(t *testing.T, cases []isRelevantEventCase) {
	for _, tc := range cases {
		convey.Convey(tc.name, t, func() {
			ev := fsnotify.Event{
				Name: filepath.Join(testConfigDir, tc.eventName),
				Op:   tc.eventOp,
			}
			convey.So(isRelevantEvent(ev, testConfigDir), convey.ShouldEqual, tc.expect)
		})
	}
}

func TestTimerCh(t *testing.T) {
	convey.Convey("should return nil when timer is nil", t, func() {
		ch := timerCh(nil)
		convey.So(ch, convey.ShouldBeNil)
	})

	convey.Convey("should return timer channel when timer is not nil", t, func() {
		timer := time.NewTimer(time.Hour)
		defer timer.Stop()
		ch := timerCh(timer)
		convey.So(ch, convey.ShouldEqual, timer.C)
	})
}

func TestHandleFsEvent(t *testing.T) {
	cases := []struct {
		name           string
		eventName      string
		eventOp        fsnotify.Op
		existingTimer  bool
		expectTimerNil bool
	}{
		{name: "should create timer when event is relevant and no timer exists",
			eventName:      testFileName,
			eventOp:        fsnotify.Write,
			existingTimer:  false,
			expectTimerNil: false,
		},
		{name: "should not create timer when event is irrelevant",
			eventName:      "unknown.txt",
			eventOp:        fsnotify.Write,
			existingTimer:  false,
			expectTimerNil: true,
		},
		{name: "should not create new timer when timer already exists",
			eventName:      testFileName,
			eventOp:        fsnotify.Write,
			existingTimer:  true,
			expectTimerNil: false,
		},
	}
	for _, tc := range cases {
		convey.Convey(tc.name, t, func() {
			var reloadTimer *time.Timer
			if tc.existingTimer {
				reloadTimer = time.NewTimer(time.Hour)
			}
			ev := fsnotify.Event{
				Name: filepath.Join(testConfigDir, tc.eventName),
				Op:   tc.eventOp,
			}
			handleFsEvent(ev, testConfigDir, &reloadTimer)
			if tc.expectTimerNil {
				convey.So(reloadTimer, convey.ShouldBeNil)
			} else {
				convey.So(reloadTimer, convey.ShouldNotBeNil)
			}
			if reloadTimer != nil {
				reloadTimer.Stop()
			}
		})
	}
}

func TestStartDynamicReload(t *testing.T) {
	convey.Convey("should return early when fsnotify watcher creation fails", t, func() {
		patches := gomonkey.ApplyFuncReturn(fsnotify.NewWatcher, nil, assert.AnError)
		defer patches.Reset()
		StartDynamicReload(context.Background(), nil)
	})
}

func TestIsRelevantEventWhenBaseNamePatterns(t *testing.T) {
	testCases := []struct {
		name           string
		baseName       string
		eventOp        fsnotify.Op
		expectedResult bool
	}{
		{name: "should return false for dotfile starting with single dot",
			baseName:       ".hidden",
			eventOp:        fsnotify.Create,
			expectedResult: false,
		},
		{name: "should return false for double dot directory",
			baseName:       "..",
			eventOp:        fsnotify.Create,
			expectedResult: false,
		},
		{name: "should return true for ..data when operation is create",
			baseName:       "..data",
			eventOp:        fsnotify.Create,
			expectedResult: true,
		},
		{name: "should return true for ..data when operation is rename",
			baseName:       "..data",
			eventOp:        fsnotify.Rename,
			expectedResult: true,
		},
		{name: "should return false for ..data when operation is write",
			baseName:       "..data",
			eventOp:        fsnotify.Write,
			expectedResult: false,
		},
	}

	for _, tc := range testCases {
		convey.Convey(tc.name, t, func() {
			ev := fsnotify.Event{
				Name: filepath.Join(testConfigDir, tc.baseName),
				Op:   tc.eventOp,
			}
			result := isRelevantEvent(ev, testConfigDir)
			convey.So(result, convey.ShouldEqual, tc.expectedResult)
		})
	}
}
