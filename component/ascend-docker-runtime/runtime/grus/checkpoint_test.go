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

// Package grus, ut checkpoint
package grus

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/containerd/go-runc"
	"github.com/opencontainers/runtime-spec/specs-go"

	"ascend-common/common-utils/hwlog"
	"ascend-docker-runtime/runtime/common"
	"ascend-docker-runtime/runtime/grus/rootfs"
	"ascend-docker-runtime/runtime/grus/runtime"
)

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwlog.InitRunLogger(&hwLogConfig, context.Background())
}

func TestScheckpoint(t *testing.T) {
	patch := gomonkey.ApplyPrivateMethod(&runtimeClient{}, "resume", func(containerID string) error {
		return errors.New("test")
	})
	defer patch.Reset()
	args := &common.Args{}
	t.Run("init runtime client failed", func(t *testing.T) {
		patches := gomonkey.ApplyFuncReturn(initRuntimeClient, &runtimeClient{}, errors.New("test"))
		defer patches.Reset()
		err := Scheckpoint(args)
		if err == nil {
			t.Fatalf("expect error, get nil")
		}
	})

	t.Run("pause container failed", func(t *testing.T) {
		patches := gomonkey.ApplyFuncReturn(initRuntimeClient, &runtimeClient{}, nil).
			ApplyPrivateMethod(&runtimeClient{}, "pause", func(containerID string) error {
				return errors.New("test")
			})
		defer patches.Reset()
		err := Scheckpoint(args)
		if err == nil {
			t.Fatalf("expect error, get nil")
		}
	})

	t.Run("use criu to runtime checkpoint container failed", func(t *testing.T) {
		patches := gomonkey.ApplyFuncReturn(initRuntimeClient, &runtimeClient{}, nil).
			ApplyPrivateMethod(&runtimeClient{}, "pause", func(containerID string) error {
				return nil
			}).
			ApplyPrivateMethod(&runtimeClient{}, "checkpoint", func(ckptPath, containerID string) error {
				return errors.New("test")
			})
		defer patches.Reset()
		err := Scheckpoint(args)
		if err == nil {
			t.Fatalf("expect error, get nil")
		}
	})

	t.Run("checkpoint rw layer of container failed", func(t *testing.T) {
		patches := gomonkey.ApplyFuncReturn(initRuntimeClient, &runtimeClient{}, nil).
			ApplyPrivateMethod(&runtimeClient{}, "pause", func(containerID string) error {
				return nil
			}).
			ApplyPrivateMethod(&runtimeClient{}, "checkpoint", func(ckptPath, containerID string) error {
				return nil
			}).
			ApplyFuncReturn(rootfsCheckpoint, errors.New("test"))
		defer patches.Reset()
		err := Scheckpoint(args)
		if err == nil {
			t.Fatalf("expect error, get nil")
		}
	})
}

func TestRuntimeClient(t *testing.T) {
	c := &runtimeClient{}
	err := c.pause("test")
	if err == nil {
		t.Fatalf("pause expect err, got nil")
	}
	err = c.resume("test")
	if err == nil {
		t.Fatalf("resume expect err, got nil")
	}
	err = c.checkpoint("/tmp/test", "test")
	if err == nil {
		t.Fatalf("checkpoint expect err, got nil")
	}

	c.client = runtime.NewRuntimeRunc()
	patches := gomonkey.ApplyFuncReturn(exec.Command, &exec.Cmd{}).
		ApplyMethodReturn(&exec.Cmd{}, "Run", errors.New("test")).
		ApplyMethodReturn(&runc.Runc{}, "Pause", errors.New("test")).
		ApplyMethodReturn(&runc.Runc{}, "Resume", errors.New("test")).
		ApplyMethodReturn(&runc.Runc{}, "State", nil, errors.New("test")).
		ApplyMethodReturn(&runc.Runc{}, "Checkpoint", errors.New("test"))
	defer patches.Reset()

	err = c.pause("test")
	if err == nil {
		t.Fatalf("pause expect err, got nil")
	}
	err = c.resume("test")
	if err == nil {
		t.Fatalf("resume expect err, got nil")
	}
	err = c.checkpoint("/tmp/test", "test")
	if err == nil {
		t.Fatalf("checkpoint expect err, got nil")
	}
	err = c.checkpoint("/tmp/test", "test")
	if err == nil {
		t.Fatalf("checkpoint expect err, got nil")
	}
}

