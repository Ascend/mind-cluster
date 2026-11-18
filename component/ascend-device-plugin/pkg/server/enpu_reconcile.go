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

package server

import (
	"context"
	"fmt"
	"time"

	"Ascend-device-plugin/pkg/kubeclient"
	"ascend-common/common-utils/hwlog"
)

// managerReconcileInterval controls periodic orphan-allocation cleanup.
const managerReconcileInterval = 30 * time.Second

// reconcilePeriodically runs immediately and then reconciles until the context is canceled.
func (m *softShareManager) reconcilePeriodically(ctx context.Context, kubeClient *kubeclient.ClientK8s) {
	ticker := time.NewTicker(managerReconcileInterval)
	defer ticker.Stop()
	for {
		if err := m.reconcileAllocations(ctx, kubeClient); err != nil && ctx.Err() == nil {
			hwlog.RunLog.Warnf("reconcile enpu-manager allocations: %v", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

// reconcileAllocations releases allocations absent from the synchronized active Pod cache.
func (m *softShareManager) reconcileAllocations(ctx context.Context, kubeClient *kubeclient.ClientK8s) error {
	enabled, err := m.enabled()
	if err != nil || !enabled {
		return err
	}

	// Take both snapshots while holding the same lock used by allocation.
	m.allocationLock.Lock()
	defer m.allocationLock.Unlock()
	ctx, cancel := context.WithTimeout(ctx, managerRequestTimeout)
	defer cancel()

	allocations, err := m.client.ListAllocations(ctx)
	if err != nil {
		return fmt.Errorf("list enpu-manager allocations: %w", err)
	}
	if len(allocations) == 0 {
		return nil
	}
	if kubeClient == nil || kubeClient.PodInformer == nil || !kubeClient.PodInformer.HasSynced() {
		return fmt.Errorf("Pod cache is not synchronized")
	}
	pods := kubeClient.GetActivePodListCache()
	liveUIDs := make(map[string]struct{}, len(pods))
	for _, pod := range pods {
		if pod.UID == "" {
			return fmt.Errorf("missing UID for Pod %s/%s", pod.Namespace, pod.Name)
		}
		liveUIDs[string(pod.UID)] = struct{}{}
	}

	for _, allocation := range allocations {
		uid := allocation.Metadata.PodUID
		if _, live := liveUIDs[uid]; live {
			continue
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := m.client.Release(ctx, allocation.AllocationID); err != nil {
			hwlog.RunLog.Warnf("release enpu-manager allocation %s for Pod UID %q: %v", allocation.AllocationID, uid, err)
			continue
		}
		hwlog.RunLog.Infof("released enpu-manager allocation %s for Pod UID %q", allocation.AllocationID, uid)
	}
	return nil
}
