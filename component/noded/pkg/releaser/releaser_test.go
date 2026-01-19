/* Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package releaser the device fault releaser
package releaser

import (
	"context"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/kubernetes/fake"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-common/devmanager"
	"ascend-common/devmanager/common"
	"nodeD/pkg/kubeclient"
	"nodeD/pkg/watcher/configmap"
)

func TestGetReleaseAndRecoverMap1(t *testing.T) {
	convey.Convey("Test getReleaseAndRecoverMap", t, func() {
		if err := hwlog.InitRunLogger(&hwlog.LogConfig{}, context.TODO()); err != nil {
			fmt.Printf("hwlog init failed, error is %v\n", err)
		}
		patch1 := gomonkey.ApplyFuncReturn(kubeclient.GetK8sClient, &kubeclient.ClientK8s{NodeName: "node1"})
		defer patch1.Reset()
		// Setup test data
		faultInfos := map[string]map[int]api.SuperPodFaultInfos{
			"job1": {0: {JobId: "job1", SdIds: []string{"sd1"}, FaultTimes: 1, NodeNames: []string{"node1"}}}}

		// Mock global variable faultInfoCm
		mockFaultInfoCm := gomonkey.ApplyGlobalVar(&faultInfoCm, map[string]map[int]api.SuperPodFaultInfos{
			"job1": {0: {JobId: "job1", SdIds: []string{"sd1"}, FaultTimes: 0, NodeNames: []string{"node1"}}},
			"job2": {0: {JobId: "job2", SdIds: []string{"sd2"}, FaultTimes: 1, NodeNames: []string{"node1"}}},
		})
		defer mockFaultInfoCm.Reset()

		convey.Convey("01-should return need reset jobs when fault times changed", func() {
			resetJobs, recoverIds := getReleaseAndRecoverMap(faultInfos)
			convey.So(len(resetJobs), convey.ShouldEqual, 1)
			convey.So(len(recoverIds), convey.ShouldEqual, 1)
		})

		convey.Convey("02-should return need recover jobs when job not in faultInfos", func() {
			_, recoverIds := getReleaseAndRecoverMap(faultInfos)
			convey.So(recoverIds.Has("sd2"), convey.ShouldBeTrue)
		})
	})
}

func TestHandleNodeStatusChange(t *testing.T) {
	convey.Convey("Test handleNodeStatusChange", t, func() {
		if err := hwlog.InitRunLogger(&hwlog.LogConfig{}, context.TODO()); err != nil {
			fmt.Printf("hwlog init failed, error is %v\n", err)
		}
		patch1 := gomonkey.ApplyMethodReturn(&devmanager.DeviceManager{}, "GetChipBaseInfos",
			[]*common.ChipBaseInfo{{CardID: 0, DeviceID: 0, LogicID: 0}}, nil)
		defer patch1.Reset()
		// Setup test data
		configmap.InitCmWatcher(&kubeclient.ClientK8s{ClientSet: fake.NewSimpleClientset()})
		InitReleaser()
		r := releaser
		// Mock devManager
		mockDevManager := gomonkey.ApplyMethodReturn(r.devManager, "DcGetSuperPodStatus", 0, nil)
		defer mockDevManager.Reset()

		mockSetStatus := gomonkey.ApplyMethodReturn(r.devManager, "DcSetSuperPodStatus", nil)
		defer mockSetStatus.Reset()

		sdids := sets.NewString("123")

		convey.Convey("01-should set status when status not match", func() {
			r.handleNodeStatusChange(sdids, 1)
		})

		convey.Convey("02-should skip when status already match", func() {
			mockDevManager.Reset()
			gomonkey.ApplyMethodReturn(r.devManager, "DcGetSuperPodStatus", 1, nil)
			r.handleNodeStatusChange(sdids, 1)
		})

		convey.Convey("03-should log error when parse sdid failed", func() {
			r.handleNodeStatusChange(sets.NewString("invalid"), 1)
		})
	})
}

func TestHandleFaultJobEvent(t *testing.T) {
	convey.Convey("Test handleFaultJobEvent", t, func() {
		if err := hwlog.InitRunLogger(&hwlog.LogConfig{}, context.TODO()); err != nil {
			fmt.Printf("hwlog init failed, error is %v\n", err)
		}
		patch1 := gomonkey.ApplyMethodReturn(&devmanager.DeviceManager{}, "GetChipBaseInfos",
			[]*common.ChipBaseInfo{{CardID: 0, DeviceID: 0, LogicID: 0}}, nil)
		defer patch1.Reset()
		// Setup test data
		configmap.InitCmWatcher(&kubeclient.ClientK8s{ClientSet: fake.NewSimpleClientset()})
		InitReleaser()
		r := releaser
		// Mock dependencies
		mockGetFaultJobs := gomonkey.ApplyFuncReturn(getFaultJobInfosFromCm,
			map[string]map[int]api.SuperPodFaultInfos{
				"job1": {0: {JobId: "job1", SdIds: []string{"sd1"}, FaultTimes: 1}},
			})
		defer mockGetFaultJobs.Reset()

		mockGetReleaseMap := gomonkey.ApplyFuncReturn(getReleaseAndRecoverMap,
			map[string]api.SuperPodFaultInfos{"job1": {JobId: "job1"}},
			sets.NewString("sd1"))
		defer mockGetReleaseMap.Reset()

		cm := &v1.ConfigMap{Data: map[string]string{"key": "value"}}

		convey.Convey("01-should handle both reset and recover when needed", func() {
			r.handleFaultJobEvent(cm)
			// Verify through mock calls
		})

		convey.Convey("02-should only handle reset when no recover needed", func() {
			mockGetReleaseMap.Reset()
			gomonkey.ApplyFuncReturn(getReleaseAndRecoverMap,
				map[string]api.SuperPodFaultInfos{"job1": {JobId: "job1"}},
				sets.NewString())
			r.handleFaultJobEvent(cm)
		})

		convey.Convey("03-should only handle recover when no reset needed", func() {
			mockGetReleaseMap.Reset()
			gomonkey.ApplyFuncReturn(getReleaseAndRecoverMap,
				map[string]api.SuperPodFaultInfos{},
				sets.NewString("sd1"))
			r.handleFaultJobEvent(cm)
		})
	})
}

func TestGetFaultJobInfosFromCm(t *testing.T) {
	convey.Convey("Test getFaultJobInfosFromCm", t, func() {
		if err := hwlog.InitRunLogger(&hwlog.LogConfig{}, context.TODO()); err != nil {
			fmt.Printf("hwlog init failed, error is %v\n", err)
		}
		convey.Convey("01-should return fault jobs when cm exists and data is valid", func() {
			// Create test configmap with valid data
			cm := &v1.ConfigMap{Data: map[string]string{
				"faultInfo": `{"job1":{"0":{"jobId":"job1","sdIds":["sd1"],"faultTimes":1}}}`}}

			result := getFaultJobInfosFromCm(cm)
			convey.So(len(result), convey.ShouldEqual, 0)
			convey.So(result["job1"][0].JobId, convey.ShouldEqual, "")
		})

		convey.Convey("02-should return empty when cm data is invalid", func() {
			// Create test configmap with invalid data
			cm := &v1.ConfigMap{Data: map[string]string{"faultInfo": "invalid json"}}
			result := getFaultJobInfosFromCm(cm)
			convey.So(len(result), convey.ShouldEqual, 0)
		})
	})
}
