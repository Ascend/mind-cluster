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

// Package server holds the implementation of registration to kubelet, k8s pod resource interface.
package server

import (
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	"Ascend-device-plugin/pkg/common"
	"Ascend-device-plugin/pkg/device/dpucontrol"
	"ascend-common/api"
	"ascend-common/common-utils/ethtool"
)

func TestUpdateDpuHealthy(t *testing.T) {
	tests := []npuFaultTestCase{
		buildFilterDpuFaultTestCase1(),
		buildFilterDpuFaultTestCase2(),
		buildFilterDpuFaultTestCase3(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hdm := &HwDevManager{
				dpuManager: &dpucontrol.DpuFilter{
					NpuWithDpuInfos: tt.npuWithDpuInfos,
				},
			}
			groupDevice := map[string][]*common.NpuDevice{
				"Ascend910": tt.deviceList,
			}
			outputs := getMockOutputForEthState(tt.npuWithDpuInfos)
			patchEth := gomonkey.ApplyFunc(ethtool.GetInterfaceOperState, func(ifaceName string) (string, error) {
				return outputs[ifaceName], nil
			})
			hdm.updateDpuHealthy(groupDevice)
			for i, dev := range tt.deviceList {
				if dpuHealthy := dev.DpuHealth; dpuHealthy != tt.wantDpuHealth[i] {
					t.Errorf("DpuHealthy = %v, want %v, for dev %v", dpuHealthy, tt.wantDpuHealth[i], dev.CardID)
				}
			}
			patchEth.Reset()
		})
	}
}

func getMockOutputForEthState(npuWithDpuInfos []dpucontrol.NpuWithDpuInfo) map[string]string {
	uniqueMap := make(map[string]string)
	for _, dpuInfos := range npuWithDpuInfos {
		for _, dpu := range dpuInfos.DpuInfo {
			if _, exist := uniqueMap[dpu.DeviceName]; !exist {
				uniqueMap[dpu.DeviceName] = dpu.Operstate
				fmt.Println(dpu)
			}
		}
	}
	return uniqueMap
}

type npuFaultTestCase struct {
	name            string
	npuWithDpuInfos []dpucontrol.NpuWithDpuInfo
	deviceList      []*common.NpuDevice
	wantDpuHealth   []string
}

func buildFilterDpuFaultTestCase1() npuFaultTestCase {
	dpuInfo1 := dpucontrol.BaseDpuInfo{
		DeviceName: "enps0", Operstate: "down", DeviceId: "0x1825", Vendor: "0x10b5",
	}
	dpuInfo2 := dpucontrol.BaseDpuInfo{
		DeviceName: "enps2", Operstate: "down", DeviceId: "0x1825", Vendor: "0x10b5",
	}
	return npuFaultTestCase{
		name: "No DPU state is up, DpuHealth is unhealthy",
		npuWithDpuInfos: []dpucontrol.NpuWithDpuInfo{
			{
				NpuId:   int32(0),
				DpuInfo: []dpucontrol.BaseDpuInfo{dpuInfo1, dpuInfo2},
			},
			{
				NpuId:   int32(1),
				DpuInfo: []dpucontrol.BaseDpuInfo{dpuInfo1, dpuInfo2},
			},
		},
		deviceList: []*common.NpuDevice{
			{
				DeviceName: "Ascend910-0",
				CardID:     int32(0),
				DpuHealth:  "",
			},
			{
				DeviceName: "Ascend910-1",
				CardID:     int32(1),
				DpuHealth:  "",
			},
		},
		wantDpuHealth: []string{v1beta1.Unhealthy, v1beta1.Unhealthy},
	}
}

