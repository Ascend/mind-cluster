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
	"bufio"
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

	"github.com/agiledragon/gomonkey/v2"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/prashantv/gostub"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"

	"ascend-common/api"
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
	stub := gostub.StubFunc(&getContainerConfig, &conCfg, nil)
	defer stub.Reset()
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
	stub := gostub.StubFunc(&getContainerConfig, &conCfg, nil)
	defer stub.Reset()
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
	stub := gostub.StubFunc(&getContainerConfig, &conCfg, nil)
	defer stub.Reset()
	stub.Stub(&ascendDockerCliName, "")
	stub.StubFunc(&doExec, nil)
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
	stub := gostub.StubFunc(&getContainerConfig, &conCfg, nil)
	defer stub.Reset()
	stub.Stub(&ascendDockerCliName, "clii")
	stub.Stub(&defaultAscendDockerCliName, "clii")
	stub.StubFunc(&doExec, nil)
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
			ApplyFuncReturn(readConfigsOfDir, []string{testStr}, []string{testStr}, nil)
		defer patches.Reset()
		convey.Convey("01-parseRuntimeOptions error, should return error", func() {
			patch := gomonkey.ApplyFuncReturn(parseRuntimeOptions, []string{testStr}, testError)
			defer patch.Reset()
			err := DoPrestartHook()
			convey.So(err, convey.ShouldBeError)
		})
		patches.ApplyFuncReturn(parseRuntimeOptions, []string{testStr}, nil)
		convey.Convey("02-parseSoftLinkMode error, should return error", func() {
			patch := gomonkey.ApplyFuncReturn(parseSoftLinkMode, "", testError)
			defer patch.Reset()
			err := DoPrestartHook()
			convey.So(err, convey.ShouldBeError)
		})
		patches.ApplyFuncReturn(parseSoftLinkMode, testStr, nil)
		convey.Convey("03-Executable error, should return error", func() {
			patch := gomonkey.ApplyFuncReturn(os.Executable, "", testError)
			defer patch.Reset()
			err := DoPrestartHook()
			convey.So(err, convey.ShouldBeError)
		})
		patches.ApplyFuncReturn(os.Executable, testStr, nil)
		convey.Convey("04-Stat error, should return error", func() {
			patch := gomonkey.ApplyFuncReturn(os.Stat, fileInfoMock{}, testError)
			defer patch.Reset()
			err := DoPrestartHook()
			convey.So(err, convey.ShouldBeError)
		})
		patches.ApplyFuncReturn(os.Stat, fileInfoMock{}, nil).
			ApplyFuncReturn(mindxcheckutils.RealFileChecker, testStr, nil).
			ApplyFuncReturn(getArgs, []string{testStr})
		convey.Convey("05-ChangeRuntimeLogMode error, should return error", func() {
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
			ApplyFuncReturn(readConfigsOfDir, []string{testStr}, []string{testStr}, nil).
			ApplyFuncReturn(parseRuntimeOptions, []string{testStr}, nil).
			ApplyFuncReturn(parseSoftLinkMode, testStr, nil).
			ApplyFuncReturn(os.Executable, testStr, nil).
			ApplyFuncReturn(os.Stat, fileInfoMock{}, nil).
			ApplyFuncReturn(mindxcheckutils.RealFileChecker, testStr, nil).
			ApplyFuncReturn(getArgs, []string{testStr}).
			ApplyFuncReturn(mindxcheckutils.ChangeRuntimeLogMode, nil)
		defer patches.Reset()
		convey.Convey("06-doExec error, should return error", func() {
			patch := gomonkey.ApplyFuncReturn(doExec, testError)
			defer patch.Reset()
			err := DoPrestartHook()
			convey.So(err, convey.ShouldBeError)
		})
		patches.ApplyFuncReturn(doExec, nil)
		convey.Convey("07-success, should return nil", func() {
			err := DoPrestartHook()
			convey.So(err, convey.ShouldBeNil)
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

	stub := gostub.Stub(&containerConfigInputStream, stateFile)
	defer stub.Reset()

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
			patch := gomonkey.ApplyFuncReturn(filepath.Abs, testStr, false)
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
			got, got1, err := readConfigsOfDir(tt.dir, tt.configs)
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
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "DoPrestartHook success case 1",
			wantErr: true,
		},
	}
	conCfg := containerConfig{
		Pid:    pidSample,
		Rootfs: ".",
		Env:    []string{api.AscendVisibleDevicesEnv + "=0-3,5,7", api.AscendRuntimeMountsEnv + "=a"},
	}
	stub := gostub.StubFunc(&getContainerConfig, &conCfg, nil)
	defer stub.Reset()
	patch := gomonkey.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) {
		return &fileInfoMock{}, nil
	})
	defer patch.Reset()
	patchReadConfigsOfDir := gomonkey.ApplyFunc(readConfigsOfDir,
		func(dir string, configs []string) ([]string, []string, error) {
			return []string{}, []string{}, nil
		})
	defer patchReadConfigsOfDir.Reset()
	patchSize := gomonkey.ApplyMethod(reflect.TypeOf(&fileInfoMock{}), "Mode", func(f *fileInfoMock) fs.FileMode {
		return fs.ModeDir
	})
	defer patchSize.Reset()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := DoPrestartHook()
			if (err == nil) == tt.wantErr {
				t.Errorf("DoPrestartHook() got = %v, want %v", err, tt.wantErr)
			}
		})
	}
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

// TestReadMountConfig tests the function readMountConfig
func TestReadMountConfig(t *testing.T) {
	convey.Convey("test readMountConfig", t, func() {
		convey.Convey("01-get abs path failed, should return error", func() {
			patch := gomonkey.ApplyFuncReturn(filepath.Abs, "", testError)
			defer patch.Reset()
			_, _, err := readMountConfig("", "")
			convey.So(err, convey.ShouldBeError)
		})
		patches := gomonkey.ApplyFuncReturn(filepath.Abs, "", nil)
		defer patches.Reset()
		convey.Convey("02-stat error, should return error", func() {
			patch := gomonkey.ApplyFuncReturn(os.Stat, fileInfoMock{}, testError)
			defer patch.Reset()
			_, _, err := readMountConfig("", "")
			convey.So(err, convey.ShouldBeError)
		})
		patches.ApplyFuncReturn(os.Stat, fileInfoMock{}, nil).
			ApplyFuncReturn(mindxcheckutils.RealFileChecker, "", nil)
		convey.Convey("03-file info is not regular, should return error", func() {
			var fm os.FileMode
			patch := gomonkey.ApplyMethodReturn(fm, "IsRegular", false)
			defer patch.Reset()
			_, _, err := readMountConfig("", "")
			convey.So(err, convey.ShouldBeError)
		})
		var fm os.FileMode
		patches.ApplyMethodReturn(fm, "IsRegular", true)
		convey.Convey("04-open file error, should return error", func() {
			patch := gomonkey.ApplyFuncReturn(os.Open, &os.File{}, testError)
			defer patch.Reset()
			_, _, err := readMountConfig("", "")
			convey.So(err, convey.ShouldBeError)
		})
		patches.ApplyFuncReturn(os.Open, &os.File{}, nil)
		convey.Convey("05-success, should return nil", func() {
			scanNum := 0
			patch := gomonkey.ApplyMethod(&bufio.Scanner{}, "Scan",
				func(scanner *bufio.Scanner) bool {
					scanNum++
					if scanNum > 1 {
						return false
					} else {
						return true
					}
				})
			defer patch.Reset()
			_, _, err := readMountConfig("", "")
			convey.So(err, convey.ShouldBeNil)
		})
	})
}
