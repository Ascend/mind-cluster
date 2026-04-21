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
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/pelletier/go-toml"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/hwlog"
	"ascend-docker-runtime/mindxcheckutils"
)

func init() {
	ctx := context.Background()
	hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, ctx)
}

var testError = errors.New("test")

// TestGetConfigVersion tests the function getConfigVersion
func TestGetConfigVersion(t *testing.T) {
	convey.Convey("Test getConfigVersion", t, func() {
		convey.Convey("version is int64", func() {
			tree, _ := toml.Load("version = 1")
			convey.So(getConfigVersion(tree), convey.ShouldEqual, int64(1))
		})
		convey.Convey("version is not int64", func() {
			tree, _ := toml.Load("version = '2'")
			convey.So(getConfigVersion(tree), convey.ShouldEqual, int64(2))
		})
		convey.Convey("version not set", func() {
			tree, _ := toml.Load("")
			convey.So(getConfigVersion(tree), convey.ShouldEqual, int64(2))
		})
	})
}

// TestGetCriRuntimePluginName tests the function getCriRuntimePluginName
func TestGetCriRuntimePluginName(t *testing.T) {
	convey.Convey("Test getCriRuntimePluginName", t, func() {
		convey.Convey("version 1", func() {
			convey.So(getCriRuntimePluginName(configVersion1), convey.ShouldEqual, version1RuntimePluginName)
		})
		convey.Convey("version 2", func() {
			convey.So(getCriRuntimePluginName(configVersion2), convey.ShouldEqual, version2RuntimePluginName)
		})
		convey.Convey("version 3", func() {
			convey.So(getCriRuntimePluginName(configVersion3), convey.ShouldEqual, version3RuntimePluginName)
		})
	})
}

// TestGetSubtreeByPath tests the function getSubtreeByPath
func TestGetSubtreeByPath(t *testing.T) {
	convey.Convey("Test getSubtreeByPath", t, func() {
		tree, _ := toml.Load(`
[plugins.cri.containerd.runtimes.runc]
runtime_type = "io.containerd.runc.v2"
`)
		convey.Convey("path exists", func() {
			keys := []string{pluginsKey, version1RuntimePluginName, containerdKey, runtimesKey, "runc"}
			subtree := getSubtreeByPath(keys, tree)
			convey.So(subtree, convey.ShouldNotBeNil)
		})
		convey.Convey("path not exists", func() {
			keys := []string{pluginsKey, version1RuntimePluginName, containerdKey, runtimesKey, "invalid"}
			subtree := getSubtreeByPath(keys, tree)
			convey.So(subtree, convey.ShouldBeNil)
		})
	})
}

// TestCopy tests the function copy
func TestCopy(t *testing.T) {
	convey.Convey("Test copy", t, func() {
		convey.Convey("tree is nil", func() {
			convey.So(copy(nil), convey.ShouldBeNil)
		})
		convey.Convey("tree is not nil", func() {
			tree, _ := toml.Load("key = 'value'")
			copyTree := copy(tree)
			convey.So(copyTree, convey.ShouldNotBeNil)
			convey.So(copyTree.Get("key"), convey.ShouldEqual, "value")
		})
	})
}

// TestGetDefaultRuntimeOptions tests the function getDefaultRuntimeOptions
func TestGetDefaultRuntimeOptions(t *testing.T) {
	convey.Convey("Test getDefaultRuntimeOptions", t, func() {
		convey.Convey("tree is nil", func() {
			options, _ := getDefaultRuntimeOptions(nil, version1RuntimePluginName)
			convey.So(options, convey.ShouldNotBeNil)
		})
		convey.Convey("runc options exist", func() {
			tree, _ := toml.Load(`
[plugins.cri.containerd.runtimes.runc]
runtime_type = "io.containerd.runc.v2"
`)
			options, _ := getDefaultRuntimeOptions(tree, version1RuntimePluginName)
			convey.So(options, convey.ShouldNotBeNil)
		})
		convey.Convey("runc options not exist", func() {
			tree, _ := toml.Load("")
			options, _ := getDefaultRuntimeOptions(tree, version1RuntimePluginName)
			convey.So(options, convey.ShouldNotBeNil)
		})
	})
}

