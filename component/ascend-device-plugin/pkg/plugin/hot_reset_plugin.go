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

// Package plugin defines the hot reset plugin interface and management
package plugin

import (
	"context"
	"time"
)

type ResetDevice struct {
	LogicID    int32
	CardID     int32
	DeviceID   int32
	PhyID      int32
	CardType   string
	IsFaultDev bool
	TokensLeft int32
}

type HookCaps struct {
	HasPreReset    bool
	HasCustomReset bool
	HasAfterReset  bool
}

type HotResetPlugin interface {
	Name() string
	PreReset(ctx context.Context, deviceList []ResetDevice) error
	CustomReset(ctx context.Context, deviceList []ResetDevice, resetErr error) error
	AfterReset(ctx context.Context, deviceList []ResetDevice, resetErr error) error
}

type HotResetPluginAdapter struct{}

func (a *HotResetPluginAdapter) Name() string { return "" }

func (a *HotResetPluginAdapter) PreReset(_ context.Context, _ []ResetDevice) error {
	return nil
}

func (a *HotResetPluginAdapter) CustomReset(_ context.Context, _ []ResetDevice, resetErr error) error {
	return resetErr
}

func (a *HotResetPluginAdapter) AfterReset(_ context.Context, _ []ResetDevice, _ error) error {
	return nil
}

const (
	PreResetTimeout    = 10 * time.Second
	CustomResetTimeout = 5 * time.Minute
	AfterResetTimeout  = 10 * time.Second
)
