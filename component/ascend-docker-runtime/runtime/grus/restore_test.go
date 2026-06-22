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

// Package grus, unit tests for restore
package grus

import (
	"ascend-docker-runtime/runtime/grus/runtime"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/opencontainers/runtime-spec/specs-go"

	"ascend-docker-runtime/runtime/common"
	"ascend-docker-runtime/runtime/grus/rootfs"
)

func TestGetEnvFromSpec(t *testing.T) {
	tests := []struct {
		name        string
		spec        *specs.Spec
		key         string
		expectedVal string
	}{
		{
			name: "env exists",
			spec: &specs.Spec{
				Process: &specs.Process{
					Env: []string{
						"KEY1=value1",
						"KEY2=value2",
					},
				},
			},
			key:         "KEY1",
			expectedVal: "value1",
		},
		{
			name: "env not exists",
			spec: &specs.Spec{
				Process: &specs.Process{
					Env: []string{
						"KEY1=value1",
						"KEY2=value2",
					},
				},
			},
			key:         "KEY3",
			expectedVal: "",
		},
		{
			name: "env with equals in value",
			spec: &specs.Spec{
				Process: &specs.Process{
					Env: []string{
						"KEY1=value=with=equals",
					},
				},
			},
			key:         "KEY1",
			expectedVal: "value=with=equals",
		},
		{
			name: "empty env list",
			spec: &specs.Spec{
				Process: &specs.Process{
					Env: []string{},
				},
			},
			key:         "KEY1",
			expectedVal: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getEnvFromSpec(tt.spec, tt.key)
			if result != tt.expectedVal {
				t.Fatalf("Expected %s, got %s", tt.expectedVal, result)
			}
		})
	}
}

func TestCreateFlagFile(t *testing.T) {
	tests := []struct {
		name        string
		spec        *specs.Spec
		rootfs      string
		expectedErr bool
		shouldExist bool
	}{
		{
			name: "flag disabled",
			spec: &specs.Spec{
				Process: &specs.Process{
					Env: []string{
						common.GRUS_SNAPSHOT_RESTORED_FLAG + "=false",
					},
				},
			},
			rootfs:      "/test/rootfs",
			expectedErr: false,
			shouldExist: false,
		},
		{
			name: "flag enabled",
			spec: &specs.Spec{
				Process: &specs.Process{
					Env: []string{
						common.GRUS_SNAPSHOT_RESTORED_FLAG + "=true",
					},
				},
			},
			rootfs:      "/test/rootfs",
			expectedErr: false,
			shouldExist: true,
		},
		{
			name: "flag enabled with uppercase",
			spec: &specs.Spec{
				Process: &specs.Process{
					Env: []string{
						common.GRUS_SNAPSHOT_RESTORED_FLAG + "=TRUE",
					},
				},
			},
			rootfs:      "/test/rootfs",
			expectedErr: false,
			shouldExist: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir, err := os.MkdirTemp("", "create-flag-test")
			if err != nil {
				t.Fatalf("Failed to create temp directory: %v", err)
			}
			defer os.RemoveAll(tempDir)

			rootfsPath := filepath.Join(tempDir, tt.rootfs)
			if err := os.MkdirAll(rootfsPath, 0755); err != nil {
				t.Fatalf("Failed to create rootfs directory: %v", err)
			}

			err = createFlagFile(tt.spec, rootfsPath)

			if (err != nil) != tt.expectedErr {
				t.Fatalf("Expected error: %v, got: %v", tt.expectedErr, err != nil)
			}

			flagFile := filepath.Join(rootfsPath, common.GRUS_RESTORE_FLAG_FILE)
			_, exists := os.Stat(flagFile)
			if tt.shouldExist && os.IsNotExist(exists) {
				t.Fatalf("Expected flag file to exist, but it doesn't")
			}
			if !tt.shouldExist && !os.IsNotExist(exists) {
				t.Fatalf("Expected flag file to not exist, but it does")
			}
		})
	}
}

