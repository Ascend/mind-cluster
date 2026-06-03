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

package server

import (
	"sync"
	"time"

	"ascend-common/common-utils/hwlog"
)

var invalidIdleTime time.Time

type IdleTimeMgr struct {
	idleTimes sync.Map
}

func NewIdleTimeMgr() *IdleTimeMgr {
	return &IdleTimeMgr{}
}

// RecordIdleTime records idle time for a device
func (m *IdleTimeMgr) RecordIdleTime(logicID int32) {
	now := time.Now()
	if _, loaded := m.idleTimes.LoadOrStore(logicID, now); loaded {
		hwlog.RunLog.Debugf("idle time already recorded, logicID: %d, skip", logicID)
		return
	}
	hwlog.RunLog.Infof("idle time recorded, logicID: %d, time: %s", logicID, now.Format(time.RFC3339))
}

// DeleteIdleTime deletes idle time for a device
func (m *IdleTimeMgr) DeleteIdleTime(logicID int32) {
	m.idleTimes.Delete(logicID)
	hwlog.RunLog.Debugf("idle time deleted, logicID: %d", logicID)
}

// IsIdleTimeExceeded checks if idle time for a device exceeds the threshold
func (m *IdleTimeMgr) IsIdleTimeExceeded(logicID int32, waitSeconds int) bool {
	val, ok := m.idleTimes.Load(logicID)
	if !ok {
		hwlog.RunLog.Debugf("idle time not recorded, logicID: %d", logicID)
		return false
	}
	idleTime := val.(time.Time)
	hwlog.RunLog.Infof("idle time checked, logicID: %d, idleTime: %s", logicID, idleTime.Format(time.RFC3339))
	return time.Since(idleTime) >= time.Duration(waitSeconds)*time.Second
}

// GetIdleTime returns idle time for a device
func (m *IdleTimeMgr) GetIdleTime(logicID int32) (time.Time, bool) {
	val, ok := m.idleTimes.Load(logicID)
	if !ok {
		return invalidIdleTime, false
	}
	return val.(time.Time), true
}
