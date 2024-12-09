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
	"os"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/containerd/containerd/services/server/config"
	"github.com/pelletier/go-toml"

	"ascend-common/common-utils/hwlog"
	"ascend-docker-runtime/mindxcheckutils"
)

const (
	convertMapToTreeFailedStr = "convert map to tree failed, error: %v"
)

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
		name            string
		srcFilePath     string
		runtimeFilePath string
		destFilePath    string
		action          string
		cgroupInfo      string
		wantErr         bool
	}{
		{
			name:       "v2  failed case 1",
			action:     addCommand,
			cgroupInfo: cgroupV2InfoStr,
			wantErr:    true,
		},
		{
			name:       "v1  failed case 2",
			action:     addCommand,
			cgroupInfo: "",
			wantErr:    true,
		},
	}
	patch := gomonkey.ApplyFunc(config.LoadConfig, func(path string, out *config.Config) error {
		return nil
	})
	defer patch.Reset()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := editContainerdConfig(tt.srcFilePath, tt.runtimeFilePath, tt.destFilePath,
				tt.action, tt.cgroupInfo); (err != nil) != tt.wantErr {
				t.Errorf("editContainerdConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestChangeCgroupV2BinaryNameConfig tests the function changeCgroupV2BinaryNameConfig
func TestChangeCgroupV2BinaryNameConfig(t *testing.T) {
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
	tests := []struct {
		name       string
		cfg        *config.Config
		binaryName string
		wantErr    bool
	}{
		{
			name: "success case 1",
			cfg: &config.Config{
				Plugins: map[string]toml.Tree{
					v1RuntimeTypeFisrtLevelPlugin: *testTree,
				},
			},
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

// TestChangeCgroupV1Config tests the function changeCgroupV1Config
func TestChangeCgroupV1Config(t *testing.T) {
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
	tests := []struct {
		name         string
		cfg          *config.Config
		runtimeValue string
		runtimeType  string
		wantErr      bool
	}{
		{
			name: "success case 1",
			cfg: &config.Config{
				Plugins: map[string]toml.Tree{
					v1RuntimeType:                 *testTreeV1RuntimeType,
					v1RuntimeTypeFisrtLevelPlugin: *testTreeV1RuntimeTypeFisrtLevelPlugin,
				},
			},
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
