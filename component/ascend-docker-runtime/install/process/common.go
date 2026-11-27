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
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml"

	"ascend-common/common-utils/hwlog"
	"ascend-docker-runtime/mindxcheckutils"
)

type commandArgs struct {
	action          string
	srcFilePath     string
	runtimeFilePath string
	destFilePath    string
}

// parseAndValidateCommand validates all paths before any engine edits its config.
// A missing source still requires a safe parent directory.

func parseAndValidateCommand(command []string) (*commandArgs, string, error) {

	if len(command) == 0 {
		return nil, "", fmt.Errorf("error param, length of command is 0")
	}

	action := command[actionPosition]

	correctParam, behavior := checkParamAndGetBehavior(action, command)
	if !correctParam {
		return nil, "", fmt.Errorf("error param")
	}

	arg := &commandArgs{
		action:       action,
		srcFilePath:  command[srcFilePosition],
		destFilePath: command[destFilePosition],
	}
	if _, err := os.Stat(arg.srcFilePath); os.IsNotExist(err) {
		if _, err := mindxcheckutils.RealDirChecker(filepath.Dir(arg.srcFilePath), true, false); err != nil {
			return nil, behavior, err
		}
	} else {
		if _, err := mindxcheckutils.RealFileChecker(arg.srcFilePath, true, false,
			mindxcheckutils.DefaultSize); err != nil {
			return nil, behavior, err
		}
	}

	if _, err := mindxcheckutils.RealDirChecker(filepath.Dir(arg.destFilePath), true, false); err != nil {
		return nil, behavior, err
	}

	if action == addCommand {
		arg.runtimeFilePath = command[runtimeFilePosition]
		if _, err := mindxcheckutils.RealFileChecker(arg.runtimeFilePath, true, false,
			mindxcheckutils.DefaultSize); err != nil {
			return nil, behavior, err
		}
	}
	return arg, behavior, nil
}

// deleteRuntimeAndEmptyParents preserves sibling keys and never deletes the root.

func deleteRuntimeAndEmptyParents(tree *toml.Tree, runtimePath []string) error {
	if tree == nil || len(runtimePath) == 0 {
		return nil
	}
	if err := deleteConfigPath(tree, runtimePath); err != nil {
		return err
	}
	for length := len(runtimePath) - 1; length > 0; length-- {
		parentPath := runtimePath[:length]
		parentNode := getSubtreeByPath(parentPath, tree)
		if parentNode == nil || len(parentNode.Keys()) != 0 {
			break
		}
		if err := deleteConfigPath(tree, parentPath); err != nil {
			return err
		}
	}
	return nil
}

func deleteConfigPath(tree *toml.Tree, path []string) error {
	if err := tree.DeletePath(path); err != nil {
		if !strings.Contains(err.Error(), tomlPathNotFoundMessage) &&
			!strings.Contains(err.Error(), tomlNoSuchKeyMessage) {
			return err
		}
	}
	return nil
}

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

// getSubtreeByPath returns a TOML table at keys, or nil for a missing or non-table value.
func getSubtreeByPath(keys []string, t *toml.Tree) *toml.Tree {
	subtree := t.GetPath(keys)
	if subtree == nil {
		return nil
	}

	switch subtree := subtree.(type) {
	case *toml.Tree:
		return subtree
	default:
		hwlog.RunLog.Errorf("invalid subtree type %T", subtree)
		return nil
	}
}

// copy clones a TOML tree so runtime options can be edited independently.
func copy(t *toml.Tree) *toml.Tree {
	if t == nil {
		return nil
	}
	copyTree, err := toml.Load(t.String())
	if err != nil {
		hwlog.RunLog.Errorf("failed to load toml: %v", err)
		return nil
	}
	return copyTree
}
