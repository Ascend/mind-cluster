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

import "ascend-common/common-utils/hwlog"

var _ Task = new(task)

type task struct {
	returnValue interface{}
	returnError error
	closeChan   chan struct{}
	fc          TaskFunc
}

// Wait implements Task interface, blocks until task is completed
func (t *task) Wait() {
	<-t.closeChan
}

// Result implements Task interface, returns task execution result
func (t *task) Result() (interface{}, error) {
	return t.returnValue, t.returnError
}

// Execute implements Task interface, runs the task function
func (t *task) Execute() {
	if t.fc == nil {
		hwlog.RunLog.Errorf("TaskFunc is  nil")
		return
	}
	t.returnValue, t.returnError = t.fc(t)
	close(t.closeChan)
}
