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
Package config provides configuration management functionalities for the ascend-faultdiag-online application.
*/
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"

	"ascend-faultdiag-online/pkg/model/enum"
	"ascend-faultdiag-online/pkg/utils/slicetool"
)

// FaultDiagConfig represents the configuration settings for the fault diagnosis system.
type FaultDiagConfig struct {
	Mode      enum.DeployMode `yaml:"mode"`       // 运行模式： "cluster" 或 "node"
	SoDir     string          `yaml:"so_dir"`     // .so 文件目录
	LogLevel  enum.LogLevel   `yaml:"log_level"`  // 日志级别：debug, info, warn, error
	QueueSize int             `yaml:"queue_size"` // 数据队列大小
}

func paramCheck(config *FaultDiagConfig) error {
	if err := slicetool.ValueIn(config.Mode, enum.DeployModes()); err != nil {
		return err
	}
	if err := slicetool.ValueIn(config.LogLevel, enum.LogLevels()); err != nil {
		return err
	}
	if config.QueueSize <= 0 {
		return fmt.Errorf("config wrong param: queue size %d must great than 0", config.QueueSize)
	}
	return nil
}

// LoadConfig reads the configuration file and returns a FaultDiagConfig instance.
func LoadConfig(path string) (*FaultDiagConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var config FaultDiagConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	if err := paramCheck(&config); err != nil {
		return nil, err
	}
	return &config, nil
}
