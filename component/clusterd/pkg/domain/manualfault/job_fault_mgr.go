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

// Package manualfault cache for hardware frequency fault with job
package manualfault

import "sync"

// JobFaultMgr an instance of JobFaultManager
var JobFaultMgr *JobFaultManager

// JobFaultManager is the job fault manager
type JobFaultManager struct {
	jobFault      map[string]*faultInfo
	slidingWindow int64 // unit: millisecond
	mutex         sync.RWMutex
}

type faultInfo struct {
	faults []*Fault
}

// Fault is the hardware frequency fault detail
type Fault struct {
	Code        string
	JobId       string
	NodeName    string
	DevName     string
	ReceiveTime int64 // unit: millisecond
}
