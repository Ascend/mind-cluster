/* Copyright(C) 2023. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package device a series of common function
package device

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"Ascend-device-plugin/pkg/common"
	"Ascend-device-plugin/pkg/kubeclient"
	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
)

const (
	logicID0             = 0
	logicID1             = 1
	logicID2             = 2
	logicID3             = 3
	lengthOfDevFaultInfo = 4
	rankID0              = 0
	lengthOfDevIdList2   = 2
	lengthOfDevIdList0   = 0
	fakePod              = "default/fake-pod"
)

// TestGetChipCountOnRing for test the default count of ring and different device
func TestGetChipCountOnRing(t *testing.T) {
	convey.Convey("test GetChipCountOnRing", t, func() {
		convey.Convey("test 910 chip count on ring success", func() {
			common.ParamOption.RealCardType = api.Ascend910A
			convey.So(getResetDevNumOnce(common.Ascend910BRingsNumTrain, common.EmptyBoardId),
				convey.ShouldEqual, common.Ascend910RingsNum)
		})
		convey.Convey("test 910B train chip count on ring success", func() {
			common.ParamOption.RealCardType = api.Ascend910B
			convey.So(getResetDevNumOnce(common.Ascend910BRingsNumTrain, common.EmptyBoardId),
				convey.ShouldEqual, common.Ascend910BRingsNumTrain)
		})
		convey.Convey("test 910B Infer chip count on ring success", func() {
			common.ParamOption.RealCardType = api.Ascend910B
			convey.So(getResetDevNumOnce(common.Ascend910BRingsNumTrain, common.EmptyBoardId),
				convey.ShouldEqual, common.Ascend910BRingsNumTrain)
		})
		convey.Convey("test 910A3 chip count on ring success", func() {
			common.ParamOption.RealCardType = api.Ascend910A3
			convey.So(getResetDevNumOnce(common.Ascend910A3RingsNum, common.EmptyBoardId),
				convey.ShouldEqual, common.Ascend910A3RingsNum)
		})
	})
}

// TestGetChipCountOnRing2 for test the default count of ring and different device
func TestGetChipCountOnRing2(t *testing.T) {
	convey.Convey("test GetChipCountOnRing", t, func() {
		convey.Convey("test 910 chip count on ring success", func() {
			common.ParamOption.RealCardType = api.Ascend910B
			convey.So(getResetDevNumOnce(common.A200TA2RingsNum, common.EmptyBoardId),
				convey.ShouldEqual, common.A200TA2RingsNum)
		})
	})
}

// TestCheckConfigMap test check config map
func TestCheckConfigMap(t *testing.T) {
	convey.Convey("test checkConfigMap", t, func() {
		convey.Convey("not cm obj will return false", func() {
			cm := "fake-cm"
			res := checkConfigMap(cm)
			convey.ShouldEqual(res, false)
		})
		convey.Convey("cm's name without request prefix return false", func() {
			cm := &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "fake-name1",
					Namespace: "fake-namespace",
				},
			}
			res := checkConfigMap(cm)
			convey.ShouldEqual(res, false)
		})
		convey.Convey("cm's name with request prefix return false", func() {
			cm := &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      common.ResetInfoCMNamePrefix + "fake-name2",
					Namespace: "fake-namespace",
				},
			}
			res := checkConfigMap(cm)
			convey.ShouldEqual(res, true)
		})
	})
}

// TestHandlePodAddEvent for test handlePodAddEvent
func TestHandlePodAddEvent(t *testing.T) {
	ascend910HotResetManager := newHotResetTools()
	convey.Convey("test handlePodAddEvent", t, func() {
		patch := gomonkey.ApplyPrivateMethod(new(HotResetTools), "getPodFromCache", func(_ *HotResetTools,
			_ string) (*v1.Pod, error) {
			return nil, errors.New("pod not found")
		})
		mokeEvent := kubeclient.Event{
			Resource: kubeclient.PodResource,
			Key:      fakePod,
		}
		defer patch.Reset()
		convey.Convey("01-add pod event, should return true", func() {
			mokeEvent.Type = kubeclient.EventTypeAdd
			ascend910HotResetManager.queue.Add(mokeEvent)
			convey.So(ascend910HotResetManager.processNextWorkItem(), convey.ShouldBeTrue)
		})
		convey.Convey("02-delete pod event, should return true", func() {
			mokeEvent.Type = kubeclient.EventTypeDelete
			ascend910HotResetManager.queue.Add(mokeEvent)
			convey.So(ascend910HotResetManager.processNextWorkItem(), convey.ShouldBeTrue)
		})
		convey.Convey("03-default pod event, should return true", func() {
			mokeEvent.Type = kubeclient.EventTypeUpdate
			ascend910HotResetManager.queue.Add(mokeEvent)
			convey.So(ascend910HotResetManager.processNextWorkItem(), convey.ShouldBeTrue)
		})
		convey.Convey("04-update cm event, should return true", func() {
			mokeEvent.Type = kubeclient.EventTypeUpdate
			mokeEvent.Resource = kubeclient.CMResource
			ascend910HotResetManager.queue.Add(mokeEvent)
			convey.So(ascend910HotResetManager.processNextWorkItem(), convey.ShouldBeTrue)
		})
		convey.Convey("04-delete cm event, should return true", func() {
			mokeEvent.Type = kubeclient.EventTypeDelete
			mokeEvent.Resource = kubeclient.CMResource
			ascend910HotResetManager.queue.Add(mokeEvent)
			convey.So(ascend910HotResetManager.processNextWorkItem(), convey.ShouldBeTrue)
		})
		convey.Convey("06-default cm event, should return true", func() {
			mokeEvent.Type = kubeclient.EventTypeAdd
			mokeEvent.Resource = kubeclient.CMResource
			ascend910HotResetManager.queue.Add(mokeEvent)
			convey.So(ascend910HotResetManager.processNextWorkItem(), convey.ShouldBeTrue)
		})

	})
}

// TestHandlePodAddEventJobNameFailed test handle pod add event
func TestHandlePodAddEventJobNameFailed(t *testing.T) {
	ascend910HotResetManager := newHotResetTools()
	event := kubeclient.Event{
		Resource: kubeclient.PodResource,
		Key:      fakePod,
		Type:     kubeclient.EventTypeAdd,
	}
	convey.Convey("test TestHandlePodAddEventJobNameFailed", t, func() {
		convey.Convey("will do nothing when get job name failed", func() {
			patch := gomonkey.ApplyPrivateMethod(new(HotResetTools), "getPodFromCache", func(_ *HotResetTools,
				_ string) (*v1.Pod, error) {
				return &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{},
					},
				}, nil
			})
			defer patch.Reset()
			ascend910HotResetManager.handlePodAddEvent(event)
			_, ok := ascend910HotResetManager.jobs[event.Key]
			convey.ShouldEqual(ok, false)
		})
		convey.Convey("will do nothing when get cm failed", func() {
			patch := gomonkey.ApplyPrivateMethod(new(HotResetTools), "getPodFromCache", func(_ *HotResetTools,
				_ string) (*v1.Pod, error) {
				return &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test2",
						Namespace: "default2",
						Labels:    map[string]string{common.ResetTaskNameKey: "test-job2"},
					},
				}, nil
			})
			defer patch.Reset()
			patch2 := gomonkey.ApplyMethod(new(HotResetTools), "GetCMFromCache", func(_ *HotResetTools,
				_ string) (*v1.ConfigMap, error) {
				return nil, errors.New("cm not found")
			})
			defer patch2.Reset()
			ascend910HotResetManager.handlePodAddEvent(event)
		})
		convey.Convey("will do nothing when pod has not been cached", func() {
			patch := gomonkey.ApplyPrivateMethod(new(HotResetTools), "getPodFromCache", func(_ *HotResetTools,
				_ string) (*v1.Pod, error) {
				return nil, errors.New("pod not found")
			})
			defer patch.Reset()
			ascend910HotResetManager.handlePodAddEvent(event)
		})
	})
}

// TestHandlePodAddEventJobNameSucceed test handle pod add event
func TestHandlePodAddEventJobNameSucceed(t *testing.T) {
	ascend910HotResetManager := newHotResetTools()
	event := kubeclient.Event{
		Resource: kubeclient.PodResource,
		Key:      fakePod,
		Type:     kubeclient.EventTypeAdd,
	}
	convey.Convey("test TestHandlePodAddEventJobNameSucceed", t, func() {
		convey.Convey("will cache job when get job name success", func() {
			patch := gomonkey.ApplyPrivateMethod(new(HotResetTools), "getPodFromCache", func(_ *HotResetTools,
				_ string) (*v1.Pod, error) {
				return &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "default",
						Labels:    map[string]string{common.ResetTaskNameKey: "test-job"},
					},
				}, nil
			})
			defer patch.Reset()
			patch2 := gomonkey.ApplyMethod(new(HotResetTools), "GetCMFromCache", func(_ *HotResetTools,
				_ string) (*v1.ConfigMap, error) {
				return nil, errors.New("cm not found")
			})
			defer patch2.Reset()
			ascend910HotResetManager.handlePodAddEvent(event)
			jobName, ok := ascend910HotResetManager.jobs[event.Key]
			convey.ShouldEqual(ok, true)
			convey.ShouldEqual(jobName, "test-job")
		})
	})
}

// TestHandlePodAddEventGetCMSucceed test handle pod add event
func TestHandlePodAddEventGetCMSucceed(t *testing.T) {
	ascend910HotResetManager := newHotResetTools()
	event := kubeclient.Event{
		Resource: kubeclient.PodResource,
		Key:      fakePod,
		Type:     kubeclient.EventTypeAdd,
	}
	convey.Convey("test TestHandlePodAddEventGetCMSucceed", t, func() {
		convey.Convey("will write to file when get cm success", func() {
			patch := gomonkey.ApplyPrivateMethod(new(HotResetTools), "getPodFromCache", func(_ *HotResetTools,
				_ string) (*v1.Pod, error) {
				return &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test3",
						Namespace: "default3",
						Labels:    map[string]string{common.ResetTaskNameKey: "test-job3"},
					},
				}, nil
			})
			defer patch.Reset()
			patch2 := gomonkey.ApplyMethod(new(HotResetTools), "GetCMFromCache", func(_ *HotResetTools,
				_ string) (*v1.ConfigMap, error) {
				return &v1.ConfigMap{
					Data: map[string]string{},
				}, nil
			})
			defer patch2.Reset()
			patch3 := gomonkey.ApplyPrivateMethod(new(HotResetTools), "writeCMToFile", func(_ *HotResetTools,
				_ *v1.ConfigMap) error {
				return nil
			})
			defer patch3.Reset()
			ascend910HotResetManager.handlePodAddEvent(event)
		})
	})
}

// TestHandlePodDeleteEvent test handle pod delete event
func TestHandlePodDeleteEvent(t *testing.T) {
	ascend910HotResetManager := newHotResetTools()
	event := kubeclient.Event{
		Resource: kubeclient.PodResource,
		Key:      fakePod,
		Type:     kubeclient.EventTypeDelete,
	}
	convey.Convey("test HandlePodDeleteEvent", t, func() {
		convey.Convey("will do nothing when jobs has not been cached", func() {
			ascend910HotResetManager.handlePodDeleteEvent(event)
		})
		convey.Convey("cached job will be deleted", func() {
			ascend910HotResetManager.jobs[event.Key] = "fake-job"
			ascend910HotResetManager.handlePodDeleteEvent(event)
			_, ok := ascend910HotResetManager.jobs[event.Key]
			convey.ShouldEqual(ok, false)
		})
		convey.Convey("invalid key slice length", func() {
			event.Key = "default/fake-pod/pod1"
			ascend910HotResetManager.jobs[event.Key] = "fake-job"
			ascend910HotResetManager.handlePodDeleteEvent(event)
		})
	})
}

// TestGetPodFromCache test get cm from cache
func TestGetPodFromCache(t *testing.T) {
	ascend910HotResetManager := newHotResetTools()
	convey.Convey("test GetCMFromCache", t, func() {
		convey.Convey("get pod from cache failed when pod is not exist", func() {
			cm, err := ascend910HotResetManager.getPodFromCache("fake-name3")
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(cm, convey.ShouldBeNil)
		})
		convey.Convey("get pod from cache sucess when item is pod", func() {
			ascend910HotResetManager.podIndexer.Add(&v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "default",
				},
			})
			cm, err := ascend910HotResetManager.getPodFromCache("default/test-pod")
			convey.So(err, convey.ShouldBeNil)
			convey.So(cm, convey.ShouldNotBeNil)
		})
	})
}

// TestGetCMFromCache test get cm from cache
func TestGetCMFromCache(t *testing.T) {
	ascend910HotResetManager := newHotResetTools()
	convey.Convey("test GetCMFromCache", t, func() {
		convey.Convey("get cm from cache failed when cm is not exist", func() {
			cm, err := ascend910HotResetManager.GetCMFromCache("fake-name4")
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(cm, convey.ShouldBeNil)
		})
		convey.Convey("get cm from cache sucess when item is cm", func() {
			ascend910HotResetManager.cmIndexer.Add(&v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cm",
					Namespace: "default",
				},
			})
			cm, err := ascend910HotResetManager.GetCMFromCache("default/test-cm")
			convey.So(err, convey.ShouldBeNil)
			convey.So(cm, convey.ShouldNotBeNil)
		})
	})
}

// TestWriteCMToFileCase1 test write cm to file
func TestWriteCMToFileCase1(t *testing.T) {
	ascend910HotResetManager := newHotResetTools()
	convey.Convey("test writeCMToFile case1", t, func() {
		convey.Convey("write cm to file failed when cm has not reset.json", func() {
			cm := &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "reset-config-test1",
				},
				Data: map[string]string{"xxx": "yyy"},
			}
			err := ascend910HotResetManager.writeCMToFile(cm)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("write cm to file failed when dir is not exist", func() {
			cm := &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "reset-config-test2",
				},
				Data: map[string]string{common.ResetInfoCMDataKey: "yyy"},
			}
			err := ascend910HotResetManager.writeCMToFile(cm)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("write cm to file success when dir is exist", func() {
			cm := &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "reset-config-test2",
				},
				Data: map[string]string{common.ResetInfoCMDataKey: "yyy", common.ResetInfoTypeKey: "zzz"},
			}
			err := os.MkdirAll(common.ResetInfoDir, os.ModePerm)
			if err != nil {
				hwlog.RunLog.Error("mkdir command failed")
			}
			err = ascend910HotResetManager.writeCMToFile(cm)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestWriteCMToFileCase2 test write cm to file
func TestWriteCMToFileCase2(t *testing.T) {
	ascend910HotResetManager := newHotResetTools()
	convey.Convey("test writeCMToFile case2", t, func() {
		convey.Convey("write cm to file failed when cm has not profilingSwitch", func() {
			cm := &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "data-trace-test1",
				},
				Data: map[string]string{"xxx": "yyy"},
			}
			err := ascend910HotResetManager.writeCMToFile(cm)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("write cm to file success when dir is exist", func() {
			cm := &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "data-trace-test1",
				},
				Data: map[string]string{common.DataTraceCmProfilingSwitchKey: "xxx"},
			}
			patch := gomonkey.ApplyFuncReturn(common.WriteToFile, nil)
			defer patch.Reset()
			err := ascend910HotResetManager.writeCMToFile(cm)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestHandleConfigMapEvent for test handleConfigMapEvent
func TestHandleConfigMapEvent(t *testing.T) {
	ascend910HotResetManager := newHotResetTools()
	convey.Convey("test handleConfigMapEvent", t, func() {
		convey.Convey("01-default event type, forget event", func() {
			mokeEvent := kubeclient.Event{Type: kubeclient.EventTypeAdd}
			ascend910HotResetManager.handleConfigMapEvent(mokeEvent)
		})
	})
}

// TestHandleCMAddEventCase1 test of handleCMAddEvent
func TestHandleCMAddEventCase1(t *testing.T) {
	ascend910HotResetManager := newHotResetTools()
	convey.Convey("test handleCMAddEvent case1", t, func() {
		convey.Convey("cm not found will do nothing", func() {
			mokeEvent := kubeclient.Event{
				Resource: kubeclient.CMResource,
				Key:      "default/reset-config-test",
				Type:     kubeclient.EventTypeAdd,
			}
			ascend910HotResetManager.queue.Add(mokeEvent)
			patch := gomonkey.ApplyMethod(new(HotResetTools), "GetCMFromCache", func(_ *HotResetTools,
				_ string) (*v1.ConfigMap, error) {
				return nil, errors.New("not found")
			})
			defer patch.Reset()
			ascend910HotResetManager.handleCMUpdateEvent(mokeEvent)
		})
		mokeEvent := kubeclient.Event{
			Resource: kubeclient.CMResource,
			Key:      "default/reset-config-test",
			Type:     kubeclient.EventTypeAdd,
		}
		ascend910HotResetManager.queue.Add(mokeEvent)
		patch1 := gomonkey.ApplyMethod(new(HotResetTools), "GetCMFromCache", func(_ *HotResetTools,
			_ string) (*v1.ConfigMap, error) {
			return &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "reset-config-test",
					Namespace: "default",
				},
				Data: map[string]string{common.ResetInfoCMDataKey: "YYY"},
			}, nil
		})
		defer patch1.Reset()

		patch2 := gomonkey.ApplyPrivateMethod(new(HotResetTools), "writeCMToFile", func(_ *HotResetTools,
			_ *v1.ConfigMap) error {
			return nil
		})
		defer patch2.Reset()

		convey.Convey("os stat return error", func() {
			ascend910HotResetManager.handleCMUpdateEvent(mokeEvent)
		})
		convey.Convey("cm obj will return false", func() {
			patch3 := gomonkey.ApplyFuncReturn(os.Stat, nil, nil)
			defer patch3.Reset()

			ascend910HotResetManager.handleCMUpdateEvent(mokeEvent)
		})
	})
}

// TestHandleCMAddEventCase2 test of handleCMAddEvent
func TestHandleCMAddEventCase2(t *testing.T) {
	ascend910HotResetManager := newHotResetTools()
	convey.Convey("test handleCMAddEvent case2", t, func() {
		mokeEvent := kubeclient.Event{
			Resource: kubeclient.CMResource,
			Key:      "default/data-trace-test",
			Type:     kubeclient.EventTypeAdd,
		}
		ascend910HotResetManager.queue.Add(mokeEvent)
		patch1 := gomonkey.ApplyMethod(new(HotResetTools), "GetCMFromCache", func(_ *HotResetTools,
			_ string) (*v1.ConfigMap, error) {
			return &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "data-trace-test",
					Namespace: "default",
				},
				Data: map[string]string{common.DataTraceCmProfilingSwitchKey: "YYY"},
			}, nil
		})
		defer patch1.Reset()

		patch2 := gomonkey.ApplyPrivateMethod(new(HotResetTools), "writeCmToFileSystem", func(_ *HotResetTools,
			_ *v1.ConfigMap, _ string, _ string, _ interface{}) error {
			return nil
		})
		defer patch2.Reset()

		convey.Convey("os stat return error", func() {
			ascend910HotResetManager.handleCMUpdateEvent(mokeEvent)
		})
		convey.Convey("cm obj will return false", func() {
			patch3 := gomonkey.ApplyFuncReturn(os.Stat, nil, nil)
			defer patch3.Reset()

			ascend910HotResetManager.handleCMUpdateEvent(mokeEvent)
		})
	})
}

// TestHandleCMDeleteEvent
func TestHandleCMDeleteEvent(t *testing.T) {
	ascend910HotResetManager := newHotResetTools()
	convey.Convey("test handleCMDeleteEvent", t, func() {
		convey.Convey("test handle delete event success", func() {
			mokeEvent := kubeclient.Event{
				Resource: "",
				Key:      "fake/event",
				Type:     "",
			}
			ascend910HotResetManager.queue.Add(mokeEvent)
			ascend910HotResetManager.handleCMDeleteEvent(mokeEvent)
		})
	})
}

// TestWriteCmToFileSystem test write file to system
func TestWriteCmToFileSystem(t *testing.T) {
	ascend910HotResetManager := newHotResetTools()
	cmName := "default/data-trace-test-default"
	fileContent := "YYY"
	convey.Convey("test handleCMDeleteEvent", t, func() {
		convey.Convey("test handle delete event success", func() {
			mokeEvent := kubeclient.Event{
				Resource: kubeclient.CMResource,
				Key:      "default/" + cmName,
				Type:     kubeclient.EventTypeAdd,
			}
			cm := &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      cmName,
					Namespace: "default",
				},
				Data: map[string]string{common.DataTraceCmProfilingSwitchKey: fileContent},
			}
			ascend910HotResetManager.queue.Add(mokeEvent)
			dir := fmt.Sprintf("%s/%s", common.DataTraceConfigDir, cm.Namespace+"."+cm.Name)
			fileFullName := filepath.Join(dir, common.DataTraceCmProfilingSwitchKey)
			err := ascend910HotResetManager.writeCmToFileSystem(cm, common.DataTraceCmProfilingSwitchKey,
				fileFullName, mokeEvent)
			convey.ShouldBeNil(err)
			_, err = os.Stat(fileFullName)
			convey.ShouldBeNil(err)
			content, err := os.ReadFile(fileFullName)
			convey.ShouldBeNil(err)
			convey.ShouldEqual(string(content), fileContent)
			err = os.RemoveAll(fileFullName)
			convey.ShouldBeNil(err)
		})
	})
}

// TestWriteCmToFileSystemWithoutKey test write file to system but cm not key
func TestWriteCmToFileSystemWithoutKey(t *testing.T) {
	ascend910HotResetManager := newHotResetTools()
	cmName := "default/data-trace-test-default"
	fileContent := "YYY"
	convey.Convey("test handleCMDeleteEvent", t, func() {
		convey.Convey("test handle delete event success", func() {
			mokeEvent := kubeclient.Event{
				Resource: kubeclient.CMResource,
				Key:      "default/" + cmName,
				Type:     kubeclient.EventTypeAdd,
			}
			cm := &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      cmName,
					Namespace: "default",
				},
				Data: map[string]string{common.DataTraceCmProfilingSwitchKey + "fake": fileContent},
			}
			ascend910HotResetManager.queue.Add(mokeEvent)
			dir := fmt.Sprintf("%s/%s", common.DataTraceConfigDir, cm.Namespace+"."+cm.Name)
			fileFullName := filepath.Join(dir, common.DataTraceCmProfilingSwitchKey)
			err := ascend910HotResetManager.writeCmToFileSystem(cm, common.DataTraceCmProfilingSwitchKey,
				fileFullName, mokeEvent)
			convey.ShouldNotBeNil(err)
		})
	})
}

func newHotResetTools() *HotResetTools {
	return &HotResetTools{
		jobs:             map[string]string{},
		noResetCmPodKeys: map[string]struct{}{},
		queue:            workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
		cmIndexer:        cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{}),
		podIndexer:       cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{}),
	}
}

// TestSyncResetCMNilPodInformer verifies that SyncResetCM does not panic when
// client.PodInformer is nil (e.g. InitPodInformer has not run yet). The
// cm-based reset flow must still be set up.
func TestSyncResetCMNilPodInformer(t *testing.T) {
	convey.Convey("PodInformer is nil, SyncResetCM should not panic", t, func() {
		hrt := newHotResetTools()
		client := &kubeclient.ClientK8s{
			Clientset: fake.NewSimpleClientset(),
			Queue:     workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
		}
		// PodInformer intentionally left nil to simulate pre-init state.
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		convey.So(func() { hrt.SyncResetCM(ctx, client) }, convey.ShouldNotPanic)
		convey.So(hrt.cmIndexer, convey.ShouldNotBeNil)
	})
}
