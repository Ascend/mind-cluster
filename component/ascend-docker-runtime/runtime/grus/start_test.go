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

// Package grus, unit tests for start
package grus

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/opencontainers/runtime-spec/specs-go"

	"ascend-docker-runtime/runtime/common"
	"ascend-docker-runtime/runtime/grus/runtime"
)

func TestReadConfigJson(t *testing.T) {
	tests := []struct {
		name        string
		prepareFile func(path string) error
		expectedErr bool
	}{
		{
			name: "file not found",
			prepareFile: func(path string) error {
				// 不创建文件
				return nil
			},
			expectedErr: true,
		},
		{
			name: "invalid JSON",
			prepareFile: func(path string) error {
				return os.WriteFile(path, []byte("invalid json"), 0644)
			},
			expectedErr: true,
		},
		{
			name: "valid JSON",
			prepareFile: func(path string) error {
				config := specs.Spec{
					Version: "1.0.2",
				}
				data, err := json.Marshal(config)
				if err != nil {
					return err
				}
				return os.WriteFile(path, data, 0644)
			},
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir, err := os.MkdirTemp("", "read-config-test")
			if err != nil {
				t.Fatalf("Failed to create temp directory: %v", err)
			}
			defer os.RemoveAll(tempDir)

			configPath := filepath.Join(tempDir, "config.json")
			if err := tt.prepareFile(configPath); err != nil {
				t.Fatalf("Failed to prepare file: %v", err)
			}

			_, err = readConfigJson(configPath)

			if (err != nil) != tt.expectedErr {
				t.Fatalf("Expected error: %v, got: %v", tt.expectedErr, err != nil)
			}
		})
	}
}

func TestReadOCIConfig(t *testing.T) {
	tests := []struct {
		name        string
		bundlePath  string
		prepareFile func(bundlePath string) error
		expectedErr bool
	}{
		{
			name:       "bundle path is empty and getwd fails",
			bundlePath: "",
			prepareFile: func(bundlePath string) error {
				// 不做任何操作，模拟 getwd 失败
				return nil
			},
			expectedErr: true,
		},
		{
			name:       "bundle path exists but config.json not found",
			bundlePath: "test-bundle",
			prepareFile: func(bundlePath string) error {
				return os.MkdirAll(bundlePath, 0755)
			},
			expectedErr: true,
		},
		{
			name:       "bundle path exists with valid config.json",
			bundlePath: "test-bundle",
			prepareFile: func(bundlePath string) error {
				if err := os.MkdirAll(bundlePath, 0755); err != nil {
					return err
				}
				config := specs.Spec{
					Version: "1.0.2",
				}
				data, err := json.Marshal(config)
				if err != nil {
					return err
				}
				return os.WriteFile(filepath.Join(bundlePath, "config.json"), data, 0644)
			},
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir, err := os.MkdirTemp("", "read-oci-config-test")
			if err != nil {
				t.Fatalf("Failed to create temp directory: %v", err)
			}
			defer os.RemoveAll(tempDir)

			bundlePath := tt.bundlePath
			if bundlePath != "" {
				bundlePath = filepath.Join(tempDir, bundlePath)
			}

			if err := tt.prepareFile(bundlePath); err != nil {
				t.Fatalf("Failed to prepare file: %v", err)
			}

			_, err = readOCIConfig(bundlePath)

			if (err != nil) != tt.expectedErr {
				t.Fatalf("Expected error: %v, got: %v", tt.expectedErr, err != nil)
			}
		})
	}
}