func buildFilterDpuFaultTestCase2() npuFaultTestCase {
	dpuInfo1 := dpucontrol.BaseDpuInfo{
		DeviceName: "enps0", Operstate: "up", DeviceId: "0x1825", Vendor: "0x10b5",
	}
	dpuInfo2 := dpucontrol.BaseDpuInfo{
		DeviceName: "enps2", Operstate: "down", DeviceId: "0x1825", Vendor: "0x10b5",
	}
	return npuFaultTestCase{
		name: "[UB type] If one of the two DPUs is in the up state, it is Subhealthy.",
		npuWithDpuInfos: []dpucontrol.NpuWithDpuInfo{
			{
				NpuId:   int32(0),
				DpuInfo: []dpucontrol.BaseDpuInfo{dpuInfo1, dpuInfo2},
			},
			{
				NpuId:   int32(1),
				DpuInfo: []dpucontrol.BaseDpuInfo{dpuInfo1, dpuInfo2},
			},
		},
		deviceList: []*common.NpuDevice{
			{
				DeviceName: "Ascend910-0",
				CardID:     int32(0),
				DpuHealth:  "",
			},
			{
				DeviceName: "Ascend910-1",
				CardID:     int32(1),
				DpuHealth:  "",
			},
		},
		wantDpuHealth: []string{api.DpuSubHealthy, api.DpuSubHealthy},
	}
}

func buildFilterDpuFaultTestCase3() npuFaultTestCase {
	dpuInfo1 := dpucontrol.BaseDpuInfo{
		DeviceName: "eth0", Operstate: "up", DeviceId: "0x1019", Vendor: "0x15b3",
	}
	dpuInfo2 := dpucontrol.BaseDpuInfo{
		DeviceName: "eth1", Operstate: "down", DeviceId: "0x1019", Vendor: "0x15b3",
	}
	return npuFaultTestCase{
		name: "[PCIe type] dpu is in the up state, it is healthy.",
		npuWithDpuInfos: []dpucontrol.NpuWithDpuInfo{
			{
				NpuId:   int32(0),
				DpuInfo: []dpucontrol.BaseDpuInfo{dpuInfo1},
			},
			{
				NpuId:   int32(1),
				DpuInfo: []dpucontrol.BaseDpuInfo{dpuInfo2},
			},
		},
		deviceList: []*common.NpuDevice{
			{
				DeviceName: "Ascend910-0",
				CardID:     int32(0),
				DpuHealth:  "",
			},
			{
				DeviceName: "Ascend910-1",
				CardID:     int32(1),
				DpuHealth:  "",
			},
		},
		wantDpuHealth: []string{v1beta1.Healthy, v1beta1.Unhealthy},
	}
}

func TestGetDpuFaultInfoOfNpu(t *testing.T) {
	tests := []npuFaultByDpuCase{
		buildGetDpuFaultInfoOfNpuCase01(),
		buildGetDpuFaultInfoOfNpuCase02(),
		buildGetDpuFaultInfoOfNpuCase03(),
		buildGetDpuFaultInfoOfNpuCase04(),
		buildGetDpuFaultInfoOfNpuCase05(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if dpuHealth := getDpuFaultInfoOfNpu(tt.npuName, tt.npuToDpuMap, tt.dpuOperstateMap); dpuHealth != tt.want {
				t.Errorf("getDpuFaultInfoOfNpu() = %v, want %v", dpuHealth, tt.want)
			}
		})
	}
}

type npuFaultByDpuCase struct {
	name            string
	npuName         string
	npuToDpuMap     map[string][]string
	dpuOperstateMap map[string]string
	want            string
}

func buildGetDpuFaultInfoOfNpuCase01() npuFaultByDpuCase {
	return npuFaultByDpuCase{
		name:    "[UB-type]the two DPUs is in up state, the DpuUnhealth is Health",
		npuName: "Ascend910-0",
		npuToDpuMap: map[string][]string{
			"0": {"enps0", "enps2"},
			"1": {"enps0", "enps2"},
			"4": {"enps1", "enps3"},
			"5": {"enps1", "enps3"},
		},
		dpuOperstateMap: map[string]string{
			"enps0": "up",
			"enps1": "up",
			"enps2": "up",
			"enps3": "up",
		},
		want: "Healthy",
	}
}

func buildGetDpuFaultInfoOfNpuCase02() npuFaultByDpuCase {
	return npuFaultByDpuCase{
		name:    "[UB-type]one dpu of the two is in up state, the DpuUnhealth is SubHealth",
		npuName: "Ascend910-8", // equal to the Ascend910-0
		npuToDpuMap: map[string][]string{
			"0": {"enps0", "enps2"},
			"1": {"enps0", "enps2"},
			"4": {"enps1", "enps3"},
			"5": {"enps1", "enps3"},
		},
		dpuOperstateMap: map[string]string{
			"enps0": "up",
			"enps1": "up",
			"enps2": "down",
			"enps3": "down",
		},
		want: "SubHealthy",
	}
}

