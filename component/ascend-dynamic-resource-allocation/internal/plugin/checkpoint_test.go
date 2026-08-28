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
	"reflect"
	"strings"
	"testing"

	drapbv1 "k8s.io/kubelet/pkg/apis/dra/v1beta1"
)

// newTestCheckpoint builds a checkpoint with two prepared claims used as the
// shared fixture for marshal/unmarshal tests.
func newTestCheckpoint() *Checkpoint {
	cp := newCheckpoint()
	cp.V1.PreparedClaims["ut-claim-a"] = PreparedDevices{
		{Device: drapbv1.Device{
			PoolName: "pool-a", DeviceName: "Ascend910-0",
			CdiDeviceIds: []string{"ascend.com/npu=0"},
		}},
		{Device: drapbv1.Device{
			PoolName: "pool-a", DeviceName: "Ascend910-1",
		}},
	}
	cp.V1.PreparedClaims["ut-claim-b"] = PreparedDevices{
		{Device: drapbv1.Device{PoolName: "pool-b", DeviceName: "Ascend910-7"}},
	}
	return cp
}

// TestNewCheckpoint verifies the constructor returns an empty but fully
// initialized checkpoint.
func TestNewCheckpoint(t *testing.T) {
	cp := newCheckpoint()

	if cp.Checksum != 0 {
		t.Errorf("Checksum = %d, want 0", cp.Checksum)
	}
	if cp.V1 == nil {
		t.Fatal("V1 = nil, want non-nil")
	}
	if cp.V1.PreparedClaims == nil {
		t.Fatal("PreparedClaims = nil, want initialized empty map")
	}
	if len(cp.V1.PreparedClaims) != 0 {
		t.Errorf("len(PreparedClaims) = %d, want 0", len(cp.V1.PreparedClaims))
	}
}

// TestMarshalUnmarshalRoundTrip verifies that a marshalled checkpoint can be
// unmarshalled back into an equivalent checkpoint that passes checksum
// verification.
func TestMarshalUnmarshalRoundTrip(t *testing.T) {
	cp := newTestCheckpoint()

	data, err := cp.MarshalCheckpoint()
	if err != nil {
		t.Fatalf("MarshalCheckpoint() error = %v, want nil", err)
	}
	if cp.Checksum == 0 {
		t.Error("Checksum after MarshalCheckpoint() = 0, want non-zero")
	}

	got := &Checkpoint{}
	if err := got.UnmarshalCheckpoint(data); err != nil {
		t.Fatalf("UnmarshalCheckpoint() error = %v, want nil", err)
	}
	if len(got.V1.PreparedClaims) != 2 {
		t.Fatalf("len(PreparedClaims) = %d, want 2", len(got.V1.PreparedClaims))
	}
	wantNames := []string{"Ascend910-0", "Ascend910-1"}
	if names := got.V1.PreparedClaims["ut-claim-a"].DeviceNames(); !reflect.DeepEqual(names, wantNames) {
		t.Errorf("DeviceNames(claim-a) = %v, want %v", names, wantNames)
	}
	if err := got.VerifyChecksum(); err != nil {
		t.Errorf("VerifyChecksum() after round trip = %v, want nil", err)
	}
}

// TestUnmarshalCheckpoint_Errors verifies that UnmarshalCheckpoint rejects
// malformed payloads.
func TestUnmarshalCheckpoint_Errors(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{"empty data", ""},
		{"garbage", "not-json"},
		{"truncated json", `{"checksum":123,"v1":{"preparedClaims":{}}`},
		{"checksum wrong type", `{"checksum":"not-a-number"}`},
		{"v1 wrong type", `{"v1":5}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cp := &Checkpoint{}

			err := cp.UnmarshalCheckpoint([]byte(tt.data))
			if err == nil {
				t.Fatalf("UnmarshalCheckpoint(%q) = nil, want error", tt.data)
			}
		})
	}
}

// TestVerifyChecksum verifies checksum verification succeeds for intact data
// and fails when the payload or the stored checksum is tampered with.
func TestVerifyChecksum(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(cp *Checkpoint)
		wantOK bool
	}{
		{
			"intact checkpoint",
			func(cp *Checkpoint) {},
			true,
		},
		{
			"payload modified after marshal",
			func(cp *Checkpoint) {
				cp.V1.PreparedClaims["ut-claim-a"][0].DeviceName = "Ascend910-9"
			},
			false,
		},
		{
			"claim added after marshal",
			func(cp *Checkpoint) {
				cp.V1.PreparedClaims["ut-claim-c"] = nil
			},
			false,
		},
		{
			"checksum tampered",
			func(cp *Checkpoint) { cp.Checksum++ },
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cp := newTestCheckpoint()
			if _, err := cp.MarshalCheckpoint(); err != nil {
				t.Fatalf("MarshalCheckpoint() error = %v, want nil", err)
			}

			tt.mutate(cp)

			err := cp.VerifyChecksum()
			if tt.wantOK && err != nil {
				t.Errorf("VerifyChecksum() = %v, want nil", err)
			}
			if !tt.wantOK && err == nil {
				t.Errorf("VerifyChecksum() = nil, want checksum mismatch error")
			}
			if !tt.wantOK && err != nil && !strings.Contains(err.Error(), "checkpoint is corrupted") {
				t.Errorf("VerifyChecksum() error = %v, want containing %q", err, "checkpoint is corrupted")
			}
		})
	}
}