func TestPrepareLogfile(t *testing.T) {
	tests := []struct {
		name        string
		prepareFile func(bundlePath string) error
		expectedErr bool
	}{
		{
			name: "log.json does not exist",
			prepareFile: func(bundlePath string) error {
				return os.MkdirAll(bundlePath, 0755)
			},
			expectedErr: false,
		},
		{
			name: "log.json already exists",
			prepareFile: func(bundlePath string) error {
				if err := os.MkdirAll(bundlePath, 0755); err != nil {
					return err
				}
				logPath := filepath.Join(bundlePath, "log.json")
				return os.WriteFile(logPath, []byte("existing log"), 0644)
			},
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir, err := os.MkdirTemp("", "prepare-log-test")
			if err != nil {
				t.Fatalf("Failed to create temp directory: %v", err)
			}
			defer os.RemoveAll(tempDir)

			bundlePath := filepath.Join(tempDir, "bundle")
			if err := tt.prepareFile(bundlePath); err != nil {
				t.Fatalf("Failed to prepare bundle: %v", err)
			}

			err = prepareLogfile(bundlePath)

			if (err != nil) != tt.expectedErr {
				t.Fatalf("Expected error: %v, got: %v", tt.expectedErr, err != nil)
			}

			logPath := filepath.Join(bundlePath, "log.json")
			if _, err := os.Stat(logPath); os.IsNotExist(err) {
				t.Fatalf("Expected log.json to exist, but it doesn't")
			}
		})
	}
}

func TestRootfsRestore(t *testing.T) {
	tests := []struct {
		name        string
		mockRestore bool
		restoreErr  error
		expectedErr bool
	}{
		{
			name:        "restore success",
			mockRestore: true,
			restoreErr:  nil,
			expectedErr: false,
		},
		{
			name:        "restore failed",
			mockRestore: true,
			restoreErr:  fmt.Errorf("restore failed"),
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patches := gomonkey.NewPatches()
			defer patches.Reset()

			if tt.mockRestore {
				patches.ApplyFuncReturn(rootfs.ApplyTarDiff, tt.restoreErr)
			}

			err := rootfsRestore("/test/ckpt", "/test/rootfs")

			if (err != nil) != tt.expectedErr {
				t.Fatalf("Expected error: %v, got: %v", tt.expectedErr, err != nil)
			}
		})
	}
}

func TestGetContainerNetns(t *testing.T) {
	tests := []struct {
		name        string
		config      *specs.Spec
		expectedNs  string
		expectedErr bool
	}{
		{
			name: "network namespace found",
			config: &specs.Spec{
				Linux: &specs.Linux{
					Namespaces: []specs.LinuxNamespace{
						{
							Type: specs.NetworkNamespace,
							Path: "/var/run/netns/test",
						},
					},
				},
			},
			expectedNs:  "/var/run/netns/test",
			expectedErr: false,
		},
		{
			name: "network namespace not found",
			config: &specs.Spec{
				Linux: &specs.Linux{
					Namespaces: []specs.LinuxNamespace{
						{
							Type: specs.PIDNamespace,
							Path: "/var/run/pidns/test",
						},
					},
				},
			},
			expectedNs:  "",
			expectedErr: true,
		},
		{
			name: "empty namespaces",
			config: &specs.Spec{
				Linux: &specs.Linux{
					Namespaces: []specs.LinuxNamespace{},
				},
			},
			expectedNs:  "",
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns, err := getContainerNetns("test-container", tt.config)

			if (err != nil) != tt.expectedErr {
				t.Fatalf("Expected error: %v, got: %v", tt.expectedErr, err != nil)
			}

			if ns != tt.expectedNs {
				t.Fatalf("Expected namespace: %s, got: %s", tt.expectedNs, ns)
			}
		})
	}
}

