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

func requestSlowNodeAlgo(slowNodeCtx *sm.SlowNodeContext, command string) error {
	filePath := sm.ClusterFilePath
	if slowNodeCtx.Deployment == enum.Node {
		filePath = sm.NodeFilePath
	}

	slownodeInput := slownode.SlowNodeInput{
		EventType: enum.SlowNodeAlgo,
		SlowNodeAlgoInput: slownode.SlowNodeAlgoInput{
			SlowNodeInputBase: slowNodeCtx.Job.SlowNodeInputBase,
			FilePath:          filePath,
		},
	}

	confJson, err := json.Marshal(slownodeInput)
	if err != nil {
		return err
	}
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
	hwlog.RunLog.Infof("[FD-OL SLOWNODE]job(name=%s, jobId=%s) %s started slow node algo successfully",
		slowNodeCtx.Job.JobName, slowNodeCtx.Job.JobId, command)
	return nil
}

// StartSlowNodeAlgo start the slow node algorithm both for node and cluster
func StartSlowNodeAlgo(slowNodeCtx *sm.SlowNodeContext) error {
	return requestSlowNodeAlgo(slowNodeCtx, start)
}

// StopSlowNodeAlgo stop the slow node algorithm both for node and cluster
func StopSlowNodeAlgo(slowNodeCtx *sm.SlowNodeContext) error {
	return requestSlowNodeAlgo(slowNodeCtx, stop)
}

func watchingStartSlowNodeAlgo(slowNodeCtx *sm.SlowNodeContext) {
	logPrefix := fmt.Sprintf("[FD-OL SLOWNODE]job(name=%s, jobId=%s)",
		slowNodeCtx.Job.JobName, slowNodeCtx.Job.JobId)
	hwlog.RunLog.Infof("%s started to watch the start slow node algo signal.", logPrefix)
	go func() {
		for {
			select {
			case <-slowNodeCtx.StartSlowNodeAlgoSign:
				hwlog.RunLog.Infof("%s received slow node algo signal, start slow node algo.", logPrefix)
				if err := StartSlowNodeAlgo(slowNodeCtx); err != nil {
					hwlog.RunLog.Errorf("%s started slow node algo failed, waiting redo job, err is: %v.",
						logPrefix, err)
					slowNodeCtx.Failed()
				} else {
					hwlog.RunLog.Infof("%s started slow node algo successfully, exiting the signal watching process.",
						logPrefix)
					// for node: step from 2 to 3
					// for cluster: step from 1 to 2
					slowNodeCtx.AddStep()
					return
				}
			case <-slowNodeCtx.StopChan:
				hwlog.RunLog.Infof("%s stopped, exiting the start slow node algo signal watching process ", logPrefix)
				return
			}
		}
	}()
}

func watchingStopSlowNodeAlgo(slowNodeCtx *sm.SlowNodeContext) {
	logPrefix := fmt.Sprintf("[FD-OL SLOWNODE]job(name=%s, jobid=%s)",
		slowNodeCtx.Job.JobName, slowNodeCtx.Job.JobId)
	hwlog.RunLog.Infof("%s started to watch the stop slow node algo signal.", logPrefix)
	go func() {
		<-slowNodeCtx.StopSlowNodeAlgoSign
		hwlog.RunLog.Infof("%s received stop slow node algo signal, stop slow node algo.", logPrefix)
		if err := StopSlowNodeAlgo(slowNodeCtx); err != nil {
			hwlog.RunLog.Errorf("%s stopped slow node algo failed: %v.", logPrefix, err)
			return
		}
		hwlog.RunLog.Infof("%s stopped slow node algo successfully.", logPrefix)
	}()
}
