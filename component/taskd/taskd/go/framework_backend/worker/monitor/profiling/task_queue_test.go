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

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestNewTaskQueue(t *testing.T) {
	convey.Convey("new task queue should not nil", t, func() {
		ctx, cancel := context.WithCancel(context.Background())
		queue := NewTaskQueue(ctx)
		cancel()
		convey.ShouldNotBeNil(queue)
		convey.ShouldNotBeNil(queue.taskChan)
	})
}

func TestTaskQueueWait(t *testing.T) {
	convey.Convey("when wait task then chan should nil", t, func() {
		ctx, cancel := context.WithCancel(context.Background())
		queue := NewTaskQueue(ctx)
		queue.Wait()
		cancel()
		convey.ShouldBeNil(queue.taskChan)
	})
}

func TestAddTask(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	queue := NewTaskQueue(ctx)
	queue.AddTask(nil, nil, 0, 0)
	cancel()
	convey.ShouldBeZeroValue(len(queue.taskChan))
}
