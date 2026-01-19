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

// Package common a series of common function
package common

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/fsnotify/fsnotify"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
)

type mockFileInfo struct {
	mode os.FileMode
	sys  interface{}
}

func (m *mockFileInfo) Name() string       { return "mock" }
func (m *mockFileInfo) Size() int64        { return 0 }
func (m *mockFileInfo) Mode() os.FileMode  { return m.mode }
func (m *mockFileInfo) ModTime() time.Time { return time.Now() }
func (m *mockFileInfo) IsDir() bool        { return false }
func (m *mockFileInfo) Sys() interface{}   { return m.sys }

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwlog.InitRunLogger(&hwLogConfig, context.Background())
}

// TestLockAllDeviceInfo for test LockAllDeviceInfo
func TestLockAllDeviceInfo(t *testing.T) {
	convey.Convey("test LockAllDeviceInfo", t, func() {
		convey.Convey("LockAllDeviceInfo success", func() {
			LockAllDeviceInfo()
			UnlockAllDeviceInfo()
		})
	})
}

// TestSetAscendRuntimeEnv for test SetAscendRuntimeEnv
func TestSetAscendRuntimeEnv(t *testing.T) {
	convey.Convey("test SetAscendRuntimeEnv", t, func() {
		id := 100
		devices := []int{id}
		SetAscendRuntimeEnv(devices, "", nil)
		ParamOption.RealCardType = api.Ascend310B
		resp := v1beta1.ContainerAllocateResponse{}
		SetAscendRuntimeEnv(devices, "", &resp)
		convey.So(resp.Envs[ascendAllowLinkEnv], convey.ShouldEqual, "True")
		convey.So(resp.Envs[AscendVisibleDevicesEnv], convey.ShouldEqual, strconv.Itoa(id))
	})
}

// TestMakeDataHash for test MakeDataHash
func TestMakeDataHash(t *testing.T) {
	convey.Convey("test MakeDataHash", t, func() {
		convey.Convey("h.Write success", func() {
			DeviceInfo := NodeDeviceInfo{DeviceList: map[string]string{HuaweiUnHealthAscend910: "Ascend910-0"}}
			ret := MakeDataHash(DeviceInfo)
			convey.So(ret, convey.ShouldNotBeEmpty)
		})
		convey.Convey("json.Marshal failed", func() {
			mockMarshal := gomonkey.ApplyFunc(json.Marshal, func(v interface{}) ([]byte, error) {
				return nil, fmt.Errorf("err")
			})
			defer mockMarshal.Reset()
			DeviceInfo := NodeDeviceInfo{DeviceList: map[string]string{HuaweiUnHealthAscend910: "Ascend910-0"}}
			ret := MakeDataHash(DeviceInfo)
			convey.So(ret, convey.ShouldBeEmpty)
		})
	})
}

// TestMapDeepCopy for test MapDeepCopy
func TestMapDeepCopy(t *testing.T) {
	convey.Convey("test MapDeepCopy", t, func() {
		convey.Convey("input nil", func() {
			ret := MapDeepCopy(nil)
			convey.So(ret, convey.ShouldNotBeNil)
		})
		convey.Convey("h.Write success", func() {
			devices := map[string]string{"100": DefaultDeviceIP}
			ret := MapDeepCopy(devices)
			convey.So(len(ret), convey.ShouldEqual, len(devices))
		})
	})
}

