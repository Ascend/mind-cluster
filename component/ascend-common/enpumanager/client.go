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

// Package enpumanager implements the node-local HTTP API.
package enpumanager

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"regexp"
	"time"
)

const (
	// operationSucceeded is the manager result code for a successful operation.
	operationSucceeded = 0
	// SocketPath is the default node-local manager socket.
	SocketPath = "/run/enpu/enpu-manager.sock"
	// maxRequestBytes bounds the complete serialized resource request.
	maxRequestBytes = 4095
	// maxResponseBytes bounds a resource API JSON response.
	maxResponseBytes = 1 << 20
	// maxMetricsBytes bounds the Prometheus text response.
	maxMetricsBytes = 8 << 20
)

var allocationIDPattern = regexp.MustCompile("^[a-fA-F0-9]{8}(-[a-fA-F0-9]{4}){3}-[a-fA-F0-9]{12}$")

// IsValidAllocationID reports whether id has the UUID shape accepted by enpu-manager.
// Callers should validate before using an ID in a request path or a state comparison.
func IsValidAllocationID(id string) bool { return allocationIDPattern.MatchString(id) }

// Client makes resource and metrics requests over a Unix socket. Deadlines belong to the caller.
type Client struct {
	http *http.Client
}

// ConnectionError identifies a failed socket dial, as distinct from an HTTP or response error.
type ConnectionError struct {
	Err error
}

func (e *ConnectionError) Error() string { return fmt.Sprintf("connect to enpu-manager: %v", e.Err) }
func (e *ConnectionError) Unwrap() error { return e.Err }

// NewClient creates a client for the supplied socket and dial timeout.
func NewClient(socketPath string, dialTimeout time.Duration) *Client {
	dialer := &net.Dialer{Timeout: dialTimeout}
	return &Client{http: &http.Client{
		Transport: &http.Transport{
			DisableCompression: true,
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				conn, err := dialer.DialContext(ctx, "unix", socketPath)
				if err != nil {
					return nil, &ConnectionError{Err: err}
				}
				return conn, nil
			},
		},
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}}
}

// AllocateRequest contains wire values supplied by the allocation caller.
type AllocateRequest struct {
	AICoreQuota   int    `json:"aicore_quota"`
	HBMQuota      int    `json:"hbm_quota"`
	SchedPolicy   int    `json:"sched_policy"`
	PhysicalID    int    `json:"phy_id"`
	PodUID        string `json:"pod_uid"`
	ContainerName string `json:"container_name,omitempty"`
}

// AllocateResponse contains the manager's allocation ID and host configuration path.
type AllocateResponse struct {
	DieID        string `json:"die_id"`
	ShmID        string `json:"shm_id"`
	AllocationID string `json:"allocation_id"`
	ConfigPath   string `json:"config_path"`
	PhysicalID   int    `json:"phy_id"`
	VirtualID    int    `json:"vnpu_id"`
}

// Allocation is a record returned by the manager's list endpoint.
type Allocation struct {
	ConfigPath   string `json:"config_path"`
	DieID        string `json:"die_id"`
	ShmID        string `json:"shm_id"`
	AllocationID string `json:"allocation_id"`
	PhysicalID   int    `json:"phy_id"`
	VirtualID    int    `json:"vnpu_id"`
	AICoreQuota  int    `json:"aicore_quota"`
	HBMQuota     int    `json:"hbm_quota"`
	HBMLimit     int    `json:"hbm_limit"`
	SchedPolicy  int    `json:"sched_policy"`
	Metadata     struct {
		PodUID string `json:"pod_uid"`
	} `json:"metadata"`
}

// operationResult carries the manager operation status code.
type operationResult struct {
	Result *int `json:"result"`
}

func (r operationResult) check() error {
	if r.Result == nil {
		return fmt.Errorf("missing enpu-manager result")
	}
	if *r.Result != operationSucceeded {
		return fmt.Errorf("enpu-manager result %d", *r.Result)
	}
	return nil
}

