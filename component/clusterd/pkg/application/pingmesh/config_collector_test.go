// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package pingmesh a series of function handle ping mesh configmap create/update/delete
package pingmesh

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"clusterd/pkg/common/constant"
)

func TestIsNeedToStop(t *testing.T) {
	convey.Convey("Testing isNeedToStop", t, func() {
		convey.Convey("when newConfigInfo is nil, should return true", func() {
			convey.So(isNeedToStop(nil), convey.ShouldBeTrue)
		})
		convey.Convey("when activate is on, should return false", func() {
			newConfigInfo := constant.ConfigPingMesh{
				"1": &constant.HccspingMeshItem{
					Activate: "on",
				},
			}
			convey.So(isNeedToStop(newConfigInfo), convey.ShouldBeFalse)
		})
	})
}

func TestUpdatePingMeshConfigCM(t *testing.T) {
	convey.Convey("Testing updatePingMeshConfigCM", t, func() {
		convey.Convey("when newConfigInfo is nil, switch status should be off", func() {
			mock1 := gomonkey.ApplyFunc(isValidConfigPingMesh, func(_ constant.ConfigPingMesh) bool {
				return true
			})
			defer mock1.Reset()
			updatePingMeshConfigCM(nil)
			convey.So(rasNetDetectInst.CheckIsOn(), convey.ShouldBeFalse)
		})
		convey.Convey("when activate is on, should return false", func() {
			newConfigInfo := constant.ConfigPingMesh{
				"1": &constant.HccspingMeshItem{
					Activate: "on",
				},
			}
			mock1 := gomonkey.ApplyFunc(isValidConfigPingMesh, func(_ constant.ConfigPingMesh) bool {
				return true
			})
			defer mock1.Reset()
			updatePingMeshConfigCM(newConfigInfo)
			convey.So(rasNetDetectInst.CheckIsOn(), convey.ShouldBeTrue)
		})
	})
}

func TestGetConfigItemBySuperPodId(t *testing.T) {
	convey.Convey("Testing getConfigItemBySuperPodId", t, func() {
		convey.Convey("case 1: config info is nil. expect return err", func() {
			item, err := getConfigItemBySuperPodId(nil, testSuperPodID)
			convey.ShouldBeNil(item)
			convey.ShouldNotBeNil(err)
		})
		convey.Convey("case 2: super pod id exist. expect return nil", func() {
			info := constant.ConfigPingMesh{
				testSuperPodID: &constant.HccspingMeshItem{
					Activate:     "",
					TaskInterval: 0,
				},
			}
			item, err := getConfigItemBySuperPodId(info, testSuperPodID)
			convey.ShouldNotBeNil(item)
			convey.ShouldBeNil(err)
		})
		convey.Convey("case 3: super pod id not exist and global exist. expect return nil", func() {
			info := constant.ConfigPingMesh{
				constant.RasGlobalKey: &constant.HccspingMeshItem{
					Activate:     "",
					TaskInterval: 0,
				},
			}
			item, err := getConfigItemBySuperPodId(info, testSuperPodID)
			convey.ShouldNotBeNil(item)
			convey.ShouldBeNil(err)
		})
		convey.Convey("case 4: super pod id not exist and global not exist. expect return error", func() {
			info := constant.ConfigPingMesh{}
			item, err := getConfigItemBySuperPodId(info, testSuperPodID)
			convey.ShouldBeNil(item)
			convey.ShouldNotBeNil(err)
		})
	})
}

func TestConfigItemEqual(t *testing.T) {
	convey.Convey("Testing configItemEqual", t, func() {
		convey.Convey("case 1: both is nil. expect return true", func() {
			convey.ShouldBeTrue(configItemEqual(nil, nil))
		})
		convey.Convey("case 2: old is nil and new is not nil. expect return false", func() {
			convey.ShouldBeFalse(configItemEqual(nil, &constant.HccspingMeshItem{}))
		})
		convey.Convey("case 3: both old and new not nil", func() {
			convey.ShouldBeFalse(configItemEqual(&constant.HccspingMeshItem{
				Activate:     "on",
				TaskInterval: 0,
			}, &constant.HccspingMeshItem{
				Activate:     "on",
				TaskInterval: 0,
			}))
		})
	})
}
