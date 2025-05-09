// Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.

// Package statistics a series of statistic function
package statistics

import (
	"encoding/json"
	"sync"
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/logs"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/job"
	"clusterd/pkg/interface/kube"
)

const (
	// JobStcCMName is the name of ConfigMap used to store job statistic data.
	JobStcCMName = "current-job-statistic"
	// JobDataCmKey is the key of ConfigMap data.
	JobDataCmKey = "data"
	// TotalJobsCmKey is the key of ConfigMap data.
	TotalJobsCmKey = "totalJob"
	// maxCMJobStatisticNum is the maximum number of job statistic data in ConfigMap.
	maxCMJobStatisticNum = 10000
	// InitVersion is the initial version of job statistic data.
	InitVersion = 0
)

// JobStcMgr used to job statistic data.
type JobStcMgr struct {
	data    constant.CurrJobStatistic
	mutex   sync.RWMutex
	version int64
}

var (
	// JobStcMgrInst is an instance of StatisticInfo used for statistic data.
	JobStcMgrInst *JobStcMgr
)

func init() {
	JobStcMgrInst = &JobStcMgr{
		data: constant.CurrJobStatistic{
			JobStatistic: make(map[string]constant.JobStatistic),
		},
		mutex:   sync.RWMutex{},
		version: InitVersion,
	}

}

// LoadConfigMapToCache Load ConfigMap To Cache
func (j *JobStcMgr) LoadConfigMapToCache(namespace, cmName string) {
	cmData, err := kube.GetConfigMap(cmName, namespace)
	if err != nil {
		if errors.IsNotFound(err) {
			hwlog.RunLog.Infof("job Statistic ConfigMap %s in namespace %s not found, skip loading in cache",
				cmName, namespace)
			return
		}
		hwlog.RunLog.Errorf("failed to get ConfigMap %s in namespace %s, err: %v", cmName, namespace, err)
		return
	}
	// parse ConfigMap data to cache
	j.mutex.Lock()
	defer j.mutex.Unlock()
	if !j.parseCMData(cmData) {
		hwlog.RunLog.Errorf("failed loaded ConfigMap %s in namespace %s into cache", cmName, namespace)
		return
	}
	hwlog.RunLog.Infof("successfully loaded ConfigMap %s in namespace %s into cache", cmName, namespace)
	hwlog.RunLog.Debugf("successfully loaded ConfigMap data: %s", cmData)
}

func (j *JobStcMgr) parseCMData(cmData *v1.ConfigMap) bool {
	oldJobDetails, ok := cmData.Data[JobDataCmKey]
	if !ok {
		hwlog.RunLog.Errorf("invalid old job statistic info, cm content: %v", cmData.Data)
		return false
	}
	tmpSlice := make([]constant.JobStatistic, 0)
	err := json.Unmarshal([]byte(oldJobDetails), &tmpSlice)
	if err != nil {
		hwlog.RunLog.Errorf("failed to unmarshal job statistic info:%s , err: %v", oldJobDetails, err)
		return false
	}
	for _, v := range tmpSlice {
		if v.StopTime != 0 {
			hwlog.RunLog.Debugf("old jobStc is stopped, delete it: %s", util.ObjToString(v))
			continue
		}
		j.data.JobStatistic[v.K8sJobID] = v
	}
	return true
}

// GetAllJobStatistic get all job statistic data
func (j *JobStcMgr) GetAllJobStatistic() (constant.CurrJobStatistic, int64) {
	j.mutex.RLock()
	defer j.mutex.RUnlock()
	res := constant.CurrJobStatistic{}
	err := util.DeepCopy(&res, &j.data)
	if err != nil {
		hwlog.RunLog.Errorf("deep copy failed: %v", err)
		return constant.CurrJobStatistic{}, j.version
	}
	return res, j.version
}

// UpdateJobStatistic update new job statistic from jobInfo
func (j *JobStcMgr) UpdateJobStatistic(jobKey string) {
	jobInfo, ok := job.GetJobCache(jobKey)
	if !ok {
		hwlog.RunLog.Debugf("jobInfo cache is empty, skip update job %s statistc", jobKey)
		return
	}

	j.mutex.Lock()
	defer j.mutex.Unlock()
	jobStc, ok := j.data.JobStatistic[jobKey]
	if !ok {
		// add into cache
		j.innerAddJobStatistic(jobKey, jobInfo)
		hwlog.RunLog.Debugf("jobStc key: %s is not in jobStatistic cache, add it", jobKey)
		return
	}

	// update cache
	if jobStc.Status != jobInfo.Status {
		hwlog.RunLog.Debugf("update jobStc, current job Status: %s, jobStc Status: %s",
			jobInfo.Status, jobStc.Status)
		newJobStc := UpdateStatistic(jobStc, jobInfo)
		j.data.JobStatistic[jobKey] = newJobStc
		j.version += 1
		// log job statistic
		logs.JobEventLog.Infof("Job Event: %s", util.ObjToString(newJobStc))

	}
}

