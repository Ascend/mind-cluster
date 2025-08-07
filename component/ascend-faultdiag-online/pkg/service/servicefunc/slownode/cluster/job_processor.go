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
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/core/context"
	"ascend-faultdiag-online/pkg/core/model/enum"
	"ascend-faultdiag-online/pkg/model/slownode"
	"ascend-faultdiag-online/pkg/service/servicefunc/slownode/algo"
	"ascend-faultdiag-online/pkg/service/servicefunc/slownode/constants"
	"ascend-faultdiag-online/pkg/service/servicefunc/slownode/slownodejob"
	globalConstants "ascend-faultdiag-online/pkg/utils/constants"
	"ascend-faultdiag-online/pkg/utils/k8s"
)

const (
	// slowNodeOn start slow node feature
	slowNodeOn = 1
	// slowNodeOff stop slow node feature
	slowNodeOff = 0

	jobSummaryPrefix = "job-summary-"
)

type jobProcessor struct {
	ctx *slownodejob.JobContext
	job *slownode.Job
}

func (j *jobProcessor) logPrefix() string {
	if j.ctx != nil {
		return fmt.Sprintf("[FD-OL SLOWNODE]job(name=%s, namespace=%s, jobId=%s)",
			j.ctx.Job.JobName, j.ctx.Job.Namespace, j.ctx.Job.JobId)
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
	// start to watch the job summary cm
	cmName := jobSummaryPrefix + j.job.JobName
	ctx := slownodejob.NewJobContext(j.job, enum.Cluster)
	slownodejob.GetJobCtxMap().Insert(j.job.KeyGenerator(), ctx)
	GetConfigMapWatcher().WatchConfigMap(j.job.Namespace, cmName)
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
	cmName := jobSummaryPrefix + j.job.JobName
	GetConfigMapWatcher().StopWatching(j.job.Namespace, cmName)
	var key = fmt.Sprintf("%s/%s", j.job.Namespace, j.job.JobName)
	slownodejob.GetJobCtxMap().Delete(key)
}

func (j *jobProcessor) start() {
	if j.ctx == nil {
		return
	}
	if j.ctx.IsRunning() {
		hwlog.RunLog.Warnf("%s started failed: already running", j.logPrefix())
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
	// clear local data & delete cm, ensure the data will not affect the new detection
	j.removeData()
	j.deleteCM()
	j.ctx.Start()
	if err := j.createOrUpdateCM(); err != nil {
		hwlog.RunLog.Errorf("%s created or updated cm feaild: %v", j.logPrefix(), err)
		return
	}
	j.ctx.StartAllProfiling()
	j.waitNodeReport()
}

func (j *jobProcessor) stop() {
	if j.ctx == nil {
		return
	}
	if !j.ctx.IsRunning() {
		hwlog.RunLog.Warnf("%s stopped failed: not running", j.logPrefix())
		return
	}
	j.ctx.RemoveAllCM()
	j.ctx.StopAllProfiling()
	algo.NewController(j.ctx).Stop()
	j.ctx.Stop()
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
	if j.ctx == nil {
		return
	}
	dir := filepath.Join(constants.ClusterFilePath, j.ctx.Job.JobId)
	if j.job.JobId == "" {
		hwlog.RunLog.Warnf("%s remove dir: %s, jobId is empty, skip", j.logPrefix(), dir)
		return
	}
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		hwlog.RunLog.Infof("%s remove dir: %s, dir is not existed, skip", j.logPrefix(), dir)
		return
	}
	if err := os.RemoveAll(dir); err != nil {
		hwlog.RunLog.Errorf("%s remove dir: %s failed: %s", j.logPrefix(), dir, err)
	} else {
		hwlog.RunLog.Infof("%s remove dir: %s successfully", j.logPrefix(), dir)
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
			jp := jobProcessor{ctx: ctx, job: ctx.Job}
			hwlog.RunLog.Infof("%s needed to restart(stop first and start)", jp.logPrefix())
			jp.stop()
			// wait the restart interval time
			time.Sleep(globalConstants.RestartInterval * time.Millisecond)
			jp.start()
		}(ctx)
	}
}
