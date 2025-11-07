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

// Package server holds the implementation of registration to kubelet, k8s pod resource interface.
package server

import (
	"Ascend-device-plugin/pkg/common"
	npuCommon "ascend-common/devmanager/common"
)

func (hdm *HwDevManager) getCardType() (string, error) {
	boardInfo, err := hdm.manager.GetDmgr().GetBoardInfo(hdm.allInfo.AllDevs[common.FirstDevice].LogicID)
	if err != nil {
		return "", err
	}

	if boardInfo.BoardId != npuCommon.A5300IBoardId && boardInfo.BoardId != npuCommon.A5300IBoardId2 {
		return "", nil
	}

	mainBoardId := hdm.manager.GetDmgr().GetMainBoardId()

	if mainBoardId == common.A5300IMainBoardId {
		return common.A5300ICardName, nil
	}
	if mainBoardId == common.A5300I4PMainBoardId {
		return common.A54P300ICardName, nil
	}

	return "", nil
}
