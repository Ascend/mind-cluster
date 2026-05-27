/*
Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.

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

package common

import (
	"sync"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestPodUIDSetChanged(t *testing.T) {
	tests := []struct {
		name string
		m    *sync.Map
		pods []*corev1.Pod
		want bool
	}{
		{
			name: "01-empty map and empty pods",
			m:    &sync.Map{},
			pods: nil,
			want: false,
		},
		{
			name: "02-empty map with pods",
			m:    &sync.Map{},
			pods: []*corev1.Pod{{ObjectMeta: metav1.ObjectMeta{UID: "uid-1"}}},
			want: true,
		},
		{
			name: "03-non-empty map with empty pods",
			m: func() *sync.Map {
				m := &sync.Map{}
				m.Store(types.UID("uid-1"), struct{}{})
				return m
			}(),
			pods: nil,
			want: true,
		},
		{
			name: "04-same uids",
			m: func() *sync.Map {
				m := &sync.Map{}
				m.Store(types.UID("uid-1"), struct{}{})
				return m
			}(),
			pods: []*corev1.Pod{{ObjectMeta: metav1.ObjectMeta{UID: "uid-1"}}},
			want: false,
		},
		{
			name: "05-pod replaced (same count, different uid)",
			m: func() *sync.Map {
				m := &sync.Map{}
				m.Store(types.UID("uid-1"), struct{}{})
				return m
			}(),
			pods: []*corev1.Pod{{ObjectMeta: metav1.ObjectMeta{UID: "uid-2"}}},
			want: true,
		},
		{
			name: "06-new pod added (more pods than map)",
			m: func() *sync.Map {
				m := &sync.Map{}
				m.Store(types.UID("uid-1"), struct{}{})
				return m
			}(),
			pods: []*corev1.Pod{
				{ObjectMeta: metav1.ObjectMeta{UID: "uid-1"}},
				{ObjectMeta: metav1.ObjectMeta{UID: "uid-2"}},
			},
			want: true,
		},
		{
			name: "07-pod removed (fewer pods than map)",
			m: func() *sync.Map {
				m := &sync.Map{}
				m.Store(types.UID("uid-1"), struct{}{})
				m.Store(types.UID("uid-2"), struct{}{})
				return m
			}(),
			pods: []*corev1.Pod{{ObjectMeta: metav1.ObjectMeta{UID: "uid-1"}}},
			want: true,
		},
		{
			name: "08-multiple same uids",
			m: func() *sync.Map {
				m := &sync.Map{}
				m.Store(types.UID("uid-1"), struct{}{})
				m.Store(types.UID("uid-2"), struct{}{})
				return m
			}(),
			pods: []*corev1.Pod{
				{ObjectMeta: metav1.ObjectMeta{UID: "uid-1"}},
				{ObjectMeta: metav1.ObjectMeta{UID: "uid-2"}},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PodUIDSetChanged(tt.m, tt.pods)
			if got != tt.want {
				t.Errorf("PodUIDSetChanged() = %v, want %v", got, tt.want)
			}
		})
	}
}
