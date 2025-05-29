/*
Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.

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
	"sync"
)

const (
	defaultTaskQueue = 8
)

type group struct {
	grPool  GrPool
	lock    sync.Mutex
	results []Task
	wg      *sync.WaitGroup
}

// newGroup creates a new task group associated with the given pool
func newGroup(gr GrPool) Group {
	return &group{
		grPool:  gr,
		results: make([]Task, 0, defaultTaskQueue),
		wg:      new(sync.WaitGroup),
	}
}

// Submit implements Group interface, adds a task to the group
func (g *group) Submit(fn TaskFunc) {
	t := g.grPool.Submit(fn)
	g.wg.Add(1)
	g.lock.Lock()
	g.results = append(g.results, t)
	g.lock.Unlock()

	go func(g *group, t Task) {
		defer g.wg.Done()
		t.Wait()
	}(g, t)
}

// Results implements Group interface, returns all tasks in the group
func (g *group) Results() []Task {
	g.lock.Lock()
	defer g.lock.Unlock()
	return g.results
}

// WaitGroup implements Group interface, waits for all tasks to complete
func (g *group) WaitGroup() {
	g.wg.Wait()
}
