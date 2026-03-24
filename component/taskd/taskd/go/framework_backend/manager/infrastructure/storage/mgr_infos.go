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

// Package storage for taskd manager backend data type
package storage

import (
	"sync"

	"taskd/common/utils"
)

// ManagerInfo for manager info
type MgrInfo struct {
	Status  map[string]string
	RWMutex sync.RWMutex
}

func (m *MgrInfo) updateMgr(newMgr *MgrInfo) error {
	m.RWMutex.Lock()
	defer m.RWMutex.Unlock()
	m.Status = newMgr.Status
	return nil
}

// SetStatusVal set manager status value
func (m *MgrInfo) SetStatusVal(key, value string) {
	m.RWMutex.Lock()
	defer m.RWMutex.Unlock()
	m.Status[key] = value
}

// GetStatusVal get manager status value
func (m *MgrInfo) GetStatusVal(key string) (string, bool) {
	m.RWMutex.RLock()
	defer m.RWMutex.RUnlock()
	val, ok := m.Status[key]
	return val, ok
}

// DeepCopy return a deep copy of MgrInfo
func (m *MgrInfo) DeepCopy() *MgrInfo {
	m.RWMutex.RLock()
	defer m.RWMutex.RUnlock()
	return &MgrInfo{
		Status:  utils.CopyStringMap(m.Status),
		RWMutex: sync.RWMutex{},
	}
}
