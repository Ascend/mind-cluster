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

// Package cluster a series of function relevant to the fd-ol deployed in cluster
package cluster

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/core/context"
	"ascend-faultdiag-online/pkg/core/model/enum"
	"ascend-faultdiag-online/pkg/model"
	"ascend-faultdiag-online/pkg/model/slownode"
	"ascend-faultdiag-online/pkg/service/servicefunc/slownode/algo"
	"ascend-faultdiag-online/pkg/service/servicefunc/slownode/constants"
	"ascend-faultdiag-online/pkg/service/servicefunc/slownode/slownodejob"
	"ascend-faultdiag-online/pkg/utils"
	globalConstants "ascend-faultdiag-online/pkg/utils/constants"
	"ascend-faultdiag-online/pkg/utils/fileutils"
	"ascend-faultdiag-online/pkg/utils/grpc"
	"ascend-faultdiag-online/pkg/utils/k8s"
)

const (
	// slowNodeOn start slow node feature
	slowNodeOn = 1
	// slowNodeOff stop slow node feature
	slowNodeOff = 0

	rescheduleNamespace  = "mindx-dl"
	rescheduleCmName     = "job-reschedule-reason"
	rescheduleRecordsKey = "recent-reschedule-records"

	restartAdvice = "please restart slow node detection job"

	maxLoopCount = 10000
)

var jobSummaryWatcher = utils.NewStorage[string]()
var rescheduleWatcher = utils.NewStorage[string]()

type jobProcessor struct {
	ctx *slownodejob.JobContext
	job *slownode.Job
}

func (j *jobProcessor) logPrefix() string {
	if j.ctx != nil {
		return j.ctx.LogPrefix()
	}
	return fmt.Sprintf("[FD-OL SLOWNODE]job(name=%s, namespace=%s, jobId=%s)",
		j.job.JobName, j.job.Namespace, j.job.JobId)
}

func (j *jobProcessor) add() {
	// create slow node context and watch the job summary cm
	_, ok := slownodejob.GetJobCtxMap().Get(j.job.KeyGenerator())
	if ok {
		hwlog.RunLog.Warnf("%s has been existed in ctxMap, ignore it", j.logPrefix())
		return // already exists, no need to create a new one
	}
	ctx := slownodejob.NewJobContext(j.job, enum.Cluster)
	slownodejob.GetJobCtxMap().Insert(j.job.KeyGenerator(), ctx)
	// start to real-time watch the job-summary
	grpcClient, err := grpc.GetClient()
	if err != nil {
		hwlog.RunLog.Errorf("%s got grpc client failed: %v, %s", j.logPrefix(), err, restartAdvice)
		slownodejob.GetJobCtxMap().Delete(j.job.KeyGenerator())
		return
	}
	registerId, err := grpcClient.SubscribeJobSummary(ctx.Job.JobName, ctx.Job.Namespace, jobSummaryProcessor)
	if err != nil {
		hwlog.RunLog.Errorf("%s started to watch the job summary failed: %v, %s", j.logPrefix(), err, restartAdvice)
		slownodejob.GetJobCtxMap().Delete(j.job.KeyGenerator())
		return
	}
	jobSummaryWatcher.Store(j.job.KeyGenerator(), registerId)
}

func (j *jobProcessor) update() {
	// query slow node context by job name and namespace
	if ctx, ok := slownodejob.GetJobCtxMap().Get(j.job.KeyGenerator()); ok {
		j.job.Servers = ctx.Job.Servers
		ctx.Update(j.job)
		j.ctx = ctx
		if j.job.SlowNode == slowNodeOn {
			j.start()
		} else {
			j.stop()
		}
	} else {
		hwlog.RunLog.Infof("%s does not exist in ctxMap, create a new job", j.logPrefix())
		// create a new slow node context
		j.add()
	}
}

