// Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.

// Package statistics a series of statistic function
package statistics

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"volcano.sh/apis/pkg/apis/batch/v1alpha1"

	ascendv1 "ascend-common/api/ascend-operator/apis/batch/v1"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/job"
	"clusterd/pkg/interface/kube"
)

func TestLoadConfigMapToCache(t *testing.T) {
	t.Run("cm successful load, Data length is 1", func(t *testing.T) {
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		patches.ApplyFunc(kube.GetConfigMap, func(name, namespace string) (*v1.ConfigMap, error) {
			return &v1.ConfigMap{
				Data: map[string]string{
					JobDataCmKey:   `[{"Name":"test-job"}]`,
					TotalJobsCmKey: "1",
				},
			}, nil
		})

		JobStcMgrInst.LoadConfigMapToCache("ns", "cm")
		assert.Equal(t, 1, len(JobStcMgrInst.data.JobStatistic))
	})

	t.Run("when cm is not found, Data length is 0", func(t *testing.T) {
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		patches.ApplyFunc(kube.GetConfigMap, func(name, namespace string) (*v1.ConfigMap, error) {
			return nil, k8serror.NewNotFound(v1.Resource("configmap"), "cm")
		})
		JobStcMgrInst.data.JobStatistic = make(map[string]constant.JobStatistic)
		JobStcMgrInst.LoadConfigMapToCache("ns", "cm")
		assert.Equal(t, 0, len(JobStcMgrInst.data.JobStatistic))
	})

	t.Run("when get cm  failed, Data length is 0", func(t *testing.T) {
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		patches.ApplyFunc(kube.GetConfigMap, func(name, namespace string) (*v1.ConfigMap, error) {
			return nil, errors.New("get cm failed")
		})
		JobStcMgrInst.data.JobStatistic = make(map[string]constant.JobStatistic)
		JobStcMgrInst.LoadConfigMapToCache("ns", "cm")
		assert.Equal(t, 0, len(JobStcMgrInst.data.JobStatistic))
	})

	t.Run("when parse cm ok", func(t *testing.T) {
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		tmpSlice := make([]constant.JobStatistic, 0)
		tmpSlice = append(tmpSlice, constant.JobStatistic{})
		cmData := &v1.ConfigMap{
			Data: map[string]string{JobDataCmKey: util.ObjToString(tmpSlice)},
		}
		patches.ApplyFunc(kube.GetConfigMap, func(name, namespace string) (*v1.ConfigMap, error) {
			return cmData, nil
		})

		JobStcMgrInst.data.JobStatistic = make(map[string]constant.JobStatistic)
		JobStcMgrInst.LoadConfigMapToCache("ns", "cm")
		assert.Equal(t, 1, len(JobStcMgrInst.data.JobStatistic))
	})
}

func TestParseCMData(t *testing.T) {
	t.Run("valid data, Data length is 2", func(t *testing.T) {
		cm := &v1.ConfigMap{
			Data: map[string]string{
				JobDataCmKey:   `[{"Name":"job1","StopTime":0},{"Name":"job2","StopTime":1}]`,
				TotalJobsCmKey: "2",
			},
		}
		result := JobStcMgrInst.parseCMData(cm)
		assert.True(t, result)
		assert.Equal(t, 1, len(JobStcMgrInst.data.JobStatistic))
	})

	t.Run("invalid jobs data, parse failed", func(t *testing.T) {
		cm := &v1.ConfigMap{
			Data: map[string]string{
				JobDataCmKey:   "1",
				TotalJobsCmKey: "2",
			},
		}
		result := JobStcMgrInst.parseCMData(cm)
		assert.False(t, result)
	})
}