// Allocate submits a creation request once and validates the creation response.
func (c *Client) Allocate(ctx context.Context, input AllocateRequest) (*AllocateResponse, error) {
	payload, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	response := struct {
		operationResult
		AllocateResponse
		PhysicalID *int `json:"phy_id"`
	}{}
	if err := c.request(ctx, http.MethodPost, "/api/v1/allocate", payload, &response); err != nil {
		return nil, err
	}
	if err := response.check(); err != nil {
		return nil, err
	}
	if !IsValidAllocationID(response.AllocationID) || response.ConfigPath == "" {
		return nil, fmt.Errorf("allocation response lacks a valid UUID or configuration path")
	}
	if response.PhysicalID == nil || *response.PhysicalID < 0 {
		return nil, fmt.Errorf("allocation response lacks a valid physical device ID")
	}
	response.AllocateResponse.PhysicalID = *response.PhysicalID
	return &response.AllocateResponse, nil
}

// ListAllocations returns a complete list or an error.
func (c *Client) ListAllocations(ctx context.Context) ([]Allocation, error) {
	response := struct {
		Allocations []struct {
			Allocation
			PhysicalID *int `json:"phy_id"`
		} `json:"allocations"`
		ErrorMsg json.RawMessage `json:"error_msg"`
	}{}
	if err := c.request(ctx, http.MethodGet, "/api/v1/allocations", nil, &response); err != nil {
		return nil, err
	}
	if len(response.ErrorMsg) != 0 || response.Allocations == nil {
		return nil, fmt.Errorf("incomplete allocation list")
	}
	allocations := make([]Allocation, 0, len(response.Allocations))
	for _, entry := range response.Allocations {
		if entry.PhysicalID == nil || *entry.PhysicalID < 0 {
			return nil, fmt.Errorf("allocation list entry lacks a valid physical device ID")
		}
		entry.Allocation.PhysicalID = *entry.PhysicalID
		allocations = append(allocations, entry.Allocation)
	}
	return allocations, nil
}

// Release requests deletion of one allocation UUID.
func (c *Client) Release(ctx context.Context, allocationID string) error {
	if !IsValidAllocationID(allocationID) {
		return fmt.Errorf("invalid allocation UUID")
	}
	var response operationResult
	if err := c.request(ctx, http.MethodDelete, "/api/v1/allocations/"+allocationID, nil, &response); err != nil {
		if isStatus(err, http.StatusNotFound) {
			return nil
		}
		return err
	}
	return response.check()
}

// httpStatus preserves a non-success HTTP response code.
type httpStatus int

func (s httpStatus) Error() string { return fmt.Sprintf("enpu-manager HTTP status %d", int(s)) }

func isStatus(err error, code int) bool {
	var status httpStatus
	return errors.As(err, &status) && int(status) == code
}

// FetchMetrics reads the manager's Prometheus text exposition.
func (c *Client) FetchMetrics(ctx context.Context) ([]byte, error) {
	return c.requestBody(ctx, http.MethodGet, "/metrics", nil, maxMetricsBytes)
}

func (c *Client) request(ctx context.Context, method, path string, payload []byte, response any) error {
	body, err := c.requestBody(ctx, method, path, payload, maxResponseBytes)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, response); err != nil {
		return fmt.Errorf("decode enpu-manager response: %w", err)
	}
	return nil
}

func (c *Client) requestBody(ctx context.Context, method, path string, payload []byte, limit int64) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, "http://localhost"+path, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	if len(payload) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}
	// Request.Write consumes the body, so measure a separate copy.
	probe := req.Clone(ctx)
	if req.GetBody != nil {
		probe.Body, err = req.GetBody()
		if err != nil {
			return nil, err
		}
	}
	var wire bytes.Buffer
	if err := probe.Write(&wire); err != nil {
		return nil, err
	}
	if wire.Len() > maxRequestBytes {
		return nil, fmt.Errorf("enpu-manager request exceeds %d bytes", maxRequestBytes)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, httpStatus(resp.StatusCode)
	}
	if resp.ContentLength > limit {
		return nil, fmt.Errorf("enpu-manager response exceeds %d bytes", limit)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > limit {
		return nil, fmt.Errorf("enpu-manager response exceeds %d bytes", limit)
	}
	return body, nil
}
