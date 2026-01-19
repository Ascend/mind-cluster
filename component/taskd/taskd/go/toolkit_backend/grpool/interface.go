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

// Task represents an asynchronous task that can be executed and waited for
type Task interface {
	// Wait blocks until the task is completed
	Wait()
	// Result returns the task execution result and error
	Result() (interface{}, error)
	// Execute runs the task function
	Execute()
}

// Group represents a collection of tasks that can be managed together
type Group interface {
	// Submit adds a new task to the group
	Submit(fn TaskFunc)
	// Results returns all tasks in the group
	Results() []Task
	// WaitGroup blocks until all tasks in the group are completed
	WaitGroup()
}

// GrPool represents a goroutine pool that can execute tasks
type GrPool interface {
	// Submit adds a new task to the pool and returns the task
	Submit(fn TaskFunc) Task
	// Group creates a new task group
	Group() Group
	// Close shuts down the goroutine pool
	Close()
}

// TaskFunc defines the function signature for task execution
type TaskFunc func(t Task) (interface{}, error)
