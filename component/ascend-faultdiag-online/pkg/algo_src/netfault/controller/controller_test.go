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

// Package controller
package controller

import (
	"bufio"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/algo_src/netfault/controllerflags"
	"ascend-faultdiag-online/pkg/algo_src/netfault/policy"
)

const (
	count0 = 0
	count1 = 1

	startSdId = 4194304
	deviceNum = 16

	maxAge        = 7
	maxBackups    = 7
	maxLineLength = 512

	testDirPermission  = 0755
	testFilePermission = 0640

	totalRows      = 20
	dataMin        = 3000
	dataMax        = 4000
	periodTimeStep = 20
	sleepTime      = 25

	delayType      = 0
	lossRateType   = 1
	disconnectType = 2

	abnormalAvgDelay   = 15000
	abnormalLossRate   = 0.79
	abnormalDisconnect = 0.80
	covertRate         = 100
)

func TestStartController(t *testing.T) {
	convey.Convey("TestStartController", t, func() {
		convey.Convey("empty path return", func() {
			clusterPath := `/cluster`
			startController(clusterPath)
		})
		convey.Convey("path not exist return", func() {
			clusterPath := `/tmp/clusterxxx`
			startController(clusterPath)
		})
		convey.Convey("path not exist", func() {
			patch0 := gomonkey.ApplyFuncReturn(os.Stat, nil, nil)
			defer patch0.Reset()
			patch1 := gomonkey.ApplyFunc(startSuperPodsDetectionAsync, func(path string) {
				return
			})
			defer patch1.Reset()
			startController("/tmp")
		})
	})
}

func TestStopController(t *testing.T) {
	convey.Convey("TestStopController", t, func() {
		convey.Convey("no parameters", func() {
			patch0 := gomonkey.ApplyMethod(reflect.TypeOf(controllerExitCond), "Wait", func(_ *sync.Cond) {
				return
			})
			defer patch0.Reset()
			stopController()
		})
	})
}

func TestReloadController(t *testing.T) {
	convey.Convey("TestReloadController", t, func() {
		convey.Convey("patch stop", func() {
			patch0 := gomonkey.ApplyFunc(stopController, func() {
				return
			})
			defer patch0.Reset()
			reloadController("/cluster")
		})
	})
}

func createSymbolicLink(t *testing.T) (originPath, symLinkPath string) {
	var fileMode0755 os.FileMode = 0755
	tmpDir := t.TempDir()
	originalPath := filepath.Join(tmpDir, "origin")
	err := os.MkdirAll(originalPath, fileMode0755)
	assert.Nil(t, err)
	symlinkPath := filepath.Join(tmpDir, "symbolic")
	// create a symlink
	err = os.Symlink(originalPath, symlinkPath)
	assert.Nil(t, err)
	return originalPath, symlinkPath
}

func TestStart(t *testing.T) {
	callStartCount := count0
	patch := gomonkey.ApplyFunc(startController, func(path string) {
		callStartCount++
	})
	defer patch.Reset()
	convey.Convey("TestStart", t, func() {
		convey.Convey("invalid path", func() {
			_, symlinkPath := createSymbolicLink(t)
			varPatch := gomonkey.ApplyGlobalVar(&clusterLevelPath, symlinkPath)
			Start()
			convey.So(callStartCount, convey.ShouldEqual, count0)
			varPatch.Reset()
		})
		convey.Convey("invalid input", func() {
			callStartCount = count0
			Start()
			convey.So(callStartCount, convey.ShouldEqual, count1)
		})
	})
}

func TestReload(t *testing.T) {
	callStartCount := count0
	patch := gomonkey.ApplyFunc(reloadController, func(path string) {
		callStartCount++
	})
	defer patch.Reset()
	convey.Convey("TestReload", t, func() {
		convey.Convey("invalid path", func() {
			_, symlinkPath := createSymbolicLink(t)
			varPatch := gomonkey.ApplyGlobalVar(&clusterLevelPath, symlinkPath)
			Reload()
			convey.So(callStartCount, convey.ShouldEqual, count0)
			varPatch.Reset()
		})
		convey.Convey("reload", func() {
			callStartCount = count0
			Reload()
			convey.So(callStartCount, convey.ShouldEqual, count1)
		})
	})
}

func TestStop(t *testing.T) {
	convey.Convey("stop", t, func() {
		convey.Convey("invalid input", func() {
			patch := gomonkey.ApplyFunc(stopController, func() {})
			defer patch.Reset()
			Stop()
		})
	})
}

