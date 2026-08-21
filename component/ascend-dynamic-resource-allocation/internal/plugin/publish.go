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

	"k8s.io/dynamic-resource-allocation/kubeletplugin"
	"k8s.io/dynamic-resource-allocation/resourceslice"
)

// ResourcePublisher wraps kubeletplugin.Helper for ResourceSlice publishing.
type ResourcePublisher struct {
	*kubeletplugin.Helper
}

// PublishResources publishes pre-built driver resources. The driver owns the
// mapping from device info to ResourceSlice contents (including the
// per-generation attributes); the publisher is a thin adapter over the
// kubelet helper and stays free of chip-generation knowledge.
func (rp *ResourcePublisher) PublishResources(ctx context.Context, resources resourceslice.DriverResources) error {
	return rp.Helper.PublishResources(ctx, resources)
}
