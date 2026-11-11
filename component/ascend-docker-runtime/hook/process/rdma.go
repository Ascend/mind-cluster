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

package process

import (
	"fmt"

	"ascend-common/rdma"
)

var discoverRdmaMountConfigs = rdma.DiscoverRdmaCollections

// Mark discovered collections for strict reading without changing ordinary collections.
func selectHookMountConfigs(configNames, hcaNames []string) ([]string, map[string]bool, error) {
	if len(hcaNames) == 0 {
		return configNames, nil, nil
	}

	collectionNames, err := discoverRdmaMountConfigs(configDir, hcaNames)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to discover RoCE RDMA mount collections: %w", err)
	}
	rdmaCollections := make(map[string]bool, len(collectionNames))
	for _, name := range collectionNames {
		rdmaCollections[name] = true
	}

	merged := make([]string, 0, len(configNames)+len(collectionNames))
	seen := make(map[string]bool, len(configNames)+len(collectionNames))
	for _, name := range append(append([]string(nil), configNames...), collectionNames...) {
		if !seen[name] {
			seen[name] = true
			merged = append(merged, name)
		}
	}
	return merged, rdmaCollections, nil
}