func buildGetDpuFaultInfoOfNpuCase03() npuFaultByDpuCase {
	return npuFaultByDpuCase{
		name:    "[UB-type]none of the two is in up state, the DpuUnhealth is UnHealth",
		npuName: "Ascend910-8", // equal to the Ascend910-0
		npuToDpuMap: map[string][]string{
			"0": {"enps0", "enps2"},
			"1": {"enps0", "enps2"},
			"4": {"enps1", "enps3"},
			"5": {"enps1", "enps3"},
		},
		dpuOperstateMap: map[string]string{
			"enps0": "down",
			"enps1": "up",
			"enps2": "down",
			"enps3": "down",
		},
		want: "Unhealthy",
	}
}

func buildGetDpuFaultInfoOfNpuCase04() npuFaultByDpuCase {
	return npuFaultByDpuCase{
		name:    "[PCIe-type]none of the DPU is in up state, the DpuUnhealth is UnHealth",
		npuName: "Ascend910-8", // equal to the Ascend910-0
		npuToDpuMap: map[string][]string{
			"0": {"eth0"},
			"1": {"eth1"},
			"4": {"eth4"},
			"5": {"eth5"},
		},
		dpuOperstateMap: map[string]string{
			"eth0": "down",
			"eth1": "down",
			"eth4": "down",
			"eth5": "down",
		},
		want: "Unhealthy",
	}
}

func buildGetDpuFaultInfoOfNpuCase05() npuFaultByDpuCase {
	return npuFaultByDpuCase{
		name:    "[PCIe-type]the DPU is in up state, the DpuUnhealth is Health",
		npuName: "Ascend910-8", // equal to the Ascend910-0
		npuToDpuMap: map[string][]string{
			"0": {"eth0"},
			"1": {"eth1"},
			"4": {"eth4"},
			"5": {"eth5"},
		},
		dpuOperstateMap: map[string]string{
			"eth0": "up",
			"eth1": "down",
			"eth4": "down",
			"eth5": "down",
		},
		want: "Healthy",
	}
}

func TestIsNpuMatched(t *testing.T) {
	tests := append(
		buildIsNpuMatchedTestCasesPart1(),
		buildIsNpuMatchedTestCasesPart2()...,
	)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isNpuMatched(tt.args.str, tt.args.target); got != tt.want {
				t.Errorf("isNpuMatched() = %v, want %v", got, tt.want)
			}
		})
	}
}

type isNpuMatchedArgs struct {
	str    string
	target string
}

type isNpuMatchedTestCase struct {
	name string
	args isNpuMatchedArgs
	want bool
}

func buildIsNpuMatchedTestCasesPart1() []isNpuMatchedTestCase {
	return []isNpuMatchedTestCase{
		{
			name: "01-should return true when target matches cardIdStr",
			args: isNpuMatchedArgs{
				str:    "Ascend910-1",
				target: "1",
			},
			want: true,
		},
		{
			name: "02-should return false when target does not match any cardIdStr",
			args: isNpuMatchedArgs{
				str:    "Ascend910-0",
				target: "3",
			},
			want: false,
		},
		{
			name: "03-should return false when input string is empty",
			args: isNpuMatchedArgs{
				str:    "",
				target: "0",
			},
			want: false,
		},
		{
			name: "04-should return false when string does not contain '-'",
			args: isNpuMatchedArgs{
				str:    "Ascend910",
				target: "0",
			},
			want: false,
		},
	}
}

func buildIsNpuMatchedTestCasesPart2() []isNpuMatchedTestCase {
	return []isNpuMatchedTestCase{
		{
			name: "05-should return false when cardId is not an integer",
			args: isNpuMatchedArgs{
				str:    "Ascend910-abc",
				target: "0",
			},
			want: false,
		},
		{
			name: "06-should handle cardIdStr with modulo NodeNum8",
			args: isNpuMatchedArgs{
				str:    "Ascend910-9", // 9 % 8 = 1
				target: "1",
			},
			want: true,
		},
		{
			name: "07-should return false if modulo does not match target",
			args: isNpuMatchedArgs{
				str:    "Ascend910-10", // 10 % 8 = 2
				target: "3",
			},
			want: false,
		},
	}
}
