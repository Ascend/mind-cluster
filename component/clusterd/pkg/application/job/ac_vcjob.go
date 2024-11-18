// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package job a series of job function
package job

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"huawei.com/npu-exporter/v6/common-utils/hwlog"
	apiCoreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"

	"clusterd/pkg/common/util"
)

// AddEvent is to handle adding new pod group
func (job *jobModel) AddEvent(agent *Agent) error {
	hwlog.RunLog.Infof("create worker for %s %s", job.Namespace, job.Name)
	bsKey := job.Uid
	if agent.BsExist(bsKey) {
		hwlog.RunLog.Errorf(" worker for %s %s is already existed", job.Namespace, job.Name)
		return nil
	}

	initCM(agent.KubeClientSet, job)

	cm, err := checkCMCreation(job.Namespace, job.Name, agent.KubeClientSet, agent.Config)
	if err != nil {
		return err
	}

	// retrieve configmap data
	jobStartString, ok := cm.Data[ConfigmapKey]
	if !ok {
		return fmt.Errorf("the key of " + ConfigmapKey + " does not exist")
	}
	var rst RankTableStatus
	if err = rst.UnmarshalToRankTable(jobStartString); err != nil {
		return err
	}
	hwlog.RunLog.Infof("jobStarting: %s", jobStartString)

	ranktable, replicasTotal, err := ranktableFactory(job, rst)
	if err != nil {
		return err
	}
	jobWorker := NewJobWorker(agent, job.Info, ranktable, replicasTotal)

	// start to report rank table build statistic for current job
	if agent.Config.DisplayStatistic {
		go jobWorker.Stat(BuildStatInterval)
	}

	// save current worker
	agent.SetBsWorker(bsKey, jobWorker)
	return nil
}

// EventUpdate : to handle job update event
func (job *jobModel) EventUpdate(agent *Agent) error {
	agent.RwMutex.RLock()
	_, exist := agent.BsWorker[job.Uid]
	agent.RwMutex.RUnlock()
	if !exist {
		// for job update, if create worker at job restart phase, the version will be incorrect
		hwlog.RunLog.Error("EventUpdate bsWorker does not exist")
		return job.AddEvent(agent)
	}
	return nil
}

// checkCMCreation check configmap
func checkCMCreation(namespace, name string, kubeClientSet kubernetes.Interface, config *Config) (
	*apiCoreV1.ConfigMap, error) {
	var cm *apiCoreV1.ConfigMap
	err := wait.PollImmediate(time.Duration(config.CmCheckTimeout)*time.Second,
		time.Duration(config.CmCheckTimeout)*time.Second,
		func() (bool, error) {
			var errTmp error
			cm, errTmp = kubeClientSet.CoreV1().ConfigMaps(namespace).
				Get(context.TODO(), fmt.Sprintf("%s-%s", ConfigmapPrefix, name), metav1.GetOptions{})
			if errTmp != nil {
				if errors.IsNotFound(errTmp) {
					return false, nil
				}
				return true, fmt.Errorf("get configmap error: %v", errTmp)
			}
			return true, nil
		})
	if err != nil {
		return nil, fmt.Errorf("failed to get configmap for job %s/%s: %v", namespace, name, err)
	}
	label910, exist := (*cm).Labels[Key910]
	if !exist || !(label910 == Val910B || label910 == Val910) {
		return nil, fmt.Errorf("invalid configmap label %s", label910)
	}

	return cm, nil
}

// DeleteWorker is to delete current worker
func (job *jobModel) DeleteWorker(namespace string, name string, uid string, agent *Agent) {
	agent.RwMutex.Lock()
	defer agent.RwMutex.Unlock()
	hwlog.RunLog.Infof("not exist + delete, current job is %s/%s/%s", namespace, name, uid)
	worker, exist := agent.BsWorker[uid]
	if !exist {
		hwlog.RunLog.Errorf("failed to delete worker for %s/%s, it's not exist", namespace, name)
		return
	}

	if agent.Config.DisplayStatistic {
		worker.CloseStat()
	}
	delete(agent.BsWorker, uid)
	hwlog.RunLog.Infof("worker for %s is deleted", uid)
	// delete configmap
	err := job.updateCMOnDeleteEvent(agent.KubeClientSet)
	if err != nil {
		hwlog.RunLog.Errorf("updateCMDelete error: %v", err)
	}
	return
}

// updateCMOnDeleteEvent handle cm update when pg delete
func (job *jobModel) updateCMOnDeleteEvent(kubeClientSet kubernetes.Interface) error {
	cm, err := kubeClientSet.CoreV1().ConfigMaps(job.Namespace).Get(context.TODO(),
		fmt.Sprintf("%s-%s", ConfigmapPrefix, job.Name), metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("get configmap error: %v", err)
	}
	changed := false
	if cm.Data[ConfigmapOperator] != OperatorDelete {
		cm.Data[ConfigmapOperator] = OperatorDelete
		changed = true
	}
	if cm.Data[DeleteTime] == "0" {
		cm.Data[DeleteTime] = getUnixTime2String()
		changed = true
	}
	if cm.Data[JobStatus] != StatusJobSucceed && cm.Data[JobStatus] != StatusJobFail {
		cm.Data[JobStatus] = StatusJobFail
		changed = true
	}
	if changed {
		if _, err = kubeClientSet.CoreV1().ConfigMaps(job.Namespace).Update(context.TODO(), cm,
			metav1.UpdateOptions{}); err != nil {
			return fmt.Errorf("update configmap delete status error: %v", err)
		}
	}
	cmData := map[string]string{
		DeleteTime:        getUnixTime2String(),
		JobStatus:         cm.Data[JobStatus],
		ConfigmapOperator: OperatorDelete,
	}
	return util.GetAndUpdateCmByTotalNum(cm.Data[cmCutNumKey], cm.Name, cm.Namespace, cmData, kubeClientSet)
}

func getUnixTime2String() string {
	timeNow64 := time.Now().Unix()
	timeNow := int(timeNow64)
	timeNowStr := strconv.Itoa(timeNow)
	return timeNowStr
}

// ranktableFactory : return the version type of ranktable according to your input parameters
func ranktableFactory(job *jobModel, rst RankTableStatus) (RankTabler, int32, error) {
	var ranktable RankTabler

	ranktable = &RankTable{ServerCount: "0", ServerList: []*ServerHccl(nil),
		RankTableStatus: RankTableStatus{Status: rst.Status}, Version: "1.0",
		UnHealthyDevice: make(map[string]string),
		UnHealthyNode:   make(map[string][]string)}

	return ranktable, job.replicas, nil
}

func getPGJobInfo(metaData metav1.Object) (string, string) {
	ownerReferences := metaData.GetOwnerReferences()
	var jobName, uid string
	for _, v := range ownerReferences {
		jobName = v.Name
		uid = string(v.UID)
		break
	}
	return jobName, uid
}
