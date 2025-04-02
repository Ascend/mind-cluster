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

// Package profiling contains functions that support dynamically collecting profiling data
package profiling

/*
#include <stdint.h>
#include <stddef.h>
*/
import "C"
import (
	"context"
	"sync"
	"time"

	"ascend-common/common-utils/hwlog"
	"taskd/common/constant"
)

// ProfTaskQueue is a global task queue
var ProfTaskQueue *TaskQueue

// BufferCompletedCallback is a callback function that is called when the buffer is completed
type BufferCompletedCallback func(buffer *C.uint8_t, size C.size_t, validSize C.size_t)

// Task a task structure that contains a callback function and parameters.
type Task struct {
	callback  BufferCompletedCallback
	buffer    *C.uint8_t
	size      C.size_t
	validSize C.size_t
}

// TaskQueue contains job need to be done
type TaskQueue struct {
	taskChan chan Task
	wg       sync.WaitGroup
	mu       sync.Mutex
	closed   bool
}

// NewTaskQueue initialize a new TaskQueue
func NewTaskQueue(ctx context.Context) *TaskQueue {
	tq := &TaskQueue{
		taskChan: make(chan Task, constant.TaskBufferSize),
	}
	// start to deal with workers, in a goroutine
	go tq.worker(ctx)
	return tq
}

// worker get a job to do
func (tq *TaskQueue) worker(ctx context.Context) {
	for {
		hwlog.RunLog.Debugf("rank:%d, current undo task num:%v", GlobalRankId, len(tq.taskChan))
		if len(tq.taskChan) >= constant.TaskBufferSize*constant.TaskThreadHold {
			hwlog.RunLog.Warnf("rank:%d, current got too many task", GlobalRankId)
		}
		select {
		case task, ok := <-tq.taskChan:
			if !ok {
				time.Sleep(time.Second)
				continue
			}
			task.callback(task.buffer, task.size, task.validSize)

			tq.wg.Done()
		case <-ctx.Done():
			hwlog.RunLog.Warn("worker queue will exit")
			return
		}
	}
}

// AddTask add task to a queue
func (tq *TaskQueue) AddTask(callback BufferCompletedCallback, buffer *C.uint8_t, size C.size_t, validSize C.size_t) {
	if tq == nil {
		return
	}
	tq.mu.Lock()
	defer tq.mu.Unlock()
	if tq.closed {
		return
	}
	tq.wg.Add(1)
	tq.taskChan <- Task{
		callback:  callback,
		buffer:    buffer,
		size:      size,
		validSize: validSize,
	}
}

// Wait wait for all done done
func (tq *TaskQueue) Wait() {
	tq.mu.Lock()
	tq.closed = true
	tq.mu.Unlock()
	tq.wg.Wait()
	close(tq.taskChan)
}