func (j *jobProcessor) delete() {
	// query all the jobs by name and namespace
	ctx, ok := slownodejob.GetJobCtxMap().Get(j.job.KeyGenerator())
	if !ok {
		hwlog.RunLog.Warnf("[FD-OL SLOWNODE]job(name=%s, namespace=%s) does not exist in context, ignore it",
			j.job.JobName, j.job.Namespace)
		return
	}
	j.ctx = ctx
	j.stop()
	grpcClient, err := grpc.GetClient()
	if err != nil {
		hwlog.RunLog.Errorf("%s got grpc client failed: %v", j.logPrefix(), err)
		slownodejob.GetJobCtxMap().Delete(j.job.KeyGenerator())
		return
	}
	if registerId, ok := jobSummaryWatcher.Load(j.job.KeyGenerator()); !ok {
		hwlog.RunLog.Warnf("%s could not got job summary watcher id", j.logPrefix())
	} else {
		grpcClient.UnsubscribeJobSummary(registerId)
		jobSummaryWatcher.Delete(j.job.KeyGenerator())
	}
	slownodejob.GetJobCtxMap().Delete(j.job.KeyGenerator())
	hwlog.RunLog.Infof("%s delete job successfully", j.logPrefix())
}

func (j *jobProcessor) start() {
	if j.ctx == nil {
		return
	}
	if j.ctx.TrainingJobStatus != enum.IsRunning {
		hwlog.RunLog.Warnf("%s started failed: training job status(%s) is not: %s", j.logPrefix(),
			j.ctx.TrainingJobStatus, enum.IsRunning)
		return
	}
	if j.job.SlowNode == slowNodeOff {
		hwlog.RunLog.Warnf("%s SlowNode is %d, no need to start", j.logPrefix(), slowNodeOff)
		return
	}
	if err := j.ctx.Start(); err != nil {
		hwlog.RunLog.Errorf("%s started failed: %v", j.logPrefix(), err)
		return
	}
	// clear local data & delete cm, ensure the data will not affect the new detection
	j.removeData()
	j.deleteCM()
	// start all profiling only depends on the training job is ready
	if err := j.loopStartAllProfiling(); err != nil {
		j.ctx.Stop()
		return
	}
	if err := j.createOrUpdateCM(); err != nil {
		hwlog.RunLog.Errorf("%s created or updated cm failed: %v, %s", j.logPrefix(), err, restartAdvice)
		j.ctx.Stop()
		j.ctx.StopAllProfiling()
		return
	}
	j.startRescheduleWatcher()
	j.waitNodeReport()
}

func (j *jobProcessor) loopStartAllProfiling() error {
	var ticker = time.NewTicker(time.Second)
	defer ticker.Stop()
	for i := 0; i < maxLoopCount; i++ {
		select {
		case _, ok := <-j.ctx.StopChan:
			if !ok {
				hwlog.RunLog.Infof("%s stopped, exiting start all profiling process", j.logPrefix())
				return errors.New("job stopped")
			}
		case <-ticker.C:
			// start all profiling per second
			err := j.ctx.StartAllProfiling()
			if err == nil {
				return nil
			}
		}
	}
	hwlog.RunLog.Errorf("%s started all profiling failed duo to reached the max retry limitation: %d, %s",
		j.logPrefix(), maxLoopCount, restartAdvice)
	return fmt.Errorf("reached the max retry limitation: %d", maxLoopCount)
}

func (j *jobProcessor) stop() {
	if j.ctx == nil {
		return
	}
	if !j.ctx.IsRunning() {
		hwlog.RunLog.Warnf("%s do not need to stop: not running", j.logPrefix())
		return
	}
	j.ctx.Stop()
	j.ctx.RemoveAllCM()
	if j.ctx.TrainingJobStatus != enum.IsCompleted {
		// training job is complete, operate the profiling will cause error
		j.ctx.StopAllProfiling()
	}
	algo.NewController(j.ctx).Stop()
	j.stopRescheduleWatcher()
	rescheduleWatcher.Delete(j.job.KeyGenerator())
	jobOnceMap.Delete(j.ctx.Job.JobId)
	j.removeData()
}

