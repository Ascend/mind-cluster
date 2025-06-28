/*
Copyright(C)2025. Huawei Technologies Co.,Ltd. All rights reserved.

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

/*
Package pingmeshv1 is using for checking hccs network
*/
package pingmeshv1

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"nodeD/pkg/pingmeshv1/types"
)

func generateJobUID(config *types.HccspingMeshConfig, destAddrs map[string]types.SuperDeviceIDs) (string, error) {
	cfg, err := json.Marshal(config)
	if err != nil {
		return "", err
	}
	address, err := json.Marshal(destAddrs)
	if err != nil {
		return "", err
	}
	address = append(address, cfg...)
	hasher := sha256.New()
	_, err = hasher.Write(address)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil

}
