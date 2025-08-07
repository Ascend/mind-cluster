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

// Package node a series of function relevant to the fd-ol deployed in node
package node

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/core/model/enum"
	"ascend-faultdiag-online/pkg/model/slownode"
	"ascend-faultdiag-online/pkg/service/servicefunc/slownode/algo"
	"ascend-faultdiag-online/pkg/service/servicefunc/slownode/common"
	"ascend-faultdiag-online/pkg/service/servicefunc/slownode/constants"
	"ascend-faultdiag-online/pkg/service/servicefunc/slownode/dataparse"
	"ascend-faultdiag-online/pkg/service/servicefunc/slownode/slownodejob"
	"ascend-faultdiag-online/pkg/utils"
	"ascend-faultdiag-online/pkg/utils/k8s"
)

type jobProcessor struct {
	ctx          *slownodejob.JobContext
	job          *slownode.Job
	nodeIp       string
	available    bool
	availableIps []string
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
	if !j.available {
		hwlog.RunLog.Infof(
			"%s ip: %s not in availabe ip list: %v, ignore it", j.logPrefix(), j.nodeIp, j.availableIps)
		return
	}
	_, ok := slownodejob.GetJobCtxMap().Get(j.job.JobId)
	if ok {
		hwlog.RunLog.Warnf("%s has been existed in context, ignore it", j.logPrefix())
		return
	}
	if utils.IsRestarted() {
		hwlog.RunLog.Warnf("%s detected pod has been restarted, restart job", j.logPrefix())
		// send cm to cluster
		j.sendRestartConfigMap()
		return
	}
	j.ctx = slownodejob.NewJobContext(j.job, enum.Node)
	slownodejob.GetJobCtxMap().Insert(j.job.JobId, j.ctx)
	j.start()
}

func (j *jobProcessor) update() {
	var ok bool
	j.ctx, ok = slownodejob.GetJobCtxMap().Get(j.job.JobId)
	if ok {
		// found ctx but not sn not in sn list: stop & delete
		if !j.available {
			hwlog.RunLog.Warnf("%s found ctx in ctxMap, but node ip: %s not in available ip list: %v, stop job",
				j.logPrefix(), j.nodeIp, j.availableIps)
			j.delete()
			return
		}
		// found ctx, available，rankIds changed: stop & start
		if !common.AreServersEqual(j.ctx.Job.Servers, j.job.Servers) {
			hwlog.RunLog.Warnf("%s found rankIds changes reload job", j.logPrefix())
			j.ctx.Update(j.job)
			j.stop()
			j.start()
			return
		}
	} else {
		// no ctx in ctx map, call add function
		hwlog.RunLog.Errorf("%s does not exist in ctxMap, create a new one", j.logPrefix())
		j.add()
	}
}

func (j *jobProcessor) delete() {
	ctx, ok := slownodejob.GetJobCtxMap().Get(j.job.JobId)
	if !ok {
		hwlog.RunLog.Warnf("%s does not exist in context, ignore it", j.logPrefix())
		return
	}
	j.ctx = ctx
	j.stop()
	ctx.RemoveAllCM()
	slownodejob.GetJobCtxMap().Delete(j.job.JobId)
}

func (j *jobProcessor) start() {
	if j.ctx == nil {
		return
	}
	if j.ctx.IsRunning() {
		hwlog.RunLog.Errorf("%s started failed: already running", j.logPrefix())
		return
	}

	j.ctx.Start()
	dataparse.NewController(j.ctx).Start()
	j.ctx.AddStep()
}

func (j *jobProcessor) stop() {
	if j.ctx == nil {
		return
	}
	if !j.ctx.IsRunning() {
		hwlog.RunLog.Errorf("%s stopped failed: not running", j.logPrefix())
		return
	}
	dataparse.NewController(j.ctx).Stop()
	algo.NewController(j.ctx).Stop()
	j.ctx.Stop()
}

func (j *jobProcessor) sendRestartConfigMap() {
	nodeIp, err := utils.GetNodeIp()
	if err != nil {
		hwlog.RunLog.Errorf("%s got node ip failed: %v", j.logPrefix(), err)
		return
	}
	var cmName = fmt.Sprintf("%s-%s", constants.NodeRestartInfoPrefix, nodeIp)
	// send to cm
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cmName,
			Namespace: j.job.Namespace,
			Labels:    map[string]string{constants.CmConsumer: constants.CmConsumerValue},
		},
		Data: map[string]string{
			constants.NodeRestartInfoCMKey: nodeIp,
		},
	}
	k8sClient, err := k8s.GetClient()
	if err != nil {
		hwlog.RunLog.Errorf("%s create k8s client failed: %v", j.logPrefix(), err)
		return
	}
	if err := k8sClient.CreateOrUpdateConfigMap(cm); err != nil {
		hwlog.RunLog.Errorf("%s create or update config map failed: %v", j.logPrefix(), err)
		return
	}
	// delete cm
	if err = k8sClient.DeleteConfigMap(cmName, j.job.Namespace); err != nil {
		hwlog.RunLog.Errorf("%s delete config map failed: %v", j.logPrefix(), err)
	} else {
		hwlog.RunLog.Errorf("%s create then delete config map: %s successfully", j.logPrefix(), cmName)
	}
}

// JobProcessor store the slow node job into the confMap in node
func JobProcessor(oldData, newData *slownode.Job, operator watch.EventType) {
	hwlog.RunLog.Infof("[FD-OL SLOWNODE]got job cm data, operator: %s, newObj: %+v, oldObj: %+v",
		operator, newData, oldData)
	if newData.JobId == "" {
		hwlog.RunLog.Infof("[FD-OL SLOWNODE]job(name=%s, jobId=%s)'s jobId is empty, ignore it",
			newData.JobName, newData.JobId)
		return
	}

	j := jobProcessor{job: newData}
	if operator == watch.Added || operator == watch.Modified {
		var err error
		j.nodeIp, err = utils.GetNodeIp()
		if err != nil {
			hwlog.RunLog.Infof("%v queried node ip failed: %s", j.logPrefix(), err)
			return
		}
		j.availableIps = make([]string, len(newData.Servers))
		j.available = false
		for i, server := range newData.Servers {
			if j.nodeIp == server.Ip {
				newData.RankIds = server.RankIds
				j.available = true
			}
			j.availableIps[i] = server.Ip
		}
		hwlog.RunLog.Infof("%s node ip is: %s, availableIps: %s", j.logPrefix(), j.nodeIp, j.availableIps)
	}

	switch operator {
	case watch.Deleted:
		j.delete()
	case watch.Added:
		j.add()
	case watch.Modified:
		j.update()
	default:
		hwlog.RunLog.Infof("%v unsupported operator: %v", j.logPrefix(), operator)
		return
	}
}