func TestAddEnv(t *testing.T) {
	tests := []struct {
		name           string
		spec           *specs.Spec
		key            string
		env            string
		expectedEnvs   []string
	}{
		{
			name:         "spec is nil",
			spec:         nil,
			key:          "TEST_KEY",
			env:          "TEST_KEY=test_value",
			expectedEnvs: nil,
		},
		{
			name: "spec.Process is nil",
			spec: &specs.Spec{},
			key:  "TEST_KEY",
			env:  "TEST_KEY=test_value",
			expectedEnvs: nil,
		},
		{
			name: "spec.Process.Env is nil",
			spec: &specs.Spec{
				Process: &specs.Process{},
			},
			key:          "TEST_KEY",
			env:          "TEST_KEY=test_value",
			expectedEnvs: nil,
		},
		{
			name: "add new env",
			spec: &specs.Spec{
				Process: &specs.Process{
					Env: []string{"EXISTING_KEY=existing_value"},
				},
			},
			key:  "NEW_KEY",
			env:  "NEW_KEY=new_value",
			expectedEnvs: []string{"EXISTING_KEY=existing_value", "NEW_KEY=new_value"},
		},
		{
			name: "update existing env",
			spec: &specs.Spec{
				Process: &specs.Process{
					Env: []string{"EXISTING_KEY=old_value", "OTHER_KEY=other_value"},
				},
			},
			key:  "EXISTING_KEY",
			env:  "EXISTING_KEY=new_value",
			expectedEnvs: []string{"EXISTING_KEY=new_value", "OTHER_KEY=other_value"},
		},
		{
			name: "empty env list",
			spec: &specs.Spec{
				Process: &specs.Process{
					Env: []string{},
				},
			},
			key:          "NEW_KEY",
			env:          "NEW_KEY=new_value",
			expectedEnvs: []string{"NEW_KEY=new_value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addEnv(tt.spec, tt.key, tt.env)

			if tt.spec == nil || tt.spec.Process == nil || tt.spec.Process.Env == nil {
				return
			}

			if len(tt.spec.Process.Env) != len(tt.expectedEnvs) {
				t.Fatalf("Expected %d env vars, got %d", len(tt.expectedEnvs), len(tt.spec.Process.Env))
			}

			for i, expectedEnv := range tt.expectedEnvs {
				if tt.spec.Process.Env[i] != expectedEnv {
					t.Fatalf("Expected env[%d]=%s, got %s", i, expectedEnv, tt.spec.Process.Env[i])
				}
			}
		})
	}
}

func TestSstart(t *testing.T) {
	tests := []struct {
		name              string
		args              *common.Args
		mockRuntimeState  bool
		mockExecRunc      bool
		mockReadOCIConfig bool
		readOCIConfigErr  bool
		hasSnapshotEnv    bool
		expectedErr       bool
	}{
		{
			name: "bundle is empty and state fails",
			args: &common.Args{
				ContainerID: "test-container",
				Root:        "/test/root",
				Bundle:      "",
			},
			mockRuntimeState: true,
			mockExecRunc:     true,
			expectedErr:      false,
		},
		{
			name: "readOCIConfig fails",
			args: &common.Args{
				ContainerID: "test-container",
				Root:        "/test/root",
				Bundle:      "/test/bundle",
			},
			mockReadOCIConfig: true,
			readOCIConfigErr:  true,
			mockExecRunc:      true,
			expectedErr:       false,
		},
		{
			name: "has snapshot image path env",
			args: &common.Args{
				ContainerID: "test-container",
				Root:        "/test/root",
				Bundle:      "/test/bundle",
			},
			mockReadOCIConfig: true,
			hasSnapshotEnv:    true,
			expectedErr:       false,
		},
		{
			name: "no snapshot image path env",
			args: &common.Args{
				ContainerID: "test-container",
				Root:        "/test/root",
				Bundle:      "/test/bundle",
			},
			mockReadOCIConfig: true,
			hasSnapshotEnv:    false,
			mockExecRunc:      true,
			expectedErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patches := gomonkey.NewPatches()
			defer patches.Reset()

			if tt.mockRuntimeState {
				patches.ApplyMethodFunc(&runtime.RuncRuntime{}, "State", func(id string) (*runtime.StateInfo, error) {
					return &runtime.StateInfo{
						Bundle: "/test/bundle",
					}, nil
				})
			}

			if tt.mockExecRunc {
				patches.ApplyFunc(common.ExecRunc, func() error {
					return nil
				})
			}

			if tt.mockReadOCIConfig {
				var spec *specs.Spec
				var err error
				if tt.readOCIConfigErr {
					err = os.ErrNotExist
				} else {
					spec = &specs.Spec{}
					if tt.hasSnapshotEnv {
						spec.Process = &specs.Process{
							Env: []string{
								common.GRUS_SNAPSHOT_IMAGE_PATH + "=/test/image/path",
								common.POD_NAME + "=" + "test-0",
							},
						}
						if err := os.MkdirAll("/test/image/path/0", 0755); err != nil {
							t.Fatalf("Failed to create temp directory: %v", err)
						}
						err = os.WriteFile("/test/image/path/0/tempfile", []byte("content"), 0660)
						if err != nil {
							t.Fatalf("Failed to create temp file: %v", err)
						}
					} else {
						spec.Process = &specs.Process{
							Env: []string{""},
						}
					}
				}
				patches.ApplyFunc(readOCIConfig, func(bundlePath string) (*specs.Spec, error) {
					return spec, err
				})
			}

			err := Sstart(tt.args)

			if (err != nil) != tt.expectedErr {
				t.Fatalf("Expected error: %v, got: %v", tt.expectedErr, err != nil)
			}
		})
	}
}
