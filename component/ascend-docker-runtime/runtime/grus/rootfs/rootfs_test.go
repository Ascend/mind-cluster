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

// Package rootfs, test
package rootfs

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/containerd/containerd"

	"ascend-common/common-utils/hwlog"
	"ascend-docker-runtime/runtime/common"
)

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwlog.InitRunLogger(&hwLogConfig, context.Background())
}

func TestGetRootfsSnapshot(t *testing.T) {
	rootfs := GetRootfsSnapshot("test", "/var/run/test.sock")
	if rootfs == nil {
		t.Fatalf("test expect not nil, got nil")
	}
	rootfs = GetRootfsSnapshot(MOCK_SNAPSHOT_KEY, "mock.sock")
	if rootfs == nil {
		t.Fatalf("mock expect not nil, got nil")
	}
	_, err := rootfs.Checkpoint("/tmp/test", "", "default")
	if err == nil {
		t.Fatalf("mock Checkpoint expect error, got nil")
	}

	_, err = rootfs.Checkpoint("/tmp/test", "test", "default")
	if err != nil {
		t.Fatalf("mock Checkpoint expect no err, got: %v", err)
	}

	err = rootfs.Restore("", "", "default")
	if err == nil {
		t.Fatalf("mock Restore expect error, got nil")
	}

	err = rootfs.Restore("/tmp/test", "test", "default")
	if err != nil {
		t.Fatalf("mock Restore expect no err, got: %v", err)
	}
}
func TestContainerSnapshot(t *testing.T) {
	dir := t.TempDir()
	rootfs := GetRootfsSnapshot(common.CONTAINERD_SNAPSHOT_KEY, "containerd.sock")
	if rootfs == nil {
		t.Fatalf("expect not nil, got nil")
	}
	ckptDir := filepath.Join(dir, "test")
	if _, err := rootfs.Checkpoint(ckptDir, "test", "default"); err == nil {
		t.Fatalf("expect not nil, got nil")
	}

	if err := os.Mkdir(ckptDir, common.COMMON_DIR_MODE); err != nil {
		t.Fatalf("expect nil, got err: %v", err)
	}
	if _, err := rootfs.Checkpoint(ckptDir, "test", "default"); err == nil {
		t.Fatalf("expect not nil, got nil")
	}
	if _, err := os.Stat(filepath.Join(ckptDir, common.ROOTFS_DIFF)); err != nil {
		t.Fatalf("expect nil, got err: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	patches := gomonkey.ApplyMethodReturn(&ContainerdRootfs{}, "GetSnapshotClient", ctx, cancel, &containerd.Client{}, errors.New("test"))
	defer patches.Reset()

	if _, err := rootfs.Checkpoint(ckptDir, "test", "default"); err == nil {
		t.Fatalf("expect not nil, got nil")
	}

	if err := rootfs.Restore(ckptDir, filepath.Join(dir, "rootfs"), ""); err != nil {
		t.Fatalf("expect nil, got err: %v", err)
	}
}
