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

package silentfault

import (
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"ascend-common/api/annotation"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/conf"
	"clusterd/pkg/domain/manualfault"
	domainpublicfault "clusterd/pkg/domain/publicfault"
	"clusterd/pkg/interface/kube"
)

func TestDeviceTypeFromName(t *testing.T) {
	convey.Convey("device type from name", t, func() {
		convey.So(deviceTypeFromName("Ascend910-0"), convey.ShouldEqual, "Ascend910")
		convey.So(deviceTypeFromName("npu-0"), convey.ShouldEqual, "npu")
		convey.So(deviceTypeFromName("0"), convey.ShouldEqual, "")
		convey.So(deviceTypeFromName("-0"), convey.ShouldEqual, "")
		convey.So(deviceTypeFromName(""), convey.ShouldEqual, "")
	})
}

func TestBuildDevNames(t *testing.T) {
	convey.Convey("build dev names", t, func() {
		convey.So(buildDevNames("Ascend910", []int32{0, 1}), convey.ShouldResemble,
			[]string{"Ascend910-0", "Ascend910-1"})
		convey.So(buildDevNames("", []int32{0}), convey.ShouldResemble, []string{"0"})
	})
}

func TestSilentDevTypeFromDetail(t *testing.T) {
	convey.Convey("silent dev type from detail", t, func() {
		info := manualfault.NodeCmInfo{Detail: map[string][]manualfault.DevCmInfo{
			"Ascend910-0": {{FaultLevel: constant.SilentFault}},
			"Ascend910-1": {{FaultLevel: constant.ManuallySeparateNPU}},
		}}
		convey.So(silentDevTypeFromDetail(info), convey.ShouldEqual, "Ascend910")
	})

	convey.Convey("no silent entry returns empty", t, func() {
		info := manualfault.NodeCmInfo{Detail: map[string][]manualfault.DevCmInfo{
			"Ascend910-1": {{FaultLevel: constant.ManuallySeparateNPU}},
		}}
		convey.So(silentDevTypeFromDetail(info), convey.ShouldEqual, "")
	})
}

func TestResolveDevTypeFromAnnotation(t *testing.T) {
	convey.Convey("resolve dev type from annotation", t, func() {
		p := gomonkey.ApplyFuncReturn(kube.GetNode, &corev1.Node{
			ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{
				annotation.NPUBaseDevInfosAnnotation: `{"Ascend910-0":{"DeviceID":0}}`,
			}},
		})
		defer p.Reset()
		convey.So(resolveDevTypeFromAnnotation("node1"), convey.ShouldEqual, "Ascend910")
	})

	convey.Convey("node nil returns empty", t, func() {
		p := gomonkey.ApplyFuncReturn(kube.GetNode, (*corev1.Node)(nil))
		defer p.Reset()
		convey.So(resolveDevTypeFromAnnotation("node1"), convey.ShouldEqual, "")
	})

	convey.Convey("invalid json returns empty", t, func() {
		p := gomonkey.ApplyFuncReturn(kube.GetNode, &corev1.Node{
			ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{
				annotation.NPUBaseDevInfosAnnotation: "not-json",
			}},
		})
		defer p.Reset()
		convey.So(resolveDevTypeFromAnnotation("node1"), convey.ShouldEqual, "")
	})
}

func TestLoadSilentDevTypeFromManualCm(t *testing.T) {
	convey.Convey("load silent dev type from manual cm", t, func() {
		cmInfo := map[string]manualfault.NodeCmInfo{
			"node1": {Detail: map[string][]manualfault.DevCmInfo{
				"Ascend910-0": {{FaultLevel: constant.SilentFault}},
			}},
			"node2": {Detail: map[string][]manualfault.DevCmInfo{
				"Ascend310-0": {{FaultLevel: constant.ManuallySeparateNPU}},
			}},
		}
		p1 := gomonkey.ApplyFuncReturn(manualfault.TryGetManualCm, &corev1.ConfigMap{}, nil)
		defer p1.Reset()
		p2 := gomonkey.ApplyFuncReturn(manualfault.ParseManualCm, cmInfo, nil)
		defer p2.Reset()

		res := loadSilentDevTypeFromManualCm()
		convey.So(res["node1"], convey.ShouldEqual, "Ascend910")
		convey.So(res["node2"], convey.ShouldEqual, "")
	})

	convey.Convey("load silent dev type get cm failed", t, func() {
		p1 := gomonkey.ApplyFuncReturn(manualfault.TryGetManualCm, nil, errors.New("boom"))
		defer p1.Reset()
		convey.So(loadSilentDevTypeFromManualCm(), convey.ShouldBeEmpty)
	})
}

func TestLoadSilentFaultCmInfo(t *testing.T) {
	convey.Convey("restore from manual cm dev type", t, func() {
		conf.SetSilentFaultPolicy(conf.SilentFaultPolicy{Enabled: true})
		ResetCache()

		cmInfo := map[string]manualfault.NodeCmInfo{
			"node1": {Detail: map[string][]manualfault.DevCmInfo{
				"Ascend910-0": {{FaultLevel: constant.SilentFault}},
			}},
		}
		p1 := gomonkey.ApplyFuncReturn(manualfault.TryGetManualCm, &corev1.ConfigMap{}, nil)
		defer p1.Reset()
		p2 := gomonkey.ApplyFuncReturn(manualfault.ParseManualCm, cmInfo, nil)
		defer p2.Reset()

		faults := map[string][]constant.NodeFault{
			"node1": {{FaultResource: constant.SilentFaultResource, FaultId: "silent-fault-node1",
				FaultLevel: constant.SilentFault, FaultDevIds: []int32{0, 1}, FaultTime: 100}},
		}
		p3 := gomonkey.ApplyMethodReturn(domainpublicfault.PubFaultCache, "GetPubFaultsForCM", faults, 1)
		defer p3.Reset()

		LoadSilentFaultCmInfo()
		info, ok := SilentFaultCmInfo.Get("node1")
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(info.DevList, convey.ShouldResemble, []string{"Ascend910-0", "Ascend910-1"})

		ResetCache()
	})

	convey.Convey("restore falls back to annotation dev type", t, func() {
		conf.SetSilentFaultPolicy(conf.SilentFaultPolicy{Enabled: true})
		ResetCache()

		p1 := gomonkey.ApplyFuncReturn(manualfault.TryGetManualCm, nil, errors.New("boom"))
		defer p1.Reset()
		p2 := gomonkey.ApplyFuncReturn(kube.GetNode, &corev1.Node{
			ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{
				annotation.NPUBaseDevInfosAnnotation: `{"Ascend910-0":{"DeviceID":0}}`,
			}},
		})
		defer p2.Reset()

		faults := map[string][]constant.NodeFault{
			"node1": {{FaultResource: constant.SilentFaultResource, FaultId: "silent-fault-node1",
				FaultLevel: constant.SilentFault, FaultDevIds: []int32{0}, FaultTime: 100}},
		}
		p3 := gomonkey.ApplyMethodReturn(domainpublicfault.PubFaultCache, "GetPubFaultsForCM", faults, 1)
		defer p3.Reset()

		LoadSilentFaultCmInfo()
		info, ok := SilentFaultCmInfo.Get("node1")
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(info.DevList, convey.ShouldResemble, []string{"Ascend910-0"})

		ResetCache()
	})

	convey.Convey("switch off skips restore", t, func() {
		conf.SetSilentFaultPolicy(conf.SilentFaultPolicy{Enabled: false})
		ResetCache()
		LoadSilentFaultCmInfo()
		convey.So(SilentFaultCmInfo.Len(), convey.ShouldEqual, 0)
	})
}
