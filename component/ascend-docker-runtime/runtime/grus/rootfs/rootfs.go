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

// Package rootfs, interface of rootfs package
package rootfs

import (
	"ascend-common/common-utils/hwlog"
	"ascend-docker-runtime/runtime/common"
)

type RootfsSnapshot interface {
	Checkpoint(ckptPath, containerID, ns string) (string, error)
	Restore(ckptPath, rootfsPath, decompressType string) error
}

var snapshot map[string]func(string) RootfsSnapshot

func init() {
	snapshot = make(map[string]func(string) RootfsSnapshot)
	snapshot[common.CONTAINERD_SNAPSHOT_KEY] = NewContainerdRootfs
	snapshot[MOCK_SNAPSHOT_KEY] = NewMockRootfs
}

func GetRootfsSnapshot(engine, engineSocket string) RootfsSnapshot {
	sn, ok := snapshot[engine]
	if ok {
		return sn(engineSocket)
	}
	hwlog.RunLog.Infof("invalid engine config: %s, ignore error and use default containerd rootfs", engine)
	return NewContainerdRootfs(common.ContainerdSock)
}