func TestGetLocalIP(t *testing.T) {
	tests := []struct {
		name         string
		mockAddrs    bool
		addrs        []net.Addr
		expectedIPv4 string
		expectedIPv6 string
		expectedErr  bool
	}{
		{
			name:      "valid IPv4 and IPv6",
			mockAddrs: true,
			addrs: []net.Addr{
				&net.IPNet{
					IP:   net.ParseIP("192.168.1.1"),
					Mask: net.CIDRMask(24, 32),
				},
				&net.IPNet{
					IP:   net.ParseIP("2001:db8::1"),
					Mask: net.CIDRMask(64, 128),
				},
			},
			expectedIPv4: "192.168.1.1",
			expectedIPv6: "2001:db8::1",
			expectedErr:  false,
		},
		{
			name:      "only IPv4",
			mockAddrs: true,
			addrs: []net.Addr{
				&net.IPNet{
					IP:   net.ParseIP("192.168.1.1"),
					Mask: net.CIDRMask(24, 32),
				},
			},
			expectedIPv4: "192.168.1.1",
			expectedIPv6: "",
			expectedErr:  false,
		},
		{
			name:      "only loopback",
			mockAddrs: true,
			addrs: []net.Addr{
				&net.IPNet{
					IP:   net.ParseIP("127.0.0.1"),
					Mask: net.CIDRMask(8, 32),
				},
			},
			expectedIPv4: "",
			expectedIPv6: "",
			expectedErr:  true,
		},
		{
			name:         "empty addresses",
			mockAddrs:    true,
			addrs:        []net.Addr{},
			expectedIPv4: "",
			expectedIPv6: "",
			expectedErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patches := gomonkey.NewPatches()
			defer patches.Reset()

			if tt.mockAddrs {
				patches.ApplyFunc(netInterfaceAddrs, func() ([]net.Addr, error) {
					return tt.addrs, nil
				})
			}

			result, err := getLocalIP()

			if (err != nil) != tt.expectedErr {
				t.Fatalf("Expected error: %v, got: %v", tt.expectedErr, err != nil)
			}

			if len(result) != 2 {
				t.Fatalf("Expected result length 2, got: %d", len(result))
			}

			if result[0] != tt.expectedIPv4 {
				t.Fatalf("Expected IPv4: %s, got: %s", tt.expectedIPv4, result[0])
			}

			if result[1] != tt.expectedIPv6 {
				t.Fatalf("Expected IPv6: %s, got: %s", tt.expectedIPv6, result[1])
			}
		})
	}
}

func TestGetContainerNewIP(t *testing.T) {
	tests := []struct {
		name             string
		config           *specs.Spec
		mockGetNetnsIP   bool
		getNetnsIPErr    error
		getNetnsIPResult []string
		expectedResult   []string
		expectedErr      bool
	}{
		{
			name: "host network mode",
			config: &specs.Spec{
				Linux: &specs.Linux{
					Namespaces: []specs.LinuxNamespace{
						{
							Type: specs.PIDNamespace,
						},
					},
				},
			},
			expectedResult: nil,
			expectedErr:    false,
		},
		{
			name: "network namespace found",
			config: &specs.Spec{
				Linux: &specs.Linux{
					Namespaces: []specs.LinuxNamespace{
						{
							Type: specs.NetworkNamespace,
							Path: "/var/run/netns/test",
						},
					},
				},
			},
			mockGetNetnsIP:   true,
			getNetnsIPResult: []string{"192.168.1.1", "2001:db8::1"},
			expectedResult:   []string{"192.168.1.1", "2001:db8::1"},
			expectedErr:      false,
		},
		{
			name: "getNetnsIP failed",
			config: &specs.Spec{
				Linux: &specs.Linux{
					Namespaces: []specs.LinuxNamespace{
						{
							Type: specs.NetworkNamespace,
							Path: "/var/run/netns/test",
						},
					},
				},
			},
			mockGetNetnsIP: true,
			getNetnsIPErr:  fmt.Errorf("get netns ip failed"),
			expectedResult: nil,
			expectedErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patches := gomonkey.NewPatches()
			defer patches.Reset()

			if tt.mockGetNetnsIP {
				patches.ApplyFunc(getNetnsIP, func(nsPath string) ([]string, error) {
					return tt.getNetnsIPResult, tt.getNetnsIPErr
				})
			}

			result, err := getContainerNewIP("test-container", tt.config)

			if (err != nil) != tt.expectedErr {
				t.Fatalf("Expected error: %v, got: %v", tt.expectedErr, err != nil)
			}

			if len(result) != len(tt.expectedResult) {
				t.Fatalf("Expected result length %d, got: %d", len(tt.expectedResult), len(result))
			}

			for i, expected := range tt.expectedResult {
				if result[i] != expected {
					t.Fatalf("Expected result[%d]: %s, got: %s", i, expected, result[i])
				}
			}
		})
	}
}

