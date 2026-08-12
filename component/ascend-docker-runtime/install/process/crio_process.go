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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml"

	"ascend-common/common-utils/hwlog"
	"ascend-docker-runtime/mindxcheckutils"
)

// crioRuntimePath returns a fresh TOML path under [crio.runtime.runtimes].
// Variadic args are appended so callers can address the runtime table itself
// or a leaf key inside it without mutating a shared backing slice.
func crioRuntimePath(name ...string) []string {
	base := []string{crioSectionKey, crioRuntimeSectionKey, crioRuntimesSectionKey}
	return append(base, name...)
}

// CriOProcess modifies the CRI-O configuration drop-in when installing or
// uninstalling the ascend runtime under the CRI-O container engine.
func CriOProcess(command []string) (string, error) {
	if len(command) == 0 {
		return "", fmt.Errorf("error param, length of command is 0")
	}
	action := command[actionPosition]
	correctParam, behavior := checkParamAndGetBehavior(action, command)
	if !correctParam {
		return "", fmt.Errorf("error param")
	}
	srcFilePath := command[srcFilePosition]
	if _, err := os.Stat(srcFilePath); !os.IsNotExist(err) {
		if _, err := mindxcheckutils.RealFileChecker(srcFilePath, true, false, mindxcheckutils.DefaultSize); err != nil {
			hwlog.RunLog.Errorf("check failed, error: %v", err)
			return behavior, err
		}
	}
	destFilePath := command[destFilePosition]
	if _, err := mindxcheckutils.RealDirChecker(filepath.Dir(destFilePath), true, false); err != nil {
		return behavior, err
	}

	runtimeFilePath := ""
	if len(command) == addCommandLength {
		runtimeFilePath = command[runtimeFilePosition]
		if _, err := mindxcheckutils.RealFileChecker(runtimeFilePath, true, false, mindxcheckutils.DefaultSize); err != nil {
			hwlog.RunLog.Errorf("failed to check, error: %v", err)
			return behavior, err
		}
	}

	arg := &commandArgs{
		action:          action,
		srcFilePath:     srcFilePath,
		runtimeFilePath: runtimeFilePath,
		destFilePath:    destFilePath,
	}

	if err := editCriOConfig(arg); err != nil {
		hwlog.RunLog.Errorf("failed to edit crio config, err: %v", err)
		return behavior, err
	}

	return behavior, nil
}

// editCriOConfig loads the CRI-O drop-in, applies the add/remove mutation and
// writes the result back. A missing source file is treated as an empty drop-in
// so a fresh /etc/crio/crio.conf.d/99-ascend-runtime.conf can be created.
func editCriOConfig(arg *commandArgs) error {
	if arg == nil {
		return errors.New("arg is nil")
	}
	configTree, err := toml.LoadFile(arg.srcFilePath)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		// Drop-in does not exist yet: start from an empty tree.
		configTree, err = toml.TreeFromMap(map[string]interface{}{})
		if err != nil {
			return err
		}
	}

	if arg.action == addCommand {
		if err = addCriORuntime(runtimeName, arg.runtimeFilePath, configTree); err != nil {
			hwlog.RunLog.Errorf("failed to add Ascend runtime, error: %v", err)
			return err
		}
	} else if arg.action == rmCommand {
		if err = removeCriORuntime(runtimeName, configTree); err != nil {
			hwlog.RunLog.Errorf("failed to remove Ascend runtime, error: %v", err)
			return err
		}
	}

	return writeTomlConfigToFile(configTree, arg.destFilePath)
}

// defaultCriORuntimeOptions returns a CRI-O runtime table with sane defaults,
// used when the config has no runc entry to inherit settings from. Mirrors the
// containerd install scene's fallback.
func defaultCriORuntimeOptions() (interface{}, error) {
	return toml.TreeFromMap(map[string]interface{}{
		crioRuntimeTypeKey: crioOciRuntimeType,
		crioRuntimeRootKey: "",
	})
}

// inheritCriORuntimeOptions copies the [crio.runtime.runtimes.runc] subtree so
// the ascend runtime inherits the site's runtime_type / runtime_root etc., only
// overriding the binary path. Mirrors the containerd install scene. Falls back
// to defaults when the runc entry is absent.
func inheritCriORuntimeOptions(tree *toml.Tree) (interface{}, error) {
	def, err := defaultCriORuntimeOptions()
	if err != nil {
		return nil, err
	}
	if tree == nil {
		return def, nil
	}
	runcOptions := getSubtreeByPath(crioRuntimePath(defaultRuntimeValue), tree)
	if runcOptions != nil {
		if copied := copy(runcOptions); copied != nil {
			return copied, nil
		}
	}
	return def, nil
}

// addCriORuntime registers the ascend OCI runtime under
// [crio.runtime.runtimes.ascend], inheriting runc's options (only the binary
// path is ascend-specific), and sets it as the CRI-O default runtime.
func addCriORuntime(name string, path string, tree *toml.Tree) error {
	if tree == nil {
		return fmt.Errorf("config tree is nil")
	}
	options, err := inheritCriORuntimeOptions(tree)
	if err != nil {
		hwlog.RunLog.Errorf("failed to inherit crio runtime options: %v", err)
		return err
	}
	tree.SetPath(crioRuntimePath(name), options)
	// Only the binary path is ascend-specific; runtime_type / runtime_root
	// follow the site's runc configuration (or the defaults above).
	tree.SetPath(crioRuntimePath(name, crioRuntimePathKey), path)
	tree.SetPath([]string{crioSectionKey, crioRuntimeSectionKey, crioDefaultRuntimeKey}, name)
	return nil
}

// removeCriORuntime restores the default runtime to runc and deletes the
// ascend runtime table together with any empty parent tables.
func removeCriORuntime(name string, tree *toml.Tree) error {
	if tree == nil {
		return nil
	}
	tree.SetPath([]string{crioSectionKey, crioRuntimeSectionKey, crioDefaultRuntimeKey}, defaultRuntimeValue)

	runtimePath := crioRuntimePath(name)
	if err := tree.DeletePath(runtimePath); err != nil {
		// If path doesn't exist, ignore error.
		if !strings.Contains(err.Error(), "path not found") &&
			!strings.Contains(err.Error(), "no such key to delete") {
			hwlog.RunLog.Errorf("failed to remove runtime, error: %v", err)
			return err
		}
	}

	// Clean up empty parent directories.
	for i := 1; i < len(runtimePath); i++ {
		parentPath := runtimePath[:len(runtimePath)-i]
		parentNode := getSubtreeByPath(parentPath, tree)
		if parentNode != nil && len(parentNode.Keys()) == 0 {
			if err := tree.DeletePath(parentPath); err != nil {
				if !strings.Contains(err.Error(), "path not found") &&
					!strings.Contains(err.Error(), "no such key to delete") {
					return err
				}
			}
		} else {
			// If parent has other keys, stop cleaning up.
			break
		}
	}

	return nil
}
