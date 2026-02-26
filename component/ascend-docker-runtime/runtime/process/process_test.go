/* Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
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
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"syscall"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/containerd/containerd/oci"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-docker-runtime/mindxcheckutils"
	"ascend-docker-runtime/runtime/dcmi"
)

const (
	// strKubeDNSPort53UDPPort represents the string of the environment variable KUBE_DNS_PORT_53_UDP_PORT
	strKubeDNSPort53UDPPort = "KUBE_DNS_PORT_53_UDP_PORT=53"
	// strKubeDNSPort53UDPProto represents the string of the environment variable KUBE_DNS_PORT_53_UDP_PROTO
	strKubeDNSPort53UDPProto             = "KUBE_DNS_PORT_53_UDP_PROTO=udp"
	fileMode0400             os.FileMode = 0400
	fileMode0600             os.FileMode = 0600
	fileMode0655             os.FileMode = 0655
	needToMkdir                          = "./test"
	fileExistErrorStr                    = "file exists"
	bundleArgStr                         = "--bundle"
	execStubLog                          = "execute stub"
	configPath                           = "./test/config.json"
	chipName                             = "910"
	testStr                              = "test"
	writeAt                              = 2
)

var (
	deviceList = []int{1}
	testError  = errors.New("test")
)

func init() {
	ctx, _ := context.WithCancel(context.Background())
	logConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	if err := hwlog.InitRunLogger(&logConfig, ctx); err != nil {
		fmt.Printf("hwlog init failed, error is %v", err)
	}
}

// TestArgsIsCreate tests the function DoProcess
func TestArgsIsCreate(t *testing.T) {
	testArgs := []string{"create", bundleArgStr, "."}
	stub := gomonkey.ApplyGlobalVar(&os.Args, testArgs)
	defer stub.Reset()

	stub.ApplyFunc(execRunc, func() error {
		t.Log(execStubLog)
		return nil
	})

	err := DoProcess()
	assert.NotNil(t, err)
}

// TestArgsIsCreateCase1 tests the function DoProcess
func TestArgsIsCreateCase1(t *testing.T) {
	testArgs := []string{"create", bundleArgStr}
	stub := gomonkey.ApplyGlobalVar(&os.Args, testArgs)
	defer stub.Reset()

	stub.ApplyFunc(execRunc, func() error {
		t.Log(execStubLog)
		return nil
	})

	err := DoProcess()
	assert.NotNil(t, err)
}

// TestArgsIsCreateCase2 tests the function DoProcess
func TestArgsIsCreateCase2(t *testing.T) {
	testArgs := []string{"create", bundleArgStr, ""}
	stub := gomonkey.ApplyGlobalVar(&os.Args, testArgs)
	defer stub.Reset()

	stub.ApplyFunc(execRunc, func() error {
		t.Log(execStubLog)
		return nil
	})

	err := DoProcess()
	assert.NotNil(t, err)
}

// TestArgsIsCreateCase3 tests the function DoProcess
func TestArgsIsCreateCase3(t *testing.T) {
	if err := os.Mkdir(needToMkdir, fileMode0655); err != nil && !strings.Contains(err.Error(), fileExistErrorStr) {
		t.Fatalf("failed to create file, error: %v", err)
	}
	f, err := os.Create(configPath)
	defer f.Close()
	if err != nil {
		t.Logf("create file error: %v", err)
	}
	err = f.Chmod(fileMode0600)
	if err != nil {
		t.Logf("chmod file error: %v", err)
	}
	testArgs := []string{"create", bundleArgStr, needToMkdir}
	stub := gomonkey.ApplyGlobalVar(&os.Args, testArgs)
	defer stub.Reset()

	stub.ApplyFunc(execRunc, func() error {
		t.Log(execStubLog)
		return nil
	})
	err = InitLogModule(context.Background())
	assert.Nil(t, err)
	err = DoProcess()
	assert.NotNil(t, err)
}

// TestArgsIsCreateCase4 tests the function DoProcess
func TestArgsIsCreateCase4(t *testing.T) {
	if err := os.Mkdir(needToMkdir, fileMode0655); err != nil && !strings.Contains(err.Error(), fileExistErrorStr) {
		t.Fatalf("failed to create file, error: %v", err)
	}
	f, err := os.Create(configPath)
	defer f.Close()
	if err != nil {
		t.Logf("create file failed, error: %v", err)
	}
	err = f.Chmod(fileMode0600)
	if err != nil {
		t.Logf("chmod file failed, error: %v", err)
	}
	testArgs := []string{"spec", bundleArgStr, needToMkdir}
	stub := gomonkey.ApplyGlobalVar(&os.Args, testArgs)
	defer stub.Reset()

	stub.ApplyFunc(execRunc, func() error {
		t.Log(execStubLog)
		return nil
	})

	err = DoProcess()
	assert.Nil(t, err)
}

// TestDoProcess test the function DoProcess
func TestDoProcess(t *testing.T) {
	convey.Convey("test DoProcess", t, func() {
		testArgs := args{
			cmd:           "create",
			bundleDirPath: "",
		}
		patches := gomonkey.ApplyFuncReturn(getArgs, &testArgs, nil)
		defer patches.Reset()
		convey.Convey("01-Getwd error, should return error", func() {
			patch := gomonkey.ApplyFuncReturn(os.Getwd, testStr, testError)
			defer patch.Reset()
			err := DoProcess()
			convey.So(err, convey.ShouldBeError)
		})
		patches.ApplyFuncReturn(os.Getwd, testStr, nil)
		convey.Convey("02-success, should return nil", func() {
			patch := gomonkey.ApplyFuncReturn(modifySpecFile, nil).
				ApplyFuncReturn(execRunc, nil)
			defer patch.Reset()
			err := DoProcess()
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestModifySpecFile tests the function modifySpecFile
func TestModifySpecFile(t *testing.T) {
	err := modifySpecFile(configPath)
	assert.NotNil(t, err)
}

// TestModifySpecFileCase1 tests the function modifySpecFile
func TestModifySpecFileCase1(t *testing.T) {
	err := InitLogModule(context.Background())
	if err != nil {
		t.Logf("init log failed, error: %v", err)
	}
	if err := os.Mkdir(needToMkdir, fileMode0400); err != nil && !strings.Contains(err.Error(), fileExistErrorStr) {
		t.Logf("mkdir error: %v", err)
	}

	err = modifySpecFile(needToMkdir)
	assert.NotNil(t, err)
	if err := os.Remove(needToMkdir); err != nil {
		t.Logf("failed to remove dir, error: %v", err)
	}
}

// TestModifySpecFileCase2 tests the function modifySpecFile
func TestModifySpecFileCase2(t *testing.T) {
	file := "./test.json"
	f, err := os.Create(file)
	defer f.Close()
	if err != nil {
		t.Log("create file error")
	}
	err = f.Chmod(fileMode0600)
	if err != nil {
		t.Logf("chmod file error: %v", err)
	}
	if err := modifySpecFile(file); err != nil {
		t.Log("run modifySpecFile failed")
	}
	if err := os.Remove(file); err != nil {
		t.Logf("remove file(%v) failed, error: %v", file, err)
	}
}

// TestModifySpecFileCase3 tests the function modifySpecFile
func TestModifySpecFileCase3(t *testing.T) {
	file := "./test_spec.json"
	if err := modifySpecFile(file); err != nil {
		t.Log("run modifySpecFile failed")
	}
}

type mockFileInfo struct {
	os.FileInfo
}

func (m mockFileInfo) Mode() os.FileMode {
	return os.ModePerm
}

func (m mockFileInfo) IsDir() bool {
	return false
}

// TestModifySpecFilePatch1 tests the function modifySpecFile patch1
func TestModifySpecFilePatch1(t *testing.T) {
	convey.Convey("test modifySpecFile patch1", t, func() {
		convey.Convey("01-open file error, should return error", func() {
			patches := gomonkey.ApplyFuncReturn(os.Stat, mockFileInfo{}, nil).
				ApplyFuncReturn(mindxcheckutils.RealFileChecker, "", nil).
				ApplyFuncReturn(os.OpenFile, &os.File{}, testError)
			defer patches.Reset()
			err := modifySpecFile("")
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("02-read file error, should return error", func() {
			patches := gomonkey.ApplyFuncReturn(os.Stat, mockFileInfo{}, nil).
				ApplyFuncReturn(mindxcheckutils.RealFileChecker, "", nil).
				ApplyFuncReturn(os.OpenFile, &os.File{}, nil).
				ApplyFuncReturn(ioutil.ReadAll, []byte{}, testError)
			defer patches.Reset()
			err := modifySpecFile("")
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("03-json truncate error, should return error", func() {
			patches := gomonkey.ApplyFuncReturn(os.Stat, mockFileInfo{}, nil).
				ApplyFuncReturn(mindxcheckutils.RealFileChecker, "", nil).
				ApplyFuncReturn(os.OpenFile, &os.File{}, nil).
				ApplyFuncReturn(ioutil.ReadAll, []byte{}, nil).
				ApplyMethodReturn(&os.File{}, "Truncate", testError)
			defer patches.Reset()
			err := modifySpecFile("")
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("04-json seek error, should return error", func() {
			patches := gomonkey.ApplyFuncReturn(os.Stat, mockFileInfo{}, nil).
				ApplyFuncReturn(mindxcheckutils.RealFileChecker, "", nil).
				ApplyFuncReturn(os.OpenFile, &os.File{}, nil).
				ApplyFuncReturn(ioutil.ReadAll, []byte{}, nil).
				ApplyMethodReturn(&os.File{}, "Truncate", nil).
				ApplyMethodReturn(&os.File{}, "Seek", int64(0), testError)
			defer patches.Reset()
			err := modifySpecFile("")
			convey.So(err, convey.ShouldBeError)
		})
	})
}

// TestModifySpecFilePatch2 tests the function modifySpecFile patch2
func TestModifySpecFilePatch2(t *testing.T) {
	convey.Convey("test modifySpecFile patch2", t, func() {
		patches := gomonkey.ApplyFuncReturn(os.Stat, mockFileInfo{}, nil).
			ApplyFuncReturn(mindxcheckutils.RealFileChecker, "", nil).
			ApplyFuncReturn(os.OpenFile, &os.File{}, nil).
			ApplyFuncReturn(ioutil.ReadAll, []byte{}, nil).
			ApplyMethodReturn(&os.File{}, "Truncate", nil).
			ApplyMethodReturn(&os.File{}, "Seek", int64(0), nil).
			ApplyFuncReturn(json.Unmarshal, nil)
		defer patches.Reset()
		convey.Convey("05-check visible device error, should return error", func() {
			patch := gomonkey.ApplyFuncReturn(checkVisibleDevice, []int{}, testError)
			defer patch.Reset()
			err := modifySpecFile("")
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("06-fail to inject hook, should return error", func() {
			patch := gomonkey.ApplyFuncReturn(addHook, testError).
				ApplyFuncReturn(checkVisibleDevice, []int{0}, nil)
			defer patch.Reset()
			err := modifySpecFile("")
			convey.So(err, convey.ShouldBeError)
		})
		patches.ApplyFuncReturn(addHook, nil)
		convey.Convey("07-fail to add device, should return error", func() {
			patch := gomonkey.ApplyFuncReturn(addDevice, testError).
				ApplyFuncReturn(checkVisibleDevice, []int{0}, nil)
			defer patch.Reset()
			err := modifySpecFile("")
			convey.So(err, convey.ShouldBeError)
		})
		patches.ApplyFuncReturn(checkVisibleDevice, []int{}, nil).
			ApplyFunc(addAscendDockerEnv, func(spec *specs.Spec) {})
		convey.Convey("08-marshal error, should return error", func() {
			patch := gomonkey.ApplyFuncReturn(json.Marshal, []byte{}, testError)
			defer patch.Reset()
			err := modifySpecFile("")
			convey.So(err, convey.ShouldBeError)
		})
		patches.ApplyFuncReturn(json.Marshal, []byte{}, nil)
		convey.Convey("09-write error, should return error", func() {
			patch := gomonkey.ApplyMethodReturn(&os.File{}, "WriteAt", 0, testError)
			defer patch.Reset()
			err := modifySpecFile("")
			convey.So(err, convey.ShouldBeError)
		})
	})
}

// TestModifySpecFilePatch3 tests the function modifySpecFile patch3
func TestModifySpecFilePatch3(t *testing.T) {
	convey.Convey("test modifySpecFile patch3", t, func() {
		patches := gomonkey.ApplyFuncReturn(os.Stat, mockFileInfo{}, nil).
			ApplyFuncReturn(mindxcheckutils.RealFileChecker, "", nil).
			ApplyFuncReturn(os.OpenFile, &os.File{}, nil).
			ApplyFuncReturn(ioutil.ReadAll, []byte{}, nil).
			ApplyMethodReturn(&os.File{}, "Truncate", nil).
			ApplyMethodReturn(&os.File{}, "Seek", int64(0), nil).
			ApplyFuncReturn(json.Unmarshal, nil).
			ApplyFuncReturn(addHook, nil).
			ApplyFuncReturn(checkVisibleDevice, []int{}, nil).
			ApplyFunc(addAscendDockerEnv, func(spec *specs.Spec) {}).
			ApplyFuncReturn(json.Marshal, []byte{}, nil).
			ApplyMethodReturn(&os.File{}, "WriteAt", 0, nil)
		defer patches.Reset()
		convey.Convey("10-success, should return nil", func() {
			err := modifySpecFile("")
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestModifySpecFilePatch4 tests the function modifySpecFile
func TestModifySpecFilePatch4(t *testing.T) {
	convey.Convey("test modifySpecFile patch3", t, func() {
		patches := gomonkey.ApplyFuncReturn(os.Stat, mockFileInfo{}, nil).
			ApplyFuncReturn(mindxcheckutils.RealFileChecker, "", nil).
			ApplyFuncReturn(os.OpenFile, &os.File{}, nil).
			ApplyFuncReturn(ioutil.ReadAll, []byte{}, nil).
			ApplyMethodReturn(&os.File{}, "Truncate", nil).
			ApplyMethodReturn(&os.File{}, "Seek", int64(0), nil)
		defer patches.Reset()
		convey.Convey("11-unmarshal error, should return error", func() {
			patch := gomonkey.ApplyFuncReturn(json.Unmarshal, testError)
			defer patch.Reset()
			err := modifySpecFile("")
			convey.So(err, convey.ShouldBeError)
		})
	})
}

func TestReadSpecFile(t *testing.T) {
	convey.Convey("test ReadSpecFile", t, func() {
		convey.Convey("01-get file stat error, should return error", func() {
			patch := gomonkey.ApplyFuncReturn(os.Stat, nil, testError)
			defer patch.Reset()
			_, err := readSpecFile("")
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("02-open file error, should return error", func() {
			patch := gomonkey.ApplyFuncReturn(os.OpenFile, nil, testError)
			defer patch.Reset()
			_, err := readSpecFile("/valid/path")
			convey.So(err, convey.ShouldBeError)
		})
		patch1 := gomonkey.ApplyFuncReturn(os.OpenFile, &os.File{}, nil)
		defer patch1.Reset()
		convey.Convey("03-check file error, should return error", func() {
			patch := patch1.ApplyFuncReturn(mindxcheckutils.CheckFileInfo, testError)
			defer patch.Reset()
			_, err := readSpecFile("/valid/path")
			convey.So(err, convey.ShouldBeError)
		})
		patch2 := patch1.ApplyFuncReturn(mindxcheckutils.CheckFileInfo, nil)
		defer patch2.Reset()
		convey.Convey("04-read file error, should return error", func() {
			patch := patch2.ApplyFuncReturn(ioutil.ReadAll, nil, testError)
			defer patch.Reset()
			_, err := readSpecFile("/valid/path")
			convey.So(err, convey.ShouldBeError)
		})
		patch3 := patch2.ApplyFuncReturn(ioutil.ReadAll, []byte("invalid json"), nil)
		defer patch3.Reset()
		convey.Convey("05-unmarshal error, should return error", func() {
			patch := patch3.ApplyFuncReturn(json.Unmarshal, testError)
			defer patch.Reset()
			_, err := readSpecFile("/valid/path")
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("06-success case", func() {
			patch := patch3.ApplyFuncReturn(json.Unmarshal, nil)
			defer patch.Reset()
			_, err := readSpecFile("/valid/path")
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestProcessDevicesAndHooks(t *testing.T) {
	convey.Convey("test processDevicesAndHooks", t, func() {
		convey.Convey("02-spec without ASCEND_VISIBLE_DEVICES, should return error", func() {
			spec := &specs.Spec{}
			err := processDevicesAndHooks(spec)
			convey.So(err, convey.ShouldBeError)
		})

		convey.Convey("02-spec with  ASCEND_VISIBLE_DEVICES, should return nil", func() {
			spec := &specs.Spec{
				Process: &specs.Process{
					Env: []string{"ASCEND_VISIBLE_DEVICES=0"},
				},
			}
			err := processDevicesAndHooks(spec)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestWriteSpecFile tests the function writeSpecFile
func TestWriteSpecFile(t *testing.T) {
	convey.Convey("test writeSpecFile", t, func() {
		testSpec := &specs.Spec{Version: "1.0.2"}
		path := "test_spec.json"
		defer os.Remove(path)
		convey.Convey("01-open file error, should return error", func() {
			err := writeSpecFile("/invalid/path", &specs.Spec{})
			convey.So(err, convey.ShouldBeError)
		})
		patch := gomonkey.ApplyFuncReturn(os.OpenFile, &os.File{}, nil)
		defer patch.Reset()
		convey.Convey("02-check file info error, should return error", func() {
			patch := gomonkey.ApplyFuncReturn(mindxcheckutils.CheckFileInfo, testError)
			defer patch.Reset()
			err := writeSpecFile(path, testSpec)
			convey.So(err, convey.ShouldBeError)
		})
		patch1 := gomonkey.ApplyFuncReturn(mindxcheckutils.CheckFileInfo, nil)
		defer patch1.Reset()
		convey.Convey("03-marshal error, should return error", func() {
			patch := gomonkey.ApplyFuncReturn(json.Marshal, []byte{}, testError)
			defer patch.Reset()
			err := writeSpecFile(path, testSpec)
			convey.So(err, convey.ShouldBeError)
		})
		patch2 := gomonkey.ApplyFuncReturn(json.Marshal, []byte("{}"), nil)
		defer patch2.Reset()
		convey.Convey("04-truncate error, should return error", func() {
			patch := gomonkey.ApplyMethodReturn(new(os.File), "Truncate", testError)
			defer patch.Reset()
			err := writeSpecFile(path, testSpec)
			convey.So(err, convey.ShouldBeError)
		})
		patch3 := gomonkey.ApplyMethodReturn(new(os.File), "Truncate", nil)
		defer patch3.Reset()
		convey.Convey("05-write error, should return error", func() {
			patch := gomonkey.ApplyMethodReturn(new(os.File), "WriteAt", 0, testError)
			defer patch.Reset()
			err := writeSpecFile(path, testSpec)
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("06-success, should return nil", func() {
			patch := gomonkey.ApplyMethodReturn(new(os.File), "WriteAt", writeAt, nil)
			defer patch.Reset()
			err := writeSpecFile(path, testSpec)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestAddHookCase1(t *testing.T) {
	var specArgs = &specs.Spec{}
	stub := gomonkey.ApplyGlobalVar(&hookCliPath, ".")
	defer stub.Reset()

	err := addHook(&dcmi.NpuV1Worker{DcMgr: &dcmi.DcV1Manager{}}, specArgs, &deviceList)
	assert.NotNil(t, err)
}

// TestAddHookCase2 tests the function addHook
func TestAddHookCase2(t *testing.T) {
	var specArgs = &specs.Spec{}
	stub := gomonkey.ApplyGlobalVar(&hookCliPath, ".")
	defer stub.Reset()
	stub.ApplyGlobalVar(&hookDefaultFile, ".")
	err := addHook(&dcmi.NpuV1Worker{DcMgr: &dcmi.DcV1Manager{}}, specArgs, &deviceList)
	assert.NotNil(t, err)
}

// TestAddHookCase3 tests the function addHook
func TestAddHookCase3(t *testing.T) {
	file := "/usr/local/bin/ascend-docker-hook"
	filenew := "/usr/local/bin/ascend-docker-hook-1"

	if err := os.Rename(file, filenew); err != nil {
		t.Log("rename ", file)
	}
	var specArgs = &specs.Spec{}
	err := addHook(&dcmi.NpuV1Worker{DcMgr: &dcmi.DcV1Manager{}}, specArgs, &deviceList)
	assert.NotNil(t, err)

	if err := os.Rename(filenew, file); err != nil {
		t.Log("rename ", file)
	}
}

// TestExecRunc tests the function execRunc
func TestExecRunc(t *testing.T) {
	stub := gomonkey.ApplyGlobalVar(&dockerRuncName, "abc-runc")
	stub.ApplyGlobalVar(&runcName, "runc123")
	defer stub.Reset()

	err := execRunc()
	assert.NotNil(t, err)
}

// TestExecRuncPatch1 tests the function execRunc
func TestExecRuncPatch1(t *testing.T) {
	convey.Convey("test execRunc patch1", t, func() {
		patches := gomonkey.ApplyFuncReturn(exec.LookPath, testStr, nil)
		defer patches.Reset()
		convey.Convey("01-EvalSymlinks error, should return error", func() {
			patch := gomonkey.ApplyFuncReturn(filepath.EvalSymlinks, testStr, testError)
			defer patch.Reset()
			err := execRunc()
			convey.So(err, convey.ShouldBeError)
		})
		patches.ApplyFuncReturn(filepath.EvalSymlinks, testStr, nil)
		convey.Convey("02-RealFileChecker error, should return error", func() {
			patch := gomonkey.ApplyFuncReturn(mindxcheckutils.RealFileChecker, testStr, testError)
			defer patch.Reset()
			err := execRunc()
			convey.So(err, convey.ShouldBeError)
		})
		patches.ApplyFuncReturn(mindxcheckutils.RealFileChecker, testStr, nil)
		convey.Convey("03-ChangeRuntimeLogMode error, should return error", func() {
			patch := gomonkey.ApplyFuncReturn(mindxcheckutils.ChangeRuntimeLogMode, testError)
			defer patch.Reset()
			convey.So(execRunc(), convey.ShouldBeError)
		})
		patches.ApplyFuncReturn(mindxcheckutils.ChangeRuntimeLogMode, nil)
		convey.Convey("04-syscall Exec error, should return error", func() {
			patch := gomonkey.ApplyFuncReturn(syscall.Exec, testError)
			defer patch.Reset()
			convey.So(execRunc(), convey.ShouldBeError)
		})
		convey.Convey("05-success, should return nil", func() {
			patch := gomonkey.ApplyFuncReturn(syscall.Exec, nil)
			defer patch.Reset()
			convey.So(execRunc(), convey.ShouldBeNil)
		})
	})
}

// TestParseDevicesCase1 tests the function parseDevices
func TestParseDevicesCase1(t *testing.T) {
	visibleDevices := "0-3,5,7"
	expectVal := []int{0, 1, 2, 3, 5, 7}
	actualVal, err := parseDevices(visibleDevices)
	if err != nil || !reflect.DeepEqual(expectVal, actualVal) {
		t.Fail()
	}
}

// TestParseDevicesCase2 tests the function parseDevices
func TestParseDevicesCase2(t *testing.T) {
	visibleDevices := "0-3-4,5,7"
	_, err := parseDevices(visibleDevices)
	assert.NotNil(t, err)
}

// TestParseDevicesCase3 tests the function parseDevices
func TestParseDevicesCase3(t *testing.T) {
	visibleDevices := "0l-3,5,7"
	_, err := parseDevices(visibleDevices)
	assert.NotNil(t, err)
}

// TestParseDevicesCase4 tests the function parseDevices
func TestParseDevicesCase4(t *testing.T) {
	visibleDevices := "0-3o,5,7"
	_, err := parseDevices(visibleDevices)
	assert.NotNil(t, err)
}

// TestParseDevicesCase5 tests the function parseDevices
func TestParseDevicesCase5(t *testing.T) {
	visibleDevices := "4-3,5,7"
	_, err := parseDevices(visibleDevices)
	assert.NotNil(t, err)
}

// TestParseDevicesCase6 tests the function parseDevices
func TestParseDevicesCase6(t *testing.T) {
	visibleDevices := "3o,5,7"
	_, err := parseDevices(visibleDevices)
	assert.NotNil(t, err)
}

// TestParseDevicesCase7 tests the function parseDevices
func TestParseDevicesCase7(t *testing.T) {
	visibleDevices := "0=3,5,7"
	_, err := parseDevices(visibleDevices)
	assert.NotNil(t, err)
}

// TestRemoveDuplication tests the function removeDuplication
func TestRemoveDuplication(t *testing.T) {
	originList := []int{1, 2, 2, 4, 5, 5, 5, 6, 8, 8}
	targetList := []int{1, 2, 4, 5, 6, 8}
	resultList := removeDuplication(originList)

	assert.EqualValues(t, targetList, resultList)
}

// TestAddEnvToDevicePlugin0 tests the function addAscendDockerEnv
func TestAddEnvToDevicePlugin0(t *testing.T) {
	devicePluginHostName := "pf2i6r"
	spec := specs.Spec{
		Process: &specs.Process{
			Env: []string{strKubeDNSPort53UDPPort,
				fmt.Sprintf("HOSTNAME=%s", devicePluginHostName),
				strKubeDNSPort53UDPProto},
		},
	}

	addAscendDockerEnv(&spec)
	assert.Contains(t, spec.Process.Env, useAscendDocker)
}

// TestAddEnvToDevicePlugin1 tests the function addAscendDockerEnv
func TestAddEnvToDevicePlugin1(t *testing.T) {
	devicePluginHostName := "pf2i6r"
	spec := specs.Spec{
		Process: &specs.Process{
			Env: []string{strKubeDNSPort53UDPPort,
				fmt.Sprintf("HOSTNAME=%s", devicePluginHostName),
				strKubeDNSPort53UDPProto},
		},
	}

	addAscendDockerEnv(&spec)
	assert.Contains(t, spec.Process.Env, useAscendDocker)
}

// TestAddEnvToDevicePlugin2 tests the function addAscendDockerEnv
func TestAddEnvToDevicePlugin2(t *testing.T) {
	convey.Convey("01-spec empty, should return immediately", t, func() {
		spec := &specs.Spec{}
		addAscendDockerEnv(spec)
		convey.So(spec.Process, convey.ShouldBeNil)
	})
}

// TestGetDeviceTypeByChipName0 tests the function GetDeviceTypeByChipName
func TestGetDeviceTypeByChipName0(t *testing.T) {
	chipName := "310B"
	devType := GetDeviceTypeByChipName(chipName)
	assert.EqualValues(t, Ascend310B, devType)
}

// TestGetDeviceTypeByChipName1 tests the function GetDeviceTypeByChipName
func TestGetDeviceTypeByChipName1(t *testing.T) {
	chipName := "310P"
	devType := GetDeviceTypeByChipName(chipName)
	assert.EqualValues(t, Ascend310P, devType)
}

// TestGetDeviceTypeByChipName2 tests the function GetDeviceTypeByChipName
func TestGetDeviceTypeByChipName2(t *testing.T) {
	chipName := "310"
	devType := GetDeviceTypeByChipName(chipName)
	assert.EqualValues(t, Ascend310, devType)
}

// TestGetDeviceTypeByChipName3 tests the function GetDeviceTypeByChipName
func TestGetDeviceTypeByChipName3(t *testing.T) {
	devType := GetDeviceTypeByChipName(chipName)
	assert.EqualValues(t, Ascend910, devType)
}

// TestGetDeviceTypeByChipName4 tests the function GetDeviceTypeByChipName
func TestGetDeviceTypeByChipName4(t *testing.T) {
	chipName := "980b"
	devType := GetDeviceTypeByChipName(chipName)
	assert.EqualValues(t, "", devType)
}

// TestGetValueByKeyCase1 tests the function getValueByKey
func TestGetValueByKeyCase1(t *testing.T) {
	data := []string{"ASCEND_VISIBLE_DEVICES=0-3,5,7"}
	word := "ASCEND_VISIBLE_DEVICES"
	expectVal := "0-3,5,7"
	actualVal := getValueByKey(data, word)
	assert.EqualValues(t, expectVal, actualVal)
}

// TestGetValueByKeyCase2 tests the function getValueByKey
func TestGetValueByKeyCase2(t *testing.T) {
	data := []string{"ASCEND_VISIBLE_DEVICES"}
	word := "ASCEND_VISIBLE_DEVICES"
	expectVal := ""
	defer func() {
		if err := recover(); err != nil {
			t.Log("exception occur")
		}
	}()
	actualVal := getValueByKey(data, word)
	assert.EqualValues(t, expectVal, actualVal)
}

// TestGetValueByKeyCase3 tests the function getValueByKey
func TestGetValueByKeyCase3(t *testing.T) {
	data := []string{"ASCEND_VISIBLE_DEVICES=0-3,5,7"}
	word := "ASCEND_VISIBLE_DEVICE"
	expectVal := ""
	actualVal := getValueByKey(data, word)
	assert.EqualValues(t, expectVal, actualVal)
}

// TestUpdateEnvAndPostHook tests the function updateEnvAndPostHook
func TestUpdateEnvAndPostHook(t *testing.T) {
	defer func() {
		if e := recover(); e != nil {
			t.Logf("%s", e)
		}
	}()
	vDeviceId := int32(100)
	vdvice := dcmi.VDeviceInfo{
		CardID:    0,
		DeviceID:  0,
		VdeviceID: vDeviceId,
	}

	spec := specs.Spec{
		Process: &specs.Process{
			Env: []string{strKubeDNSPort53UDPPort,
				fmt.Sprintf("%s=0", ascendVisibleDevices),
				strKubeDNSPort53UDPProto},
		},
		Hooks: &specs.Hooks{},
	}
	convey.Convey("test updateEnvAndPostHook patch2", t, func() {
		convey.Convey("get executable path failed, should return error", func() {
			patch := gomonkey.ApplyFuncReturn(os.Executable, "", errors.New("executable failed"))
			defer patch.Reset()
			err := updateEnvAndPostHook(&spec, vdvice, &deviceList)
			convey.ShouldContain(err.Error(), "cannot get the path of docker-destroy:")
		})
		convey.Convey("deviceIdList is nil, should return error", func() {
			patch := gomonkey.ApplyFuncReturn(mindxcheckutils.RealFileChecker, "", errors.New("check failed")).
				ApplyFuncReturn(os.Executable, "", nil)
			defer patch.Reset()
			err := updateEnvAndPostHook(&spec, vdvice, &deviceList)
			convey.ShouldContain(err.Error(), "failed to check docker-destroy executable file at")
		})
		convey.Convey("updateEnvAndPostHook success, should return nil", func() {
			patch := gomonkey.ApplyFuncReturn(mindxcheckutils.RealFileChecker, "", nil).
				ApplyFuncReturn(os.Executable, "", nil)
			defer patch.Reset()
			err := updateEnvAndPostHook(&spec, vdvice, &deviceList)
			assert.Nil(t, err)
			assert.Contains(t, spec.Process.Env, "ASCEND_VISIBLE_DEVICES=0")
			assert.Contains(t, spec.Process.Env, "ASCEND_RUNTIME_OPTIONS=VIRTUAL")
			assert.Contains(t, spec.Hooks.Poststop[0].Path, destroyHookCli)
		})
	})
}

// TestUpdateEnvAndPostHookPatch1 tests the function updateEnvAndPostHook
func TestUpdateEnvAndPostHookPatch1(t *testing.T) {
	convey.Convey("test updateEnvAndPostHook patch1", t, func() {
		convey.Convey("01-deviceIdList is nil, should return nil", func() {
			spec := &specs.Spec{}
			err := updateEnvAndPostHook(spec, dcmi.VDeviceInfo{}, nil)
			convey.ShouldBeNil(err)
			convey.So(spec.Process, convey.ShouldBeNil)
		})
	})
}

// TestAddDeviceToSpec0 tests the function addDeviceToSpec
func TestAddDeviceToSpec0(t *testing.T) {
	devPath := "/dev/davinci0"
	statStub := gomonkey.ApplyFunc(oci.DeviceFromPath, func(name string) (*specs.LinuxDevice, error) {
		return &specs.LinuxDevice{
			Path: devPath,
		}, nil
	})
	defer statStub.Reset()

	spec := specs.Spec{
		Linux: &specs.Linux{
			Devices: []specs.LinuxDevice{},
			Resources: &specs.LinuxResources{
				Devices: []specs.LinuxDeviceCgroup{},
			},
		},
	}

	err := addDeviceToSpec(&spec, devPath, devPath)
	assert.Nil(t, err)
	assert.Contains(t, spec.Linux.Devices[0].Path, devPath)
}

// TestAddDeviceToSpecPatch1 tests the function addDeviceToSpec
func TestAddDeviceToSpecPatch1(t *testing.T) {
	convey.Convey("test addDeviceToSpec patch1", t, func() {
		convey.Convey("01-DeviceFromPath error, should return error", func() {
			patch := gomonkey.ApplyFuncReturn(oci.DeviceFromPath, nil, testError)
			defer patch.Reset()
			convey.So(addDeviceToSpec(nil, "", ""), convey.ShouldBeError)
		})
		testSp := specs.Spec{
			Linux: &specs.Linux{
				Devices: make([]specs.LinuxDevice, 0),
				Resources: &specs.LinuxResources{
					Devices: make([]specs.LinuxDeviceCgroup, 0),
				},
			},
		}
		patches := gomonkey.ApplyFuncReturn(oci.DeviceFromPath, &specs.LinuxDevice{
			Type:  testStr,
			Major: 0,
			Minor: 0,
		}, nil)
		defer patches.Reset()
		convey.Convey("02-virtualDavinciName success, should return nil", func() {
			dContainerPath, err := getMountPath("0", virtualDavinciName)
			convey.So(err, convey.ShouldBeNil)
			convey.So(addDeviceToSpec(&testSp, "0", dContainerPath), convey.ShouldBeNil)
		})
		convey.Convey("03-davinciManagerDocker success, should return nil", func() {
			dContainerPath, err := getMountPath("0", davinciManagerDocker)
			convey.So(err, convey.ShouldBeNil)
			convey.So(addDeviceToSpec(&testSp, "", dContainerPath), convey.ShouldBeNil)
		})
		convey.Convey("04-notRenamePath success, should return nil", func() {
			dPath := devicePath + dvppCmdList
			convey.So(addDeviceToSpec(&testSp, dPath, dPath), convey.ShouldBeNil)
		})
	})
}

// TestAddAscend310BManagerDevice tests the function addAscend310BManagerDevice
func TestAddAscend310BManagerDevice(t *testing.T) {
	statStub := gomonkey.ApplyFunc(addDeviceToSpec, func(spec *specs.Spec, dHostPath string,
		dContainerPath string) error {
		return nil
	})
	defer statStub.Reset()

	pathStub := gomonkey.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) {
		return nil, nil
	})
	defer pathStub.Reset()

	spec := specs.Spec{
		Linux: &specs.Linux{
			Devices: []specs.LinuxDevice{},
			Resources: &specs.LinuxResources{
				Devices: []specs.LinuxDeviceCgroup{},
			},
		},
	}

	err := addAscend310BManagerDevice(&spec)
	assert.Nil(t, err)
}

// TestGetMountPath tests the function getMountPath
func TestGetMountPath(t *testing.T) {
	convey.Convey("test getMountPath", t, func() {
		convey.Convey("test virtualDavinciName device type success", func() {
			testPath := "0"
			expectPath := devicePath + davinciName + testPath
			dContainerPath, err := getMountPath(testPath, virtualDavinciName)
			convey.So(err, convey.ShouldBeNil)
			convey.So(dContainerPath, convey.ShouldEqual, expectPath)
		})
		convey.Convey("test virtualDavinciName device type error", func() {
			testPath := "0a0"
			dContainerPath, err := getMountPath(testPath, virtualDavinciName)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(dContainerPath, convey.ShouldEqual, "")
		})
		convey.Convey("test davinciManagerDocker device type success", func() {
			testPath := ""
			expectPath := devicePath + davinciManager
			dContainerPath, err := getMountPath(testPath, davinciManagerDocker)
			convey.So(err, convey.ShouldBeNil)
			convey.So(dContainerPath, convey.ShouldEqual, expectPath)
		})
		convey.Convey("test default device type success", func() {
			testPath := ""
			expectPath := testPath
			dContainerPath, err := getMountPath(testPath, "")
			convey.So(err, convey.ShouldBeNil)
			convey.So(dContainerPath, convey.ShouldEqual, expectPath)
		})
	})
}

// TestAddAscend310BManagerDevicePatch1 tests the function addAscend310BManagerDevice
func TestAddAscend310BManagerDevicePatch1(t *testing.T) {
	convey.Convey("test addAscend310BManagerDevice patch1", t, func() {
		convey.Convey("01-addDeviceToSpec error, should return error", func() {
			patch := gomonkey.ApplyFuncReturn(addDeviceToSpec, testError)
			defer patch.Reset()
			convey.So(addAscend310BManagerDevice(nil), convey.ShouldBeError)
		})
		patches := gomonkey.ApplyFuncReturn(addDeviceToSpec, nil)
		defer patches.Reset()
		convey.Convey("02-Stat error, should return error", func() {
			patch := gomonkey.ApplyFuncReturn(os.Stat, mockFileInfo{}, testError)
			defer patch.Reset()
			convey.So(addAscend310BManagerDevice(nil), convey.ShouldBeError)
		})
	})
}

// TestAddCommonManagerDevice tests the function addCommonManagerDevice
func TestAddCommonManagerDevice(t *testing.T) {
	statStub := gomonkey.ApplyFunc(addDeviceToSpec, func(spec *specs.Spec, dHostPath string,
		dContainerPath string) error {
		return nil
	})
	defer statStub.Reset()

	tests := []struct {
		name        string
		deviceType  string
		expectError bool
	}{
		{
			name:        "Ascend910A2（910B）",
			deviceType:  api.Ascend910B,
			expectError: false,
		},
		{
			name:        "Ascend910A3",
			deviceType:  api.Ascend910A3,
			expectError: false,
		},
		{
			name:        "Ascend910A5",
			deviceType:  api.Ascend910A5,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := specs.Spec{
				Linux: &specs.Linux{
					Devices: []specs.LinuxDevice{},
					Resources: &specs.LinuxResources{
						Devices: []specs.LinuxDeviceCgroup{},
					},
				},
			}

			err := addCommonManagerDevice(&spec, tt.deviceType)
			if tt.expectError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

// TestAddCommonManagerDevicePatch1 tests the function addCommonManagerDevice
func TestAddCommonManagerDevicePatch1(t *testing.T) {
	convey.Convey("test addCommonManagerDevice patch1", t, func() {
		convey.Convey("01-addDeviceToSpec error, should return error", func() {
			patch := gomonkey.ApplyFunc(addDeviceToSpec, func(spec *specs.Spec, dHostPath, dContainerPath string) error {
				return testError
			})
			defer patch.Reset()

			spec := specs.Spec{
				Linux: &specs.Linux{
					Devices: []specs.LinuxDevice{},
					Resources: &specs.LinuxResources{
						Devices: []specs.LinuxDeviceCgroup{},
					},
				},
			}
			err := addCommonManagerDevice(&spec, Ascend910)
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("02-Ascend910A5 skip dvpp_cmdlist, should return nil", func() {
			callCount := 0
			patch := gomonkey.ApplyFunc(addDeviceToSpec, func(spec *specs.Spec, dHostPath, dContainerPath string) error {
				if strings.Contains(dContainerPath, "dvpp_cmdlist") {
					callCount++
				}
				return nil
			})
			defer patch.Reset()

			spec := specs.Spec{
				Linux: &specs.Linux{
					Devices: []specs.LinuxDevice{},
					Resources: &specs.LinuxResources{
						Devices: []specs.LinuxDeviceCgroup{},
					},
				},
			}
			err := addCommonManagerDevice(&spec, Ascend910A5)
			convey.So(err, convey.ShouldBeNil)
			convey.So(callCount, convey.ShouldEqual, 0) // Verify that dvpp_cmdlist was not added.
		})
	})
}

// TestAddManagerDevice tests the function addManagerDevice
func TestAddManagerDevice(t *testing.T) {
	devPath := "/dev/mockdevice"
	statStub := gomonkey.ApplyFunc(oci.DeviceFromPath, func(dPath string) (*specs.LinuxDevice, error) {
		return &specs.LinuxDevice{
			Path: devPath,
		}, nil
	})
	defer statStub.Reset()

	patchGetChipName := gomonkey.ApplyMethod(reflect.TypeOf(&dcmi.NpuV1Worker{}), "GetChipName", func(f *dcmi.NpuV1Worker) (string, error) {
		return chipName, nil
	})
	defer patchGetChipName.Reset()

	patchGetProductType := gomonkey.ApplyMethod(reflect.TypeOf(&dcmi.NpuV1Worker{}), "GetProductType", func(f *dcmi.NpuV1Worker) (string, error) {
		return "", nil
	})
	defer patchGetProductType.Reset()

	spec := specs.Spec{
		Linux: &specs.Linux{
			Devices: []specs.LinuxDevice{},
			Resources: &specs.LinuxResources{
				Devices: []specs.LinuxDeviceCgroup{},
			},
		},
	}
	ctx, _ := context.WithCancel(context.Background())
	err := InitLogModule(ctx)
	assert.Nil(t, err)
	err = addManagerDevice(&dcmi.NpuV1Worker{DcMgr: &dcmi.DcV1Manager{}}, &spec)
	assert.Nil(t, err)
}

// TestAddUBDevice tests the function addUBDevice
func TestAddUBDevice(t *testing.T) {
	specInstance := specs.Spec{
		Linux: &specs.Linux{
			Devices: []specs.LinuxDevice{},
			Resources: &specs.LinuxResources{
				Devices: []specs.LinuxDeviceCgroup{},
			},
		},
		Process: &specs.Process{
			Env: []string{},
		},
	}

	convey.Convey("test addUBDevice", t, func() {
		convey.Convey("test addUBDevice should add nothing when ub directory not exists", func() {
			mockStat := gomonkey.ApplyFunc(os.Stat, func(_ string) (fs.FileInfo, error) {
				return nil, os.ErrNotExist
			})
			defer mockStat.Reset()

			err := addUBDevice(&specInstance)
			convey.So(err, convey.ShouldBeNil)
			convey.So(specInstance.Linux.Devices, convey.ShouldBeEmpty)
		})

		convey.Convey("test addUBDevice should add ub devices when ub directory exists", func() {
			mockStat := gomonkey.ApplyFunc(os.Stat, func(_ string) (fs.FileInfo, error) {
				return mockFileInfo{}, nil
			})
			defer mockStat.Reset()
			mockAddDevicesInDir := gomonkey.ApplyFunc(addDevicesInDir, func(spec *specs.Spec, dirPath string) error {
				spec.Linux.Devices = append(spec.Linux.Devices, specs.LinuxDevice{
					Path: "UBDevicePath",
				})
				return nil
			})
			defer mockAddDevicesInDir.Reset()

			err := addUBDevice(&specInstance)
			convey.So(err, convey.ShouldBeNil)
			convey.So(specInstance.Linux.Devices, convey.ShouldNotBeEmpty)
		})
	})
}

// TestAddUBDeviceWithError tests the function addUBDevice
func TestAddUBDeviceWithError(t *testing.T) {
	specInstance := specs.Spec{
		Linux: &specs.Linux{
			Devices: []specs.LinuxDevice{},
			Resources: &specs.LinuxResources{
				Devices: []specs.LinuxDeviceCgroup{},
			},
		},
		Process: &specs.Process{
			Env: []string{},
		},
	}

	convey.Convey("test addUBDevice should return error when addDevicesInDir return error", t, func() {
		mockStat := gomonkey.ApplyFunc(os.Stat, func(_ string) (fs.FileInfo, error) {
			return mockFileInfo{}, nil
		})
		defer mockStat.Reset()
		mockAddDevicesInDir := gomonkey.ApplyFunc(addDevicesInDir, func(spec *specs.Spec, dirPath string) error {
			return fmt.Errorf("read device dir error")
		})
		defer mockAddDevicesInDir.Reset()

		err := addUBDevice(&specInstance)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

type FakeDirEntry struct {
	name string
	typ  fs.FileMode
	info fs.FileInfo
}

func (f FakeDirEntry) Name() string               { return f.name }
func (f FakeDirEntry) IsDir() bool                { return false }
func (f FakeDirEntry) Type() fs.FileMode          { return f.typ }
func (f FakeDirEntry) Info() (fs.FileInfo, error) { return f.info, nil }

// TestAddDevicesInDir tests the function addDevicesInDir
func TestAddDevicesInDir(t *testing.T) {
	specInstance := specs.Spec{
		Linux: &specs.Linux{
			Devices: []specs.LinuxDevice{},
			Resources: &specs.LinuxResources{
				Devices: []specs.LinuxDeviceCgroup{},
			},
		},
		Process: &specs.Process{
			Env: []string{},
		},
	}

	convey.Convey("test addDevicesInDir", t, func() {
		mockReadDir := gomonkey.ApplyFunc(os.ReadDir, func(path string) ([]os.DirEntry, error) {
			return []os.DirEntry{
				FakeDirEntry{name: "udma1", typ: fs.ModeDevice},
			}, nil
		})
		defer mockReadDir.Reset()
		patch := gomonkey.ApplyFunc(addDeviceToSpec, func(spec *specs.Spec, dPath string, deviceType string) error {
			spec.Linux.Devices = append(spec.Linux.Devices, specs.LinuxDevice{
				Path: "UBDevicePath",
			})
			return nil
		})
		defer patch.Reset()

		err := addDevicesInDir(&specInstance, "/dev/udurma")
		convey.So(err, convey.ShouldBeNil)
		convey.So(specInstance.Linux.Devices, convey.ShouldNotBeEmpty)
	})
}

// TestAddDevicesInDirWithError tests the function addDevicesInDir
func TestAddDevicesInDirWithError(t *testing.T) {
	specInstance := specs.Spec{
		Linux: &specs.Linux{
			Devices: []specs.LinuxDevice{},
			Resources: &specs.LinuxResources{
				Devices: []specs.LinuxDeviceCgroup{},
			},
		},
		Process: &specs.Process{
			Env: []string{},
		},
	}

	convey.Convey("test addDevicesInDir with error", t, func() {
		convey.Convey("test addDevicesInDir should return error when readDir error", func() {
			mockReadDir := gomonkey.ApplyFunc(os.ReadDir, func(path string) ([]os.DirEntry, error) {
				return nil, fmt.Errorf("dir read error")
			})
			defer mockReadDir.Reset()

			err := addDevicesInDir(&specInstance, "/dev/udurma")
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(specInstance.Linux.Devices, convey.ShouldBeEmpty)
		})

		convey.Convey("test addDevicesInDir should return error when addDeviceToSpec error", func() {
			mockReadDir := gomonkey.ApplyFunc(os.ReadDir, func(path string) ([]os.DirEntry, error) {
				return []os.DirEntry{
					FakeDirEntry{name: "udma1", typ: fs.ModeDevice},
				}, nil
			})
			defer mockReadDir.Reset()
			patch := gomonkey.ApplyFunc(addDeviceToSpec, func(spec *specs.Spec, dPath string, deviceType string) error {
				return fmt.Errorf("add device to spec error")
			})
			defer patch.Reset()

			err := addDevicesInDir(&specInstance, "/dev/udurma")
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(specInstance.Linux.Devices, convey.ShouldBeEmpty)
		})
	})
}

// TestAddDevice tests the function addDevice
func TestAddDevice(t *testing.T) {
	devPath := "/dev/davinci1"
	statStub := gomonkey.ApplyFunc(oci.DeviceFromPath, func(name string) (*specs.LinuxDevice, error) {
		return &specs.LinuxDevice{
			Path: devPath,
		}, nil
	})
	defer statStub.Reset()

	manageDeviceStub := gomonkey.ApplyFunc(addManagerDevice, func(w dcmi.WorkerInterface, spec *specs.Spec) error {
		return nil
	})
	defer manageDeviceStub.Reset()

	spec := specs.Spec{
		Linux: &specs.Linux{
			Devices: []specs.LinuxDevice{},
			Resources: &specs.LinuxResources{
				Devices: []specs.LinuxDeviceCgroup{},
			},
		},
		Process: &specs.Process{
			Env: []string{strKubeDNSPort53UDPPort,
				"ASCEND_VISIBLE_DEVICES=1",
				"ASCEND_RUNTIME_OPTIONS=",
				strKubeDNSPort53UDPProto},
		},
	}

	ctx, _ := context.WithCancel(context.Background())
	err := InitLogModule(ctx)
	assert.Nil(t, err)
	err = addDevice(&dcmi.NpuV1Worker{DcMgr: &dcmi.DcV1Manager{}}, &spec, deviceList)
	assert.Nil(t, err)
	assert.Contains(t, spec.Linux.Devices[0].Path, devPath)
}

// TestAddDevicePatch1 tests the function addDevice
func TestAddDevicePatch1(t *testing.T) {
	convey.Convey("test addDevice patch1", t, func() {
		patches := gomonkey.ApplyFuncReturn(getValueByKey, testStr)
		patches.Reset()
		testSp := &specs.Spec{
			Process: &specs.Process{
				Env: make([]string, 0),
			},
		}
		convey.Convey("01-addDeviceToSpec error, should return error", func() {
			patch := gomonkey.ApplyFuncReturn(strings.Contains, true).
				ApplyFuncReturn(addDeviceToSpec, testError)
			defer patch.Reset()
			convey.So(addDevice(&dcmi.NpuV1Worker{DcMgr: &dcmi.DcV1Manager{}}, testSp, make([]int, 0)), convey.ShouldBeError)
		})
		convey.Convey("02-addManagerDevice error, should return error", func() {
			patch := gomonkey.ApplyFuncReturn(addDeviceToSpec, nil).
				ApplyFuncReturn(addManagerDevice, testError)
			defer patch.Reset()
			convey.So(addDevice(&dcmi.NpuV1Worker{DcMgr: &dcmi.DcV1Manager{}}, testSp, make([]int, 0)), convey.ShouldBeError)
		})
	})
}

// TestAddHook tests the function addHook
func TestAddHook(t *testing.T) {
	patch := gomonkey.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) {
		return nil, nil
	})
	defer patch.Reset()
	patchRealFileCheck := gomonkey.ApplyFunc(mindxcheckutils.RealFileChecker, func(path string,
		checkParent, allowLink bool, size int) (string, error) {
		return "", nil
	})
	defer patchRealFileCheck.Reset()
	tests := []struct {
		name         string
		spec         *specs.Spec
		deviceIdList *[]int
		wantErr      bool
	}{
		{
			name:         "success case 1",
			deviceIdList: &[]int{0},
			spec: &specs.Spec{
				Process: &specs.Process{
					Env: []string{ascendRuntimeOptions + "=VIRTUAL"},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := addHook(&dcmi.NpuV1Worker{DcMgr: &dcmi.DcV1Manager{}}, tt.spec, tt.deviceIdList); (err != nil) != tt.wantErr {
				t.Errorf("addHook() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestAddHookPatch1 tests the function addHook
func TestAddHookPatch1(t *testing.T) {
	convey.Convey("test addHook patch1", t, func() {
		convey.Convey("01-deviceList is nil, should return nil", func() {
			convey.So(addHook(&dcmi.NpuV1Worker{DcMgr: &dcmi.DcV1Manager{}}, &specs.Spec{}, nil), convey.ShouldBeNil)
		})
		testIn := make([]int, 1)
		convey.Convey("02-Executable error, should return error", func() {
			patch := gomonkey.ApplyFuncReturn(os.Executable, testStr, testError)
			defer patch.Reset()
			convey.So(addHook(&dcmi.NpuV1Worker{DcMgr: &dcmi.DcV1Manager{}}, &specs.Spec{}, &testIn), convey.ShouldBeError)
		})
		patches := gomonkey.ApplyFuncReturn(os.Executable, testStr, nil).
			ApplyFuncReturn(mindxcheckutils.RealFileChecker, testStr, nil)
		defer patches.Reset()
		convey.Convey("03-Stat error, should return error", func() {
			patch := gomonkey.ApplyFuncReturn(os.Stat, mockFileInfo{}, testError)
			defer patch.Reset()
			convey.So(addHook(&dcmi.NpuV1Worker{DcMgr: &dcmi.DcV1Manager{}}, &specs.Spec{}, &testIn), convey.ShouldBeError)
		})
		patches.ApplyFuncReturn(os.Stat, mockFileInfo{}, nil)
		convey.Convey("04-over MaxCommandLength, should return error", func() {
			testSp := specs.Spec{
				Hooks: &specs.Hooks{
					Prestart: make([]specs.Hook, MaxCommandLength+1),
				},
			}
			convey.So(addHook(&dcmi.NpuV1Worker{DcMgr: &dcmi.DcV1Manager{}}, &testSp, &testIn), convey.ShouldBeError)
		})
		convey.Convey("05-hook path contains hookCli, should return error", func() {
			testSp := specs.Spec{
				Hooks: &specs.Hooks{
					Prestart: []specs.Hook{specs.Hook{
						Path: hookCli,
					}},
				},
				Process: &specs.Process{
					Env: make([]string, MaxCommandLength+1),
				},
			}
			convey.So(addHook(&dcmi.NpuV1Worker{DcMgr: &dcmi.DcV1Manager{}}, &testSp, &testIn), convey.ShouldBeError)
		})
	})
}

// TestAddHookPatch2 tests the function addHook
func TestAddHookPatch2(t *testing.T) {
	convey.Convey("test addHook patch2", t, func() {
		patches := gomonkey.ApplyFuncReturn(os.Executable, testStr, nil).
			ApplyFuncReturn(mindxcheckutils.RealFileChecker, testStr, nil).
			ApplyFuncReturn(os.Stat, mockFileInfo{}, nil)
		defer patches.Reset()
		testIn := make([]int, 1)
		testSp := specs.Spec{
			Hooks: &specs.Hooks{
				Prestart: []specs.Hook{specs.Hook{
					Path: hookCli,
				}},
			},
			Process: &specs.Process{
				Env: make([]string, 1),
			},
		}
		convey.Convey("06-CreateVDevice error, return error", func() {
			patch := gomonkey.ApplyFuncReturn(dcmi.CreateVDevice, dcmi.VDeviceInfo{}, testError)
			defer patch.Reset()
			convey.So(addHook(&dcmi.NpuV1Worker{DcMgr: &dcmi.DcV1Manager{}}, &testSp, &testIn), convey.ShouldBeError)
		})
		patches.ApplyFuncReturn(dcmi.CreateVDevice, dcmi.VDeviceInfo{VdeviceID: 0}, nil)
		convey.Convey("07-success, should return nil", func() {
			patch := gomonkey.ApplyFuncReturn(updateEnvAndPostHook, nil)
			defer patch.Reset()
			convey.So(addHook(&dcmi.NpuV1Worker{DcMgr: &dcmi.DcV1Manager{}}, &testSp, &testIn), convey.ShouldBeNil)
		})
	})
}

// TestParseAscendDevices tests the function parseAscendDevices
func TestParseAscendDevices(t *testing.T) {
	patchGetChipName := gomonkey.ApplyMethod(reflect.TypeOf(&dcmi.NpuV1Worker{}), "GetChipName", func(f *dcmi.NpuV1Worker) (string, error) {
		return chipName, nil
	})
	defer patchGetChipName.Reset()
	tests := []struct {
		name           string
		visibleDevices string
		want           []int
		wantErr        bool
	}{
		{
			name:           "parseAscendDevices success case 1",
			visibleDevices: "Ascend910-0",
			want:           []int{0},
			wantErr:        false,
		},
		{
			name:           "parseAscendDevices success case 2",
			visibleDevices: "npu-0",
			want:           []int{0},
			wantErr:        false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseAscendDevices(tt.visibleDevices)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseAscendDevices() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equalf(t, tt.want, got, "parseAscendDevices(%v)", tt.visibleDevices)
		})
	}
}

// TestParseAscendDevicesPatch1 tests the function parseAscendDevices
func TestParseAscendDevicesPatch1(t *testing.T) {
	patchGetChipName := gomonkey.ApplyMethod(reflect.TypeOf(&dcmi.NpuV1Worker{}), "GetChipName", func(f *dcmi.NpuV1Worker) (string, error) {
		return testStr, nil
	})
	defer patchGetChipName.Reset()
	convey.Convey("test parseAscendDevices patch1", t, func() {
		convey.Convey("01-matchGroups is nil, should return error", func() {
			devs := "test,"
			_, err := parseAscendDevices(devs)
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("02-Atoi error, should return error", func() {
			devs := "Ascend910-8,"
			patch := gomonkey.ApplyFuncReturn(strconv.Atoi, 0, testError)
			defer patch.Reset()
			_, err := parseAscendDevices(devs)
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("03-GetChipName error, should return error", func() {
			devs := testStr
			_, err := parseAscendDevices(devs)
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("04-GetDeviceTypeByChipName error, should return error", func() {
			devs := testStr
			_, err := parseAscendDevices(devs)
			convey.So(err, convey.ShouldBeError)
		})
	})
}

func specWithEnv(envValue string) *specs.Spec {
	return &specs.Spec{
		Process: &specs.Process{
			Env: []string{envValue},
		},
	}
}

// TestCheckVisibleDevice tests the function checkVisibleDevice
func TestCheckVisibleDevice(t *testing.T) {
	patchGetChipName := gomonkey.ApplyMethod(reflect.TypeOf(&dcmi.NpuV1Worker{}), "GetChipName", func(f *dcmi.NpuV1Worker) (string, error) {
		return chipName, nil
	})
	defer patchGetChipName.Reset()
	tests := []struct {
		name    string
		spec    *specs.Spec
		want    []int
		wantErr bool
	}{
		{"success with Ascend910-0", specWithEnv(ascendVisibleDevices + "=Ascend910-0"), []int{0}, false},
		{"success with 0", specWithEnv(ascendVisibleDevices + "=0"), []int{0}, false},
		{"success with npu-0", specWithEnv(ascendVisibleDevices + "=npu-0"), []int{0}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := checkVisibleDevice(tt.spec)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkVisibleDevice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equalf(t, tt.want, got, "checkVisibleDevice(%v)", tt.spec)
		})
	}
}

// TestCheckVisibleDevicePatch1 tests the function checkVisibleDevice
func TestCheckVisibleDevicePatch1(t *testing.T) {
	convey.Convey("test checkVisibleDevice patch1", t, func() {
		testSpec := &specs.Spec{
			Process: &specs.Process{
				Env: []string{ascendVisibleDevices + "=Ascend910-0"},
			},
		}
		convey.Convey("01-visible devices is empty, should return nil)", func() {
			patch := gomonkey.ApplyFuncReturn(getValueByDeviceKey, "")
			defer patch.Reset()
			_, err := checkVisibleDevice(testSpec)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("02-parseAscendDevices error, should return error", func() {
			patch := gomonkey.ApplyFuncReturn(getValueByDeviceKey, "Ascend").
				ApplyFuncReturn(parseAscendDevices, []int{0}, testError)
			defer patch.Reset()
			_, err := checkVisibleDevice(testSpec)
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("03-parseDevices error, should return error", func() {
			patch := gomonkey.ApplyFuncReturn(getValueByDeviceKey, testStr).
				ApplyFuncReturn(parseDevices, []int{0}, testError)
			defer patch.Reset()
			_, err := checkVisibleDevice(testSpec)
			convey.So(err, convey.ShouldBeError)
		})
	})
}

func TestGetDeviceTypeByChipName5(t *testing.T) {
	tests := []struct {
		name     string
		chipName string
		expected string
	}{
		{
			name:     "Ascend910A5 chip type",
			chipName: "Ascend950",
			expected: Ascend910A5,
		},
		{
			name:     "Ascend910A5 chip with suffix",
			chipName: "Ascend950XX",
			expected: Ascend910A5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devType := GetDeviceTypeByChipName(tt.chipName)
			assert.EqualValues(t, tt.expected, devType)
		})
	}
}

// TestGetCommonManagerDevices tests the function getCommonManagerDevices
func TestGetCommonManagerDevices(t *testing.T) {
	var (
		ascend910A5DeviceLen = 1
		defaultDeviceLen     = 2
	)

	convey.Convey("test getCommonManagerDevices", t, func() {
		convey.Convey("01-Ascend910A5 should only return hisi_hdc", func() {
			devices := getCommonManagerDevices(Ascend910A5)
			convey.So(len(devices), convey.ShouldEqual, ascend910A5DeviceLen)
			convey.So(devices[0], convey.ShouldEqual, hisiHdc)
		})
		convey.Convey("02-other devices should return both devmm_svm and hisi_hdc", func() {
			devices := getCommonManagerDevices(Ascend910)
			convey.So(len(devices), convey.ShouldEqual, defaultDeviceLen)
			convey.So(devices, convey.ShouldContain, devmmSvm)
			convey.So(devices, convey.ShouldContain, hisiHdc)
		})
		convey.Convey("03-unknown device type should return default device list", func() {
			devices := getCommonManagerDevices("UnknownDevice")
			convey.So(len(devices), convey.ShouldEqual, defaultDeviceLen)
			convey.So(devices, convey.ShouldContain, devmmSvm)
			convey.So(devices, convey.ShouldContain, hisiHdc)
		})
	})
}
