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

// Package snapshot for the control of concurrent request of checkpoint
package snapshot

import (
	"fmt"
	"sync"
)

type semaphore struct {
	mu      sync.Mutex
	idStore map[string]struct{}
	max     uint
}

func newSemaphore(max uint) *semaphore {
	if max == 0 {
		return nil
	}

	s := &semaphore{
		idStore: make(map[string]struct{}, max),
		max:     max,
	}
	return s
}

func (s *semaphore) acquire(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.idStore) >= int(s.max) {
		return fmt.Errorf("reach max, wait release and retry")
	}
	if _, ok := s.idStore[id]; ok {
		return fmt.Errorf("id: %s, has acquired", id)
	}
	s.idStore[id] = struct{}{}
	return nil
}

func (s *semaphore) release(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.idStore, id)
}
