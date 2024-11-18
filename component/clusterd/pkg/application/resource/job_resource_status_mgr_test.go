// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

//go:build !race

// Package resource a series of resource test function
package resource

import (
	"context"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/npu-exporter/v6/common-utils/hwlog"

	"clusterd/pkg/application/job"
	"clusterd/pkg/interface/grpc/common"
	"clusterd/pkg/interface/grpc/pb"
	"clusterd/pkg/interface/kube"
)

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwlog.InitRunLogger(&hwLogConfig, context.Background())
}

func mockJobMgr() *JobSourceStatusManager {
	return &JobSourceStatusManager{
		publisher:  nil,
		notifyChan: make(chan *common.Notifier, notifyChanLen),
	}
}

// TestNewJobWorker test NewJobWorker
func TestNewJobSourceStatusManager(t *testing.T) {
	convey.Convey("test NewJobSourceStatusManager", t, func() {
		result := NewJobSourceStatusManager(nil)
		convey.So(result, convey.ShouldNotBeNil)
	})
}

// TestNotifySignalSend test NotifySignalSend
func TestNotifySignalSend(t *testing.T) {
	convey.Convey("test NotifySignalSend", t, func() {
		mgr := mockJobMgr()
		mgr.NotifySignalSend(&common.Notifier{
			CreateTimeStamp:     0,
			ProcessManageSignal: pb.ProcessManageSignal{},
		})
		convey.So(len(mgr.notifyChan), convey.ShouldEqual, 1)
	})
}

// TestGetJobNameAndNameSpace test GetJobNameAndNameSpace
func TestGetJobNameAndNameSpace(t *testing.T) {
	convey.Convey("test GetJobNameAndNameSpace", t, func() {
		mgr := mockJobMgr()
		convey.Convey("case jobMgr is nil", func() {
			kube.JobMgr = nil
			name, namespace := mgr.GetJobNameAndNameSpace("test_task_id")
			convey.So(name, convey.ShouldEqual, "")
			convey.So(namespace, convey.ShouldEqual, "")
		})
		convey.Convey("case task id not exist", func() {
			kube.JobMgr = &job.Agent{
				BsWorker: make(map[string]job.PodWorker),
			}
			kube.JobMgr.BsWorker["test_task_id"] = &job.Worker{}
			name, namespace := mgr.GetJobNameAndNameSpace("test_task_id_not_exist")
			convey.So(name, convey.ShouldEqual, "")
			convey.So(namespace, convey.ShouldEqual, "")
		})
		convey.Convey("case task id exist", func() {
			kube.JobMgr = &job.Agent{
				BsWorker: make(map[string]job.PodWorker),
			}
			kube.JobMgr.BsWorker["test_task_id"] = &job.Worker{
				WorkerInfo: job.WorkerInfo{},
				Info: job.Info{
					Name:      "test_name",
					Namespace: "test_namespace",
				},
			}
			name, namespace := mgr.GetJobNameAndNameSpace("test_task_id")
			convey.So(name, convey.ShouldEqual, "test_name")
			convey.So(namespace, convey.ShouldEqual, "test_namespace")
		})

	})
}

// TestGetJobDeviceNumPerNode test GetJobDeviceNumPerNode
func TestGetJobDeviceNumPerNode(t *testing.T) {
	convey.Convey("test TestGetJobDeviceNumPerNode", t, func() {
		mgr := mockJobMgr()
		convey.Convey("case jobMgr is nil", func() {
			kube.JobMgr = nil
			deviceNum := mgr.GetJobDeviceNumPerNode("test_task_id")
			convey.So(deviceNum, convey.ShouldEqual, -1)
		})
		convey.Convey("case task id not exist", func() {
			kube.JobMgr = &job.Agent{
				BsWorker: make(map[string]job.PodWorker),
			}
			kube.JobMgr.BsWorker["test_task_id"] = &job.Worker{}
			deviceNum := mgr.GetJobDeviceNumPerNode("test_task_id_not_exist")
			convey.So(deviceNum, convey.ShouldEqual, -1)
		})
		convey.Convey("case task id exist", func() {
			kube.JobMgr = &job.Agent{
				BsWorker: make(map[string]job.PodWorker),
			}
			kube.JobMgr.BsWorker["test_task_id"] = &job.Worker{
				WorkerInfo: job.WorkerInfo{
					CMData: &job.RankTable{},
				},
				Info: job.Info{
					Name:      "test_name",
					Namespace: "test_namespace",
				},
			}
			mockDeviceNum := gomonkey.ApplyMethod(reflect.TypeOf(&job.RankTable{}), "GetJobDeviceNumPerNode",
				func(_ *job.RankTable) int {
					return 1
				})
			defer mockDeviceNum.Reset()
			deviceNum := mgr.GetJobDeviceNumPerNode("test_task_id")
			convey.So(deviceNum, convey.ShouldEqual, 1)
		})

	})
}

func TestGetJobHealthy(t *testing.T) {
	convey.Convey("Test GetJobHealthy", t, func() {
		mgr := mockJobMgr()
		convey.Convey("JobMgr is nil", func() {
			var jobMgr *job.Agent = nil
			mockJobMgr := gomonkey.ApplyGlobalVar(&kube.JobMgr, jobMgr)
			defer mockJobMgr.Reset()
			isHealth, faultRanks := mgr.GetJobHealthy("")
			convey.So(isHealth, convey.ShouldBeFalse)
			convey.So(faultRanks, convey.ShouldBeNil)
		})
		convey.Convey("JobMgr is not nil and jobId is not exist", func() {
			jobMgr := mockAgentEmpty()
			jobMgr.BsWorker["test_task_id"] = &job.Worker{}
			mockJobMgr := gomonkey.ApplyGlobalVar(&kube.JobMgr, jobMgr)
			defer mockJobMgr.Reset()
			isHealth, faultRanks := mgr.GetJobHealthy("")
			convey.So(isHealth, convey.ShouldBeFalse)
			convey.So(faultRanks, convey.ShouldBeNil)
		})
		convey.Convey("JobMgr is not nil and jobId is exist", func() {
			jobMgr := mockAgentEmpty()
			jobMgr.BsWorker["test_task_id"] = &job.Worker{}
			patch := gomonkey.ApplyMethod(reflect.TypeOf(new(job.WorkerInfo)), "GetJobHealth",
				func(_ *job.WorkerInfo) (bool, []string) { return true, []string{} }).
				ApplyGlobalVar(&kube.JobMgr, jobMgr)
			defer patch.Reset()
			isHealth, faultRanks := mgr.GetJobHealthy("test_task_id")
			convey.So(isHealth, convey.ShouldBeTrue)
			convey.So(faultRanks, convey.ShouldNotBeNil)
		})
	})
}
