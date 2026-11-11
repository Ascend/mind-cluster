/*
Copyright(C) 2026. Huawei Technologies CO.,Ltd. All rights reserved.

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

package workload

import (
	"slices"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"infer-operator/pkg/common"
)

// fakeWorkLoad implements WorkLoadInterface for pure function unit tests.
// It does not depend on a K8s client or informer.
type fakeWorkLoad struct {
	objectMeta metav1.ObjectMeta
	ready      bool
}

func (f *fakeWorkLoad) GetWorkLoadObjMeta() metav1.ObjectMeta {
	return f.objectMeta
}

func (f *fakeWorkLoad) SetWorkLoadObjMeta(objectMeta metav1.ObjectMeta) {
	f.objectMeta = objectMeta
}

func (f *fakeWorkLoad) GetWorkLoadReplicas() int32 {
	return common.DefaultReplicas
}

func (f *fakeWorkLoad) IsWorkLoadReady() bool {
	return f.ready
}

// readyWorkLoad builds a ready instance, index is the value of the instance index label.
func readyWorkLoad(index string) *fakeWorkLoad {
	return &fakeWorkLoad{
		objectMeta: metav1.ObjectMeta{
			Name:   "workload-" + index,
			Labels: map[string]string{common.InstanceIndexLabelKey: index},
		},
		ready: true,
	}
}

// unreadyWorkLoad builds an unready instance.
func unreadyWorkLoad(index string) *fakeWorkLoad {
	return &fakeWorkLoad{
		objectMeta: metav1.ObjectMeta{
			Name:   "workload-" + index,
			Labels: map[string]string{common.InstanceIndexLabelKey: index},
		},
		ready: false,
	}
}

// invalidWorkLoad builds an instance whose index label is missing or illegal.
func invalidWorkLoad(index string, withLabel bool) *fakeWorkLoad {
	meta := metav1.ObjectMeta{Name: "workload-invalid"}
	if withLabel {
		meta.Labels = map[string]string{common.InstanceIndexLabelKey: index}
	}
	return &fakeWorkLoad{objectMeta: meta, ready: true}
}

func TestParseInstanceIndex(t *testing.T) {
	tests := []struct {
		name      string
		meta      metav1.ObjectMeta
		wantIndex int
		wantOK    bool
	}{
		{
			name:      "valid positive index",
			meta:      metav1.ObjectMeta{Labels: map[string]string{common.InstanceIndexLabelKey: "3"}},
			wantIndex: 3,
			wantOK:    true,
		},
		{
			name:      "valid zero index",
			meta:      metav1.ObjectMeta{Labels: map[string]string{common.InstanceIndexLabelKey: "0"}},
			wantIndex: 0,
			wantOK:    true,
		},
		{
			name:      "valid negative index",
			meta:      metav1.ObjectMeta{Labels: map[string]string{common.InstanceIndexLabelKey: "-1"}},
			wantIndex: -1,
			wantOK:    true,
		},
		{
			name:   "missing index label",
			meta:   metav1.ObjectMeta{Labels: map[string]string{"other": "0"}},
			wantOK: false,
		},
		{
			name:   "nil labels",
			meta:   metav1.ObjectMeta{},
			wantOK: false,
		},
		{
			name:   "non integer index label",
			meta:   metav1.ObjectMeta{Labels: map[string]string{common.InstanceIndexLabelKey: "abc"}},
			wantOK: false,
		},
		{
			name:   "empty index label",
			meta:   metav1.ObjectMeta{Labels: map[string]string{common.InstanceIndexLabelKey: ""}},
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			index, ok := parseInstanceIndex(tt.meta)
			if ok != tt.wantOK {
				t.Fatalf("parseInstanceIndex() ok = %v, want %v", ok, tt.wantOK)
			}
			if ok && index != tt.wantIndex {
				t.Fatalf("parseInstanceIndex() index = %d, want %d", index, tt.wantIndex)
			}
		})
	}
}

func TestSelectIndicesToDelete(t *testing.T) {
	tests := []struct {
		name       string
		workloads  []WorkLoadInterface
		indexLimit int
		want       []int
	}{
		{
			name: "UT-D1 shrink count not positive keeps all instances",
			workloads: []WorkLoadInterface{
				readyWorkLoad("0"), readyWorkLoad("1"), readyWorkLoad("2"), readyWorkLoad("-1"),
			},
			indexLimit: 4,
			want:       nil,
		},
		{
			name: "UT-D2 unready count reaches shrink count",
			workloads: []WorkLoadInterface{
				readyWorkLoad("0"), unreadyWorkLoad("1"), readyWorkLoad("2"), readyWorkLoad("3"),
				unreadyWorkLoad("4"), readyWorkLoad("5"), readyWorkLoad("6"),
			},
			indexLimit: 5,
			want:       []int{4, 1},
		},
		{
			name: "UT-D3 unready fewer than shrink count filled by ready",
			workloads: []WorkLoadInterface{
				readyWorkLoad("0"), readyWorkLoad("1"), unreadyWorkLoad("2"), readyWorkLoad("3"),
				readyWorkLoad("4"), readyWorkLoad("5"), readyWorkLoad("6"), readyWorkLoad("7"),
			},
			indexLimit: 5,
			want:       []int{2, 7, 6},
		},
		{
			name: "UT-D4 all ready degrades to highest indices",
			workloads: []WorkLoadInterface{
				readyWorkLoad("0"), readyWorkLoad("1"), readyWorkLoad("2"), readyWorkLoad("3"),
				readyWorkLoad("4"), readyWorkLoad("5"), readyWorkLoad("6"), readyWorkLoad("7"),
			},
			indexLimit: 5,
			want:       []int{7, 6, 5},
		},
		{
			name: "UT-D5 all unready deletes highest indices",
			workloads: []WorkLoadInterface{
				unreadyWorkLoad("0"), unreadyWorkLoad("1"), unreadyWorkLoad("2"), unreadyWorkLoad("3"),
				unreadyWorkLoad("4"), unreadyWorkLoad("5"), unreadyWorkLoad("6"), unreadyWorkLoad("7"),
			},
			indexLimit: 5,
			want:       []int{7, 6, 5},
		},
		{
			name: "UT-D6 invalid or missing index label skipped",
			workloads: []WorkLoadInterface{
				readyWorkLoad("0"), readyWorkLoad("1"), readyWorkLoad("2"),
				invalidWorkLoad("abc", true), invalidWorkLoad("", false),
			},
			indexLimit: 3,
			want:       []int{2, 1},
		},
		{
			name: "UT-D7 negative index outranks unready ones",
			workloads: []WorkLoadInterface{
				readyWorkLoad("-1"), readyWorkLoad("0"), unreadyWorkLoad("1"),
				readyWorkLoad("2"), unreadyWorkLoad("3"),
			},
			indexLimit: 3,
			want:       []int{-1, 3},
		},
		{
			name: "UT-D8 output follows the delete priority order",
			workloads: []WorkLoadInterface{
				readyWorkLoad("0"), unreadyWorkLoad("1"), readyWorkLoad("2"),
				unreadyWorkLoad("3"), readyWorkLoad("4"), readyWorkLoad("5"),
			},
			indexLimit: 2,
			want:       []int{3, 1, 5, 4},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := selectIndicesToDelete(tt.workloads, tt.indexLimit)
			if !slices.Equal(got, tt.want) {
				t.Fatalf("selectIndicesToDelete() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSelectIndicesToCreate(t *testing.T) {
	tests := []struct {
		name           string
		currentIndices []int
		target         int
		want           []int
	}{
		{
			name:           "UT-C1 current count equals target",
			currentIndices: []int{0, 1, 2, 3, 4},
			target:         5,
			want:           nil,
		},
		{
			name:           "UT-C2 current count larger than target",
			currentIndices: []int{0, 1, 2, 3, 4},
			target:         3,
			want:           nil,
		},
		{
			name:           "UT-C3 fill gaps first",
			currentIndices: []int{0, 2, 3},
			target:         5,
			want:           []int{1, 4},
		},
		{
			name:           "UT-C4 fill gaps then append after max index",
			currentIndices: []int{0, 2},
			target:         5,
			want:           []int{1, 3, 4},
		},
		{
			name:           "UT-C5 empty current indices",
			currentIndices: []int{},
			target:         3,
			want:           []int{0, 1, 2},
		},
		{
			name:           "UT-C6 duplicated indices are deduplicated",
			currentIndices: []int{0, 0, 2, 2, 3},
			target:         5,
			want:           []int{1, 4},
		},
		{
			name:           "UT-C7 terminating index counts and is not a gap",
			currentIndices: []int{0, 1, 2},
			target:         4,
			want:           []int{3},
		},
		{
			name:           "UT-C8 target is zero",
			currentIndices: []int{0, 1, 2},
			target:         0,
			want:           nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SelectIndicesToCreate(tt.currentIndices, tt.target)
			if !slices.Equal(got, tt.want) {
				t.Fatalf("SelectIndicesToCreate() = %v, want %v", got, tt.want)
			}
			if !slices.IsSorted(got) {
				t.Fatalf("SelectIndicesToCreate() = %v, want ascending order", got)
			}
		})
	}
}

func TestSelectIndicesToCreateSkipsNegativeIndices(t *testing.T) {
	tests := []struct {
		name           string
		currentIndices []int
		target         int
		want           []int
	}{
		{
			name:           "negative index does not occupy the replica count",
			currentIndices: []int{-1, 0, 1},
			target:         3,
			want:           []int{2},
		},
		{
			name:           "only negative indices treated as no instance",
			currentIndices: []int{-2, -1},
			target:         1,
			want:           []int{0},
		},
		{
			name:           "negative index is not a gap",
			currentIndices: []int{0, 1, 2, -1},
			target:         4,
			want:           []int{3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SelectIndicesToCreate(tt.currentIndices, tt.target)
			if !slices.Equal(got, tt.want) {
				t.Fatalf("SelectIndicesToCreate() = %v, want %v", got, tt.want)
			}
		})
	}
}
