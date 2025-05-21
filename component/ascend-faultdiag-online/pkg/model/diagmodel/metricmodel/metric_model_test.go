/*
Copyright(C)2025. Huawei Technologies Co.,Ltd. All rights reserved.

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

// Package metricmodel provides the test case for the metricmodel package.
package metricmodel

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"ascend-faultdiag-online/pkg/model/enum"
	"ascend-faultdiag-online/pkg/utils/constants"
)

func TestGetDomainItemKey(t *testing.T) {
	testCases := []struct {
		name     string
		item     DomainItem
		expected string
	}{
		{
			name: "ValidNPUDomain",
			item: DomainItem{
				DomainType: enum.NpuDomain,
				Value:      "usage",
			},
			expected: "npu" + constants.ValueSeparator + "usage",
		},
		{
			name: "ValidHostDomain",
			item: DomainItem{
				DomainType: enum.HostDomain,
				Value:      "usage",
			},
			expected: "host" + constants.ValueSeparator + "usage",
		},
		{
			name: "ValidNetworkDomain",
			item: DomainItem{
				DomainType: enum.NetworkDomain,
				Value:      "usage",
			},
			expected: "network" + constants.ValueSeparator + "usage",
		},
		{
			name: "ValidNpuChipDomain",
			item: DomainItem{
				DomainType: enum.NpuChipDomain,
				Value:      "usage",
			},
			expected: "npu_chip" + constants.ValueSeparator + "usage",
		},
		{
			name: "EmptyValue",
			item: DomainItem{
				DomainType: enum.HostDomain,
				Value:      "",
			},
			expected: "host" + constants.ValueSeparator,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.item.GetDomainItemKey()
			assert.Equal(t, tc.expected, result)
		})
	}
}
