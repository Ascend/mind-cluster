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

// Package mount — mount.go
//
// Mount config types: MountEntry and the UB type marker.
package mount

// ubType marks a mount entry group as UB user-space driver files, gated by
// the external DisableUBMounts option.
const ubType = "UB"

// MountEntry is a normalized group of mount entries read from a mount source,
// before glob expansion and existence checks.
type MountEntry struct {
	// Paths are the host paths in this group, or glob patterns when they
	// contain wildcard characters (*?[).
	Paths []string `json:"path"`

	// Type marks the group; "UB" for UB user-space files (gated by
	// DisableUBMounts), empty otherwise.
	Type string `json:"type,omitempty"`
}
