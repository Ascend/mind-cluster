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

// Package common this for util method
package common

import (
	"testing"

	"ascend-common/api"
)

func TestGetMaxNpuCountPerNode(t *testing.T) {
	tests := []struct {
		name        string
		mainBoardId int
		expected    int
		description string
	}{
		{
			name:        "Override for A5UBX MainBoard",
			mainBoardId: A5UBXMainBoardId,
			expected:    16,
			description: "Should return overridden value 16 for A5UBX",
		},
		{
			name:        "Override for A5TX MainBoard",
			mainBoardId: A5TXMainBoardId,
			expected:    4,
			description: "Should return overridden value 4 for A5TX",
		},
		{
			name:        "Override for A5DY MainBoard",
			mainBoardId: A5DYMainBoardId,
			expected:    4,
			description: "Should return overridden value 4 for A5DY",
		},
		{
			name:        "Unknown MainBoard - return default",
			mainBoardId: 99999,
			expected:    api.NpuCountPerNode,
			description: "Should return default api.NpuCountPerNode when MainBoard ID not in override map",
		},
		{
			name:        "Zero MainBoard ID",
			mainBoardId: 0,
			expected:    api.NpuCountPerNode,
			description: "Should return default for zero/unknown MainBoard ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetMaxNpuCountPerNode(tt.mainBoardId)
			if result != tt.expected {
				t.Errorf("GetMaxNpuCountPerNode(%d) = %d, want %d (%s)",
					tt.mainBoardId, result, tt.expected, tt.description)
			}
		})
	}
}

func TestGetMaxNpuCountPerNodeConcurrency(t *testing.T) {
	const goroutines = 100
	done := make(chan bool)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			testIds := []int{A5UBXMainBoardId, A5TXMainBoardId, A5DYMainBoardId, 12345, 67890}
			for _, testId := range testIds {
				_ = GetMaxNpuCountPerNode(testId)
			}
			done <- true
		}(i)
	}

	for i := 0; i < goroutines; i++ {
		<-done
	}
}