func (j *jobProcessor) createOrUpdateCM() error {
	if j.ctx == nil {
		return errors.New("createOrUpdateCM failed: ctx is nil")
	}
	dataBytes, err := json.MarshalIndent(j.ctx.Job, "", "  ")
	if err != nil {
		return err
	}
	cmName := constants.NodeJobPrefix + "-" + j.ctx.Job.JobId
	j.ctx.AllCMNames.Store(cmName, struct{}{})
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cmName,
			Namespace: j.job.Namespace,
			Labels:    map[string]string{constants.CmConsumer: constants.CmConsumerValue},
		},
		Data: map[string]string{
			constants.NodeJobCMKey: string(dataBytes),
		},
	}
	k8sClient, err := k8s.GetClient()
	if err != nil {
		return err
	}
	if err := k8sClient.CreateOrUpdateConfigMap(cm); err != nil {
		return err
	}
	hwlog.RunLog.Infof("%s create or update configmap: [key: %s, value: %s] successfully",
		j.logPrefix(), cm.Name, cm.Data)
	return nil
}

// removeData all the data producted by this job
func (j *jobProcessor) removeData() {
	if j.ctx == nil || j.ctx.Job == nil {
		return
	}
	dir := filepath.Join(constants.ClusterFilePath, j.ctx.Job.JobId)
	if j.ctx.Job.JobId == "" {
		hwlog.RunLog.Warnf("%s remove dir: %s, jobId is empty, skip", j.logPrefix(), dir)
		return
	}
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		hwlog.RunLog.Infof("%s remove dir: %s, dir is not existed, skip", j.logPrefix(), dir)
		return
	}
	absDir, err := fileutils.CheckPath(dir)
	if err != nil {
		hwlog.RunLog.Warnf("%s remove dir: %s failed: %v", j.logPrefix(), dir, err)
		return
	}
	if err := os.RemoveAll(absDir); err != nil {
		hwlog.RunLog.Errorf("%s remove dir: %s failed: %s", j.logPrefix(), absDir, err)
	} else {
		hwlog.RunLog.Infof("%s remove dir: %s successfully", j.logPrefix(), absDir)
	}
}

func (j *jobProcessor) deleteCM() {
	if j.ctx == nil {
		hwlog.RunLog.Errorf("%s deleted cm failed: ctx is nil", j.logPrefix())
		return
	}
	k8sClient, err := k8s.GetClient()
	if err != nil {
		hwlog.RunLog.Errorf("%s got k8s client failed: %v", j.logPrefix(), err)
		return
	}
	cmName := constants.NodeJobPrefix + "-" + j.ctx.Job.JobId
	if err := k8sClient.DeleteConfigMap(cmName, j.ctx.Job.Namespace); err != nil {
		hwlog.RunLog.Errorf("%s deleted cm: %v  failed: %v", j.logPrefix(), cmName, err)
		return
	}
	hwlog.RunLog.Infof("%s deleted cm: %v successfully", j.logPrefix(), cmName)
}

// waitNodeReport wait goroutine, timeout or there is node report data parse result
func (j *jobProcessor) waitNodeReport() {
	go func() {
		hwlog.RunLog.Infof("%s started to wait the nodes report, timeout: %ds", j.logPrefix(),
			context.FdCtx.Config.NodeReportTimeout)
		select {
		case <-time.After(time.Duration(context.FdCtx.Config.NodeReportTimeout) * time.Second):
			hwlog.RunLog.Infof("%s no node report util timeout: %d, stop slow node detection", j.logPrefix(),
				context.FdCtx.Config.NodeReportTimeout)
			j.stop()
		case <-j.ctx.NodeReportSignal:
			hwlog.RunLog.Infof("%s detected node report, exit wait node report process", j.logPrefix())
			return
		case <-j.ctx.StopChan:
			hwlog.RunLog.Infof("%s job stopped, exit wait node report process", j.logPrefix())
			return
		}
	}()
}

func (j *jobProcessor) startRescheduleWatcher() {
	var cmWatcher = k8s.GetCmWatcher()
	registerId := cmWatcher.Subscribe(rescheduleNamespace, rescheduleCmName, j.rescheduleProcessor)
	rescheduleWatcher.Store(j.job.KeyGenerator(), registerId)
}

