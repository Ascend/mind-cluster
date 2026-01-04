/*
 * Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * 		http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package v1alpha1

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGpuConfigValidate(t *testing.T) {
	tests := map[string]struct {
		gpuConfig *NpuConfig
		expected  error
	}{
		"empty NpuConfig": {
			gpuConfig: &NpuConfig{},
			expected:  errors.New("no sharing strategy set"),
		},
		"empty NpuConfig.Sharing": {
			gpuConfig: &NpuConfig{
				Sharing: &NpuSharing{},
			},
			expected: errors.New("unknown GPU sharing strategy: "),
		},
		"unknown GPU sharing strategy": {
			gpuConfig: &NpuConfig{
				Sharing: &NpuSharing{
					Strategy: "unknown",
				},
			},
			expected: errors.New("unknown GPU sharing strategy: unknown"),
		},
		"empty NpuConfig.Sharing.TimeSlicingConfig": {
			gpuConfig: &NpuConfig{
				Sharing: &NpuSharing{
					Strategy:          TimeSlicingStrategy,
					TimeSlicingConfig: &TimeSlicingConfig{},
				},
			},
			expected: errors.New("unknown time-slice interval: "),
		},
		"valid NpuConfig with TimeSlicing": {
			gpuConfig: &NpuConfig{
				Sharing: &NpuSharing{
					Strategy: TimeSlicingStrategy,
					TimeSlicingConfig: &TimeSlicingConfig{
						Interval: MediumTimeSlice,
					},
				},
			},
			expected: nil,
		},
		"negative NpuConfig.Sharing.SpacePartitioningConfig.PartitionCount": {
			gpuConfig: &NpuConfig{
				Sharing: &NpuSharing{
					Strategy: SpacePartitioningStrategy,
					SpacePartitioningConfig: &SpacePartitioningConfig{
						PartitionCount: -1,
					},
				},
			},
			expected: errors.New("invalid partition count: -1"),
		},
		"valid NpuConfig with SpacePartitioning": {
			gpuConfig: &NpuConfig{
				Sharing: &NpuSharing{
					Strategy: SpacePartitioningStrategy,
					SpacePartitioningConfig: &SpacePartitioningConfig{
						PartitionCount: 1000,
					},
				},
			},
			expected: nil,
		},
		"default NpuConfig": {
			gpuConfig: DefaultGpuConfig(),
			expected:  nil,
		},
		"invalid TimeSlicingConfig ignored with strategy is SpacePartitioning": {
			gpuConfig: &NpuConfig{
				Sharing: &NpuSharing{
					Strategy:          SpacePartitioningStrategy,
					TimeSlicingConfig: &TimeSlicingConfig{},
					SpacePartitioningConfig: &SpacePartitioningConfig{
						PartitionCount: 1,
					},
				},
			},
			expected: nil,
		},
		"invalid SpacePartitioningConfig ignored with strategy is TimeSlicing": {
			gpuConfig: &NpuConfig{
				Sharing: &NpuSharing{
					Strategy: TimeSlicingStrategy,
					TimeSlicingConfig: &TimeSlicingConfig{
						Interval: MediumTimeSlice,
					},
					SpacePartitioningConfig: &SpacePartitioningConfig{
						PartitionCount: -1,
					},
				},
			},
			expected: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			err := test.gpuConfig.Validate()
			assert.Equal(t, test.expected, err)
		})
	}
}
