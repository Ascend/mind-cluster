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
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"Ascend-device-plugin/pkg/common"
	"Ascend-device-plugin/pkg/enpuconfig"
	"ascend-common/common-utils/hwlog"
	"ascend-common/enpumanager"
)

// managerRequestTimeout bounds manager operations in one allocation or reconciliation.
const managerRequestTimeout = 10 * time.Second

// managerConnectionExitCode requests a plugin restart after a manager connection failure.
const managerConnectionExitCode = 1

// managerDialTimeout bounds a resource API socket connection.
const managerDialTimeout = time.Second

// softShareManagerClient provides allocation creation, lookup and release.
type softShareManagerClient interface {
	enpuconfig.ResourceClient
	Release(context.Context, string) error
}

// softShareManager shares mode selection and serialization across plugin servers.
type softShareManager struct {
	client         softShareManagerClient
	socketPath     string
	seen           atomic.Bool
	localWarned    atomic.Bool
	allocationLock sync.Mutex
}

// newSoftShareManager initializes the shared resource client and mode state.
func newSoftShareManager() *softShareManager {
	return &softShareManager{
		client:     enpumanager.NewClient(enpumanager.SocketPath, managerDialTimeout),
		socketPath: enpumanager.SocketPath,
	}
}

// enabled remembers manager mode once the socket has been observed by this process.
// Installing or uninstalling enpu-manager requires a device plugin restart, so one observation
// is enough to decide the mode for this process; the sticky flag also keeps manager mode during
// an enpu-manager service restart, which is allowed while soft share tasks are running.
func (m *softShareManager) enabled() (bool, error) {
	if m.seen.Load() {
		return true, nil
	}
	if _, err := os.Stat(m.socketPath); err != nil {
		if os.IsNotExist(err) {
			return m.seen.Load(), nil
		}
		return false, fmt.Errorf("check enpu-manager socket: %w", err)
	}
	// Keep using manager after the socket has been observed, including during a restart.
	m.seen.Store(true)
	return true, nil
}

// requireManager reports whether a soft share allocation can use enpu-manager.
// Once the socket exists the manager is installed for this process, so the allocation must use
// it and a later UDS failure fails the allocation; there is no fallback to the local
// configuration. A missing socket means enpu-manager is not installed on this node, which keeps
// the legacy local configuration and only warns once.
func (m *softShareManager) requireManager() (bool, error) {
	ready, err := m.enabled()
	if err != nil || ready {
		return ready, err
	}
	if !m.localWarned.Swap(true) {
		hwlog.RunLog.Warnf("enpu-manager socket %s is absent, soft share uses local configuration",
			m.socketPath)
	}
	return false, nil
}

// managerOperationError requests a restart on manager connection failures during allocation.
func managerOperationError(err error) error {
	var connectionErr *enpumanager.ConnectionError
	if errors.As(err, &connectionErr) && common.IsSupportSoftShareDevice() {
		hwlog.RunLog.Errorf("enpu-manager UDS connection failed: %v", err)
		os.Exit(managerConnectionExitCode)
	}
	return err
}
