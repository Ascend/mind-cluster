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

// Package controllerflags is used for control the controller state
package controllerflags

import "sync"

type taskState struct {
	state   bool
	rwMutex sync.RWMutex
}

// GetState get isControllerStart state
func (s *taskState) GetState() bool {
	s.rwMutex.RLock()
	defer s.rwMutex.RUnlock()
	return s.state
}

// SetState Set isControllerStart state
// true: controller退出了
// false: controller没有退出
func (s *taskState) SetState(state bool) {
	s.rwMutex.Lock()
	defer s.rwMutex.Unlock()
	s.state = state
}

// IsControllerExited controllerExited state 初始值设置为true，表示controller is exited!
var IsControllerExited = taskState{state: true}

// IsControllerStarted controllerStarted state
var IsControllerStarted = taskState{state: false}