// TestAddRuntime tests the function addRuntime
func TestAddRuntime(t *testing.T) {
	convey.Convey("Test addRuntime", t, func() {
		convey.Convey("tree is nil", func() {
			err := addRuntime(runtimeName, "/path/to/runtime", nil, version1RuntimePluginName)
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("add runtime successfully", func() {
			tree, _ := toml.Load(`
[plugins.cri.containerd.runtimes.runc]
runtime_type = "io.containerd.runc.v2"
`)
			err := addRuntime(runtimeName, "/path/to/runtime", tree, version1RuntimePluginName)
			convey.So(err, convey.ShouldBeNil)
			convey.So(tree.GetPath([]string{pluginsKey, version1RuntimePluginName, containerdKey,
				defaultRuntimeNameKey}),
				convey.ShouldEqual, runtimeName)
			// Check if runtime was added correctly
			runtimeConfig := tree.GetPath([]string{pluginsKey, version1RuntimePluginName, containerdKey,
				runtimesKey, runtimeName})
			convey.So(runtimeConfig, convey.ShouldNotBeNil)
			// Check if binary path was set correctly
			binaryPath := tree.GetPath([]string{pluginsKey, version1RuntimePluginName, containerdKey, runtimesKey,
				runtimeName, optionsKey, binaryNameKey})
			convey.So(binaryPath, convey.ShouldEqual, "/path/to/runtime")
		})
	})
}

// TestRemoveRuntime tests the function removeRuntime
func TestRemoveRuntime(t *testing.T) {
	convey.Convey("Test removeRuntime", t, func() {
		convey.Convey("tree is nil", func() {
			err := removeRuntime(runtimeName, nil, version1RuntimePluginName)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("remove runtime successfully", func() {
			tree, _ := toml.Load(`
[plugins.cri.containerd.runtimes.ascend]
runtime_type = "io.containerd.runc.v2"

[plugins.cri.containerd]
default_runtime_name = "ascend"
`)
			err := removeRuntime(runtimeName, tree, version1RuntimePluginName)
			convey.So(err, convey.ShouldBeNil)
			convey.So(tree.GetPath([]string{pluginsKey, version1RuntimePluginName, containerdKey, defaultRuntimeNameKey}),
				convey.ShouldEqual, "runc")
		})
	})
}

// TestWriteContainerdConfigToFile tests the function writeContainerdConfigToFile
func TestWriteContainerdConfigToFile(t *testing.T) {
	convey.Convey("Test writeContainerdConfigToFile", t, func() {
		convey.Convey("tree is nil", func() {
			err := writeContainerdConfigToFile(nil, "test.toml")
			convey.So(err, convey.ShouldBeError)
		})
	})
}

// TestEditContainerdConfig tests the function editContainerdConfig
func TestEditContainerdConfig(t *testing.T) {
	convey.Convey("Test editContainerdConfig", t, func() {
		convey.Convey("arg is nil", func() {
			err := editContainerdConfig(nil)
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("load file error", func() {
			patches := gomonkey.ApplyFuncReturn(toml.LoadFile, nil, testError)
			defer patches.Reset()
			err := editContainerdConfig(&commandArgs{srcFilePath: "test.toml"})
			convey.So(err, convey.ShouldBeError)
		})
	})
}

// TestContainerdProcess tests the function ContainerdProcess
func TestContainerdProcess(t *testing.T) {
	convey.Convey("Test ContainerdProcess", t, func() {
		convey.Convey("command is empty", func() {
			_, err := ContainerdProcess([]string{})
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("invalid param", func() {
			_, err := ContainerdProcess([]string{"invalid"})
			convey.So(err, convey.ShouldBeError)
		})
	})
}

// TestContainerdProcessFileCheck tests file check in ContainerdProcess
func TestContainerdProcessFileCheck(t *testing.T) {
	emptyStr := ""
	destFileTest := "aaa.txt.pid"

	convey.Convey("Test ContainerdProcess file check", t, func() {
		convey.Convey("file not exists and dir check fail", func() {
			patches := gomonkey.ApplyFuncReturn(os.Stat, nil, os.ErrNotExist).
				ApplyFuncReturn(mindxcheckutils.RealDirChecker, "", testError).
				ApplyFuncReturn(checkParamAndGetBehavior, true, "test")
			defer patches.Reset()
			cmds := []string{"test", oldJson, destFileTest, emptyStr, emptyStr}
			_, err := ContainerdProcess(cmds)
			convey.So(err, convey.ShouldBeError)
		})
	})
}

// TestContainerdProcessSuccess tests successful ContainerdProcess
func TestContainerdProcessSuccess(t *testing.T) {
	emptyStr := ""
	destFileTest := "aaa.txt.pid"

	convey.Convey("Test ContainerdProcess success", t, func() {
		// Create a valid toml tree for mocking
		tomlTree, _ := toml.Load("")

		patches := gomonkey.ApplyFuncReturn(os.Stat, &FileInfoMockContainerd{}, nil).
			ApplyFuncReturn(mindxcheckutils.RealFileChecker, "", nil).
			ApplyFuncReturn(mindxcheckutils.RealDirChecker, "", nil).
			ApplyFuncReturn(toml.LoadFile, tomlTree, nil).
			ApplyFuncReturn(writeContainerdConfigToFile, nil)
		defer patches.Reset()

		cmds := []string{"add", oldJson, destFileTest, "/path/to/runtime", emptyStr, emptyStr}
		ret, err := ContainerdProcess(cmds)
		convey.So(err, convey.ShouldBeNil)
		convey.So(ret, convey.ShouldEqual, "install")
	})
}

func initTestLog() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwlog.InitRunLogger(&hwLogConfig, context.Background())
}

// FileInfoMock is a mock for os.FileInfo
type FileInfoMockContainerd struct{}

func (f *FileInfoMockContainerd) Name() string       { return "test" }
func (f *FileInfoMockContainerd) Size() int64        { return 0 }
func (f *FileInfoMockContainerd) Mode() os.FileMode  { return 0 }
func (f *FileInfoMockContainerd) ModTime() time.Time { return time.Time{} }
func (f *FileInfoMockContainerd) IsDir() bool        { return false }
func (f *FileInfoMockContainerd) Sys() interface{}   { return nil }
