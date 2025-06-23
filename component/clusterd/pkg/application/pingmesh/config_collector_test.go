// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package pingmesh a series of function handle ping mesh configmap create/update/delete
package pingmesh

import (
	"testing"

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
			updatePingMeshConfigCM(nil)
			convey.So(rasNetDetectInst.CheckIsOn(), convey.ShouldBeFalse)
		})
		convey.Convey("when activate is on, should return false", func() {
			newConfigInfo := constant.ConfigPingMesh{
				"1": &constant.HccspingMeshItem{
					Activate: "on",
				},
			}
			updatePingMeshConfigCM(newConfigInfo)
			convey.So(rasNetDetectInst.CheckIsOn(), convey.ShouldBeTrue)
		})
	})
}
