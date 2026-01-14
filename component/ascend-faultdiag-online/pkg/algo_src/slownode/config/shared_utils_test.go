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

// Package config is used for file reading and writing, as well as data processing.
package config

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"

	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	"ascend-faultdiag-online/pkg/algo_src/slownode/jobdetectionmanager"
)

const (
	testDir       = "testdata"
	testFileName  = "testfile.txt"
	logLineLength = 256
)

func init() {
	config := hwlog.LogConfig{
		OnlyToStdout:  true,
		MaxLineLength: logLineLength,
	}
	err := hwlog.InitRunLogger(&config, context.TODO())
	if err != nil {
		fmt.Println(err)
	}
}

func testSymlinkDetection(testFilePath string) {
	convey.Convey("test symlink", func() {
		symlinkPath := filepath.Join(testDir, "symlink.txt")
		convey.So(os.Symlink(testFilePath, symlinkPath), convey.ShouldBeNil)
		defer os.Remove(symlinkPath)

		convey.Convey("fileOrDir=false, got result: true", func() {
			result := CheckExistDirectoryOrFile(symlinkPath, false, "cluster", "testJob")
			convey.So(result, convey.ShouldBeFalse)
		})
	})
}

func TestGetLocalIP(t *testing.T) {
	convey.Convey("test GetLocalIP", t, func() {
		correctIp := "127.0.0.1"
		convey.Convey("got correct ip from env", func() {
			patch := gomonkey.ApplyFuncReturn(os.Getenv, correctIp)
			defer patch.Reset()
			ip, err := GetLocalIP()
			convey.So(err, convey.ShouldBeNil)
			convey.So(correctIp, convey.ShouldEqual, ip)
		})
		// if got wrong ip from env, query the ip from interfaceAddrs
		patches := gomonkey.ApplyFuncReturn(os.Getenv, "wrongip")
		defer patches.Reset()

		convey.Convey("call net.InterfaceAddrs got error", func() {
			patch := gomonkey.ApplyFuncReturn(
				net.InterfaceAddrs,
				[]net.Addr{},
				errors.New("mock net.InterfaceAddrs failed"),
			)
			defer patch.Reset()
			ip, err := GetLocalIP()
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldEqual, "mock net.InterfaceAddrs failed")
			convey.So(ip, convey.ShouldBeEmpty)
		})
		convey.Convey("no valid ip address found in interfaceAddrs", func() {
			patch := gomonkey.ApplyFuncReturn(
				net.InterfaceAddrs,
				[]net.Addr{&net.IPNet{IP: net.ParseIP("::1")}},
				nil,
			)
			defer patch.Reset()
			ip, err := GetLocalIP()
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldEqual, "[SLOWNODE ALGO]no valid IP address found")
			convey.So(ip, convey.ShouldBeEmpty)
		})
		convey.Convey("found correct ip address in interfaceAddrs", func() {
			ip, err := GetLocalIP()
			convey.So(err, convey.ShouldBeNil)
			convey.So(ip, convey.ShouldNotBeEmpty)
		})
	})

}

func TestCheckExistCpOrEp(t *testing.T) {
	convey.Convey("test checkExistCpOrEp", t, func() {
		var rankJson, jobPath = map[string]any{}, "path"
		convey.Convey("empty rankJson", func() {
			res := checkExistCpOrEp(rankJson, jobPath)
			convey.So(res, convey.ShouldBeFalse)
		})

		convey.Convey("wrong type", func() {
			// notmap[string]any
			rankJson = map[string]any{"test": "test"}
			res := checkExistCpOrEp(rankJson, jobPath)
			convey.So(res, convey.ShouldBeTrue)
			// no field group_name
			rankJson = map[string]any{"test1": map[string]any{"test2": ""}}
			res = checkExistCpOrEp(rankJson, jobPath)
			convey.So(res, convey.ShouldBeTrue)
			// field group_name is not string
			rankJson = map[string]any{"test1": map[string]any{"group_name": struct{}{}}}
			res = checkExistCpOrEp(rankJson, jobPath)
			convey.So(res, convey.ShouldBeTrue)
			// no global_ranks
			rankJson = map[string]any{"test1": map[string]any{"group_name": "cp"}}
			res = checkExistCpOrEp(rankJson, jobPath)
			convey.So(res, convey.ShouldBeTrue)
			// global_ranks is not []any
			rankJson = map[string]any{"test1": map[string]any{
				"group_name":   "cp",
				"global_ranks": "test",
			}}
			res = checkExistCpOrEp(rankJson, jobPath)
			convey.So(res, convey.ShouldBeTrue)
		})
		convey.Convey("wrong data", func() {
			// group_name is not cp or exp
			rankJson = map[string]any{"test1": map[string]any{"group_name": "www"}}
			res := checkExistCpOrEp(rankJson, jobPath)
			convey.So(res, convey.ShouldBeFalse)
			// the length of global_ranks is greater than 1
			rankJson = map[string]any{"test1": map[string]any{
				"group_name":   "cp",
				"global_ranks": []any{"t1", "t2"},
			}}
			res = checkExistCpOrEp(rankJson, jobPath)
			convey.So(res, convey.ShouldBeTrue)
		})
	})
}

