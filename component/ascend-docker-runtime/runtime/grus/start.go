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

// Package grus, start of grus
package grus

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/opencontainers/runtime-spec/specs-go"

	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	"ascend-docker-runtime/runtime/common"
	"ascend-docker-runtime/runtime/grus/runtime"
)

func readConfigJson(path string) (*specs.Spec, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var config specs.Spec
	if err = json.Unmarshal(file, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func readOCIConfig(bundlePath string) (*specs.Spec, error) {
	var err error
	if bundlePath == "" {
		if bundlePath, err = os.Getwd(); err != nil {
			hwlog.RunLog.Errorf("get workdir failed: %v", err)
			return nil, err
		}
	}
	configSpec, err := readConfigJson(filepath.Join(bundlePath, "config.json"))
	if err != nil {
		hwlog.RunLog.Errorf("get config.json spec for path: %s, err: %v", bundlePath, err)
		return nil, err
	}
	return configSpec, nil
}

func addEnv(spec *specs.Spec, key, env string) {
	if spec == nil || spec.Process == nil || spec.Process.Env == nil {
		return
	}
	for i, envLine := range spec.Process.Env {
		words := strings.SplitN(envLine, "=", common.ENV_KEY_VALUE_MAX_PARTS)
		if len(words) != common.ENV_KEY_VALUE_MAX_PARTS {
			hwlog.RunLog.Errorf("environment error: %v", envLine)
		}
		if words[0] == key {
			spec.Process.Env[i] = env
			return
		}
	}
	spec.Process.Env = append(spec.Process.Env, env)

}

func Sstart(a *common.Args) error {
	if a.Bundle == "" {
		runtimeAPI := runtime.GetRuntime(common.RuntimeNameRunc, a.Root)
		if si, err := runtimeAPI.State(a.ContainerID); err == nil {
			a.Bundle = si.Bundle
		}
		hwlog.RunLog.Infof("sstart args - containerID: %s, root: %s, bundle: %s", a.ContainerID, a.Root, a.Bundle)
	}

	spec, err := readOCIConfig(a.Bundle)
	if err != nil {
		hwlog.RunLog.Errorf("ignore get checkpoint path of %s error: %v", a.ContainerID, err)
		return common.ExecRunc()
	}

	imgPath := getEnvFromSpec(spec, common.GRUS_SNAPSHOT_IMAGE_PATH)
	podName := getEnvFromSpec(spec, common.POD_NAME)
	podIndex, _ := utils.GetLastNumberFromString(podName)
	files, _ := os.ReadDir(filepath.Join(imgPath, podIndex))
	if imgPath != "" && podIndex != "" && len(files) != 0 {
		hwlog.RunLog.Infof("ignore start cmd for %s, because current is restore", a.ContainerID)
		return nil
	}

	return common.ExecRunc()
}
