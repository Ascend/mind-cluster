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

// Package slownode a series of function relevant to the fd-ol deployed in cluster
package slownode

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/global"
	"ascend-faultdiag-online/pkg/model/enum"
	"ascend-faultdiag-online/pkg/model/feature/slownode"
	sm "ascend-faultdiag-online/pkg/module/slownode"
)

// ClusterProcessSlowNodeJob store the slow node feat config into the confMap in cluster
func ClusterProcessSlowNodeJob(oldData, newData *slownode.SlowNodeJob, operator string) {
	hwlog.RunLog.Infof("[FD-OL SLOWNODE]got job cm data, oldObj: %+v, newObj: %+v, operator: %s",
		oldData, newData, operator)

	if newData.JobName == "" {
		hwlog.RunLog.Errorf("[FD-OL SLOWNODE]job(name=%s): jobName is empty, ignore it.", newData.JobName)
		return
	}
	var delayTime = 2
	if operator == DeleteOperator {
		delayTime = 0
	}
	if err := getJobIdDelayed(newData, delayTime); err != nil {
		hwlog.RunLog.Errorf("[FD-OL SLOWNODE]job(name=%s, jobId=%s) queried job id failed: %s",
			newData.JobName, newData.JobId, err)
		return
	}
	switch operator {
	case AddOperator:
		processAddJob(newData)
	case UpdateOperator:
		processUpdateJob(newData)
	case DeleteOperator:
		processDeleteJob(newData)
	default:
		return
	}
}

func processAddJob(job *slownode.SlowNodeJob) {
	slowNodeCtxMap := sm.GetSlowNodeCtxMap()
	_, ok := slowNodeCtxMap.Get(job.JobId)
	if ok {
		hwlog.RunLog.Warnf(
			"[FD-OL SLOWNODE]job(name=%s, jobId=%s) has been existed in SlowNodeCtxMap, ignore it.",
			job.JobName, job.JobId)
		return
	}
	slowNodeCtx := sm.NewSlowNodeContext(job, enum.Cluster)
	nodesIp, err := global.K8sClient.GetWorkerNodesIPByLabel(keyJobName, job.JobName)
	if err != nil {
		hwlog.RunLog.Errorf("[FD-OL SLOWNODE]job(name=%s, jobId=%s) get all suitable nodes ip failed: %s",
			job.JobName, job.JobId, err)
		return
	}
	hwlog.RunLog.Infof("[FD-OL SLOWNODE]job(name=%s, jobId=%s) get all suitable nodes ip: %s",
		job.JobName, job.JobId, nodesIp)
	slowNodeCtx.InsertNodesIP(nodesIp)
	slowNodeCtxMap.Insert(job.JobId, slowNodeCtx)
	if job.SlowNode == SlowNodeOn {
		clusterStart(slowNodeCtx)
	}
}

func processUpdateJob(job *slownode.SlowNodeJob) {
	slowNodeCtx, ok := sm.GetSlowNodeCtxMap().Get(job.JobId)
	if !ok {
		hwlog.RunLog.Errorf("[FD-OL SLOWNODE]job(name=%s, jobId=%s) does not exist in SlowNodeCtxMap, ignore it.",
			job.JobName, job.JobId)
		return
	}
	slowNodeCtx.Update(job)
	if !slowNodeCtx.IsRunning() && job.SlowNode == SlowNodeOn {
		clusterStart(slowNodeCtx)
	} else if slowNodeCtx.IsRunning() && slowNodeCtx.Job.SlowNode == SlowNodeOff {
		clusterStop(slowNodeCtx)
	}
}

func processDeleteJob(job *slownode.SlowNodeJob) {
	slowNodeCtxMap := sm.GetSlowNodeCtxMap()
	slowNodeCtx, ok := slowNodeCtxMap.Get(job.JobId)
	if !ok {
		hwlog.RunLog.Errorf("[FD-OL SLOWNODE]job(name=%s, jobId=%s) does not exist in SlowNodeCtxMap, ignore it.",
			job.JobName, job.JobId)
		return
	}
	if slowNodeCtx.IsRunning() {
		clusterStop(slowNodeCtx)
	}
	slowNodeCtxMap.Delete(job.JobId)
	// delete cmKeys
	eliminateCM(slowNodeCtx)
}