func (j *jobProcessor) rescheduleProcessor(oldCm, newCm *corev1.ConfigMap, op watch.EventType) {
	if op != watch.Added && op != watch.Modified {
		return
	}
	hwlog.RunLog.Infof("%v got reschedule data: %v", j.logPrefix(), newCm)
	// the format of rescheduleRecordsKey, refer to:
	// https://www.hiascend.com/document/detail/zh/mindcluster/71RC1/clustersched/dlug/dl_resume_060.html
	// convert configMap to map[string]any
	var data map[string]model.RescheduleData
	records, exists := newCm.Data[rescheduleRecordsKey]
	if !exists || records == "" {
		hwlog.RunLog.Warnf("%v %v not in cm data: %+v or records is empty",
			j.logPrefix(), rescheduleRecordsKey, newCm.Data)
		return
	}
	if err := json.Unmarshal([]byte(records), &data); err != nil {
		hwlog.RunLog.Errorf("%v convert reschedule-records: %s to map[string]any failed: %v",
			j.logPrefix(), records, err)
		return
	}
	for _, rescheduleData := range data {
		if strings.HasSuffix(rescheduleData.JobId, j.ctx.Job.JobId) {
			if rescheduleData.TotalRescheduleTimes != j.ctx.GetRescheduleCount() {
				hwlog.RunLog.Infof("%v detected the TotalRescheduleTimes: %d changed(local count: %d), "+
					"stop and start slow node detection",
					j.logPrefix(), rescheduleData.TotalRescheduleTimes, j.ctx.GetRescheduleCount())
				j.ctx.SetRescheduleCount(rescheduleData.TotalRescheduleTimes)
				j.stop()
				j.start()
			}
			return
		}
	}
}

func (j *jobProcessor) stopRescheduleWatcher() {
	registerId, exists := rescheduleWatcher.Load(j.job.KeyGenerator())
	if !exists {
		return
	}
	var cmWatcher = k8s.GetCmWatcher()
	cmWatcher.Unsubscribe(rescheduleNamespace, rescheduleCmName, registerId)
}

// JobProcessor store the slow node feat config into the confMap in cluster
func JobProcessor(oldData, newData *slownode.Job, operator watch.EventType) {
	hwlog.RunLog.Infof("[FD-OL SLOWNODE]got job cm data, operator: %s, newData: %+v, oldData: %+v",
		operator, newData, oldData)

	if newData.JobName == "" {
		hwlog.RunLog.Errorf("[FD-OL SLOWNODE]job(name=%s): jobName is empty, ignore it", newData.JobName)
		return
	}

	jp := jobProcessor{job: newData}

	switch operator {
	case watch.Added:
		jp.add()
	case watch.Modified:
		jp.update()
	case watch.Deleted:
		jp.delete()
	default:
		return
	}
}

// JobRestartProcessor got the node restart config map, loop the context and restart the correspond job
func JobRestartProcessor(oldNodeIp, newNodeIp *string, operator watch.EventType) {
	if operator != watch.Added && operator != watch.Modified {
		return
	}
	hwlog.RunLog.Infof("[FD-OL SLOWNODE]got node: %s restarted info", *newNodeIp)
	// loop local ctx, found the correspond job
	ctxList := slownodejob.GetJobCtxMap().GetByNodeIp(*newNodeIp)
	if len(ctxList) == 0 {
		hwlog.RunLog.Infof("[FD-OL SLOWNODE]no job needs to be restarted: no job ctx found")
		return
	}
	for _, ctx := range ctxList {
		go func(ctx *slownodejob.JobContext) {
			if ctx == nil || ctx.Job == nil {
				return
			}
			jp := jobProcessor{ctx: ctx, job: ctx.Job}
			hwlog.RunLog.Infof("%s needed to restart(stop first and start)", jp.logPrefix())
			jp.stop()
			// wait the restart interval time
			time.Sleep(globalConstants.RestartInterval * time.Millisecond)
			jp.start()
		}(ctx)
	}
}
