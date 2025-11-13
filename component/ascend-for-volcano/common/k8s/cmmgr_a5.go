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

// Package k8s is using for the k8s operation like configmap informer.
package k8s

import (
	"strings"
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
)

func (cmMgr *ClusterInfoWitchCm) dealClusterDpuInfo(cm *v1.ConfigMap, operator string) {
	if !strings.HasPrefix(cm.Name, util.DpuCmInfoNamePrefixByClusterd) {
		return
	}
	dpuInfoMap, err := getDataFromCM[map[string]DpuCMInfo](cm, cm.Name)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("%s get dpu data from CM failed, cmName=%s, err=%v", util.DpuLogPrefix, cm.Name,
			err)
		return
	}
	klog.V(util.LogDebugLev).Infof("%s get dpu info :%v.", util.DpuLogPrefix, dpuInfoMap)
	cmMgr.dpuInfosFromCm.Lock()
	for dpuCmName, oneNodeDpuInfo := range dpuInfoMap {
		nodeName := strings.TrimPrefix(dpuCmName, util.DpuCmInfoNamePrefixByDp)
		klog.V(util.LogDebugLev).Infof("%s operator:%v.", util.DpuLogPrefix, operator)
		switch operator {
		case util.AddOperator, util.UpdateOperator:
			oneNodeDpuInfo.CacheUpdateTime = time.Now().Unix()
			klog.V(util.LogDebugLev).Infof("%s node name: %v, one node dpu info:%+v.",
				util.DpuLogPrefix, nodeName, oneNodeDpuInfo)
			cmMgr.dpuInfosFromCm.Dpus[nodeName] = oneNodeDpuInfo
		case util.DeleteOperator:
			delete(cmMgr.dpuInfosFromCm.Dpus, nodeName)
		default:
			klog.V(util.LogWarningLev).Infof("%s unknown operator:%v.", util.DpuLogPrefix, operator)
		}
	}
	cmMgr.dpuInfosFromCm.Unlock()
}

// GetDpuInfos get dpu infos
func GetDpuInfos(nodeList []*api.NodeInfo) map[string]DpuCMInfo {
	dpuInfos := make(map[string]DpuCMInfo)
	cmManager.dpuInfosFromCm.Lock()
	for _, nodeInfo := range nodeList {
		if tmpDpuInfo, ok := cmManager.dpuInfosFromCm.Dpus[nodeInfo.Name]; ok {
			dpuInfos[nodeInfo.Name] = tmpDpuInfo
		}
	}
	cmManager.dpuInfosFromCm.Unlock()
	return dpuInfos
}
