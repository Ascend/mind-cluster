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
	"reflect"
	"testing"
	"time"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
)

func TestRecordPodErrorAndPodErrorCards(t *testing.T) {
	ResetPodError()
	defer ResetPodError()

	now := time.Now().Unix()
	RecordPodError("node-1", []string{"Ascend910-0", "Ascend910-1"}, now)
	want := map[string]int64{"Ascend910-0": now, "Ascend910-1": now}
	if got := PodErrorCards("node-1"); !reflect.DeepEqual(got, want) {
		t.Fatalf("PodErrorCards() = %v, want %v", got, want)
	}

	// a later failure time overwrites an earlier one on the same card
	RecordPodError("node-1", []string{"Ascend910-0"}, now+10)
	want = map[string]int64{"Ascend910-0": now + 10, "Ascend910-1": now}
	if got := PodErrorCards("node-1"); !reflect.DeepEqual(got, want) {
		t.Fatalf("PodErrorCards() after overwrite = %v, want %v", got, want)
	}
}

func TestRecordPodErrorSkipsInvalidInput(t *testing.T) {
	ResetPodError()
	defer ResetPodError()

	now := time.Now().Unix()
	RecordPodError("", []string{"Ascend910-0"}, now)
	RecordPodError("node-1", nil, now)
	RecordPodError("node-1", []string{"Ascend910-0"}, 0)
	if got := PodErrorCards("node-1"); len(got) != 0 {
		t.Fatalf("expected no records for invalid input, got %v", got)
	}
}

func TestResetPodError(t *testing.T) {
	ResetPodError()
	defer ResetPodError()

	RecordPodError("node-1", []string{"Ascend910-0"}, time.Now().Unix())
	if got := PodErrorCards("node-1"); len(got) != 1 {
		t.Fatalf("expected 1 record before reset, got %v", got)
	}
	ResetPodError()
	if got := PodErrorCards("node-1"); len(got) != 0 {
		t.Fatalf("expected no records after reset, got %v", got)
	}
}

func TestExpirePodErrorLocked(t *testing.T) {
	ResetPodError()
	defer ResetPodError()

	now := int64(1_000_000)
	expired := now - util.MaxPodErrorToleranceAfterTime - 1
	kept := now - util.MaxPodErrorToleranceAfterTime + 1

	podErrorCache.mu.Lock()
	podErrorCache.records = map[string]map[string]int64{
		"node-1": {"expired-card": expired, "kept-card": kept},
		"node-2": {"expired-card": expired},
	}
	podErrorCache.mu.Unlock()

	expirePodErrorLocked(now)

	podErrorCache.mu.Lock()
	defer podErrorCache.mu.Unlock()

	node1, ok := podErrorCache.records["node-1"]
	if !ok || len(node1) != 1 || node1["kept-card"] != kept {
		t.Fatalf("node-1 after expire = %v, want only kept-card=%d", node1, kept)
	}
	if _, ok := podErrorCache.records["node-2"]; ok {
		t.Fatalf("node-2 should be fully expired and removed")
	}
}