func TestRootfsCheckpoint(t *testing.T) {
	args := &common.Args{}
	patches := gomonkey.ApplyMethodReturn(&rootfs.ContainerdRootfs{}, "Checkpoint", "", errors.New("test"))
	defer patches.Reset()

	err := rootfsCheckpoint(args, "default")
	if err == nil {
		t.Fatalf("rootfsCheckpoint expect err, got nil")
	}
}

func TestSresume(t *testing.T) {
	patch := gomonkey.ApplyPrivateMethod(&runtimeClient{}, "resume", func(containerID string) error {
		return errors.New("test")
	})
	defer patch.Reset()
	args := &common.Args{}
	t.Run("init runtime client failed", func(t *testing.T) {
		patches := gomonkey.ApplyFuncReturn(initRuntimeClient, &runtimeClient{}, errors.New("test"))
		defer patches.Reset()
		err := Sresume(args)
		if err == nil {
			t.Fatalf("expect error, get nil")
		}
	})

	t.Run("get container's status failed", func(t *testing.T) {
		patches := gomonkey.ApplyFuncReturn(initRuntimeClient, &runtimeClient{}, nil).
			ApplyPrivateMethod(&runtimeClient{}, "state", func(containerID string) (*runtime.StateInfo, error) {
				return nil, errors.New("test")
			})
		defer patches.Reset()
		err := Sresume(args)
		if err == nil {
			t.Fatalf("expect error, get nil")
		}
	})

	t.Run("if container not paused, skip resume", func(t *testing.T) {
		patches := gomonkey.ApplyFuncReturn(initRuntimeClient, &runtimeClient{}, nil).
			ApplyPrivateMethod(&runtimeClient{}, "state", func(containerID string) (*runtime.StateInfo, error) {
				return &runtime.StateInfo{Status: "test"}, nil
			})
		defer patches.Reset()
		err := Sresume(args)
		if err != nil {
			t.Fatalf("expect nil, get error")
		}
	})

	t.Run("resume failed", func(t *testing.T) {
		patches := gomonkey.ApplyFuncReturn(initRuntimeClient, &runtimeClient{}, nil).
			ApplyPrivateMethod(&runtimeClient{}, "state", func(containerID string) (*runtime.StateInfo, error) {
				return &runtime.StateInfo{Status: "paused"}, nil
			}).
			ApplyPrivateMethod(&runtimeClient{}, "resume", func(containerID string) error {
				return errors.New("test")
			})
		defer patches.Reset()
		err := Sresume(args)
		if err == nil {
			t.Fatalf("expect error, get nil")
		}
	})

	t.Run("resume success", func(t *testing.T) {
		patches := gomonkey.ApplyFuncReturn(initRuntimeClient, &runtimeClient{}, nil).
			ApplyPrivateMethod(&runtimeClient{}, "state", func(containerID string) (*runtime.StateInfo, error) {
				return &runtime.StateInfo{Status: "paused"}, nil
			}).
			ApplyPrivateMethod(&runtimeClient{}, "resume", func(containerID string) error {
				return nil
			})
		defer patches.Reset()
		err := Sresume(args)
		if err != nil {
			t.Fatalf("expect nil, get error")
		}
	})
}

