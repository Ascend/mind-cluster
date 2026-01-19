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

	"gopkg.in/yaml.v2"

	"ascend-common/common-utils/utils"
	"ascend-faultdiag-online/pkg/core/model/enum"
	"ascend-faultdiag-online/pkg/utils/slicetool"
)

// Cluster represents the configuration for the cluster mode.
type Cluster struct {
	// Node report timeout in seconds
	NodeReportTimeout int `yaml:"node_report_timeout"`
	// All nodes report timeout in seconds
	AllNodesReportTimeout int `yaml:"all_nodes_report_timeout"`
}

// FaultDiagConfig represents the configuration settings for the fault diagnosis system.
type FaultDiagConfig struct {
	Mode      enum.DeployMode `yaml:"mode"`       // deploy mode: cluster or node
	LogLevel  enum.LogLevel   `yaml:"log_level"`  // 日志级别：debug, info, warn, error
	QueueSize int             `yaml:"queue_size"` // 数据队列大小

	Cluster `yaml:",inline"`
}

func paramCheck(config *FaultDiagConfig) error {
	if err := slicetool.ValueIn(config.Mode, []enum.DeployMode{enum.Cluster, enum.Node}); err != nil {
		return err
	}
	if err := slicetool.ValueIn(config.LogLevel, enum.LogLevels()); err != nil {
		return err
	}
	if config.QueueSize <= 0 {
		return fmt.Errorf("config wrong param: queue size %d must great than 0", config.QueueSize)
	}
	if config.Mode == enum.Cluster {
		if config.NodeReportTimeout <= 0 {
			return fmt.Errorf("config wrong param: node report timeout %d must great than 0", config.NodeReportTimeout)
		}
		if config.AllNodesReportTimeout <= 0 {
			return fmt.Errorf("config wrong param: all nodes report timeout %d must great than 0",
				config.AllNodesReportTimeout)
		}
	}
	return nil
}

// LoadConfig reads the configuration file and returns a FaultDiagConfig instance.
func LoadConfig(path string) (*FaultDiagConfig, error) {
	data, err := utils.LoadFile(path)
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
