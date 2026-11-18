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

package enpuconfig

import (
	"fmt"
	"path"
	"strings"
)

// CleanConfigPath checks that a host path is inside the supplied allowlisted directory.
func CleanConfigPath(root, hostPath string) (string, error) {
	if !path.IsAbs(root) || path.Clean(root) == "/" || strings.ContainsRune(root, 0) {
		return "", fmt.Errorf("invalid enpu-manager configuration root")
	}
	if !path.IsAbs(hostPath) || strings.ContainsRune(hostPath, 0) {
		return "", fmt.Errorf("enpu-manager configuration path must be absolute")
	}
	cleaned := path.Clean(hostPath)
	if !strings.HasPrefix(cleaned, path.Clean(root)+"/") {
		return "", fmt.Errorf("enpu-manager configuration path is outside the allowlisted directory")
	}
	return cleaned, nil
}
