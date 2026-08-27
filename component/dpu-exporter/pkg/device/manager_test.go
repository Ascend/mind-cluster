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

package device

import (
	"errors"
	"strings"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
)

// TestAutoInit covers all branches of the package-level AutoInit factory
// (requires -gcflags=all=-l for gomonkey, as build/test.sh sets globally).
func TestAutoInit(t *testing.T) {
	// unsupported card type
	if _, err := AutoInit("broadcom"); err == nil || !strings.Contains(err.Error(), "unsupported cardType") {
		t.Fatalf("AutoInit(unsupported) error = %v, want unsupported cardType", err)
	}

	// NewHwDpuManager failure
	p1 := gomonkey.NewPatches()
	defer p1.Reset()
	p1.ApplyFunc(NewHwDpuManager, func() (*HwDpuManager, error) {
		return nil, errors.New("manager init failed")
	})
	if _, err := AutoInit(CardTypeHuawei); err == nil || !strings.Contains(err.Error(), "init huawei dpu manager failed") {
		t.Fatalf("AutoInit(manager error) = %v, want init huawei dpu manager failed", err)
	}
	p1.Reset()

	// device discovery failure on the manager
	p2 := gomonkey.NewPatches()
	defer p2.Reset()
	p2.ApplyFunc(NewHwDpuManager, func() (*HwDpuManager, error) {
		return newFakeScriptManager(t, fakeHinicadm5Script("Card num:0")), nil
	})
	if _, err := AutoInit(CardTypeHuawei); err == nil || !strings.Contains(err.Error(), "huawei dpu auto init failed") {
		t.Fatalf("AutoInit(auto init error) = %v, want huawei dpu auto init failed", err)
	}
	p2.Reset()

	// success
	p3 := gomonkey.NewPatches()
	defer p3.Reset()
	p3.ApplyFunc(NewHwDpuManager, func() (*HwDpuManager, error) {
		return newFakeScriptManager(t, fakeHinicadm5Script(fakeInfoOutput)), nil
	})
	dm, err := AutoInit(CardTypeHuawei)
	if err != nil {
		t.Fatalf("AutoInit(huawei) error = %v", err)
	}
	if got := dm.GetDpuList(); len(got) != 1 || got[0].CardName != "hinic0" {
		t.Errorf("GetDpuList = %+v, want 1 hinic0 card", got)
	}
	if dm.GetCardType() != CardTypeHuawei {
		t.Errorf("GetCardType = %s, want %s", dm.GetCardType(), CardTypeHuawei)
	}
}