// ClusterProcessDataProfilingResult process the data profiling callback from FD-OL in node
func ClusterProcessDataProfilingResult(oldData, newData *slownode.NodeDataProfilingResult, operator string) {
	hwlog.RunLog.Infof("[FD-OL SLOWNODE]got data profiling cm data, oldObj: %+v, newObj: %+v, operator: %s",
		oldData, newData, operator)
	if (operator != AddOperator && operator != UpdateOperator) || !newData.FinishedInitialProfiling {
		return
	}
	slowNodeCtx, ok := sm.GetSlowNodeCtxMap().Get(newData.JobId)
	if !ok || !slowNodeCtx.IsRunning() {
		hwlog.RunLog.Errorf(
			"[FD-OL SLOWNODE]process data profiling callback: job(name=%s, jobId=%s) is not exit or not running.",
			newData.JobName, newData.JobId)
		return
	}
	// write the tp/pp data into local file
	fileName := newData.NodeIP + parallelGroupSuffix
	dir := fmt.Sprintf("%s/%s", sm.ClusterFilePath, newData.JobId)
	clusterWriteFile(dir, fileName, newData.ParallelGroupInfo)
	// all saved files matches the nodeIPs, stop heavy profiling and strat slow node algo
	if isMatchNodeIPs(newData.JobId, slowNodeCtx.WorkerNodesIP) {
		hwlog.RunLog.Infof("[FD-OL SLOWNODE]job(name=%s, jobId=%s) has been wroten all the parallel group data, "+
			"stop heavy profiling and start slow node algo", newData.JobName, slowNodeCtx.Job.JobId)
		if slowNodeCtx.IsStartedHeavyProfiling() {
			slowNodeCtx.StopHeavyProfiling()
		}
		slowNodeCtx.StartDataParse()
	}
}

// ClusterProcessSlowNodeAlgoResult process the slow node algo result, write to file
func ClusterProcessSlowNodeAlgoResult(oldData, newData *slownode.NodeSlowNodeAlgoResult, operator string) {
	hwlog.RunLog.Infof("[FD-OL SLOWNODE]got node slow node algo cm data, oldObj: %+v, newObj: %+v, operator: %s",
		oldData, newData, operator)
	if operator != AddOperator && operator != UpdateOperator {
		return
	}
	slowNodeCtx, ok := sm.GetSlowNodeCtxMap().Get(newData.JobId)
	if !ok || !slowNodeCtx.IsRunning() {
		hwlog.RunLog.Errorf(
			"[FD-OL SLOWNODE]process slow node algo result: job(name=%s, jobId=%s) is not exit or not running.",
			newData.JobName, newData.JobId)
		return
	}
	fileName := newData.NodeRank + slownodeAlgoResultSuffix
	dir := fmt.Sprintf("%s/%s/%s", sm.ClusterFilePath, newData.JobId, nodeLevelDetectionResult)
	var data = map[string]any{
		slownodeAlgoResultPrefix + newData.JobId: map[string]any{
			newData.NodeRank: newData,
		},
	}
	clusterWriteFile(dir, fileName, data)
}

func getJobId(job *slownode.SlowNodeJob, checkRunning bool) error {
	if job.JobId != "" {
		return nil
	}
	cmName := jobSummaryPrefix + job.JobName
	for retryCount := 1; retryCount <= maxRetryCount; retryCount++ {
		cm, err := global.K8sClient.GetConfigMap(cmName, defaultNamespace)
		if err != nil {
			if errors.IsNotFound(err) {
				hwlog.RunLog.Infof(
					"[FD-OL SLOWNODE]job(name=%s, jobId=%s) queried cm(name=%s, namespace=%s) failed"+
						", cm is not found. retry %d/%d",
					job.JobName, job.JobId, cmName, defaultNamespace, retryCount, maxRetryCount)
				time.Sleep(time.Duration(1<<uint(retryCount)) * time.Second)
				continue
			}
			return fmt.Errorf("queried cm(name=%s, namespace=%s) failed: %s", cmName, defaultNamespace, err)
		}
		if checkRunning && cm.Data[keyJobStatus] != isRunning {
			hwlog.RunLog.Infof(
				"[FD-OL SLOWNODE]job(name=%s, jobId=%s) queried cm(name=%s, namespace=%s) failed"+
					", job is not running. retry %d/%d",
				job.JobName, job.JobId, cmName, defaultNamespace, retryCount, maxRetryCount)
			time.Sleep(time.Duration(1<<uint(retryCount)) * time.Second)
			continue
		}
		if cm.Data[keyJobId] != "" {
			job.JobId = cm.Data[keyJobId]
			return nil
		}
		return fmt.Errorf("no job_id in cm: %s", cm.Data)
	}
	return fmt.Errorf("queried cm(name=%s, namespace=%s) failed, reached the max retries", cmName, defaultNamespace)
}

func getJobIdDelayed(job *slownode.SlowNodeJob, delay int) error {
	checkRunning := true
	if delay == 0 {
		checkRunning = false
	}
	if err := getJobId(job, checkRunning); err != nil {
		return err
	}
	if delay != 0 {
		time.Sleep(time.Duration(delay) * time.Second)
		if err := getJobId(job, checkRunning); err != nil {
			return err
		}
	}
	return nil
}

func clusterStart(slowNodeCtx *sm.SlowNodeContext) {
	slowNodeCtx.Start()
	if err := createOrUpdateCM(slowNodeCtx); err != nil {
		hwlog.RunLog.Errorf("[FD-OL SLOWNODE]job(name=%s, jobId=%s) create or update cm feaild: %s",
			slowNodeCtx.Job.JobName, slowNodeCtx.Job.JobId, err)
		return
	}
	go watchingStartSlowNodeAlgo(slowNodeCtx)
	go watchingStopSlowNodeAlgo(slowNodeCtx)
	go clusterWatchingDataParse(slowNodeCtx)
	slowNodeCtx.StartAllProfiling()
}