func TestUpdateJobStcByPGUpdate(t *testing.T) {
	t.Run("update when exist jobInfo, update to running", func(t *testing.T) {
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		var jobKey = "test-1"
		job.SaveJobCache(jobKey, constant.JobInfo{
			Status: job.StatusJobRunning,
			Key:    jobKey,
		})
		defer job.DeleteJobCache(jobKey)
		JobStcMgrInst.data.JobStatistic[jobKey] = constant.JobStatistic{Status: job.StatusJobPending}
		JobStcMgrInst.UpdateStcByPGUpdate(jobKey)
		assert.Equal(t, job.StatusJobRunning, JobStcMgrInst.data.JobStatistic[(jobKey)].Status)
		JobStcMgrInst.data.JobStatistic = make(map[string]constant.JobStatistic)
	})

	t.Run("update get job cache failed,", func(t *testing.T) {
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		var jobKey = "test-1"
		patches.ApplyFunc(job.GetJobCache, func(key string) (constant.JobInfo, bool) {
			return constant.JobInfo{}, false
		})
		JobStcMgrInst.data.JobStatistic[jobKey] = constant.JobStatistic{Status: job.StatusJobPending}
		JobStcMgrInst.UpdateStcByPGUpdate(jobKey)
		assert.NotEqual(t, job.StatusJobRunning, JobStcMgrInst.data.JobStatistic[(jobKey)].Status)
		JobStcMgrInst.data.JobStatistic = make(map[string]constant.JobStatistic)
	})

	t.Run("update get job when get jobStc failed,", func(t *testing.T) {
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		var jobKey = "test-1"
		patches.ApplyFunc(job.GetJobCache, func(key string) (constant.JobInfo, bool) {
			return constant.JobInfo{}, true
		})
		JobStcMgrInst.data.JobStatistic = make(map[string]constant.JobStatistic)
		JobStcMgrInst.UpdateStcByPGUpdate(jobKey)
		assert.NotEqual(t, job.StatusJobRunning, JobStcMgrInst.data.JobStatistic[(jobKey)].Status)
	})
}

func TestDeleteJobStatistic(t *testing.T) {
	patches := gomonkey.NewPatches()
	defer patches.Reset()
	var jobKey = "test-1"
	job.SaveJobCache(jobKey, constant.JobInfo{CustomJobID: jobKey})
	defer job.DeleteJobCache(jobKey)
	patches.ApplyFunc(kube.UpdateConfigMap, func(cm *v1.ConfigMap) (*v1.ConfigMap, error) {
		return &v1.ConfigMap{
			Data: map[string]string{
				JobDataCmKey:   `[{"Name":"test-job"}]`,
				TotalJobsCmKey: "1",
			},
		}, nil
	})
	JobStcMgrInst.data.JobStatistic = make(map[string]constant.JobStatistic)

	t.Run("normal delete", func(t *testing.T) {
		JobStcMgrInst.data.JobStatistic[jobKey] = constant.JobStatistic{StopTime: 1}
		JobStcMgrInst.DeleteJobStatistic(jobKey)
		assert.Equal(t, 0, len(JobStcMgrInst.data.JobStatistic))
	})

	t.Run("delete when stop time is 0, update stop time", func(t *testing.T) {
		JobStcMgrInst.data.JobStatistic[jobKey] = constant.JobStatistic{StopTime: 0}
		JobStcMgrInst.DeleteJobStatistic(jobKey)
		assert.Equal(t, 0, len(JobStcMgrInst.data.JobStatistic))
	})
}

func TestPreDeleteJobStatistic(t *testing.T) {
	patches := gomonkey.NewPatches()
	defer patches.Reset()
	var jobKey = "test"
	job.SaveJobCache(jobKey, constant.JobInfo{Status: "failed"})
	defer job.DeleteJobCache(jobKey)
	JobStcMgrInst.data.JobStatistic[jobKey] = constant.JobStatistic{Status: "Running"}
	JobStcMgrInst.PreDeleteJobStatistic(jobKey)
	assert.Equal(t, "failed", JobStcMgrInst.data.JobStatistic[jobKey].Status)
}

