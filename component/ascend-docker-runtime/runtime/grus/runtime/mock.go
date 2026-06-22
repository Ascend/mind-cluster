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

// Package runtime, mock runtime for ut test
package runtime

import "fmt"

const (
	// MockName mock runtime name
	MockName = "mock"
)

// MockRuntime is the mock implementation of RuntimeAPI for unit testing
type MockRuntime struct {
}

// NewRuntimeMock creates a new mock runtime instance
func NewRuntimeMock() RuntimeAPI {
	return &MockRuntime{}
}

// Init initializes the mock runtime
func (r *MockRuntime) Init(rootPath string) {
}

// Pause pauses the container with the specified ID
func (r *MockRuntime) Pause(id string) error {
	if id == "" {
		return fmt.Errorf("mock runtime Error")
	}
	return nil
}

// Resume resumes the paused container with the specified ID
func (r *MockRuntime) Resume(id string) error {
	if id == "" {
		return fmt.Errorf("mock runtime Error")
	}
	return nil
}

// State returns the state information of the container with the specified ID
func (r *MockRuntime) State(id string) (*StateInfo, error) {
	if id == "" {
		return nil, fmt.Errorf("mock runtime Error")
	}
	return &StateInfo{}, nil
}

// Checkpoint creates a checkpoint of the container at the specified path
func (r *MockRuntime) Checkpoint(ckptPath, id string) error {
	if ckptPath == "" {
		return fmt.Errorf("mock runtime Error")
	}
	return nil
}

// Restore restores the container from a checkpoint at the specified path
func (r *MockRuntime) Restore(ckptPath, id, ns string, externalEnvs []string) error {
	if ckptPath == "" {
		return fmt.Errorf("mock runtime Error")
	}
	return nil
}
