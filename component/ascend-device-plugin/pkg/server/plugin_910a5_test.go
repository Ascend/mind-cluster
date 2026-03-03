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

// Package server holds the implementation of registration to kubelet, k8s device plugin interface and grpc service.
package server

import (
	"context"
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	"Ascend-device-plugin/pkg/common"
	"Ascend-device-plugin/pkg/device"
	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-common/devmanager"
	commonv1 "ascend-common/devmanager/common"
)

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	if err := hwlog.InitRunLogger(&hwLogConfig, context.Background()); err != nil {
		return
	}
	common.ParamOption.PresetVDevice = true
}

// TestSetHcclTopoFilePathEnv_RealCardTypeNotA5 test set hccl topo file env (not a5)
func TestSetHcclTopoFilePathEnv_RealCardTypeNotA5(t *testing.T) {
	ps := NewPluginServer("1", devices, []string{common.HiAIManagerDevice},
		device.NewHwAscend910Manager())
	resp := &v1beta1.ContainerAllocateResponse{}
	convey.Convey("test setHcclTopoFilePathEnv  case 1 RealCardType is not A5", t, func() {
		realCardType := common.ParamOption.RealCardType
		common.ParamOption.RealCardType = api.Ascend910
		defer func() {
			common.ParamOption.RealCardType = realCardType
		}()
		ps.setHcclTopoFilePathEnv(resp, common.NpuAllInfo{})
		convey.So(len(resp.Envs), convey.ShouldEqual, 0)
	})
}

// TestSetHcclTopoFilePathEnv_AllDevsEmpty test set hccl topo file env (dev is empty)
func TestSetHcclTopoFilePathEnv_AllDevsEmpty(t *testing.T) {
	ps := NewPluginServer("1", devices, []string{common.HiAIManagerDevice},
		device.NewHwAscend910Manager())
	resp := &v1beta1.ContainerAllocateResponse{}
	realCardType := common.ParamOption.RealCardType
	common.ParamOption.RealCardType = api.Ascend910A5
	defer func() {
		common.ParamOption.RealCardType = realCardType
	}()
	convey.Convey("test setHcclTopoFilePathEnv case 2 AllDevs len is 0", t, func() {
		ps.setHcclTopoFilePathEnv(resp, common.NpuAllInfo{})
		convey.So(len(resp.Envs), convey.ShouldEqual, 0)
	})
}

// TestSetServerHcclTopoFilePathEnv test set hccl topo file env (success)
func TestSetServerHcclTopoFilePathEnv(t *testing.T) {
	ps := NewPluginServer("1", devices, []string{common.HiAIManagerDevice},
		device.NewHwAscend910Manager())
	resp := &v1beta1.ContainerAllocateResponse{}
	realCardType := common.ParamOption.RealCardType
	common.ParamOption.RealCardType = api.Ascend910A5
	defer func() {
		common.ParamOption.RealCardType = realCardType
	}()

	mockDmgr := gomonkey.ApplyMethodReturn(ps.manager, "GetDmgr", &devmanager.DeviceManager{})
	defer mockDmgr.Reset()
	convey.Convey("test setHcclTopoFilePathEnv case 3 GetSuperPodInfo error", t, func() {
		patch2 := gomonkey.ApplyMethodReturn(ps.manager.GetDmgr(), "GetSuperPodInfo",
			commonv1.CgoSuperPodInfo{}, errors.New("fake error"))
		defer patch2.Reset()
		ps.setHcclTopoFilePathEnv(resp, common.NpuAllInfo{})
		convey.So(len(resp.Envs), convey.ShouldEqual, 0)
	})
	allDevs := []common.NpuDevice{{LogicID: 0}}
	allNPUInfo := common.NpuAllInfo{AllDevs: allDevs}
	mockGetNPUs := gomonkey.ApplyMethodReturn(ps.manager, "GetNPUs", common.NpuAllInfo{AllDevs: allDevs}, nil)
	defer mockGetNPUs.Reset()
	convey.Convey("test setHcclTopoFilePathEnv case 4 superPodType illegal", t, func() {
		cardType := common.ParamOption.CardType
		common.ParamOption.CardType = "server"
		defer func() {
			common.ParamOption.CardType = cardType
		}()
		patch2 := gomonkey.ApplyMethodReturn(ps.manager.GetDmgr(), "GetSuperPodInfo",
			commonv1.CgoSuperPodInfo{SuperPodType: 10}, nil)
		defer patch2.Reset()
		ps.setHcclTopoFilePathEnv(resp, allNPUInfo)
		convey.So(len(resp.Envs), convey.ShouldEqual, 0)
	})
	convey.Convey("test setHcclTopoFilePathEnv case 5 set env ok", t, func() {
		cardType := common.ParamOption.CardType
		common.ParamOption.CardType = "server"
		defer func() {
			common.ParamOption.CardType = cardType
		}()
		patch2 := gomonkey.ApplyMethodReturn(ps.manager.GetDmgr(), "GetSuperPodInfo",
			commonv1.CgoSuperPodInfo{SuperPodType: 1}, nil)
		defer patch2.Reset()
		ps.setHcclTopoFilePathEnv(resp, allNPUInfo)
		convey.So(len(resp.Envs), convey.ShouldEqual, 1)
	})
}

// TestSetServerHcclTopoFilePathEnv test set hccl topo file env (standard card)
func TestSetCardHcclTopoFilePathEnv(t *testing.T) {
	ps := NewPluginServer("1", devices, []string{common.HiAIManagerDevice},
		device.NewHwAscend910Manager())
	resp := &v1beta1.ContainerAllocateResponse{}
	realCardType := common.ParamOption.RealCardType
	common.ParamOption.RealCardType = api.Ascend910A5
	defer func() {
		common.ParamOption.RealCardType = realCardType
	}()
	allDevs := []common.NpuDevice{{LogicID: 0}}
	allNPUInfo := common.NpuAllInfo{AllDevs: allDevs}
	convey.Convey("test setHcclTopoFilePathEnv case 6 standard card 300I-Atlas350", t, func() {
		cardType := common.ParamOption.CardType
		common.ParamOption.CardType = common.A5300ICardName
		defer func() {
			common.ParamOption.CardType = cardType
		}()

		ps.setHcclTopoFilePathEnv(resp, allNPUInfo)
		convey.So(len(resp.Envs), convey.ShouldEqual, 1)
		convey.So(resp.Envs[common.HcclTopoFilePathKey], convey.ShouldNotBeEmpty)
	})

	convey.Convey("test setHcclTopoFilePathEnv case 7 standard card 300I-Atlas350-4p", t, func() {
		cardType := common.ParamOption.CardType
		common.ParamOption.CardType = common.A54P300ICardName
		defer func() {
			common.ParamOption.CardType = cardType
		}()

		ps.setHcclTopoFilePathEnv(resp, allNPUInfo)
		convey.So(len(resp.Envs), convey.ShouldEqual, 1)
		convey.So(resp.Envs[common.HcclTopoFilePathKey], convey.ShouldNotBeEmpty)
	})
}