func TestReadRuntimeLog(t *testing.T) {
	tests := []struct {
		name        string
		prepareFile func(logPath string) error
		expectedLog string
	}{
		{
			name: "file not found",
			prepareFile: func(logPath string) error {
				return nil
			},
			expectedLog: "",
		},
		{
			name: "file exists with long content",
			prepareFile: func(logPath string) error {
				longContent := strings.Repeat("a", common.READ_MAX_LEN+100)
				return os.WriteFile(logPath, []byte(longContent), 0644)
			},
			expectedLog: strings.Repeat("a", common.READ_MAX_LEN),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir, err := os.MkdirTemp("", "read-log-test")
			if err != nil {
				t.Fatalf("Failed to create temp directory: %v", err)
			}
			defer os.RemoveAll(tempDir)

			logPath := filepath.Join(tempDir, "log.json")
			if err := tt.prepareFile(logPath); err != nil {
				t.Fatalf("Failed to prepare log file: %v", err)
			}

			result := readRuntimeLog(logPath)

			if result != tt.expectedLog {
				t.Fatalf("Expected log: %s, got: %s", tt.expectedLog, result)
			}
		})
	}
}

func TestValidateSnapshotImagePath(t *testing.T) {
	tests := []struct {
		name         string
		imagePath    string
		containerID  string
		prepareDir   func(path string) error
		expectedPath string
		expectedErr  bool
	}{
		{
			name:         "empty image path",
			imagePath:    "",
			containerID:  "test-container",
			expectedPath: "",
			expectedErr:  false,
		},
		{
			name:         "relative path",
			imagePath:    "relative/path",
			containerID:  "test-container",
			expectedPath: "",
			expectedErr:  false,
		},
		{
			name:         "absolute path not exist",
			imagePath:    "/nonexistent/path",
			containerID:  "test-container",
			expectedPath: "",
			expectedErr:  true,
		},
		{
			name:        "absolute path is file",
			imagePath:   "/file",
			containerID: "test-container",
			prepareDir: func(path string) error {
				return os.WriteFile(path, []byte("content"), 0644)
			},
			expectedPath: "",
			expectedErr:  true,
		},
		{
			name:        "absolute path is directory",
			imagePath:   "/test/dir",
			containerID: "test-container",
			prepareDir: func(path string) error {
				return os.MkdirAll(path, 0755)
			},
			expectedPath: "/tmp/restore_test/test/dir",
			expectedErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := os.MkdirAll("/tmp/restore_test", 0755)
			if err != nil {
				t.Fatalf("Failed to create temp directory: %v", err)
			}
			defer os.RemoveAll("/tmp/restore_test")

			imagePath := tt.imagePath
			if filepath.IsAbs(imagePath) {
				imagePath = filepath.Join("/tmp/restore_test", imagePath)
			}

			if tt.prepareDir != nil {
				if err := tt.prepareDir(imagePath); err != nil {
					t.Fatalf("Failed to prepare directory: %v", err)
				}
			}

			result, err := validateSnapshotImagePath(imagePath, tt.containerID)

			if (err != nil) != tt.expectedErr {
				t.Fatalf("Expected error: %v, got: %v", tt.expectedErr, err != nil)
			}

			if result != tt.expectedPath {
				t.Fatalf("Expected path: %s, got: %s", tt.expectedPath, result)
			}
		})
	}
}