func TestRegisterDetectionCallback(t *testing.T) {
	convey.Convey("Test RegisterDetectionCallback", t, func() {
		convey.Convey("should return when input is nil", func() {
			RegisterDetectionCallback(nil)
			convey.So(callbackFunc, convey.ShouldBeNil)
		})

		convey.Convey("should set callbackFunc when input is valid", func() {
			var callCount int
			callback := func(string) {
				callCount++
			}
			RegisterDetectionCallback(callback)
			convey.So(callbackFunc, convey.ShouldNotBeNil)
			callbackFunc("test")
			convey.So(callCount, convey.ShouldEqual, count1)
		})
	})
}

func TestStartControllerA3LossDisconnect(t *testing.T) {
	// Initialize the log
	_, err := initLogAndJonDir()
	assert.NoError(t, err)

	currentDir, err := os.Getwd()
	assert.NoError(t, err)
	controllerflags.IsControllerExited.SetState(false)
	testDataDir := filepath.Join(currentDir, "data-A3")
	clusterLevelPath := filepath.Join(testDataDir, "cluster")
	hwlog.RunLog.Infof("[TEST]clusterLevelPath is: %s", clusterLevelPath)

	// Construct the input file
	superPodDir := filepath.Join(clusterLevelPath, "super-pod-0")
	err = creatConf(superPodDir)
	assert.NoError(t, err)
	err = creatSuperPodJson(superPodDir)
	assert.NoError(t, err)
	err = creatPingResult(superPodDir, disconnectType)
	assert.NoError(t, err)

	go startController(clusterLevelPath)
	time.Sleep(sleepTime * time.Second)
	// Stop the detection task, log print: "[NETFAULT ALGO]net fault detection complete!"
	controllerflags.IsControllerExited.SetState(true)
	stopController()

	// Assert that the file exists
	testFiles := []string{
		filepath.Join(superPodDir, "cathelper.conf"),
		filepath.Join(superPodDir, "network_fault.json"),
		filepath.Join(superPodDir, "ping_result_1.csv"),
		filepath.Join(superPodDir, "super-pod-0.json"),
	}
	err = checkFileExist(testFiles)
	assert.NoError(t, err)
	_, err = os.Stat(filepath.Join(superPodDir, "ping_list_1.json"))
	assert.True(t, os.IsNotExist(err)) // The ping_list file will be deleted after the detection task stops

	res, err := readRest(filepath.Join(superPodDir, "network_fault.json"))
	assert.NoError(t, err)
	assert.Equal(t, 1, len(res))
	assert.Equal(t, 1, len(res[0]))
	dataMap := res[0][0]
	assert.Equal(t, abnormalDisconnect*covertRate, dataMap["avgLossRate"])
	assert.Equal(t, float64(disconnectType), dataMap["faultType"])
	assert.Equal(t, fmt.Sprintf("%d", startSdId), dataMap["srcId"])
	assert.Equal(t, fmt.Sprintf("%d", startSdId+1), dataMap["dstId"])

	err = os.RemoveAll(testDataDir)
	assert.NoError(t, err)

}

func TestStartControllerA3LossRate(t *testing.T) {
	// Initialize the log
	_, err := initLogAndJonDir()
	assert.NoError(t, err)

	currentDir, err := os.Getwd()
	assert.NoError(t, err)
	controllerflags.IsControllerExited.SetState(false)
	testDataDir := filepath.Join(currentDir, "data-A3")
	clusterLevelPath := filepath.Join(testDataDir, "cluster")
	hwlog.RunLog.Infof("[TEST]clusterLevelPath is: %s", clusterLevelPath)

	// Construct the input file
	superPodDir := filepath.Join(clusterLevelPath, "super-pod-0")
	err = creatConf(superPodDir)
	assert.NoError(t, err)
	err = creatSuperPodJson(superPodDir)
	assert.NoError(t, err)
	err = creatPingResult(superPodDir, lossRateType)
	assert.NoError(t, err)

	go startController(clusterLevelPath)
	time.Sleep(sleepTime * time.Second)
	// Stop the detection task, log print: "[NETFAULT ALGO]net fault detection complete!"
	controllerflags.IsControllerExited.SetState(true)
	stopController()

	// Assert that the file exists
	testFiles := []string{
		filepath.Join(superPodDir, "cathelper.conf"),
		filepath.Join(superPodDir, "network_fault.json"),
		filepath.Join(superPodDir, "ping_result_1.csv"),
		filepath.Join(superPodDir, "super-pod-0.json"),
	}
	err = checkFileExist(testFiles)
	assert.NoError(t, err)
	_, err = os.Stat(filepath.Join(superPodDir, "ping_list_1.json"))
	assert.True(t, os.IsNotExist(err)) // The ping_list file will be deleted after the detection task stops

	res, err := readRest(filepath.Join(superPodDir, "network_fault.json"))
	assert.NoError(t, err)
	assert.Equal(t, 1, len(res))
	assert.Equal(t, 1, len(res[0]))
	dataMap := res[0][0]
	assert.Equal(t, abnormalLossRate*covertRate, dataMap["avgLossRate"])
	assert.Equal(t, float64(lossRateType), dataMap["faultType"])
	assert.Equal(t, fmt.Sprintf("%d", startSdId), dataMap["srcId"])
	assert.Equal(t, fmt.Sprintf("%d", startSdId+1), dataMap["dstId"])

	err = os.RemoveAll(testDataDir)
	assert.NoError(t, err)

}