func clusterStop(slowNodeCtx *sm.SlowNodeContext) {
	slowNodeCtx.Job.SlowNode = SlowNodeOff
	if err := createOrUpdateCM(slowNodeCtx); err != nil {
		hwlog.RunLog.Errorf("[FD-OL SLOWNODE]job(name=%s, jobId=%s) create or update cm feaild: %s",
			slowNodeCtx.Job.JobName, slowNodeCtx.Job.JobId, err)
		return
	}
	slowNodeCtx.StopAllProfiling()
	slowNodeCtx.StopSlowNodeAlgo()
	slowNodeCtx.Stop()
}

func createOrUpdateCM(slowNodeCtx *sm.SlowNodeContext) error {
	dataBytes, err := json.MarshalIndent(slowNodeCtx.Job, "", "  ")
	if err != nil {
		hwlog.RunLog.Errorf("[FD-OL SLOWNODE]json marshal job failed: %v", err)
		return err
	}
	cmName := sm.NodeSlowNodeJobPrefix + "-" + slowNodeCtx.Job.JobId
	slowNodeCtx.AllCMNAMEs.Store(cmName, struct{}{})
	cm := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cmName,
			Namespace: slowNodeCtx.Job.Namespace,
			Labels:    map[string]string{sm.CmConsumer: sm.CmConsumerValue},
		},
		Data: map[string]string{
			sm.NodeSlowNodeJobCMKey: string(dataBytes),
		},
	}
	if err := global.K8sClient.CreateOrUpdateConfigMap(cm); err != nil {
		return err
	}
	hwlog.RunLog.Infof("[FD-OL SLOWNODE]create or update configmap: [key: %s, value: %s] successfully",
		cm.Name, cm.Data)
	return nil
}

func clusterWriteFile(dir, fileName string, data map[string]any) {

	filePath := fmt.Sprintf("%s/%s", dir, fileName)
	bytes, err := json.Marshal(data)
	if err != nil {
		hwlog.RunLog.Errorf("[FD-OL SLOWNODE]json marshal data[%+v] failed: %v", data, err)
		return
	}

	if fileInfo, err := os.Lstat(dir); err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(dir, os.ModePerm) // create directory, including necessary parent directories
			if err != nil {
				hwlog.RunLog.Errorf("[FD-OL SLOWNODE]created directory failed: %v", err)
				return
			}
		}
	} else if (fileInfo.Mode() & os.ModeSymlink) != 0 {
		hwlog.RunLog.Errorf("[FD-OL SLOWNODE]the path %s is a symlink, unsupport", dir)
		return
	}

	// write the data
	err = os.WriteFile(filePath, bytes, FileMode)
	if err != nil {
		hwlog.RunLog.Errorf("[FD-OL SLOWNODE]wrote bytes to file(%s) failed: %v", filePath, err)
		return
	}
	hwlog.RunLog.Infof("[FD-OL SLOWNODE]wrote data to file(%s) successfully", filePath)
}

func isMatchNodeIPs(jobId string, nodesIP []string) bool {
	dir := fmt.Sprintf("%s/%s", sm.ClusterFilePath, jobId)
	files, err := os.ReadDir(dir)
	if err != nil {
		hwlog.RunLog.Errorf("[FD-OL SLOWNODE]read the file list in dir: %s failed: %s", dir, err)
		return false
	}
	var filesMap = map[string]bool{}
	for _, file := range files {
		filesMap[file.Name()] = true
	}
	for _, nodeIP := range nodesIP {
		file := fmt.Sprintf("%s_parallel_group.json", nodeIP)
		if !filesMap[file] {
			return false
		}
	}
	return true
}

// clusterWatchingDataParse is watching the StartDataParseSign, if got the sign,
// request merging paralle group info to data parse
func clusterWatchingDataParse(slowNodeCtx *sm.SlowNodeContext) {
	logPrefix := fmt.Sprintf("[FD-OL SLOWNODE]job(name=%s, jobId=%s)", slowNodeCtx.Job.JobName, slowNodeCtx.Job.JobId)
	hwlog.RunLog.Infof("%s started watching start data parse(merge paralle group info) signal.", logPrefix)
	for {
		select {
		case <-slowNodeCtx.StartDataParseSign:
			hwlog.RunLog.Infof("%s received data parse(merge paralle group info) signal.", logPrefix)
			if err := startDataParse(slowNodeCtx); err != nil {
				hwlog.RunLog.Errorf("%s started data parse(merge paralle group info) failed: %v.", logPrefix, err)
				continue
			}
			hwlog.RunLog.Infof(
				"%s started data parse(merge paralle group info) successfully, exit signal watching process.",
				logPrefix)
			return
		case <-slowNodeCtx.StopChan:
			hwlog.RunLog.Infof("%s stopped, exit start data parse(merge paralle group info) signal watching process ",
				logPrefix)
			return
		}
	}
}
