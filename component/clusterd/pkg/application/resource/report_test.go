// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

//go:build !race

// Package resource a series of resource test function
package resource

import (
	"context"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/application/faultmanager"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/device"
	"clusterd/pkg/domain/node"
	"clusterd/pkg/domain/switchinfo"
	"clusterd/pkg/interface/kube"
)

const time2s = 2 * time.Second

func init() {
	hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, context.Background())
}

func TestAddNewMessageTotal(t *testing.T) {
	convey.Convey("test updateChan message", t, func() {
		convey.Convey("test updateChan message", func() {

		})
		AddNewMessageTotal()
		var num = 1
		if updateChan != nil {
			num, _ = <-updateChan
		}
		convey.So(num, convey.ShouldEqual, constant.AllProcessType)
	})
}

func TestStopReport(t *testing.T) {
	convey.Convey("test stop report", t, func() {
		cycleTicker = time.NewTicker(1 * time.Second)
		convey.So(StopReport, convey.ShouldNotPanic)
		updateChan = make(chan int, 1)
	})
}

func TestUpdateCmWithEmpty(t *testing.T) {
	convey.Convey("Test updateCmWithEmpty", t, func() {
		convey.Convey("update config by deviceArr", func() {
			clientSet := fake.NewSimpleClientset()
			mockUpdateConfig := gomonkey.ApplyFunc(updateConfig, func(_, _ string) {
				cm := initFakeCMByDataMap("test-cm", "vcjob", map[string]string{"key": "value"})
				clientSet.CoreV1().ConfigMaps(cm.Namespace).Create(context.TODO(), cm, metav1.CreateOptions{})
			})
			defer mockUpdateConfig.Reset()
			updateAllCm([]string{"device"}, []string{"node"}, []string{"switch"})
			cm, err := clientSet.CoreV1().ConfigMaps("vcjob").Get(context.TODO(),
				"test-cm", metav1.GetOptions{})
			convey.So(err, convey.ShouldBeNil)
			convey.So(len(cm.Data), convey.ShouldEqual, 1)
		})
	})
}

func initFakeCMByDataMap(name, namespace string, m map[string]string) *v1.ConfigMap {
	return &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: m,
	}
}

func TestReport(t *testing.T) {
	convey.Convey("test stop report", t, func() {
		gomonkey.ApplyFunc(kube.UpdateConfigMap, func(cm *v1.ConfigMap) (*v1.ConfigMap, error) {
			return nil, nil
		})
		ctx, cancel := context.WithCancel(context.Background())
		faultmanager.NewFaultProcessCenter(ctx)
		updateChan = make(chan int, 5)
		updateChan <- constant.AllProcessType
		updateChan <- constant.DeviceProcessType
		updateChan <- constant.NodeProcessType
		updateChan <- constant.SwitchProcessType

		go Report(ctx)
		time.Sleep(time2s)
		cancel()
		convey.So(len(updateChan), convey.ShouldEqual, 0)
	})
}

func TestUpdateAllCm(t *testing.T) {
	convey.Convey("TestUpdateAllCm", t, func() {
		deviceArr := device.GetSafeData(map[string]*constant.DeviceInfo{
			"test": {},
		})
		nodeArr := node.GetSafeData(map[string]*constant.NodeInfo{
			"test": {},
		})
		switchArr := switchinfo.GetSafeData(map[string]*constant.SwitchInfo{
			"test": {},
		})
		gomonkey.ApplyFunc(kube.UpdateConfigMap, func(cm *v1.ConfigMap) (*v1.ConfigMap, error) {
			return nil, nil
		})
		updateAllCm(deviceArr, nodeArr, switchArr)
	})
}
