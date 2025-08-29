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

// Package dataparse start fd and run slownode
package dataparse

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/core/context"
	"ascend-faultdiag-online/pkg/core/model/enum"
	"ascend-faultdiag-online/pkg/model/slownode"
	"ascend-faultdiag-online/pkg/service/servicefunc/slownode/constants"
	"ascend-faultdiag-online/pkg/service/servicefunc/slownode/slownodejob"
)

// Controller is to control the data parse
type Controller struct {
	ctx *slownodejob.JobContext
}

// NewController return a new Controller pointer
func NewController(ctx *slownodejob.JobContext) *Controller {
	return &Controller{ctx: ctx}
}

func (d *Controller) request(command enum.Command) error {
	dataParseInput := slownode.DataParseInput{
		FilePath:          constants.NodeFilePath,
		ParallelGroupPath: make([]string, 0),
		RankIds:           d.ctx.Job.RankIds,
	}
	dataParseInput.JobName = d.ctx.Job.JobName
	dataParseInput.JobId = d.ctx.Job.JobId
	if d.ctx.Deployment == enum.Cluster {
		dataParseInput.FilePath = constants.ClusterFilePath
		nodeIps := d.ctx.GetReportedNodeIps()
		dataParseInput.ParallelGroupPath = make([]string, len(nodeIps))
		for i, ip := range nodeIps {
			dataParseInput.ParallelGroupPath[i] = ip + constants.ParallelGroupSuffix
		}
	}
	input := slownode.ReqInput{
		EventType:      enum.DataParse,
		DataParseInput: dataParseInput,
	}
	confJson, err := json.Marshal(input)
	if err != nil {
		return err
	}
	hwlog.RunLog.Infof("%v %s data parse, confJson: %s", d.ctx.LogPrefix(), command, string(confJson))
	apiPath := fmt.Sprintf("feature/slownode/%s/%s", d.ctx.Deployment, command)
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
	hwlog.RunLog.Infof("%v %s data parse success, resp: %s", d.ctx.LogPrefix(), command, resp)
	return nil
}

// Start start the data parse
func (d *Controller) Start() {
	if err := d.request(enum.Start); err != nil {
		hwlog.RunLog.Errorf("%v started data parse failed: %v", d.ctx.LogPrefix(), err)
	}
}

// Stop stop the data parse
func (d *Controller) Stop() {
	if err := d.request(enum.Stop); err != nil {
		hwlog.RunLog.Errorf("%v stopped data parse failed: %v", d.ctx.LogPrefix(), err)
	}
}

// MergeParallelGroupInfoWatcher watching the merge parallel group info signal
// no need to watch the stop signal, merge parallel group info would not run forever
// only occurs in cluster
func (d *Controller) MergeParallelGroupInfoWatcher() {
	go func() {
		hwlog.RunLog.Infof("%s started watching merge parallel group info signal, timeout: %d",
			d.ctx.LogPrefix(), context.FdCtx.Config.AllNodesReportTimeout)
		select {
		case <-d.ctx.MergeParallelGroupInfoSignal:
			d.handleMergeSignal("received signal")
		case <-time.After(time.Duration(context.FdCtx.Config.AllNodesReportTimeout) * time.Second):
			d.handleMergeSignal(fmt.Sprintf("timeout after %d seconds", context.FdCtx.Config.AllNodesReportTimeout))
		case _, ok := <-d.ctx.StopChan:
			if !ok {
				hwlog.RunLog.Infof("%s stopped, exiting merge signal watcher", d.ctx.LogPrefix())
				return
			}
		}
	}()
}

func (d *Controller) handleMergeSignal(triggerReason string) {
	hwlog.RunLog.Infof("%s %s, merging parallel group info (reported nodes: %v)",
		d.ctx.LogPrefix(), triggerReason, d.ctx.GetReportedNodeIps())

	d.ctx.AddStep() // Advance cluster step (e.g., from 1 to 2)
	d.ctx.StopHeavyProfiling()
	d.Start()
	hwlog.RunLog.Infof("%s merge succeeded, exiting watcher", d.ctx.LogPrefix())
}
