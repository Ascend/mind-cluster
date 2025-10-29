/*
Copyright(C)2025. Huawei Technologies Co.,Ltd. All rights reserved.

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

// Package node provides some func to process the data in node
package node

import (
	"encoding/json"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/model/slownode"
	"ascend-faultdiag-online/pkg/service/servicefunc/slownode/algo"
	"ascend-faultdiag-online/pkg/service/servicefunc/slownode/common"
	"ascend-faultdiag-online/pkg/service/servicefunc/slownode/constants"
	"ascend-faultdiag-online/pkg/service/servicefunc/slownode/slownodejob"
	"ascend-faultdiag-online/pkg/utils"
	globalConstants "ascend-faultdiag-online/pkg/utils/constants"
	"ascend-faultdiag-online/pkg/utils/fileutils"
	"ascend-faultdiag-online/pkg/utils/k8s"
)

// AlgoCallbackProcessor process the algo callback
func AlgoCallbackProcessor(message string) {
	var data = map[string]map[string]any{}
	if err := json.Unmarshal([]byte(message), &data); err != nil {
		hwlog.RunLog.Errorf("[FD-OL SLOWNODE]json unmarshal data: %s failed: %v", message, err)
		return
	}
	var nodeResult = slownode.NodeAlgoResult{}
	if err := common.ConvertMaptoStruct(data, &nodeResult); err != nil {
		hwlog.RunLog.Errorf("[FD-OL SLOWNODE]convert map data: %s to struct failed: %v", data, err)
		return
	}

	ctx, ok := slownodejob.GetJobCtxMap().Get(nodeResult.JobId)
	if !ok {
		hwlog.RunLog.Warnf(
			"[FD-OL SLOWNODE]job(jobId=%s) is not exited, exit slow node algo callback process", nodeResult.JobId)
		return
	}
	if ctx == nil || ctx.Job == nil {
		hwlog.RunLog.Error("[FD-OL SLOWNODE]process slow node algo callback: invalid nil context or job")
		return
	}
	if !ctx.IsRunning() {
		hwlog.RunLog.Errorf("%s process slow node algo callback: not running", ctx.LogPrefix())
		return
	}
	if nodeResult.IsSlow == constants.IsDegradation {
		hwlog.RunLog.Infof("%s detected node: %v is degradation, slow rank ids: %v",
			ctx.LogPrefix(), nodeResult.NodeRank, nodeResult.SlowCalculateRanks)
	}
	nodeResult.JobName = ctx.Job.JobName
	nodeResult.Namespace = ctx.Job.Namespace
	cmName := fmt.Sprintf("%s-%s-%s", constants.NodeAlgoResultPrefix, nodeResult.JobId, nodeResult.NodeRank)
	if err := callbackReport(&nodeResult, cmName, constants.NodeAlgoResultCMKey, ctx); err != nil {
		hwlog.RunLog.Errorf("%s reported node algo result failed: %v", ctx.LogPrefix(), err)
	}
}

// DataParseCallbackProcessor process the data parse callback
func DataParseCallbackProcessor(message string) {
	var dataParseResult slownode.DataParseResult
	if err := json.Unmarshal([]byte(message), &dataParseResult); err != nil {
		hwlog.RunLog.Errorf("[FD-OL SLOWNODE]json marshal data parse callback data: %v failed: %s", message, err)
		return
	}
	ctx, ok := slownodejob.GetJobCtxMap().Get(dataParseResult.JobId)
	if !ok {
		hwlog.RunLog.Errorf(
			"[FD-OL SLOWNODE]job(name=%s, jobId=%s) is not existed, exit data parse callback process",
			dataParseResult.JobName, dataParseResult.JobId)
		return
	}
	if ctx == nil || ctx.Job == nil {
		hwlog.RunLog.Error("[FD-OL SLOWNODE]process data parse callback: invalid nil context or job")
		return
	}
	if !ctx.IsRunning() {
		hwlog.RunLog.Errorf("%s process data parse callback: not running", ctx.LogPrefix())
		return
	}
	if dataParseResult.IsFinished && dataParseResult.StepCount >= constants.MinStepCount &&
		ctx.Step() == slownodejob.NodeStep1 {
		if err := dataProfilingReport(ctx); err != nil {
			hwlog.RunLog.Errorf("%s reported node data profiling result failed: %v", ctx.LogPrefix(), err)
			return
		}
		ctx.RealRankIds = dataParseResult.RankIds
		// step from 1 to 2
		ctx.AddStep()
		algo.NewController(ctx).Start()
	}
}

// nodeReportProfiling report the data parsed result to cm
func dataProfilingReport(ctx *slownodejob.JobContext) error {
	parallelGroupInfo, err := parallelGroupInfoReader(ctx.Job.JobId)
	if err != nil {
		return err
	}
	nodeIp, err := utils.GetNodeIp()
	if err != nil {
		return err
	}
	// finish profiling
	var profilingResult = slownode.NodeDataProfilingResult{
		Namespace:                ctx.Job.Namespace,
		FinishedInitialProfiling: true,
		FinishedTime:             time.Now().Unix(),
		NodeIp:                   nodeIp,
		ParallelGroupInfo:        parallelGroupInfo,
	}
	profilingResult.JobName = ctx.Job.JobName
	profilingResult.JobId = ctx.Job.JobId
	cmName := fmt.Sprintf("%s-%s-%s", constants.NodeDataProfilingResultPrefix, ctx.Job.JobId, nodeIp)
	return callbackReport(&profilingResult, cmName, constants.NodeDataProfilingResultCMKey, ctx)
}

func callbackReport[T slownode.NodeAlgoResult | slownode.NodeDataProfilingResult](
	data *T, cmName, cmKey string, ctx *slownodejob.JobContext) error {
	dataBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	err = createOrUpdateCM(cmName, ctx.Job.Namespace, cmKey, dataBytes)
	if err != nil {
		return err
	}
	k8sClient, err := k8s.GetClient()
	if err != nil {
		return err
	}
	if err = k8sClient.DeleteConfigMap(cmName, ctx.Job.Namespace); err != nil {
		return err
	}
	return nil
}

func createOrUpdateCM(name, namespace, cmKey string, data []byte) error {
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    map[string]string{constants.CmConsumer: constants.CmConsumerValue},
		},
		Data: map[string]string{
			cmKey: string(data),
		},
	}
	k8sClient, err := k8s.GetClient()
	if err != nil {
		hwlog.RunLog.Errorf("[FD-OL SLOWNODE]create k8s client failed: %v", err)
		return err
	}
	if err := k8sClient.CreateOrUpdateConfigMap(cm); err != nil {
		return err
	}
	hwlog.RunLog.Infof("[FD-OL SLOWNODE]create or update configmap: [key: %s, value: %+v] successfully",
		cm.Name, cm.Data)
	return nil
}

func parallelGroupInfoReader(jobId string) (map[string]any, error) {
	filePath := fmt.Sprintf("%s/%s/%s", constants.NodeFilePath, jobId, constants.ParallelGroupInfo)
	fileContent, err := fileutils.ReadLimitBytes(filePath, globalConstants.Size10M)
	if err != nil {
		return nil, err
	}
	var data = map[string]any{}
	err = json.Unmarshal(fileContent, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}
