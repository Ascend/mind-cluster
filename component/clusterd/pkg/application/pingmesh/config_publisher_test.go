/*
Copyright(C)2025. Huawei Technologies Co.,Ltd. All rights reserved.

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

/*
Package pingmesh a series of function handle ping mesh configmap create/update/delete
*/
package pingmesh

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/superpod"
)

const (
	taskIntervalUnit = 10
)

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwlog.InitRunLogger(&hwLogConfig, context.Background())
}

func TestGetRasConfigBySuperPodId(t *testing.T) {
	convey.Convey("Test getRasConfigBySuperPodId", t, func() {
		cf := &ConfigPingMeshCmManager{}
		testSuperPodID := "test-pod-123"
		testInterval := 5
		testConfigCM := &constant.HccspingMeshItem{
			TaskInterval: testInterval,
			Activate:     constant.RasNetDetectOnStr,
		}
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		convey.Convey("When getConfigItemBySuperPodId succeeds", func() {
			patches.ApplyFunc(getConfigItemBySuperPodId, func(configInfo constant.ConfigPingMesh, superPodID string) (*constant.HccspingMeshItem, error) {
				convey.So(superPodID, convey.ShouldEqual, testSuperPodID)
				return testConfigCM, nil
			})
			result := cf.getRasConfigBySuperPodId(testSuperPodID)
			convey.Convey("Should return correct config", func() {
				convey.So(result, convey.ShouldNotBeNil)
				convey.So(result.Period, convey.ShouldEqual, testInterval*taskIntervalUnit)
				convey.So(result.NetFault, convey.ShouldEqual, constant.RasNetDetectOnStr)
			})
		})
		convey.Convey("When getConfigItemBySuperPodId fails", func() {
			expectedErr := fmt.Errorf("config not found")
			patches.ApplyFunc(getConfigItemBySuperPodId, func(configInfo constant.ConfigPingMesh, superPodID string) (*constant.HccspingMeshItem, error) {
				return nil, expectedErr
			})
			result := cf.getRasConfigBySuperPodId(testSuperPodID)
			convey.Convey("Should return default config and log error", func() {
				convey.So(result, convey.ShouldNotBeNil)
				convey.So(result, convey.ShouldPointTo, &rasConfig)
			})
		})
		convey.Convey("When Activate is not ON", func() {
			offConfigCM := &constant.HccspingMeshItem{
				TaskInterval: testInterval,
				Activate:     "off",
			}
			patches.ApplyFunc(getConfigItemBySuperPodId, func(configInfo constant.ConfigPingMesh, superPodID string) (*constant.HccspingMeshItem, error) {
				return offConfigCM, nil
			})
			result := cf.getRasConfigBySuperPodId(testSuperPodID)
			convey.Convey("Should return config with NetFault off", func() {
				convey.So(result, convey.ShouldNotBeNil)
				convey.So(result.NetFault, convey.ShouldNotEqual, constant.RasNetDetectOnStr)
			})
		})
	})
}

func TestUpdateConfigData(t *testing.T) {
	convey.Convey("Test UpdateConfigData", t, func() {
		cpm := &ConfigPingMeshCmManager{
			RWMutex: sync.RWMutex{},
			cacheStatus: &constant.CacheStatus{
				Inited: false,
			},
		}
		configCMInfo := constant.ConfigPingMesh{
			"1": nil,
		}
		cpm.UpdateConfigData(configCMInfo)
		convey.So(len(cpm.configCMInfo), convey.ShouldEqual, 1)
	})
}

func TestUpdateConfigFileWhenCmUpdated(t *testing.T) {
	convey.Convey("Test UpdateConfigFileWhenCmUpdated case 1", t, func() {
		cf := &ConfigPingMeshCmManager{}
		mock1 := gomonkey.ApplyFunc((*RasNetFaultCmManager).CheckIsOn, func(_ *RasNetFaultCmManager) bool {
			return false
		})
		defer mock1.Reset()
		cf.updateConfigFileWhenCmUpdated()
	})

	convey.Convey("Test UpdateConfigFileWhenCmUpdated case 1", t, func() {
		cf := &ConfigPingMeshCmManager{}
		mock1 := gomonkey.ApplyFunc((*RasNetFaultCmManager).CheckIsOn, func(_ *RasNetFaultCmManager) bool {
			return false
		})
		defer mock1.Reset()
		mock2 := gomonkey.ApplyFunc(superpod.ListClusterDevice, func() []*api.SuperPodDevice {
			device := &api.SuperPodDevice{
				SuperPodID: "1",
			}
			return []*api.SuperPodDevice{device}
		})
		defer mock2.Reset()
		mock3 := gomonkey.ApplyFunc((*ConfigPingMeshCmManager).getRasConfigBySuperPodId,
			func(_ *ConfigPingMeshCmManager, _ string) *constant.CathelperConf { return nil })
		defer mock3.Reset()
		mock4 := gomonkey.ApplyFunc(handlerSuperPodRoce, func(_ map[int]string) {})
		defer mock4.Reset()
		mock5 := gomonkey.ApplyFunc(superpod.GetAllSuperPodIDWithAcceleratorType, func() map[int]string {
			return make(map[int]string)
		})
		defer mock5.Reset()
		cf.updateConfigFileWhenCmUpdated()
	})

}