func TestUpdateStatistic(t *testing.T) {
	now := time.Now().Unix()
	baseJobStc := constant.JobStatistic{PodFirstRunningTime: 0, PodLastFaultTime: 0}
	t.Run("Pending status - no change", func(t *testing.T) {
		jobInfo := constant.JobInfo{Status: job.StatusJobPending}
		result := updateStatistic(baseJobStc, jobInfo)
		assert.Equal(t, job.StatusJobPending, result.Status)
	})
	t.Run("Running status - first running", func(t *testing.T) {
		jobInfo := constant.JobInfo{
			Status: job.StatusJobRunning,
			PreServerList: []constant.ServerHccl{{DeviceList: []constant.Device{{DeviceID: "1"}, {DeviceID: "2"}}},
				{DeviceList: []constant.Device{{DeviceID: "3"}}}}}
		result := updateStatistic(baseJobStc, jobInfo)
		assert.NotZero(t, result.PodFirstRunningTime)
		assert.Equal(t, int64(3), result.CardNums)
	})
	t.Run("Running status - recover from fault", func(t *testing.T) {
		jobStc := baseJobStc
		jobStc.PodLastFaultTime = now - 1
		jobStc.PodFirstRunningTime = now - 1 - 1
		jobInfo := constant.JobInfo{Status: job.StatusJobRunning}
		result := updateStatistic(jobStc, jobInfo)
		assert.NotZero(t, result.PodLastRunningTime)
	})
	t.Run("Completed status - set stop time", func(t *testing.T) {
		jobInfo := constant.JobInfo{Status: job.StatusJobCompleted}
		result := updateStatistic(baseJobStc, jobInfo)
		assert.NotZero(t, result.StopTime)
	})
	t.Run("Failed status - normal failure", func(t *testing.T) {
		jobInfo := constant.JobInfo{Status: job.StatusJobFail}
		result := updateStatistic(baseJobStc, jobInfo)
		assert.NotZero(t, result.PodLastFaultTime)
		assert.Equal(t, int64(1), result.PodFaultTimes)
	})
	t.Run("Failed status with pre-delete", func(t *testing.T) {
		jobInfo := constant.JobInfo{Status: job.StatusJobFail, IsPreDelete: true}
		result := updateStatistic(baseJobStc, jobInfo)
		assert.Equal(t, result.PodLastFaultTime, result.StopTime)
	})
	t.Run("Unknown status - no change", func(t *testing.T) {
		jobInfo := constant.JobInfo{Status: "Unknown"}
		result := updateStatistic(baseJobStc, jobInfo)
		assert.Equal(t, baseJobStc.PodFirstRunningTime, result.PodFirstRunningTime)
	})
}

func TestCheckPGCreateTimeout(t *testing.T) {
	var jobKey = "test-key"
	t.Run("get jobStc failed", func(t *testing.T) {
		JobStcMgrInst.data.JobStatistic = make(map[string]constant.JobStatistic)
		JobStcMgrInst.getPGCreateTimeoutReason(jobKey)
		assert.Equal(t, 0, len(JobStcMgrInst.data.JobStatistic))
	})

	t.Run("get job event failed", func(t *testing.T) {
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		patches.ApplyFunc(kube.GetJobEvent, func(namespace, name, jobType string) (*v1.EventList, error) {
			return &v1.EventList{}, errors.New("error")
		})
		JobStcMgrInst.data.JobStatistic[jobKey] = constant.JobStatistic{ScheduleProcess: jobCreated}
		JobStcMgrInst.getPGCreateTimeoutReason(jobKey)
		assert.Contains(t, JobStcMgrInst.data.JobStatistic[jobKey].ScheduleFailReason, "unknown reason")

	})

	t.Run("pg create timeout", func(t *testing.T) {
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		errMsg := "pg failed"
		patches.ApplyFunc(kube.GetJobEvent, func(namespace, name, jobType string) (*v1.EventList, error) {
			eList := &v1.EventList{}
			eList.Items = append(eList.Items, v1.Event{Message: errMsg})
			return eList, nil
		})
		JobStcMgrInst.data.JobStatistic[jobKey] = constant.JobStatistic{ScheduleProcess: jobCreated}
		JobStcMgrInst.getPGCreateTimeoutReason(jobKey)
		assert.Equal(t, errMsg, JobStcMgrInst.data.JobStatistic[jobKey].ScheduleFailReason)
	})
}

func TestAddJobStatistic(t *testing.T) {
	t.Run("add new job success and delete old job", func(t *testing.T) {
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		jobKey := "test-key"
		jobName := "test-job"
		jobNamespace := "test-namespace"
		oldJobKey := "test-old"
		jobInfo := &ascendv1.AscendJob{
			ObjectMeta: metav1.ObjectMeta{
				Name:      jobName,
				Namespace: jobNamespace,
			},
		}
		JobStcMgrInst.data.JobStatistic = make(map[string]constant.JobStatistic)
		JobStcMgrInst.data.JobStatistic[oldJobKey] = constant.JobStatistic{
			K8sJobID:  oldJobKey,
			Name:      jobName,
			Namespace: jobNamespace,
			StopTime:  time.Now().Unix(),
		}
		patches.ApplyFunc(initStcJob, func(jobMeta metav1.Object, jobID string) constant.JobStatistic {
			return constant.JobStatistic{}
		})
		JobStcMgrInst.addJobStatistic(jobKey, jobInfo)
		assert.Equal(t, 1, len(JobStcMgrInst.data.JobStatistic))
	})
}

