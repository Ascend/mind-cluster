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
	"errors"
	"fmt"

	resourceapi "k8s.io/api/resource/v1"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	coreclientset "k8s.io/client-go/kubernetes"
	"k8s.io/dynamic-resource-allocation/kubeletplugin"

	"ascend-common/common-utils/hwlog"
	draFlags "ascend-dynamic-resource-allocation/internal/flags"
	"ascend-dynamic-resource-allocation/pkg/consts"
)

// AscendDraPlugin implements the kubelet DRA plugin interface for Ascend devices.
type AscendDraPlugin struct {
	client coreclientset.Interface
	*ResourcePublisher
	state *DeviceState
	*DraHealthManager
	cancelCtx func(error)
}

// NewAscendDraPlugin creates a plugin instance with kube client, state and health manager.
func NewAscendDraPlugin(
	draConfig *draFlags.DRAConfig,
	cancelCtx context.CancelCauseFunc,
	specs CdiSpecInterface) (*AscendDraPlugin, error) {
	clientSets, err := draConfig.KubeClientConfig.NewClientSets()
	if err != nil {
		return nil, fmt.Errorf("create client: %v", err)
	}
	draPlugin := &AscendDraPlugin{
		client:            clientSets.Core,
		cancelCtx:         cancelCtx,
		ResourcePublisher: &ResourcePublisher{},
		DraHealthManager:  NewDraHealthChecker(draConfig.DraHealthzConfig),
	}

	state, err := NewDeviceState(draConfig.DraOption, specs)
	if err != nil {
		return nil, err
	}
	draPlugin.state = state
	hwlog.RunLog.Info("dra plugin instance created")
	return draPlugin, nil
}

// RegisterService registers the plugin with kubelet.
func (adp *AscendDraPlugin) RegisterService(ctx context.Context, draConfig *draFlags.DRAConfig) error {
	helper, err := kubeletplugin.Start(
		ctx,
		adp,
		kubeletplugin.KubeClient(adp.client),
		kubeletplugin.NodeName(draConfig.DraOption.NodeName),
		kubeletplugin.DriverName(consts.DriverName),
		kubeletplugin.RegistrarDirectoryPath(draConfig.DraOption.KubeletRegistrarDirectoryPath),
		kubeletplugin.PluginDataDirectoryPath(draConfig.DraOption.DriverPluginPath()),
	)
	adp.ResourcePublisher.Helper = helper
	hwlog.RunLog.Infof("kubelet plugin registered, node=%s, driver=%s",
		draConfig.DraOption.NodeName, consts.DriverName)
	return err
}

// LoadCheckpoint loads and verifies the prepared-claim state before the
// plugin is registered with kubelet.
func (adp *AscendDraPlugin) LoadCheckpoint() error {
	if adp.state == nil {
		return errors.New("device state is not initialized")
	}
	return adp.state.LoadCheckpoint()
}

// PrepareResourceClaims is the kubelet DRA callback for allocating devices
// for one or more claims. It is idempotent: a previously prepared claim
// returns the same prepared devices from the checkpoint.
func (adp *AscendDraPlugin) PrepareResourceClaims(ctx context.Context, claims []*resourceapi.ResourceClaim) (map[types.UID]kubeletplugin.PrepareResult, error) {
	if len(claims) > 0 {
		hwlog.RunLog.Infof("PrepareResourceClaims is called, number of claims: %d", len(claims))
	}
	result := make(map[types.UID]kubeletplugin.PrepareResult)
	for _, claim := range claims {
		result[claim.UID] = adp.prepareResourceClaim(ctx, claim)
	}

	return result, nil
}

func (adp *AscendDraPlugin) prepareResourceClaim(_ context.Context, claim *resourceapi.ResourceClaim) kubeletplugin.PrepareResult {
	preparedPBs, err := adp.state.Prepare(claim)
	if err != nil {
		return kubeletplugin.PrepareResult{
			Err: fmt.Errorf("error preparing devices for claim %v: %w", claim.UID, err),
		}
	}
	var prepared []kubeletplugin.Device
	for _, preparedPB := range preparedPBs {
		prepared = append(prepared, kubeletplugin.Device{
			Requests:     preparedPB.GetRequestNames(),
			PoolName:     preparedPB.GetPoolName(),
			DeviceName:   preparedPB.GetDeviceName(),
			CDIDeviceIDs: preparedPB.GetCdiDeviceIds(),
		})
	}
	hwlog.RunLog.Infof("Returning newly prepared devices for claim '%v', count=%d", claim.UID, len(prepared))
	for i, dev := range prepared {
		hwlog.RunLog.Debugf("  prepared[%d]: request=%v pool=%s device=%s cdi=%v",
			i, dev.Requests, dev.PoolName, dev.DeviceName, dev.CDIDeviceIDs)
	}
	return kubeletplugin.PrepareResult{Devices: prepared}
}

// UnprepareResourceClaims is the kubelet DRA callback for releasing the
// devices previously allocated to one or more claims. It is idempotent: a
// missing claim is a no-op.
func (adp *AscendDraPlugin) UnprepareResourceClaims(ctx context.Context, claims []kubeletplugin.NamespacedObject) (map[types.UID]error, error) {
	if len(claims) > 0 {
		hwlog.RunLog.Infof("UnprepareResourceClaims is called, number of claims: %d", len(claims))
	}
	result := make(map[types.UID]error)

	for _, claim := range claims {
		result[claim.UID] = adp.unprepareResourceClaim(ctx, claim)
	}

	return result, nil
}

func (adp *AscendDraPlugin) unprepareResourceClaim(_ context.Context, claim kubeletplugin.NamespacedObject) error {
	if err := adp.state.Unprepare(string(claim.UID)); err != nil {
		return fmt.Errorf("error unpreparing devices for claim %v: %w", claim.UID, err)
	}
	hwlog.RunLog.Infof("devices unprepared for claim %v", claim.UID)
	return nil
}

// HandleError reports a background error to the kubelet runtime handler.
// Recoverable errors (ErrRecoverable) are logged at warning level;
// otherwise the error is fatal, the plugin context is cancelled to shut
// the driver down.
func (adp *AscendDraPlugin) HandleError(ctx context.Context, err error, msg string) {
	utilruntime.HandleErrorWithContext(ctx, err, msg)
	if errors.Is(err, kubeletplugin.ErrRecoverable) {
		hwlog.RunLog.Warnf("recoverable background error: %s, err: %v", msg, err)
		return
	}
	hwlog.RunLog.Errorf("fatal background error: %s, err: %v", msg, err)
	if adp.cancelCtx != nil {
		adp.cancelCtx(fmt.Errorf("fatal background error: %w", err))
	}
}
