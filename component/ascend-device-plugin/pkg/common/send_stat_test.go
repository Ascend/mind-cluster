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

// Package common a series of common function
package common

import (
	"testing"
	"time"
)

const (
	numInt10 = 10
	numInt2  = 2
)

// TestNewSendStats tests the NewSendStats function.
func TestNewSendStats(t *testing.T) {
	// Test valid input
	stats := NewSendStats(numInt10)
	if stats.recordLen != numInt10 {
		t.Errorf("Expected record length %d, but got %d", numInt10, stats.recordLen)
	}

	// Test invalid record length (less than or equal to 0)
	stats = NewSendStats(0)
	if stats.recordLen != DefaultSendRecordLength {
		t.Errorf("Expected default record length %d, but got %d",
			DefaultSendRecordLength, stats.recordLen)
	}

	// Test invalid record length (greater than MaxSendRecordLength)
	stats = NewSendStats(MaxSendRecordLength + 1)
	if stats.recordLen != DefaultSendRecordLength {
		t.Errorf("Expected default record length %d, but got %d",
			DefaultSendRecordLength, stats.recordLen)
	}
}

// TestRecordSendResult tests the RecordSendResult function.
func TestRecordSendResult(t *testing.T) {
	stats := NewSendStats(numInt10)
	// Record a successful send result
	stats.RecordSendResult(true)
	if len(stats.sendResults) != 0 {
		t.Errorf("Expected sendResults length 0, but got %d", len(stats.sendResults))
	}

	for i := 0; i <= numInt10; i++ {
		stats.RecordSendResult(false)
	}
	if len(stats.sendResults) != numInt10 {
		t.Errorf("Expected sendResults length not to exceed record length %d, but got %d",
			numInt10, len(stats.sendResults))
	}
}

// TestGetConsecutiveFailures tests the GetConsecutiveFailures function.
func TestGetConsecutiveFailures(t *testing.T) {
	stats := NewSendStats(numInt10)
	// Record some results
	for i := 0; i < numInt10; i++ {
		stats.RecordSendResult(true)
	}
	for i := 0; i < numInt10; i++ {
		stats.RecordSendResult(false)
	}
	count := stats.GetConsecutiveFailures()
	if count != numInt10 {
		t.Errorf("Expected consecutive failures count %d, but got %d", numInt10, count)
	}

	// Test the case with no failures
	successStats := &SendStats{
		sendResults: []sendResult{
			{
				sendTime: time.Now(),
				success:  true,
			},
		},
		recordLen: numInt10}
	count = successStats.GetConsecutiveFailures()
	if count != 0 {
		t.Errorf("Expected consecutive failures count 0 when no failures, but got %d", count)
	}
}

// TestGetLastSendStatus tests the GetLastSendStatus function.
func TestGetLastSendStatus(t *testing.T) {
	stats := NewSendStats(numInt10)
	// Record a successful send result
	stats.RecordSendResult(true)
	status := stats.GetLastSendStatus()
	if !status {
		t.Errorf("Expected last send status to be true, but got false")
	}

	// Record a failed send result
	stats.RecordSendResult(false)
	status = stats.GetLastSendStatus()
	if status {
		t.Errorf("Expected last send status to be false, but got true")
	}

	// Test the case with no results
	emptyStats := NewSendStats(numInt10)
	status = emptyStats.GetLastSendStatus()
	if !status {
		t.Errorf("Expected last send status to be true when no results, but got false")
	}
}

// TestString tests the String function.
func TestString(t *testing.T) {
	stats := NewSendStats(numInt10)
	// Record some results
	for i := 0; i < numInt10; i++ {
		stats.RecordSendResult(true)
	}
	str := stats.String()
	if str == "" {
		t.Errorf("Expected non - empty string representation, but got empty string")
	}
}
