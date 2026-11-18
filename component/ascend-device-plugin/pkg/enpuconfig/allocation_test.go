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

package enpuconfig

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"ascend-common/enpumanager"
)

const testAllocationID = "f312ad4d-d12a-4eca-a2c3-84a1c9f17ba7"

func testResourceClient(server *httptest.Server) *Provider {
	return &Provider{ResourceClient: enpumanager.NewClient(server.Listener.Addr().String(), 0)}
}
func newUnixServer(t *testing.T, handler http.Handler) *httptest.Server {
	t.Helper()
	dir, err := os.MkdirTemp(".", "uds-")
	if err != nil {
		t.Fatal(err)
	}
	socket := filepath.Join(dir, "manager.sock")
	listener, err := net.Listen("unix", socket)
	if err != nil {
		os.Remove(dir)
		t.Fatal(err)
	}
	server := httptest.NewUnstartedServer(handler)
	server.Listener.Close()
	server.Listener = listener
	server.Start()
	t.Cleanup(func() { server.Close(); os.Remove(socket); os.Remove(dir) })
	return server
}
func TestEnsureAllocationCreatesThenReuses(t *testing.T) {
	const uid = "pod-1"
	const config = "/etc/enpu/vcann-rt/allocations/example/npu_info.config"
	const id = testAllocationID
	var creates atomic.Int32
	server := newUnixServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/allocations":
			if creates.Load() == 0 {
				_, _ = fmt.Fprint(w, `{"allocations":[]}`)
				return
			}
			_, _ = fmt.Fprintf(w, `{"allocations":[{"allocation_id":%q,"config_path":%q,"phy_id":0,
				"vnpu_id":5,"aicore_quota":30,"metadata":{"pod_uid":%q}}]}`, id, config, uid)
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/allocate":
			creates.Add(1)
			var req struct {
				enpumanager.AllocateRequest
				AllocationID json.RawMessage `json:"allocation_id"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.AllocationID != nil ||
				req.PodUID != uid || req.PhysicalID != 0 || req.AICoreQuota != 30 {
				t.Errorf("unexpected creation request: %+v, %v", req, err)
			}
			_, _ = fmt.Fprintf(w, `{"result":0,"allocation_id":%q,"config_path":%q,"phy_id":0,"vnpu_id":5}`, id, config)
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	c := testResourceClient(server)
	for i := 0; i < 2; i++ {
		result, err := c.EnsureAllocation(context.Background(), enpumanager.AllocateRequest{PodUID: uid, AICoreQuota: 30})
		if err != nil || result.AllocationID != id || result.ConfigPath != config || result.VirtualID != 5 ||
			result.Reused != (i == 1) {
			t.Fatalf("unexpected allocation: %+v, %v", result, err)
		}
	}
	if creates.Load() != 1 {
		t.Fatalf("expected one creation, got %d", creates.Load())
	}
}

func TestEnsureAllocationRejectsUnusableLookup(t *testing.T) {
	const record = `{"allocation_id":"` + testAllocationID + `","phy_id":0,"metadata":{"pod_uid":"pod-1"}}`
	for _, tt := range []struct {
		name, response string
	}{
		{"incomplete list", `{"allocations":[],"error_msg":"truncated"}`},
		{"duplicate pod UID", `{"allocations":[` + record + `,` + record + `]}`},
		{"missing physical device", `{"allocations":[{"allocation_id":"` + testAllocationID + `","metadata":{"pod_uid":"pod-1"}}]}`},
		{"invalid allocation ID", `{"allocations":[{"allocation_id":"invalid","phy_id":0,"metadata":{"pod_uid":"pod-1"}}]}`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			server := newUnixServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("must not create an allocation after an unusable lookup, got %s", r.Method)
					http.NotFound(w, r)
					return
				}
				_, _ = fmt.Fprint(w, tt.response)
			}))
			defer server.Close()
			result, err := testResourceClient(server).EnsureAllocation(context.Background(), enpumanager.AllocateRequest{PodUID: "pod-1"})
			if err == nil || result != nil {
				t.Fatalf("expected lookup error, got %+v, %v", result, err)
			}
		})
	}
}

func TestCleanConfigPath(t *testing.T) {
	for _, input := range []string{"", "relative", "/config", "/config-other/npu_info.config",
		"/config/../outside", "/config/inside/../../outside"} {
		if _, err := CleanConfigPath("/config", input); err == nil {
			t.Errorf("accepted path %q", input)
		}
	}
	result, err := CleanConfigPath("/config", "/config/./job/npu_info.config")
	if err != nil || result != "/config/job/npu_info.config" {
		t.Fatalf("unexpected path %q, %v", result, err)
	}
}

// Use the real HTTP client with the test server to exercise wire encoding.
