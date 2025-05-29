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
	"context"
)

const (
	defaultWorkers = 8
	bufferMultiple = 2
)

var _ GrPool = new(grPool)

type grPool struct {
	workers uint32
	tasks   chan Task
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewPool creates a new goroutine pool with specified number of workers
// workers: number of goroutines in the pool (0 means use default)
// ctx: context for controlling pool lifecycle
func NewPool(workers uint32, ctx context.Context) GrPool {
	if workers == 0 {
		workers = defaultWorkers
	}
	g := &grPool{
		workers: workers,
	}
	g.ctx, g.cancel = context.WithCancel(ctx)
	g.initPool()
	return g
}

func (gr *grPool) initPool() {
	gr.tasks = make(chan Task, gr.workers*bufferMultiple)
	for i := 0; i < int(gr.workers); i++ {
		gr.startWorker()
	}
}

func (gr *grPool) startWorker() {
	go func(g *grPool) {
		var t Task
		for {
			select {
			case t = <-g.tasks:
				t.Execute()
			case <-gr.ctx.Done():
				return
			}
		}
	}(gr)
}

// Submit implements GrPool interface, adds a task to the pool
func (gr *grPool) Submit(fc TaskFunc) Task {
	t := &task{
		closeChan: make(chan struct{}),
		fc:        fc,
	}
	gr.tasks <- t
	return t
}

// Group implements GrPool interface, creates a new task group
func (gr *grPool) Group() Group {
	return newGroup(gr)
}

// Close implements GrPool interface, shuts down the pool
func (gr *grPool) Close() {
	if gr.cancel != nil {
		gr.cancel()
	}
}
