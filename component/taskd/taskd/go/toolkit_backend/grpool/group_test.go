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
// package grpool is a Go package that provides a simple and efficient way to manage goroutines.
package grpool

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestNewGroup(t *testing.T) {
	gr := NewPool(1, context.Background())
	g := newGroup(gr)
	if g == nil {
		t.Error("Expected new group to be created, but got nil")
	}
}

func TestGroupSubmit(t *testing.T) {
	gr := NewPool(1, context.Background())
	g := newGroup(gr)

	done := make(chan struct{})
	fn := func(t Task) (interface{}, error) {
		close(done)
		return nil, nil
	}

	g.Submit(fn)
	<-done

	if len(g.Results()) != 1 {
		t.Errorf("Expected 1 result, but got %d", len(g.Results()))
	}
}

func TestGroupResults(t *testing.T) {
	gr := NewPool(1, context.Background())
	g := newGroup(gr)

	fn := func(t Task) (interface{}, error) {
		return nil, nil
	}

	g.Submit(fn)
	results := g.Results()
	if len(results) != 1 {
		t.Errorf("Expected 1 result, but got %d", len(results))
	}
}

func TestGroupWaitGroup(t *testing.T) {
	gr := NewPool(1, context.Background())
	g := newGroup(gr)

	var wg sync.WaitGroup
	wg.Add(1)
	fn := func(t Task) (interface{}, error) {
		wg.Done()
		return nil, nil
	}

	g.Submit(fn)
	wg.Wait()

	done := make(chan struct{})
	go func() {
		g.WaitGroup()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Error("WaitGroup took too long to complete")
	}
}
