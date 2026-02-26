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

// Package server holds the implementation of registration to kubelet, k8s device plugin interface and grpc service.
package server

import (
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	"Ascend-device-plugin/pkg/common"
	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
)

// setHcclTopoFilePathEnv set hccl topo file path env
func (ps *PluginServer) setHcclTopoFilePathEnv(resp *v1beta1.ContainerAllocateResponse, allNPUInfo common.NpuAllInfo) {
	if common.ParamOption.RealCardType != api.Ascend910A5 {
		hwlog.RunLog.Debugf("set hccl topo file path failed, card type not %s", api.HuaweiNPU)
		return
	}
	if len(allNPUInfo.AllDevs) == 0 {
		hwlog.RunLog.Error("set hccl topo file path failed, AllDevs len is 0")
		return
	}
	productTypeKey := ps.getProductTypeKey(common.ParamOption.CardType, allNPUInfo)
	if productTypeKey == -1 {
		hwlog.RunLog.Errorf("get product type key failed: type %v", common.ParamOption.CardType)
		return
	}
	if resp.Envs == nil {
		resp.Envs = make(map[string]string)
	}
	if path, exist := hcclTopoFilePathMap[productTypeKey]; exist {
		resp.Envs[common.HcclTopoFilePathKey] = path
		return
	}
	hwlog.RunLog.Errorf("get hccl topo file path failed, Product type<%s>", common.ParamOption.CardType)
}

func (ps *PluginServer) getProductTypeKey(cardType string, allNPUInfo common.NpuAllInfo) int8 {
	switch cardType {
	case common.A5300ICardName:
		return common.ProductType1PCard
	case common.A54P300ICardName:
		return common.ProductType4PCard
	default:
		superPodInfo, err := ps.manager.GetDmgr().GetSuperPodInfo(allNPUInfo.AllDevs[0].LogicID)
		if err != nil {
			hwlog.RunLog.Errorf("set hccl topo file path failed, get super pod info err<%v>", err)
			return -1
		}
		return int8(superPodInfo.SuperPodType)
	}
}
