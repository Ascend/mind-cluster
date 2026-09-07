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

package kubeclient

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	fake "k8s.io/client-go/kubernetes/fake"
	clientgotesting "k8s.io/client-go/testing"

	"ascend-common/common-utils/hwlog"
	draFlags "ascend-dynamic-resource-allocation/internal/flags"
)

// init initializes the run logger so error paths exercised by the tests can
// write logs instead of panicking on a nil hwlog.RunLog.
func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwlog.InitRunLogger(&hwLogConfig, context.Background())
}

// newFakeClientK8s builds a ClientK8s backed by a fake clientset whose
// patch-nodes reactor fails with the given errors in order before succeeding,
// and records every patch action it receives.
func newFakeClientK8s(nodeName string, patchErrs []error) (*ClientK8s, *[]clientgotesting.PatchAction) {
	node := &corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: nodeName}}
	clientset := fake.NewSimpleClientset(node)
	var actions []clientgotesting.PatchAction
	clientset.Fake.PrependReactor("patch", "nodes",
		func(action clientgotesting.Action) (bool, runtime.Object, error) {
			actions = append(actions, action.(clientgotesting.PatchAction))
			if len(actions) <= len(patchErrs) {
				return true, nil, patchErrs[len(actions)-1]
			}
			return true, node, nil
		})
	return &ClientK8s{ClientSet: clientset, NodeName: nodeName}, &actions
}

// TestAddAnnotation verifies the retry behavior of AddAnnotation: the patch is
// retried until it succeeds or retryTime attempts are exhausted.
func TestAddAnnotation(t *testing.T) {
	const nodeName = "ut-node"

	tests := []struct {
		name        string
		patchErrs   []error
		wantErr     string // empty means success is expected
		wantAttempt int
	}{
		{
			name:        "patch succeeds on first attempt",
			patchErrs:   nil,
			wantAttempt: 1,
		},
		{
			name:        "patch succeeds after two failures",
			patchErrs:   []error{errors.New("api server timeout"), errors.New("connection refused")},
			wantAttempt: retryTime,
		},
		{
			name:        "all patch attempts fail",
			patchErrs:   []error{errors.New("err 1"), errors.New("err 2"), errors.New("err 3")},
			wantErr:     "err 3",
			wantAttempt: retryTime,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, actions := newFakeClientK8s(nodeName, tt.patchErrs)

			err := client.AddAnnotation("huawei.com/ascend-dra.version", "ut-value")
			if got := len(*actions); got != tt.wantAttempt {
				t.Errorf("AddAnnotation() made %d patch attempts, want %d", got, tt.wantAttempt)
			}
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("AddAnnotation() error = %v, want nil", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("AddAnnotation() error = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}

// TestAddAnnotation_PatchPayload verifies the JSON patch document sent to the
// API server, including the JSON Pointer escaping of the annotation key.
func TestAddAnnotation_PatchPayload(t *testing.T) {
	const nodeName = "ut-node"
	const key = "huawei.com/ascend-dra.version"
	const value = `{"version":"v6.0.0"}`

	client, actions := newFakeClientK8s(nodeName, nil)

	if err := client.AddAnnotation(key, value); err != nil {
		t.Fatalf("AddAnnotation() error = %v, want nil", err)
	}

	if got := len(*actions); got != 1 {
		t.Fatalf("got %d patch actions, want 1", got)
	}
	patchAction := (*actions)[0]
	if patchAction.GetName() != nodeName {
		t.Errorf("patch target node = %q, want %q", patchAction.GetName(), nodeName)
	}
	if patchAction.GetPatchType() != types.JSONPatchType {
		t.Errorf("patch type = %v, want %v", patchAction.GetPatchType(), types.JSONPatchType)
	}

	var patch []map[string]interface{}
	if err := json.Unmarshal(patchAction.GetPatch(), &patch); err != nil {
		t.Fatalf("Unmarshal patch %s error = %v", patchAction.GetPatch(), err)
	}
	if len(patch) != 1 {
		t.Fatalf("patch document has %d operations, want 1", len(patch))
	}
	// The '/' inside the annotation key must be escaped as '~1' per RFC 6902.
	wantPatch := map[string]interface{}{
		"op":    "add",
		"path":  "/metadata/annotations/huawei.com~1ascend-dra.version",
		"value": value,
	}
	for k, want := range wantPatch {
		got, ok := patch[0][k]
		if !ok {
			t.Errorf("patch operation misses %q field, got %v", k, patch[0])
			continue
		}
		if got != want {
			t.Errorf("patch %q = %v, want %v", k, got, want)
		}
	}
}

// TestNewClientK8s verifies ClientK8s creation from a kubeconfig file and
// failure outside a cluster without a kubeconfig.
func TestNewClientK8s(t *testing.T) {
	t.Run("success from kubeconfig file", func(t *testing.T) {
		cfg := &draFlags.KubeClientConfig{
			KubeConfig:   writeTestKubeconfig(t),
			KubeAPIQPS:   20,
			KubeAPIBurst: 40,
		}

		client, err := NewClientK8s(cfg, "ut-node")
		if err != nil {
			t.Fatalf("NewClientK8s() error = %v, want nil", err)
		}
		if client.ClientSet == nil {
			t.Error("NewClientK8s().ClientSet is nil, want non-nil clientset")
		}
		if client.NodeName != "ut-node" {
			t.Errorf("NewClientK8s().NodeName = %q, want %q", client.NodeName, "ut-node")
		}
	})

	t.Run("error outside cluster without kubeconfig", func(t *testing.T) {
		t.Setenv("KUBERNETES_SERVICE_HOST", "")
		t.Setenv("KUBERNETES_SERVICE_PORT", "")

		client, err := NewClientK8s(&draFlags.KubeClientConfig{}, "ut-node")
		if err == nil {
			t.Fatalf("NewClientK8s() = %+v, want error", client)
		}
		if !strings.Contains(err.Error(), "create client configuration") {
			t.Errorf("NewClientK8s() error = %v, want containing %q", err, "create client configuration")
		}
	})
}

// writeTestKubeconfig writes a minimal valid kubeconfig file and returns
// its path.
func writeTestKubeconfig(t *testing.T) string {
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
