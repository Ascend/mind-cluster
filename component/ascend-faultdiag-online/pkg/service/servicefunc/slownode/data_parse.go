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

// Package slownode start fd and run slownode
package slownode

import (
	"encoding/json"
	"errors"
	"fmt"

	"ascend-common/common-utils/hwlog"
	api "ascend-faultdiag-online/pkg/api/v1"
	"ascend-faultdiag-online/pkg/global/globalctx"
	"ascend-faultdiag-online/pkg/model/enum"
	"ascend-faultdiag-online/pkg/model/feature/slownode"
	sm "ascend-faultdiag-online/pkg/module/slownode"
)

func requestDataParse(slowNodeCtx *sm.SlowNodeContext, command string) error {
	dataParseInput := slownode.DataParseInput{
		FilePath:          sm.NodeFilePath,
		JobName:           slowNodeCtx.Job.JobName,
		JobId:             slowNodeCtx.Job.JobId,
		ParallelGroupPath: make([]string, 0),
	}
	if slowNodeCtx.Deployment == enum.Cluster {
		dataParseInput.FilePath = sm.ClusterFilePath
		dataParseInput.ParallelGroupPath = make([]string, len(slowNodeCtx.WorkerNodesIP))
		for i, nodeIP := range slowNodeCtx.WorkerNodesIP {
			dataParseInput.ParallelGroupPath[i] = nodeIP + parallelGroupSuffix
		}
	}
	input := slownode.SlowNodeInput{
		EventType:      enum.DataParse,
		DataParseInput: dataParseInput,
	}
	confJson, err := json.Marshal(input)
	if err != nil {
		return err
	}
	hwlog.RunLog.Infof("[FD-OL SLOWNODE]job(name=%s, namespace=%s) %s data parse, confJson: %s",
		slowNodeCtx.Job.JobName, slowNodeCtx.Job.Namespace, command, string(confJson))
	apiPath := fmt.Sprintf("feature/slownode/%s/%s", slowNodeCtx.Deployment, command)
	resp, err := api.Request(globalctx.Fdctx, apiPath, string(confJson))
	if err != nil {
		return err
	}
	var res = slownode.ApiRes{}
	if err = json.Unmarshal([]byte(resp), &res); err != nil {
		return err
	}
	if res.Status != success {
		return errors.New(res.Msg)
	}
	hwlog.RunLog.Infof("[FD-OL SLOWNODE]job(name=%s, namespace=%s) %s data parse success, resp: %s",
		slowNodeCtx.Job.JobName, slowNodeCtx.Job.Namespace, command, resp)
	return nil
}

func startDataParse(slowNodeCtx *sm.SlowNodeContext) error {
	return requestDataParse(slowNodeCtx, start)
}

func stopDataParse(slowNodeCtx *sm.SlowNodeContext) error {
	return requestDataParse(slowNodeCtx, stop)
}

func watchingStartDataParse(slowNodeCtx *sm.SlowNodeContext) {
	logPrefix := fmt.Sprintf("[FD-OL SLOWNODE]job(name=%s, namespace=%s)",
		slowNodeCtx.Job.JobName, slowNodeCtx.Job.Namespace)
	hwlog.RunLog.Infof("%s started to watch the start data parse signal.", logPrefix)
	go func() {
		for {
			select {
			case <-slowNodeCtx.StartDataParseSign:
				hwlog.RunLog.Infof("%s received data parse signal, start data parse.", logPrefix)
				if err := startDataParse(slowNodeCtx); err != nil {
					hwlog.RunLog.Errorf(
						"%s started data parse failed, waiting the redo job, err is: %v.", logPrefix, err)
					slowNodeCtx.Failed()
				} else {
					hwlog.RunLog.Infof(
						"%s started data parse successfully, exit signal watching process.", logPrefix)
					// step from 0 to 1
					slowNodeCtx.AddStep()
					return
				}
			case <-slowNodeCtx.StopChan:
				hwlog.RunLog.Infof("%s stopped, exit start data parse signal watching process ", logPrefix)
				return
			}
		}
	}()
}

func watchingStopDataParse(slowNodeCtx *sm.SlowNodeContext) {
	logPrefix := fmt.Sprintf("[FD-OL SLOWNODE]job(name=%s, namespace=%s)",
		slowNodeCtx.Job.JobName, slowNodeCtx.Job.Namespace)
	hwlog.RunLog.Infof("%s started to watch the stop data parse signal.", logPrefix)
	go func() {
		<-slowNodeCtx.StopDataParseSign
		hwlog.RunLog.Infof("%s received stop data parse signal, stop data parse.", logPrefix)
		if err := stopDataParse(slowNodeCtx); err != nil {
			hwlog.RunLog.Errorf("%s stopped data parse failed: %v.", logPrefix, err)
			return
		}
		hwlog.RunLog.Infof("%s stopped data parse successfully.", logPrefix)
	}()
}
