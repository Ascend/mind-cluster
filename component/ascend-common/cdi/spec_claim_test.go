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

// Package cdi — spec_claim_test.go
//
// Tests for GenerateClaimSpec and DeleteClaimSpec.
package cdi

import (
	"fmt"
	"testing"
)

// =============================================================================
// GenerateClaimSpec + DeleteClaimSpec
// =============================================================================

func TestClaimSpec_Lifecycle(t *testing.T) {
	cleanup := setupMocks()
	defer cleanup()

	// Create.
	specName, ids := writeClaimSpec(t, "claim-1", []int{0, 1}, "")
	if specName == "" {
		t.Error("specName must not be empty")
	}
	if fmt.Sprintf("%v", ids) != "[ascend.com/npu=0 ascend.com/npu=1]" {
		t.Errorf("ids = %v", ids)
	}

	// Overwrite with more devices works atomically.
	specName2, ids2 := writeClaimSpec(t, "claim-1", []int{0, 1, 2}, "")
	if specName2 == "" || len(ids2) != 3 {
		t.Errorf("overwrite failed: %q %v", specName2, ids2)
	}

	// Delete.
	if err := DeleteClaimSpec("", "claim-1"); err != nil {
		t.Fatal("delete failed:", err)
	}
}

func TestClaimSpec_Errors(t *testing.T) {
	t.Run("empty claimUID", func(t *testing.T) {
		_, _, err := GenerateClaimSpec(BuildSpecConfig{DeviceConfig: DeviceConfig{DeviceIDs: []int{0}, DevType: "Ascend910"}, MountConfig: testMountSource(t)}, "")
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestClaimSpec_EmptyLogicIDs(t *testing.T) {
	_, ids, err := GenerateClaimSpec(BuildSpecConfig{DeviceConfig: DeviceConfig{DeviceIDs: []int{}, DevType: "Ascend910"}, MountConfig: testMountSource(t)}, "claim-empty")
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 0 {
		t.Errorf("got %d ids", len(ids))
	}
}

func TestClaimSpec_AutoCreateDir(t *testing.T) {
	cleanup := setupMocks()
	defer cleanup()

	// Directory creation is now managed by the CDI library's default cache.
	specName, ids := writeClaimSpec(t, "claim-dir", []int{0}, "")
	if specName == "" || len(ids) != 1 {
		t.Fatalf("unexpected result: specName=%q ids=%v", specName, ids)
	}
}

func TestDeleteClaimSpec_EdgeCases(t *testing.T) {
	t.Run("non-existent spec", func(t *testing.T) {
		if err := DeleteClaimSpec("", "x"); err != nil {
			t.Fatal("should be idempotent:", err)
		}
	})
	t.Run("empty claimUID", func(t *testing.T) {
		if err := DeleteClaimSpec("", ""); err != nil {
			t.Fatal(err)
		}
	})
}
