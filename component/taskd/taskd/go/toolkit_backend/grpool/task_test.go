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
	"errors"
	"testing"
	"time"
)

func TestTaskWait(t *testing.T) {
	task := &task{
		closeChan: make(chan struct{}),
	}

	done := make(chan struct{})
	go func() {
		task.Wait()
		close(done)
	}()

	close(task.closeChan)
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Error("Wait took too long to complete")
	}
}

func TestTaskResult(t *testing.T) {
	expectedValue := "test"
	expectedError := errors.New("test error")
	task := &task{
		returnValue: expectedValue,
		returnError: expectedError,
	}

	value, err := task.Result()
	if value != expectedValue {
		t.Errorf("Expected value %v, but got %v", expectedValue, value)
	}
	if err != expectedError {
		t.Errorf("Expected error %v, but got %v", expectedError, err)
	}
}

func TestTaskExecute(t *testing.T) {
	expectedValue := "test"
	expectedError := errors.New("test error")
	fn := func(t Task) (interface{}, error) {
		return expectedValue, expectedError
	}
	task := &task{
		fc:        fn,
		closeChan: make(chan struct{}),
	}

	task.Execute()
	value, err := task.Result()
	if value != expectedValue {
		t.Errorf("Expected value %v, but got %v", expectedValue, value)
	}
	if err != expectedError {
		t.Errorf("Expected error %v, but got %v", expectedError, err)
	}
}
