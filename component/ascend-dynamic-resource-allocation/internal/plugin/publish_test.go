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

// PublishResources success-path coverage requires a fully configured
// kubeletplugin.Helper: all of its fields are unexported and only
// kubeletplugin.Start populates them, which in turn needs real gRPC
// registration sockets and a reachable API server. That is integration-test
// territory. The unit tests below cover the delegation contract that
// publish.go itself implements: the call reaches the embedded helper and
// its error is propagated unchanged, with the wrapper adding no state,
// validation or failure modes of its own.

import (
	"context"
	"strings"
	"testing"

	resourceapi "k8s.io/api/resource/v1"
	kubeletplugin "k8s.io/dynamic-resource-allocation/kubeletplugin"
	"k8s.io/dynamic-resource-allocation/resourceslice"
)

const wantNoNodeNameErr = "no NodeName was set to publish resources"

// TestResourcePublisher_PublishResources_DelegatesToHelper verifies the
// adapter forwards the call to the embedded helper and propagates its error
// unchanged. A zero-value helper fails the vendored nodeName validation
// deterministically before touching any client or mutex, making this path
// environment-independent.
func TestResourcePublisher_PublishResources_DelegatesToHelper(t *testing.T) {
	rp := &ResourcePublisher{Helper: &kubeletplugin.Helper{}}

	err := rp.PublishResources(context.Background(), resourceslice.DriverResources{})

	if err == nil {
		t.Fatal("PublishResources() = nil, want error")
	}
	if !strings.Contains(err.Error(), wantNoNodeNameErr) {
		t.Errorf("PublishResources() error = %v, want containing %q", err, wantNoNodeNameErr)
	}

	// The adapter is a stateless pass-through: a repeated call must behave
	// identically instead of panicking or caching the first result.
	if err := rp.PublishResources(context.Background(), resourceslice.DriverResources{}); err == nil {
		t.Error("second PublishResources() = nil, want same error again")
	} else if !strings.Contains(err.Error(), wantNoNodeNameErr) {
		t.Errorf("second PublishResources() error = %v, want containing %q", err, wantNoNodeNameErr)
	}
}

// TestResourcePublisher_PublishResources_AcceptsDriverResources verifies the
// adapter accepts a populated DriverResources payload: the vendored nodeName
// validation fires before any resource inspection, so the payload must flow
// through without the adapter itself validating, mutating or rejecting it.
func TestResourcePublisher_PublishResources_AcceptsDriverResources(t *testing.T) {
	rp := &ResourcePublisher{Helper: &kubeletplugin.Helper{}}
	resources := resourceslice.DriverResources{
		Pools: map[string]resourceslice.Pool{
			"pool-a": {
				Generation: 1,
				Slices: []resourceslice.Slice{{
					Devices: []resourceapi.Device{
						{Name: "Ascend910-0"},
						{Name: "Ascend910-1"},
					},
				}},
			},
		},
	}

	err := rp.PublishResources(context.Background(), resources)

	if err == nil {
		t.Fatal("PublishResources() = nil, want error")
	}
	if !strings.Contains(err.Error(), wantNoNodeNameErr) {
		t.Errorf("PublishResources() error = %v, want containing %q", err, wantNoNodeNameErr)
	}
}
