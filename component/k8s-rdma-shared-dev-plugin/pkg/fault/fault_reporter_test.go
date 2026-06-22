/*
   Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package fault for fault check and fault report
package fault

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/client-go/kubernetes"

	"ascend-common/common-utils/hwlog"
	util "github.com/Mellanox/k8s-rdma-shared-dev-plugin/pkg/utils"
)

func init() {
	_ = hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, context.Background())
}

func TestIsDpuInfoCfgEqualDifferentDPUListLength(t *testing.T) {
	a := &DpuInfoCfg{}
	a.DPUInfo.DPUList = []DPUItem{{HcaName: "mlx5_0"}}
	b := &DpuInfoCfg{}
	b.DPUInfo.DPUList = []DPUItem{{HcaName: "mlx5_0"}, {HcaName: "mlx5_1"}}

	convey.Convey("When DPU list lengths differ", t, func() {
		convey.So(isDpuInfoCfgEqual(a, b), convey.ShouldBeFalse)
	})
}

func TestIsDpuInfoCfgEqualSameDPUList(t *testing.T) {
	a := &DpuInfoCfg{}
	a.DPUInfo.DPUList = []DPUItem{{HcaName: "mlx5_0", EthName: "enp0s1"}}
	b := &DpuInfoCfg{}
	b.DPUInfo.DPUList = []DPUItem{{HcaName: "mlx5_0", EthName: "enp0s1"}}

	convey.Convey("When DPU lists are identical", t, func() {
		convey.So(isDpuInfoCfgEqual(a, b), convey.ShouldBeTrue)
	})
}

func TestIsDpuInfoCfgEqualDifferentFaultList(t *testing.T) {
	a := &DpuInfoCfg{}
	a.DPUInfo.DPUList = []DPUItem{{
		HcaName:   "mlx5_0",
		FaultList: []FaultDetail{{FaultCode: "21000022"}},
	}}
	b := &DpuInfoCfg{}
	b.DPUInfo.DPUList = []DPUItem{{
		HcaName:   "mlx5_0",
		FaultList: []FaultDetail{{FaultCode: "21000023"}},
	}}

	convey.Convey("When fault lists differ", t, func() {
		convey.So(isDpuInfoCfgEqual(a, b), convey.ShouldBeFalse)
	})
}

func TestIsDpuInfoCfgEqualDifferentNodeEvent(t *testing.T) {
	a := &DpuInfoCfg{}
	a.DPUInfo.NodeEvent = &NodeEvent{NodeName: "node1"}
	b := &DpuInfoCfg{}
	b.DPUInfo.NodeEvent = &NodeEvent{NodeName: "node2"}

	convey.Convey("When node events differ", t, func() {
		convey.So(isDpuInfoCfgEqual(a, b), convey.ShouldBeFalse)
	})
}

func TestIsDpuInfoCfgEqualNilNodeEvents(t *testing.T) {
	a := &DpuInfoCfg{}
	b := &DpuInfoCfg{}

	convey.Convey("When both node events are nil", t, func() {
		convey.So(isDpuInfoCfgEqual(a, b), convey.ShouldBeTrue)
	})
}

func TestIsDPUItemEqualSame(t *testing.T) {
	a := &DPUItem{HcaName: "mlx5_0", EthName: "enp0s1", IpAddr: "10.0.0.1", DeviceID: "0xa222", VendorID: "0x19e5"}
	b := &DPUItem{HcaName: "mlx5_0", EthName: "enp0s1", IpAddr: "10.0.0.1", DeviceID: "0xa222", VendorID: "0x19e5"}

	convey.Convey("When DPU items are identical", t, func() {
		convey.So(isDPUItemEqual(a, b), convey.ShouldBeTrue)
	})
}

func TestIsDPUItemEqualDifferentHcaName(t *testing.T) {
	a := &DPUItem{HcaName: "mlx5_0"}
	b := &DPUItem{HcaName: "mlx5_1"}

	convey.Convey("When HCA names differ", t, func() {
		convey.So(isDPUItemEqual(a, b), convey.ShouldBeFalse)
	})
}

func TestIsDPUItemEqualDifferentEthName(t *testing.T) {
	a := &DPUItem{HcaName: "mlx5_0", EthName: "enp0s1"}
	b := &DPUItem{HcaName: "mlx5_0", EthName: "enp0s2"}

	convey.Convey("When eth names differ", t, func() {
		convey.So(isDPUItemEqual(a, b), convey.ShouldBeFalse)
	})
}

func TestIsDPUItemEqualDifferentIpAddr(t *testing.T) {
	a := &DPUItem{HcaName: "mlx5_0", IpAddr: "10.0.0.1"}
	b := &DPUItem{HcaName: "mlx5_0", IpAddr: "10.0.0.2"}

	convey.Convey("When IP addresses differ", t, func() {
		convey.So(isDPUItemEqual(a, b), convey.ShouldBeFalse)
	})
}

func TestIsDPUItemEqualDifferentDeviceID(t *testing.T) {
	a := &DPUItem{HcaName: "mlx5_0", DeviceID: "0xa222"}
	b := &DPUItem{HcaName: "mlx5_0", DeviceID: "0xa223"}

	convey.Convey("When device IDs differ", t, func() {
		convey.So(isDPUItemEqual(a, b), convey.ShouldBeFalse)
	})
}

func TestIsDPUItemEqualDifferentVendorID(t *testing.T) {
	a := &DPUItem{HcaName: "mlx5_0", VendorID: "0x19e5"}
	b := &DPUItem{HcaName: "mlx5_0", VendorID: "0x19e6"}

	convey.Convey("When vendor IDs differ", t, func() {
		convey.So(isDPUItemEqual(a, b), convey.ShouldBeFalse)
	})
}

func TestIsDPUItemEqualDifferentFaultList(t *testing.T) {
	a := &DPUItem{HcaName: "mlx5_0", FaultList: []FaultDetail{{FaultCode: "21000022"}}}
	b := &DPUItem{HcaName: "mlx5_0", FaultList: []FaultDetail{{FaultCode: "21000023"}}}

	convey.Convey("When fault lists differ", t, func() {
		convey.So(isDPUItemEqual(a, b), convey.ShouldBeFalse)
	})
}

func TestUpdateLastReportedDataNil(t *testing.T) {
	lastReportedDataMu.Lock()
	savedLastReportedData := lastReportedData
	lastReportedDataMu.Unlock()
	t.Cleanup(func() {
		lastReportedDataMu.Lock()
		lastReportedData = savedLastReportedData
		lastReportedDataMu.Unlock()
	})

	convey.Convey("When updating with nil", t, func() {
		updateLastReportedData(nil)
		lastReportedDataMu.RLock()
		convey.So(lastReportedData, convey.ShouldBeNil)
		lastReportedDataMu.RUnlock()
	})
}

func TestUpdateLastReportedDataSuccess(t *testing.T) {
	lastReportedDataMu.Lock()
	savedLastReportedData := lastReportedData
	lastReportedDataMu.Unlock()
	t.Cleanup(func() {
		lastReportedDataMu.Lock()
		lastReportedData = savedLastReportedData
		lastReportedDataMu.Unlock()
	})

	dpuCfg := DpuInfoCfg{}
	dpuCfg.DPUInfo.DPUList = []DPUItem{{HcaName: "mlx5_0"}}
	dpuCfg.UpdateTime = time.Now()

	convey.Convey("When updating with valid data", t, func() {
		updateLastReportedData(&dpuCfg)
		lastReportedDataMu.RLock()
		convey.So(lastReportedData, convey.ShouldNotBeNil)
		convey.So(lastReportedData.DPUInfo.DPUList[0].HcaName, convey.ShouldEqual, "mlx5_0")
		lastReportedDataMu.RUnlock()
	})
}

func TestReportToConfigMapGetNodeNameError(t *testing.T) {
	lastReportedDataMu.Lock()
	savedLastReportedData := lastReportedData
	lastReportedDataMu.Unlock()
	t.Cleanup(func() {
		lastReportedDataMu.Lock()
		lastReportedData = savedLastReportedData
		lastReportedDataMu.Unlock()
	})

	patches := gomonkey.ApplyFunc(util.GetNodeName, func() (string, error) {
		return "", errors.New("node name not set")
	})
	defer patches.Reset()

	lastReportedDataMu.Lock()
	lastReportedData = nil
	lastReportedDataMu.Unlock()

	convey.Convey("When GetNodeName returns error", t, func() {
		err := ReportToConfigMap(DpuInfoCfg{})
		convey.Convey("Then error should be returned", func() {
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldContainSubstring, "failed to get node name")
		})
	})
}

func TestReportToConfigMapNoChanges(t *testing.T) {
	lastReportedDataMu.Lock()
	savedLastReportedData := lastReportedData
	lastReportedDataMu.Unlock()
	t.Cleanup(func() {
		lastReportedDataMu.Lock()
		lastReportedData = savedLastReportedData
		lastReportedDataMu.Unlock()
	})

	patches := gomonkey.ApplyFunc(util.GetNodeName, func() (string, error) {
		return "test-node", nil
	})
	defer patches.Reset()

	dpuCfg := DpuInfoCfg{}
	dpuCfg.DPUInfo.DPUList = []DPUItem{{HcaName: "mlx5_0"}}
	dpuCfg.UpdateTime = time.Now()

	lastReportedDataMu.Lock()
	lastReportedData = &dpuCfg
	lastReportedDataMu.Unlock()

	convey.Convey("When no changes detected and within force update interval", t, func() {
		err := ReportToConfigMap(dpuCfg)
		convey.Convey("Then no update should be needed", func() {
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestReportToConfigMapForceUpdate(t *testing.T) {
	lastReportedDataMu.Lock()
	savedLastReportedData := lastReportedData
	lastReportedDataMu.Unlock()
	t.Cleanup(func() {
		lastReportedDataMu.Lock()
		lastReportedData = savedLastReportedData
		lastReportedDataMu.Unlock()
	})

	patches := gomonkey.ApplyFunc(util.GetNodeName, func() (string, error) {
		return "test-node", nil
	})
	patches.ApplyFunc(CreateOrUpdateConfigMap, func(cmName, cfgJSONStr string) error {
		return nil
	})
	defer patches.Reset()

	oldTime := time.Now().Add(-10 * time.Minute)
	dpuCfg := DpuInfoCfg{}
	dpuCfg.DPUInfo.DPUList = []DPUItem{{HcaName: "mlx5_0"}}
	dpuCfg.UpdateTime = time.Now()

	lastReportedDataMu.Lock()
	lastReportedData = &DpuInfoCfg{}
	lastReportedData.DPUInfo.DPUList = []DPUItem{{HcaName: "mlx5_0"}}
	lastReportedData.UpdateTime = oldTime
	lastReportedDataMu.Unlock()

	convey.Convey("When force update interval has passed", t, func() {
		err := ReportToConfigMap(dpuCfg)
		convey.Convey("Then update should be triggered", func() {
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestReportToConfigMapDataChanged(t *testing.T) {
	lastReportedDataMu.Lock()
	savedLastReportedData := lastReportedData
	lastReportedDataMu.Unlock()
	t.Cleanup(func() {
		lastReportedDataMu.Lock()
		lastReportedData = savedLastReportedData
		lastReportedDataMu.Unlock()
	})

	patches := gomonkey.ApplyFunc(util.GetNodeName, func() (string, error) {
		return "test-node", nil
	})
	patches.ApplyFunc(CreateOrUpdateConfigMap, func(cmName, cfgJSONStr string) error {
		return nil
	})
	defer patches.Reset()

	dpuCfg := DpuInfoCfg{}
	dpuCfg.DPUInfo.DPUList = []DPUItem{{HcaName: "mlx5_1"}}
	dpuCfg.UpdateTime = time.Now()

	lastReportedDataMu.Lock()
	lastReportedData = &DpuInfoCfg{}
	lastReportedData.DPUInfo.DPUList = []DPUItem{{HcaName: "mlx5_0"}}
	lastReportedData.UpdateTime = time.Now()
	lastReportedDataMu.Unlock()

	convey.Convey("When data has changed", t, func() {
		err := ReportToConfigMap(dpuCfg)
		convey.Convey("Then update should be triggered", func() {
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestReportToConfigMapCreateOrUpdateError(t *testing.T) {
	lastReportedDataMu.Lock()
	savedLastReportedData := lastReportedData
	lastReportedDataMu.Unlock()
	t.Cleanup(func() {
		lastReportedDataMu.Lock()
		lastReportedData = savedLastReportedData
		lastReportedDataMu.Unlock()
	})

	patches := gomonkey.ApplyFunc(util.GetNodeName, func() (string, error) {
		return "test-node", nil
	})
	patches.ApplyFunc(CreateOrUpdateConfigMap, func(cmName, cfgJSONStr string) error {
		return errors.New("configmap update failed")
	})
	defer patches.Reset()

	lastReportedDataMu.Lock()
	lastReportedData = nil
	lastReportedDataMu.Unlock()

	convey.Convey("When CreateOrUpdateConfigMap returns error", t, func() {
		err := ReportToConfigMap(DpuInfoCfg{})
		convey.Convey("Then error should be returned", func() {
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldContainSubstring, "configmap update failed")
		})
	})
}

func TestReportToConfigMapFirstReport(t *testing.T) {
	lastReportedDataMu.Lock()
	savedLastReportedData := lastReportedData
	lastReportedDataMu.Unlock()
	t.Cleanup(func() {
		lastReportedDataMu.Lock()
		lastReportedData = savedLastReportedData
		lastReportedDataMu.Unlock()
	})

	patches := gomonkey.ApplyFunc(util.GetNodeName, func() (string, error) {
		return "test-node", nil
	})
	patches.ApplyFunc(CreateOrUpdateConfigMap, func(cmName, cfgJSONStr string) error {
		return nil
	})
	defer patches.Reset()

	lastReportedDataMu.Lock()
	lastReportedData = nil
	lastReportedDataMu.Unlock()

	dpuCfg := DpuInfoCfg{}
	dpuCfg.DPUInfo.DPUList = []DPUItem{{HcaName: "mlx5_0"}}
	dpuCfg.UpdateTime = time.Now()

	convey.Convey("When first report (lastReportedData is nil)", t, func() {
		err := ReportToConfigMap(dpuCfg)
		convey.Convey("Then update should be triggered", func() {
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestStartFaultReportingContextCancel(t *testing.T) {
	savedK8sClientset := k8sClientset
	savedAppCtx := appCtx
	t.Cleanup(func() {
		k8sClientset = savedK8sClientset
		appCtx = savedAppCtx
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	convey.Convey("When context is cancelled", t, func() {
		StartFaultReporting(ctx, &kubernetes.Clientset{})
		convey.So(true, convey.ShouldBeTrue)
	})
}

func TestStartFaultReportingReceiveFromChannel(t *testing.T) {
	savedK8sClientset := k8sClientset
	savedAppCtx := appCtx
	t.Cleanup(func() {
		k8sClientset = savedK8sClientset
		appCtx = savedAppCtx
		select {
		case <-faultResultChan:
		default:
		}
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	k8sClientset = &kubernetes.Clientset{}
	appCtx = ctx

	patches := gomonkey.ApplyFunc(ReportToConfigMap, func(dpuCfg DpuInfoCfg) error {
		cancel()
		return nil
	})
	defer patches.Reset()

	convey.Convey("When receiving from fault result channel", t, func() {
		faultResultChan <- DpuInfoCfg{}
		StartFaultReporting(ctx, &kubernetes.Clientset{})
		convey.So(true, convey.ShouldBeTrue)
	})
}

func TestStartFaultReportingReportError(t *testing.T) {
	savedK8sClientset := k8sClientset
	savedAppCtx := appCtx
	t.Cleanup(func() {
		k8sClientset = savedK8sClientset
		appCtx = savedAppCtx
		select {
		case <-faultResultChan:
		default:
		}
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	k8sClientset = &kubernetes.Clientset{}
	appCtx = ctx

	patches := gomonkey.ApplyFunc(ReportToConfigMap, func(dpuCfg DpuInfoCfg) error {
		cancel()
		return errors.New("report error")
	})
	defer patches.Reset()

	convey.Convey("When reporting faults fails", t, func() {
		faultResultChan <- DpuInfoCfg{}
		StartFaultReporting(ctx, &kubernetes.Clientset{})
		convey.So(true, convey.ShouldBeTrue)
	})
}
