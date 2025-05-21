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
Package config provides some test case for the config package.
*/
package config

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"

	"ascend-faultdiag-online/pkg/model/enum"
)

func TestParamCheck(t *testing.T) {
	testCases := []struct {
		name     string
		config   FaultDiagConfig
		expected error
	}{
		{
			name: "ValidConfig",
			config: FaultDiagConfig{
				Mode:      enum.Node,
				SoDir:     "/usr/lib",
				LogLevel:  enum.LgInfo,
				QueueSize: 10,
			},
			expected: nil,
		},
		{
			name: "InvalidMode",
			config: FaultDiagConfig{
				Mode:      "invalid_mode",
				SoDir:     "/usr/lib",
				LogLevel:  enum.LgInfo,
				QueueSize: 10,
			},
			expected: errors.New("the parameter invalid_mode is not in the list: [cluster node]"),
		},
		{
			name: "InvalidLogLevel",
			config: FaultDiagConfig{
				Mode:      enum.Cluster,
				SoDir:     "/usr/lib",
				LogLevel:  "invalid_level",
				QueueSize: 10,
			},
			expected: errors.New("the parameter invalid_level is not in the list: [info debug warn error]"),
		},
		{
			name: "InvalidQueueSize",
			config: FaultDiagConfig{
				Mode:      enum.Cluster,
				SoDir:     "/usr/lib",
				LogLevel:  enum.LgDebug,
				QueueSize: 0,
			},
			expected: fmt.Errorf("config wrong param: queue size 0 must great than 0"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := paramCheck(&tc.config)
			if tc.expected == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	validConfig := FaultDiagConfig{
		Mode:      enum.Cluster,
		SoDir:     "/usr/lib",
		LogLevel:  enum.LgInfo,
		QueueSize: 10,
	}

	// convert struct data to yaml
	data, err := yaml.Marshal(validConfig)
	assert.NoError(t, err)

	// create a temp file
	tmpFile, err := os.CreateTemp("", "config_*.yaml")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name()) // delete the temp file after the test

	// write the yaml file
	_, err = tmpFile.Write(data)
	assert.NoError(t, err)
	err = tmpFile.Close()
	assert.NoError(t, err)

	// read the config data
	loadedConfig, err := LoadConfig(tmpFile.Name())
	assert.NoError(t, err)
	assert.Equal(t, validConfig, *loadedConfig)

	// test case 1：non-existent file
	_, err = LoadConfig("non_existent.yaml")
	assert.Error(t, err)

	// test case 2：invalid yaml
	tmpFileInvalid, err := os.CreateTemp("", "config_invalid_*.yaml")
	assert.NoError(t, err)
	defer os.Remove(tmpFileInvalid.Name())

	_, err = tmpFileInvalid.Write([]byte("invalid_yaml: ::::"))
	assert.NoError(t, err)
	err = tmpFileInvalid.Close()
	assert.NoError(t, err)

	_, err = LoadConfig(tmpFileInvalid.Name())
	assert.Error(t, err)
}