func TestStartControllerA3Delay(t *testing.T) {
	// Initialize the log
	_, err := initLogAndJonDir()
	assert.NoError(t, err)

	currentDir, err := os.Getwd()
	assert.NoError(t, err)
	controllerflags.IsControllerExited.SetState(false)
	testDataDir := filepath.Join(currentDir, "data-A3")
	clusterLevelPath := filepath.Join(testDataDir, "cluster")
	hwlog.RunLog.Infof("[TEST]clusterLevelPath is: %s", clusterLevelPath)

	// Construct the input file
	superPodDir := filepath.Join(clusterLevelPath, "super-pod-0")
	err = creatConf(superPodDir)
	assert.NoError(t, err)
	err = creatSuperPodJson(superPodDir)
	assert.NoError(t, err)
	err = creatPingResult(superPodDir, delayType)
	assert.NoError(t, err)

	go startController(clusterLevelPath)
	time.Sleep(sleepTime * time.Second)
	// Stop the detection task, log print: "[NETFAULT ALGO]net fault detection complete!"
	controllerflags.IsControllerExited.SetState(true)
	stopController()

	// Assert that the file exists
	testFiles := []string{
		filepath.Join(superPodDir, "cathelper.conf"),
		filepath.Join(superPodDir, "network_fault.json"),
		filepath.Join(superPodDir, "ping_result_1.csv"),
		filepath.Join(superPodDir, "super-pod-0.json"),
	}
	err = checkFileExist(testFiles)
	assert.NoError(t, err)
	_, err = os.Stat(filepath.Join(superPodDir, "ping_list_1.json"))
	assert.True(t, os.IsNotExist(err)) // The ping_list file will be deleted after the detection task stops

	res, err := readRest(filepath.Join(superPodDir, "network_fault.json"))
	assert.NoError(t, err)
	assert.Equal(t, 1, len(res))
	assert.Equal(t, 1, len(res[0]))
	dataMap := res[0][0]
	assert.Equal(t, float64(abnormalAvgDelay), dataMap["avgDelay"])
	assert.Equal(t, float64(delayType), dataMap["faultType"])
	assert.Equal(t, fmt.Sprintf("%d", startSdId), dataMap["srcId"])
	assert.Equal(t, fmt.Sprintf("%d", startSdId+1), dataMap["dstId"])

	err = os.RemoveAll(testDataDir)
	assert.NoError(t, err)

}

func readRest(filePath string) ([][]map[string]interface{}, error) {
	var detectionRes [][]map[string]interface{}
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var jsonLines []string
	var isNextLineJSON bool

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.Contains(line, "result(") && strings.HasSuffix(line, "):") {
			isNextLineJSON = true
			continue
		}
		if !isNextLineJSON {
			continue
		}
		if line == "[]" {
			continue
		}
		jsonLines = append(jsonLines, line)
		if !strings.HasPrefix(line, "]") {
			continue
		}
		jsonStr := strings.Join(jsonLines, "")
		var results []map[string]interface{}
		if err := json.Unmarshal([]byte(jsonStr), &results); err != nil {
			return nil, err
		}
		detectionRes = append(detectionRes, results)
		jsonLines = []string{}
		isNextLineJSON = false
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return detectionRes, nil
}

