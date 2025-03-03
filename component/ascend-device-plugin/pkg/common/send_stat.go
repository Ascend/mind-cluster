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
	"fmt"
	"sync"
	"time"
)

type sendResult struct {
	sendTime time.Time
	success  bool
}

// SendStats is a structure used to collect and analyze send failure statistics.
type SendStats struct {
	rwLock      sync.RWMutex
	sendResults []sendResult
	recordLen   int
}

// NewSendStats creates a new instance of SendStats.
func NewSendStats(recordLength int) *SendStats {
	if recordLength <= 0 || recordLength > MaxSendRecordLength {
		recordLength = DefaultSendRecordLength
	}
	return &SendStats{
		sendResults: make([]sendResult, 0, recordLength+1),
		recordLen:   recordLength,
	}
}

// RecordSendResult records the result of a send operation.
func (s *SendStats) RecordSendResult(success bool) {
	s.rwLock.Lock()
	defer s.rwLock.Unlock()
	if success {
		s.sendResults = s.sendResults[:0]
		return
	}
	result := sendResult{
		sendTime: time.Now(),
		success:  success,
	}
	s.sendResults = append(s.sendResults, result)
	if len(s.sendResults) > s.recordLen {
		s.sendResults = s.sendResults[len(s.sendResults)-s.recordLen:]
	}
}

// GetConsecutiveFailures counts the number of consecutive send failures.
func (s *SendStats) GetConsecutiveFailures() int {
	s.rwLock.RLock()
	defer s.rwLock.RUnlock()

	count := 0
	for i := len(s.sendResults) - 1; i >= 0; i-- {
		if s.sendResults[i].success {
			break
		}
		count++
	}
	return count
}

// GetLastSendStatus return last send status
func (s *SendStats) GetLastSendStatus() bool {
	s.rwLock.RLock()
	defer s.rwLock.RUnlock()

	n := len(s.sendResults)
	if n == 0 {
		return true
	}
	return s.sendResults[n-1].success
}

// String format stat info to string
func (s *SendStats) String() string {
	return fmt.Sprintf("consecutiveFailureCount=%d, lastSendSuccess=%v",
		s.GetConsecutiveFailures(), s.GetLastSendStatus())
}
