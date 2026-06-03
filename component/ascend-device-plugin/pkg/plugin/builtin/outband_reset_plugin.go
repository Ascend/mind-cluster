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

// Package builtin provides built-in hot reset plugins
package builtin

import (
	"context"
	"fmt"
	"time"

	"Ascend-device-plugin/pkg/common"
	"Ascend-device-plugin/pkg/plugin"
	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-common/devmanager"
)

const (
	outbandResetPluginName   = "outbandReset"
	beforeOutBandRescanDelay = 3
	outBandBootPollTimeout   = 150 * time.Second
)

type OutBandResetPlugin struct {
	plugin.HotResetPluginAdapter
	dmgr devmanager.DeviceInterface
}

func NewOutBandResetPlugin(dmgr devmanager.DeviceInterface) *OutBandResetPlugin {
	return &OutBandResetPlugin{dmgr: dmgr}
}

func (p *OutBandResetPlugin) Name() string {
	return outbandResetPluginName
}

func (p *OutBandResetPlugin) CustomReset(_ context.Context, deviceList []plugin.ResetDevice,
	resetErr error) error {
	if resetErr == nil {
		return nil
	}
	for _, dev := range deviceList {
		if dev.CardType != api.Ascend910A3 {
			return resetErr
		}
	}
	hwlog.RunLog.Infof("custom reset error: %v, continue out band reset", resetErr)
	var lastErr error
	for _, dev := range deviceList {
		if err := p.resetDeviceOutBand(dev.LogicID); err != nil {
			lastErr = err
			hwlog.RunLog.Errorf("out band reset failed for logicID %d: %v", dev.LogicID, err)
			continue
		}
		if err := p.waitRingResetComplete(deviceList); err != nil {
			lastErr = err
			hwlog.RunLog.Errorf("wait ring boot failed for logicID %d: %v", dev.LogicID, err)
			continue
		}
		hwlog.RunLog.Infof("out band reset success, logicID: %d", dev.LogicID)
	}
	return lastErr
}

func (p *OutBandResetPlugin) waitRingResetComplete(deviceList []plugin.ResetDevice) error {
	deadline := time.Now().Add(outBandBootPollTimeout)
	for {
		if p.allDevicesBooted(deviceList) {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("wait ring boot complete timeout, devices: %d", len(deviceList))
		}
		time.Sleep(1 * time.Second)
	}
}

func (p *OutBandResetPlugin) allDevicesBooted(deviceList []plugin.ResetDevice) bool {
	for _, dev := range deviceList {
		bootState, err := p.dmgr.GetDeviceBootStatus(dev.LogicID)
		if err != nil || bootState != common.BootStartFinish {
			return false
		}
	}
	return true
}

func (p *OutBandResetPlugin) resetDeviceOutBand(logicID int32) error {
	if err := p.dmgr.GetOutBandChannelState(logicID); err != nil {
		return fmt.Errorf("out band channel state error: %w", err)
	}
	if err := p.dmgr.PreResetSoc(logicID); err != nil {
		return fmt.Errorf("pre reset soc failed: %w", err)
	}
	if err := p.dmgr.SetDeviceResetOutBand(logicID); err != nil {
		return fmt.Errorf("reset out band failed: %w", err)
	}
	time.Sleep(beforeOutBandRescanDelay * time.Second)
	if err := p.dmgr.RescanSoc(logicID); err != nil {
		return fmt.Errorf("rescan device failed: %w", err)
	}
	return nil
}