func TestScreate(t *testing.T) {
	tests := []struct {
		name               string
		args               *common.Args
		mockReadOCIConfig  bool
		readOCIConfigErr   error
		mockRuntimeRestore bool
		runtimeRestoreErr  error
		expectedErr        bool
	}{
		{
			name: "bundle is empty",
			args: &common.Args{
				Bundle: "",
			},
			expectedErr: true,
		},
		{
			name: "readOCIConfig failed",
			args: &common.Args{
				ContainerID: "test-container",
				Root:        "/test/root",
				Bundle:      "/test/bundle",
			},
			mockReadOCIConfig: true,
			readOCIConfigErr:  fmt.Errorf("read config failed"),
			expectedErr:       true,
		},
		{
			name: "restore success",
			args: &common.Args{
				ContainerID: "test-container",
				Root:        "/test/root",
				Bundle:      "/test/bundle",
			},
			mockReadOCIConfig:  true,
			mockRuntimeRestore: true,
			expectedErr:        false,
		},
		{
			name: "restore failed",
			args: &common.Args{
				ContainerID: "test-container",
				Root:        "/test/root",
				Bundle:      "/test/bundle",
			},
			mockReadOCIConfig:  true,
			mockRuntimeRestore: true,
			runtimeRestoreErr:  fmt.Errorf("restore failed"),
			expectedErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir, err := os.MkdirTemp("", "screate-test")
			if err != nil {
				t.Fatalf("Failed to create temp directory: %v", err)
			}
			defer os.RemoveAll(tempDir)

			if tt.args.Bundle != "" {
				tt.args.Bundle = filepath.Join(tempDir, tt.args.Bundle)
				if err := os.MkdirAll(tt.args.Bundle, 0755); err != nil {
					t.Fatalf("Failed to create bundle directory: %v", err)
				}
			}

			patches := gomonkey.NewPatches()
			defer patches.Reset()

			if tt.mockReadOCIConfig {
				var spec *specs.Spec
				var err error
				if tt.readOCIConfigErr != nil {
					err = tt.readOCIConfigErr
				} else {
					spec = &specs.Spec{
						Root: &specs.Root{
							Path: "rootfs",
						},
						Process: &specs.Process{
							Env: []string{
								common.GRUS_SNAPSHOT_IMAGE_PATH + "=" + filepath.Join(tempDir, "snapshot"),
								common.POD_NAME + "=" + "test-0",
							},
						},
						Mounts: []specs.Mount{},
					}
					if err := os.MkdirAll(filepath.Join(tempDir, "snapshot/0"), 0755); err != nil {
						t.Fatalf("Failed to create snapshot directory: %v", err)
					}
					err = os.WriteFile(filepath.Join(tempDir, "snapshot/0", "tempfile"), []byte("content"), 0660)
					if err != nil {
						t.Fatalf("Failed to create temp file: %v", err)
					}
				}
				patches.ApplyFunc(readOCIConfig, func(bundlePath string) (*specs.Spec, error) {
					return spec, err
				})
			}

			if tt.mockRuntimeRestore {
				patches.ApplyMethodFunc(&runtime.RuncRuntime{}, "Restore", func(ckptPath, id, ns string, externalEnvs []string) error {
					return tt.runtimeRestoreErr
				})
			}

			patches.ApplyFunc(rootfsRestore, func(ckptPath, rootfsPath string) error {
				return nil
			})

			patches.ApplyFunc(createFlagFile, func(spec *specs.Spec, rootfs string) error {
				return nil
			})

			patches.ApplyFunc(prepareLogfile, func(bundlePath string) error {
				return nil
			})

			patches.ApplyFunc(getContainerNewIP, func(conID string, config *specs.Spec) ([]string, error) {
				return nil, nil
			})

			patches.ApplyFunc(common.ExecRunc, func() error {
				return nil
			})

			patches.ApplyFunc(common.WriteSpecFile, func(path string, spec *specs.Spec) error {
				return nil
			})

			err = Screate(tt.args)

			if (err != nil) != tt.expectedErr {
				t.Fatalf("Expected error: %v, got: %v", tt.expectedErr, err != nil)
			}
		})
	}
}
