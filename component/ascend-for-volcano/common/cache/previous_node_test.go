/*
Copyright(C)2026. Huawei Technologies Co.,Ltd. All rights reserved.

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

package cache

import (
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/types"
)

func TestRecordAndGetPreferredNode(t *testing.T) {
	cache := NewPodNodeAffinityCache()

	cache.RecordAssignment(types.UID("owner-1"), "0", "node-a")
	cache.RecordAssignment(types.UID("owner-1"), "1", "node-b")
	cache.RecordAssignment(types.UID("owner-2"), "0", "node-c")

	if node := cache.GetPreferredNode(types.UID("owner-1"), "0"); node != "node-a" {
		t.Errorf("expected node-a, got %s", node)
	}
	if node := cache.GetPreferredNode(types.UID("owner-1"), "1"); node != "node-b" {
		t.Errorf("expected node-b, got %s", node)
	}
	if node := cache.GetPreferredNode(types.UID("owner-2"), "0"); node != "node-c" {
		t.Errorf("expected node-c, got %s", node)
	}
	if node := cache.GetPreferredNode(types.UID("owner-1"), "9"); node != "" {
		t.Errorf("expected empty string, got %s", node)
	}
	if node := cache.GetPreferredNode(types.UID("owner-99"), "0"); node != "" {
		t.Errorf("expected empty string, got %s", node)
	}
}

func TestRecordOverwrite(t *testing.T) {
	cache := NewPodNodeAffinityCache()

	cache.RecordAssignment(types.UID("owner-1"), "0", "node-a")
	cache.RecordAssignment(types.UID("owner-1"), "0", "node-b")

	if node := cache.GetPreferredNode(types.UID("owner-1"), "0"); node != "node-b" {
		t.Errorf("expected node-b after overwrite, got %s", node)
	}
}

func TestSize(t *testing.T) {
	cache := NewPodNodeAffinityCache()

	if cache.Size() != 0 {
		t.Errorf("expected 0, got %d", cache.Size())
	}

	cache.RecordAssignment(types.UID("owner-1"), "0", "node-a")
	cache.RecordAssignment(types.UID("owner-1"), "1", "node-b")

	if cache.Size() != 2 {
		t.Errorf("expected 2, got %d", cache.Size())
	}
}

func TestEvictExpired(t *testing.T) {
	cache := NewPodNodeAffinityCache()

	cache.RecordAssignment(types.UID("owner-active"), "0", "node-a")

	// Manually set an old timestamp to simulate a deleted owner
	cache.OwnerToRankNodes["owner-stale"] = map[string]*RankNodeEntry{"0": {Node: "node-b"}}
	cache.UpdateTime["owner-stale"] = time.Now().Unix() - 100000

	cache.EvictExpired(24 * time.Hour)

	if node := cache.GetPreferredNode(types.UID("owner-active"), "0"); node != "node-a" {
		t.Errorf("expected node-a to survive, got %s", node)
	}
	if node := cache.GetPreferredNode(types.UID("owner-stale"), "0"); node != "" {
		t.Errorf("expected empty for expired owner, got %s", node)
	}
	if cache.Size() != 1 {
		t.Errorf("expected 1 entry remaining, got %d", cache.Size())
	}
}