func checkFileExist(filePaths []string) error {
	for _, filePath := range filePaths {
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func initLogAndJonDir() (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	jobDir := filepath.Join(currentDir, "job_dir")
	hwLogConfig := hwlog.LogConfig{
		LogFileName:   filepath.Join(jobDir, "test.log"),
		LogLevel:      0, // "Log level, -1-debug, 0-info, 1-warning, 2-error, 3-critical(default 0)"
		MaxAge:        maxAge,
		MaxBackups:    maxBackups,
		MaxLineLength: maxLineLength,
	}
	if err := hwlog.InitRunLogger(&hwLogConfig, context.TODO()); err != nil {
		return "", fmt.Errorf("hwlog init failed, error is %v", err)
	}
	return jobDir, nil
}

func creatConf(dataDir string) error {
	if err := os.MkdirAll(dataDir, testDirPermission); err != nil {
		return err
	}
	lines := []string{
		"suppressedPeriod=0",
		"networkType=1",
		"pingType=0",
		"pingTimes=5",
		"pingInterval=1",
		fmt.Sprintf("period=%d", periodTimeStep),
		"netFault=on",
	}
	content := strings.Join(lines, "\n")
	if err := os.WriteFile(filepath.Join(dataDir, "cathelper.conf"), []byte(content), testFilePermission); err != nil {
		return err
	}
	return nil
}

func creatSuperPodJson(dataDir string) error {
	nodeDeviceMap := make(map[string]*policy.NodeDevice)
	deviceMap := make(map[string]string)
	for id := 0; id < deviceNum; id++ {
		davidId := fmt.Sprintf("%d", id)
		deviceMap[davidId] = fmt.Sprintf("%d", startSdId+id)
	}
	key := "node-1"
	nodeDeviceMap[key] = &policy.NodeDevice{
		NodeName:  key,
		ServerID:  "1",
		DeviceMap: deviceMap,
	}
	info := policy.SuperPodInfo{
		Version:       "A3",
		SuperPodID:    "0",
		NodeDeviceMap: nodeDeviceMap,
	}
	jsonData, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return err
	}
	if err = os.WriteFile(filepath.Join(dataDir, "super-pod-0.json"), jsonData, testFilePermission); err != nil {
		return err
	}
	return nil
}

func creatPingResult(dataDir string, faultType int) error {
	file, err := os.Create(filepath.Join(dataDir, "ping_result_1.csv"))
	if err != nil {
		return err
	}
	defer file.Close()
	if err := file.Chmod(testFilePermission); err != nil {
		return err
	}
	writer := csv.NewWriter(file)
	defer writer.Flush()

	headers := []string{"pingTaskId", "srcType", "srcAddr", "dstType", "dstAddr", "minDelay", "maxDelay", "avgDelay",
		"minLossRate", "maxLossRate", "avgLossRate", "timestamp"}
	if err := writer.Write(headers); err != nil {
		return err
	}
	rows := make([][]string, 0, totalRows)
	// Current time plus 1 minute to ensure it is later than the detection start time
	baseTime := time.Now().Add(1 * time.Minute)
	for i := 0; i < totalRows+1; i++ {
		// Generate timestamp for current line (base time + i*20 seconds)
		currentTime := baseTime.Add(time.Duration(i) * (periodTimeStep * time.Second))
		timestampMs := currentTime.UnixMilli()
		timestampStr := fmt.Sprintf("%d", timestampMs)
		// Generate a random integer between 3000 and 4000 (inclusive)
		avgDelay := rand.Intn(dataMax-dataMin+1) + dataMin
		avgLossRate := "0.000"
		if i == totalRows {
			switch faultType {
			case delayType:
				avgDelay = abnormalAvgDelay
			case lossRateType:
				avgLossRate = fmt.Sprintf("%.3f", abnormalLossRate)
			case disconnectType:
				avgLossRate = fmt.Sprintf("%.3f", abnormalDisconnect)
			default:
				return fmt.Errorf("invalid faultType %d", faultType)
			}
		}
		row := []string{
			"0", "0", fmt.Sprintf("%d", startSdId), "0", fmt.Sprintf("%d", startSdId+1),
			"3100", "6820", fmt.Sprintf("%d", avgDelay),
			avgLossRate, avgLossRate, avgLossRate,
			timestampStr,
		}
		rows = append(rows, row)
	}
	return writer.WriteAll(rows)
}
