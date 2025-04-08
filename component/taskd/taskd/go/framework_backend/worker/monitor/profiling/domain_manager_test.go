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
	"github.com/stretchr/testify/assert"

	"taskd/common/constant"
)

func TestGetProfilingSwitchInvalidJson(t *testing.T) {
	t.Run("invalid json content", func(t *testing.T) {
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		// Mock read file content
		patches.ApplyFunc(ioutil.ReadFile, func(path string) ([]byte, error) {
			return []byte("{invalid json}"), nil
		})
		result := GetProfilingSwitch("any_path")
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
			return json.Marshal(SwitchProfiling{
				CommunicationOperator: "ON",
				Step:                  "OFF",
				SaveCheckpoint:        "ON",
				FP:                    "OFF",
				DataLoader:            "ON",
			})
		})
		result := GetProfilingSwitch("any_path")
		expected := SwitchProfiling{
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
		result := GetProfilingSwitch("any_path")
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

		patches.ApplyFunc(GetProfilingSwitch, func(path string) SwitchProfiling {
			return allOffSwitch()
		})
		mu := sync.Mutex{}
		var called bool
		patches.ApplyFunc(changeProfileSwitchStatus, func(profilingSwitches SwitchProfiling) {
			mu.Lock()
			called = true
			mu.Unlock()
		})

		// start manage domain
		go ManageDomainEnableStatus(ctx)
		time.Sleep(constant.DomainCheckInterval)
		cancel()
		time.Sleep(constant.CheckProfilingCacheInterval)
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

		patches.ApplyFunc(GetProfilingSwitch, func(path string) SwitchProfiling {
			return SwitchProfiling{Step: constant.SwitchON}
		})
		mu := sync.Mutex{}
		var called bool
		patches.ApplyFunc(changeProfileSwitchStatus, func(profilingSwitches SwitchProfiling) {
			mu.Lock()
			called = true
			mu.Unlock()
		})

		// start manage domain
		go ManageDomainEnableStatus(ctx)
		time.Sleep(constant.DomainCheckInterval)
		cancel()
		time.Sleep(constant.CheckProfilingCacheInterval)
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
		patches.ApplyFunc(GetProfilingSwitch, func(path string) SwitchProfiling {
			return allOnSwitch()
		})
		mu := sync.Mutex{}
		var called bool
		patches.ApplyFunc(changeProfileSwitchStatus, func(profilingSwitches SwitchProfiling) {
			mu.Lock()
			called = true
			mu.Unlock()
		})

		// start manage domain
		go ManageDomainEnableStatus(ctx)
		time.Sleep(constant.DomainCheckInterval)
		cancel()
		time.Sleep(constant.CheckProfilingCacheInterval)
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
		patches.ApplyFunc(EnableMarkerDomain, func(domainName string, status string) error {
			enableMarkerCall = true
			return nil
		})

		// start manage domain
		changeProfileSwitchStatus(allOnSwitch())

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
		patches.ApplyFunc(EnableMarkerDomain, func(domainName string, status string) error {
			enableMarkerCall = true
			return nil
		})

		// start manage domain
		changeProfileSwitchStatus(allOffSwitch())

		assert.Equal(t, true, disableMspCall)
		assert.Equal(t, false, enableMspCall)
		assert.Equal(t, false, enableMarkerCall)
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
		patches.ApplyFunc(EnableMarkerDomain, func(domainName string, status string) error {
			enableMarkerCall = true
			return nil
		})

		// start manage domain
		changeProfileSwitchStatus(SwitchProfiling{Step: constant.SwitchON})

		assert.Equal(t, false, disableMspCall)
		assert.Equal(t, true, enableMspCall)
		assert.Equal(t, true, enableMarkerCall)
	})

}

func allOffSwitch() SwitchProfiling {
	return SwitchProfiling{
		CommunicationOperator: constant.SwitchOFF,
		Step:                  constant.SwitchOFF,
		SaveCheckpoint:        constant.SwitchOFF,
		FP:                    constant.SwitchOFF,
		DataLoader:            constant.SwitchOFF,
	}
}

func allOnSwitch() SwitchProfiling {
	return SwitchProfiling{
		CommunicationOperator: constant.SwitchON,
		Step:                  constant.SwitchON,
		SaveCheckpoint:        constant.SwitchON,
		FP:                    constant.SwitchON,
		DataLoader:            constant.SwitchON,
	}
}
