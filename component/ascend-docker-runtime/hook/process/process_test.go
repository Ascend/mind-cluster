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
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/assert"
)

const (
	pidSample                              = 123
	fileMode0600               os.FileMode = 0600
	ascendVisibleDeviceTestStr             = "ASCEND_VISIBLE_DEVICES=0-3,5,7"
	configFile                             = "config.json"
	strRepeatTimes                         = 129
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
		Env:    []string{"ASCEND_VISIBLE_DEVICES=0l-3,5,7"},
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
			"ASCEND_VISIBLE_DEVICES=0",
			"ASCEND_RUNTIME_OPTIONS=VIRTUAL,NODRV",
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

// TestGetValueByKeyCase1 test the function getValueByKey
func TestGetValueByKeyCase1(t *testing.T) {
	data := []string{ascendVisibleDeviceTestStr}
	word := "ASCEND_VISIBLE_DEVICES"
	expectVal := "0-3,5,7"
	actualVal := getValueByKey(data, word)
	if actualVal != expectVal {
		t.Fail()
	}
}

// TestGetValueByKeyCase2 test the function getValueByKey
func TestGetValueByKeyCase2(t *testing.T) {
	data := []string{"ASCEND_VISIBLE_DEVICES"}
	word := "ASCEND_VISIBLE_DEVICES"
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

// fileInfoMock is used to test
type fileInfoMock struct {
	os.FileInfo
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
		Env:    []string{"ASCEND_VISIBLE_DEVICES=0-3,5,7", "ASCEND_RUNTIME_MOUNTS=a"},
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
