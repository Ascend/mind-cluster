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

// Package utils for common utils
package utils

import (
	"os"
	"strings"
)

// NormalizeVarRunPath replaces the /var/run prefix only when it is the standard
// symlink to /run. It does not resolve other symlinks or validate the path;
// callers must still perform path and permission checks before using the result.
func NormalizeVarRunPath(path string) string {
	return normalizeVarRunPath(path, os.Readlink)
}

func normalizeVarRunPath(path string, readlink func(string) (string, error)) string {
	const prefix = "/var/run"
	if path != prefix && !strings.HasPrefix(path, prefix+"/") {
		return path
	}
	target, err := readlink(prefix)
	if err == nil && (target == "/run" || target == "../run") {
		return "/run" + strings.TrimPrefix(path, prefix)
	}
	return path
}
