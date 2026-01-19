/* Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
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

/*
Package utils is a DT collection for func in pp_network_detection
*/
package utils

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPPNetworkDetection(t *testing.T) {
	/*
		this is an integration test for func bellow:
		1. PPNetworkDetection
		2. updateBadConnectNumbers
		3. processAbnormalCard
		4. updateBadConnectNumbersForIndex

		some func used but not well tested are bellow:
		1. initializeConnectNumbers
		2. TestProcessAbnormalCard
		3. calculatePPValue
		4. findAbnormalNodes
	*/

	var testCases = []struct {
		slowSendRanks []int
		PPrankss      [][]int
		expect        []int
	}{
		{},
		{[]int{1, 2, 3}, [][]int{{1, 2}, {3, 4}, {4, 5}}, []int{}},
		{[]int{}, [][]int{{1, 2, 3, 4, 5, 6}}, []int{}},
		{[]int{1, 2, 3}, [][]int{{1, 2, 3, 4, 5, 6}}, []int{3}},
		{[]int{1}, [][]int{{1, 2, 3, 4, 5, 6}}, []int{1}},
	}
	for _, tc := range testCases {
		assert.ElementsMatch(t, tc.expect, PPNetworkDetection(tc.slowSendRanks, tc.PPrankss))
	}
}

func TestInitializeConnectNumbers(t *testing.T) {
	var testCases = []struct {
		length            int
		connectNumbers    []int
		badConnectNumbers []int
		err               error
	}{
		{0, nil, nil, errors.New("PPranks len is zero")},
		{1, []int{1}, []int{0}, nil},
		{2, []int{1, 1}, []int{0, 0}, nil},
		{3, []int{1, 2, 1}, []int{0, 0, 0}, nil},
	}

	for _, tc := range testCases {
		connectNumbers, badConnectNumbers, err := initializeConnectNumbers(tc.length)
		assert.Equal(t, tc.connectNumbers, connectNumbers)
		assert.Equal(t, tc.badConnectNumbers, badConnectNumbers)
		assert.Equal(t, tc.err, err)
	}
}

func TestProcessAbnormalCard(t *testing.T) {
	var testCases = []struct {
		abnormalCard int
		// the length of PPranks and badConnectNumbers are the same
		PPranks []int
		// fixed value, see: initializeConnectNumbers
		badConnectNumbers []int
		expect            []int
	}{
		{1, []int{1, 2, 3, 4}, []int{1, 2, 2, 1}, []int{2, 3, 2, 1}},
		{4, []int{1, 2, 3, 4}, []int{1, 2, 2, 1}, []int{1, 2, 2, 1}},
		{2, []int{1, 2, 3, 4}, []int{1, 2, 2, 1}, []int{1, 3, 3, 1}},
	}

	for _, tc := range testCases {
		processAbnormalCard(tc.abnormalCard, tc.PPranks, tc.badConnectNumbers)
		assert.Equal(t, tc.expect, tc.badConnectNumbers)
	}
}

func TestCalculatePPValue(t *testing.T) {
	var testCases = []struct {
		connectNumbers    []int
		badConnectNumbers []int
		expect            []float64
	}{
		{[]int{}, []int{}, []float64{}},
		{[]int{1, 2, 3, 4}, []int{}, []float64{0, 0, 0, 0}},
		{[]int{}, []int{1, 2, 3, 4}, []float64{}},
		{[]int{1, 2, 3, 4}, []int{1}, []float64{1, 0, 0, 0}},
		{[]int{1, 2, 3, 4}, []int{1, 2, 3, 4, 5, 6}, []float64{1, 1, 1, 1}},
		{[]int{1, 2, 3, 4}, []int{0, 0, 0, 0, 0, 0}, []float64{0, 0, 0, 0}},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expect, calculatePPValue(tc.connectNumbers, tc.badConnectNumbers))
	}
}

func TestFindAbnormalNodes(t *testing.T) {
	// the length of PPranks and ppValue are the same
	var testCases = []struct {
		PPranks []int
		ppValue []float64
		expect  []int
	}{
		{[]int{}, []float64{}, []int{}},
		{[]int{1, 2, 3, 4}, []float64{0, 0.5, 0, 0}, []int{}},
		{[]int{1, 1, 1, 1}, []float64{0.5, 0.5, 0.5, 0.5}, []int{}},
		{[]int{2, 2, 2, 2}, []float64{0.5, 1, 1, 1}, []int{2}},
		{[]int{2, 2, 2, 2}, []float64{1, 0.5, 1, 1}, []int{2, 2}},
		{[]int{2, 2, 2, 2}, []float64{1, 0.5, 1, 0.5}, []int{2, 2}},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expect, findAbnormalNodes(tc.PPranks, tc.ppValue))
	}
}