func TestGetOldJobStc(t *testing.T) {
	jobName := "test-job"
	jobNamespace := "test-namespace"
	oldJobKey := "test-old"
	jobInfo := &ascendv1.AscendJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: jobNamespace,
		},
	}
	t.Run("get a old job key", func(t *testing.T) {
		JobStcMgrInst.data.JobStatistic = make(map[string]constant.JobStatistic)
		JobStcMgrInst.data.JobStatistic[oldJobKey] = constant.JobStatistic{
			K8sJobID:  oldJobKey,
			Name:      jobName,
			Namespace: jobNamespace,
			StopTime:  time.Now().Unix(),
		}
		key := JobStcMgrInst.getOldStopJobStc(jobInfo)
		assert.Equal(t, oldJobKey, key)
		JobStcMgrInst.data.JobStatistic = make(map[string]constant.JobStatistic)
	})

	t.Run("get key is null", func(t *testing.T) {
		JobStcMgrInst.data.JobStatistic = make(map[string]constant.JobStatistic)
		JobStcMgrInst.data.JobStatistic[oldJobKey] = constant.JobStatistic{
			K8sJobID:  oldJobKey,
			Name:      jobName,
			Namespace: jobNamespace,
			StopTime:  0,
		}
		key := JobStcMgrInst.getOldStopJobStc(jobInfo)
		assert.Equal(t, "", key)
		JobStcMgrInst.data.JobStatistic = make(map[string]constant.JobStatistic)
	})
}

func TestInitStcJob(t *testing.T) {
	t.Run("init job success", func(t *testing.T) {
		jobName := "test-job"
		jobNamespace := "test-namespace"
		jobKey := "test-key"
		jobInfo := &ascendv1.AscendJob{
			ObjectMeta: metav1.ObjectMeta{
				Name:      jobName,
				Namespace: jobNamespace,
			},
		}
		jobStc := initStcJob(jobInfo, jobKey)
		assert.Equal(t, jobName, jobStc.Name)
	})
}

func TestGetIntValue(t *testing.T) {
	t.Run("get value ok", func(t *testing.T) {
		val := getIntValue("5%")
		assert.Equal(t, 5, val)
	})
	t.Run("get value failed", func(t *testing.T) {
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		patches.ApplyFunc(strconv.Atoi, func(s string) (int, error) {
			return 0, errors.New("error")
		})
		val := getIntValue("5%")
		assert.Equal(t, 0, val)
	})
}

func TestJobStcByACJobUpdate(t *testing.T) {
	t.Run("need check pg create timeout", func(t *testing.T) {
		jobKey := "test-key"
		JobStcMgrInst.data.JobStatistic[jobKey] = constant.JobStatistic{
			K8sJobID:        jobKey,
			ScheduleProcess: jobInit,
		}
		JobStcMgrInst.CheckTimeoutMap.Range(func(key, value interface{}) bool {
			JobStcMgrInst.CheckTimeoutMap.Delete(key)
			return true
		})
		JobStcMgrInst.JobStcByACJobUpdate(jobKey)
		_, ok := JobStcMgrInst.CheckTimeoutMap.Load(jobKey)
		assert.Equal(t, true, ok)
		defer JobStcMgrInst.CheckTimeoutMap.Delete(jobKey)
	})

}

func TestUpdateStcByACJobUpdate(t *testing.T) {
	t.Run("need check pg create timeout", func(t *testing.T) {
		JobStcMgrInst.data.JobStatistic = make(map[string]constant.JobStatistic)
		jobKey := "test-key"
		JobStcMgrInst.data.JobStatistic[jobKey] = constant.JobStatistic{
			K8sJobID:        jobKey,
			ScheduleProcess: jobInit,
		}
		ret := JobStcMgrInst.updateStcByACJobUpdate(jobKey)
		assert.Equal(t, true, ret)
	})
	t.Run("job key not in cache", func(t *testing.T) {
		JobStcMgrInst.data.JobStatistic = make(map[string]constant.JobStatistic)
		jobKey := "test-key"
		ret := JobStcMgrInst.updateStcByACJobUpdate(jobKey)
		assert.Equal(t, false, ret)
	})

	t.Run("job scheduler >=  jobCreated", func(t *testing.T) {
		JobStcMgrInst.data.JobStatistic = make(map[string]constant.JobStatistic)
		jobKey := "test-job"
		JobStcMgrInst.data.JobStatistic[jobKey] = constant.JobStatistic{
			K8sJobID:        jobKey,
			ScheduleProcess: jobCreated,
		}
		ret := JobStcMgrInst.updateStcByACJobUpdate(jobKey)
		assert.Equal(t, false, ret)
	})
}

