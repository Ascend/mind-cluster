/* Copyright(C) 2022. Huawei Technologies Co.,Ltd. All rights reserved.
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
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/containerd/containerd/oci"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/stretchr/testify/assert"

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
)

var (
	deviceList = []int{1}
)

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

func TestAddHookCase1(t *testing.T) {
	var specArgs = &specs.Spec{}
	stub := gomonkey.ApplyGlobalVar(&hookCliPath, ".")
	defer stub.Reset()

	err := addHook(specArgs, &deviceList)
	assert.NotNil(t, err)
}

// TestAddHookCase2 tests the function addHook
func TestAddHookCase2(t *testing.T) {
	var specArgs = &specs.Spec{}
	stub := gomonkey.ApplyGlobalVar(&hookCliPath, ".")
	defer stub.Reset()
	stub.ApplyGlobalVar(&hookDefaultFile, ".")
	err := addHook(specArgs, &deviceList)
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
	err := addHook(specArgs, &deviceList)
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
	devicePluginHostName := devicePlugin + "pf2i6r"
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

	updateEnvAndPostHook(&spec, vdvice, &deviceList)
	assert.Contains(t, spec.Process.Env, "ASCEND_VISIBLE_DEVICES=0")
	assert.Contains(t, spec.Process.Env, "ASCEND_RUNTIME_OPTIONS=VIRTUAL")
	assert.Contains(t, spec.Hooks.Poststop[0].Path, destroyHookCli)
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

	err := addDeviceToSpec(&spec, devPath, notRenameDeviceType)
	assert.Nil(t, err)
	assert.Contains(t, spec.Linux.Devices[0].Path, devPath)
}

// TestAddAscend310BManagerDevice tests the function addAscend310BManagerDevice
func TestAddAscend310BManagerDevice(t *testing.T) {
	statStub := gomonkey.ApplyFunc(addDeviceToSpec, func(spec *specs.Spec, dPath string, deviceType string) error {
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

// TestAddCommonManagerDevice tests the function addCommonManagerDevice
func TestAddCommonManagerDevice(t *testing.T) {
	statStub := gomonkey.ApplyFunc(addDeviceToSpec, func(spec *specs.Spec, dPath string, deviceType string) error {
		return nil
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

	err := addCommonManagerDevice(&spec)
	assert.Nil(t, err)
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

	dcmiStub := gomonkey.ApplyFunc(dcmi.GetChipName, func() (string, error) {
		return chipName, nil
	})
	defer dcmiStub.Reset()

	productStub := gomonkey.ApplyFunc(dcmi.GetProductType, func(w dcmi.WorkerInterface) (string, error) {
		return "", nil
	})
	defer productStub.Reset()

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
	err = addManagerDevice(&spec)
	assert.Nil(t, err)
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

	manageDeviceStub := gomonkey.ApplyFunc(addManagerDevice, func(spec *specs.Spec) error {
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
	err = addDevice(&spec, deviceList)
	assert.Nil(t, err)
	assert.Contains(t, spec.Linux.Devices[0].Path, devPath)
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
			if err := addHook(tt.spec, tt.deviceIdList); (err != nil) != tt.wantErr {
				t.Errorf("addHook() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestParseAscendDevices tests the function parseAscendDevices
func TestParseAscendDevices(t *testing.T) {
	patchRealFileCheck := gomonkey.ApplyFunc(dcmi.GetChipName, func() (string, error) {
		return chipName, nil
	})
	defer patchRealFileCheck.Reset()
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

// TestCheckVisibleDevice tests the function checkVisibleDevice
func TestCheckVisibleDevice(t *testing.T) {
	patchRealFileCheck := gomonkey.ApplyFunc(dcmi.GetChipName, func() (string, error) {
		return chipName, nil
	})
	defer patchRealFileCheck.Reset()
	tests := []struct {
		name    string
		spec    *specs.Spec
		want    []int
		wantErr bool
	}{
		{
			name: "checkVisibleDevice success case 1",
			spec: &specs.Spec{
				Process: &specs.Process{
					Env: []string{ascendVisibleDevices + "=Ascend910-0"},
				},
			},
			wantErr: false,
			want:    []int{0},
		},
		{
			name: "checkVisibleDevice success case 2",
			spec: &specs.Spec{
				Process: &specs.Process{
					Env: []string{ascendVisibleDevices + "=0"},
				},
			},
			wantErr: false,
			want:    []int{0},
		},
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
