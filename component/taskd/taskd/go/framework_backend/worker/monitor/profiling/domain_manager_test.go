/* Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.


   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

// Package profiling contains functions that support dynamically collecting profiling data
package profiling

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"sync"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"

	"taskd/common/constant"
	"taskd/common/utils"
	"taskd/framework_backend/manager/infrastructure/storage"
	"taskd/toolkit_backend/net"
	"taskd/toolkit_backend/net/common"
)

const checkDomainRunInterval = 2 * constant.DomainCheckInterval

func TestGetProfilingSwitchInvalidJson(t *testing.T) {
	t.Run("invalid json content", func(t *testing.T) {
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		// Mock read file content
		patches.ApplyFunc(ioutil.ReadFile, func(path string) ([]byte, error) {
			return []byte("{invalid json}"), nil
		})
		result, _ := utils.GetProfilingSwitch("any_path")
		expected := allOffSwitch()
		if result != expected {
			t.Errorf("expect %+v，actual %+v", expected, result)
		}
	})
}

func TestGetProfilingSwitchValidJson(t *testing.T) {
	t.Run("valid json content", func(t *testing.T) {
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		// Mock read file content
		patches.ApplyFunc(ioutil.ReadFile, func(path string) ([]byte, error) {
			return json.Marshal(constant.ProfilingSwitch{
				CommunicationOperator: "ON",
				Step:                  "OFF",
				SaveCheckpoint:        "ON",
				FP:                    "OFF",
				DataLoader:            "ON",
			})
		})
		result, _ := utils.GetProfilingSwitch("any_path")
		expected := constant.ProfilingSwitch{
			CommunicationOperator: "ON",
			Step:                  "OFF",
			SaveCheckpoint:        "ON",
			FP:                    "OFF",
			DataLoader:            "ON",
		}
		if result != expected {
			t.Errorf("expect %+v，actual %+v", expected, result)
		}
	})
}

func TestGetProfilingSwitchReadFileFailed(t *testing.T) {
	t.Run("read file failed", func(t *testing.T) {
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		// Mock read file content
		patches.ApplyFunc(ioutil.ReadFile, func(path string) ([]byte, error) {
			return nil, errors.New("file error")
		})
		result, _ := utils.GetProfilingSwitch("any_path")
		expected := allOffSwitch()
		if result != expected {
			t.Errorf("expect %+v，actual %+v", expected, result)
		}
	})
}

func TestManageDomainEnableStatusOffAll(t *testing.T) {

	t.Run("off all switch", func(t *testing.T) {
		// create ctx
		ctx, cancel := context.WithCancel(context.Background())
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		patches.ApplyFunc(utils.GetProfilingSwitch, func(path string) (constant.ProfilingSwitch, error) {
			return allOffSwitch(), nil
		})
		mu := sync.Mutex{}
		var called bool
		patches.ApplyFunc(changeProfileSwitchStatus, func(profilingSwitches constant.ProfilingDomainCmd) {
			mu.Lock()
			called = true
			mu.Unlock()
		})

		// start manage domain
		go ManageDomainEnableStatus(ctx)
		time.Sleep(checkDomainRunInterval)
		cancel()
		mu.Lock()
		assert.True(t, called)
		mu.Unlock()
	})

}

func TestManageDomainEnableStatusAnyOn(t *testing.T) {

	t.Run("any switch is on, except communicate", func(t *testing.T) {
		// create ctx
		ctx, cancel := context.WithCancel(context.Background())
		patches := gomonkey.NewPatches()
		defer patches.Reset()

		patches.ApplyFunc(utils.GetProfilingSwitch, func(path string) (constant.ProfilingSwitch, error) {
			return constant.ProfilingSwitch{Step: constant.SwitchON}, nil
		})
		mu := sync.Mutex{}
		var called bool
		patches.ApplyFunc(changeProfileSwitchStatus, func(profilingSwitches constant.ProfilingDomainCmd) {
			mu.Lock()
			called = true
			mu.Unlock()
		})

		// start manage domain
		go ManageDomainEnableStatus(ctx)
		time.Sleep(checkDomainRunInterval)
		cancel()
		mu.Lock()
		assert.True(t, called)
		mu.Unlock()
	})

}

func TestManageDomainEnableStatusOnAll(t *testing.T) {
	t.Run("all switch is on", func(t *testing.T) {
		// create ctx
		ctx, cancel := context.WithCancel(context.Background())
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		patches.ApplyFunc(utils.GetProfilingSwitch, func(path string) (constant.ProfilingSwitch, error) {
			return allOnSwitch(), nil
		})
		mu := sync.Mutex{}
		var called bool
		patches.ApplyFunc(changeProfileSwitchStatus, func(profilingSwitches constant.ProfilingDomainCmd) {
			mu.Lock()
			called = true
			mu.Unlock()
		})

		// start manage domain
		go ManageDomainEnableStatus(ctx)
		time.Sleep(checkDomainRunInterval)
		cancel()
		mu.Lock()
		assert.True(t, called)
		mu.Unlock()
	})

}

func TestChangeProfileSwitchStatusAllOn(t *testing.T) {
	t.Run("all switch is on", func(t *testing.T) {
		patches := gomonkey.NewPatches()
		defer patches.Reset()

		var disableMspCall bool
		patches.ApplyFunc(DisableMsptiActivity, func() error {
			disableMspCall = true
			return nil
		})

		var enableMspCall bool
		patches.ApplyFunc(EnableMsptiMarkerActivity, func() error {
			enableMspCall = true
			return nil
		})

		var enableMarkerCall bool
		patches.ApplyFunc(EnableMarkerDomain, func(domainName string, status bool) error {
			enableMarkerCall = true
			return nil
		})

		// start manage domain
		changeProfileSwitchStatus(constant.ProfilingDomainCmd{
			DefaultDomainAble: true,
			CommDomainAble:    true,
		})

		assert.Equal(t, false, disableMspCall)
		assert.Equal(t, true, enableMspCall)
		assert.Equal(t, true, enableMarkerCall)
	})

}

func TestChangeProfileSwitchStatusAllOff(t *testing.T) {
	t.Run("all switch is on", func(t *testing.T) {
		patches := gomonkey.NewPatches()
		defer patches.Reset()

		var disableMspCall bool
		patches.ApplyFunc(DisableMsptiActivity, func() error {
			disableMspCall = true
			return nil
		})

		var enableMspCall bool
		patches.ApplyFunc(EnableMsptiMarkerActivity, func() error {
			enableMspCall = true
			return nil
		})

		var enableMarkerCall bool
		patches.ApplyFunc(EnableMarkerDomain, func(domainName string, status bool) error {
			enableMarkerCall = true
			return nil
		})

		// start manage domain
		changeProfileSwitchStatus(constant.ProfilingDomainCmd{})

		assert.Equal(t, true, disableMspCall)
		assert.Equal(t, false, enableMspCall)
		assert.Equal(t, true, enableMarkerCall)
	})

}

func TestChangeProfileSwitchStatusAnyOff(t *testing.T) {
	t.Run("all switch is on", func(t *testing.T) {
		patches := gomonkey.NewPatches()
		defer patches.Reset()

		var disableMspCall bool
		patches.ApplyFunc(DisableMsptiActivity, func() error {
			disableMspCall = true
			return nil
		})

		var enableMspCall bool
		patches.ApplyFunc(EnableMsptiMarkerActivity, func() error {
			enableMspCall = true
			return nil
		})

		var enableMarkerCall bool
		patches.ApplyFunc(EnableMarkerDomain, func(domainName string, status bool) error {
			enableMarkerCall = true
			return nil
		})

		// start manage domain
		changeProfileSwitchStatus(constant.ProfilingDomainCmd{DefaultDomainAble: true})

		assert.Equal(t, false, disableMspCall)
		assert.Equal(t, true, enableMspCall)
		assert.Equal(t, true, enableMarkerCall)
	})

}

func allOffSwitch() constant.ProfilingSwitch {
	return constant.ProfilingSwitch{
		CommunicationOperator: constant.SwitchOFF,
		Step:                  constant.SwitchOFF,
		SaveCheckpoint:        constant.SwitchOFF,
		FP:                    constant.SwitchOFF,
		DataLoader:            constant.SwitchOFF,
	}
}

func allOnSwitch() constant.ProfilingSwitch {
	return constant.ProfilingSwitch{
		CommunicationOperator: constant.SwitchON,
		Step:                  constant.SwitchON,
		SaveCheckpoint:        constant.SwitchON,
		FP:                    constant.SwitchON,
		DataLoader:            constant.SwitchON,
	}
}

func TestNotifyMgrSwitchChange(t *testing.T) {
	NetTool = &net.NetInstance{}
	called := false
	gomonkey.ApplyMethod(NetTool, "SyncSendMessage",
		func(nt *net.NetInstance, uuid, mtype, msgBody string, dst *common.Position) (*common.Ack, error) {
			called = true
			return nil, nil
		})

	notifyMgrSwitchChange(constant.ProfilingResult{})
	convey.ShouldBeTrue(called)
}

func TestProcessMsg(t *testing.T) {
	patches := gomonkey.NewPatches()
	defer patches.Reset()
	ProcessMsg(0, &common.Message{
		Body: utils.ObjToString(storage.MsgBody{
			Code: constant.ProfilingAllOnCmdCode,
		}),
	})
	var profilingSwitch constant.ProfilingDomainCmd
	select {
	case msg := <-CmdChan:
		profilingSwitch = msg
	default:
	}
	convey.ShouldBeTrue(profilingSwitch.CommDomainAble)
	convey.ShouldBeTrue(profilingSwitch.DefaultDomainAble)
}
