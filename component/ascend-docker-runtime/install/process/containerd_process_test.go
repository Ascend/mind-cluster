/* Copyright(C) 2024. Huawei Technologies Co.,Ltd. All rights reserved.
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

package process

import (
	"context"
	"errors"
	"os"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/containerd/containerd/services/server/config"
	"github.com/pelletier/go-toml"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/hwlog"
	"ascend-docker-runtime/mindxcheckutils"
)

const (
	convertMapToTreeFailedStr = "convert map to tree failed, error: %v"
)

var testError = errors.New("test")

// TestContainerdProcess tests the function ContainerdProcess
func TestContainerdProcess(t *testing.T) {
	tests := getTestDockerProcessCases()
	initTestLog(t)
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			if tt.Name == "success case 4" {
				patch := gomonkey.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) {
					return &FileInfoMock{}, nil
				})
				defer patch.Reset()
				patchRealDirCheck := gomonkey.ApplyFunc(mindxcheckutils.RealDirChecker, func(path string,
					checkParent, allowLink bool) (string, error) {
					return "", nil
				})
				defer patchRealDirCheck.Reset()
				patchRealFileCheck := gomonkey.ApplyFunc(mindxcheckutils.RealFileChecker, func(path string,
					checkParent, allowLink bool, size int) (string, error) {
					return "", nil
				})
				defer patchRealFileCheck.Reset()
			}
			got, got1 := ContainerdProcess(tt.Command)
			if (got1 == nil) == tt.WantErr {
				t.Errorf("ContainerdProcess() got = %v, want %v", got, tt.WantErr)
			}
			if got != tt.WantResult {
				t.Errorf("ContainerdProcess() got1 = %v, want %v", got1, tt.WantResult)
			}
		})
	}
}

// TestContainerdProcess1 tests the function ContainerdProcess patch1
func TestContainerdProcess1(t *testing.T) {
	convey.Convey("test ContainerdProcess patch1", t, func() {
		convey.Convey("01-fail to check, return error", func() {
			patch := gomonkey.ApplyFunc(checkParamAndGetBehavior, func(action string, command []string) (bool, string) {
				return false, ""
			})
			defer patch.Reset()
			ret, err := ContainerdProcess([]string{"command"})
			if ret != "" || err == nil {
				t.Errorf("want , got %v", ret)
			}
		})
		convey.Convey("02-file not exists and fails to check, return error", func() {
			patches := gomonkey.ApplyFuncReturn(os.Stat, nil, os.ErrNotExist).
				ApplyFuncReturn(mindxcheckutils.RealDirChecker, "", testError).
				ApplyFuncReturn(checkParamAndGetBehavior, true, "test")
			defer patches.Reset()
			emptyStr := ""
			destFileTest := "aaa.txt.pid"
			cmds := []string{"test2", oldJson, destFileTest, emptyStr, emptyStr, emptyStr, emptyStr}
			_, err := ContainerdProcess(cmds)
			convey.So(err, convey.ShouldBeError)
		})
	})
}

// TestContainerdProcess2 tests the function ContainerdProcess patch2
func TestContainerdProcess2(t *testing.T) {
	emptyStr := ""
	destFileTest := "aaa.txt.pid"
	convey.Convey("test ContainerdProcess patch2", t, func() {
		convey.Convey("03-file exists and fails to check", func() {
			patches := gomonkey.ApplyFuncReturn(os.Stat, nil, nil).
				ApplyFuncReturn(mindxcheckutils.RealFileChecker, "", testError).
				ApplyFuncReturn(checkParamAndGetBehavior, true, "test3")
			defer patches.Reset()
			cmds := []string{"test3", oldJson, destFileTest, emptyStr, emptyStr, emptyStr, emptyStr}
			ret, err := ContainerdProcess(cmds)
			convey.So(ret, convey.ShouldEqual, "test3")
			convey.So(err, convey.ShouldBeError)
		})
	})
}

// TestContainerdProcess3 tests the function ContainerdProcess patch3
func TestContainerdProcess3(t *testing.T) {
	emptyStr := ""
	destFileTest := "aaa.txt.pid"
	convey.Convey("test ContainerdProcess patch3", t, func() {
		convey.Convey("04-file exists, file check pass, dir check fail, should return error", func() {
			patches := gomonkey.ApplyFuncReturn(os.Stat, nil, nil).
				ApplyFuncReturn(mindxcheckutils.RealFileChecker, "", nil).
				ApplyFuncReturn(checkParamAndGetBehavior, true, "test")
			defer patches.Reset()
			p4 := gomonkey.ApplyFuncReturn(mindxcheckutils.RealDirChecker, "", testError)
			defer p4.Reset()
			cmds := []string{"test", oldJson, destFileTest, emptyStr, emptyStr, emptyStr, emptyStr}
			_, err := ContainerdProcess(cmds)
			convey.So(err, convey.ShouldBeError)
		})
	})
}

// TestContainerdProcess4 tests the function ContainerdProcess patch4
func TestContainerdProcess4(t *testing.T) {
	emptyStr := ""
	destFileTest := "aaa.txt.pid"
	convey.Convey("test ContainerdProcess patch4", t, func() {
		convey.Convey("05-file exists, dir check pass, file check fail, should return error", func() {
			patches := gomonkey.ApplyFuncReturn(os.Stat, nil, nil).
				ApplyFuncReturn(mindxcheckutils.RealFileChecker, "", testError).
				ApplyFuncReturn(checkParamAndGetBehavior, true, "test").
				ApplyFuncReturn(mindxcheckutils.RealDirChecker, "", nil)
			defer patches.Reset()
			cmds := []string{"add", oldJson, destFileTest, emptyStr, emptyStr, emptyStr, emptyStr}
			_, err := ContainerdProcess(cmds)
			convey.So(err, convey.ShouldBeError)
		})
	})
}

// TestContainerdProcess5 tests the function ContainerdProcess patch5
func TestContainerdProcess5(t *testing.T) {
	emptyStr := ""
	destFileTest := "aaa.txt.pid"
	convey.Convey("test ContainerdProcess patch5", t, func() {
		convey.Convey("06-file exists, dir check pass, file check pass, should return nil", func() {
			patches := gomonkey.ApplyFuncReturn(os.Stat, nil, nil).
				ApplyFuncReturn(mindxcheckutils.RealFileChecker, "", nil).
				ApplyFuncReturn(checkParamAndGetBehavior, true, "test").
				ApplyFuncReturn(mindxcheckutils.RealDirChecker, "", nil).
				ApplyFuncReturn(editContainerdConfig, nil)
			defer patches.Reset()
			cmds := []string{"add", oldJson, destFileTest, emptyStr, emptyStr, emptyStr, emptyStr}
			_, err := ContainerdProcess(cmds)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func initTestLog(t *testing.T) {
	backups := 2
	logMaxAge := 365
	fileMaxSize := 2
	runLogConfig := hwlog.LogConfig{
		LogFileName: "./test/run.log",
		LogLevel:    0,
		MaxBackups:  backups,
		FileMaxSize: fileMaxSize,
		MaxAge:      logMaxAge,
	}
	if err := hwlog.InitRunLogger(&runLogConfig, context.Background()); err != nil {
		t.Fatalf("hwlog init failed, error is %v", err)
	}
}

// TestEditContainerdConfig tests the function editContainerdConfig
func TestEditContainerdConfig(t *testing.T) {
	tests := []struct {
		name    string
		args    *commandArgs
		wantErr bool
	}{
		{
			name: "v2  failed case 1",
			args: &commandArgs{
				action:     addCommand,
				cgroupInfo: cgroupV2InfoStr,
			},
			wantErr: true,
		},
		{
			name: "v1  failed case 2",
			args: &commandArgs{
				action:     addCommand,
				cgroupInfo: "",
			},
			wantErr: true,
		},
	}
	patch := gomonkey.ApplyFunc(config.LoadConfig, func(path string, out *config.Config) error {
		return nil
	})
	defer patch.Reset()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := editContainerdConfig(tt.args); (err != nil) != tt.wantErr {
				t.Errorf("editContainerdConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestEditContainerdConfig1 tests the function editContainerdConfig patch1
func TestEditContainerdConfig1(t *testing.T) {
	convey.Convey("test editContainerdConfig patch1", t, func() {
		convey.Convey("01-loadConfig failed, should return error", func() {
			patches := gomonkey.ApplyFunc(config.LoadConfig, func(path string, out *config.Config) error {
				return testError
			})
			defer patches.Reset()
			err := editContainerdConfig(&commandArgs{srcFilePath: "", runtimeFilePath: "", destFilePath: "",
				action: addCommand, cgroupInfo: cgroupV2InfoStr, osName: "", osVersion: ""})
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("02-changeCgroupV2BinaryNameConfig failed", func() {
			patches := gomonkey.ApplyFunc(changeCgroupV2BinaryNameConfig, func(cfg *config.Config, binaryName string) error {
				return testError
			})
			defer patches.Reset()
			err := editContainerdConfig(&commandArgs{srcFilePath: "", runtimeFilePath: "", destFilePath: "",
				action: addCommand, cgroupInfo: cgroupV2InfoStr, osName: "", osVersion: ""})
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("03-changeCgroupV1Config failed", func() {
			patches := gomonkey.ApplyFunc(changeCgroupV1Config, func(cfg *config.Config,
				runtimeValue, runtimeType string) error {
				return testError
			})
			defer patches.Reset()
			err := editContainerdConfig(&commandArgs{srcFilePath: "", runtimeFilePath: "", destFilePath: "",
				action: addCommand, cgroupInfo: "", osName: "", osVersion: ""})
			convey.So(err, convey.ShouldBeError)
		})
	})
}

// TestEditContainerdConfig2 tests the function editContainerdConfig patch2
func TestEditContainerdConfig2(t *testing.T) {
	convey.Convey("test editContainerdConfig patch2", t, func() {
		convey.Convey("04-writeContainerdConfigToFile fail, should return error", func() {
			patches := gomonkey.ApplyFuncReturn(config.LoadConfig, nil).
				ApplyFuncReturn(changeCgroupV2BinaryNameConfig, nil).
				ApplyFuncReturn(writeContainerdConfigToFile, testError)
			defer patches.Reset()
			err := editContainerdConfig(&commandArgs{srcFilePath: "", runtimeFilePath: "", destFilePath: "",
				action: addCommand, cgroupInfo: cgroupV2InfoStr, osName: "", osVersion: ""})
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("05-all pass, should return nil", func() {
			patches := gomonkey.ApplyFuncReturn(config.LoadConfig, nil).
				ApplyFuncReturn(changeCgroupV2BinaryNameConfig, nil).
				ApplyFuncReturn(writeContainerdConfigToFile, nil)
			defer patches.Reset()
			err := editContainerdConfig(&commandArgs{srcFilePath: "", runtimeFilePath: "", destFilePath: "",
				action: addCommand, cgroupInfo: cgroupV2InfoStr, osName: "", osVersion: ""})
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func defaultConfigV2(t *testing.T) *config.Config {
	testMap := map[string]interface{}{
		containerdKey: map[string]interface{}{
			runtimesKey: map[string]interface{}{
				runcKey: map[string]interface{}{
					runcOptionsKey: map[string]interface{}{
						binaryNameKey: "",
					},
				},
			},
		},
	}
	testTree, err := toml.TreeFromMap(testMap)
	if err != nil {
		t.Fatalf(convertMapToTreeFailedStr, err)
	}
	return &config.Config{
		Plugins: map[string]toml.Tree{
			v1RuntimeTypeFisrtLevelPlugin: *testTree,
		},
	}
}

// TestChangeCgroupV2BinaryNameConfig tests the function changeCgroupV2BinaryNameConfig
func TestChangeCgroupV2BinaryNameConfig(t *testing.T) {
	tests := []struct {
		name       string
		cfg        *config.Config
		binaryName string
		wantErr    bool
	}{
		{
			name:    "success case 1",
			cfg:     defaultConfigV2(t),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := changeCgroupV2BinaryNameConfig(tt.cfg, tt.binaryName); (err != nil) != tt.wantErr {
				t.Errorf("changeCgroupV2BinaryNameConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestChangeCgroupV2BinaryNameConfig1 tests the function changeCgroupV2BinaryNameConfig patch1
func TestChangeCgroupV2BinaryNameConfig1(t *testing.T) {
	convey.Convey("test changeCgroupV2BinaryNameConfig patch1", t, func() {
		convey.Convey("01-getMap runtimes error, should return error", func() {
			patches := gomonkey.ApplyFuncReturn(getMap, nil, testError)
			defer patches.Reset()
			err := changeCgroupV2BinaryNameConfig(defaultConfigV2(t), "")
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("02-getMap runc error, should return error", func() {
			patches := gomonkey.ApplyFunc(getMap, func(input interface{}, key string) (interface{}, error) {
				if key == runcKey {
					return nil, testError
				} else {
					return nil, nil
				}
			})
			defer patches.Reset()
			err := changeCgroupV2BinaryNameConfig(defaultConfigV2(t), "")
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("03-getMap runcOptions error, should return error", func() {
			patches := gomonkey.ApplyFunc(getMap, func(input interface{}, key string) (interface{}, error) {
				if key == runcOptionsKey {
					return nil, testError
				} else {
					return nil, nil
				}
			})
			defer patches.Reset()
			err := changeCgroupV2BinaryNameConfig(defaultConfigV2(t), "")
			convey.So(err, convey.ShouldBeError)
		})
	})
}

// TestChangeCgroupV2BinaryNameConfig2 tests the function changeCgroupV2BinaryNameConfig patch2
func TestChangeCgroupV2BinaryNameConfig2(t *testing.T) {
	convey.Convey("test changeCgroupV2BinaryNameConfig patch1", t, func() {
		convey.Convey("01-getMap runtimes error, should return error", func() {
			patches := gomonkey.ApplyFuncReturn(getMap, nil, nil)
			defer patches.Reset()
			err := changeCgroupV2BinaryNameConfig(defaultConfigV2(t), "")
			convey.So(err, convey.ShouldBeError)
		})
	})
}

func defaultConfigV1(t *testing.T) *config.Config {
	testMapV1RuntimeTypeFisrtLevelPlugin := map[string]interface{}{
		containerdKey: map[string]interface{}{
			runtimesKey: map[string]interface{}{
				runcKey: map[string]interface{}{},
			},
		},
	}
	testTreeV1RuntimeTypeFisrtLevelPlugin, err := toml.TreeFromMap(testMapV1RuntimeTypeFisrtLevelPlugin)
	if err != nil {
		t.Fatalf(convertMapToTreeFailedStr, err)
	}
	testMapV1RuntimeType := map[string]interface{}{}
	testTreeV1RuntimeType, err := toml.TreeFromMap(testMapV1RuntimeType)
	if err != nil {
		t.Fatalf(convertMapToTreeFailedStr, err)
	}
	return &config.Config{
		Plugins: map[string]toml.Tree{
			v1RuntimeType:                 *testTreeV1RuntimeType,
			v1RuntimeTypeFisrtLevelPlugin: *testTreeV1RuntimeTypeFisrtLevelPlugin,
		},
	}
}

// TestChangeCgroupV1Config tests the function changeCgroupV1Config
func TestChangeCgroupV1Config(t *testing.T) {
	tests := []struct {
		name         string
		cfg          *config.Config
		runtimeValue string
		runtimeType  string
		wantErr      bool
	}{
		{
			name:    "success case 1",
			cfg:     defaultConfigV1(t),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := changeCgroupV1Config(tt.cfg, tt.runtimeValue, tt.runtimeType); (err != nil) != tt.wantErr {
				t.Errorf("changeCgroupV1Config() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestGetMap tests the function getMap
func TestGetMap(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		key     string
		want    interface{}
		wantErr bool
	}{
		{
			name:    "invalid case 1",
			input:   "",
			wantErr: true,
		},
		{
			name:    "invalid case 2",
			input:   map[string]interface{}{},
			key:     v1NeedChangeKeyRuntimeType,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getMap(tt.input, tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("getMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getMap() got = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestWriteContainerdConfigToFile tests the function writeContainerdConfigToFile
func TestWriteContainerdConfigToFile(t *testing.T) {
	tests := []struct {
		name         string
		cfg          config.Config
		destFilePath string
		wantErr      bool
	}{
		{
			name:    "marshal failed case 1",
			cfg:     config.Config{},
			wantErr: true,
		},
		{
			name:         "success case 2",
			cfg:          config.Config{},
			destFilePath: "config.toml",
			wantErr:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := writeContainerdConfigToFile(tt.cfg, tt.destFilePath); (err != nil) != tt.wantErr {
				t.Errorf("writeContainerdConfigToFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestWriteContainerdConfigToFile1 tests the function writeContainerdConfigToFile patch1
func TestWriteContainerdConfigToFile1(t *testing.T) {
	convey.Convey("test writeContainerdConfigToFile patch1", t, func() {
		convey.Convey("01-marshal error, should return error", func() {
			patches := gomonkey.ApplyFuncReturn(toml.Marshal, []byte{}, testError)
			defer patches.Reset()
			err := writeContainerdConfigToFile(config.Config{}, "")
			convey.So(err, convey.ShouldBeError)
		})
	})
}

func emptyConfig() *config.Config {
	return &config.Config{}
}

func pluginConfig() *config.Config {
	return &config.Config{
		Plugins: map[string]toml.Tree{v1RuntimeTypeFisrtLevelPlugin: toml.Tree{}},
	}
}

type EmptyStruct struct{}

// TestChangeCgroupV1RuntimeTypeConfig tests the function changeCgroupV1RuntimeTypeConfig
func TestChangeCgroupV1RuntimeTypeConfig(t *testing.T) {
	convey.Convey("Test changeCgroupV1RuntimeTypeConfig", t, func() {
		convey.Convey("01-plugin no ok, should return error", func() {
			err := changeCgroupV1RuntimeTypeConfig(emptyConfig(), "")
			convey.ShouldBeError(err)
		})
		convey.Convey("02-get map runtimesKey error, should return error", func() {
			patch1 := gomonkey.ApplyFunc(getMap, func(input interface{}, key string) (interface{}, error) {
				if key == runtimesKey {
					return nil, testError
				} else {
					return nil, nil
				}
			})
			defer patch1.Reset()
			tr := toml.Tree{}
			m := make(map[string]interface{})
			m[containerdKey] = EmptyStruct{}
			patch2 := gomonkey.ApplyMethodReturn(&tr, "ToMap", m)
			defer patch2.Reset()
			err := changeCgroupV1RuntimeTypeConfig(pluginConfig(), "")
			convey.ShouldBeError(err)
		})
	})
}

// TestChangeCgroupV1RuntimeTypeConfig1 tests the function changeCgroupV1RuntimeTypeConfig patch1
func TestChangeCgroupV1RuntimeTypeConfig1(t *testing.T) {
	convey.Convey("Test changeCgroupV1RuntimeTypeConfig", t, func() {
		m := make(map[string]interface{})
		m[containerdKey] = EmptyStruct{}
		convey.Convey("03-get map runc error, should return error", func() {
			patch1 := gomonkey.ApplyFunc(getMap, func(input interface{}, key string) (interface{}, error) {
				if key == runcKey {
					return nil, testError
				} else {
					return nil, nil
				}
			})
			defer patch1.Reset()
			tr := toml.Tree{}
			patch2 := gomonkey.ApplyMethodReturn(&tr, "ToMap", m)
			defer patch2.Reset()
			err := changeCgroupV1RuntimeTypeConfig(pluginConfig(), "")
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("03-runc config assert error, should return error", func() {
			patch1 := gomonkey.ApplyFunc(getMap, func(input interface{}, key string) (interface{}, error) {
				return nil, nil
			})
			defer patch1.Reset()
			tr := toml.Tree{}
			patch2 := gomonkey.ApplyMethodReturn(&tr, "ToMap", m)
			defer patch2.Reset()
			err := changeCgroupV1RuntimeTypeConfig(pluginConfig(), "")
			convey.So(err, convey.ShouldBeError)
		})
	})
}