func TestCheckNodeLevelJobOpenEpOrCp(t *testing.T) {
	convey.Convey("test checkNodeLevelJobOpenEpOrCp", t, func() {
		var conf = AlgoInputConfig{}
		convey.Convey("permission & softlink check failed", func() {
			patch := gomonkey.ApplyFuncReturn(CheckFileOrDirectoryIsSoftLink, true)
			defer patch.Reset()
			convey.So(checkNodeLevelJobOpenEpOrCp(conf), convey.ShouldBeTrue)
		})
		patches := gomonkey.ApplyFuncReturn(CheckFileOrDirectoryIsSoftLink, false)
		patches.ApplyFuncReturn(CheckFileOrDirectoryReadMode, true)
		defer patches.Reset()
		convey.Convey("open file failed", func() {
			patch := gomonkey.ApplyFuncReturn(os.Open, nil, errors.New("mock open failed"))
			defer patch.Reset()
			convey.So(checkNodeLevelJobOpenEpOrCp(conf), convey.ShouldBeTrue)
		})
		patches.ApplyFuncReturn(os.Open, &os.File{}, nil)
		convey.Convey("Readdirnames", func() {
			patch := gomonkey.ApplyMethodReturn(&os.File{}, "Readdirnames", nil, errors.New("mock Readdirnames failed"))
			defer patch.Reset()
			convey.So(checkNodeLevelJobOpenEpOrCp(conf), convey.ShouldBeTrue)
		})
		patches.ApplyMethodReturn(&os.File{}, "Close", nil)
		convey.Convey("no rank dir found", func() {
			patch := gomonkey.ApplyMethodReturn(&os.File{}, "Readdirnames", nil, nil)
			defer patch.Reset()
			convey.So(checkNodeLevelJobOpenEpOrCp(conf), convey.ShouldBeTrue)
		})
		patches.ApplyMethodReturn(&os.File{}, "Readdirnames", []string{"0", "1"}, nil)
		convey.Convey("load file failed", func() {
			patch := gomonkey.ApplyFuncReturn(utils.LoadFile, []byte{}, errors.New("mock LoadFile failed"))
			defer patch.Reset()
			convey.So(checkNodeLevelJobOpenEpOrCp(conf), convey.ShouldBeTrue)
		})
		patches.ApplyFuncReturn(utils.LoadFile, []byte{}, nil)
		convey.Convey("json unmarsha failed", func() {
			patch := gomonkey.ApplyFuncReturn(json.Unmarshal, errors.New("mock json unmarshal failed"))
			defer patch.Reset()
			convey.So(checkNodeLevelJobOpenEpOrCp(conf), convey.ShouldBeTrue)
		})
		patches.ApplyFuncReturn(json.Unmarshal, nil).ApplyFuncReturn(checkExistCpOrEp, false)
		convey.Convey("normal", func() {
			convey.So(checkNodeLevelJobOpenEpOrCp(conf), convey.ShouldBeFalse)
		})
	})
}

