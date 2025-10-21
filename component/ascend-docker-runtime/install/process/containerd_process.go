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

	"github.com/containerd/containerd/services/server/config"
	"github.com/pelletier/go-toml"

	"ascend-common/common-utils/hwlog"
	"ascend-docker-runtime/mindxcheckutils"
)

type commandArgs struct {
	action          string
	srcFilePath     string
	runtimeFilePath string
	destFilePath    string
	cgroupInfo      string
	osName          string
	osVersion       string
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
	if _, err := os.Stat(srcFilePath); os.IsNotExist(err) {
		if _, err := mindxcheckutils.RealDirChecker(filepath.Dir(srcFilePath), true, false); err != nil {
			hwlog.RunLog.Errorf("check failed, error: %v", err)
			return behavior, err
		}
	} else {
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
		cgroupInfo:      command[len(command)-cgroupInfoIndexFromEnd],
		osName:          command[len(command)-osNameIndexFromEnd],
		osVersion:       command[len(command)-osVersionIndexFromEnd],
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
		hwlog.RunLog.Error("arg is nil")
		return errors.New("arg is nil")
	}
	cfg := config.Config{}
	if err := config.LoadConfig(arg.srcFilePath, &cfg); err != nil {
		hwlog.RunLog.Errorf("failed to load configuration file: %v", err)
		return err
	}
	if strings.Contains(arg.cgroupInfo, cgroupV2InfoStr) {
		hwlog.RunLog.Info("it is cgroup v2")
		binaryName := ""
		if arg.action == addCommand {
			binaryName = arg.runtimeFilePath
		}
		err := changeCgroupV2BinaryNameConfig(&cfg, binaryName)
		if err != nil {
			hwlog.RunLog.Errorf("failed to change cgroup v2 config, error: %v", err)
			return err
		}
	} else {
		hwlog.RunLog.Info("it is cgroup v1")
		runtimeValue := defaultRuntimeValue
		runtimeType := v2RuncRuntimeType
		if arg.action == addCommand {
			runtimeValue = arg.runtimeFilePath
			runtimeType = v1RuntimeType
			if arg.osName == openEulerStr && arg.osVersion == openEulerVersionForV2RuntimeType {
				runtimeType = v2RuncRuntimeType
			}
		}
		err := changeCgroupV1Config(&cfg, runtimeValue, runtimeType)
		if err != nil {
			hwlog.RunLog.Errorf("failed to change cgroup v1 config, error: %v", err)
			return err
		}
	}
	err := writeContainerdConfigToFile(cfg, arg.destFilePath)
	if err != nil {
		hwlog.RunLog.Errorf("failed to write configuration file: %v", err)
		return err
	}
	return nil
}

func changeCgroupV2BinaryNameConfig(cfg *config.Config, binaryName string) error {
	value, ok := cfg.Plugins[v1RuntimeTypeFirstLevelPlugin]
	if !ok {
		hwlog.RunLog.Errorf(notFindPluginLogStr, v1RuntimeTypeFirstLevelPlugin, cfg.Plugins)
		return fmt.Errorf(notFindPluginErrorStr, v1RuntimeTypeFirstLevelPlugin)
	}
	valueMap := value.ToMap()
	containerdConfig := valueMap[containerdKey]
	runtimesConfig, err := getMap(containerdConfig, runtimesKey)
	if err != nil {
		hwlog.RunLog.Errorf(getMapFaileLogStr, runtimesKey, err)
		return err
	}
	runcConfig, err := getMap(runtimesConfig, runcKey)
	if err != nil {
		hwlog.RunLog.Errorf(getMapFaileLogStr, runcKey, err)
		return err
	}
	runcOptionsConfig, err := getMap(runcConfig, runcOptionsKey)
	if err != nil {
		hwlog.RunLog.Errorf(getMapFaileLogStr, runcOptionsKey, err)
		return err
	}
	runcOptionsConfigMap, ok := runcOptionsConfig.(map[string]interface{})
	if !ok {
		hwlog.RunLog.Errorf(convertConfigFailLogStr, runcOptionsKey, runcOptionsConfig)
		return fmt.Errorf(convertConfigFailErrorStr, runcOptionsKey, runcOptionsConfig)
	}
	runcOptionsConfigMap[binaryNameKey] = binaryName
	newTree, err := toml.TreeFromMap(valueMap)
	if err != nil {
		hwlog.RunLog.Errorf(convertTreeFailLogStr, err)
		return err
	}
	cfg.Plugins[v1RuntimeTypeFirstLevelPlugin] = *newTree
	return nil
}

