/*
 * Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at

 * http://www.apache.org/licenses/LICENSE-2.0

 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package plugin

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	resourceapi "k8s.io/api/resource/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	kubeletplugin "k8s.io/dynamic-resource-allocation/kubeletplugin"

	"ascend-common/common-utils/healthz"
	draFlags "ascend-dynamic-resource-allocation/internal/flags"
)

// RegisterService is not covered here: kubeletplugin.Start creates real
// registration sockets and a node registrar that talks to the API server,
// neither of which is available in unit tests.

// fakeCdiSpec is an in-memory CdiSpecInterface fake recording calls.
type fakeCdiSpec struct {
	writeErr  error
	deleteErr error
	cdiIDs    []string
	writes    []string
	deletes   []string
}

func (f *fakeCdiSpec) WriteClaimSpec(claimUID string, _ []string) ([]string, error) {
	f.writes = append(f.writes, claimUID)
	if f.writeErr != nil {
		return nil, f.writeErr
	}
	return f.cdiIDs, nil
}

func (f *fakeCdiSpec) DeleteClaimSpec(claimUID string) error {
	f.deletes = append(f.deletes, claimUID)
	return f.deleteErr
}

// newTestDeviceState creates a DeviceState with its checkpoint directory
// under a fresh temp dir and the given spec fake.
func newTestDeviceState(t *testing.T, specs CdiSpecInterface) *DeviceState {
	t.Helper()
	draOption := &draFlags.DRAOption{
		NodeName:                    "ut-node",
		KubeletPluginsDirectoryPath: t.TempDir(),
	}
	state, err := NewDeviceState(draOption, specs)
	if err != nil {
		t.Fatalf("NewDeviceState() error = %v, want nil", err)
	}
	return state
}

// allocatedClaim builds a claim with one allocated device result.
func allocatedClaim(uid, device string) *resourceapi.ResourceClaim {
	return &resourceapi.ResourceClaim{
		ObjectMeta: metav1.ObjectMeta{UID: types.UID(uid)},
		Status: resourceapi.ResourceClaimStatus{
			Allocation: &resourceapi.AllocationResult{
				Devices: resourceapi.DeviceAllocationResult{
					Results: []resourceapi.DeviceRequestAllocationResult{
						{Request: "req-0", Pool: "pool-a", Device: device},
					},
				},
			},
		},
	}
}

// writePluginTestKubeconfig writes a minimal valid kubeconfig file.
func writePluginTestKubeconfig(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "kubeconfig")
	content := `apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://127.0.0.1:6443
  name: ut-cluster
contexts:
- context:
    cluster: ut-cluster
    user: ut-user
  name: ut-context
current-context: ut-context
users:
- name: ut-user
  user:
    token: ut-token
`
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
	return path
}

// TestNewAscendDraPlugin verifies plugin construction succeeds with a valid
// kubeconfig and fails when the kube client cannot be created.
func TestNewAscendDraPlugin(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		draConfig := &draFlags.DRAConfig{
			DraOption: &draFlags.DRAOption{
				NodeName:                    "ut-node",
				KubeletPluginsDirectoryPath: t.TempDir(),
			},
			KubeClientConfig: &draFlags.KubeClientConfig{
				KubeConfig:   writePluginTestKubeconfig(t),
				KubeAPIQPS:   5,
				KubeAPIBurst: 10,
			},
			DraHealthzConfig: &healthz.Config{EnableHealthz: false},
		}

		adp, err := NewAscendDraPlugin(draConfig, func(error) {}, &fakeCdiSpec{})

		if err != nil {
			t.Fatalf("NewAscendDraPlugin() error = %v, want nil", err)
		}
		if adp.state == nil {
			t.Error("NewAscendDraPlugin().state = nil, want non-nil device state")
		}
		if adp.DraHealthManager.DraHealthzConfig != draConfig.DraHealthzConfig {
			t.Error("DraHealthzConfig was not stored on the plugin")
		}
	})

	t.Run("client creation fails", func(t *testing.T) {
		t.Setenv("KUBERNETES_SERVICE_HOST", "")
		t.Setenv("KUBERNETES_SERVICE_PORT", "")
		draConfig := &draFlags.DRAConfig{
			DraOption:        &draFlags.DRAOption{},
			KubeClientConfig: &draFlags.KubeClientConfig{},
			DraHealthzConfig: &healthz.Config{},
		}

		adp, err := NewAscendDraPlugin(draConfig, func(error) {}, &fakeCdiSpec{})

		if err == nil {
			t.Fatalf("NewAscendDraPlugin() = %+v, want error", adp)
		}
		if !strings.Contains(err.Error(), "create client") {
			t.Errorf("NewAscendDraPlugin() error = %v, want containing %q", err, "create client")
		}
	})
}

// TestPrepareResourceClaims_Errors verifies preparing no claims and an
// unallocated claim.
func TestPrepareResourceClaims_Errors(t *testing.T) {
	t.Run("empty claims", func(t *testing.T) {
		adp := &AscendDraPlugin{state: newTestDeviceState(t, &fakeCdiSpec{})}

		result, err := adp.PrepareResourceClaims(context.Background(), nil)

		if err != nil {
			t.Fatalf("PrepareResourceClaims(nil) error = %v, want nil", err)
		}
		if len(result) != 0 {
			t.Errorf("len(result) = %d, want 0", len(result))
		}
	})

	t.Run("claim not allocated", func(t *testing.T) {
		adp := &AscendDraPlugin{state: newTestDeviceState(t, &fakeCdiSpec{})}
		claim := &resourceapi.ResourceClaim{
			ObjectMeta: metav1.ObjectMeta{UID: types.UID("uid-unallocated")},
		}

		result, err := adp.PrepareResourceClaims(context.Background(),
			[]*resourceapi.ResourceClaim{claim})

		if err != nil {
			t.Fatalf("PrepareResourceClaims() error = %v, want nil", err)
		}
		got := result[claim.UID]
		if got.Err == nil {
			t.Fatal("PrepareResult.Err = nil, want error")
		}
		if !strings.Contains(got.Err.Error(), "claim not yet allocated") {
			t.Errorf("PrepareResult.Err = %v, want containing %q", got.Err, "claim not yet allocated")
		}
	})
}

// TestPrepareResourceClaims_Success verifies a prepared claim maps allocated
// devices into the kubelet result, and that a CDI write failure surfaces as
// a per-claim error.
func TestPrepareResourceClaims_Success(t *testing.T) {
	t.Run("allocated claim", func(t *testing.T) {
		specs := &fakeCdiSpec{cdiIDs: []string{"ascend.com/npu=0"}}
		adp := &AscendDraPlugin{state: newTestDeviceState(t, specs)}
		claim := allocatedClaim("uid-ok", "Ascend910-0")

		result, err := adp.PrepareResourceClaims(context.Background(),
			[]*resourceapi.ResourceClaim{claim})

		if err != nil {
			t.Fatalf("PrepareResourceClaims() error = %v, want nil", err)
		}
		got := result[claim.UID]
		if got.Err != nil {
			t.Fatalf("PrepareResult.Err = %v, want nil", got.Err)
		}
		if len(got.Devices) != 1 {
			t.Fatalf("len(Devices) = %d, want 1", len(got.Devices))
		}
		dev := got.Devices[0]
		if dev.PoolName != "pool-a" || dev.DeviceName != "Ascend910-0" {
			t.Errorf("device = %s/%s, want pool-a/Ascend910-0", dev.PoolName, dev.DeviceName)
		}
		if len(dev.Requests) != 1 || dev.Requests[0] != "req-0" {
			t.Errorf("Requests = %v, want [req-0]", dev.Requests)
		}
		if len(dev.CDIDeviceIDs) != 1 || dev.CDIDeviceIDs[0] != "ascend.com/npu=0" {
			t.Errorf("CDIDeviceIDs = %v, want [ascend.com/npu=0]", dev.CDIDeviceIDs)
		}
	})

	t.Run("write spec fails", func(t *testing.T) {
		specs := &fakeCdiSpec{writeErr: errors.New("boom")}
		adp := &AscendDraPlugin{state: newTestDeviceState(t, specs)}
		claim := allocatedClaim("uid-wf", "Ascend910-1")

		result, err := adp.PrepareResourceClaims(context.Background(),
			[]*resourceapi.ResourceClaim{claim})

		if err != nil {
			t.Fatalf("PrepareResourceClaims() error = %v, want nil", err)
		}
		got := result[claim.UID]
		if got.Err == nil {
			t.Fatal("PrepareResult.Err = nil, want error")
		}
		if !strings.Contains(got.Err.Error(), "boom") {
			t.Errorf("PrepareResult.Err = %v, want containing %q", got.Err, "boom")
		}
	})
}

// TestUnprepareResourceClaims verifies unpreparing no claims and an unknown
// claim is a no-op.
func TestUnprepareResourceClaims(t *testing.T) {
	t.Run("empty claims", func(t *testing.T) {
		adp := &AscendDraPlugin{state: newTestDeviceState(t, &fakeCdiSpec{})}

		result, err := adp.UnprepareResourceClaims(context.Background(), nil)

		if err != nil {
			t.Fatalf("UnprepareResourceClaims(nil) error = %v, want nil", err)
		}
		if len(result) != 0 {
			t.Errorf("len(result) = %d, want 0", len(result))
		}
	})

	t.Run("unknown claim is no-op", func(t *testing.T) {
		specs := &fakeCdiSpec{}
		adp := &AscendDraPlugin{state: newTestDeviceState(t, specs)}
		obj := kubeletplugin.NamespacedObject{UID: types.UID("uid-unknown")}

		result, err := adp.UnprepareResourceClaims(context.Background(),
			[]kubeletplugin.NamespacedObject{obj})

		if err != nil {
			t.Fatalf("UnprepareResourceClaims() error = %v, want nil", err)
		}
		if err := result[obj.UID]; err != nil {
			t.Errorf("result[uid] = %v, want nil (idempotent)", err)
		}
	})
}

// TestUnprepareResourceClaims_AfterPrepare verifies a prepared claim is
// released and that a CDI delete failure surfaces as an error.
func TestUnprepareResourceClaims_AfterPrepare(t *testing.T) {
	t.Run("prepared claim released", func(t *testing.T) {
		specs := &fakeCdiSpec{cdiIDs: []string{"ascend.com/npu=0"}}
		adp := &AscendDraPlugin{state: newTestDeviceState(t, specs)}
		claim := allocatedClaim("uid-rel", "Ascend910-0")
		ctx := context.Background()
		if _, err := adp.PrepareResourceClaims(ctx, []*resourceapi.ResourceClaim{claim}); err != nil {
			t.Fatalf("PrepareResourceClaims() error = %v, want nil", err)
		}
		obj := kubeletplugin.NamespacedObject{UID: claim.UID}

		result, err := adp.UnprepareResourceClaims(ctx, []kubeletplugin.NamespacedObject{obj})

		if err != nil {
			t.Fatalf("UnprepareResourceClaims() error = %v, want nil", err)
		}
		if err := result[obj.UID]; err != nil {
			t.Errorf("result[uid] = %v, want nil", err)
		}
		if len(specs.deletes) != 1 || specs.deletes[0] != string(claim.UID) {
			t.Errorf("DeleteClaimSpec calls = %v, want [%s]", specs.deletes, claim.UID)
		}
	})

	t.Run("delete spec fails", func(t *testing.T) {
		specs := &fakeCdiSpec{cdiIDs: []string{"ascend.com/npu=0"}}
		adp := &AscendDraPlugin{state: newTestDeviceState(t, specs)}
		claim := allocatedClaim("uid-df", "Ascend910-1")
		ctx := context.Background()
		if _, err := adp.PrepareResourceClaims(ctx, []*resourceapi.ResourceClaim{claim}); err != nil {
			t.Fatalf("PrepareResourceClaims() error = %v, want nil", err)
		}
		specs.deleteErr = errors.New("boom-del")
		obj := kubeletplugin.NamespacedObject{UID: claim.UID}

		result, err := adp.UnprepareResourceClaims(ctx, []kubeletplugin.NamespacedObject{obj})

		if err != nil {
			t.Fatalf("UnprepareResourceClaims() error = %v, want nil", err)
		}
		if err := result[obj.UID]; err == nil {
			t.Fatal("result[uid] = nil, want error")
		} else if !strings.Contains(err.Error(), "boom-del") {
			t.Errorf("result[uid] = %v, want containing %q", err, "boom-del")
		}
	})
}

// TestHandleError verifies fatal errors cancel the plugin context while
// recoverable errors and a nil cancel function do not.
func TestHandleError(t *testing.T) {
	t.Run("fatal error cancels context", func(t *testing.T) {
		ctx, cancel := context.WithCancelCause(context.Background())
		adp := &AscendDraPlugin{cancelCtx: cancel}

		adp.HandleError(context.Background(), errors.New("boom"), "ut-msg")

		cause := context.Cause(ctx)
		if cause == nil {
			t.Fatal("context not cancelled after fatal error")
		}
		if !strings.Contains(cause.Error(), "fatal background error") {
			t.Errorf("context cause = %v, want containing %q", cause, "fatal background error")
		}
	})

	t.Run("recoverable error does not cancel", func(t *testing.T) {
		ctx, cancel := context.WithCancelCause(context.Background())
		adp := &AscendDraPlugin{cancelCtx: cancel}

		adp.HandleError(context.Background(),
			wrapRecoverable("boom"), "ut-msg")

		if context.Cause(ctx) != nil {
			t.Error("context cancelled after recoverable error, want untouched")
		}
	})

	t.Run("nil cancelCtx does not panic", func(t *testing.T) {
		adp := &AscendDraPlugin{}

		adp.HandleError(context.Background(), errors.New("boom"), "ut-msg")
	})
}

// wrapRecoverable builds an error satisfying
// errors.Is(err, kubeletplugin.ErrRecoverable), mirroring how the kubelet
// plugin reports recoverable errors.
func wrapRecoverable(msg string) error {
	return errors.Join(kubeletplugin.ErrRecoverable, errors.New(msg))
}
