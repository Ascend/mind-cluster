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

// Package process deal the docker or containerd scene installation
package process

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"ascend-docker-runtime/mindxcheckutils"
)

var reserveDefaultRuntime = false

// DockerProcess modifies the docker configuration file when installing or uninstalling the docker scenario.
func DockerProcess(command []string) (string, error) {
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
			return behavior, err
		}
	} else {
		if _, err := mindxcheckutils.RealFileChecker(srcFilePath, true, false, mindxcheckutils.DefaultSize); err != nil {
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
			return behavior, err
		}
	}

	setReserveDefaultRuntime(command)

	// check file permission
	writeContent, err := createJsonString(srcFilePath, runtimeFilePath, action)
	if err != nil {
		return behavior, err
	}
	return behavior, writeJson(destFilePath, writeContent)
}

func createJsonString(srcFilePath, runtimeFilePath, action string) ([]byte, error) {
	var writeContent []byte
	if _, err := os.Stat(srcFilePath); err == nil {
		daemon, err := modifyDaemon(srcFilePath, runtimeFilePath, action)
		if err != nil {
			return nil, err
		}
		writeContent, err = json.MarshalIndent(daemon, "", "        ")
		if err != nil {
			return nil, err
		}
	} else if os.IsNotExist(err) {
		// not existed
		if !reserveDefaultRuntime {
			writeContent = []byte(fmt.Sprintf(commonTemplate, runtimeFilePath))
		} else {
			writeContent = []byte(fmt.Sprintf(noDefaultTemplate, runtimeFilePath))
		}
	} else {
		return nil, err
	}
	return writeContent, nil
}

func writeJson(destFilePath string, writeContent []byte) error {
	if _, err := os.Stat(destFilePath); os.IsNotExist(err) {
		const perm = 0600
		file, err := os.OpenFile(destFilePath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, perm)
		if err != nil {
			return fmt.Errorf("create target file failed")
		}
		_, err = file.Write(writeContent)
		if err != nil {
			closeErr := file.Close()
			return fmt.Errorf("write target file failed with close err %v", closeErr)
		}
		err = file.Close()
		if err != nil {
			return fmt.Errorf("close target file failed")
		}
		return nil
	} else {
		return fmt.Errorf("target file already existed")
	}
}

func modifyDaemon(srcFilePath, runtimeFilePath, action string) (map[string]interface{}, error) {
	// existed...
	daemon, err := loadOriginJson(srcFilePath)
	if err != nil {
		return nil, err
	}

	if _, ok := daemon["runtimes"]; !ok && action == addCommand {
		daemon["runtimes"] = map[string]interface{}{}
	}
	runtimeValue := daemon["runtimes"]
	runtimeConfig, runtimeConfigOk := runtimeValue.(map[string]interface{})
	if !runtimeConfigOk && action == addCommand {
		return nil, fmt.Errorf("extract runtime failed")
	}
	if action == addCommand {
		runtimeConfig, daemon, err = addDockerDaemon(runtimeConfig, daemon, runtimeFilePath)
	} else if action == rmCommand {
		runtimeConfig, daemon, err = rmDockerDaemon(runtimeConfig, daemon, runtimeConfigOk)
	} else {
		return nil, fmt.Errorf("param error")
	}
	return daemon, err
}

func addDockerDaemon(runtimeConfig, daemon map[string]interface{}, runtimeFilePath string,
) (map[string]interface{}, map[string]interface{}, error) {
	if runtimeConfig == nil {
		return nil, daemon, fmt.Errorf("runtime config is nil")
	}
	if _, ok := runtimeConfig["ascend"]; !ok {
		runtimeConfig["ascend"] = map[string]interface{}{}
	}
	ascendConfig, ok := runtimeConfig["ascend"].(map[string]interface{})
	if !ok {
		return nil, nil, fmt.Errorf("extract ascend failed")
	}
	ascendConfig["path"] = runtimeFilePath
	if _, ok := ascendConfig["runtimeArgs"]; !ok {
		ascendConfig["runtimeArgs"] = []string{}
	}
	if !reserveDefaultRuntime && daemon != nil {
		daemon[defaultRuntimeKey] = "ascend"
	}
	return runtimeConfig, daemon, nil
}

func rmDockerDaemon(runtimeConfig, daemon map[string]interface{}, runtimeConfigOk bool,
) (map[string]interface{}, map[string]interface{}, error) {
	if runtimeConfigOk {
		delete(runtimeConfig, "ascend")
	}
	if value, ok := daemon[defaultRuntimeKey]; ok && value == "ascend" {
		delete(daemon, defaultRuntimeKey)
	}
	return runtimeConfig, daemon, nil
}

func loadOriginJson(srcFilePath string) (map[string]interface{}, error) {
	if fileInfo, err := os.Stat(srcFilePath); err != nil {
		return nil, err
	} else if fileInfo.Size() > maxFileSize {
		return nil, fmt.Errorf("file size too large")
	}

	file, err := os.Open(srcFilePath)
	if err != nil {
		return nil, fmt.Errorf("open daemon.json failed, err: %v", err)
	}
	content, err := ioutil.ReadAll(file)
	if err != nil {
		closeErr := file.Close()
		return nil, fmt.Errorf("read daemon.json failed, close file err is %v", closeErr)
	}
	err = file.Close()
	if err != nil {
		return nil, fmt.Errorf("close daemon.json failed, err: %v", err)
	}

	var daemon map[string]interface{}
	err = json.Unmarshal(content, &daemon)
	if err != nil {
		return nil, fmt.Errorf("load daemon.json failed, err: %v", err)
	}
	return daemon, nil
}

func setReserveDefaultRuntime(command []string) {
	reserveCmdPostion := len(command) - reserveIndexFromEnd
	if command[reserveCmdPostion] == "yes" {
		reserveDefaultRuntime = true
	}
}
