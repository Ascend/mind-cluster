/* Copyright(C) 2024. Huawei Technologies Co.,Ltd. All rights reserved.
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
	"os"

	"github.com/pelletier/go-toml"

	"ascend-common/common-utils/hwlog"
)

func checkParamAndGetBehavior(action string, command []string) (bool, string) {
	correctParam, behavior := false, ""
	if action == addCommand && len(command) == addCommandLength {
		correctParam = true
		behavior = "install"
	}
	if action == rmCommand && len(command) == rmCommandLength {
		correctParam = true
		behavior = "uninstall"
	}
	return correctParam, behavior
}

// CheckParamLength whether the param length is valid
func CheckParamLength(command []string) bool {
	return len(command) == addCommandLength || len(command) == rmCommandLength
}

// writeTomlConfigToFile marshals the toml tree to destFilePath (created or
// truncated, mode perm). Shared by the containerd and CRI-O install scenes.
// Close/Sync errors are captured via the named return so flush-time failures
// (disk full, NFS, quota) are not silently swallowed.
func writeTomlConfigToFile(configTree *toml.Tree, destFilePath string) (err error) {
	if configTree == nil {
		return fmt.Errorf("config tree is nil")
	}

	tomlData, err := configTree.Marshal()
	if err != nil {
		return fmt.Errorf("unable to convert to TOML: %v", err)
	}

	file, err := os.OpenFile(destFilePath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, perm)
	if err != nil {
		hwlog.RunLog.Errorf("failed to open file for writing: %v", err)
		return err
	}
	defer func() {
		if cerr := file.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("failed to close config file: %v", cerr)
		}
	}()

	if _, err = file.Write(tomlData); err != nil {
		hwlog.RunLog.Errorf("failed to write config to file: %v", err)
		return err
	}
	if err = file.Sync(); err != nil {
		hwlog.RunLog.Errorf("failed to sync config file: %v", err)
		return err
	}

	return nil
}
