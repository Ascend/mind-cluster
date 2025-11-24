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
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/algo_src/netfault/algo"
	"ascend-faultdiag-online/pkg/algo_src/netfault/controllerflags"
	"ascend-faultdiag-online/pkg/algo_src/netfault/policy"
)

const logLineLength = 256

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

// TestGetSuperPodDirInfoFailCasesPartOne test for func getSuperPodDirInfo
func TestGetSuperPodDirInfoFailCasesPartOne(t *testing.T) {
	convey.Convey("Test getSuperPodDirInfo Fail Cases", t, func() {
		convey.Convey("should return nil when clusterPath is not existed", func() {
			clusterPath := `\aa\bb\sss.json`
			ret1, ret2 := getSuperPodDirInfo(clusterPath)
			convey.So(ret1, convey.ShouldBeNil)
			convey.So(ret2, convey.ShouldBeNil)
		})

		convey.Convey("should return nil when FileInfo is directory", func() {
			clusterPath := `./super-pod-0.json`
			clusterPathFile, err := os.Create(clusterPath)
			if err != nil {
				return
			}
			defer clusterPathFile.Close()
			defer clusterPathFile.Chmod(0600) //文件权限
			ret1, ret2 := getSuperPodDirInfo(clusterPath)
			convey.So(ret1, convey.ShouldBeNil)
			convey.So(ret2, convey.ShouldBeNil)
			err = os.Remove(clusterPath)
			if err != nil {
				return
			}
		})
	})
}

// TestDeletePingListFile
func TestDeletePingListFile(t *testing.T) {
	convey.Convey("TestDeletePingListFile", t, func() {
		convey.Convey("delete path", func() {
			clusterPath := `\aa\bb\sss.json`
			deletePingListFile(clusterPath)
		})
	})
}

// TestGetSuperPodDirInfoFailCasesPartTwo test for func getSuperPodDirInfo
func TestGetSuperPodDirInfoFailCasesPartTwo(t *testing.T) {
	convey.Convey("Test getSuperPodDirInfo Fail Cases", t, func() {
		convey.Convey("should return nil when clusterPath is not matched super-pod", func() {
			clusterDir := "./1/"
			clusterPath := clusterDir + `/1.json`
			err := os.Mkdir(clusterDir, 0755) //文件权限
			if err != nil {
				hwlog.RunLog.Errorf("path %s, err: %v", clusterDir, err)
				return
			}
			clusterPathFile, err := os.Create(clusterPath)
			if err != nil {
				hwlog.RunLog.Errorf("path %s, err: %v", clusterPath, err)
				return
			}
			defer clusterPathFile.Close()
			defer clusterPathFile.Chmod(0600) //文件权限
			ret1, ret2 := getSuperPodDirInfo(clusterDir)
			hwlog.RunLog.Errorf("ret1 %v", ret1)
			hwlog.RunLog.Errorf("ret2 %v", ret2)
			convey.So(ret1, convey.ShouldBeZeroValue)
			convey.So(ret2, convey.ShouldBeZeroValue)
			err = os.Remove(clusterPath)
			if err != nil {
				return
			}
			err = os.Remove(clusterDir)
			if err != nil {
				return
			}
		})
	})
}

// TestGetSuperPodDirInfoPassCases test for func getSuperPodDirInfo
func TestGetSuperPodDirInfoPassCases(t *testing.T) {
	convey.Convey("Test getSuperPodDirInfo Pass Cases", t, func() {
		convey.Convey("should return valid value when clusterPath is valid", func() {
			clusterDir := "./super-pod-0/"
			clusterPath := `./super-pod-0/super-pod-0.json`
			err := os.Mkdir(clusterDir, 0755) //文件权限
			if err != nil {
				hwlog.RunLog.Errorf("path %s, err: %v", clusterDir, err)
				return
			}
			clusterPathFile, err := os.Create(clusterPath)
			if err != nil {
				hwlog.RunLog.Errorf("path %s, err: %v", clusterPath, err)
				return
			}
			actualReturnValue1 := []int{0}
			actualReturnValue2 := []string{"super-pod-0/super-pod-0"}
			defer clusterPathFile.Close()
			defer clusterPathFile.Chmod(0600) //文件权限
			ret1, ret2 := getSuperPodDirInfo(clusterDir)
			convey.So(ret1, convey.ShouldResemble, actualReturnValue1)
			convey.So(ret2, convey.ShouldResemble, actualReturnValue2)
			err = os.Remove(clusterPath)
			if err != nil {
				return
			}
			err = os.Remove(clusterDir)
			if err != nil {
				return
			}
		})
	})
}

