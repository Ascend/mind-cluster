// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

//go:build !race

// Package resource a series of resource test function
package resource

import (
	"reflect"
	"sync"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/client-go/kubernetes"

	"clusterd/pkg/application/job"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/device"
	"clusterd/pkg/domain/node"
	"clusterd/pkg/domain/switchinfo"
	"clusterd/pkg/interface/kube"
)

func mockAgentEmpty() *job.Agent {
	return &job.Agent{
		Config:        &job.Config{},
		BsWorker:      map[string]job.PodWorker{},
		KubeClientSet: &kubernetes.Clientset{},
		RwMutex:       sync.RWMutex{},
	}
}

func TestSaveDeviceInfoCM(t *testing.T) {
	convey.Convey("Test saveDeviceInfoCM", t, func() {
		patch := gomonkey.ApplyFunc(device.BusinessDataIsNotEqual,
			func(_ *constant.DeviceInfo, _ *constant.DeviceInfo) bool { return true }).
			ApplyFunc(updateJobDeviceHealth, func(_ string, _ map[string]string) {
				kube.JobMgr.BsWorker["test_task_id"] = &job.Worker{}
			}).ApplyFunc(AddNewMessageTotal, func() {}).ApplyGlobalVar(&kube.JobMgr, mockAgentEmpty())
		defer patch.Reset()
		saveDeviceInfoCM(&constant.DeviceInfo{})
		convey.So(len(kube.JobMgr.BsWorker), convey.ShouldEqual, 1)
	})
}

func TestUpdateJobDeviceHealth(t *testing.T) {
	convey.Convey("Test updateJobDeviceHealth", t, func() {
		nodeName := "node"
		deviceList := map[string]string{"NetworkUnhealthy": "value1", "Unhealthy": "value2", "otherKey": "value3"}
		convey.Convey("JobMgr is nil", func() {
			var jobMgr *job.Agent = nil
			mockJobMgr := gomonkey.ApplyGlobalVar(&kube.JobMgr, jobMgr)
			defer mockJobMgr.Reset()
			updateJobDeviceHealth(nodeName, deviceList)
			convey.So(kube.JobMgr, convey.ShouldBeNil)
		})
		convey.Convey("deviceList is nil", func() {
			mockJobMgr := gomonkey.ApplyGlobalVar(&kube.JobMgr, mockAgentEmpty())
			defer mockJobMgr.Reset()
			updateJobDeviceHealth(nodeName, map[string]string{})
			convey.So(len(kube.JobMgr.BsWorker), convey.ShouldEqual, 0)
		})
		convey.Convey("deviceList is not nil", func() {
			patch := gomonkey.ApplyMethod(reflect.TypeOf(new(job.Agent)), "UpdateJobDeviceStatus",
				func(_ *job.Agent, _ string, _, _ string) {
					kube.JobMgr.BsWorker["test_task_id"] = &job.Worker{}
				}).ApplyGlobalVar(&kube.JobMgr, mockAgentEmpty())
			defer patch.Reset()
			updateJobDeviceHealth(nodeName, deviceList)
			convey.So(len(kube.JobMgr.BsWorker), convey.ShouldEqual, 1)
		})
	})
}

func TestSaveSwitchInfoCM(t *testing.T) {
	convey.Convey("Test saveSwitchInfoCM", t, func() {
		patch := gomonkey.ApplyFunc(switchinfo.BusinessDataIsNotEqual,
			func(_, _ *constant.SwitchInfo) bool { return true }).
			ApplyFunc(updateJobNodeHealth, func(_ string, _ bool) {
				kube.JobMgr.BsWorker["test_task_id"] = &job.Worker{}
			}).ApplyFunc(AddNewMessageTotal, func() {}).ApplyGlobalVar(&kube.JobMgr, mockAgentEmpty())
		defer patch.Reset()
		saveSwitchInfoCM(&constant.SwitchInfo{})
		convey.So(len(kube.JobMgr.BsWorker), convey.ShouldEqual, 1)
	})
}

func TestSaveNodeInfoCM(t *testing.T) {
	convey.Convey("Test saveNodeInfoCM", t, func() {
		patch := gomonkey.ApplyFunc(node.BusinessDataIsNotEqual,
			func(_ *constant.NodeInfo, _ *constant.NodeInfo) bool {
				return true
			}).
			ApplyFunc(updateJobNodeHealth, func(_ string, _ bool) {
				kube.JobMgr.BsWorker["test_task_id"] = &job.Worker{}
			}).ApplyFunc(AddNewMessageTotal, func() {}).ApplyGlobalVar(&kube.JobMgr, mockAgentEmpty())
		defer patch.Reset()
		saveNodeInfoCM(&constant.NodeInfo{})
		convey.So(len(kube.JobMgr.BsWorker), convey.ShouldEqual, 1)
	})
}
