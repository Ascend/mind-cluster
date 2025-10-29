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

// Package cluster a series of function relevant to process the result of slow node node
package cluster

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"k8s.io/apimachinery/pkg/watch"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/model/slownode"
	"ascend-faultdiag-online/pkg/service/servicefunc/slownode/common"
	"ascend-faultdiag-online/pkg/service/servicefunc/slownode/constants"
	"ascend-faultdiag-online/pkg/service/servicefunc/slownode/dataparse"
	"ascend-faultdiag-online/pkg/service/servicefunc/slownode/slownodejob"
	"ascend-faultdiag-online/pkg/utils/fileutils"
)

var (
	jobOnceMap sync.Map // jobId: sync.Once
)

const (
	algoResultSuffix = "_Result.json"
	algoResultPrefix = "slownode_"
)

// DataProfilingResultProcessor process the data profiling callback from FD-OL in node
func DataProfilingResultProcessor(oldData, newData *slownode.NodeDataProfilingResult, operator watch.EventType) {
	if newData == nil {
		hwlog.RunLog.Error("[FD-OL SLOWNODE]data profiling result is nil")
		return
	}
	if (operator != watch.Added && operator != watch.Modified) || !newData.FinishedInitialProfiling {
		return
	}
	hwlog.RunLog.Infof("[FD-OL SLOWNODE]got data profiling cm data, operator: %s, data: %+v", operator, newData)
	ctx, ok := slownodejob.GetJobCtxMap().Get(newData.KeyGenerator())
	if !ok {
		hwlog.RunLog.Errorf("[FD-OL SLOWNODE]process data profiling callback: job(name=%s, jobId=%s) is not exited",
			newData.JobName, newData.JobId)
		return
	}
	if ctx == nil || ctx.Job == nil {
		hwlog.RunLog.Error("[FD-OL SLOWNODE]process data profiling callback: invalid nil context or job")
		return
	}
	if !ctx.IsRunning() {
		hwlog.RunLog.Errorf("%s process data profiling callback: not running", ctx.LogPrefix())
		return
	}
	if ctx.Step() != slownodejob.ClusterStep1 {
		// ClusterStep1 means started all profiling
		// ClusterStep2 means started merge parallel group info
		hwlog.RunLog.Warnf("%s has been started merge paralle group info, ignore the data profiling result",
			ctx.LogPrefix())
		return
	}
	// if the first node report the profiling result, start merge paralle group info watcher
	once, _ := jobOnceMap.LoadOrStore(newData.JobId, &sync.Once{})
	once.(*sync.Once).Do(func() {
		dataparse.NewController(ctx).MergeParallelGroupInfoWatcher()
		ctx.NodeReportSignal <- struct{}{}
	})
	// write the tp/pp data into local file
	fileName := newData.NodeIp + constants.ParallelGroupSuffix
	dir := fmt.Sprintf("%s/%s", constants.ClusterFilePath, newData.JobId)
	if err := writeFile(dir, fileName, newData.ParallelGroupInfo); err != nil {
		hwlog.RunLog.Errorf("%s write parallel group info to file failed: %v", ctx.LogPrefix(), err)
		return
	}
	ctx.AddReportedNodeIp(newData.NodeIp)
	hwlog.RunLog.Infof("%s wrote parallel group info to file(%s) successfully", ctx.LogPrefix(), fileName)
	// all saved files matches the nodeIps, stop heavy profiling and strat slow node algo
	if ctx.AllNodesReported() {
		hwlog.RunLog.Infof("%s has been wroten all the parallel group data, "+
			"stop heavy profiling and start slow node algo", ctx.LogPrefix())
		ctx.TriggerMerge()
	}
}

// AlgoResultProcessor process the slow node algo result, write to file
func AlgoResultProcessor(oldData, newData *slownode.NodeAlgoResult, operator watch.EventType) {
	if newData == nil {
		hwlog.RunLog.Error("[FD-OL SLOWNODE]data node algo result is nil")
		return
	}
	if operator != watch.Added && operator != watch.Modified {
		return
	}
	hwlog.RunLog.Infof("[FD-OL SLOWNODE]got node slow node algo cm data, operator: %s, newObj: %+v", operator, newData)
	if err := common.NodeRankValidator(newData.NodeRank); err != nil {
		hwlog.RunLog.Errorf("[FD-OL SLOWNODE]invalid node rank: %s, err: %v", newData.NodeRank, err)
		return
	}
	var key = newData.KeyGenerator()
	ctx, ok := slownodejob.GetJobCtxMap().Get(key)
	if !ok || !ctx.IsRunning() {
		hwlog.RunLog.Errorf("[FD-OL SLOWNODE]process slow node algo result: job(name=%s, jobId=%s) is not exited",
			newData.JobName, newData.JobId)
		return
	}
	if ctx == nil || ctx.Job == nil {
		hwlog.RunLog.Error("[FD-OL SLOWNODE]process slow node algo result: invalid nil context or job")
		return
	}
	if !ctx.IsRunning() {
		hwlog.RunLog.Errorf("%s process slow node algo result: not running", ctx.LogPrefix())
		return
	}
	if _, ok := ctx.ReportedNodeIps.Load(newData.NodeRank); !ok {
		hwlog.RunLog.Errorf("%sgnores node: [%s] algo result, this node had not been reported the parallel group "+
			"info due to the timeout", ctx.LogPrefix(), newData.NodeRank)
		return
	}
	fileName := newData.NodeRank + algoResultSuffix
	dir := fmt.Sprintf("%s/%s/%s", constants.ClusterFilePath, newData.JobId, constants.NodeLevelDetectionResult)
	var data = map[string]any{
		algoResultPrefix + newData.JobId: map[string]any{
			newData.NodeRank: newData,
		},
	}
	if err := writeFile(dir, fileName, data); err != nil {
		hwlog.RunLog.Errorf("%s write slow node algo result to file failed: %v", ctx.LogPrefix(), err)
		return
	}
	hwlog.RunLog.Infof("%s wrote slow node algo result to file(%s) successfully", ctx.LogPrefix(), fileName)
}

func writeFile(dir, fileName string, data map[string]any) error {
	filePath := fmt.Sprintf("%s/%s", dir, fileName)
	absPath, err := fileutils.CheckPath(filePath)
	if err != nil {
		return err
	}
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	if fileInfo, err := os.Lstat(dir); err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(dir, os.ModePerm) // create directory, including necessary parent directories
			if err != nil {
				return err
			}
		}
	} else if (fileInfo.Mode() & os.ModeSymlink) != 0 {
		return err
	}
	// write the data
	return os.WriteFile(absPath, bytes, constants.FileMode)
}
