/* Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package schedulingexception is for collecting scheduling exception

package schedulingexception

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"volcano.sh/apis/pkg/apis/batch/v1alpha1"
	"volcano.sh/apis/pkg/apis/scheduling/v1beta1"

	batchv1 "ascend-common/api/ascend-operator/apis/batch/v1"
	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/job"
	"clusterd/pkg/domain/pod"
	"clusterd/pkg/interface/kube"
)

var (
	collector    *Collector
	newCollector sync.Once
)

type Config struct {
	CheckInterval int64
}

// CheckSchedulingException is for checking scheduling exception
func CheckSchedulingException(ctx context.Context, config *Config) {
	newCollector.Do(func() {
		if config == nil {
			hwlog.RunLog.Error("scheduling exception collector config is nil")
			return
		}
		if config.CheckInterval <= 0 {
			hwlog.RunLog.Error("scheduling exception collector check interval is invalid")
			config.CheckInterval = defaultCheckInterval
		}
		collector = &Collector{
			jobExceptions: map[string]*jobExceptionInfo{},
			checkInterval: config.CheckInterval,
		}
	})
	if collector == nil {
		hwlog.RunLog.Error("scheduling exception collector is nil")
		return
	}
	go collector.Start(ctx)
}

// Collector is for collecting scheduling exception
type Collector struct {
	isRunning     bool
	checkInterval int64
	jobExceptions map[string]*jobExceptionInfo
}

type exceptionReport struct {
	JobExceptions map[string]*jobExceptionInfo `json:"jobs"`
}

// Start is for starting scheduling exception collector
func (c *Collector) Start(ctx context.Context) {
	if c.isRunning {
		hwlog.RunLog.Info("scheduling exception collector is already running")
		return
	}
	c.isRunning = true

	defer func() {
		c.isRunning = false
	}()

	ticker := time.NewTicker(time.Duration(c.checkInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			hwlog.RunLog.Info("scheduling exception collector context done")
			return
		case <-ticker.C:
			c.checkJobs()
		}
	}
}

func (c *Collector) checkJobs() {
	allJobs := job.GetAllJobCache()
	exceptionJobs := make(map[string]*jobExceptionInfo, 0)

	for jobKey, jobInfo := range allJobs {
		cond := c.processPodGroup(jobKey, jobInfo)
		if cond == nil {
			continue
		}
		if _, exist := c.jobExceptions[jobKey]; !exist {
			hwlog.RunLog.Infof("job %s has scheduling exception", jobKey)
			c.jobExceptions[jobKey] = &jobExceptionInfo{
				JobName:   jobInfo.Name,
				JobType:   jobInfo.JobType,
				NameSpace: jobInfo.NameSpace,
				Condition: *cond,
			}
		}

		hwlog.RunLog.Infof("pgname: %v, cond: %v, pre cond: %v", jobInfo.PgName, cond, c.jobExceptions[jobKey].Condition)
		key := jobInfo.Name + "." + jobKey
		if cond.Equal(c.jobExceptions[jobKey].Condition) {
			exceptionJobs[key] = c.jobExceptions[jobKey]
		}
		c.jobExceptions[jobKey].Condition = *cond
	}

	allMetaObjs := c.processJobs(exceptionJobs, allJobs)
	c.cleanupJobs(allJobs, allMetaObjs)

	report := exceptionReport{
		JobExceptions: exceptionJobs,
	}
	if err := updateConfigMap(report); err != nil {
		hwlog.RunLog.Errorf("update scheduling exception configmap failed: %v", err)
		return
	}
	hwlog.RunLog.Infof("updated scheduling exception configmap with %d jobs", len(exceptionJobs))
}

func (c *Collector) processPodGroup(jobKey string, jobInfo constant.JobInfo) *conditionDetail {
	if jobInfo.Status == job.StatusJobRunning || jobInfo.Status == job.StatusJobCompleted {
		delete(c.jobExceptions, jobKey)
		return nil
	}

	obj, err := kube.GetObject(kube.PodGroupGVK(), fmt.Sprintf("%s/%s", jobInfo.NameSpace, jobInfo.PgName))
	if err != nil {
		hwlog.RunLog.Warnf("get podgroup %s failed: %v", jobInfo.PgName, err)
		return nil
	}

	pg, ok := obj.(*v1beta1.PodGroup)
	if !ok {
		hwlog.RunLog.Warnf("convert object to PodGroup failed for job %s", jobKey)
		return nil
	}

	pods := pod.GetPodByJobId(jobKey)
	indices := c.analyzePodGroupConditions(pg)
	cond := c.analyzePodGroupPhase(pg, pods, indices, jobInfo)
	return cond
}

func (c *Collector) analyzePodGroupConditions(pg *v1beta1.PodGroup) conditionIndices {
	indices := conditionIndices{
		jobEnqueueFailedIndex:     invalidIndex,
		jobValidFailedIndex:       invalidIndex,
		predicatedNodesErrorIndex: invalidIndex,
		batchOrderFailedIndex:     invalidIndex,
		notEnoughResourcesIndex:   invalidIndex,
	}

	for index, condition := range pg.Status.Conditions {
		if condition.Type != v1beta1.PodGroupUnschedulableType || condition.Status != corev1.ConditionTrue {
			continue
		}

		switch condition.Reason {
		case jobEnqueueFailedReason:
			indices.jobEnqueueFailedIndex = index
		case jobValidateFailedReason:
			indices.jobValidFailedIndex = index
		case nodePredicateFailedReason:
			indices.predicatedNodesErrorIndex = index
		case batchOrderFailedReason:
			indices.batchOrderFailedIndex = index
		case notEnoughResourcesReason:
			indices.notEnoughResourcesIndex = index
		default:
			hwlog.RunLog.Warnf("unknown condition reason: %s", condition.Reason)
		}
	}

	hwlog.RunLog.Infof("pg phase: %s, conditionIndices %#v", pg.Status.Phase, indices)
	return indices
}

func (c *Collector) analyzePodGroupPhase(pg *v1beta1.PodGroup, pods map[string]corev1.Pod, indices conditionIndices,
	jobInfo constant.JobInfo) *conditionDetail {
	switch pg.Status.Phase {
	case "":
		return c.processPodGroupCreated()
	case v1beta1.PodGroupUnknown:
		return c.processPodGroupUnknown(pods)
	case v1beta1.PodGroupInqueue:
		return c.processPodGroupInqueue(pg, pods, indices, jobInfo)
	case v1beta1.PodGroupPending:
		return c.processPodGroupPending(pg, indices)
	case v1beta1.PodGroupRunning:
		return c.processPodGroupRunning(pods)
	default:
		return nil
	}
}

func (c *Collector) processPodGroupCreated() *conditionDetail {
	return &conditionDetail{
		Status: podGroupCreated,
		Reason: "PgNotInitialized",
		Message: "pg phase is empty, volcano-scheduler maybe not running, you can check the status by" +
			"executing command: kubectl get pod -n volcano-system -l app=volcano-scheduler",
	}
}

func (c *Collector) processPodGroupUnknown(pods map[string]corev1.Pod) *conditionDetail {
	for _, p := range pods {
		if p.Status.Phase == corev1.PodPending {
			return &conditionDetail{
				Status: podGroupUnknown,
				Reason: "PodPending",
				Message: fmt.Sprintf("pod is pending, please check the pod status, "+
					"executing command: kubectl describe pod -n %s %s", p.Namespace, p.Name),
			}
		}
		if p.Status.Phase == corev1.PodFailed {
			return &conditionDetail{
				Status:  podGroupUnknown,
				Reason:  "PodFailed",
				Message: fmt.Sprintf("pod is failed, please check the pod status, executing command: kubectl describe pod -n %s %s", p.Namespace, p.Name),
			}
		}
	}
	return nil
}

func (c *Collector) processPodGroupInqueue(pg *v1beta1.PodGroup, pods map[string]corev1.Pod, indices conditionIndices,
	jobInfo constant.JobInfo) *conditionDetail {
	if indices.batchOrderFailedIndex >= 0 {
		cond := &conditionDetail{
			Status:  podGroupInqueue,
			Reason:  pg.Status.Conditions[indices.batchOrderFailedIndex].Reason,
			Message: pg.Status.Conditions[indices.batchOrderFailedIndex].Message,
		}
		if indices.predicatedNodesErrorIndex >= 0 {
			cond.Message += fmt.Sprintf("; %s", pg.Status.Conditions[indices.predicatedNodesErrorIndex].Message)
		}
		return cond
	}

	if indices.predicatedNodesErrorIndex >= 0 {
		return &conditionDetail{
			Status:  podGroupInqueue,
			Reason:  pg.Status.Conditions[indices.predicatedNodesErrorIndex].Reason,
			Message: pg.Status.Conditions[indices.predicatedNodesErrorIndex].Message,
		}
	}

	if indices.jobValidFailedIndex >= 0 {
		return &conditionDetail{
			Status:  podGroupInqueue,
			Reason:  pg.Status.Conditions[indices.jobValidFailedIndex].Reason,
			Message: pg.Status.Conditions[indices.jobValidFailedIndex].Message,
		}
	}

	if indices.notEnoughResourcesIndex >= 0 {
		return c.processNotEnoughResources(pg, pods, indices, jobInfo)
	}

	return nil
}

func (c *Collector) processNotEnoughResources(pg *v1beta1.PodGroup, pods map[string]corev1.Pod,
	indices conditionIndices, jobInfo constant.JobInfo) *conditionDetail {
	msg := pg.Status.Conditions[indices.notEnoughResourcesIndex].Message
	if len(pods) < int(pg.Spec.MinMember) {
		msg += fmt.Sprintf(" the number of pods is less than minMember, "+
			"you can check the job status by excuting command: kubectl describe [job-type, "+
			"eg vcjob,acjob] -n %s %s",
			jobInfo.NameSpace, jobInfo.Name)
	}
	return &conditionDetail{
		Status:  podGroupInqueue,
		Reason:  pg.Status.Conditions[indices.notEnoughResourcesIndex].Reason,
		Message: msg,
	}
}

func (c *Collector) processPodGroupPending(pg *v1beta1.PodGroup, indices conditionIndices) *conditionDetail {
	if indices.jobEnqueueFailedIndex >= 0 {
		return &conditionDetail{
			Status:  podGroupPending,
			Reason:  pg.Status.Conditions[indices.jobEnqueueFailedIndex].Reason,
			Message: pg.Status.Conditions[indices.jobEnqueueFailedIndex].Message,
		}
	}

	return &conditionDetail{
		Status: podGroupPending,
		Reason: jobEnqueueFailedReason,
		Message: fmt.Sprintf("the resources such as cpu, memory is not enough in Queue, "+
			"you can check them by executing command: kubectl describe q %s, "+
			"or view the log of volcano-scheduler", pg.Spec.Queue),
	}
}

func (c *Collector) processPodGroupRunning(pods map[string]corev1.Pod) *conditionDetail {
	for _, p := range pods {
		if p.Status.Phase == corev1.PodFailed {
			return c.processPodFailed(p)
		}
		if p.Status.Phase == corev1.PodPending {
			return &conditionDetail{
				Status: podGroupRunning,
				Reason: "PodPending",
				Message: fmt.Sprintf("pod is pending, "+
					"please check the pod status using the following command: kubectl describe pod -n %s"+
					" %s", p.Namespace, p.Name),
			}
		}
	}
	return nil
}

func (c *Collector) processPodFailed(p corev1.Pod) *conditionDetail {
	if p.Status.Reason == "UnexpectedAdmissionError" && strings.Contains(p.Status.Message, "not get valid pod") {
		return &conditionDetail{
			Status: podGroupRunning,
			Reason: "PodFailed",
			Message: "The pod is not valid for allocation. " +
				"Please check whether the annotation huawei." +
				"com/Ascend910=<some-value> is present in the pod. If it is missing, " +
				"inspect the Volcano scheduler logs for the keyword `Failed to get plugin volcano" +
				"-npu_***" +
				"` using the following command: kubectl logs -f -n volcano-system -l app=volcano" +
				"-scheduler, if exist, " +
				"check the config in configmap volcano-scheduler-configmap and the plugin " +
				" in the dir of `\\plugins` of the container",
		}
	}

	return &conditionDetail{
		Status: podGroupRunning,
		Reason: "PodFailed",
		Message: fmt.Sprintf("pod is failed, "+
			"please check the pod status using the following command: kubectl describe pod -n %s"+
			" %s", p.Namespace, p.Name),
	}
}

func (c *Collector) processJobs(exceptionJobs map[string]*jobExceptionInfo, pgInfos map[string]constant.JobInfo) map[string]metav1.Object {
	allMetaObjs := make(map[string]metav1.Object, 0)
	for gvk, jobs := range kube.ListObjects(kube.AcJobGVK(), kube.VcJobGVK()) {
		for _, jobObj := range jobs {
			metaObj, ok := jobObj.(metav1.Object)
			if !ok {
				continue
			}
			if _, exist := pgInfos[string(metaObj.GetUID())]; exist {
				continue
			}

			allMetaObjs[string(metaObj.GetUID())] = metaObj
			hwlog.RunLog.Infof("job %s/%s is %s", metaObj.GetNamespace(), metaObj.GetName(), metaObj.GetUID())

			cond := c.processJobObject(jobObj, metaObj)
			if cond == nil {
				continue
			}
			jobKey := string(metaObj.GetUID())
			if _, exist := c.jobExceptions[jobKey]; !exist {
				c.jobExceptions[jobKey] = &jobExceptionInfo{
					NameSpace: metaObj.GetNamespace(),
					JobName:   metaObj.GetName(),
					JobType:   gvk.String(),
					Condition: *cond,
				}
			}

			key := metaObj.GetName() + "." + jobKey
			if cond.Equal(c.jobExceptions[jobKey].Condition) {
				exceptionJobs[key] = c.jobExceptions[jobKey]
			}

			c.jobExceptions[jobKey].Condition = *cond
		}
	}

	return allMetaObjs
}

func (c *Collector) processJobObject(jobObj interface{}, metaObj metav1.Object) *conditionDetail {
	var cond *conditionDetail
	switch j := jobObj.(type) {
	case *v1alpha1.Job:
		cond = c.processVcJob(j)
	case *batchv1.AscendJob:
		cond = c.processAscendJob(j)
	default:
		hwlog.RunLog.Warn("unknown job type")
		return nil
	}

	return cond
}

func (c *Collector) processVcJob(vcJob *v1alpha1.Job) *conditionDetail {
	if vcJob.Status.State.Phase == "" {
		return &conditionDetail{
			Status: jobStatusEmpty,
			Reason: "JobNoInitialized",
			Message: "job condition is empty, volcano-controller maybe not running, " +
				"you can check the status by executing command: kubectl get pod -n volcano-system -l" +
				" app=volcano-controller",
		}
	} else if vcJob.Status.State.Phase == v1alpha1.Pending {
		return &conditionDetail{
			Status: jobStatusInitialized,
			Reason: "JobPending",
			Message: fmt.Sprintf("job is pending, "+
				"you can check the job status by executing command: kubectl describe [job-type, "+
				"eg vcjob,acjob] -n %s %s", vcJob.GetNamespace(), vcJob.GetName()),
		}
	}
	return nil
}

func (c *Collector) processAscendJob(ascendJob *batchv1.AscendJob) *conditionDetail {
	if len(ascendJob.Status.Conditions) == 0 {
		return &conditionDetail{
			Status: jobStatusEmpty,
			Reason: "JobNoInitialized",
			Message: "job condition is empty, ascend-operator maybe not running, " +
				"you can check the status by executing command: kubectl get pod -n mindx-dl -l app" +
				"=controller-manager",
		}
	}
	failedConditionIndex, createdConditionIndex := invalidIndex, invalidIndex
	for index, condition := range ascendJob.Status.Conditions {
		if condition.Type == "Failed" && condition.Status == corev1.ConditionTrue {
			failedConditionIndex = index
		}
		if condition.Type == "Created" && condition.Status == corev1.ConditionTrue {
			createdConditionIndex = index
		}
	}
	if failedConditionIndex >= 0 {
		return &conditionDetail{
			Status:  jobStatusFailed,
			Reason:  ascendJob.Status.Conditions[failedConditionIndex].Reason,
			Message: ascendJob.Status.Conditions[failedConditionIndex].Message,
		}
	}
	if createdConditionIndex >= 0 {
		return &conditionDetail{
			Status:  jobStatusInitialized,
			Reason:  ascendJob.Status.Conditions[createdConditionIndex].Reason,
			Message: ascendJob.Status.Conditions[createdConditionIndex].Message,
		}
	}
	return nil
}

func (c *Collector) cleanupJobs(allJobs map[string]constant.JobInfo, allMetaObjs map[string]metav1.Object) {
	for key := range c.jobExceptions {
		_, existInJobSummary := allJobs[key]
		_, existInMetaObjs := allMetaObjs[key]
		if !existInJobSummary && !existInMetaObjs {
			hwlog.RunLog.Infof("job %s is not exist, delete from cache", key)
			delete(c.jobExceptions, key)
		}
	}
}

func updateConfigMap(report exceptionReport) error {
	data := make(map[string]string)
	for key, jobInfo := range report.JobExceptions {
		data[key] = util.ObjToString(jobInfo)
	}

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cmName,
			Namespace: cmNamespace,
		},
		Data: data,
	}

	_, err := kube.CreateConfigMap(cm)
	if err == nil {
		return nil
	}

	if !errors.IsAlreadyExists(err) {
		return fmt.Errorf("create configmap failed: %v", err)
	}

	existingCM, err := kube.GetConfigMap(cmName, cmNamespace)
	if err != nil {
		return fmt.Errorf("get configmap failed: %v", err)
	}

	existingCM.Data = cm.Data
	_, err = kube.UpdateConfigMap(existingCM)
	if err != nil {
		return fmt.Errorf("update configmap failed: %v", err)
	}

	return nil
}
