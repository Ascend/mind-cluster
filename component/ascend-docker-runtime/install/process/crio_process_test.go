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
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/pelletier/go-toml"
	"github.com/smartystreets/goconvey/convey"

	"ascend-docker-runtime/mindxcheckutils"
)

// TestAddCriORuntime tests the function addCriORuntime
func TestAddCriORuntime(t *testing.T) {
	convey.Convey("Test addCriORuntime", t, func() {
		convey.Convey("tree is nil, should return error", func() {
			err := addCriORuntime(runtimeName, "/path/to/runtime", nil)
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("empty tree: ascend uses default options, only path overridden", func() {
			tree, _ := toml.Load("")
			err := addCriORuntime(runtimeName, "/usr/local/Ascend/Ascend-Docker-Runtime/ascend-docker-runtime", tree)
			convey.So(err, convey.ShouldBeNil)
			convey.So(tree.GetPath([]string{crioSectionKey, crioRuntimeSectionKey, crioDefaultRuntimeKey}),
				convey.ShouldEqual, runtimeName)
			convey.So(tree.GetPath(crioRuntimePath(runtimeName, crioRuntimePathKey)),
				convey.ShouldEqual, "/usr/local/Ascend/Ascend-Docker-Runtime/ascend-docker-runtime")
			convey.So(tree.GetPath(crioRuntimePath(runtimeName, crioRuntimeTypeKey)), convey.ShouldEqual, crioOciRuntimeType)
			convey.So(tree.GetPath(crioRuntimePath(runtimeName, crioRuntimeRootKey)), convey.ShouldEqual, "")
		})
		convey.Convey("inherits runc runtime_type/runtime_root when runc entry present", func() {
			tree, _ := toml.Load(`
[crio.runtime.runtimes.runc]
runtime_type = "oci"
runtime_root = "/run/runc"
`)
			err := addCriORuntime(runtimeName, "/path/to/runtime", tree)
			convey.So(err, convey.ShouldBeNil)
			convey.So(tree.GetPath(crioRuntimePath(runtimeName, crioRuntimePathKey)),
				convey.ShouldEqual, "/path/to/runtime")
			convey.So(tree.GetPath(crioRuntimePath(runtimeName, crioRuntimeTypeKey)), convey.ShouldEqual, "oci")
			convey.So(tree.GetPath(crioRuntimePath(runtimeName, crioRuntimeRootKey)), convey.ShouldEqual, "/run/runc")
		})
		convey.Convey("re-add is idempotent", func() {
			tree, _ := toml.Load("")
			convey.So(addCriORuntime(runtimeName, "/path/a", tree), convey.ShouldBeNil)
			convey.So(addCriORuntime(runtimeName, "/path/b", tree), convey.ShouldBeNil)
			convey.So(tree.GetPath(crioRuntimePath(runtimeName, crioRuntimePathKey)), convey.ShouldEqual, "/path/b")
		})
	})
}

// TestRemoveCriORuntime tests the function removeCriORuntime
func TestRemoveCriORuntime(t *testing.T) {
	convey.Convey("Test removeCriORuntime", t, func() {
		convey.Convey("tree is nil, should return nil", func() {
			convey.So(removeCriORuntime(runtimeName, nil), convey.ShouldBeNil)
		})
		convey.Convey("remove runtime successfully", func() {
			tree, _ := toml.Load(`
[crio.runtime.runtimes.ascend]
runtime_type = "oci"
runtime_path = "/path/to/runtime"

[crio.runtime]
default_runtime = "ascend"
`)
			err := removeCriORuntime(runtimeName, tree)
			convey.So(err, convey.ShouldBeNil)
			convey.So(tree.GetPath([]string{crioSectionKey, crioRuntimeSectionKey, crioDefaultRuntimeKey}),
				convey.ShouldEqual, defaultRuntimeValue)
			convey.So(tree.GetPath(crioRuntimePath(runtimeName)), convey.ShouldBeNil)
		})
		convey.Convey("remove non-existent runtime is tolerated", func() {
			tree, _ := toml.Load("")
			err := removeCriORuntime(runtimeName, tree)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestEditCriOConfig tests the function editCriOConfig
func TestEditCriOConfig(t *testing.T) {
	convey.Convey("Test editCriOConfig", t, func() {
		convey.Convey("arg is nil, should return error", func() {
			err := editCriOConfig(nil)
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("load file error (not not-exist), should return error", func() {
			patches := gomonkey.ApplyFuncReturn(toml.LoadFile, nil, testError)
			defer patches.Reset()
			err := editCriOConfig(&commandArgs{srcFilePath: "test.toml", action: addCommand,
				runtimeFilePath: "/path/to/runtime"})
			convey.So(err, convey.ShouldBeError)
		})
	})
}

// TestCriOProcess tests the function CriOProcess
func TestCriOProcess(t *testing.T) {
	convey.Convey("Test CriOProcess", t, func() {
		convey.Convey("command is empty, should return error", func() {
			_, err := CriOProcess([]string{})
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("invalid param, should return error", func() {
			_, err := CriOProcess([]string{"invalid"})
			convey.So(err, convey.ShouldBeError)
		})
	})
}

// TestCriOProcessFileCheck tests file check in CriOProcess
func TestCriOProcessFileCheck(t *testing.T) {
	emptyStr := ""
	destFileTest := "aaa.txt.pid"

	convey.Convey("Test CriOProcess file check", t, func() {
		convey.Convey("file not exists and dir check fail", func() {
			patches := gomonkey.ApplyFuncReturn(os.Stat, nil, os.ErrNotExist).
				ApplyFuncReturn(mindxcheckutils.RealDirChecker, "", testError).
				ApplyFuncReturn(checkParamAndGetBehavior, true, "test")
			defer patches.Reset()
			cmds := []string{"test", emptyStr, destFileTest, emptyStr, emptyStr}
			_, err := CriOProcess(cmds)
			convey.So(err, convey.ShouldBeError)
		})
	})
}

// TestCriOProcessRm tests successful CriOProcess (rm / uninstall scene)
func TestCriOProcessRm(t *testing.T) {
	emptyStr := ""
	destFileTest := "aaa.txt.pid"

	convey.Convey("Test CriOProcess rm (uninstall)", t, func() {
		// Source drop-in already carries the ascend runtime entry.
		srcTree, _ := toml.Load(`
[crio.runtime.runtimes.ascend]
runtime_type = "oci"
runtime_path = "/path/to/runtime"

[crio.runtime]
default_runtime = "ascend"
`)
		patches := gomonkey.ApplyFuncReturn(os.Stat, &FileInfoMockCriO{}, nil).
			ApplyFuncReturn(mindxcheckutils.RealFileChecker, "", nil).
			ApplyFuncReturn(mindxcheckutils.RealDirChecker, "", nil).
			ApplyFuncReturn(toml.LoadFile, srcTree, nil).
			ApplyFuncReturn(writeTomlConfigToFile, nil)
		defer patches.Reset()

		cmds := []string{"rm", emptyStr, destFileTest, emptyStr, emptyStr}
		ret, err := CriOProcess(cmds)
		convey.So(err, convey.ShouldBeNil)
		convey.So(ret, convey.ShouldEqual, "uninstall")
	})
}

// TestCriOProcessSuccess tests successful CriOProcess (add scene)
func TestCriOProcessSuccess(t *testing.T) {
	emptyStr := ""
	destFileTest := "aaa.txt.pid"

	convey.Convey("Test CriOProcess success", t, func() {
		tomlTree, _ := toml.Load("")

		patches := gomonkey.ApplyFuncReturn(os.Stat, &FileInfoMockCriO{}, nil).
			ApplyFuncReturn(mindxcheckutils.RealFileChecker, "", nil).
			ApplyFuncReturn(mindxcheckutils.RealDirChecker, "", nil).
			ApplyFuncReturn(toml.LoadFile, tomlTree, nil).
			ApplyFuncReturn(writeTomlConfigToFile, nil)
		defer patches.Reset()

		cmds := []string{"add", emptyStr, destFileTest, "/path/to/runtime", emptyStr, emptyStr}
		ret, err := CriOProcess(cmds)
		convey.So(err, convey.ShouldBeNil)
		convey.So(ret, convey.ShouldEqual, "install")
	})
}

// TestWriteTomlConfigToFile tests the shared writer
func TestWriteTomlConfigToFile(t *testing.T) {
	convey.Convey("Test writeTomlConfigToFile", t, func() {
		convey.Convey("tree is nil, should return error", func() {
			err := writeTomlConfigToFile(nil, "test.toml")
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("write a real tree, content persisted", func() {
			tree, _ := toml.Load(`runtime_type = "oci"`)
			tmpFile := filepath.Join(os.TempDir(), "ascend-crio-ut.conf")
			defer os.Remove(tmpFile)
			convey.So(writeTomlConfigToFile(tree, tmpFile), convey.ShouldBeNil)
			data, rErr := os.ReadFile(tmpFile)
			convey.So(rErr, convey.ShouldBeNil)
			convey.So(string(data), convey.ShouldContainSubstring, "oci")
		})
	})
}

// FileInfoMockCriO is a mock for os.FileInfo
type FileInfoMockCriO struct{}

func (f *FileInfoMockCriO) Name() string       { return "test" }
func (f *FileInfoMockCriO) Size() int64        { return 0 }
func (f *FileInfoMockCriO) Mode() os.FileMode  { return 0 }
func (f *FileInfoMockCriO) ModTime() time.Time { return time.Time{} }
func (f *FileInfoMockCriO) IsDir() bool        { return false }
func (f *FileInfoMockCriO) Sys() interface{}   { return nil }
