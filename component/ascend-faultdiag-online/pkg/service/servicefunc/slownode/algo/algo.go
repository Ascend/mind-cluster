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

// Package algo start fd and run slownode
package algo

import (
	"encoding/json"
	"errors"
	"fmt"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/core/context"
	"ascend-faultdiag-online/pkg/core/model/enum"
	"ascend-faultdiag-online/pkg/model/slownode"
	"ascend-faultdiag-online/pkg/service/servicefunc/slownode/constants"
	"ascend-faultdiag-online/pkg/service/servicefunc/slownode/slownodejob"
)

// Controller is to control the algo
type Controller struct {
	ctx *slownodejob.JobContext
}

// NewController return a new AlgoContorller pointer
func NewController(ctx *slownodejob.JobContext) *Controller {
	return &Controller{ctx: ctx}
}

func (a *Controller) request(command enum.Command) error {
	input := slownode.ReqInput{
		EventType: enum.SlowNodeAlgo,
		AlgoInput: slownode.AlgoInput{
			InputBase: a.ctx.Job.InputBase,
			FilePath:  constants.ClusterFilePath,
		},
	}
	if a.ctx.Deployment == enum.Node {
		input.AlgoInput.FilePath = constants.NodeFilePath
		input.AlgoInput.RankIds = a.ctx.RealRankIds
	}

	confJson, err := json.Marshal(input)
	if err != nil {
		return err
	}
	apiPath := fmt.Sprintf("feature/slownode/%s/%s", a.ctx.Deployment, command)
	resp, err := context.FdCtx.Request(apiPath, string(confJson))
	if err != nil {
		return err
	}
	var res = slownode.ApiRes{}
	if err = json.Unmarshal([]byte(resp), &res); err != nil {
		return err
	}
	if res.Status != enum.Success {
		return errors.New(res.Msg)
	}
	hwlog.RunLog.Infof("[FD-OL SLOWNODE]job(name=%s, jobId=%s) %s slow node algo successfully",
		a.ctx.Job.JobName, a.ctx.Job.JobId, command)
	return nil
}

// Start start the slow node algorithm both for node and cluster
func (a *Controller) Start() {
	if err := a.request(enum.Start); err != nil {
		hwlog.RunLog.Errorf("[FD-OL SLOWNODE]job(name=%s, jobId=%s) started slow node algo failed: %v",
			a.ctx.Job.JobName, a.ctx.Job.JobId, err)
		return
	}
}

// Stop stop the slow node algorithm both for node and cluster
func (a *Controller) Stop() {
	if err := a.request(enum.Stop); err != nil {
		hwlog.RunLog.Errorf("[FD-OL SLOWNODE]job(name=%s, jobId=%s) stopped slow node algo failed: %v",
			a.ctx.Job.JobName, a.ctx.Job.JobId, err)
		return
	}
}
