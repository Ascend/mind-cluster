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

// Package worker for taskd worker backend
package worker

// NewTaskDWorker return a taskd worker instance
func NewTaskDWorker(rankId int) *BaseWorker {
	return &BaseWorker{RankId: rankId}
}

// BaseWorker the class of taskd worker backend
type BaseWorker struct {
	RankId int
}

// Start start taskd worker
func (w *BaseWorker) Start() string {
	return "start worker!"
}

// Process process task
func (w *BaseWorker) Process() {
	return
}