func (j *JobStcMgr) innerAddJobStatistic(jobKey string, jobInfo constant.JobInfo) {
	if len(j.data.JobStatistic) == maxCMJobStatisticNum {
		hwlog.RunLog.Warnf("exceeded the maximum number of tasksjob cache, can not add job: %s", jobKey)
		return
	}

	// add into cache
	jobStc := InitStatistic(jobInfo, jobKey)
	j.data.JobStatistic[jobKey] = jobStc
	j.version += 1
	// log job statistic
	logs.JobEventLog.Infof("Job Event: %s", util.ObjToString(jobStc))
}

// AddJobStatistic add new job statistic from jobInfo add
func (j *JobStcMgr) AddJobStatistic(jobKey string) {
	if len(j.data.JobStatistic) == maxCMJobStatisticNum {
		hwlog.RunLog.Warnf("exceeded the maximum number of tasksjob cache, can not add job: %s", jobKey)
		return
	}
	jobInfo, ok := job.GetJobCache(jobKey)
	if !ok {
		hwlog.RunLog.Debugf("jobInfo cache is empty, skip add %s job statisc", jobKey)
		return
	}

	j.mutex.Lock()
	defer j.mutex.Unlock()
	OldJobStcKey := j.getOldJobStc(jobInfo)
	if OldJobStcKey != "" {
		delete(j.data.JobStatistic, OldJobStcKey)
		j.version += 1
		hwlog.RunLog.Infof("delete old jobStc, key :%s", OldJobStcKey)
	}

	// add into cache
	if _, ok := j.data.JobStatistic[jobKey]; ok {
		return
	}
	jobStc := InitStatistic(jobInfo, jobKey)
	j.data.JobStatistic[jobKey] = jobStc
	j.version += 1
	// log job statistic
	hwlog.RunLog.Debugf("add Job statistic: %s", util.ObjToString(jobStc))
	logs.JobEventLog.Infof("Job Event: %s", util.ObjToString(jobStc))
}

func (j *JobStcMgr) getOldJobStc(jobInfo constant.JobInfo) string {
	oldKey := ""
	for _, v := range j.data.JobStatistic {
		if v.NameSpace != jobInfo.NameSpace || v.Name != jobInfo.Name || v.StopTime == 0 {
			continue
		}
		oldKey = v.K8sJobID
		break
	}
	return oldKey
}

// PreDeleteJobStatistic when the jobInfo is IsPreDelete
func (j *JobStcMgr) PreDeleteJobStatistic(jobKey string) {
	jobInfo, ok := job.GetJobCache(jobKey)
	if !ok {
		hwlog.RunLog.Debugf("jobInfo cache is empty, skip delete job %s statistc", jobKey)
		return
	}

	j.mutex.Lock()
	defer j.mutex.Unlock()
	jobStc := j.data.JobStatistic[jobKey]
	if jobStc.Status != jobInfo.Status {
		hwlog.RunLog.Debugf("update jobStc, current job Status: %s, jobStc Status: %s",
			jobInfo.Status, jobStc.Status)
		newJobStc := UpdateStatistic(jobStc, jobInfo)
		j.data.JobStatistic[jobKey] = newJobStc
		logs.JobEventLog.Infof("Job Event: %s", util.ObjToString(newJobStc))
		j.version += 1
	}
}

// DeleteJobStatistic delete jobStc cache, when jobInfo cache delete
func (j *JobStcMgr) DeleteJobStatistic(jobKey string) {
	jobInfo, ok := job.GetJobCache(jobKey)
	if !ok {
		hwlog.RunLog.Debugf("jobInfo cache is empty, job key: %s", jobKey)
		return
	}

	j.mutex.Lock()
	defer j.mutex.Unlock()
	jobStc, ok := j.data.JobStatistic[jobKey]
	if !ok {
		hwlog.RunLog.Debugf("job Statistic cache is empty, skip delete job %s statistc", jobKey)
		return
	}
	hwlog.RunLog.Infof("delete jobStc, current job Status: %s, jobStc Key: %s",
		jobInfo.Status, jobKey)
	//  update the stop time if stopTime not exist
	if jobStc.StopTime == 0 {
		if jobStc.PodLastFaultTime > jobStc.PodLastRunningTime {
			jobStc.StopTime = jobStc.PodLastFaultTime
		} else {
			jobStc.StopTime = time.Now().Unix()
		}
		j.data.JobStatistic[jobKey] = jobStc
		logs.JobEventLog.Infof("Job Event: %s", util.ObjToString(jobStc))
	}

	delete(j.data.JobStatistic, jobKey)
	j.version += 1
}