func changeCgroupV1Config(cfg *config.Config, runtimeValue, runtimeType string) error {
	err := changeCgroupV1RuntimeConfig(cfg, runtimeValue)
	if err != nil {
		hwlog.RunLog.Errorf("failed to change cgroup V1 runtime config, error: %v", err)
		return err
	}
	return changeCgroupV1RuntimeTypeConfig(cfg, runtimeType)
}

func changeCgroupV1RuntimeConfig(cfg *config.Config, runtimeValue string) error {
	value, ok := cfg.Plugins[v1RuntimeType]
	if !ok {
		hwlog.RunLog.Errorf(notFindPluginLogStr, v1RuntimeType, cfg.Plugins)
		return fmt.Errorf(notFindPluginErrorStr, v1RuntimeType)
	}
	valueMap := value.ToMap()
	valueMap[v1NeedChangeKeyRuntime] = runtimeValue
	newTree, err := toml.TreeFromMap(valueMap)
	if err != nil {
		hwlog.RunLog.Errorf(convertTreeFailLogStr, err)
		return err
	}
	cfg.Plugins[v1RuntimeType] = *newTree
	return nil
}

func changeCgroupV1RuntimeTypeConfig(cfg *config.Config, runtimeType string) error {
	value, ok := cfg.Plugins[v1RuntimeTypeFirstLevelPlugin]
	if !ok {
		hwlog.RunLog.Errorf(notFindPluginLogStr, v1RuntimeTypeFirstLevelPlugin, cfg.Plugins)
		return fmt.Errorf(notFindPluginErrorStr, v1RuntimeTypeFirstLevelPlugin)
	}
	valueMap := value.ToMap()
	containerdConfig := valueMap[containerdKey]
	runtimesConfig, err := getMap(containerdConfig, runtimesKey)
	if err != nil {
		hwlog.RunLog.Errorf(getMapFaileLogStr, runtimesKey, err)
		return err
	}
	runcConfig, err := getMap(runtimesConfig, runcKey)
	if err != nil {
		hwlog.RunLog.Errorf(getMapFaileLogStr, runcKey, err)
		return err
	}
	runcConfigMap, ok := runcConfig.(map[string]interface{})
	if !ok {
		hwlog.RunLog.Errorf(convertConfigFailLogStr, runcKey, runcConfig)
		return fmt.Errorf(convertConfigFailErrorStr, runcKey, runcConfig)
	}
	runcConfigMap[v1NeedChangeKeyRuntimeType] = runtimeType
	newTree, err := toml.TreeFromMap(valueMap)
	if err != nil {
		hwlog.RunLog.Errorf(convertTreeFailLogStr, err)
		return err
	}
	cfg.Plugins[v1RuntimeTypeFirstLevelPlugin] = *newTree
	return nil
}

func getMap(input interface{}, key string) (interface{}, error) {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		hwlog.RunLog.Errorf(convertConfigFailLogStr, key, input)
		return nil, fmt.Errorf(convertConfigFailErrorStr, key, input)
	}
	output, ok := inputMap[key]
	if !ok {
		hwlog.RunLog.Errorf("can not find config %v, config is: %+v", key, inputMap)
		return nil, fmt.Errorf("can not find config: %v", key)
	}
	return output, nil
}

func writeContainerdConfigToFile(cfg config.Config, destFilePath string) error {
	tomlString, err := toml.Marshal(cfg)
	if err != nil {
		hwlog.RunLog.Errorf("failed to marshall to toml, error: %v", err)
		return err
	}
	file, err := os.OpenFile(destFilePath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, perm)
	if err != nil {
		hwlog.RunLog.Errorf("failed to create file, error: %v", err)
		return err
	}
	defer func() {
		err := file.Close()
		if err != nil {
			hwlog.RunLog.Errorf("failed to close file, error: %v", err)
		}
	}()
	_, err = file.Write(tomlString)
	if err != nil {
		hwlog.RunLog.Errorf("failed to write, error: %v", err)
		return err
	}
	return nil
}
