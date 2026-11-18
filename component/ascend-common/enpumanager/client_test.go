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

package enpumanager

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

const testAllocationID = "f312ad4d-d12a-4eca-a2c3-84a1c9f17ba7"

func TestResourceClientWireContract(t *testing.T) {
	var createCalls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/allocate":
			createCalls++
			var req struct {
				AllocateRequest
				Container string `json:"container_name"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.PodUID != "pod-1" || req.Container != "infer" ||
				req.PhysicalID != 0 || req.AICoreQuota != 30 || req.HBMQuota != 256 || req.SchedPolicy != 1 {
				t.Errorf("unexpected allocation request: %+v, %v", req, err)
			}
			_, _ = fmt.Fprintf(w, `{"result":0,"allocation_id":%q,"phy_id":0,"config_path":"/var/test/npu_info.config"}`, testAllocationID)
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/allocations":
			_, _ = fmt.Fprintf(w, `{"allocations":[{"allocation_id":%q,"phy_id":0,"metadata":{"pod_uid":"pod-1"}}]}`, testAllocationID)
		case r.Method == http.MethodDelete && r.URL.Path == "/api/v1/allocations/"+testAllocationID:
			_, _ = fmt.Fprint(w, `{"result":0}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	c := testResourceClient(server)
	ctx := context.Background()
	result, err := c.Allocate(ctx, AllocateRequest{
		PodUID: "pod-1", ContainerName: "infer", AICoreQuota: 30, HBMQuota: 256, SchedPolicy: 1,
	})
	if err != nil || result.AllocationID != testAllocationID || result.ConfigPath == "" || createCalls != 1 {
		t.Fatalf("unexpected allocation: %+v, %v (calls=%d)", result, err, createCalls)
	}
	allocations, err := c.ListAllocations(ctx)
	if err != nil || len(allocations) != 1 || allocations[0].Metadata.PodUID != "pod-1" {
		t.Fatalf("unexpected allocations: %+v, %v", allocations, err)
	}
	if err := c.Release(ctx, allocations[0].AllocationID); err != nil {
		t.Fatal(err)
	}
}

func TestAllocateRejectsMissingPhysicalDevice(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprintf(w, `{"result":0,"allocation_id":%q,"config_path":"/config/npu_info.config"}`, testAllocationID)
	}))
	defer server.Close()
	result, err := testResourceClient(server).Allocate(context.Background(), AllocateRequest{})
	if err == nil || result != nil {
		t.Fatalf("expected missing phy_id error, got %+v, %v", result, err)
	}
}

func TestResourceClientRejectsUnusableResponses(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_, _ = fmt.Fprint(w, `{"allocations":[],"error_msg":"list truncated"}`)
		case http.MethodPost:
			_, _ = fmt.Fprint(w, `{"allocation_id":"`+testAllocationID+`","config_path":"/var/test/npu_info.config"}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	c := testResourceClient(server)
	if list, err := c.ListAllocations(context.Background()); err == nil || list != nil {
		t.Fatalf("truncated response returned %v, %v", list, err)
	}
	if result, err := c.Allocate(context.Background(), AllocateRequest{}); err == nil || result != nil {
		t.Fatalf("response missing result returned %v, %v", result, err)
	}
	if err := c.Release(context.Background(), testAllocationID); err != nil {
		t.Fatalf("HTTP 404 should be treated as already released: %v", err)
	}
	_, err := NewClient(filepath.Join(t.TempDir(), "absent.sock"), 0).ListAllocations(context.Background())
	var connectErr *ConnectionError
	if !errors.As(err, &connectErr) {
		t.Fatalf("expected a UDS connection error, got %v", err)
	}
}

func TestIsValidAllocationID(t *testing.T) {
	valid := []string{
		testAllocationID,
		strings.ToUpper(testAllocationID),
		// Shape-only check: a nil UUID and non-RFC-4122 version nibbles are still accepted.
		"00000000-0000-0000-0000-000000000000",
	}
	for _, input := range valid {
		if !IsValidAllocationID(input) {
			t.Errorf("IsValidAllocationID(%q) = false, want true", input)
		}
	}
	invalid := []string{
		"",
		"f312ad4dd12a4ecaa2c384a1c9f17ba7",
		"f312ad4d-d12a-4eca-a2c3-84a1c9f17ba",
		"f312ad4d-d12a-4eca-a2c3-84a1c9f17ba7\n",
		"../../etc/passwd",
		"f312ad4d/d12a/4eca/a2c3/84a1c9f17ba7",
	}
	for _, input := range invalid {
		if IsValidAllocationID(input) {
			t.Errorf("IsValidAllocationID(%q) = true, want false", input)
		}
	}
}

func testResourceClient(server *httptest.Server) *Client {
	return &Client{http: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		clone := r.Clone(r.Context())
		clone.URL.Host = server.Listener.Addr().String()
		return server.Client().Transport.RoundTrip(clone)
	})}}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestFetchMetricsReadsSocketAndRejectsHTTPFailure(t *testing.T) {
	// Use a relative socket path so a long CI checkout does not exceed sockaddr_un's limit.
	dir, err := os.MkdirTemp(".", "uds-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Remove(dir); err != nil {
			t.Error(err)
		}
	})
	socket := filepath.Join(dir, "m.sock")
	var status atomic.Int32
	listener, err := net.Listen("unix", socket)
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/metrics" {
			http.NotFound(w, r)
			return
		}
		if code := status.Load(); code != 0 {
			w.WriteHeader(int(code))
			return
		}
		_, _ = io.WriteString(w, "metric 1\n")
	}))
	_ = server.Listener.Close()
	server.Listener = listener
	server.Start()
	defer func() {
		server.Close()
		if err := os.Remove(socket); err != nil && !os.IsNotExist(err) {
			t.Error(err)
		}
	}()

	body, err := NewClient(socket, 0).FetchMetrics(context.Background())
	if err != nil || string(body) != "metric 1\n" {
		t.Fatalf("unexpected response: %q, %v", body, err)
	}
	status.Store(http.StatusServiceUnavailable)
	if _, err := NewClient(socket, 0).FetchMetrics(context.Background()); !isStatus(err, http.StatusServiceUnavailable) {
		t.Fatal("expected HTTP error")
	}
}

func TestFetchMetricsRejectsOversizedBodyWithoutContentLength(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		_, _ = fmt.Fprint(w, strings.Repeat("x", maxMetricsBytes+1))
	}))
	defer server.Close()
	f := testResourceClient(server)
	if _, err := f.FetchMetrics(context.Background()); err == nil {
		t.Fatal("expected response size error")
	}
}
