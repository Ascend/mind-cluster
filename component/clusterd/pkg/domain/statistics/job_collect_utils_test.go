// Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.

// Package statistics a series of statistic function
package statistics

import (
	"strconv"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"

	"clusterd/pkg/common/constant"
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
			return nil, errors.NewNotFound(v1.Resource("configmap"), "cm")
		})
		JobStcMgrInst.data.JobStatistic = make(map[string]constant.JobStatistic)
		JobStcMgrInst.LoadConfigMapToCache("ns", "cm")
		assert.Equal(t, 0, len(JobStcMgrInst.data.JobStatistic))
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

func TestUpdateJobStatistic(t *testing.T) {
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	// Mock job cache
	patches.ApplyFunc(job.GetJobCache, func(key string) (constant.JobInfo, bool) {
		return constant.JobInfo{
			Status:      "Running",
			CustomJobID: "test-1",
		}, true
	})

	t.Run("update when exist jobInfo, update to running", func(t *testing.T) {
		JobStcMgrInst.data.JobStatistic["test-1"] = constant.JobStatistic{Status: "Pending"}
		JobStcMgrInst.UpdateJobStatistic("test-1")
		assert.Equal(t, "Running", JobStcMgrInst.data.JobStatistic[("test-1")].Status)
	})

	t.Run("update when not exist jobInfo, add new jobStc", func(t *testing.T) {
		for k := range JobStcMgrInst.data.JobStatistic {
			delete(JobStcMgrInst.data.JobStatistic, k)
		}
		JobStcMgrInst.UpdateJobStatistic("test-2")
		assert.Equal(t, 1, len(JobStcMgrInst.data.JobStatistic))
	})
}

func TestAddJobStatistic(t *testing.T) {
	patches := gomonkey.NewPatches()
	defer patches.Reset()
	for k := range JobStcMgrInst.data.JobStatistic {
		delete(JobStcMgrInst.data.JobStatistic, k)
	}

	patches.ApplyFunc(job.GetJobCache, func(key string) (constant.JobInfo, bool) {
		return constant.JobInfo{
			Name:        "test-job",
			NameSpace:   "ns",
			CustomJobID: "cid-1",
		}, true
	})

	t.Run("add new job success", func(t *testing.T) {
		JobStcMgrInst.AddJobStatistic("test-key")
		assert.Equal(t, 1, len(JobStcMgrInst.data.JobStatistic))
		delete(JobStcMgrInst.data.JobStatistic, "test-key")
	})

	t.Run("add new job when exist old job, delete old job", func(t *testing.T) {
		JobStcMgrInst.data.JobStatistic["test-key"] = constant.JobStatistic{
			Name:      "test-job",
			NameSpace: "ns",
			StopTime:  1,
		}
		JobStcMgrInst.AddJobStatistic("test-key")
		assert.Equal(t, 1, len(JobStcMgrInst.data.JobStatistic))
		delete(JobStcMgrInst.data.JobStatistic, "test-key")
	})
}

func TestDeleteJobStatistic(t *testing.T) {
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyFunc(job.GetJobCache, func(key string) (constant.JobInfo, bool) {
		return constant.JobInfo{CustomJobID: "test-1"}, true
	})

	patches.ApplyFunc(kube.UpdateConfigMap, func(cm *v1.ConfigMap) (*v1.ConfigMap, error) {
		return &v1.ConfigMap{
			Data: map[string]string{
				JobDataCmKey:   `[{"Name":"test-job"}]`,
				TotalJobsCmKey: "1",
			},
		}, nil
	})
	for k := range JobStcMgrInst.data.JobStatistic {
		delete(JobStcMgrInst.data.JobStatistic, k)
	}

	t.Run("normal delete", func(t *testing.T) {
		JobStcMgrInst.data.JobStatistic[("test-1")] = constant.JobStatistic{StopTime: 1}
		JobStcMgrInst.DeleteJobStatistic("test-1")
		assert.Equal(t, 0, len(JobStcMgrInst.data.JobStatistic))
	})

	t.Run("delete when stop time is 0, update stop time", func(t *testing.T) {
		JobStcMgrInst.data.JobStatistic[("test-1")] = constant.JobStatistic{StopTime: 0}
		JobStcMgrInst.DeleteJobStatistic("test-1")
		assert.Equal(t, 0, len(JobStcMgrInst.data.JobStatistic))
	})
}

func TestPreDeleteJobStatistic(t *testing.T) {
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyFunc(job.GetJobCache, func(key string) (constant.JobInfo, bool) {
		return constant.JobInfo{Status: "failed"}, true
	})

	JobStcMgrInst.data.JobStatistic["test"] = constant.JobStatistic{Status: "Running"}
	JobStcMgrInst.PreDeleteJobStatistic("test")
	assert.Equal(t, "failed", JobStcMgrInst.data.JobStatistic["test"].Status)
}

func TestInnerAddJobStatistic(t *testing.T) {
	t.Run("exceed max limit", func(t *testing.T) {
		JobStcMgrInst.data.JobStatistic = make(map[string]constant.JobStatistic)
		for i := 0; i < maxCMJobStatisticNum; i++ {
			JobStcMgrInst.data.JobStatistic[strconv.Itoa(i)] = constant.JobStatistic{}
		}
		JobStcMgrInst.innerAddJobStatistic("test", constant.JobInfo{})
		assert.Equal(t, maxCMJobStatisticNum, len(JobStcMgrInst.data.JobStatistic))
	})

	t.Run("normal add", func(t *testing.T) {
		JobStcMgrInst.data.JobStatistic = make(map[string]constant.JobStatistic)
		JobStcMgrInst.innerAddJobStatistic("test", constant.JobInfo{})
		assert.Equal(t, 1, len(JobStcMgrInst.data.JobStatistic))
	})
}

func TestGetAllJobStatistic(t *testing.T) {
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