func TestCheckCurJobOpenCpOrEp(t *testing.T) {
	convey.Convey("test CheckCurJobOpenCpOrEp", t, func() {
		var conf = AlgoInputConfig{}
		convey.Convey("CheckExistDirectoryOrFile is false", func() {
			patch := gomonkey.ApplyFuncReturn(CheckExistDirectoryOrFile, false)
			defer patch.Reset()
			ans := CheckCurJobOpenCpOrEp(conf, "cluster")
			convey.So(ans, convey.ShouldBeTrue)
		})
		convey.Convey("normal", func() {
			patch := gomonkey.ApplyFuncReturn(CheckExistDirectoryOrFile, true)
			patch.ApplyFuncReturn(checkNodeLevelJobOpenEpOrCp, true)
			defer patch.Reset()
			ans := CheckCurJobOpenCpOrEp(conf, "cluster")
			convey.So(ans, convey.ShouldBeTrue)
		})
	})
}

func TestTransferFloatArrayToInt(t *testing.T) {
	convey.Convey("test TransferFloatArrayToInt", t, func() {
		var npuIds []any
		convey.Convey("npuIds is empty", func() {
			ans := TransferFloatArrayToInt(npuIds)
			convey.So(ans, convey.ShouldBeNil)
		})
		convey.Convey("elment is not type of float64", func() {
			npuIds = []any{"data"}
			ans := TransferFloatArrayToInt(npuIds)
			convey.So(ans, convey.ShouldBeNil)
		})
		convey.Convey("normal", func() {
			npuIds = []any{0.}
			ans := TransferFloatArrayToInt(npuIds)
			convey.So(ans, convey.ShouldNotBeNil)
		})
	})
}

func TestLoopDetectionIntervalCheckSwitch(t *testing.T) {
	convey.Convey("test LoopDetectionIntervalCheckSwitch", t, func() {
		var detectionUsed int64 = 1
		var detectionInterval = 10
		var jobName, level string = "jobName", "cluster"
		convey.Convey("detectionUsed is greater than detectionInterval", func() {
			LoopDetectionIntervalCheckSwitch(int64(detectionInterval), int(detectionUsed), jobName, level)
		})

		patch := gomonkey.ApplyFuncReturn(jobdetectionmanager.GetDetectionLoopStatusClusterLevel, false)
		defer patch.Reset()
		LoopDetectionIntervalCheckSwitch(detectionUsed, detectionInterval, jobName, level)
		LoopDetectionIntervalCheckSwitch(detectionUsed, detectionInterval, jobName, level)
	})
}

func TestCheckExistDirectoryOrFile(t *testing.T) {
	var fileMode0644 os.FileMode = 0644
	var fileMode0755 os.FileMode = 0755
	// create test dir and file
	err := os.MkdirAll(testDir, fileMode0755)
	assert.Nil(t, err)
	testFilePath := filepath.Join(testDir, testFileName)
	err = os.WriteFile(testFilePath, []byte("test"), fileMode0644)
	assert.Nil(t, err)
	defer os.RemoveAll(testDir)

	patches := gomonkey.ApplyFunc(jobdetectionmanager.GetDetectionLoopStatusClusterLevel, func(string) bool { return true })
	defer patches.Reset()
	patches.ApplyFunc(jobdetectionmanager.GetDetectionLoopStatusNodeLevel, func(string) bool { return true })

	convey.Convey("test CheckExistDirectoryOrFile func", t, func() {
		testFileExists(testFilePath)
		testDirectoryExists(testDir)
		testSymlinkDetection(testFilePath)
	})
}

func testFileExists(testFilePath string) {
	convey.Convey("file exists", func() {
		convey.Convey("fileOrDir=false, got result: true", func() {
			result := CheckExistDirectoryOrFile(testFilePath, false, "cluster", "testJob")
			convey.So(result, convey.ShouldBeTrue)
		})

		convey.Convey("fileOrDir=true got result: false", func() {
			result := CheckExistDirectoryOrFile(testFilePath, true, "cluster", "testJob")
			convey.So(result, convey.ShouldBeFalse)
		})
	})
}

func testDirectoryExists(testDir string) {
	convey.Convey("dir exists", func() {
		convey.Convey("fileOrDir=true, got result: true", func() {
			result := CheckExistDirectoryOrFile(testDir, true, "node", "testJob")
			convey.So(result, convey.ShouldBeTrue)
		})

		convey.Convey("fileOrDir=false, got result: false", func() {
			result := CheckExistDirectoryOrFile(testDir, false, "node", "testJob")
			convey.So(result, convey.ShouldBeFalse)
		})
	})
}