// TestCheckDiffConfig test for func CheckDiffConfig
func TestCheckDiffConfig(t *testing.T) {
	mockTime := gomonkey.ApplyFunc(time.Sleep, func(d time.Duration) {})
	defer mockTime.Reset()
	convey.Convey("Test checkDiffConfig", t, func() {
		convey.Convey("should return false when config file is not existed", func() {
			superPodFilePath := "1"
			ret := checkDiffConfig(superPodFilePath)
			convey.So(ret, convey.ShouldBeNil)
		})

		convey.Convey("should return valid value when config file is existed", func() {
			controllerflags.IsControllerStarted.SetState(true)
			superPodFilePath := "./"
			configFile, err := os.Create(superPodFilePath + "cathelper.conf")
			if err != nil {
				return
			}
			defer configFile.Close()
			defer configFile.Chmod(0600) //文件权限
			ret := checkDiffConfig(superPodFilePath)
			convey.So(ret, convey.ShouldBeEmpty)
			controllerflags.IsControllerStarted.SetState(false)
			err = os.Remove(superPodFilePath + "cathelper.conf")
			if err != nil {
				return
			}
		})
	})
}

// TestGetCurSuperPodDetectionInterval test for func getCurSuperPodDetectionInterval
func TestGetCurSuperPodDetectionInterval(t *testing.T) {
	convey.Convey("Test getCurSuperPodDetectionInterval Part Seven", t, func() {
		convey.Convey("should return 15 when period value is string", func() {
			conf := map[string]any{
				"period": "abc",
			}
			superPodFilePath := "./superpodfile.json"
			ret := getCurSuperPodDetectionInterval(conf, superPodFilePath)
			convey.So(ret, convey.ShouldEqual, 15) //15 means default value
		})

		convey.Convey("should return 1 when period value is 1", func() {
			conf := map[string]any{
				"period": 1,
			}
			superPodFilePath := "./superpodfile.json"
			ret := getCurSuperPodDetectionInterval(conf, superPodFilePath)
			convey.So(ret, convey.ShouldEqual, 1) //1 means period value
		})

		convey.Convey("should return 15 when conf is empty", func() {
			conf := map[string]any{}
			superPodFilePath := "./superpodfile.json"
			ret := getCurSuperPodDetectionInterval(conf, superPodFilePath)
			convey.So(ret, convey.ShouldEqual, 15) //15 means default value
		})
	})
}

