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

package device

import (
	"context"
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-common/devmanager"
	dc "ascend-common/devmanager/common"
)

// initLog initialises the hwlog run logger so generation code that calls
// hwlog.RunLog does not panic on a nil logger during tests.
func initLog() {
	hwLogConfig := hwlog.LogConfig{OnlyToStdout: true}
	hwlog.InitRunLogger(&hwLogConfig, context.Background())
}

func init() { initLog() }

// errSentinel is a stable sentinel error reused by gomonkey stubs.
var errSentinel = errors.New("test sentinel error")

// new910WithMock builds an Ascend910Generation backed by a DeviceManagerMock
// whose DevType is api.Ascend910 so GetDevType returns a meaningful value.
func new910WithMock() (*Ascend910Generation, *devmanager.DeviceManagerMock) {
	mock := &devmanager.DeviceManagerMock{DevType: api.Ascend910}
	g := NewAscend910Generation()
	g.SetDmgr(mock)
	return g, mock
}

// new950WithMock builds an Ascend950Generation backed by a DeviceManagerMock.
func new950WithMock() (*Ascend950Generation, *devmanager.DeviceManagerMock) {
	mock := &devmanager.DeviceManagerMock{DevType: api.Ascend910A5}
	g := NewAscend950Generation()
	g.SetDmgr(mock)
	return g, mock
}

// Verify both successful creation and failures that must retain the created identity for cleanup.
func TestBasicManagerCreateVerification(t *testing.T) {
	for _, tc := range []struct {
		name      string
		cores     float32
		status    uint32
		queryErr  error
		missing   bool
		wantError bool
	}{
		{name: "unused A2 device", cores: 5, status: dc.InvalidID},
		{name: "wrong actual cores", cores: 10, wantError: true},
		{name: "already in use", cores: 5, status: 1, wantError: true},
		{name: "query failure", queryErr: errSentinel, wantError: true},
		{name: "created device missing", missing: true, wantError: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mock := &devmanager.DeviceManagerMock{DevType: api.Ascend910B}
			patches := gomonkey.NewPatches()
			defer patches.Reset()
			patches.ApplyMethodReturn(mock, "GetLogicIDFromPhysicID", int32(7), nil)
			patches.ApplyMethodReturn(mock, "CreateVirtualDevice", dc.CgoCreateVDevOut{VDevID: 212}, nil)
			info := dc.VirtualDevInfo{}
			if !tc.missing {
				info.TotalResource.VDevNum = 1
				info.VDevInfo = []dc.CgoVDevQueryStru{{VDevID: 212, QueryInfo: dc.CgoVDevQueryInfo{
					Name: "future-template", Status: tc.status, Computing: dc.CgoComputingResource{Aic: tc.cores},
				}}}
			}
			patches.ApplyMethodReturn(mock, "GetVirtualDeviceInfo", info, tc.queryErr)
			v, err := NewDynamicVNPUManager(mock).Create(7, "future-template", 5)
			if (err != nil) != tc.wantError || v == nil || v.VDevID != 212 || v.LogicID != 7 || v.CDIDeviceName != "dynamic-vnpu-212-7" {
				t.Fatalf("created=%+v error=%v", v, err)
			}
		})
	}
}

func TestBasicManagerDestroyVerification(t *testing.T) {
	for _, tc := range []struct {
		name                 string
		wantCalls            int
		queryErr, destroyErr error
		present, wantError   bool
	}{
		{name: "release", present: true, wantCalls: 1},
		{name: "already absent"},
		{name: "query failure", queryErr: errSentinel, wantError: true},
		{name: "destroy failure", present: true, destroyErr: errSentinel, wantError: true, wantCalls: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mock := &devmanager.DeviceManagerMock{DevType: api.Ascend910B}
			patches := gomonkey.NewPatches()
			defer patches.Reset()
			info := dc.VirtualDevInfo{}
			if tc.present {
				info.TotalResource.VDevNum = 1
				info.VDevInfo = []dc.CgoVDevQueryStru{{VDevID: 212}}
			}
			patches.ApplyMethodReturn(mock, "GetVirtualDeviceInfo", info, tc.queryErr)
			calls := 0
			patches.ApplyMethod(mock, "DestroyVirtualDevice", func(_ *devmanager.DeviceManagerMock, logic int32, id uint32) error {
				if logic != 7 || id != 212 {
					t.Fatalf("destroy wrong identity: %d/%d", logic, id)
				}
				calls++
				return tc.destroyErr
			})
			err := NewDynamicVNPUManager(mock).Destroy(7, 212)
			if (err != nil) != tc.wantError || calls != tc.wantCalls {
				t.Fatalf("destroy calls=%d error=%v", calls, err)
			}
		})
	}
}
