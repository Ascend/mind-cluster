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
	"os"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/utils"
	"ascend-faultdiag-online/pkg/core/model/enum"
)

func TestParamCheck(t *testing.T) {
	var queueSize = 10
	convey.Convey("test paramCheck", t, func() {
		config := &FaultDiagConfig{
			Mode:      enum.Cluster,
			LogLevel:  enum.LgDebug,
			QueueSize: queueSize,
			Cluster: Cluster{
				NodeReportTimeout:     30,
				AllNodesReportTimeout: 60,
			},
		}
		convey.Convey(("invalid mod"), func() {
			config.Mode = "invalid mode"
			err := paramCheck(config)
			convey.So(err.Error(), convey.ShouldEqual, "the parameter invalid mode is not in the list: [cluster node]")
			config.Mode = enum.Cluster
		})
		convey.Convey(("valid config"), func() {
			err := paramCheck(config)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey(("invalid log level"), func() {
			config.LogLevel = "invalid"
			err := paramCheck(config)
			convey.So(err.Error(), convey.ShouldEqual,
				"the parameter invalid is not in the list: [info debug warn error]")
			config.LogLevel = enum.LgInfo // reset to valid value
		})
		convey.Convey(("invalid QueueSize"), func() {
			config.QueueSize = -1
			err := paramCheck(config)
			convey.So(err.Error(), convey.ShouldEqual, "config wrong param: queue size -1 must great than 0")
			config.QueueSize = queueSize // reset to valid value
		})
		convey.Convey(("invalid NodeReportTimeout"), func() {
			config.NodeReportTimeout = -1
			err := paramCheck(config)
			convey.So(err.Error(), convey.ShouldEqual, "config wrong param: node report timeout -1 must great than 0")
			config.NodeReportTimeout = 30 // reset to valid value
		})
		convey.Convey(("invalid AllNodesReportTimeout"), func() {
			config.AllNodesReportTimeout = 0
			err := paramCheck(config)
			convey.So(err.Error(), convey.ShouldEqual,
				"config wrong param: all nodes report timeout 0 must great than 0")
			config.AllNodesReportTimeout = 60 // reset to valid value
		})
	})
}

func writeTempFile(filePath string, data string) (*os.File, error) {
	tmpFile, err := os.CreateTemp("", filePath)
	if err != nil {
		return nil, err
	}
	_, err = tmpFile.Write([]byte(data))
	if err != nil {
		return nil, err
	}
	err = tmpFile.Close()
	if err != nil {
		return nil, err
	}
	return tmpFile, nil
}

func removeFile(file *os.File) error {
	if file == nil {
		return errors.New("file is nil")
	}
	return os.Remove(file.Name())
}

func TestLoadConfig(t *testing.T) {
	convey.Convey("test LoadConfig", t, func() {
		testLoadConfigWithSuccess()
		testLoadConfigWithError()

	})
}

func testLoadConfigWithError() {
	convey.Convey("load file failed", func() {
		patch := gomonkey.ApplyFunc(utils.LoadFile, func(path string) ([]byte, error) {
			return nil, errors.New("load file error")
		})
		defer patch.Reset()
		_, err := LoadConfig("non_existent_file.yaml")
		convey.So(err.Error(), convey.ShouldEqual, "load file error")
	})
	convey.Convey("file yaml unmarshal failed", func() {
		yamlData := "}"
		tmpFile, err := writeTempFile("invalid_file.yaml", yamlData)
		convey.So(err, convey.ShouldBeNil)
		_, err = LoadConfig(tmpFile.Name())
		convey.So(err.Error(), convey.ShouldEqual, "yaml: did not find expected node content")
		convey.So(removeFile(tmpFile), convey.ShouldBeNil)
	})
	convey.Convey("param check failed", func() {
		yamlData := "log_level: info\nqueue_size: 10\nnode_report_timeout: -1\nall_nodes_report_timeout: 1200"
		tmpFile, err := writeTempFile("valid_config.yaml", yamlData)
		convey.So(err, convey.ShouldBeNil)
		loadedConfig, err := LoadConfig(tmpFile.Name())
		convey.So(loadedConfig, convey.ShouldBeNil)
		convey.So(err.Error(), convey.ShouldEqual, "the parameter  is not in the list: [cluster node]")
		convey.So(removeFile(tmpFile), convey.ShouldBeNil)
	})
}

func testLoadConfigWithSuccess() {
	var NodeReportTimeout = 7200
	var AllNodesReportTimeout = 1200
	var queueSize = 10
	var yamlData = `
mode: cluster
log_level: info
queue_size: 10
node_report_timeout: 7200
all_nodes_report_timeout: 1200`
	convey.Convey("test load config with success", func() {
		tmpFile, err := writeTempFile("valid_config.yaml", yamlData)
		convey.So(err, convey.ShouldBeNil)
		loadedConfig, err := LoadConfig(tmpFile.Name())
		convey.So(err, convey.ShouldBeNil)
		convey.So(loadedConfig.Mode, convey.ShouldEqual, enum.Cluster)
		convey.So(loadedConfig.LogLevel, convey.ShouldEqual, enum.LgInfo)
		convey.So(loadedConfig.QueueSize, convey.ShouldEqual, queueSize)
		convey.So(loadedConfig.NodeReportTimeout, convey.ShouldEqual, NodeReportTimeout)
		convey.So(loadedConfig.AllNodesReportTimeout, convey.ShouldEqual, AllNodesReportTimeout)
		convey.So(removeFile(tmpFile), convey.ShouldBeNil)
	})
}
