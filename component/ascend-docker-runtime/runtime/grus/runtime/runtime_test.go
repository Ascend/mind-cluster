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

// Package runtime, ut of runtime
package runtime

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/containerd/go-runc"

	"ascend-common/common-utils/hwlog"
)

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwlog.InitRunLogger(&hwLogConfig, context.Background())
}

func TestGetRuntime(t *testing.T) {
	client := GetRuntime("nil", "/test")
	if client == nil {
		t.Fatalf("test expect not nil, got nil")
	}
	client = GetRuntime(MockName, "/test")
	if client == nil {
		t.Fatalf("mock expect not nil, got nil")
	}
	if err := client.Pause(""); err == nil {
		t.Fatalf("mock expect error, got nil")
	}
	if err := client.Pause("test"); err != nil {
		t.Fatalf("mock expect no error, got: %v", err)
	}

	if err := client.Resume(""); err == nil {
		t.Fatalf("mock expect error, got nil")
	}
	if err := client.Resume("test"); err != nil {
		t.Fatalf("mock expect no error, got: %v", err)
	}

	if _, err := client.State(""); err == nil {
		t.Fatalf("mock expect error, got nil")
	}
	if _, err := client.State("test"); err != nil {
		t.Fatalf("mock expect no error, got: %v", err)
	}

	if err := client.Checkpoint("", "test"); err == nil {
		t.Fatalf("mock expect error, got nil")
	}
	if err := client.Checkpoint("./test", "test"); err != nil {
		t.Fatalf("mock expect no error, got: %v", err)
	}

	if err := client.Restore("", "test", "default", nil); err == nil {
		t.Fatalf("mock expect error, got nil")
	}
	if err := client.Restore("./test", "test", "default", nil); err != nil {
		t.Fatalf("mock expect no error, got: %v", err)
	}
}

func TestRuncRuntime(t *testing.T) {
	tmpDir := t.TempDir()
	client := GetRuntime(RuncName, tmpDir)
	if client == nil {
		t.Fatalf("expect not nil, got nil")
	}

	patches := gomonkey.ApplyFuncReturn(exec.Command, &exec.Cmd{}).
		ApplyMethodReturn(&exec.Cmd{}, "Run", errors.New("test")).
		ApplyMethodReturn(&runc.Runc{}, "Pause", errors.New("test")).
		ApplyMethodReturn(&runc.Runc{}, "Resume", errors.New("test")).
		ApplyMethodReturn(&runc.Runc{}, "State", nil, errors.New("test")).
		ApplyMethodReturn(&runc.Runc{}, "Checkpoint", errors.New("test"))

	defer patches.Reset()

	if err := client.Pause("test"); err == nil {
		t.Fatalf("expect err, got nil")
	}

	if err := client.Resume("test"); err == nil {
		t.Fatalf("expect err, got nil")
	}

	if _, err := client.State("test"); err == nil {
		t.Fatalf("expect err, got nil")
	}

	if err := client.Checkpoint(filepath.Join(tmpDir, "result"), "test"); err == nil {
		t.Fatalf("expect err, got nil")
	}

	os.Args = append(os.Args, "create")
	if err := client.Restore(filepath.Join(tmpDir, "result"), "test", "default", nil); err == nil {
		t.Fatalf("expect err, got nil")
	}
	os.Args = os.Args[0 : len(os.Args)-1]
}

func TestRuncRuntimeWithEmptyClient(t *testing.T) {
	tmpDir := t.TempDir()
	client := NewRuntimeRunc()
	if client == nil {
		t.Fatalf("expect not nil, got nil")
	}

	var err error
	expected := "runc client not init"

	if err = client.Pause("test"); err == nil {
		t.Fatalf("expect err, got nil")
	}
	if err.Error() != expected {
		t.Fatalf("expect %s, got %v", expected, err)
	}

	if err = client.Resume("test"); err == nil {
		t.Fatalf("expect err, got nil")
	}
	if err.Error() != expected {
		t.Fatalf("expect %s, got %v", expected, err)
	}

	if _, err = client.State("test"); err == nil {
		t.Fatalf("expect err, got nil")
	}
	if err.Error() != expected {
		t.Fatalf("expect %s, got %v", expected, err)
	}

	if err = client.Checkpoint(filepath.Join(tmpDir, "result"), "test"); err == nil {
		t.Fatalf("expect err, got nil")
	}
	if err.Error() != expected {
		t.Fatalf("expect %s, got %v", expected, err)
	}
}