func TestCheckDumpLogForNpuError(t *testing.T) {
	tests := []struct {
		name        string
		prepareLog  func(ckptPath string) error
		expectedErr bool
	}{
		{
			name: "log file not found",
			prepareLog: func(ckptPath string) error {
				// 不创建日志文件
				return nil
			},
			expectedErr: false,
		},
		{
			name: "log file exists without NPU error",
			prepareLog: func(ckptPath string) error {
				logDir := filepath.Join(ckptPath, "image", "work")
				if err := os.MkdirAll(logDir, 0755); err != nil {
					return err
				}
				logFile := filepath.Join(logDir, "dump.log")
				return os.WriteFile(logFile, []byte("some normal log content"), 0644)
			},
			expectedErr: false,
		},
		{
			name: "log file exists with NPU error",
			prepareLog: func(ckptPath string) error {
				logDir := filepath.Join(ckptPath, "image", "work")
				if err := os.MkdirAll(logDir, 0755); err != nil {
					return err
				}
				logFile := filepath.Join(logDir, "dump.log")
				return os.WriteFile(logFile, []byte("some log content [npu-plugin fini-dump err] some more content"), 0644)
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建临时目录
			ckptPath, err := os.MkdirTemp("", "checkpoint-test")
			if err != nil {
				t.Fatalf("Failed to create temp directory: %v", err)
			}
			defer os.RemoveAll(ckptPath)

			// 准备日志文件
			if err := tt.prepareLog(ckptPath); err != nil {
				t.Fatalf("Failed to prepare log: %v", err)
			}

			// 调用函数
			err = checkDumpLogForNpuError(ckptPath)

			// 验证结果
			if (err != nil) != tt.expectedErr {
				t.Fatalf("Expected error: %v, got: %v", tt.expectedErr, err != nil)
			}
		})
	}
}

func TestUpdateCkptEnvs(t *testing.T) {
	tests := []struct {
		name         string
		conSpec      *specs.Spec
		expectedErr  bool
		expectedEnvs map[string]string
	}{
		{
			name:        "conSpec is nil",
			conSpec:     nil,
			expectedErr: true,
		},
		{
			name: "no dev/shm mount",
			conSpec: &specs.Spec{
				Mounts: []specs.Mount{
					{
						Destination: "/other/path",
						Source:      "/host/path",
					},
				},
			},
			expectedErr: false,
			expectedEnvs: map[string]string{
				"CRIU_CALL_BY_GRUS": "1",
				"CRIU_LOG_LEVEL":    "1",
			},
		},
		{
			name: "with dev/shm mount",
			conSpec: &specs.Spec{
				Mounts: []specs.Mount{
					{
						Destination: "/dev/shm",
						Source:      "/host/shm",
					},
				},
			},
			expectedErr: false,
			expectedEnvs: map[string]string{
				"CRIU_CALL_BY_GRUS":       "1",
				"CRIU_LOG_LEVEL":          "1",
				"SNAPSHOT_LINK_REMAP_SRC": "/host/shm",
				"SNAPSHOT_LINK_REMAP_DST": "test-ckpt/rootfs-external-diff.tar",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 保存原始环境变量
			originalEnvs := make(map[string]string)
			for key := range tt.expectedEnvs {
				originalEnvs[key] = os.Getenv(key)
			}
			defer func() {
				// 恢复原始环境变量
				for key, value := range originalEnvs {
					if value == "" {
						os.Unsetenv(key)
					} else {
						os.Setenv(key, value)
					}
				}
			}()

			// 清除测试相关的环境变量
			for key := range tt.expectedEnvs {
				os.Unsetenv(key)
			}

			// 创建 runtimeClient
			client := &runtimeClient{
				conSpec: tt.conSpec,
			}

			// 调用方法
			err := client.updateCkptEnvs("test-ckpt")

			// 验证错误
			if (err != nil) != tt.expectedErr {
				t.Fatalf("Expected error: %v, got: %v", tt.expectedErr, err != nil)
			}

			// 验证环境变量
			if !tt.expectedErr {
				for key, expectedValue := range tt.expectedEnvs {
					actualValue := os.Getenv(key)
					if actualValue != expectedValue {
						t.Fatalf("Expected %s=%s, got %s=%s", key, expectedValue, key, actualValue)
					}
				}
			}
		})
	}
}
