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

// Package process — inject.go
//
// InjectEdits: applies CDI container edits into an OCI runtime spec.
package process

import (
	"fmt"

	"github.com/opencontainers/runtime-spec/specs-go"

	"tags.cncf.io/container-device-interface/pkg/cdi"
	cdispec "tags.cncf.io/container-device-interface/specs-go"
)

// InjectEdits applies a CDI Spec's container edits to an in-memory OCI spec.
// Shared edits (Spec.ContainerEdits) and per-device edits are both injected
// via the standard CDI library's ContainerEdits.Apply().
func InjectEdits(spec *specs.Spec, cdidSpec *cdispec.Spec) error {
	if spec == nil {
		return fmt.Errorf("cdi: spec cannot be nil")
	}
	if cdidSpec == nil {
		return nil
	}

	// Spec-level shared edits.
	if err := (&cdi.ContainerEdits{ContainerEdits: &cdidSpec.ContainerEdits}).Apply(spec); err != nil {
		return err
	}

	// Per-device edits.
	for _, dev := range cdidSpec.Devices {
		if err := (&cdi.ContainerEdits{ContainerEdits: &dev.ContainerEdits}).Apply(spec); err != nil {
			return err
		}
	}

	return nil
}
