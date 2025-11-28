// Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.

// Package statistics a series of statistic function
package statistics

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"ascend-common/api"
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

	acJobType = api.AscendJob

	acJobCreateTimeout = 5
	pgCreateTimeout    = 5
	pgInQueueTimeout   = 10
	checkJobStateLoop  = 1

	maxEventMsgLen = 1024

	jobInit    = "0%"
	jobCreated = "5%"
	pgCreated  = "10%"
	pgInQueue  = "35%"
	pgRunning  = "80%"
	jobRunning = "100%"
)

// JobStcMgr used to job statistic data.
type JobStcMgr struct {
	data            constant.CurrJobStatistic
	mutex           sync.RWMutex
	version         int64
	CheckTimeoutMap sync.Map
}

type jobCheckInfo struct {
	ScheduleChangeTime int64
	Schedule           string
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
		mutex:           sync.RWMutex{},
		version:         InitVersion,
		CheckTimeoutMap: sync.Map{},
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

func (j *JobStcMgr) checkTimeout() {
	var stateTimeoutMap = map[string]int64{
		jobInit:    acJobCreateTimeout,
		jobCreated: pgCreateTimeout,
		pgCreated:  pgInQueueTimeout,
	}
	var timeoutReasonMap = map[string]func(jobID string){
		jobInit:    j.getACJobCreateTimeoutReason,
		jobCreated: j.getPGCreateTimeoutReason,
		pgCreated:  j.getPGInQueueTimeoutReason,
	}
	nowTime := time.Now().Unix()
	j.CheckTimeoutMap.Range(func(key, value interface{}) bool {
		jobID, ok := key.(string)
		if !ok {
			return true
		}
		info, ok := value.(jobCheckInfo)
		if !ok {
			return true
		}
		if nowTime-info.ScheduleChangeTime > stateTimeoutMap[info.Schedule] {
			timeoutReasonMap[info.Schedule](jobID)
			j.CheckTimeoutMap.Delete(jobID)
		}
		return true
	})
}

// CheckJobScheduleTimeout check schedule timeout
func (j *JobStcMgr) CheckJobScheduleTimeout(ctx context.Context) {
	ticker := time.NewTicker(checkJobStateLoop * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			j.checkTimeout()
		case <-ctx.Done():
			hwlog.RunLog.Info("stop check timeout")
			return
		}
	}
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

// UpdateStcByPGUpdate update new job statistic from jobInfo
func (j *JobStcMgr) UpdateStcByPGUpdate(jobKey string) {
	jobInfo, ok := job.GetJobCache(jobKey)
	if !ok {
		hwlog.RunLog.Debugf("jobInfo cache is empty, skip update job %s statistc", jobKey)
		return
	}
	j.mutex.Lock()
	defer j.mutex.Unlock()
	jobStc, ok := j.data.JobStatistic[jobKey]
	if !ok {
		hwlog.RunLog.Debugf("jobStc key: %s is not in jobStatistic cache", jobKey)
		return
	}
	// update cache
	jobStc.ScheduleProcess = pgCreated
	if jobStc.Status != jobInfo.Status {
		hwlog.RunLog.Debugf("update jobStc, current job Status: %s, jobStc Status: %s",
			jobInfo.Status, jobStc.Status)
		newJobStc := updateStatistic(jobStc, jobInfo)
		j.data.JobStatistic[jobKey] = newJobStc
		j.version += 1
		// log job statistic
		logs.JobEventLog.Infof("Job Event: %s", util.ObjToString(newJobStc))
	}
}

// UpdateStcByPGCreate add new job statistic from pg add
func (j *JobStcMgr) UpdateStcByPGCreate(jobKey string) {
	if j.updateStcByPGCreate(jobKey) {
		JobStcMgrInst.CheckTimeoutMap.Store(jobKey, jobCheckInfo{
			ScheduleChangeTime: time.Now().Unix(),
			Schedule:           pgCreated,
		})
	}
}

func (j *JobStcMgr) updateStcByPGCreate(jobKey string) bool {
	j.mutex.Lock()
	defer j.mutex.Unlock()
	// add into cache
	jobStc, ok := j.data.JobStatistic[jobKey]
	if !ok {
		hwlog.RunLog.Debugf("jobStc cache is empty, skip update %s job statisc", jobKey)
		return false
	}
	if getIntValue(jobStc.ScheduleProcess) >= getIntValue(pgCreated) {
		return false
	}
	jobStc.ScheduleProcess = pgCreated
	jobStc.ScheduleFailReason = ""
	j.data.JobStatistic[jobKey] = jobStc
	j.version++
	// log job statistic
	hwlog.RunLog.Debugf("update Job statistic: %v", jobStc)
	logs.JobEventLog.Infof("Job Event: %s", util.ObjToString(jobStc))
	return true
}

func (j *JobStcMgr) getOldStopJobStc(jobInfo metav1.Object) string {
	for _, v := range j.data.JobStatistic {
		if v.Namespace != jobInfo.GetNamespace() ||
			v.Name != jobInfo.GetName() ||
			v.StopTime == 0 {
			continue
		}
		return v.K8sJobID
	}
	return ""
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
	jobStc, ok := j.data.JobStatistic[jobKey]
	if !ok {
		hwlog.RunLog.Debugf("jobStc cache is empty, skip update %s job statisc", jobKey)
		return
	}
	if jobStc.Status != jobInfo.Status {
		hwlog.RunLog.Debugf("update jobStc, current job Status: %s, jobStc Status: %s",
			jobInfo.Status, jobStc.Status)
		newJobStc := updateStatistic(jobStc, jobInfo)
		j.data.JobStatistic[jobKey] = newJobStc
		logs.JobEventLog.Infof("Job Event: %s", util.ObjToString(newJobStc))
		j.version += 1
	}
}

// DeleteJobStatistic delete jobStc cache, when jobInfo cache delete
func (j *JobStcMgr) DeleteJobStatistic(jobKey string) {
	j.mutex.Lock()
	defer j.mutex.Unlock()
	jobStc, ok := j.data.JobStatistic[jobKey]
	if !ok {
		hwlog.RunLog.Debugf("job Statistic cache is empty, skip delete job %s statistc", jobKey)
		return
	}
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
	j.version++
	j.CheckTimeoutMap.Delete(jobKey)
	hwlog.RunLog.Infof("delete jobStc, current jobStc Key: %s", jobKey)
}

// JobStcByACJobUpdate update jobStc when acJob update
func (j *JobStcMgr) JobStcByACJobUpdate(jobKey string) {
	if j.updateStcByACJobUpdate(jobKey) {
		JobStcMgrInst.CheckTimeoutMap.Store(jobKey, jobCheckInfo{
			ScheduleChangeTime: time.Now().Unix(),
			Schedule:           jobCreated,
		})
	}
}

// JobStcByJobDelete handel jobStc when job delete
func (j *JobStcMgr) JobStcByJobDelete(jobKey string) {
	j.mutex.Lock()
	defer j.mutex.Unlock()
	jobStc, ok := j.data.JobStatistic[jobKey]
	if !ok {
		hwlog.RunLog.Debugf("jobStc key: %s is not in jobStatistic cache", jobKey)
		return
	}
	//  the created job has configuration error
	if getIntValue(jobStc.ScheduleProcess) < getIntValue(pgCreated) && jobStc.StopTime == 0 {
		jobStc.Status = job.StatusJobFail
		jobStc.StopTime = time.Now().Unix()
		j.data.JobStatistic[jobKey] = jobStc
		j.version++
		logs.JobEventLog.Infof("Job Event: %s", util.ObjToString(jobStc))
	}
	// job has been stopped
	if jobStc.StopTime != 0 {
		delete(j.data.JobStatistic, jobKey)
		j.version++
		j.CheckTimeoutMap.Delete(jobKey)
		hwlog.RunLog.Infof("delete jobStc, current jobStc Key: %s, jobStc: %v", jobKey, jobStc)
	}
}

func (j *JobStcMgr) updateStcByACJobUpdate(jobKey string) bool {
	j.mutex.Lock()
	defer j.mutex.Unlock()
	jobStc, ok := j.data.JobStatistic[jobKey]
	if !ok {
		hwlog.RunLog.Warnf("jobStc key: %s is not in jobStatistic cache", jobKey)
		return false
	}
	if getIntValue(jobStc.ScheduleProcess) >= getIntValue(jobCreated) {
		return false
	}
	jobStc.ScheduleProcess = jobCreated
	jobStc.ScheduleFailReason = ""
	j.data.JobStatistic[jobKey] = jobStc
	j.version++
	hwlog.RunLog.Debugf("update Job statistic: %v", jobStc)
	logs.JobEventLog.Infof("Job Event: %s", util.ObjToString(jobStc))
	return true
}

// JobStcByACJobCreate jobStc when acJob add
func (j *JobStcMgr) JobStcByACJobCreate(jobKey string) {
	jobInfo := GetJob(jobKey)
	j.addJobStatistic(jobKey, jobInfo)
	j.CheckTimeoutMap.Store(jobKey, jobCheckInfo{
		ScheduleChangeTime: time.Now().Unix(),
		Schedule:           jobInit,
	})
}

// JobStcByVCJobCreate jobStc when vcJob add
func (j *JobStcMgr) JobStcByVCJobCreate(jobKey string) {
	jobInfo := GetJob(jobKey)
	j.addJobStatistic(jobKey, jobInfo)
}

func (j *JobStcMgr) addJobStatistic(jobKey string, jobInfo metav1.Object) {
	if jobInfo == nil {
		hwlog.RunLog.Warnf("job info is nil, cannot add job: %s", jobKey)
		return
	}
	if len(j.data.JobStatistic) == maxCMJobStatisticNum {
		hwlog.RunLog.Warnf("exceeded max tasksjob cache, cannot add job: %s", jobKey)
		return
	}
	j.mutex.Lock()
	defer j.mutex.Unlock()
	oldKey := j.getOldStopJobStc(jobInfo)
	if oldKey != "" && oldKey != jobKey {
		delete(j.data.JobStatistic, oldKey)
		j.version++
		j.CheckTimeoutMap.Delete(oldKey)
		hwlog.RunLog.Infof("delete old jobStc, key: %s", oldKey)
	}

	if _, ok := j.data.JobStatistic[jobKey]; ok {
		return
	}

	jobStc := initStcJob(jobInfo, jobKey)
	j.data.JobStatistic[jobKey] = jobStc
	j.version++
	hwlog.RunLog.Debugf("add Job statistic: %v, jobKey: %s", jobStc, jobKey)
	logs.JobEventLog.Infof("Job Event: %s", util.ObjToString(jobStc))
}

func (j *JobStcMgr) getPGInQueueTimeoutReason(jobKey string) {
	return
}

func (j *JobStcMgr) getPGCreateTimeoutReason(jobKey string) {
	j.mutex.Lock()
	defer j.mutex.Unlock()
	jobStc, ok := j.data.JobStatistic[jobKey]
	if !ok {
		hwlog.RunLog.Debugf("jobStc key: %s is not in jobStatistic cache", jobKey)
		return
	}
	if getIntValue(jobStc.ScheduleProcess) >= getIntValue(pgCreated) {
		return
	}
	events, err := kube.GetJobEvent(jobStc.Namespace, jobStc.Name, acJobType)
	if err != nil || len(events.Items) == 0 {
		jobStc.ScheduleFailReason = fmt.Sprintf("unknown reason, check %v event manually", api.AscendJob)
		j.data.JobStatistic[jobKey] = jobStc
		j.version++
		return
	}

	lastEvent := events.Items[len(events.Items)-1]
	if len(lastEvent.Message) >= maxEventMsgLen {
		lastEvent.Message = lastEvent.Message[:maxEventMsgLen]
	}
	jobStc.ScheduleFailReason = lastEvent.Message
	j.data.JobStatistic[jobKey] = jobStc
	j.version++
	hwlog.RunLog.Debugf("update Job statistic: %v", jobStc)
	logs.JobEventLog.Infof("Job Event: %s", util.ObjToString(jobStc))
}

func (j *JobStcMgr) getACJobCreateTimeoutReason(jobKey string) {
	j.mutex.Lock()
	defer j.mutex.Unlock()
	jobStc, ok := j.data.JobStatistic[jobKey]
	if !ok {
		hwlog.RunLog.Debugf("jobStc key: %s is not in jobStatistic cache", jobKey)
		return
	}
	if getIntValue(jobStc.ScheduleProcess) >= getIntValue(jobCreated) {
		return
	}
	jobStc.ScheduleFailReason =
		fmt.Sprintf("the acJob status is empty, it may be that the %v has not processed it", api.AscendOperator)
	j.data.JobStatistic[jobKey] = jobStc
	j.version++
	hwlog.RunLog.Infof("update Job statistic: %v", jobStc)
	logs.JobEventLog.Infof("Job Event: %s", util.ObjToString(jobStc))
}

func initStcJob(jobMeta metav1.Object, jobID string) constant.JobStatistic {
	jobStc := constant.JobStatistic{
		CustomJobID:     jobMeta.GetAnnotations()[job.CustomJobID],
		K8sJobID:        jobID,
		Status:          job.StatusJobPending,
		Name:            jobMeta.GetName(),
		Namespace:       jobMeta.GetNamespace(),
		ScheduleProcess: jobInit,
	}
	return jobStc
}

// updateStatistic update job statistic info
func updateStatistic(jobStc constant.JobStatistic, jobInfo constant.JobInfo) constant.JobStatistic {
	jobStc.Status = jobInfo.Status
	switch jobInfo.Status {
	case job.StatusJobPending:
		// job scheduling
		return jobStc
	case job.StatusJobRunning:
		nowTime := time.Now().Unix()
		jobStc.StopTime = 0
		// job start running success
		if jobStc.PodFirstRunningTime == 0 {
			jobStc.PodFirstRunningTime = nowTime
			cardNum := 0
			for _, serverList := range jobInfo.PreServerList {
				cardNum += len(serverList.DeviceList)
			}
			jobStc.CardNums = int64(cardNum)
			jobStc.ScheduleProcess = jobRunning
			jobStc.ScheduleFailReason = ""
			return jobStc
		}
		// job recover success
		if jobStc.PodLastFaultTime > jobStc.PodLastRunningTime {
			jobStc.PodLastRunningTime = nowTime
			return jobStc
		}
	case job.StatusJobCompleted:
		jobStc.StopTime = time.Now().Unix()
		return jobStc

	case job.StatusJobFail:
		if jobStc.PodLastRunningTime >= jobStc.PodLastFaultTime {
			jobStc.PodLastFaultTime = time.Now().Unix()
			jobStc.PodFaultTimes += 1
			if jobInfo.IsPreDelete {
				jobStc.StopTime = jobStc.PodLastFaultTime
			}
			return jobStc
		}
	default:
		return jobStc
	}
	return jobStc
}

func getIntValue(value string) int {
	value = strings.TrimSuffix(value, "%")
	val, err := strconv.Atoi(value)
	if err != nil {
		hwlog.RunLog.Errorf("convert value to int error, value is %s", value)
	}
	return val
}
