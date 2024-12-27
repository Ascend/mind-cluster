/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.

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

/*
Package rescheduling is using for HuaWei Ascend pin fault rescheduling.
*/
package rescheduling

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

func (dealCM *DealReSchedulerConfigmap) setCMData(value map[string]string) {
	dealCM.CMData = value
}

func (dealCM *DealReSchedulerConfigmap) setCMName(value string) {
	dealCM.CMName = value
}

func (dealCM *DealReSchedulerConfigmap) setCMNameSpace(value string) {
	dealCM.CMNameSpace = value
}

func newReSchedulerCM() *DealReSchedulerConfigmap {
	dealCM := &DealReSchedulerConfigmap{}
	dealCM.setCMName(CmName)
	dealCM.setCMNameSpace(CmNameSpace)
	dealCM.setCMData(newEmptyCacheData())
	klog.V(util.LogInfoLev).Infof("configmap %s in %s has been created", CmName, CmNameSpace)
	return dealCM
}

func (dealCM *DealReSchedulerConfigmap) updateReSchedulerCMCache(cmDate map[string]string) {
	if dealCM == nil {
		dealCM = newReSchedulerCM()
	}
	dealCM.setCMData(cmDate)
	klog.V(util.LogInfoLev).Infof("configmap %s in cache %s has been update", CmName, CmNameSpace)
}

func (dealCM *DealReSchedulerConfigmap) createEmptyReCM(kubeClient kubernetes.Interface,
	jobType string) (map[string]string, error) {
	cmData := newEmptyCacheData()

	var faultCM = &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      CmName,
			Namespace: CmNameSpace,
		},
		Data: cmData,
	}
	err := util.CreateOrUpdateConfigMap(kubeClient, faultCM, CmName, CmNameSpace)
	if err != nil {
		return cmData, err
	}
	return cmData, nil
}

func newEmptyCacheData() map[string]string {
	cacheData := make(map[string]string, util.MapInitNum)
	cacheData[CmFaultNodeKind] = ""
	cacheData[CmFaultJob] = ""
	cacheData[CmNodeRankTimeMapKind] = ""
	cacheData[CmJobRemainRetryTimes] = ""
	checkCode := util.MakeDataHash(cacheData)
	cacheData[CmCheckCode] = checkCode
	return cacheData
}
