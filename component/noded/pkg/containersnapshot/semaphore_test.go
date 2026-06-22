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

// Package snapshot for ut of semaphore
package snapshot

import (
	"fmt"
	"sync"
	"testing"

	"nodeD/pkg/common"
)

func TestNewSemaphore(t *testing.T) {
	t.Run("MaxIsZero", func(t *testing.T) {
		s := newSemaphore(0)
		if s != nil {
			t.Errorf("Expected nil, got %v", s)
		}
	})

	t.Run("MaxIsPositive", func(t *testing.T) {
		s := newSemaphore(common.MaxCheckpointRequest)
		if s == nil {
			t.Errorf("Expected non-nil, got %v", s)
		}
		if len(s.idStore) != 0 {
			t.Errorf("Expected empty idStore, got %v", s.idStore)
		}
		if s.max != common.MaxCheckpointRequest {
			t.Errorf("Expected max %d, got %d", common.MaxCheckpointRequest, s.max)
		}
	})
}

func TestAcquire(t *testing.T) {
	t.Run("AcquireSuccess", func(t *testing.T) {
		s := newSemaphore(common.MaxCheckpointRequest)
		err := s.acquire("id1")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if _, ok := s.idStore["id1"]; !ok {
			t.Errorf("Expected id1 in idStore, got %v", s.idStore)
		}
	})

	t.Run("AcquireWhenMaxReached", func(t *testing.T) {
		s := newSemaphore(1)
		s.acquire("id1")
		err := s.acquire("id2")
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
		errMsg := "reach max, wait release and retry"
		if err.Error() != errMsg {
			t.Errorf("errmsg changed: %s != %v", errMsg, err)
		}
		if _, ok := s.idStore["id2"]; ok {
			t.Errorf("Expected id2 not in idStore, got %v", s.idStore)
		}
	})

	t.Run("AcquireSameID", func(t *testing.T) {
		s := newSemaphore(common.MaxCheckpointRequest)
		s.acquire("id1")
		err := s.acquire("id1")
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
		errMsg := "id: id1, has acquired"
		if err.Error() != errMsg {
			t.Errorf("errmsg changed: %s != %v", errMsg, err)
		}
		if len(s.idStore) != 1 {
			t.Errorf("Expected 1 entry in idStore, got %v", s.idStore)
		}
	})
}

func TestRelease(t *testing.T) {
	t.Run("ReleaseSuccess", func(t *testing.T) {
		s := newSemaphore(common.MaxCheckpointRequest)
		s.acquire("id1")
		s.release("id1")
		if _, ok := s.idStore["id1"]; ok {
			t.Errorf("Expected id1 not in idStore, got %v", s.idStore)
		}
	})

	t.Run("ReleaseNonExistID", func(t *testing.T) {
		s := newSemaphore(common.MaxCheckpointRequest)
		s.release("id1")
		if _, ok := s.idStore["id1"]; ok {
			t.Errorf("Expected id1 not in idStore, got %v", s.idStore)
		}
	})
}

func TestConcurrentAcquire(t *testing.T) {
	t.Run("ConcurrentAcquireSuccess", func(t *testing.T) {
		s := newSemaphore(common.MaxCheckpointRequest)
		var wg sync.WaitGroup
		var mu sync.Mutex
		var errors []error

		for i := 0; i < common.MaxCheckpointRequest; i++ {
			wg.Add(1)
			go func(id string) {
				defer wg.Done()
				err := s.acquire(id)
				mu.Lock()
				if err != nil {
					errors = append(errors, err)
				}
				mu.Unlock()
			}(fmt.Sprintf("id%d", i))
		}

		wg.Wait()

		if len(errors) > 0 {
			t.Errorf("Expected no errors, got %v", errors)
		}
		if len(s.idStore) != common.MaxCheckpointRequest {
			t.Errorf("Expected 8 entries in idStore, got %v", s.idStore)
		}
	})

	t.Run("ConcurrentAcquireWhenMaxReached", func(t *testing.T) {
		const TestMaxCnt = 3
		s := newSemaphore(TestMaxCnt)
		var wg sync.WaitGroup
		var mu sync.Mutex
		var errors []error

		for i := 0; i < common.MaxCheckpointRequest; i++ {
			wg.Add(1)
			go func(id string) {
				defer wg.Done()
				err := s.acquire(id)
				mu.Lock()
				if err != nil {
					errors = append(errors, err)
				}
				mu.Unlock()
			}(fmt.Sprintf("id%d", i))
		}

		wg.Wait()

		expectedErrors := common.MaxCheckpointRequest - TestMaxCnt
		if len(errors) != expectedErrors {
			t.Errorf("Expected %d errors, got %v", expectedErrors, errors)
		}
		if len(s.idStore) != TestMaxCnt {
			t.Errorf("Expected 3 entries in idStore, got %v", s.idStore)
		}
	})

	t.Run("ConcurrentAcquireSameID", func(t *testing.T) {
		s := newSemaphore(common.MaxCheckpointRequest)
		var wg sync.WaitGroup
		var mu sync.Mutex
		var errors []error

		for i := 0; i < common.MaxCheckpointRequest; i++ {
			wg.Add(1)
			go func(id string) {
				defer wg.Done()
				err := s.acquire(id)
				mu.Lock()
				if err != nil {
					errors = append(errors, err)
				}
				mu.Unlock()
			}("id1")
		}

		wg.Wait()

		if len(errors) != common.MaxCheckpointRequest-1 {
			t.Errorf("Expected 7 errors, got %v", errors)
		}
		if len(s.idStore) != 1 {
			t.Errorf("Expected 1 entry in idStore, got %v", s.idStore)
		}
	})
}
