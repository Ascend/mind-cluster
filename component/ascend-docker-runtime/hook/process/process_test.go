/* Copyright(C) 2022. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package main
package process

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"

	"ascend-common/api"
	"ascend-common/cdi/mount"
	"ascend-common/common-utils/hwlog"
	"ascend-docker-runtime/mindxcheckutils"
)

const (
	pidSample                              = 123
	fileMode0600               os.FileMode = 0600
	ascendVisibleDeviceTestStr             = api.AscendVisibleDevicesEnv + "=0-3,5,7"
	configFile                             = "config.json"
	strRepeatTimes                         = 129
	testStr                                = "test"
)

var (
	containerConfigInputStream = os.Stdin
	testError                  = errors.New("test")
)

// TestMain initializes the logger before tests run individually or in any order.
func TestMain(m *testing.M) {
	if err := hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "initialize test logger: %v\n", err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}

// TestDoPrestartHookCase1 test function DoPrestartHook
func TestDoPrestartHookCase1(t *testing.T) {
	err := DoPrestartHook()
	assert.NotNil(t, err)
}

// TestDoPrestartHookCase2 test function DoPrestartHook
func TestDoPrestartHookCase2(t *testing.T) {
	conCfg := containerConfig{
		Pid:    pidSample,
		Rootfs: ".",
		Env:    []string{api.AscendVisibleDevicesEnv + "=0l-3,5,7"},
	}
	patch := gomonkey.ApplyFuncReturn(getContainerConfig, &conCfg, nil)
	defer patch.Reset()
	err := DoPrestartHook()
	assert.NotNil(t, err)
}

// TestDoPrestartHookCase3 test function DoPrestartHook
func TestDoPrestartHookCase3(t *testing.T) {
	conCfg := containerConfig{
		Pid:    pidSample,
		Rootfs: ".",
		Env:    []string{"ASCEND_VISIBLE_DEVICE=0-3,5,7"},
	}
	patch := gomonkey.ApplyFuncReturn(getContainerConfig, &conCfg, nil)
	defer patch.Reset()
	err := DoPrestartHook()
	assert.Nil(t, err)
}

// TestDoPrestartHookCase4 test function DoPrestartHook
func TestDoPrestartHookCase4(t *testing.T) {
	conCfg := containerConfig{
		Pid:    pidSample,
		Rootfs: ".",
		Env: []string{
			api.AscendVisibleDevicesEnv + "=0",
			api.AscendRuntimeOptionsEnv + "=VIRTUAL,NODRV",
		},
	}
	err := InitLogModule(context.Background())
	if err != nil {
		t.Log("failed")
	}
	patch := gomonkey.ApplyFuncReturn(getContainerConfig, &conCfg, nil).
		ApplyGlobalVar(&ascendDockerCliName, "").
		ApplyFuncReturn(doExec, nil)
	defer patch.Reset()
	err = DoPrestartHook()
	assert.NotNil(t, err)
}

// TestDoPrestartHookCase5 test function DoPrestartHook
func TestDoPrestartHookCase5(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			t.Log("exception", err)
		}
	}()
	conCfg := containerConfig{
		Pid:    pidSample,
		Rootfs: ".",
		Env:    []string{ascendVisibleDeviceTestStr},
	}
	patch := gomonkey.ApplyFuncReturn(getContainerConfig, &conCfg, nil).
		ApplyGlobalVar(&ascendDockerCliName, "clii").
		ApplyGlobalVar(&defaultAscendDockerCliName, "clii").
		ApplyFuncReturn(doExec, nil)
	defer patch.Reset()
	err := DoPrestartHook()
	assert.NotNil(t, err)
}

// TestDoPrestartHookPatch1 test function DoPrestartHook
func TestDoPrestartHookPatch1(t *testing.T) {
	convey.Convey("test DoPrestartHook patch1", t, func() {
		ctrCfg := &containerConfig{}
		patches := gomonkey.ApplyFuncReturn(getContainerConfig, ctrCfg, nil).
			ApplyFuncReturn(getValueByKey, testStr).
			ApplyFuncReturn(parseMounts, []string{testStr}).
			ApplyFuncReturn(parseUBDrvMount, true, nil).
			ApplyFuncReturn(os.Stat, fileInfoMock{}, nil).
			ApplyFuncReturn(readConfigsOfDir, []string{testStr}, []string{testStr}, nil).
			ApplyFunc(addUBMount, func(fileMountList []string, dirMountList []string) ([]string, []string) {
				return fileMountList, dirMountList
			})
		defer patches.Reset()
		convey.Convey("01-parseUBDrvMount error, should return error", func() {
			patch := gomonkey.ApplyFuncReturn(parseUBDrvMount, true, testError)
			defer patch.Reset()
			err := DoPrestartHook()
			convey.So(err, convey.ShouldBeError)
		})
		patches.ApplyFuncReturn(parseUBDrvMount, true, nil)
		convey.Convey("02-readConfigsOfDir error, should return error", func() {
			patch := gomonkey.ApplyFuncReturn(readConfigsOfDir, []string{}, []string{}, testError)
			defer patch.Reset()
			err := DoPrestartHook()
			convey.So(err, convey.ShouldBeError)
		})
		patches.ApplyFuncReturn(readConfigsOfDir, []string{testStr}, []string{testStr}, nil)
		convey.Convey("03-parseRuntimeOptions error, should return error", func() {
			patch := gomonkey.ApplyFuncReturn(parseRuntimeOptions, []string{testStr}, testError)
			defer patch.Reset()
			err := DoPrestartHook()
			convey.So(err, convey.ShouldBeError)
		})
		patches.ApplyFuncReturn(parseRuntimeOptions, []string{testStr}, nil)
		convey.Convey("04-parseSoftLinkMode error, should return error", func() {
			patch := gomonkey.ApplyFuncReturn(parseSoftLinkMode, "", testError)
			defer patch.Reset()
			err := DoPrestartHook()
			convey.So(err, convey.ShouldBeError)
		})
		patches.ApplyFuncReturn(parseSoftLinkMode, testStr, nil)
		convey.Convey("05-Executable error, should return error", func() {
			patch := gomonkey.ApplyFuncReturn(os.Executable, "", testError)
			defer patch.Reset()
			err := DoPrestartHook()
			convey.So(err, convey.ShouldBeError)
		})
		patches.ApplyFuncReturn(os.Executable, testStr, nil)
		convey.Convey("06-Stat error, should return error", func() {
			patch := gomonkey.ApplyFuncReturn(os.Stat, fileInfoMock{}, testError)
			defer patch.Reset()
			err := DoPrestartHook()
			convey.So(err, convey.ShouldBeError)
		})
		patches.ApplyFuncReturn(os.Stat, fileInfoMock{}, nil).
			ApplyFuncReturn(getArgs, []string{testStr})
		convey.Convey("07-RealFileChecker error, should return error", func() {
			patch := gomonkey.ApplyFuncReturn(mindxcheckutils.RealFileChecker, testStr, testError)
			defer patch.Reset()
			err := DoPrestartHook()
			convey.So(err, convey.ShouldBeError)
		})
		patches.ApplyFuncReturn(mindxcheckutils.RealFileChecker, testStr, nil)
		convey.Convey("08-ChangeRuntimeLogMode error, should return error", func() {
			patch := gomonkey.ApplyFuncReturn(mindxcheckutils.ChangeRuntimeLogMode, testError)
			defer patch.Reset()
			err := DoPrestartHook()
			convey.So(err, convey.ShouldBeError)
		})
	})
}

// TestDoPrestartHookPatch2 test function DoPrestartHook
func TestDoPrestartHookPatch2(t *testing.T) {
	convey.Convey("test DoPrestartHook patch2", t, func() {
		ctrCfg := &containerConfig{}
		patches := gomonkey.ApplyFuncReturn(getContainerConfig, ctrCfg, nil).
			ApplyFuncReturn(getValueByKey, testStr).
			ApplyFuncReturn(parseMounts, []string{testStr}).
			ApplyFuncReturn(parseUBDrvMount, true, nil).
			ApplyFuncReturn(readConfigsOfDir, []string{testStr}, []string{testStr}, nil).
			ApplyFunc(addUBMount, func(fileMountList []string, dirMountList []string) ([]string, []string) {
				return fileMountList, dirMountList
			}).
			ApplyFuncReturn(parseRuntimeOptions, []string{testStr}, nil).
			ApplyFuncReturn(parseSoftLinkMode, testStr, nil).
			ApplyFuncReturn(os.Executable, testStr, nil).
			ApplyFuncReturn(os.Stat, fileInfoMock{}, nil).
			ApplyFuncReturn(mindxcheckutils.RealFileChecker, testStr, nil).
			ApplyFuncReturn(mindxcheckutils.ChangeRuntimeLogMode, nil)
		defer patches.Reset()
		convey.Convey("06-doExec error, should return error", func() {
			patch := gomonkey.ApplyFuncVarReturn(&doExec, testError)
			defer patch.Reset()
			err := DoPrestartHook()
			convey.So(err, convey.ShouldBeError)
		})
		patches.ApplyFuncVarReturn(&doExec, nil)
		convey.Convey("07-success, should return nil", func() {
			err := DoPrestartHook()
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("08-allowLink forced to True when allowLink=False and mountUBDrv=True", func() {
			var capturedAllowLink string
			patchSoftLink := gomonkey.ApplyFuncReturn(parseSoftLinkMode, "False", nil)
			defer patchSoftLink.Reset()
			// Keep getArgs real: gomonkey's ARM64 jump overwrites X10, which holds
			// its final string argument's length. Capture the generated CLI arguments instead.
			patchDoExec := gomonkey.ApplyFuncVar(&doExec, func(_ string, args, _ []string) error {
				for i := 0; i+1 < len(args); i++ {
					if args[i] == "--allow-link" {
						capturedAllowLink = args[i+1]
						break
					}
				}
				return nil
			})
			defer patchDoExec.Reset()
			err := DoPrestartHook()
			convey.So(err, convey.ShouldBeNil)
			convey.So(capturedAllowLink, convey.ShouldEqual, "True")
		})
	})
}

// TestGetValueByKeyCase1 test the function getValueByKey
func TestGetValueByKeyCase1(t *testing.T) {
	data := []string{ascendVisibleDeviceTestStr}
	word := api.AscendVisibleDevicesEnv
	expectVal := "0-3,5,7"
	actualVal := getValueByKey(data, word)
	if actualVal != expectVal {
		t.Fail()
	}
}

// TestGetValueByKeyCase2 test the function getValueByKey
func TestGetValueByKeyCase2(t *testing.T) {
	data := []string{api.AscendVisibleDevicesEnv}
	word := api.AscendVisibleDevicesEnv
	expectVal := ""
	defer func() {
		if err := recover(); err != nil {
			t.Log("exception occur")
		}
	}()
	actualVal := getValueByKey(data, word)
	if actualVal != expectVal {
		t.Fail()
	}
}

// TestGetValueByKeyCase3 test the function getValueByKey
func TestGetValueByKeyCase3(t *testing.T) {
	data := []string{ascendVisibleDeviceTestStr}
	word := "ASCEND_VISIBLE_DEVICE"
	expectVal := ""
	actualVal := getValueByKey(data, word)
	if actualVal != expectVal {
		t.Fail()
	}
}

// TestParseOciSpecFileCase1 test the function parseOciSpecFile
func TestParseOciSpecFileCase1(t *testing.T) {
	file := "file"
	_, err := parseOciSpecFile(file)
	if err == nil {
		t.Fail()
	}
}

// TestParseOciSpecFileCase2 test the function parseOciSpecFile
func TestParseOciSpecFileCase2(t *testing.T) {
	file := "file"
	f, err := os.Create(file)
	defer os.Remove(file)
	defer f.Close()
	if err != nil {
		t.Log("create file failed")
	}
	err = f.Chmod(fileMode0600)
	if err != nil {
		t.Logf("chmod file error: %v", err)
	}
	_, err = parseOciSpecFile(file)
	if err == nil {
		t.Fail()
	}
}

// TestParseOciSpecFileCase3 test the function parseOciSpecFile
func TestParseOciSpecFileCase3(t *testing.T) {
	file, err := os.Create(configFile)
	if err != nil {
		t.Log("create file failed")
		t.FailNow()
	}
	defer os.Remove(configFile)
	defer file.Close()
	err = file.Chmod(fileMode0600)
	if err != nil {
		t.Log("chmod file failed")
		t.FailNow()
	}
	testSpec := specs.Spec{}
	jsonData, err := json.MarshalIndent(testSpec, "", "    ")
	if err != nil {
		t.Logf("failed to MarshalIndent, err: %v", err)
		t.FailNow()
	}
	_, err = file.Write(jsonData)
	if err != nil {
		t.Logf("failed to Write, err: %v", err)
		t.FailNow()
	}
	_, err = parseOciSpecFile(configFile)
	assert.Equal(t, fmt.Errorf("invalid OCI spec for empty process"), err)
}

// TestGetContainerConfig test the function getContainerConfig
func TestGetContainerConfig(t *testing.T) {
	cmd := exec.Command("runc", "spec")
	if err := cmd.Run(); err != nil {
		t.Log("runc spec failed")
	}
	defer func() {
		if err := recover(); err != nil {
			t.Log("exception", err)
		}
	}()
	defer os.Remove(configFile)
	stateFile, err := os.Open(configFile)
	if err != nil {
		t.Log("open file failed")
	}
	defer stateFile.Close()

	patch := gomonkey.ApplyGlobalVar(&containerConfigInputStream, stateFile)
	defer patch.Reset()

	getContainerConfig()
}

// TestGetContainerConfigPatch1 test the function getContainerConfig
func TestGetContainerConfigPatch1(t *testing.T) {
	convey.Convey("test getContainerConfig patch1", t, func() {
		patches := gomonkey.ApplyMethodReturn(json.NewDecoder(containerConfigInputStream),
			"Decode", nil)
		defer patches.Reset()
		convey.Convey("01-RealFileChecker error, should return error", func() {
			patch := gomonkey.ApplyFuncReturn(mindxcheckutils.RealFileChecker, testStr, testError)
			defer patch.Reset()
			_, err := getContainerConfig()
			convey.So(err, convey.ShouldBeError)
		})
		patches.ApplyFuncReturn(mindxcheckutils.RealFileChecker, testStr, nil)
		convey.Convey("02-parseOciSpecFile error, should return error", func() {
			patch := gomonkey.ApplyFuncReturn(parseOciSpecFile, &specs.Spec{}, testError)
			defer patch.Reset()
			_, err := getContainerConfig()
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("03-over MaxCommandLength, should return error", func() {
			sp := &specs.Spec{
				Process: &specs.Process{
					Env: make([]string, MaxCommandLength+1),
				},
			}
			patch := gomonkey.ApplyFuncReturn(parseOciSpecFile, sp, nil)
			defer patch.Reset()
			_, err := getContainerConfig()
			convey.So(err, convey.ShouldBeError)
		})
	})
}

// TestGetContainerConfigPatch2 test the function getContainerConfig
func TestGetContainerConfigPatch2(t *testing.T) {
	convey.Convey("test getContainerConfig patch2", t, func() {
		testSp := &specs.Spec{
			Process: &specs.Process{
				Env: make([]string, 1),
			},
			Root: &specs.Root{
				Path: testStr,
			},
		}
		patches := gomonkey.ApplyFuncReturn(mindxcheckutils.RealFileChecker, testStr, nil).
			ApplyFuncReturn(parseOciSpecFile, testSp, nil).
			ApplyMethodReturn(json.NewDecoder(containerConfigInputStream),
				"Decode", nil)
		defer patches.Reset()
		convey.Convey("04-rfs not abs path, success, should return nil", func() {
			patch := gomonkey.ApplyFuncReturn(filepath.Abs, testStr, nil)
			defer patch.Reset()
			_, err := getContainerConfig()
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// fileInfoMock is used to test
type fileInfoMock struct {
	os.FileInfo
}

func (f fileInfoMock) Mode() os.FileMode {
	return os.ModePerm
}

// TestReadConfigsOfDir tests the function readConfigsOfDir
func TestReadConfigsOfDir(t *testing.T) {
	patch := gomonkey.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) {
		return &fileInfoMock{}, nil
	})
	defer patch.Reset()
	patchSize := gomonkey.ApplyMethod(reflect.TypeOf(&fileInfoMock{}), "Mode", func(f *fileInfoMock) fs.FileMode {
		return fs.ModeDir
	})
	defer patchSize.Reset()
	tests := []struct {
		name    string
		dir     string
		configs []string
		want    []string
		want1   []string
		wantErr bool
	}{
		{
			name:  "readConfigsOfDir success case 1",
			want:  []string{},
			want1: []string{},
		},
		{
			name:    "readConfigsOfDir fail case 2",
			configs: []string{"base"},
			wantErr: true,
			want:    nil,
			want1:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := readConfigsOfDir(tt.dir, tt.configs, nil)
			if (err == nil) == tt.wantErr {
				t.Errorf("readConfigsOfDir() got = %v, want %v", err, tt.wantErr)
			}
			assert.Equalf(t, tt.want, got, "readConfigsOfDir(%v, %v)", tt.dir, tt.configs)
			assert.Equalf(t, tt.want1, got1, "readConfigsOfDir(%v, %v)", tt.dir, tt.configs)
		})
	}
}

// TestDoPrestartHook tests the function DoPrestartHook
func TestDoPrestartHook(t *testing.T) {
	patches := prepareRDMAHook(t, []string{api.AscendVisibleDevicesEnv + "=0-3,5,7",
		api.AscendRuntimeMountsEnv + "=a", api.AscendUBDrvMountEnv + "=False"})
	readCalls, execCalls := 0, 0
	readConfigsStub := readConfigsOfDir
	readConfigsStub = func(dir string, configNames []string, rdmaCollections map[string]bool) ([]string, []string, error) {
		readCalls++
		assert.Equal(t, configDir, dir)
		assert.Equal(t, []string{"a"}, configNames)
		assert.Nil(t, rdmaCollections)
		return []string{"/host/driver.so"}, []string{"/host/config"}, nil
	}
	patches.ApplyFunc(readConfigsOfDir, readConfigsStub)
	doExec = func(_ string, args, _ []string) error {
		execCalls++
		assert.Equal(t, []string{"/usr/local/bin/ascend-docker-cli", "--allow-link", "False", "--pid", "123",
			"--rootfs", "/rootfs", "--mount-file", "/host/driver.so", "--mount-dir", "/host/config"}, args)
		return nil
	}
	assert.NoError(t, DoPrestartHook())
	assert.Equal(t, 1, readCalls)
	assert.Equal(t, 1, execCalls)
}

// TestParseRuntimeOptions tests the function parseRuntimeOptions
func TestParseRuntimeOptions(t *testing.T) {
	virtualValue := "VIRTUAL"
	tests := []struct {
		name           string
		runtimeOptions string
		want           []string
		wantErr        bool
	}{
		{
			name:           "too long case 1",
			runtimeOptions: strings.Repeat("a", strRepeatTimes),
			wantErr:        true,
		},
		{
			name:           "invalid case 2",
			runtimeOptions: "a",
			wantErr:        true,
		},
		{
			name:           "success case 3",
			runtimeOptions: virtualValue,
			wantErr:        false,
			want:           []string{virtualValue},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseRuntimeOptions(tt.runtimeOptions)
			if (err == nil) == tt.wantErr {
				t.Errorf("DoPrestartHook() got = %v, want %v", err, tt.wantErr)
			}
			assert.Equalf(t, tt.want, got, "parseRuntimeOptions(%v)", tt.runtimeOptions)
		})
	}
}

// TestAddUBMount tests the function addUBMount
func TestAddUBMount(t *testing.T) {
	convey.Convey("test addUBMount", t, func() {
		convey.Convey("test addUBMount should do nothing when mount file and directory not exist", func() {
			var fileMountList []string
			var dirMountList []string
			mockStat := gomonkey.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) {
				return nil, os.ErrNotExist
			})
			defer mockStat.Reset()
			fileMountList, dirMountList = addUBMount(fileMountList, dirMountList)
			convey.So(fileMountList, convey.ShouldBeEmpty)
			convey.So(dirMountList, convey.ShouldBeEmpty)
		})
		convey.Convey("test addUBMount should append list when mount directory exists", func() {
			var fileMountList []string
			var dirMountList []string
			mockStat := mockStat()
			defer mockStat.Reset()
			fileMountList, dirMountList = addUBMount(fileMountList, dirMountList)
			convey.So(fileMountList, convey.ShouldBeEmpty)
			convey.So(dirMountList, convey.ShouldNotBeEmpty)
		})
	})
}

// TestContainsGlob tests the containsGlob function
func TestContainsGlob(t *testing.T) {
	convey.Convey("test containsGlob", t, func() {
		convey.Convey("should return true for patterns with *", func() {
			convey.So(containsGlob("/usr/lib64/libummu*"), convey.ShouldBeTrue)
		})
		convey.Convey("should return true for patterns with ?", func() {
			convey.So(containsGlob("/usr/lib64/libummu?.so"), convey.ShouldBeTrue)
		})
		convey.Convey("should return true for patterns with [", func() {
			convey.So(containsGlob("/usr/lib64/lib[abc].so"), convey.ShouldBeTrue)
		})
		convey.Convey("should return false for exact paths", func() {
			convey.So(containsGlob("/usr/lib64/libummu.so"), convey.ShouldBeFalse)
		})
	})
}

// TestExpandGlobPathWhenValidFile tests expandGlobPath with a valid glob match
func TestExpandGlobPathWhenValidFile(t *testing.T) {
	convey.Convey("test expandGlobPath when glob matches valid files", t, func() {
		mockGlob := gomonkey.ApplyFunc(filepath.Glob, func(pattern string) ([]string, error) {
			if strings.Contains(pattern, "ubrma") {
				return []string{"/usr/lib64/ubrma/lib.so"}, nil
			}
			return []string{"/usr/lib64/libummu.so"}, nil
		})
		defer mockGlob.Reset()
		mockEvalSymlinks := gomonkey.ApplyFunc(filepath.EvalSymlinks,
			func(path string) (string, error) { return path, nil })
		defer mockEvalSymlinks.Reset()
		mockStat := gomonkey.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) {
			return mockFileInfo{mode: os.ModeTemporary}, nil
		})
		defer mockStat.Reset()
		mockLstat := gomonkey.ApplyFunc(os.Lstat, func(name string) (os.FileInfo, error) {
			return mockFileInfo{mode: 0}, nil
		})
		defer mockLstat.Reset()
		fileList, dirList := expandGlobPath("/usr/lib64/libummu*")
		convey.So(len(fileList), convey.ShouldEqual, 1)
		convey.So(dirList, convey.ShouldBeEmpty)
	})
}

// TestExpandGlobPathWhenSymlinkOutsideDir tests expandGlobPath skips symlink outside expected dir
func TestExpandGlobPathWhenSymlinkOutsideDir(t *testing.T) {
	convey.Convey("test expandGlobPath when symlink points outside expected dir", t, func() {
		mockGlob := gomonkey.ApplyFunc(filepath.Glob, func(pattern string) ([]string, error) {
			return []string{"/usr/lib64/libummu.so"}, nil
		})
		defer mockGlob.Reset()
		mockEvalSymlinks := gomonkey.ApplyFunc(filepath.EvalSymlinks,
			func(path string) (string, error) { return "/opt/other/libummu.so", nil })
		defer mockEvalSymlinks.Reset()
		fileList, dirList := expandGlobPath("/usr/lib64/libummu*")
		convey.So(fileList, convey.ShouldBeEmpty)
		convey.So(dirList, convey.ShouldBeEmpty)
	})
}

// TestExpandGlobPathWhenSymlinkNotOwnedByRoot tests expandGlobPath skips symlink not owned by root
func TestExpandGlobPathWhenSymlinkNotOwnedByRoot(t *testing.T) {
	convey.Convey("test expandGlobPath when symlink is not owned by root", t, func() {
		mockGlob := gomonkey.ApplyFunc(filepath.Glob, func(pattern string) ([]string, error) {
			return []string{"/usr/lib64/libummu.so"}, nil
		})
		defer mockGlob.Reset()
		mockEvalSymlinks := gomonkey.ApplyFunc(filepath.EvalSymlinks,
			func(path string) (string, error) { return path, nil })
		defer mockEvalSymlinks.Reset()
		mockStat := gomonkey.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) {
			return mockFileInfo{mode: os.ModeTemporary}, nil
		})
		defer mockStat.Reset()
		mockLstat := gomonkey.ApplyFunc(os.Lstat, func(name string) (os.FileInfo, error) {
			if name != "/usr/lib64/libummu.so" {
				return mockFileInfo{mode: os.ModeDir}, nil
			}
			return mockFileInfo{mode: os.ModeSymlink}, nil
		})
		defer mockLstat.Reset()
		mockGetFileUID := gomonkey.ApplyFunc(getFileUID, func(info os.FileInfo) uint32 {
			return 1000
		})
		defer mockGetFileUID.Reset()
		fileList, dirList := expandGlobPath("/usr/lib64/libummu*")
		convey.So(fileList, convey.ShouldBeEmpty)
		convey.So(dirList, convey.ShouldBeEmpty)
	})
}

// TestExpandGlobPathWhenRealFileNotOwnedByRoot tests expandGlobPath skips when real file is not owned by root
func TestExpandGlobPathWhenRealFileNotOwnedByRoot(t *testing.T) {
	convey.Convey("test expandGlobPath when real file is not owned by root", t, func() {
		mockGlob := gomonkey.ApplyFunc(filepath.Glob, func(pattern string) ([]string, error) {
			return []string{"/usr/lib64/libummu.so"}, nil
		})
		defer mockGlob.Reset()
		mockEvalSymlinks := gomonkey.ApplyFunc(filepath.EvalSymlinks,
			func(path string) (string, error) { return path, nil })
		defer mockEvalSymlinks.Reset()
		mockStat := gomonkey.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) {
			return mockFileInfo{mode: os.ModeTemporary}, nil
		})
		defer mockStat.Reset()
		mockLstat := gomonkey.ApplyFunc(os.Lstat, func(name string) (os.FileInfo, error) {
			if name != "/usr/lib64/libummu.so" {
				return mockFileInfo{mode: os.ModeDir}, nil
			}
			return mockFileInfo{mode: os.ModeSymlink}, nil
		})
		defer mockLstat.Reset()
		callCount := 0
		mockGetFileUID := gomonkey.ApplyFunc(getFileUID, func(info os.FileInfo) uint32 {
			callCount++
			if callCount == 1 {
				return 0 // symlink owner is root
			}
			return 1000 // real file owner is not root
		})
		defer mockGetFileUID.Reset()
		fileList, dirList := expandGlobPath("/usr/lib64/libummu*")
		convey.So(fileList, convey.ShouldBeEmpty)
		convey.So(dirList, convey.ShouldBeEmpty)
	})
}

// TestGetArgs tests the function getArgs
func TestGetArgs(t *testing.T) {
	tests := []struct {
		name            string
		cliPath         string
		containerConfig *containerConfig
		fileMountList   []string
		dirMountList    []string
		allowLink       string
		want            []string
	}{
		{
			name:            "success case 1",
			containerConfig: &containerConfig{},
			fileMountList:   []string{"test"},
			dirMountList:    []string{"test"},
			allowLink:       "true",
			cliPath:         "testcli",
			want: []string{"testcli", "--allow-link", "true", "--pid", "0", "--rootfs", "",
				"--mount-file", "test", "--mount-dir", "test"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, getArgs(tt.cliPath, tt.containerConfig, tt.fileMountList,
				tt.dirMountList, tt.allowLink), "getArgs(%v, %v, %v, %v, %v)",
				tt.cliPath, tt.containerConfig, tt.fileMountList, tt.dirMountList, tt.allowLink)
		})
	}
}

// TestParseMounts tests the function parseMounts
func TestParseMounts(t *testing.T) {
	tests := []struct {
		name   string
		mounts string
		want   []string
	}{
		{
			name:   "base case 1",
			mounts: "",
			want:   []string{baseConfig},
		},
		{
			name:   "base case 2",
			mounts: strings.Repeat("a", strRepeatTimes),
			want:   []string{baseConfig},
		},
		{
			name:   "other case 3",
			mounts: "testList,testList1",
			want:   []string{"testlist", "testlist1"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, parseMounts(tt.mounts), "parseMounts(%v)", tt.mounts)
		})
	}
}

// TestParseMountEntry tests the function parseMountEntry
func TestParseMountEntry(t *testing.T) {
	convey.Convey("test parseMountEntry", t, func() {
		convey.Convey("legacy entry without colon keeps same-path mount", func() {
			hostPath, containerPath, err := parseMountEntry("/usr/local/Ascend/driver")
			convey.So(err, convey.ShouldBeNil)
			convey.So(hostPath, convey.ShouldEqual, "/usr/local/Ascend/driver")
			convey.So(containerPath, convey.ShouldBeEmpty)
		})
		convey.Convey("mapping entry should be split correctly", func() {
			hostPath, containerPath, err := parseMountEntry("/data/models:/workspace/models")
			convey.So(err, convey.ShouldBeNil)
			convey.So(hostPath, convey.ShouldEqual, "/data/models")
			convey.So(containerPath, convey.ShouldEqual, "/workspace/models")
		})
		convey.Convey("entry with spaces around parts should be trimmed", func() {
			hostPath, containerPath, err := parseMountEntry(" /data/a : /container/a ")
			convey.So(err, convey.ShouldBeNil)
			convey.So(hostPath, convey.ShouldEqual, "/data/a")
			convey.So(containerPath, convey.ShouldEqual, "/container/a")
		})
		convey.Convey("blank entry should return error", func() {
			_, _, err := parseMountEntry("")
			convey.So(err, convey.ShouldBeError)
			_, _, err = parseMountEntry("   ")
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("empty part should return error", func() {
			_, _, err := parseMountEntry(":/container/a")
			convey.So(err, convey.ShouldBeError)
			_, _, err = parseMountEntry("/data/a:")
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("relative container path should return error", func() {
			_, _, err := parseMountEntry("/data/a:container/a")
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("container path contains '..' should return error", func() {
			_, _, err := parseMountEntry("/data/a:/container/../etc")
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("container path of exactly MaxContainerPathLen should pass", func() {
			_, _, err := parseMountEntry("/data/a:/container/" + strings.Repeat("a", mount.MaxContainerPathLen-11))
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("too long container path should return error", func() {
			_, _, err := parseMountEntry("/data/a:/container/" + strings.Repeat("a", mount.MaxContainerPathLen-10))
			convey.So(err, convey.ShouldBeError)
		})
	})
}

// TestBuildMountEntry tests the function buildMountEntry
func TestBuildMountEntry(t *testing.T) {
	convey.Convey("test buildMountEntry", t, func() {
		convey.Convey("empty container path keeps host path only", func() {
			convey.So(buildMountEntry("/data/a", ""), convey.ShouldEqual, "/data/a")
		})
		convey.Convey("mapping entry should be reassembled", func() {
			convey.So(buildMountEntry("/data/a", "/container/a"),
				convey.ShouldEqual, "/data/a:/container/a")
		})
	})
}

// TestReadMountConfigGlobMapping verifies glob entries cannot carry a mapping
func TestReadMountConfigGlobMapping(t *testing.T) {
	patch := gomonkey.ApplyFuncReturn(mindxcheckutils.RealFileChecker, "", nil)
	defer patch.Reset()

	convey.Convey("test readMountConfig rejects glob entries with mapping", t, func() {
		tmpDir := t.TempDir()
		listPath := filepath.Join(tmpDir, "base.list")
		if err := os.WriteFile(listPath, []byte(tmpDir+"/src/*:/data/m\n"), 0600); err != nil {
			t.Fatalf("write list failed: %v", err)
		}
		_, _, err := readMountConfig(tmpDir, "base")
		convey.So(err, convey.ShouldBeError)
	})
}

func TestReadMountConfigFailures(t *testing.T) {
	for _, test := range []struct{ stage, want string }{
		{"path", "failed to assemble base config file path"},
		{"check", testError.Error()},
		{"open", "failed to open base configuration file"},
	} {
		t.Run(test.stage, func(t *testing.T) {
			openCalls := 0
			patches := gomonkey.NewPatches().
				ApplyFuncReturn(mindxcheckutils.RealFileChecker, "", nil).
				ApplyFunc(os.Open, func(string) (*os.File, error) {
					openCalls++
					return nil, testError
				})
			defer patches.Reset()
			switch test.stage {
			case "path":
				patches.ApplyFuncReturn(filepath.Abs, "", testError)
			case "check":
				patches.ApplyFuncReturn(mindxcheckutils.RealFileChecker, "", testError)
			}
			files, dirs, err := readMountConfig("/config", "base")
			assert.ErrorContains(t, err, test.want)
			assert.Nil(t, files)
			assert.Nil(t, dirs)
			wantOpenCalls := 0
			if test.stage == "open" {
				wantOpenCalls = 1
			}
			assert.Equal(t, wantOpenCalls, openCalls)
		})
	}
}

func TestReadMountConfigClassifiesEntries(t *testing.T) {
	patches := gomonkey.ApplyFuncReturn(mindxcheckutils.RealFileChecker, "", nil)
	defer patches.Reset()
	dir := t.TempDir()
	sourcePath := filepath.Join(dir, "driver.so")
	if err := os.WriteFile(sourcePath, []byte("driver"), 0600); err != nil {
		t.Fatal(err)
	}
	contents := "# driver list\n\n" + sourcePath + ":/container/driver.so\n" + dir + "\n" + filepath.Join(dir, "missing") + "\n"
	if err := os.WriteFile(filepath.Join(dir, "base.list"), []byte(contents), 0600); err != nil {
		t.Fatal(err)
	}
	files, dirs, err := readMountConfig(dir, "base")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, []string{sourcePath + ":/container/driver.so"}, files)
	assert.Equal(t, []string{dir}, dirs)
}

func TestParseUBDrvMount(t *testing.T) {
	convey.Convey("test parseUBDrvMount", t, func() {
		convey.Convey("should return True for 'True' input", func() {
			result, err := parseUBDrvMount("True")
			convey.So(err, convey.ShouldBeNil)
			convey.So(result, convey.ShouldBeTrue)
		})
		convey.Convey("should return False for 'False' input", func() {
			result, err := parseUBDrvMount("False")
			convey.So(err, convey.ShouldBeNil)
			convey.So(result, convey.ShouldBeFalse)
		})
		convey.Convey("should return True for empty input", func() {
			result, err := parseUBDrvMount("")
			convey.So(err, convey.ShouldBeNil)
			convey.So(result, convey.ShouldBeTrue)
		})
		convey.Convey("should return error for invalid input", func() {
			_, err := parseUBDrvMount("invalid")
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestShouldMountUBDriverFiles tests shouldMountUBDriverFiles
func TestShouldMountUBDriverFiles(t *testing.T) {
	convey.Convey("test shouldMountUBDriverFiles", t, func() {
		convey.Convey("should return false when mountUBDrv is False", func() {
			result := shouldMountUBDriverFiles(false)
			convey.So(result, convey.ShouldBeFalse)
		})
		convey.Convey("should return true when UB config file exists", func() {
			patches := gomonkey.ApplyFuncReturn(os.Stat, mockFileInfo{}, nil)
			defer patches.Reset()
			result := shouldMountUBDriverFiles(true)
			convey.So(result, convey.ShouldBeTrue)
		})
		convey.Convey("should return false when UB config file does not exist", func() {
			patches := gomonkey.ApplyFuncReturn(os.Stat, mockFileInfo{}, os.ErrNotExist)
			defer patches.Reset()
			result := shouldMountUBDriverFiles(true)
			convey.So(result, convey.ShouldBeFalse)
		})
		convey.Convey("should return false when stat returns other error", func() {
			patches := gomonkey.ApplyFuncReturn(os.Stat, mockFileInfo{}, os.ErrPermission)
			defer patches.Reset()
			result := shouldMountUBDriverFiles(true)
			convey.So(result, convey.ShouldBeFalse)
		})
	})
}

type mockFileInfo struct {
	mode os.FileMode
	sys  interface{}
}

func (m mockFileInfo) Name() string       { return "mock" }
func (m mockFileInfo) Size() int64        { return 0 }
func (m mockFileInfo) Mode() os.FileMode  { return m.mode }
func (m mockFileInfo) ModTime() time.Time { return time.Now() }
func (m mockFileInfo) IsDir() bool        { return false }
func (m mockFileInfo) Sys() interface{}   { return m.sys }

func mockStat() *gomonkey.Patches {
	return gomonkey.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) {
		mockMountItemMap := map[string]mockFileInfo{
			topoDirPath: {mode: os.ModeDir},
		}
		return mockMountItemMap[name], nil
	})
}

func TestExpandGlobPathKeepsOrdinaryLinkedParentBehavior(t *testing.T) {
	dir := t.TempDir()
	realDir, linkedDir := filepath.Join(dir, "real"), filepath.Join(dir, "linked")
	if err := os.Mkdir(realDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(realDir, linkedDir); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(realDir, "driver.so"), []byte("driver"), 0644); err != nil {
		t.Fatal(err)
	}
	files, dirs := expandGlobPath(filepath.Join(linkedDir, "*.so"))
	if len(files) != 0 || len(dirs) != 0 {
		t.Fatalf("ordinary glob must retain its original parent comparison: files=%v dirs=%v", files, dirs)
	}
}

func TestDeduplicateMountEntries(t *testing.T) {
	tests := []struct {
		name    string
		entries []string
		want    []string
	}{
		{"empty", nil, []string{}},
		{"identical", []string{"/host/a:/container/a", "/host/b:/container/b", "/host/a:/container/a"},
			[]string{"/host/a:/container/a", "/host/b:/container/b"}},
		{"normalized", []string{"/host/a", "/host/a:/host/a", "/host//a:/host/./a"}, []string{"/host/a"}},
		{"distinct targets", []string{"/host/a:/container/a", "/host/a:/container/b"},
			[]string{"/host/a:/container/a", "/host/a:/container/b"}},
		{"distinct sources", []string{"/host/first:/container/a", "/host/second:/container/./a", "/host/first:/container/a"},
			[]string{"/host/first:/container/a", "/host/second:/container/./a"}},
		{"symlink spelling", []string{"/host/lib.so", "/host/lib.so.1", "/host/lib.so"},
			[]string{"/host/lib.so", "/host/lib.so.1"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files, dirs := deduplicateMountEntries(tt.entries, nil)
			assert.Equal(t, tt.want, files)
			assert.Empty(t, dirs)
		})
	}
}

func TestDeduplicateMountEntriesSharesFileAndDirectoryKeys(t *testing.T) {
	files, dirs := deduplicateMountEntries([]string{"/host/first:/container/item"},
		[]string{"/host/first:/container/item", "/host/second:/container/item"})
	assert.Equal(t, []string{"/host/first:/container/item"}, files)
	assert.Equal(t, []string{"/host/second:/container/item"}, dirs)
}
