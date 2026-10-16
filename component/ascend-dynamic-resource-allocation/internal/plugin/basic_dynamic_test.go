/*
 * Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at

 * http://www.apache.org/licenses/LICENSE-2.0

 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package plugin

import (
	"context"
	"encoding/json"
	"testing"

	resourceapi "k8s.io/api/resource/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"ascend-common/devmanager"
	dc "ascend-common/devmanager/common"
	"ascend-dynamic-resource-allocation/internal/device"
	"ascend-dynamic-resource-allocation/pkg/consts"
)

type basicHardware struct {
	devmanager.DeviceInterface
	devices           map[uint32]string
	cores             float32
	creates, destroys int
	destroyErr        error
}

func (h *basicHardware) GetLogicIDFromPhysicID(id int32) (int32, error) { return id, nil }
func (h *basicHardware) CreateVirtualDevice(_ int32, in dc.CgoCreateVDevRes) (dc.CgoCreateVDevOut, error) {
	h.creates++
	h.devices[212] = in.TemplateName
	return dc.CgoCreateVDevOut{VDevID: 212}, nil
}
func (h *basicHardware) GetVirtualDeviceInfo(int32) (dc.VirtualDevInfo, error) {
	info := dc.VirtualDevInfo{}
	info.TotalResource.VDevNum = uint32(len(h.devices))
	for id, name := range h.devices {
		info.VDevInfo = append(info.VDevInfo, dc.CgoVDevQueryStru{VDevID: id, QueryInfo: dc.CgoVDevQueryInfo{
			Name: name, Status: dc.InvalidID, Computing: dc.CgoComputingResource{Aic: h.cores},
		}})
	}
	return info, nil
}
func (h *basicHardware) DestroyVirtualDevice(_ int32, id uint32) error {
	if h.destroyErr != nil {
		return h.destroyErr
	}
	h.destroys++
	delete(h.devices, id)
	return nil
}

func basicDynamicClaim() *resourceapi.ResourceClaim {
	claim := allocatedClaim("basic-dynamic", "npu-7")
	share := types.UID("basic-share")
	r := &claim.Status.Allocation.Devices.Results[0]
	r.ShareID = &share
	r.ConsumedCapacity = map[resourceapi.QualifiedName]resource.Quantity{consts.AICoreNum: resource.MustParse("5")}
	payload, _ := json.Marshal(dynamicVNPUConfig{Template: "vir05_1c_16g"})
	claim.Status.Allocation.Devices.Config = []resourceapi.DeviceAllocationConfiguration{{
		Source: resourceapi.AllocationConfigSourceClaim,
		DeviceConfiguration: resourceapi.DeviceConfiguration{Opaque: &resourceapi.OpaqueDeviceConfiguration{
			Driver: consts.DriverName, Parameters: runtime.RawExtension{Raw: payload},
		}},
	}}
	return claim
}

func TestBasicDynamicCreateInjectReuseAndRelease(t *testing.T) {
	h := &basicHardware{devices: map[uint32]string{}, cores: 5}
	specs := &fakeCdiSpec{cdiIDs: []string{"ascend.com/npu=212"}}
	state := newTestDeviceState(t, specs)
	state.dynamicVNPUManager = device.NewDynamicVNPUManager(h)
	adp := &AscendDraPlugin{state: state}
	claim := basicDynamicClaim()
	for i := 0; i < 2; i++ {
		r := adp.prepareResourceClaim(context.Background(), claim)
		if r.Err != nil || len(r.Devices) != 1 {
			t.Fatalf("result=%+v", r)
		}
		d := r.Devices[0]
		if d.ShareID == nil || *d.ShareID != "basic-share" || len(d.CDIDeviceIDs) != 1 || d.CDIDeviceIDs[0] != "ascend.com/npu=212" {
			t.Fatalf("device=%+v", d)
		}
	}
	if state.checkpoint.V1.PreparedClaims[string(claim.UID)][0].PhyID != 7 ||
		state.ValidateMode(false) != nil || state.ValidateMode(true) == nil {
		t.Fatal("checkpoint lost physical identity or allocation mode protection")
	}
	if h.creates != 1 || len(h.devices) != 1 {
		t.Fatal("repeated Prepare created another vNPU")
	}
	if err := state.Unprepare(string(claim.UID)); err != nil {
		t.Fatal(err)
	}
	if len(h.devices) != 0 || h.destroys != 1 || len(state.checkpoint.V1.PreparedClaims) != 0 {
		t.Fatal("release incomplete")
	}
	if err := state.Unprepare(string(claim.UID)); err != nil {
		t.Fatal(err)
	}
	if h.destroys != 1 {
		t.Fatal("release was not idempotent")
	}
}

func TestBasicActualCapacityMismatchIsRolledBack(t *testing.T) {
	h := &basicHardware{devices: map[uint32]string{}, cores: 10}
	specs := &fakeCdiSpec{cdiIDs: []string{"ascend.com/npu=212"}}
	state := newTestDeviceState(t, specs)
	state.dynamicVNPUManager = device.NewDynamicVNPUManager(h)
	if _, err := state.Prepare(basicDynamicClaim()); err == nil {
		t.Fatal("wrong actual capacity accepted")
	}
	if h.destroys != 1 || len(h.devices) != 0 || len(specs.writes) != 0 {
		t.Fatal("invalid vNPU was injected or leaked")
	}
}