func TestJobStcByJobDelete(t *testing.T) {
	t.Run("delete job not in cache", func(t *testing.T) {
		jobKey := "test-key"
		JobStcMgrInst.data.JobStatistic = make(map[string]constant.JobStatistic)
		JobStcMgrInst.JobStcByJobDelete(jobKey)
		assert.Equal(t, 0, len(JobStcMgrInst.data.JobStatistic))
	})

	t.Run("delete job , job scheduleProcess <= pgCreated", func(t *testing.T) {
		jobKey := "test-key"
		JobStcMgrInst.data.JobStatistic = make(map[string]constant.JobStatistic)
		JobStcMgrInst.data.JobStatistic[jobKey] = constant.JobStatistic{
			K8sJobID:        jobKey,
			ScheduleProcess: jobInit,
		}
		JobStcMgrInst.JobStcByJobDelete(jobKey)
		assert.Equal(t, 0, len(JobStcMgrInst.data.JobStatistic))
	})

	t.Run("delete job in cache, stopTime != 0", func(t *testing.T) {
		jobKey := "test-key"
		JobStcMgrInst.data.JobStatistic = make(map[string]constant.JobStatistic)
		JobStcMgrInst.data.JobStatistic[jobKey] = constant.JobStatistic{
			K8sJobID:        jobKey,
			ScheduleProcess: pgRunning,
			StopTime:        time.Now().Unix(),
		}
		JobStcMgrInst.JobStcByJobDelete(jobKey)
		assert.Equal(t, 0, len(JobStcMgrInst.data.JobStatistic))
	})
}

func TestJobStcByACJobAdd(t *testing.T) {
	t.Run("ac job add ok", func(t *testing.T) {
		jobKey := "jobKey"
		jobInfo := &ascendv1.AscendJob{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-name",
				Namespace: "test-namespace",
				UID:       types.UID(jobKey),
			},
		}
		SaveJob(jobInfo)
		defer DeleteJob(jobInfo)
		JobStcMgrInst.data.JobStatistic = make(map[string]constant.JobStatistic)
		JobStcMgrInst.JobStcByACJobCreate(jobKey)
		assert.Equal(t, 1, len(JobStcMgrInst.data.JobStatistic))
	})
}

func TestJobStcByVCJobAdd(t *testing.T) {
	t.Run("vc job add ok", func(t *testing.T) {
		jobKey := "jobKey"
		jobInfo := &v1alpha1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-name",
				Namespace: "test-namespace",
				UID:       types.UID(jobKey),
			},
		}
		SaveJob(jobInfo)
		defer DeleteJob(jobInfo)
		JobStcMgrInst.data.JobStatistic = make(map[string]constant.JobStatistic)
		JobStcMgrInst.JobStcByVCJobCreate(jobKey)
		assert.Equal(t, 1, len(JobStcMgrInst.data.JobStatistic))
	})
}

func TestGetAllJobStatistic(t *testing.T) {
	JobStcMgrInst.data.JobStatistic = make(map[string]constant.JobStatistic)
	t.Run("get data ok", func(t *testing.T) {
		JobStcMgrInst.data.JobStatistic["test"] = constant.JobStatistic{Status: "Running"}
		JobStcMgrInst.version = 1
		data, version := JobStcMgrInst.GetAllJobStatistic()
		assert.Equal(t, "Running", data.JobStatistic["test"].Status)
		assert.Equal(t, "Running", JobStcMgrInst.data.JobStatistic["test"].Status)
		assert.Equal(t, int64(1), version)
	})

	t.Run("change data after get data", func(t *testing.T) {
		JobStcMgrInst.data.JobStatistic["test"] = constant.JobStatistic{Status: "Running"}
		JobStcMgrInst.version = 1
		data, version := JobStcMgrInst.GetAllJobStatistic()
		delete(data.JobStatistic, "test")
		assert.Equal(t, 0, len(data.JobStatistic))
		assert.Equal(t, 1, len(JobStcMgrInst.data.JobStatistic))
		assert.Equal(t, int64(1), version)
	})
}