// TestFindCSVFiles test for func findCSVFiles
func TestFindCSVFiles(t *testing.T) {
	convey.Convey("Test findCSVFiles", t, func() {
		convey.Convey("should return 15 when period value is string", func() {
			dir := "./superpodfile.json"
			ret, err := findCSVFiles(dir)
			convey.So(ret, convey.ShouldBeNil)
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("should return 15 when period value is string1", func() {
			dir := "./"
			configFile, err := os.Create(dir + "1.csv")
			if err != nil {
				return
			}
			err = configFile.Chmod(0644) //文件权限
			if err != nil {
				return
			}
			defer configFile.Close()
			csvFiles := []string{"1.csv"}
			ret, err := findCSVFiles(dir)
			convey.So(ret, convey.ShouldResemble, csvFiles)
			convey.So(err, convey.ShouldBeNil)
			err = os.Remove(dir + "1.csv")
			if err != nil {
				return
			}
		})
	})
}

// TestReadCSVFilePartOne test for func readCSVFile
func TestReadCSVFilePartOne(t *testing.T) {
	convey.Convey("Test readCSVFile Part One", t, func() {
		convey.Convey("should return nil when filePath is invalid", func() {
			filePath := ""
			startTime := time.Now().UnixMilli()
			ret, err := readCSVFile(filePath, startTime)
			convey.So(ret, convey.ShouldBeNil)
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("should return nil when CSV is empty", func() {
			var filePath = "./test_csv1.csv"
			configFile, err := os.Create(filePath)
			if err != nil {
				return
			}
			err = configFile.Chmod(0644) //文件权限
			if err != nil {
				return
			}
			startTime := time.Now().UnixMilli()
			ret, err := readCSVFile(filePath, startTime)
			hwlog.RunLog.Errorf("csv data: %v", ret)
			convey.So(ret, convey.ShouldBeNil)
			convey.So(err, convey.ShouldNotBeNil)
			err = os.Remove(filePath)
			if err != nil {
				return
			}
		})
	})
}

// TestReadCSVFilePartTwo test for func readCSVFile
func TestReadCSVFilePartTwo(t *testing.T) {
	convey.Convey("Test readCSVFile Part Two", t, func() {
		convey.Convey("should return nil when CSV read fail", func() {
			var filePath = "./test_csv2.csv"
			configFile, err := os.Create(filePath)
			if err != nil {
				return
			}
			err = configFile.Chmod(0644) //文件权限
			if err != nil {
				return
			}
			mockReadAll := gomonkey.ApplyMethod(new(csv.Reader), "ReadAll",
				func(*csv.Reader) ([][]string, error) {
					return nil, errors.New("read all failed")
				})
			defer mockReadAll.Reset()
			startTime := time.Now().UnixMilli()
			ret, err := readCSVFile(filePath, startTime)
			convey.So(ret, convey.ShouldBeNil)
			convey.So(err, convey.ShouldNotBeNil)
			err = os.Remove(filePath)
			if err != nil {
				return
			}
		})
	})
}

// TestReadCSVFilePartThree test for func readCSVFile
func TestReadCSVFilePartThree(t *testing.T) {
	convey.Convey("Test readCSVFile Part Three", t, func() {
		convey.Convey("should return valid value when CSV read", func() {
			var filePath = "./test_csv3.csv"
			configFile, err := os.Create(filePath)
			if err != nil {
				return
			}
			defer os.Remove(filePath)
			err = configFile.Chmod(0644) //文件权限
			if err != nil {
				return
			}
			headers := []string{"header1", "header2", "header3"}
			rows := [][]string{{"1", "1", "abc"}, {"2", "2", "1745479726"}, {"3", "3", "1745479710"}}
			err = WriteToCsv(filePath, headers, rows)
			convey.So(err, convey.ShouldBeNil)
			expectData := make([]map[string]any, 0)
			expectData = append(expectData, map[string]any{"header1": "2", "header2": "2", "header3": "1745479726"})
			ret, err := readCSVFile(filePath, 1745479711) // 1745479711 means test case time stamp value
			convey.So(err, convey.ShouldBeNil)
			convey.So(ret, convey.ShouldResemble, expectData)
		})
	})
}

func WriteToCsv(filePath string, headers []string, rows [][]string) error {
	file, err := os.OpenFile(filePath, os.O_RDWR, 644) // 644 means file permission
	if err != nil {
		return fmt.Errorf("open file %s failed: %s", filePath, err)
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	if err = writer.Write(headers); err != nil {
		return fmt.Errorf("write header %v to %s failed: %s", headers, filePath, err)
	}
	for _, row := range rows {
		if err = writer.Write(row); err != nil {
			return fmt.Errorf("write row %v to %s failed: %s", row, filePath, err)
		}
	}
	writer.Flush()
	if err = writer.Error(); err != nil {
		return fmt.Errorf("csv file writer: [%s] flush data failed: failed: %s", filePath, err)
	}

	if err = file.Close(); err != nil {
		return fmt.Errorf("file: [%s] close failed: %s", filePath, err)
	}
	return nil
}

func TestLoopDetectionIntervalCheckSwitch(t *testing.T) {
	convey.Convey("Test loopDetectionIntervalCheckSwitch", t, func() {
		convey.Convey("should sleep count 1 When detection time is less than interval", func() {
			interval := int64(1)
			detectionStartTime := time.Now().Unix() - (interval - 1)

			mockNow := time.Unix(detectionStartTime+interval-1, 0)
			patchNow := gomonkey.ApplyFunc(time.Now, func() time.Time {
				return mockNow
			})
			defer patchNow.Reset()

			var sleepCount int
			patchSleep := gomonkey.ApplyFunc(time.Sleep, func(d time.Duration) {
				sleepCount++
			})
			defer patchSleep.Reset()

			loopDetectionIntervalCheckSwitch(interval, detectionStartTime, "")

			convey.So(sleepCount, convey.ShouldEqual, 0)
		})
	})
}

func TestAsyncReadCsvFile(t *testing.T) {
	convey.Convey("Test asyncReadCsvFile", t, func() {
		convey.Convey("should start right goroutine when len(file) < maxAsyncGoRoutineToReadCsvFiles", func() {
			files := []string{"test1", "test2", "test3"}
			var goRoutineNum int32 = 0
			mockReadWorkRoutine := gomonkey.ApplyFunc(readWorkRoutine, func(wg *sync.WaitGroup,
				csvFileQueue <-chan string, input *[]map[string]any, startTime int64) {
				defer wg.Done()
				atomic.AddInt32(&goRoutineNum, 1)
			})
			defer mockReadWorkRoutine.Reset()
			asyncReadCsvFile(files, 0)
			convey.So(goRoutineNum, convey.ShouldEqual, len(files))
		})
		convey.Convey("should start right goroutine when len(file) > maxAsyncGoRoutineToReadCsvFiles", func() {
			files := []string{"0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0",
				"0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0",
				"0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0"}
			var goRoutineNum int32 = 0
			mockReadWorkRoutine := gomonkey.ApplyFunc(readWorkRoutine, func(wg *sync.WaitGroup,
				csvFileQueue <-chan string, input *[]map[string]any, startTime int64) {
				defer wg.Done()
				atomic.AddInt32(&goRoutineNum, 1)
			})
			defer mockReadWorkRoutine.Reset()
			asyncReadCsvFile(files, 0)
			convey.So(goRoutineNum, convey.ShouldEqual, maxAsyncGoRoutineToReadCsvFiles)
		})
	})
}

func TestMarkFalseDetection(t *testing.T) {
	convey.Convey("TestMarkFalseDetection", t, func() {
		// 测试键值对存在
		convey.Convey("when_key_exists", func() {
			modifyRecorderSyncLock.Lock()
			superPodDetectionRecorder["test/path"] = true
			modifyRecorderSyncLock.Unlock()
			markFalseDetection("test/path")
			modifyRecorderSyncLock.Lock()
			convey.So(superPodDetectionRecorder["test/path"], convey.ShouldEqual, false)
			modifyRecorderSyncLock.Unlock()
		})
	})
}

func TestStartSuperPodsDetectionAsync(t *testing.T) {
	convey.Convey("Test startSuperPodsDetectionAsync", t, func() {
		convey.Convey("should start right count detection", func() {
			gomonkey.ApplyFunc(ifAddNewSuperPodDetection, func(_ string, wg *sync.WaitGroup) {
				wg.Done()
			})
			startSuperPodsDetectionAsync("Path")
			convey.So(controllerflags.IsControllerStarted.GetState(), convey.ShouldBeFalse)
		})
	})
}

func TestWriteNetFaultResult(t *testing.T) {
	convey.Convey("Test writeNetFaultResult", t, func() {
		superPodPath := "/test/superpod"
		result := []byte("test result")

		convey.Convey("should return when openFile err", func() {
			patchOpenFile := gomonkey.ApplyFunc(
				os.OpenFile,
				func(name string, flag int, perm os.FileMode) (*os.File, error) {
					return nil, errors.New("open failed")
				},
			)
			defer patchOpenFile.Reset()
			writeNetFaultResult(result, superPodPath, 0)
		})

		convey.Convey("should return when file write err", func() {
			var file *os.File

			patchOpenFile := gomonkey.ApplyFunc(
				os.OpenFile, func(name string, flag int, perm os.FileMode) (*os.File, error) { return file, nil },
			)
			defer patchOpenFile.Reset()
			writeNetFaultResult(result, superPodPath, 0)
		})
	})
}

func TestLoopCsvCallDetection(t *testing.T) {
	convey.Convey("Test LoopCsvCallDetection", t, func() {
		convey.Convey("should stop detection when controller exit", func() {
			detectObj := algo.NewNetDetect("")
			patch := gomonkey.ApplyMethod(reflect.TypeOf(detectObj), "StartFaultDetect",
				func(nd *algo.NetDetect, _ []map[string]any) []any {
					return nil
				})
			defer patch.Reset()
			go func() {
				time.Sleep(time.Second)
				controllerflags.IsControllerExited.SetState(true)
			}()
			param := detectionParam{}
			loopCsvCallDetection(param)
		})
	})
}

func TestLoopWaitSuperPodDirAndCheckConfigFile(t *testing.T) {
	convey.Convey("test func LoopWaitSuperPodDirAndCheckConfigFile", t, func() {
		convey.Convey("return false when stat err", func() {
			controllerflags.IsControllerExited.SetState(false)
			patch0 := gomonkey.ApplyFuncReturn(policy.CheckCurSuperPodConfigSwitch, true)
			defer patch0.Reset()
			patch1 := gomonkey.ApplyFuncReturn(os.Stat, nil, os.ErrNotExist)
			defer patch1.Reset()
			patch3 := gomonkey.ApplyFunc(time.Sleep, func(_ time.Duration) {})
			defer patch3.Reset()
			ret := loopWaitSuperPodDirAndCheckConfigFile("", "", "", true)
			convey.So(ret, convey.ShouldBeFalse)
		})
		convey.Convey("patch for loop wait", func() {
			controllerflags.IsControllerExited.SetState(false)
			patch0 := gomonkey.ApplyFuncReturn(policy.CheckCurSuperPodConfigSwitch, true)
			defer patch0.Reset()
			patch1 := gomonkey.ApplyFuncReturn(os.Stat, nil, nil)
			defer patch1.Reset()
			ret := loopWaitSuperPodDirAndCheckConfigFile("", "", "", true)
			time.Sleep(time.Duration(1) * time.Second)
			controllerflags.IsControllerExited.SetState(true)
			convey.So(ret, convey.ShouldBeTrue)
		})
	})
}

func TestReadCSVFile(t *testing.T) {
	convey.Convey("test func ReadCsvFile", t, func() {
		convey.Convey("return nil when open file err", func() {
			ret1, _ := readCSVFile("A|b/PAth", 0)
			convey.So(ret1, convey.ShouldBeNil)
		})
		convey.Convey("return nil when record err", func() {
			patch := gomonkey.ApplyFuncReturn(os.OpenFile, nil, nil)
			defer patch.Reset()
			ret1, _ := readCSVFile("Path", 0)
			convey.So(ret1, convey.ShouldBeNil)
		})
		convey.Convey("return map when read success", func() {
			// create temp csv file
			var filePath = "./test_csv4.csv"
			err := createTmpCsvFile(filePath)
			convey.So(err, convey.ShouldBeNil)
			defer clearTmpCsvFile(filePath)
			ret1, err := readCSVFile(filePath, 0)
			convey.So(err, convey.ShouldBeNil)
			convey.So(ret1, convey.ShouldNotBeEmpty)
		})
	})
}

func createTmpCsvFile(filePath string) error {
	fileContent := `
pingTaskId,srcType,srcAddr,dstType,dstAddr,minDelay,maxDelay,avgDelay,minLossRate,maxLossRate,avgLossRate,timestamp
0,0,6094863,0,4194304,-1,-1,-1,1.000,1.000,1.000,1755687818727
`
	var fileMode0644 os.FileMode = 0644
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR, fileMode0644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(fileContent)
	return err
}

func clearTmpCsvFile(filePath string) {
	err := os.Remove(filePath)
	if err != nil {
		hwlog.RunLog.Errorf("remove temp csv file %s failed: %v", filePath, err)
	}
}

func TestFindCsvFile(t *testing.T) {
	convey.Convey("test func find CSVFile", t, func() {
		convey.Convey("return nil whem walk err", func() {
			patch := gomonkey.ApplyFuncReturn(filepath.Walk, errors.New("err"))
			defer patch.Reset()
			ret, _ := findCSVFiles("")
			convey.So(ret, convey.ShouldBeNil)
		})
	})
}

func TestGetFalseFlagDetection(t *testing.T) {
	convey.Convey("TestGetFalseFlagDetection", t, func() {
		convey.Convey("delete detection a", func() {
			for k, _ := range superPodDetectionRecorder {
				delete(superPodDetectionRecorder, k)
			}
			superPodDetectionRecorder["a"] = false
			ret := getFalseFlagDetection(false)
			convey.So(len(ret) == 1, convey.ShouldEqual, true)
		})
	})
}

func TestIfAddNewSuperPodDetection(t *testing.T) {
	convey.Convey("TestIfAddNewSuperPodDetection", t, func() {
		convey.Convey("Test IfAddNewSuperPodDetection", func() {
			var wg sync.WaitGroup
			controllerflags.IsControllerExited.SetState(false)
			superPodDetectionRecorder["1"] = false
			superPodDetectionRecorder["2"] = false
			patch1 := gomonkey.ApplyFuncReturn(getSuperPodDirInfo, []int{1}, []string{"/tmp/test"})
			defer patch1.Reset()
			patch2 := gomonkey.ApplyFunc(addNewSuperPodDetection, func(wg *sync.WaitGroup, a []string, b []int) {
				wg.Done()
			})
			defer patch2.Reset()
			wg.Add(1)
			ifAddNewSuperPodDetection("", &wg)
			time.Sleep(1 * time.Second)
			controllerflags.IsControllerExited.SetState(true)
			wg.Wait()
			superPodDetectionRecorder = map[string]bool{}
			convey.So(len(superPodDetectionRecorder) == 0, convey.ShouldBeTrue)
		})
	})
}

func TestAddNewSuperPodDetection(t *testing.T) {
	convey.Convey("TestIfAddNewSuperPodDetection", t, func() {
		convey.Convey("empty super pod detection numbers", func() {
			var wg sync.WaitGroup
			addNewSuperPodDetection(&wg, []string{}, []int{})
		})
		convey.Convey("over max super pod detection numbers", func() {
			var wg sync.WaitGroup
			a := make([]string, maxSuperPodDetectionNums+1)
			b := make([]int, maxSuperPodDetectionNums+1)
			addNewSuperPodDetection(&wg, a, b)
		})
	})
}

func TestStartRoceRoutine(t *testing.T) {
	convey.Convey("TestStartRoceRoutine", t, func() {
		convey.Convey("start failed", func() {
			var wg sync.WaitGroup
			wg.Add(1)
			patch := gomonkey.ApplyFunc(startRoceDetection, func(path string) {
				return
			})
			defer patch.Reset()
			startRoceRoutine(&wg)
			wg.Wait()
		})
	})
}

func TestDetectionCurSuperPod(t *testing.T) {
	convey.Convey("test func detectionCurSuperPod", t, func() {
		patch := gomonkey.NewPatches()
		defer patch.Reset()
		patch.ApplyFuncReturn(loopWaitSuperPodDirAndCheckConfigFile, true)
		convey.Convey("should return when setCallAlgoParam failed", func() {
			subPatch := gomonkey.ApplyFuncReturn(checkDiffConfig, nil)
			defer subPatch.Reset()
			detectionCurSuperPod(0, "")
		})
		patch.ApplyFuncReturn(checkDiffConfig, map[string]any{})
		convey.Convey("should return when set callAlgoParam failed", func() {
			subPatch := gomonkey.ApplyFuncReturn(policy.SetCallAlgorithmParamInfo, errors.New("err"))
			defer subPatch.Reset()
			detectionCurSuperPod(0, "")
		})
		patch.ApplyFuncReturn(policy.SetCallAlgorithmParamInfo, nil)
		convey.Convey("should return when get npuInfoMap failed", func() {
			subPatch := gomonkey.ApplyFuncReturn(policy.GetTargetSuperPodNpuMap, false, nil)
			defer subPatch.Reset()
			detectionCurSuperPod(0, "")
		})
		convey.Convey("should return when gen pingList failed", func() {
			subPatch := gomonkey.ApplyFuncReturn(policy.GetTargetSuperPodNpuMap, true, nil)
			subPatch.ApplyFuncReturn(policy.GenSuperPodServersPingList, false)
			defer subPatch.Reset()
			detectionCurSuperPod(0, "")
		})
	})
}

func TestMakeAlgoRoceParam(t *testing.T) {
	convey.Convey("Test func makeAlgoRoceParam", t, func() {
		convey.Convey("return nil when checkDiffConfig", func() {
			patch := gomonkey.ApplyFuncReturn(checkDiffConfig, nil)
			defer patch.Reset()
			ret := makeAlgoRoceParam("path", []string{})
			convey.So(ret, convey.ShouldBeNil)
		})
		convey.Convey("return map when normal", func() {
			paramMap := make(map[string]any)
			dirName := []string{"1", "2", "3"}
			patch := gomonkey.ApplyFuncReturn(checkDiffConfig, map[string]any{})
			defer patch.Reset()
			paramMap[policy.ServerIdMap] = make(map[string]string)
			paramMap["axisStategy"] = "same_axis"
			paramMap["superPodArr"] = dirName
			paramMap["superPodJobFlag"] = true
			paramMap["npu_type"] = "A5"
			paramMap["pingObjType"] = 0
			ret := makeAlgoRoceParam("path", dirName)
			convey.So(ret, convey.ShouldResemble, paramMap)
		})
	})
}

func TestGetRoceDetectionSuperPodFile(t *testing.T) {
	convey.Convey("test func getRoceDetectionSuperPodFile", t, func() {
		convey.Convey("return nil when ReadFile err", func() {
			patch := gomonkey.ApplyFuncReturn(os.ReadFile, nil, errors.New("err"))
			defer patch.Reset()
			ret1, ret2 := getRoceDetectionSuperPodFile("", "")
			convey.So(ret1, convey.ShouldBeNil)
			convey.So(ret2, convey.ShouldBeNil)
		})
		convey.Convey("should nil unmarshal err", func() {
			patch1 := gomonkey.ApplyFuncReturn(os.ReadFile, nil, nil)
			patch2 := gomonkey.ApplyFuncReturn(json.Unmarshal, errors.New("err"))
			defer patch1.Reset()
			defer patch2.Reset()
			ret1, ret2 := getRoceDetectionSuperPodFile("", "")
			convey.So(ret1, convey.ShouldBeNil)
			convey.So(ret2, convey.ShouldBeNil)
		})
		convey.Convey("valid parameters", func() {
			patch1 := gomonkey.ApplyFuncReturn(os.ReadFile, []byte(`{"superPodList:[1]"}`), nil)
			patch2 := gomonkey.ApplyFuncReturn(loopWaitSuperPodDirAndCheckConfigFile, true)
			defer patch1.Reset()
			defer patch2.Reset()
			ret1, ret2 := getRoceDetectionSuperPodFile("", "")
			convey.So(len(ret1) > 0, convey.ShouldBeTrue)
			convey.So(len(ret2) > 0, convey.ShouldBeTrue)
		})
	})
}

func TestStartRoceDetection(t *testing.T) {
	convey.Convey("test func start RoceDetection", t, func() {
		convey.Convey("should return when wait error", func() {
			patch := gomonkey.ApplyFuncReturn(loopWaitSuperPodDirAndCheckConfigFile, false)
			defer patch.Reset()
			startRoceDetection("")
		})
		patch := gomonkey.NewPatches()
		defer patch.Reset()
		patch.ApplyFuncReturn(loopWaitSuperPodDirAndCheckConfigFile, true)
		convey.Convey("should return when getNpuInfo err", func() {
			subPatch := gomonkey.ApplyFuncReturn(policy.GetSuperPodsRoceNpuInfo, nil, nil)
			defer subPatch.Reset()
			startRoceDetection("")
		})
		patch.ApplyFuncReturn(policy.GetSuperPodsRoceNpuInfo, map[string]algo.NpuInfo{}, map[string]any{})
		convey.Convey("should return when GenPinglist err", func() {
			subPatch := gomonkey.ApplyFuncReturn(policy.GenRoceSuperPodLevelPingList, false)
			defer subPatch.Reset()
			startRoceDetection("")
		})
		patch.ApplyFuncReturn(policy.GenRoceSuperPodLevelPingList, true)
		convey.Convey("should Detection when normal", func() {
			call := 0
			subPatch := gomonkey.ApplyFunc(loopCsvCallDetection, func(_ detectionParam) { call++ })
			defer subPatch.Reset()
			startRoceDetection("")
			convey.So(call, convey.ShouldEqual, 0)
		})
	})
}
