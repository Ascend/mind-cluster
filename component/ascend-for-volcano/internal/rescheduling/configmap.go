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
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
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

func (dealCM *DealReSchedulerConfigmap) newReSchedulerCMFromEnv(env *plugin.ScheduleEnv, jobType string) error {
	reCmData, getErr := util.GetConfigMapWithRetry(env.FrameAttr.KubeClient, CmNameSpace, CmName)
	if getErr != nil {
		if !errors.IsNotFound(getErr) {
			klog.V(util.LogErrorLev).Infof("newReSchedulerCMFromEnv :%s.", util.SafePrint(getErr))
			return getErr
		}
		klog.V(util.LogDebugLev).Infof("%s's configmap %s not in env", RePropertyName, CmName)
		cmData, err := dealCM.createEmptyReCM(env.FrameAttr.KubeClient, jobType)
		if err != nil {
			return fmt.Errorf("create %s configmap %s configmap failed: %s", RePropertyName, CmName, util.SafePrint(err))
		}
		dealCM.setCMName(CmName)
		dealCM.setCMNameSpace(CmNameSpace)
		dealCM.setCMData(cmData)
		klog.V(util.LogInfoLev).Infof("configmap %s in %s has been created", CmName, CmNameSpace)
		return nil
	}

	if err := checkReSchedulerCMCheckCode(reCmData.Data); err != nil {
		klog.V(util.LogErrorLev).Infof("newReSchedulerCMFromEnv: %v", err)
		return fmt.Errorf("newReSchedulerCMFromEnv: %v", err)
	}
	klog.V(util.LogInfoLev).Infof("%s configmap %s: checkCode success", RePropertyName, CmName)
	dealCM.setCMName(reCmData.Name)
	dealCM.setCMNameSpace(reCmData.Namespace)
	dealCM.setCMData(reCmData.Data)
	return nil
}

func checkReSchedulerCMCheckCode(data map[string]string) error {
	checkCode, ok := data[CmCheckCode]
	if !ok {
		return fmt.Errorf("configmap %s in %s has no checkcode", CmName, CmNameSpace)
	}
	delete(data, CmCheckCode)
	curCheckCode := util.MakeDataHash(data)
	if checkCode != curCheckCode {
		klog.V(util.LogErrorLev).Infof("checkCode err:(%s), calc(%s), data: %#v", checkCode, curCheckCode, data)
		return fmt.Errorf("checkCode does not match")
	}
	return nil
}

func (dealCM *DealReSchedulerConfigmap) createEmptyReCM(kubeClient kubernetes.Interface,
	jobType string) (map[string]string, error) {
	cmData := make(map[string]string, util.MapInitNum)
	cmData[CmFaultNodeKind] = ""
	cmData[jobType] = ""
	cmData[CmNodeHeartbeatKind] = ""
	cmData[CmNodeRankTimeMapKind] = ""
	cmData[CmJobRemainRetryTimes] = ""
	checkCode := util.MakeDataHash(cmData)
	cmData[CmCheckCode] = checkCode

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