func TestCheckACJobCreateTimeout(t *testing.T) {
	t.Run("acJob not in cache", func(t *testing.T) {
		jobKey := "test-key"
		JobStcMgrInst.data.JobStatistic = make(map[string]constant.JobStatistic)
		JobStcMgrInst.getACJobCreateTimeoutReason(jobKey)
		assert.Equal(t, 0, len(JobStcMgrInst.data.JobStatistic))
	})

	t.Run("acJob create timeout", func(t *testing.T) {
		jobKey := "test-key"
		JobStcMgrInst.data.JobStatistic = make(map[string]constant.JobStatistic)
		JobStcMgrInst.data.JobStatistic[jobKey] = constant.JobStatistic{
			K8sJobID:        jobKey,
			ScheduleProcess: jobInit,
		}
		JobStcMgrInst.getACJobCreateTimeoutReason(jobKey)
		assert.Contains(t, JobStcMgrInst.data.JobStatistic[jobKey].ScheduleFailReason, "ascend-Operator")
	})
}

func TestCheckJobScheduleTimeout(t *testing.T) {
	t.Run("check timeout ok", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		JobStcMgrInst.data.JobStatistic = make(map[string]constant.JobStatistic)
		JobStcMgrInst.CheckTimeoutMap.Range(func(key, value interface{}) bool {
			JobStcMgrInst.CheckTimeoutMap.Delete(key)
			return true
		})
		go JobStcMgrInst.CheckJobScheduleTimeout(ctx)
		time.Sleep(time.Second)
	})
}

func TestCheckTimeout(t *testing.T) {
	t.Run("check timeout ok", func(t *testing.T) {
		jobKey := "jobKey"
		JobStcMgrInst.data.JobStatistic = make(map[string]constant.JobStatistic)
		JobStcMgrInst.data.JobStatistic[jobKey] = constant.JobStatistic{
			K8sJobID:        jobKey,
			ScheduleProcess: jobInit,
		}
		JobStcMgrInst.CheckTimeoutMap.Store(jobKey, jobCheckInfo{
			ScheduleChangeTime: time.Now().Unix() - acJobCreateTimeout - acJobCreateTimeout,
			Schedule:           jobInit,
		})
		JobStcMgrInst.checkTimeout()
		assert.NotEqual(t, "", JobStcMgrInst.data.JobStatistic[jobKey].ScheduleFailReason)
	})
}

func TestUpdateByPGAdd(t *testing.T) {
	t.Run("need check pg in queue timeout", func(t *testing.T) {
		JobStcMgrInst.data.JobStatistic = make(map[string]constant.JobStatistic)
		jobKey := "test-key"
		JobStcMgrInst.data.JobStatistic[jobKey] = constant.JobStatistic{
			K8sJobID:        jobKey,
			ScheduleProcess: jobInit,
		}
		ret := JobStcMgrInst.updateStcByPGCreate(jobKey)
		assert.Equal(t, true, ret)
	})
	t.Run("job key not in cache", func(t *testing.T) {
		JobStcMgrInst.data.JobStatistic = make(map[string]constant.JobStatistic)
		jobKey := "test-key"
		ret := JobStcMgrInst.updateStcByPGCreate(jobKey)
		assert.Equal(t, false, ret)
	})

	t.Run("job scheduler >=  pgCreated", func(t *testing.T) {
		JobStcMgrInst.data.JobStatistic = make(map[string]constant.JobStatistic)
		jobKey := "test-job"
		JobStcMgrInst.data.JobStatistic[jobKey] = constant.JobStatistic{
			K8sJobID:        jobKey,
			ScheduleProcess: pgCreated,
		}
		ret := JobStcMgrInst.updateStcByPGCreate(jobKey)
		assert.Equal(t, false, ret)
	})
}

func TestUpdateStcByPGAdd(t *testing.T) {
	t.Run("need check pg in queue timeout", func(t *testing.T) {
		JobStcMgrInst.data.JobStatistic = make(map[string]constant.JobStatistic)
		jobKey := "test-key"
		JobStcMgrInst.data.JobStatistic[jobKey] = constant.JobStatistic{
			K8sJobID:        jobKey,
			ScheduleProcess: jobInit,
		}
		JobStcMgrInst.CheckTimeoutMap.Range(func(key, value interface{}) bool {
			JobStcMgrInst.CheckTimeoutMap.Delete(key)
			return true
		})
		JobStcMgrInst.UpdateStcByPGCreate(jobKey)
		_, ok := JobStcMgrInst.CheckTimeoutMap.Load(jobKey)
		assert.Equal(t, true, ok)
		defer JobStcMgrInst.CheckTimeoutMap.Delete(jobKey)
	})

}
