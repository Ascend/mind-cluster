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

// Package timedetector is a DT collection for func in history_abnormal_detection
package timedetector

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"ascend-common/common-utils/hwlog"
)

const (
	logLineLength = 256
	delta         = 1e-9
)

func init() {
	config := hwlog.LogConfig{
		OnlyToStdout:  true,
		MaxLineLength: logLineLength,
	}
	err := hwlog.InitRunLogger(&config, context.TODO())
	if err != nil {
		fmt.Println(err)
	}
}

func TestCalculateMean(t *testing.T) {
	var testCases = []struct {
		dataList []float64
		expect   float64
	}{
		{[]float64{}, 0.0},
		{[]float64{1, 2, 3, 4, 5, 6}, 3.5},
		{[]float64{1, -1, 2, -2, 3, -3}, 0},
		{[]float64{3, 3, 4}, 3.333333333333333},
		{[]float64{0.1, 0.2, 0.3}, 0.20000},
		{[]float64{42.00001}, 42.00001},
	}

	for _, tc := range testCases {
		assert.InDelta(t, tc.expect, calculateMean(tc.dataList), delta)
	}
}

func TestCalculateStandardDeviation(t *testing.T) {
	var testCases = []struct {
		data   []float64
		mean   float64
		expect float64
	}{
		{},
		{[]float64{0.1, 0.2, 0.3}, 0.2, 0.0816496580927726},
		{[]float64{0.1}, 0.1, 0},
		{[]float64{0.1, 0.2}, 0.1, 0.07071067811865477},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expect, calculateStandardDeviation(tc.data, tc.mean))
	}
}

func TestGetBound(t *testing.T) {
	var testCases = []struct {
		Nsigma                int
		degradationPercentage float64
		mean                  float64
		stdDev                float64
		expectLowerBound      float64
		expectUpperBound      float64
	}{
		{},
		{0, 0, 0.1, 0.777, -2.231, 2.431},
		{0, 0.7, 0.1, 0.777, 0.03, 0.17},
		{1, 0, 0.1, 0.777, -0.677, 0.877},
		{1, 0.7, 0.1, 0.777, 0.03, 0.17},
	}

	for _, tc := range testCases {
		assert.InDelta(t, tc.expectLowerBound,
			getLowerBound(tc.Nsigma, tc.degradationPercentage, tc.mean, tc.stdDev), delta)
		assert.InDelta(t, tc.expectUpperBound,
			getUpperBound(tc.Nsigma, tc.degradationPercentage, tc.mean, tc.stdDev), delta)
	}
}
