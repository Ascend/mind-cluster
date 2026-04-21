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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml"

	"ascend-common/common-utils/hwlog"
	"ascend-docker-runtime/mindxcheckutils"
)

func getConfigVersion(t *toml.Tree) int64 {
	switch v := t.Get("version").(type) {
	case int64:
		return v
	default:
		return configVersion2
	}
}

func getCriRuntimePluginName(configVersion int64) string {
	switch configVersion {
	case configVersion1:
		return version1RuntimePluginName
	case configVersion2:
		return version2RuntimePluginName
	default:
		return version3RuntimePluginName
	}
}

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

type commandArgs struct {
	action          string
	srcFilePath     string
	runtimeFilePath string
	destFilePath    string
}

// ContainerdProcess modifies the containerd configuration file when installing or uninstalling the containerd scenario.
func ContainerdProcess(command []string) (string, error) {
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

	err := editContainerdConfig(arg)
	if err != nil {
		hwlog.RunLog.Errorf("failed to edit containerd config, err: %v", err)
		return behavior, err
	}

	return behavior, nil
}

func editContainerdConfig(arg *commandArgs) error {
	if arg == nil {
		return errors.New("arg is nil")
	}
	configTree, err := toml.LoadFile(arg.srcFilePath)
	if err != nil {
		return err
	}
	version := getConfigVersion(configTree)
	criRuntimePluginName := getCriRuntimePluginName(version)
	if arg.action == addCommand {
		// Add Ascend runtime
		err = addRuntime(runtimeName, arg.runtimeFilePath, configTree, criRuntimePluginName)
		if err != nil {
			hwlog.RunLog.Errorf("failed to add Ascend runtime, error: %v", err)
			return err
		}
	} else if arg.action == rmCommand {
		// Remove Ascend runtime
		err = removeRuntime(runtimeName, configTree, criRuntimePluginName)
		if err != nil {
			hwlog.RunLog.Errorf("failed to remove Ascend runtime, error: %v", err)
			return err
		}
	}

	// Save config to file
	err = writeContainerdConfigToFile(configTree, arg.destFilePath)
	if err != nil {
		hwlog.RunLog.Errorf("failed to write configuration file: %v", err)
		return err
	}

	return nil
}

func getDefaultRuntimeOptions(t *toml.Tree, criRuntimePluginName string) (interface{}, error) {
	defaultOptions, err := toml.TreeFromMap(map[string]interface{}{
		"runtime_type":                    v2RuncRuntimeType,
		"runtime_root":                    "",
		"runtime_engine":                  "",
		"privileged_without_host_devices": false,
	})
	if err != nil {
		hwlog.RunLog.Errorf("TreeFromMap failed: %v", err)
		return nil, err
	}
	defaultOptions.SetPath([]string{optionsKey, systemdCgroupKey}, true)
	if t == nil {
		hwlog.RunLog.Warn("Tree is nil, could not infer options from runtimes runc")
		return defaultOptions, nil
	}
	options := getSubtreeByPath([]string{pluginsKey, criRuntimePluginName, containerdKey, runtimesKey,
		defaultRuntimeValue}, t)
	if options != nil {
		hwlog.RunLog.Infof("Using options from runtime runc: %v", options)
		newOptions := copy(options)
		if newOptions != nil {
			return newOptions, nil
		}
	}
	hwlog.RunLog.Warn("Could not infer options from runtimes runc")
	return defaultOptions, nil
}

// addRuntime adds a runtime to the containerd config
func addRuntime(name string, path string, tree *toml.Tree, criRuntimePluginName string) error {
	if tree == nil {
		return fmt.Errorf("config tree is nil")
	}
	// Create default runtime options
	defaultOptions, err := getDefaultRuntimeOptions(tree, criRuntimePluginName)
	if err != nil {
		hwlog.RunLog.Errorf("getDefaultRuntimeOptions failed: %v", err)
		return err
	}
	// Set runtime options
	tree.SetPath([]string{pluginsKey, criRuntimePluginName, containerdKey, runtimesKey, name}, defaultOptions)
	// Set binary path
	tree.SetPath([]string{pluginsKey, criRuntimePluginName, containerdKey, runtimesKey, name,
		optionsKey, binaryNameKey}, path)
	// Set default_runtime_name
	tree.SetPath([]string{pluginsKey, criRuntimePluginName, containerdKey, defaultRuntimeNameKey}, name)
	return nil
}

// removeRuntime removes a runtime from the containerd config
func removeRuntime(name string, tree *toml.Tree, criRuntimePluginName string) error {
	if tree == nil {
		return nil
	}
	// Set default_runtime_name
	tree.SetPath([]string{pluginsKey, criRuntimePluginName, containerdKey, defaultRuntimeNameKey}, defaultRuntimeValue)
	// Remove runtime configuration
	runtimePath := []string{pluginsKey, criRuntimePluginName, containerdKey, runtimesKey, name}
	err := tree.DeletePath(runtimePath)
	if err != nil {
		// If path doesn't exist, ignore error
		if !strings.Contains(err.Error(), "path not found") {
			hwlog.RunLog.Errorf("failed to remove runtime, error: %v", err)
			return err
		}
	}

	// Clean up empty parent directories
	for i := 1; i < len(runtimePath); i++ {
		parentPath := runtimePath[:len(runtimePath)-i]
		parentNode := getSubtreeByPath(parentPath, tree)
		if parentNode != nil && len(parentNode.Keys()) == 0 {
			tree.DeletePath(parentPath)
		} else {
			// If parent has other keys, stop cleaning up
			break
		}
	}

	return nil
}

func writeContainerdConfigToFile(configTree *toml.Tree, destFilePath string) error {
	if configTree == nil {
		return fmt.Errorf("config tree is nil")
	}

	// Marshal config to TOML
	tomlData, err := configTree.Marshal()
	if err != nil {
		return fmt.Errorf("unable to convert to TOML: %v", err)
	}

	// Write to file
	file, err := os.OpenFile(destFilePath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, perm)
	if err != nil {
		hwlog.RunLog.Errorf("failed to open file for writing: %v", err)
		return err
	}
	defer file.Close()

	_, err = file.Write(tomlData)
	if err != nil {
		hwlog.RunLog.Errorf("failed to write config to file: %v", err)
		return err
	}

	return nil
}
