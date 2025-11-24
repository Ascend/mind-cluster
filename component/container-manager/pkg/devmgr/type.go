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

// Package devmgr hwDevMgr struct
package devmgr

import (
	"ascend-common/devmanager"
	"container-manager/pkg/common"
)

// HwDevMgr manages huawei npu device
type HwDevMgr struct {
	// node dimension
	devType  string
	boardId  uint32
	devUsage string
	workMode string
	// device dimension
	npuInfos map[int32]*common.NPUInfo // key: phy id; value: npu info
	// interacting with dcmi interface
	dmgr devmanager.DeviceInterface
}
