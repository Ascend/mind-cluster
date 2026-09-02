/* Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
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
	"fmt"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"Ascend-device-plugin/pkg/common"
	"Ascend-device-plugin/pkg/device"
	"ascend-common/api"
	"ascend-common/api/annotation"
	"ascend-common/api/label"
	"ascend-common/devmanager"
	npuCommon "ascend-common/devmanager/common"
)

const (
	dieTypeVDIE = 1
	dieTypeDDIE = 2
)

const (
	labChipName910B  = "Ascend910B"
	labMemorySize8G  = 8192
	labBoardID0x28   = 0x28
	labBoardID0x29   = 0x29
	labDummyProdType = "dummy-product"
	labDriverVersion = "1.0.0"
	labAICoreCount   = 56
)

func newTestHwDevManager() *HwDevManager {
	return &HwDevManager{
		manager: device.NewHwAscend910Manager(),
		allInfo: common.NpuAllInfo{
			AllDevs: []common.NpuDevice{{LogicID: 0}},
		},
	}
}

func setupDMgrStub() *gomonkey.Patches {
	return gomonkey.ApplyMethod(
		reflect.TypeOf(new(device.HwAscend910Manager)),
		"GetDmgr",
		func(_ *device.HwAscend910Manager) devmanager.DeviceInterface {
			return &devmanager.DeviceManagerMock{}
		},
	)
}

func stubMgrMethod(name string, ret interface{}) *gomonkey.Patches {
	return gomonkey.ApplyMethodReturn(&device.HwAscend910Manager{}, name, ret)
}

func TestChipNameLabeler(t *testing.T) {
	hdm := newTestHwDevManager()
	patch := setupDMgrStub()
	defer patch.Reset()
	labeler := &chipNameLabeler{hdm: hdm}
	convey.Convey("should return error when get valid chip info fails", t, func() {
		stub := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{},
			"GetValidChipInfo", npuCommon.ChipInfo{}, fmt.Errorf("mock error"))
		defer stub.Reset()
		labels := make(map[string]string)
		err := labeler.Write(labels, &label.NodeContext{})
		convey.So(err, convey.ShouldNotBeNil)
	})
	convey.Convey("should write chip name labels when get valid chip info succeeds", t, func() {
		stub := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{},
			"GetValidChipInfo", npuCommon.ChipInfo{Name: labChipName910B}, nil)
		defer stub.Reset()
		labels := make(map[string]string)
		err := labeler.Write(labels, &label.NodeContext{})
		convey.So(err, convey.ShouldBeNil)
		convey.So(labels[label.NPUChipNameLabel], convey.ShouldEqual, labChipName910B)
		convey.So(labels[label.NPUChipNameLabelDeprecated], convey.ShouldEqual, labChipName910B)
	})
}

func TestChipMemoryLabeler(t *testing.T) {
	hdm := newTestHwDevManager()
	patch := setupDMgrStub()
	defer patch.Reset()
	labeler := &chipMemoryLabeler{hdm: hdm}
	convey.Convey("should skip when heterogeneous chips present", t, func() {
		labels := make(map[string]string)
		err := labeler.Write(labels, &label.NodeContext{IsHeterogeneous: true})
		convey.So(err, convey.ShouldBeNil)
		convey.So(labels, convey.ShouldBeEmpty)
	})
	convey.Convey("should skip when no on-chip memory", t, func() {
		stub := gomonkey.ApplyFuncReturn(common.HasOnChipMemory, false)
		defer stub.Reset()
		labels := make(map[string]string)
		err := labeler.Write(labels, &label.NodeContext{})
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("should skip when no devices found", t, func() {
		stub := gomonkey.ApplyFuncReturn(common.HasOnChipMemory, true)
		defer stub.Reset()
		hdmNoDev := &chipMemoryLabeler{hdm: &HwDevManager{
			manager: hdm.manager,
			allInfo: common.NpuAllInfo{AllDevs: nil},
		}}
		labels := make(map[string]string)
		err := hdmNoDev.Write(labels, &label.NodeContext{})
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("should write memory labels when HBM info succeeds", t, func() {
		stub1 := gomonkey.ApplyFuncReturn(common.HasOnChipMemory, true)
		defer stub1.Reset()
		stub2 := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{},
			"GetDeviceHbmInfo", &npuCommon.HbmInfo{MemorySize: labMemorySize8G}, nil)
		defer stub2.Reset()
		labels := make(map[string]string)
		err := labeler.Write(labels, &label.NodeContext{})
		convey.So(err, convey.ShouldBeNil)
		convey.So(labels[label.NPUChipMemoryLabel], convey.ShouldEqual, "8G")
	})
}

func TestChipBoardIDLabeler(t *testing.T) {
	hdm := newTestHwDevManager()
	patch := setupDMgrStub()
	defer patch.Reset()
	labeler := &chipBoardIDLabeler{hdm: hdm}
	convey.Convey("should skip when heterogeneous chips present", t, func() {
		labels := make(map[string]string)
		err := labeler.Write(labels, &label.NodeContext{IsHeterogeneous: true})
		convey.So(err, convey.ShouldBeNil)
		convey.So(labels, convey.ShouldBeEmpty)
	})
	convey.Convey("should skip when empty board id", t, func() {
		stub := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{},
			"GetBoardInfo", npuCommon.BoardInfo{BoardId: common.EmptyBoardId}, nil)
		defer stub.Reset()
		labels := make(map[string]string)
		err := labeler.Write(labels, &label.NodeContext{})
		convey.So(err, convey.ShouldBeNil)
		convey.So(labels, convey.ShouldBeEmpty)
	})
	convey.Convey("should write board id label when board info succeeds", t, func() {
		stub := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{},
			"GetBoardInfo", npuCommon.BoardInfo{BoardId: labBoardID0x28}, nil)
		defer stub.Reset()
		labels := make(map[string]string)
		err := labeler.Write(labels, &label.NodeContext{})
		convey.So(err, convey.ShouldBeNil)
		convey.So(labels[label.NPUChipBoardIDLabel], convey.ShouldEqual, "0x28")
	})
}

func TestChipProductTypeLabeler(t *testing.T) {
	hdm := newTestHwDevManager()
	labeler := &chipProductTypeLabeler{hdm: hdm}
	convey.Convey("should skip when heterogeneous chips present", t, func() {
		labels := make(map[string]string)
		err := labeler.Write(labels, &label.NodeContext{IsHeterogeneous: true})
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("should skip when product type is empty", t, func() {
		stub := gomonkey.ApplyPrivateMethod(hdm, "getProductType",
			func(_ *HwDevManager) string { return "" })
		defer stub.Reset()
		labels := make(map[string]string)
		err := labeler.Write(labels, &label.NodeContext{})
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("should write product type label when succeeds", t, func() {
		stub1 := gomonkey.ApplyPrivateMethod(hdm, "getProductType",
			func(_ *HwDevManager) string { return labDummyProdType })
		defer stub1.Reset()
		stub2 := gomonkey.ApplyFuncReturn(common.IsContainAll300IDuo, false)
		defer stub2.Reset()
		labels := make(map[string]string)
		err := labeler.Write(labels, &label.NodeContext{})
		convey.So(err, convey.ShouldBeNil)
		convey.So(labels[label.NPUChipProductTypeLabel], convey.ShouldEqual, labDummyProdType)
	})
	convey.Convey("should write deprecated label for All300IDuo", t, func() {
		stub1 := gomonkey.ApplyPrivateMethod(hdm, "getProductType",
			func(_ *HwDevManager) string { return labDummyProdType })
		defer stub1.Reset()
		stub2 := gomonkey.ApplyFuncReturn(common.IsContainAll300IDuo, true)
		defer stub2.Reset()
		labels := make(map[string]string)
		err := labeler.Write(labels, &label.NodeContext{})
		convey.So(err, convey.ShouldBeNil)
		convey.So(labels[label.NPUChipProductTypeLabelDeprecated], convey.ShouldEqual, api.A300IDuoLabel)
	})
}

func TestAcceleratorTypeLabeler(t *testing.T) {
	hdm := newTestHwDevManager()
	patch := setupDMgrStub()
	defer patch.Reset()
	labeler := &acceleratorTypeLabeler{hdm: hdm}
	convey.Convey("should skip when heterogeneous chips present", t, func() {
		labels := make(map[string]string)
		err := labeler.Write(labels, &label.NodeContext{IsHeterogeneous: true})
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("should skip when card type is not 910B", t, func() {
		common.ParamOption.RealCardType = api.Ascend910A5
		labels := make(map[string]string)
		err := labeler.Write(labels, &label.NodeContext{})
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("should write A300I label when board id matches", t, func() {
		common.ParamOption.RealCardType = api.Ascend910B
		stub := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{},
			"GetBoardInfo", npuCommon.BoardInfo{BoardId: labBoardID0x28}, nil)
		defer stub.Reset()
		labels := make(map[string]string)
		err := labeler.Write(labels, &label.NodeContext{})
		convey.So(err, convey.ShouldBeNil)
		convey.So(labels[label.AcceleratorTypeKeyDeprecated], convey.ShouldEqual, label.A300IA2Label)
	})
	convey.Convey("should write A300I label when board id is 64GB", t, func() {
		common.ParamOption.RealCardType = api.Ascend910B
		stub := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{},
			"GetBoardInfo", npuCommon.BoardInfo{BoardId: labBoardID0x29}, nil)
		defer stub.Reset()
		labels := make(map[string]string)
		err := labeler.Write(labels, &label.NodeContext{})
		convey.So(err, convey.ShouldBeNil)
		convey.So(labels[label.AcceleratorTypeKeyDeprecated], convey.ShouldEqual, label.A300IA2Label)
	})
}

func TestServerTypeLabeler(t *testing.T) {
	hdm := newTestHwDevManager()
	labeler := &serverTypeLabeler{hdm: hdm}
	convey.Convey("should write new prefix when device is not old type", t, func() {
		common.ParamOption.RealCardType = api.Ascend910A5
		common.ParamOption.AiCoreCount = labAICoreCount
		labels := make(map[string]string)
		err := labeler.Write(labels, &label.NodeContext{})
		convey.So(err, convey.ShouldBeNil)
		convey.So(labels[label.NPUServerTypeLabel], convey.ShouldContainSubstring, "npu-")
	})
	convey.Convey("should write labels when old device type and labels missing", t, func() {
		common.ParamOption.RealCardType = api.Ascend910
		common.ParamOption.AiCoreCount = labAICoreCount
		labels := make(map[string]string)
		err := labeler.Write(labels, &label.NodeContext{
			Node: &v1.Node{},
		})
		convey.So(err, convey.ShouldBeNil)
		convey.So(labels, convey.ShouldNotBeEmpty)
	})
}

func TestDriverVersionLabeler(t *testing.T) {
	hdm := newTestHwDevManager()
	patch := setupDMgrStub()
	defer patch.Reset()
	labeler := &driverVersionLabeler{hdm: hdm}
	convey.Convey("should skip when driver version is empty", t, func() {
		stub := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{},
			"GetDcmiVersion", "")
		defer stub.Reset()
		labels := make(map[string]string)
		err := labeler.Write(labels, &label.NodeContext{})
		convey.So(err, convey.ShouldBeNil)
		convey.So(labels, convey.ShouldBeEmpty)
	})
	convey.Convey("should write driver version labels when succeeds", t, func() {
		stub := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{},
			"GetDcmiVersion", labDriverVersion)
		defer stub.Reset()
		labels := make(map[string]string)
		err := labeler.Write(labels, &label.NodeContext{})
		convey.So(err, convey.ShouldBeNil)
		convey.So(labels[label.NPUDriverVersionLabel], convey.ShouldEqual, labDriverVersion)
		convey.So(labels[label.NPUDriverVersionLabelDeprecated], convey.ShouldEqual, labDriverVersion)
	})
}

func TestAcceleratorLabeler(t *testing.T) {
	hdm := newTestHwDevManager()
	labeler := &acceleratorLabeler{hdm: hdm}
	convey.Convey("should write accelerator label when card type in map", t, func() {
		common.ParamOption.RealCardType = api.Ascend910
		labels := make(map[string]string)
		err := labeler.Write(labels, &label.NodeContext{})
		convey.So(err, convey.ShouldBeNil)
		convey.So(labels[label.AcceleratorLabelKeyDeprecated], convey.ShouldNotBeEmpty)
	})
	convey.Convey("should skip when card type not in map", t, func() {
		common.ParamOption.RealCardType = "unknown"
		labels := make(map[string]string)
		err := labeler.Write(labels, &label.NodeContext{})
		convey.So(err, convey.ShouldBeNil)
		convey.So(labels, convey.ShouldBeEmpty)
	})
}

func TestTopologyLabeler(t *testing.T) {
	hdm := newTestHwDevManager()
	labeler := &topologyLabeler{hdm: hdm}
	convey.Convey("should write superpod id for A3 device", t, func() {
		common.ParamOption.RealCardType = api.Ascend910A3
		stub := stubMgrMethod("GetSuperPodID", int32(1))
		defer stub.Reset()
		labels := make(map[string]string)
		err := labeler.Write(labels, &label.NodeContext{})
		convey.So(err, convey.ShouldBeNil)
		convey.So(labels[label.TopoLabelSuperPodId], convey.ShouldEqual, "1")
	})
	convey.Convey("should skip superpod id for A3 when invalid", t, func() {
		common.ParamOption.RealCardType = api.Ascend910A3
		stub := stubMgrMethod("GetSuperPodID", int32(-1))
		defer stub.Reset()
		labels := make(map[string]string)
		err := labeler.Write(labels, &label.NodeContext{})
		convey.So(err, convey.ShouldBeNil)
		convey.So(labels, convey.ShouldBeEmpty)
	})
	convey.Convey("should write all topology labels for A5 1D device", t, func() {
		common.ParamOption.RealCardType = api.Ascend910A5
		stub1 := stubMgrMethod("GetSuperPodID", int32(2))
		defer stub1.Reset()
		stub2 := stubMgrMethod("GetSuperPodType", int32(common.ProductType1D))
		defer stub2.Reset()
		stub3 := stubMgrMethod("GetRackID", int32(3))
		defer stub3.Reset()
		stub4 := stubMgrMethod("GetServerIndex", int32(4))
		defer stub4.Reset()
		labels := make(map[string]string)
		err := labeler.Write(labels, &label.NodeContext{})
		convey.So(err, convey.ShouldBeNil)
		convey.So(labels[label.TopoLabelSuperPodId], convey.ShouldEqual, "2")
		convey.So(labels[label.TopoLabelRackId], convey.ShouldEqual, "3")
		convey.So(labels[label.TopoLabelServerId], convey.ShouldEqual, "4")
	})
	convey.Convey("should skip rack id for A5 non-1D/2D device", t, func() {
		common.ParamOption.RealCardType = api.Ascend910A5
		stub1 := stubMgrMethod("GetSuperPodID", int32(2))
		defer stub1.Reset()
		stub2 := stubMgrMethod("GetSuperPodType", int32(0))
		defer stub2.Reset()
		stub3 := stubMgrMethod("GetServerIndex", int32(4))
		defer stub3.Reset()
		labels := make(map[string]string)
		err := labeler.Write(labels, &label.NodeContext{})
		convey.So(err, convey.ShouldBeNil)
		convey.So(labels[label.TopoLabelSuperPodId], convey.ShouldEqual, "2")
		convey.So(labels[label.TopoLabelServerId], convey.ShouldEqual, "4")
	})
}

func TestBaseInfoAnnotator(t *testing.T) {
	hdm := newTestHwDevManager()
	annotator := &baseInfoAnnotator{hdm: hdm}
	convey.Convey("should write base info annotations when succeeds", t, func() {
		stub := gomonkey.ApplyPrivateMethod(hdm, "getNpuBaseInfo",
			func(_ *HwDevManager) map[string]*common.NpuBaseInfo {
				return map[string]*common.NpuBaseInfo{"0": {}}
			})
		defer stub.Reset()
		annotations := make(map[string]string)
		err := annotator.Write(annotations, &label.NodeContext{})
		convey.So(err, convey.ShouldBeNil)
		convey.So(annotations[annotation.NPUBaseDevInfosAnnotation], convey.ShouldNotBeEmpty)
		convey.So(annotations[annotation.BaseDevInfoAnnoDeprecated], convey.ShouldNotBeEmpty)
	})
}

func TestVdieDdieAnnotator(t *testing.T) {
	hdm := newTestHwDevManager()
	annotator := &vdieDdieAnnotator{hdm: hdm}
	convey.Convey("should write vdie and ddie annotations", t, func() {
		stub := gomonkey.ApplyPrivateMethod(hdm, "getDieIDAnnotations",
			func(_ *HwDevManager) map[int32]map[string]string {
				return map[int32]map[string]string{
					dieTypeVDIE: {"0": "vdie-0"},
					dieTypeDDIE: {"0": "ddie-0"},
				}
			})
		defer stub.Reset()
		annotations := make(map[string]string)
		err := annotator.Write(annotations, &label.NodeContext{})
		convey.So(err, convey.ShouldBeNil)
		convey.So(annotations[annotation.NPUChipVdieIDsAnnotation], convey.ShouldNotBeEmpty)
		convey.So(annotations[annotation.NPUChipDdieIDsAnnotation], convey.ShouldNotBeEmpty)
	})
	convey.Convey("should skip when die ids are empty", t, func() {
		stub := gomonkey.ApplyPrivateMethod(hdm, "getDieIDAnnotations",
			func(_ *HwDevManager) map[int32]map[string]string {
				return map[int32]map[string]string{}
			})
		defer stub.Reset()
		annotations := make(map[string]string)
		err := annotator.Write(annotations, &label.NodeContext{})
		convey.So(err, convey.ShouldBeNil)
		convey.So(annotations, convey.ShouldBeEmpty)
	})
}

func TestSerialNumberAnnotator(t *testing.T) {
	hdm := newTestHwDevManager()
	annotator := &serialNumberAnnotator{hdm: hdm}
	convey.Convey("should write serial number annotation", t, func() {
		stub := gomonkey.ApplyPrivateMethod(hdm, "getChipSerialNumbers",
			func(_ *HwDevManager) map[string]string { return map[string]string{"0": "SN123"} })
		defer stub.Reset()
		annotations := make(map[string]string)
		err := annotator.Write(annotations, &label.NodeContext{})
		convey.So(err, convey.ShouldBeNil)
		convey.So(annotations[annotation.NPUChipSerialNumberAnnotation], convey.ShouldNotBeEmpty)
	})
	convey.Convey("should skip when serial numbers are empty", t, func() {
		stub := gomonkey.ApplyPrivateMethod(hdm, "getChipSerialNumbers",
			func(_ *HwDevManager) map[string]string { return map[string]string{} })
		defer stub.Reset()
		annotations := make(map[string]string)
		err := annotator.Write(annotations, &label.NodeContext{})
		convey.So(err, convey.ShouldBeNil)
		convey.So(annotations, convey.ShouldBeEmpty)
	})
}

func TestServerTypeAnnotator(t *testing.T) {
	hdm := newTestHwDevManager()
	annotator := &serverTypeAnnotator{hdm: hdm}
	convey.Convey("should write server type annotation", t, func() {
		common.ParamOption.RealCardType = api.Ascend910A5
		annotations := make(map[string]string)
		err := annotator.Write(annotations, &label.NodeContext{})
		convey.So(err, convey.ShouldBeNil)
		convey.So(annotations[annotation.ServerTypeKeyDeprecated], convey.ShouldNotBeEmpty)
	})
}

func TestSuperPodInfoAnnotator(t *testing.T) {
	hdm := newTestHwDevManager()
	annotator := &superPodInfoAnnotator{hdm: hdm}
	convey.Convey("should write all annotations for A5 1D device", t, func() {
		common.ParamOption.RealCardType = api.Ascend910A5
		stub1 := stubMgrMethod("GetSuperPodID", int32(1))
		defer stub1.Reset()
		stub2 := stubMgrMethod("GetServerIndex", int32(2))
		defer stub2.Reset()
		stub3 := stubMgrMethod("GetSuperPodType", int32(common.ProductType1D))
		defer stub3.Reset()
		stub4 := stubMgrMethod("GetRackID", int32(3))
		defer stub4.Reset()
		annotations := make(map[string]string)
		err := annotator.Write(annotations, &label.NodeContext{})
		convey.So(err, convey.ShouldBeNil)
		convey.So(annotations[annotation.SuperPodIDKeyDeprecated], convey.ShouldEqual, "1")
		convey.So(annotations[annotation.ServerIndexKeyDeprecated], convey.ShouldEqual, "2")
		convey.So(annotations[annotation.RackIDKeyDeprecated], convey.ShouldEqual, "3")
	})
	convey.Convey("should skip rack ID for non-A5 device", t, func() {
		common.ParamOption.RealCardType = api.Ascend910
		stub1 := stubMgrMethod("GetSuperPodID", int32(1))
		defer stub1.Reset()
		stub2 := stubMgrMethod("GetServerIndex", int32(2))
		defer stub2.Reset()
		annotations := make(map[string]string)
		err := annotator.Write(annotations, &label.NodeContext{})
		convey.So(err, convey.ShouldBeNil)
		convey.So(annotations[annotation.SuperPodIDKeyDeprecated], convey.ShouldEqual, "1")
		convey.So(annotations[annotation.ServerIndexKeyDeprecated], convey.ShouldEqual, "2")
	})
}

func TestTopologyAnnotator(t *testing.T) {
	hdm := newTestHwDevManager()
	patch := setupDMgrStub()
	defer patch.Reset()
	anno := &topologyAnnotator{hdm: hdm}

	convey.Convey("skips heterogeneous nodes", t, func() {
		stub := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{},
			"GetNodeTopo", npuCommon.TopoDefault8Chip)
		defer stub.Reset()
		annotations := make(map[string]string)
		err := anno.Write(annotations, &label.NodeContext{Node: &v1.Node{}, IsHeterogeneous: true})
		convey.So(err, convey.ShouldBeNil)
		convey.So(annotations, convey.ShouldBeEmpty)
	})

	convey.Convey("keeps an existing manually-set npu.topology untouched", t, func() {
		stub := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{},
			"GetNodeTopo", npuCommon.TopoDefault8Chip)
		defer stub.Reset()
		annotations := make(map[string]string)
		ctx := &label.NodeContext{Node: &v1.Node{ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{annotation.NPUTopologyAnnotation: "[[0,1]]"},
		}}}
		err := anno.Write(annotations, ctx)
		convey.So(err, convey.ShouldBeNil)
		convey.So(annotations, convey.ShouldBeEmpty)
		convey.So(ctx.Node.Annotations[annotation.NPUTopologyAnnotation], convey.ShouldEqual, "[[0,1]]")
	})

	convey.Convey("writes npu.topology on lookup hit", t, func() {
		stub := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{},
			"GetNodeTopo", npuCommon.TopoDefault8Chip)
		defer stub.Reset()
		annotations := make(map[string]string)
		err := anno.Write(annotations, &label.NodeContext{Node: &v1.Node{}})
		convey.So(err, convey.ShouldBeNil)
		convey.So(annotations[annotation.NPUTopologyAnnotation], convey.ShouldEqual, npuCommon.TopoDefault8Chip)
	})

	convey.Convey("skips when the topo lookup misses", t, func() {
		stub := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{},
			"GetNodeTopo", "")
		defer stub.Reset()
		annotations := make(map[string]string)
		err := anno.Write(annotations, &label.NodeContext{Node: &v1.Node{}})
		convey.So(err, convey.ShouldBeNil)
		convey.So(annotations, convey.ShouldBeEmpty)
	})
}

func TestCardTypeAnnotator(t *testing.T) {
	hdm := newTestHwDevManager()
	annotator := &cardTypeAnnotator{hdm: hdm}
	convey.Convey("should write card type annotation when succeeds", t, func() {
		stub := gomonkey.ApplyPrivateMethod(hdm, "getCardType",
			func(_ *HwDevManager) (string, error) { return "A300I", nil })
		defer stub.Reset()
		annotations := make(map[string]string)
		err := annotator.Write(annotations, &label.NodeContext{})
		convey.So(err, convey.ShouldBeNil)
		convey.So(annotations[annotation.CardTypeKeyDeprecated], convey.ShouldEqual, "A300I")
	})
	convey.Convey("should skip when card type is empty", t, func() {
		stub := gomonkey.ApplyPrivateMethod(hdm, "getCardType",
			func(_ *HwDevManager) (string, error) { return "", nil })
		defer stub.Reset()
		annotations := make(map[string]string)
		err := annotator.Write(annotations, &label.NodeContext{})
		convey.So(err, convey.ShouldBeNil)
		convey.So(annotations, convey.ShouldBeEmpty)
	})
	convey.Convey("should skip when getCardType returns error", t, func() {
		stub := gomonkey.ApplyPrivateMethod(hdm, "getCardType",
			func(_ *HwDevManager) (string, error) { return "", fmt.Errorf("mock error") })
		defer stub.Reset()
		annotations := make(map[string]string)
		err := annotator.Write(annotations, &label.NodeContext{})
		convey.So(err, convey.ShouldBeNil)
		convey.So(annotations, convey.ShouldBeEmpty)
	})
}

func TestWriteValue(t *testing.T) {
	convey.Convey("should skip when value is empty", t, func() {
		m := make(map[string]string)
		writeValue(m, "", "key1", "key2")
		convey.So(m, convey.ShouldBeEmpty)
	})
	convey.Convey("should write to all keys when value is not empty", t, func() {
		m := make(map[string]string)
		writeValue(m, "val", "key1", "key2")
		convey.So(m["key1"], convey.ShouldEqual, "val")
		convey.So(m["key2"], convey.ShouldEqual, "val")
	})
}

func TestSanitizeLabelValue(t *testing.T) {
	convey.Convey("should keep valid label value unchanged", t, func() {
		result := sanitizeLabelValue("valid-label.value_123")
		convey.So(result, convey.ShouldEqual, "valid-label.value_123")
	})
	convey.Convey("should remove invalid characters", t, func() {
		result := sanitizeLabelValue("test@#$label")
		convey.So(result, convey.ShouldEqual, "testlabel")
	})
	convey.Convey("should replace multiple spaces with single dash", t, func() {
		result := sanitizeLabelValue("test  label  value")
		convey.So(result, convey.ShouldEqual, "test-label-value")
	})
}
