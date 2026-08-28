/*
 * Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at

 * http://www.apache.org/licenses/LICENSE-2.0

 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package device

import (
	"context"
	"errors"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-common/devmanager"
)

// initLog initialises the hwlog run logger so generation code that calls
// hwlog.RunLog does not panic on a nil logger during tests.
func initLog() {
	hwLogConfig := hwlog.LogConfig{OnlyToStdout: true}
	hwlog.InitRunLogger(&hwLogConfig, context.Background())
}

func init() { initLog() }

// errSentinel is a stable sentinel error reused by gomonkey stubs.
var errSentinel = errors.New("test sentinel error")

// new910WithMock builds an Ascend910Generation backed by a DeviceManagerMock
// whose DevType is api.Ascend910 so GetDevType returns a meaningful value.
func new910WithMock() (*Ascend910Generation, *devmanager.DeviceManagerMock) {
	mock := &devmanager.DeviceManagerMock{DevType: api.Ascend910}
	g := NewAscend910Generation()
	g.SetDmgr(mock)
	return g, mock
}

// new950WithMock builds an Ascend950Generation backed by a DeviceManagerMock.
func new950WithMock() (*Ascend950Generation, *devmanager.DeviceManagerMock) {
	mock := &devmanager.DeviceManagerMock{DevType: api.Ascend910A5}
	g := NewAscend950Generation()
	g.SetDmgr(mock)
	return g, mock
}
