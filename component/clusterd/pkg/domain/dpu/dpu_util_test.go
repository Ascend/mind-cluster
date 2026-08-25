// Copyright (c) Huawei Technologies Co., Ltd. 2026-2026. All rights reserved.

// Package dpu a series of dpu test function
package dpu

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
)

func mustMarshal(v interface{}) string {
	data, _ := json.Marshal(v)
	return string(data)
}

func init() {
	hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, context.Background())
}

func TestParseDpuInfoCM(t *testing.T) {
	convey.Convey("test parse dpu info cm", t, func() {
		convey.Convey("valid cm with dpu info", func() {
			dpuInfoCfg := constant.DpuInfoCfg{
				DPUInfo: constant.DpuInfoBody{
					DPUList: []constant.DpuItem{
						{DeviceID: "0x8200", HcaName: "mlx5_0", FaultList: []constant.DpuFaultDetail{
							{FaultCode: "21000023", FaultLevel: "SubHealth"},
						}},
					},
				},
				UpdateTime: 1234567890,
			}
			data := mustMarshal(dpuInfoCfg)
			cm := &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: "dpuinfo-node1"},
				Data:       map[string]string{DpuInfoCMDataKey: data},
			}
			info, err := ParseDpuInfoCM(cm)
			convey.So(err, convey.ShouldBeNil)
			convey.So(info.CmName, convey.ShouldEqual, "dpuinfo-node1")
			convey.So(info.UpdateTime, convey.ShouldEqual, 1234567890)
			convey.So(len(info.DPUInfo.DPUList), convey.ShouldEqual, 1)
		})

		convey.Convey("cm without dpu info key", func() {
			cm := &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: "dpuinfo-node2"},
				Data:       map[string]string{"other": "val"},
			}
			_, err := ParseDpuInfoCM(cm)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestDeepCopy(t *testing.T) {
	convey.Convey("test deep copy dpu info", t, func() {
		convey.Convey("copy non-nil dpu info", func() {
			original := &constant.DpuInfo{
				DpuInfoCfg: constant.DpuInfoCfg{
					DPUInfo: constant.DpuInfoBody{
						DPUList: []constant.DpuItem{
							{DeviceID: "0x8200", FaultList: []constant.DpuFaultDetail{
								{FaultCode: "21000023"},
							}},
						},
					},
					UpdateTime: 999,
				},
				CmName: "dpuinfo-node1",
			}
			copied := DeepCopy(original)
			convey.So(copied.CmName, convey.ShouldEqual, "dpuinfo-node1")
			convey.So(copied.DPUInfo.DPUList[0].DeviceID, convey.ShouldEqual, "0x8200")
			// modify copy, original should not change
			copied.DPUInfo.DPUList[0].DeviceID = "changed"
			convey.So(original.DPUInfo.DPUList[0].DeviceID, convey.ShouldEqual, "0x8200")
		})

		convey.Convey("copy nil", func() {
			convey.So(DeepCopy(nil), convey.ShouldBeNil)
		})
	})
}

func TestGetSafeData(t *testing.T) {
	convey.Convey("test get safe data", t, func() {
		convey.Convey("empty map returns empty slice", func() {
			result := GetSafeData(map[string]*constant.DpuInfo{})
			convey.So(len(result), convey.ShouldEqual, 0)
		})

		convey.Convey("single node returns one chunk", func() {
			dpuInfos := map[string]*constant.DpuInfo{
				"dpuinfo-node1": {
					DpuInfoCfg: constant.DpuInfoCfg{
						DPUInfo: constant.DpuInfoBody{
							DPUList: []constant.DpuItem{
								{DeviceID: "0x8200"},
							},
						},
						UpdateTime: 100,
					},
					CmName: "dpuinfo-node1",
				},
			}
			result := GetSafeData(dpuInfos)
			convey.So(len(result), convey.ShouldEqual, 1)
			convey.So(len(result[0]) > 0, convey.ShouldBeTrue)
		})
	})
}