// TestGetDeviceFromPodAnnotation for test GetDeviceFromPodAnnotation
func TestGetDeviceFromPodAnnotation(t *testing.T) {
	convey.Convey("test GetDeviceFromPodAnnotation", t, func() {
		convey.Convey("input invalid pod", func() {
			_, err := GetDeviceFromPodAnnotation(nil, api.Ascend910)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("annotationTag not exist", func() {
			_, err := GetDeviceFromPodAnnotation(&v1.Pod{}, api.Ascend910)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("annotationTag exist", func() {
			pod := v1.Pod{}
			pod.Annotations = map[string]string{api.ResourceNamePrefix + api.Ascend910: "Ascend910-0"}
			_, err := GetDeviceFromPodAnnotation(&pod, api.Ascend910)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func createFile(filePath string) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := f.Chmod(SocketChmod); err != nil {
		return err
	}
	return nil
}

// TestGetDefaultDevices2 for GetDefaultDevices
func TestGetDefaultDevices2(t *testing.T) {
	convey.Convey("test TestGetDefaultDevices2", t, func() {
		mockStat := gomonkey.ApplyFunc(getDavinciManagerPath, func() (string, error) {
			return HiAIManagerDevice, nil
		})
		defer mockStat.Reset()
		ParamOption = Option{
			ProductTypes: []string{Atlas200ISoc},
		}
		defer func() {
			ParamOption = Option{}
		}()
		convey.Convey("set200SocDefaultDevices return err", func() {
			patch := gomonkey.ApplyFunc(set200SocDefaultDevices, func() ([]string, error) {
				return []string{Atlas200ISocVPC}, fmt.Errorf("err")
			})
			defer patch.Reset()
			_, err := GetDefaultDevices(true)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("set200SocDefaultDevices return nil", func() {
			patch := gomonkey.ApplyFunc(set200SocDefaultDevices, func() ([]string, error) {
				return []string{Atlas200ISocVPC}, nil
			})
			defer patch.Reset()
			_, err := GetDefaultDevices(true)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestGetDefaultDevices for GetDefaultDevices
func TestGetDefaultDevices(t *testing.T) {
	convey.Convey("pods is nil", t, func() {
		mockStat := gomonkey.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) {
			return nil, fmt.Errorf("err")
		})
		defer mockStat.Reset()
		_, err := GetDefaultDevices(true)
		convey.So(err, convey.ShouldNotBeNil)
	})
	if _, err := os.Stat(HiAIHDCDevice); err != nil {
		if err = createFile(HiAIHDCDevice); err != nil {
			t.Fatal("TestGetDefaultDevices Run Failed")
		}
	}

	if _, err := os.Stat(HiAIManagerDevice); err != nil {
		if err = createFile(HiAIManagerDevice); err != nil {
			t.Fatal("TestGetDefaultDevices Run Failed")
		}
	}

	if _, err := os.Stat(HiAISVMDevice); err != nil {
		if err = createFile(HiAISVMDevice); err != nil {
			t.Fatal("TestGetDefaultDevices Run Failed")
		}
	}

	if _, err := os.Stat(DvppCmdlistDevice); err != nil {
		if err = createFile(DvppCmdlistDevice); err != nil {
			t.Fatal("TestGetDefaultDevices Run Failed")
		}
	}

	defaultDevices, err := GetDefaultDevices(true)
	if err != nil {
		t.Errorf("TestGetDefaultDevices Run Failed")
	}
	defaultMap := make(map[string]string)
	defaultMap[HiAIHDCDevice] = ""
	defaultMap[HiAIManagerDevice] = ""
	defaultMap[HiAISVMDevice] = ""
	defaultMap[HiAi200RCEventSched] = ""
	defaultMap[HiAi200RCHiDvpp] = ""
	defaultMap[HiAi200RCLog] = ""
	defaultMap[HiAi200RCMemoryBandwidth] = ""
	defaultMap[HiAi200RCSVM0] = ""
	defaultMap[HiAi200RCTsAisle] = ""
	defaultMap[HiAi200RCUpgrade] = ""
	defaultMap[DvppCmdlistDevice] = ""

	for _, str := range defaultDevices {
		if _, ok := defaultMap[str]; !ok {
			t.Errorf("TestGetDefaultDevices Run Failed")
		}
	}
	t.Logf("TestGetDefaultDevices Run Pass")
}

// TestSet200SocDefaultDevices for test set200SocDefaultDevices
func TestSet200SocDefaultDevices(t *testing.T) {
	convey.Convey("test set200SocDefaultDevices", t, func() {
		convey.Convey("os.Stat err", func() {
			mockStat := gomonkey.ApplyFuncReturn(os.Stat, nil, errors.New("failed"))
			defer mockStat.Reset()
			_, err := set200SocDefaultDevices()
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("device is exist", func() {
			mockStat := gomonkey.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) {
				if name == HiAi200RCEventSched {
					return nil, fmt.Errorf("err")
				}
				return nil, nil
			})
			defer mockStat.Reset()
			ret, err := set200SocDefaultDevices()
			convey.So(ret, convey.ShouldNotBeNil)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestSet310BDefaultDevices for test set310BDefaultDevices
func TestSet310BDefaultDevices(t *testing.T) {
	convey.Convey("test set310BDefaultDevices", t, func() {
		convey.Convey("os.Stat err", func() {
			mockStat := gomonkey.ApplyFuncReturn(os.Stat, nil, errors.New("failed"))
			defer mockStat.Reset()
			convey.So(len(set310BDefaultDevices()), convey.ShouldEqual, 0)
		})
		convey.Convey("device is exist", func() {
			mockStat := gomonkey.ApplyFuncReturn(os.Stat, nil, nil)
			defer mockStat.Reset()
			convey.So(len(set310BDefaultDevices()), convey.ShouldNotEqual, 0)
		})
	})
}

func TestFilterPods1(t *testing.T) {
	convey.Convey("test FilterPods part1", t, func() {
		convey.Convey("The number of container exceeds the upper limit", func() {
			pods := []v1.Pod{{Spec: v1.PodSpec{Containers: make([]v1.Container, MaxContainerLimit+1)}}}
			res := FilterPods(pods, api.Ascend910, nil)
			convey.So(res, convey.ShouldBeEmpty)
		})
		convey.Convey("annotationTag not exist", func() {
			pods := []v1.Pod{{Spec: v1.PodSpec{Containers: []v1.Container{{Resources: v1.
				ResourceRequirements{Limits: v1.ResourceList{}}}}}}}
			res := FilterPods(pods, api.Ascend910, nil)
			convey.So(res, convey.ShouldBeEmpty)
		})
		convey.Convey("limitsDevNum exceeds the upper limit", func() {
			limits := resource.NewQuantity(MaxDevicesNum*MaxAICoreNum+1, resource.DecimalExponent)
			pods := []v1.Pod{{Spec: v1.PodSpec{Containers: []v1.Container{{Resources: v1.
				ResourceRequirements{Limits: v1.ResourceList{api.ResourceNamePrefix + Ascend910vir2: *limits}}}}}}}
			res := FilterPods(pods, Ascend910vir2, nil)
			convey.So(res, convey.ShouldBeEmpty)
		})
		convey.Convey("no assigned flag", func() {
			limits := resource.NewQuantity(1, resource.DecimalExponent)
			pods := []v1.Pod{
				{Spec: v1.PodSpec{Containers: []v1.Container{{Resources: v1.ResourceRequirements{Limits: v1.
					ResourceList{api.ResourceNamePrefix + api.Ascend910: *limits}}}}}}}
			res := FilterPods(pods, api.Ascend910, nil)
			convey.So(res, convey.ShouldBeEmpty)
		})
		convey.Convey("had assigned flag", func() {
			limits := resource.NewQuantity(1, resource.DecimalExponent)
			pods := []v1.Pod{
				{Spec: v1.PodSpec{Containers: []v1.Container{{Resources: v1.ResourceRequirements{Limits: v1.
					ResourceList{api.HuaweiAscend910: *limits}}}}},
					ObjectMeta: metav1.ObjectMeta{Name: "test3", Namespace: "test3",
						Annotations: map[string]string{
							PodPredicateTime: "1", api.HuaweiAscend910: api.Ascend910 + "-1"}},
				},
			}
			res := FilterPods(pods, api.Ascend910, nil)
			convey.So(len(res), convey.ShouldEqual, 1)
		})
	})
}

func TestFilterPods2(t *testing.T) {
	convey.Convey("test FilterPods part2", t, func() {
		limits := resource.NewQuantity(1, resource.DecimalExponent)
		pods := []v1.Pod{
			{Spec: v1.PodSpec{Containers: []v1.Container{{Resources: v1.ResourceRequirements{Limits: v1.
				ResourceList{api.HuaweiAscend910: *limits}}}}},
				ObjectMeta: metav1.ObjectMeta{Name: "test3", Namespace: "test3",
					Annotations: map[string]string{
						PodPredicateTime: "1", api.HuaweiAscend910: api.Ascend910 + "-1"},
					DeletionTimestamp: &metav1.Time{}},
				Status: v1.PodStatus{ContainerStatuses: make([]v1.ContainerStatus, 1),
					Reason: "UnexpectedAdmissionError"},
			},
		}
		convey.Convey("DeletionTimestamp is not nil", func() {
			res := FilterPods(pods, api.Ascend910, nil)
			convey.So(res, convey.ShouldBeEmpty)
		})
		pods[0].DeletionTimestamp = nil
		convey.Convey("The number of container status exceeds the upper limit", func() {
			pods[0].Status.ContainerStatuses = make([]v1.ContainerStatus, 1)
			res := FilterPods(pods, api.Ascend910, nil)
			convey.So(res, convey.ShouldBeEmpty)
		})
		convey.Convey("Waiting.Message is not nil", func() {
			pods[0].Status.ContainerStatuses = []v1.ContainerStatus{{State: v1.ContainerState{Waiting: &v1.
				ContainerStateWaiting{Message: "PreStartContainer check failed"}}}}
			res := FilterPods(pods, api.Ascend910, nil)
			convey.So(res, convey.ShouldBeEmpty)
		})
		convey.Convey("pod.Status.Reason is UnexpectedAdmissionError", func() {
			pods[0].Status = v1.PodStatus{ContainerStatuses: []v1.ContainerStatus{},
				Reason: "UnexpectedAdmissionError"}
			res := FilterPods(pods, api.Ascend910, nil)
			convey.So(res, convey.ShouldBeEmpty)
		})
		convey.Convey("conditionFunc return false", func() {
			pods[0].Status = v1.PodStatus{}
			mockConitionFunc := func(pod *v1.Pod) bool {
				return false
			}
			res := FilterPods(pods, api.Ascend910, mockConitionFunc)
			convey.So(res, convey.ShouldBeEmpty)
		})
	})
}

// TestVerifyPath for VerifyPath
func TestVerifyPath(t *testing.T) {
	convey.Convey("TestVerifyPath", t, func() {
		convey.Convey("filepath.Abs failed", func() {
			mock := gomonkey.ApplyFunc(filepath.Abs, func(path string) (string, error) {
				return "", fmt.Errorf("err")
			})
			defer mock.Reset()
			_, ret := VerifyPathAndPermission("", 0)
			convey.So(ret, convey.ShouldBeFalse)
		})
		convey.Convey("os.Stat failed", func() {
			mock := gomonkey.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) {
				return nil, fmt.Errorf("err")
			})
			defer mock.Reset()
			_, ret := VerifyPathAndPermission("./", 0)
			convey.So(ret, convey.ShouldBeFalse)
		})
		convey.Convey("filepath.EvalSymlinks failed", func() {
			mock := gomonkey.ApplyFunc(filepath.EvalSymlinks, func(path string) (string, error) {
				return "", fmt.Errorf("err")
			})
			defer mock.Reset()
			_, ret := VerifyPathAndPermission("./", 0)
			convey.So(ret, convey.ShouldBeFalse)
		})
	})
}

// TestCheckPodNameAndSpace for test CheckPodNameAndSpace
func TestCheckPodNameAndSpace(t *testing.T) {
	convey.Convey("test CheckPodNameAndSpace", t, func() {
		convey.Convey("beyond max length", func() {
			podPara, maxLength := "abc", 1
			convey.So(CheckPodNameAndSpace(podPara, maxLength), convey.ShouldNotBeNil)
		})
		convey.Convey("device is exist", func() {
			podPara, maxLength := "abc", PodNameMaxLength
			convey.So(CheckPodNameAndSpace(podPara, maxLength), convey.ShouldBeNil)
			podPara = "abc_d"
			convey.So(CheckPodNameAndSpace(podPara, maxLength), convey.ShouldNotBeNil)
		})
	})
}

// TestWatchFile for test watchFile
func TestWatchFile(t *testing.T) {
	convey.Convey("TestWatchFile", t, func() {
		convey.Convey("fsnotify.NewWatcher ok", func() {
			watcher, err := NewFileWatch()
			convey.So(err, convey.ShouldBeNil)
			convey.So(watcher, convey.ShouldNotBeNil)
		})
		convey.Convey("fsnotify.NewWatcher failed", func() {
			mock := gomonkey.ApplyFunc(fsnotify.NewWatcher, func() (*fsnotify.Watcher, error) {
				return nil, fmt.Errorf("error")
			})
			defer mock.Reset()
			watcher, err := NewFileWatch()
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(watcher, convey.ShouldBeNil)
		})
		watcher, _ := NewFileWatch()
		convey.Convey("stat failed", func() {
			convey.So(watcher, convey.ShouldNotBeNil)
			mock := gomonkey.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) {
				return nil, fmt.Errorf("err")
			})
			defer mock.Reset()
			err := watcher.WatchFile("")
			convey.So(err, convey.ShouldNotBeNil)
		})
		mockStat := gomonkey.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) {
			return nil, nil
		})
		defer mockStat.Reset()
		convey.Convey("Add failed", func() {
			convey.So(watcher, convey.ShouldNotBeNil)
			mockWatchFile := gomonkey.ApplyMethod(reflect.TypeOf(new(fsnotify.Watcher)), "Add",
				func(_ *fsnotify.Watcher, name string) error { return fmt.Errorf("err") })
			defer mockWatchFile.Reset()
			err := watcher.WatchFile("")
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("Add ok", func() {
			convey.So(watcher, convey.ShouldNotBeNil)
			mockWatchFile := gomonkey.ApplyMethod(reflect.TypeOf(new(fsnotify.Watcher)), "Add",
				func(_ *fsnotify.Watcher, name string) error { return nil })
			defer mockWatchFile.Reset()
			err := watcher.WatchFile("")
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestGetDeviceListID for test GetDeviceListID
func TestGetDeviceListID(t *testing.T) {
	convey.Convey("TestGetDeviceListID", t, func() {
		convey.Convey("device num excceed max num", func() {
			devices := make([]string, MaxDevicesNum+1)
			_, _, ret := GetDeviceListID(devices, "")
			convey.So(ret, convey.ShouldNotBeNil)
		})
		convey.Convey("device name is invalid", func() {
			devices := []string{"Ascend910"}
			_, _, ret := GetDeviceListID(devices, "")
			convey.So(ret, convey.ShouldNotBeNil)
		})
		convey.Convey("physical device", func() {
			devices := []string{"Ascend910-0"}
			_, ascendVisibleDevices, ret := GetDeviceListID(devices, "")
			convey.So(ret, convey.ShouldBeNil)
			convey.So(len(ascendVisibleDevices), convey.ShouldEqual, 1)
		})
		convey.Convey("virtual device", func() {
			devices := []string{"Ascend910-2c-100-0"}
			_, ascendVisibleDevices, ret := GetDeviceListID(devices, VirtualDev)
			convey.So(ret, convey.ShouldBeNil)
			convey.So(len(ascendVisibleDevices), convey.ShouldEqual, 1)
		})
	})
}

// TestGetPodConfiguration for test GetPodConfiguration
func TestGetPodConfiguration(t *testing.T) {
	convey.Convey("TestGetPodConfiguration", t, func() {
		convey.Convey("Marshal failed", func() {
			mockMarshal := gomonkey.ApplyFunc(json.Marshal, func(v interface{}) ([]byte, error) {
				return nil, fmt.Errorf("err")
			})
			defer mockMarshal.Reset()
			devices := map[int]string{100: DefaultDeviceIP}
			phyDevMapVirtualDev := map[int]int{100: 0}
			deviceType := "Ascend910-2c"
			superPodID := int32(1)
			info := ServerInfo{
				ServerID:   DefaultDeviceIP,
				DeviceType: deviceType,
				SuperPodID: superPodID,
			}
			ret := GetPodConfiguration(phyDevMapVirtualDev, devices, "pod-name", info, nil)
			convey.So(ret, convey.ShouldBeEmpty)
		})
		convey.Convey("Marshal ok", func() {
			devices := map[int]string{100: DefaultDeviceIP}
			phyDevMapVirtualDev := map[int]int{100: 0}
			deviceType := "Ascend910-2c"
			superPodID := int32(1)
			info := ServerInfo{
				ServerID:   DefaultDeviceIP,
				DeviceType: deviceType,
				SuperPodID: superPodID,
			}
			ret := GetPodConfiguration(phyDevMapVirtualDev, devices, "pod-name", info, nil)
			convey.So(ret, convey.ShouldNotBeEmpty)
		})
	})
}

// TestNewSignWatcher for test NewSignWatcher
func TestNewSignWatcher(t *testing.T) {
	convey.Convey("TestNewSignWatcher", t, func() {
		signChan := NewSignWatcher(syscall.SIGHUP)
		convey.So(signChan, convey.ShouldNotBeNil)
	})
}

// TestCheckFileUserSameWithProcess for test CheckFileUserSameWithProcess
func TestCheckFileUserSameWithProcess(t *testing.T) {
	convey.Convey("CheckFileUserSameWithProcess", t, func() {
		var testMode os.FileMode = 0660
		loggerPath := "/home/test"
		convey.Convey("user is root", func() {
			mockFunc := gomonkey.ApplyFuncReturn(os.Getuid, RootUID)
			defer mockFunc.Reset()
			convey.So(CheckFileUserSameWithProcess(loggerPath), convey.ShouldBeTrue)
		})
		convey.Convey("user is not root, logger path is unavailable", func() {
			mockFunc := gomonkey.ApplyFuncReturn(os.Getuid, 1).
				ApplyFuncReturn(os.Lstat, &mockFileInfo{}, fmt.Errorf("get path stat failed"))
			defer mockFunc.Reset()
			convey.So(CheckFileUserSameWithProcess(loggerPath), convey.ShouldBeFalse)
		})
		convey.Convey("user is not root, logger file is unavailable", func() {
			mockFunc := gomonkey.ApplyFuncReturn(os.Getuid, 1).
				ApplyFuncReturn(os.Lstat, &mockFileInfo{mode: testMode, sys: "invalid-type"}, nil)
			defer mockFunc.Reset()
			convey.So(CheckFileUserSameWithProcess(loggerPath), convey.ShouldBeFalse)
		})
		convey.Convey("user is not root, logger file stat uid or gid is not equal curUid", func() {
			mockFunc := gomonkey.ApplyFuncReturn(os.Getuid, 1).
				ApplyFuncReturn(os.Lstat, &mockFileInfo{mode: testMode, sys: &syscall.Stat_t{Uid: RootUID}}, nil)
			defer mockFunc.Reset()
			convey.So(CheckFileUserSameWithProcess(loggerPath), convey.ShouldBeFalse)
		})
		convey.Convey("user is not root, both logger file stat uid and gid are equal curUid", func() {
			mockFunc := gomonkey.ApplyFuncReturn(os.Getuid, 1).
				ApplyFuncReturn(os.Lstat, &mockFileInfo{mode: testMode, sys: &syscall.Stat_t{Uid: 1, Gid: 1}}, nil)
			defer mockFunc.Reset()
			convey.So(CheckFileUserSameWithProcess(loggerPath), convey.ShouldBeTrue)
		})
	})
}

// TestIsContainAtlas300IDuo for test IsContainAtlas300IDuo
func TestIsContainAtlas300IDuo(t *testing.T) {
	convey.Convey("IsContainAtlas300IDuo", t, func() {
		convey.Convey("IsContainAtlas300IDuo success", func() {
			ParamOption.ProductTypes = nil
			convey.So(IsContainAtlas300IDuo(), convey.ShouldBeFalse)
			ParamOption.ProductTypes = []string{Atlas300IDuo}
			convey.So(IsContainAtlas300IDuo(), convey.ShouldBeTrue)
		})
	})
}

// TestRecordFaultInfoList for test RecordFaultInfoList
func TestRecordFaultInfoList(t *testing.T) {
	convey.Convey("test RecordFaultInfoList success", t, func() {
		RecordFaultInfoList([]*TaskDevInfo{{}})
	})
}

// TestInt32Join for test Int32Join
func TestInt32Join(t *testing.T) {
	convey.Convey("test Int32Join success", t, func() {
		convey.So(Int32Join([]int32{0, 1}, ","), convey.ShouldEqual, "0,1")
	})
}

// TestGetDeviceRunMode for test GetDeviceRunMode
func TestGetDeviceRunMode(t *testing.T) {
	convey.Convey("test GetDeviceRunMode", t, func() {
		convey.Convey("device mode is Ascend310", func() {
			ParamOption.RealCardType = api.Ascend310
			ret, err := GetDeviceRunMode()
			convey.So(ret, convey.ShouldEqual, api.Ascend310)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("device mode is Ascend910, when card real type is Ascend910", func() {
			ParamOption.RealCardType = api.Ascend910
			ret, err := GetDeviceRunMode()
			convey.So(ret, convey.ShouldEqual, api.Ascend910)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("device mode is Ascend910, when card real type is Ascend910B", func() {
			ParamOption.RealCardType = api.Ascend910B
			ret, err := GetDeviceRunMode()
			convey.So(ret, convey.ShouldEqual, api.Ascend910)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("device mode is Ascend910, when card real type is Atlas A3", func() {
			ParamOption.RealCardType = api.Ascend910A3
			ret, err := GetDeviceRunMode()
			convey.So(ret, convey.ShouldEqual, api.Ascend910)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("device mode is Ascend310P", func() {
			ParamOption.RealCardType = api.Ascend310P
			ret, err := GetDeviceRunMode()
			convey.So(ret, convey.ShouldEqual, api.Ascend310P)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("device mode is invalid", func() {
			ParamOption.RealCardType = ""
			_, err := GetDeviceRunMode()
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestCheckDeviceName for test CheckDeviceName
func TestCheckDeviceName(t *testing.T) {
	convey.Convey("test CheckDeviceName", t, func() {
		convey.Convey("device name is valid", func() {
			convey.So(CheckDeviceName(api.Ascend910+"-0", api.Ascend910), convey.ShouldBeTrue)
		})
		convey.Convey("device name is invalid", func() {
			convey.So(CheckDeviceName("", api.Ascend910), convey.ShouldBeFalse)
		})
	})
}

// TestGetJobNameOfPod
func TestGetJobNameOfPod(t *testing.T) {
	convey.Convey("test GetJobNameOfPod", t, func() {
		const fakeJobName = "job1"
		convey.Convey("pod has vcjob name", func() {
			pod := &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						ResetTaskNameKey: fakeJobName,
					},
				},
			}
			jobName := GetJobNameOfPod(pod)
			convey.ShouldEqual(jobName, fakeJobName)
		})
		convey.Convey("pod has acjob name", func() {
			pod := &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						ResetTaskNameKeyInLabel: fakeJobName,
					},
				},
			}
			jobName := GetJobNameOfPod(pod)
			convey.ShouldEqual(jobName, fakeJobName)
		})
		convey.Convey("pod has no name", func() {
			pod := &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{},
				},
			}
			jobName := GetJobNameOfPod(pod)
			convey.ShouldEqual(jobName, "")
		})
	})
}

// TestCheckDeviceName for test GetSyncMapLen
func TestGetSyncMapLen(t *testing.T) {
	t.Run("TestGetSyncMapLen", func(t *testing.T) {
		m := &sync.Map{}
		m.Store("key1", "value1")
		m.Store("key2", "value2")
		if got := GetSyncMapLen(m); got != 2 {
			t.Errorf("GetSyncMapLen() = %v, want %v", got, 2)
		}
	})
}

// TestObjToString for test ObjToString
func TestObjToString(t *testing.T) {
	t.Run("TestObjToString", func(t *testing.T) {
		mp := map[string]string{"hello": "world"}
		want := `{"hello":"world"}`
		if got := ObjToString(mp); got != want {
			t.Errorf("ObjToString failed")
		}
	})
}

// TestKeys for test Keys
func TestKeys(t *testing.T) {
	t.Run("TestKeys", func(t *testing.T) {
		mp := map[string]string{"key": "value"}
		want := "key"
		if got := Keys(mp); got[0] != want {
			t.Errorf("Keys() = %v, want %v", got[0], want)
		}
	})
}

// TestSetDeviceByPathWhen200RC for test setDeviceByPathWhen200RC
func TestSetDeviceByPathWhen200RC(t *testing.T) {
	convey.Convey("test setDeviceByPathWhen200RC", t, func() {
		// 01-stub os.Stat, set device success, devices should be updated
		mockStat := gomonkey.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) {
			return nil, nil
		})
		defer mockStat.Reset()
		devices := []string{}
		setDeviceByPathWhen200RC(&devices)
		convey.So(len(devices), convey.ShouldEqual, 7)
		convey.So(devices[0], convey.ShouldEqual, HiAi200RCEventSched)
		if len(devices) >= 7 {
			convey.So(devices[6], convey.ShouldEqual, HiAi200RCUpgrade)
		}
	})
}

// TestGetPodNameFromEnv for test GetPodNameFromEnv
func TestGetPodNameFromEnv(t *testing.T) {
	convey.Convey("test getPodNameFromEnv", t, func() {
		convey.Convey("01-check pod name and space failed, should return error", func() {
			mockGetEnv := gomonkey.ApplyFuncReturn(os.Getenv, "master-1")
			defer mockGetEnv.Reset()
			mockCheck := gomonkey.ApplyFuncReturn(CheckPodNameAndSpace, errors.New("failed"))
			defer mockCheck.Reset()
			_, err := GetPodNameFromEnv()
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("02-check pod name and space success, should return nil", func() {
			mockGetEnv := gomonkey.ApplyFuncReturn(os.Getenv, "master-1")
			defer mockGetEnv.Reset()
			podName, err := GetPodNameFromEnv()
			convey.So(err, convey.ShouldBeNil)
			convey.So(podName == "master-1", convey.ShouldBeTrue)
		})
	})
}

// TestIsContainAll300IDuo for test IsContainAll300IDuo
func TestIsContainAll300IDuo(t *testing.T) {
	convey.Convey("test IsContainAll300IDuo", t, func() {
		convey.Convey("01-productType length is zero, should return false", func() {
			mockParam := gomonkey.ApplyGlobalVar(&ParamOption, Option{})
			defer mockParam.Reset()
			convey.So(IsContainAll300IDuo(), convey.ShouldBeFalse)
		})
		convey.Convey("02-productType has Atlas300IDuo, should return true", func() {
			mockParam := gomonkey.ApplyGlobalVar(&ParamOption, Option{ProductTypes: []string{Atlas300IDuo}})
			defer mockParam.Reset()
			convey.So(IsContainAll300IDuo(), convey.ShouldBeTrue)
		})
		convey.Convey("03-productType has other type, should return false", func() {
			mockParam := gomonkey.ApplyGlobalVar(&ParamOption, Option{ProductTypes: []string{Atlas300IDuo, "other type"}})
			defer mockParam.Reset()
			convey.So(IsContainAll300IDuo(), convey.ShouldBeFalse)
		})
	})
}

// TestIntInList for test IntInList
func TestIntInList(t *testing.T) {
	convey.Convey("test intInList", t, func() {
		list := []int32{1, 2, 3}
		// 01-list has target number, should return true
		convey.So(IntInList(2, list), convey.ShouldBeTrue)
		// 01-list has not target number, should return false
		convey.So(IntInList(4, list), convey.ShouldBeFalse)
	})
}

// TestCompareStringSetMap for test CompareStringSetMap
func TestCompareStringSetMap(t *testing.T) {
	tests := []struct {
		name     string
		map1     map[string]sets.String
		map2     map[string]sets.String
		expected bool
	}{
		{name: "Both maps are nil, should return true", map1: nil, map2: nil, expected: true},
		{name: "One map is nil and the other is not, should return false",
			map1: nil, map2: map[string]sets.String{}, expected: false},
		{name: "Maps have different lengths, should return false",
			map1: map[string]sets.String{"key1": sets.NewString("subKey1")},
			map2: map[string]sets.String{"key1": sets.NewString("subKey1"),
				"key2": sets.NewString("subKey1")}, expected: false},
		{name: "Maps have different keys, should return false",
			map1: map[string]sets.String{"key1": sets.NewString("subKey1")},
			map2: map[string]sets.String{"key2": sets.NewString("subKey1")}, expected: false},
		{name: "Maps have different values, should return false",
			map1: map[string]sets.String{"key": sets.NewString("subKey1")},
			map2: map[string]sets.String{"key": sets.NewString("subKey2")}, expected: false},
		{name: "Maps are identical, should return true",
			map1: map[string]sets.String{"key": sets.NewString("subKey")},
			map2: map[string]sets.String{"key": sets.NewString("subKey")}, expected: true},
		{name: "Maps have same keys and values but different order, should return true",
			map1: map[string]sets.String{"key": sets.NewString("subKey1", "subKey2")},
			map2: map[string]sets.String{"key": sets.NewString("subKey2", "subKey1")}, expected: true},
		{name: "Both maps are empty, should return true",
			map1: map[string]sets.String{}, map2: map[string]sets.String{}, expected: true},
	}

	convey.Convey("Test CompareStringSetMap function", t, func() {
		for _, tc := range tests {
			convey.Convey(tc.name, func() {
				res := CompareStringSetMap(tc.map1, tc.map2)
				convey.So(res, convey.ShouldEqual, tc.expected)
			})
		}
	})
}

// TestGetSuperDeviceID for test getSuperDeviceID
func TestGetSuperDeviceID(t *testing.T) {
	convey.Convey("test getSuperDeviceID", t, func() {
		deviceId := 2
		allDevices := []NpuDevice{
			{PhyID: 0, SuperDeviceID: 1},
		}
		suberDeviceId := getSuperDeviceID(0, allDevices)
		convey.So(suberDeviceId, convey.ShouldEqual, 1)
		suberDeviceId = getSuperDeviceID(deviceId, allDevices)
		convey.So(suberDeviceId, convey.ShouldEqual, SdIdAbnormal)
	})
}

func TestTriggerUpdate(t *testing.T) {
	convey.Convey("trigger update success", t, func() {
		verifyUpdateTrigger(t)
		TriggerUpdate("test trigger update")
		convey.So(verifyUpdateTrigger(t), convey.ShouldBeTrue)
	})
	convey.Convey("not trigger update", t, func() {
		verifyUpdateTrigger(t)
		if updateTriggerChan == nil {
			t.Error("updateTriggerChan is nil")
		}
		updateTriggerChan <- struct{}{}
		TriggerUpdate("test trigger update")
		convey.So(verifyUpdateTrigger(t), convey.ShouldBeTrue)
	})
}

func verifyUpdateTrigger(t *testing.T) bool {
	if updateTriggerChan == nil {
		t.Error("updateTriggerChan is nil")
	}
	select {
	case <-updateTriggerChan:
		return true
	default:
		return false
	}
}
