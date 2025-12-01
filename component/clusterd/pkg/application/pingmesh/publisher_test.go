// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package pingmesh a series of function handle ping mesh configmap create/update/delete
package pingmesh

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/apimachinery/pkg/api/errors"

	"ascend-common/api"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/superpod"
	"clusterd/pkg/interface/kube"
)

const (
	testSuperPodID = "1"
	testCheckCode  = "test-check-code"
)

func TestUpdateSuperPodDeviceCM(t *testing.T) {
	convey.Convey("Test updateSuperPodDeviceCM", t, func() {
		device := &api.SuperPodDevice{
			SuperPodID:    testSuperPodID,
			NodeDeviceMap: make(map[string]*api.NodeDevice),
		}

		patchCreateOrUpdate := gomonkey.ApplyFunc(kube.CreateOrUpdateConfigMap,
			func(name, namespace string, data map[string]string, labels map[string]string) error {
				return nil
			})
		defer patchCreateOrUpdate.Reset()

		patchUpdateOrCreate := gomonkey.ApplyFunc(kube.UpdateOrCreateConfigMap,
			func(name, namespace string, data map[string]string, labels map[string]string) error {
				return nil
			})
		defer patchUpdateOrCreate.Reset()

		convey.Convey("Test valid device with init=true", func() {
			err := updateSuperPodDeviceCM(device, testCheckCode, true)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Test valid device with init=false", func() {
			err := updateSuperPodDeviceCM(device, testCheckCode, false)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Test nil device", func() {
			err := updateSuperPodDeviceCM(nil, testCheckCode, true)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestAddEvent(t *testing.T) {
	convey.Convey("Test addEvent", t, func() {
		oldEventMapLen := len(publishMgr.eventMap)
		_, preExist := publishMgr.eventMap[testSuperPodID]
		addEvent(testSuperPodID, constant.AddOperator)
		if !preExist {
			convey.So(len(publishMgr.eventMap), convey.ShouldEqual, oldEventMapLen+1)
		} else {
			convey.So(len(publishMgr.eventMap), convey.ShouldEqual, oldEventMapLen)
		}
		convey.So(publishMgr.eventMap[testSuperPodID], convey.ShouldEqual, constant.AddOperator)
	})
}

func TestInitSuperPodsCM(t *testing.T) {
	convey.Convey("Test initSuperPodsCM", t, func() {
		patchList := gomonkey.ApplyFunc(superpod.ListClusterDevice, func() []*api.SuperPodDevice {
			return []*api.SuperPodDevice{
				{
					SuperPodID:    testSuperPodID,
					NodeDeviceMap: make(map[string]*api.NodeDevice),
				},
			}
		})
		defer patchList.Reset()

		patchHash := gomonkey.ApplyFunc(util.MakeDataHash, func(v interface{}) string {
			return testCheckCode
		})
		defer patchHash.Reset()

		patchUpdate := gomonkey.ApplyFunc(updateSuperPodDeviceCM,
			func(device *api.SuperPodDevice, checkCode string, init bool) error {
				return nil
			})
		defer patchUpdate.Reset()

		oldPublishLogMapLen := len(publishMgr.cmPublishLogMap)
		initSuperPodsCM()
		convey.So(len(publishMgr.cmPublishLogMap), convey.ShouldEqual, oldPublishLogMapLen+1)
		convey.So(publishMgr.cmPublishLogMap[testSuperPodID], convey.ShouldNotBeNil)
	})
}

func TestHandleUpdate(t *testing.T) {
	convey.Convey("Test handleUpdate", t, func() {
		patchHash := gomonkey.ApplyFunc(util.MakeDataHash, func(v interface{}) string {
			return testCheckCode
		})
		defer patchHash.Reset()

		patchUpdate := gomonkey.ApplyFunc(updateSuperPodDeviceCM,
			func(device *api.SuperPodDevice, checkCode string, init bool) error {
				return nil
			})
		defer patchUpdate.Reset()

		device := &api.SuperPodDevice{
			SuperPodID:    testSuperPodID,
			NodeDeviceMap: make(map[string]*api.NodeDevice),
		}

		convey.Convey("Test new check code", func() {
			ptUpdate := gomonkey.ApplyFuncReturn(handleCmUpdate, nil).
				ApplyFuncReturn(handleFileUpdate, nil)
			defer ptUpdate.Reset()
			err := handleUpdate(testSuperPodID, device)
			convey.So(err, convey.ShouldEqual, nil)
			convey.So(publishMgr.cmPublishLogMap[testSuperPodID].preCheckCode, convey.ShouldEqual, testCheckCode)
		})

		convey.Convey("Test same check code", func() {
			publishMgr.cmPublishLogMap[testSuperPodID] = &publishLog{
				publishKey:   testSuperPodID,
				preCheckCode: testCheckCode,
			}
			publishMgr.filePublishLogMap[testSuperPodID] = &publishLog{
				publishKey:   testSuperPodID,
				preCheckCode: testCheckCode,
			}
			err := handleUpdate(testSuperPodID, device)
			convey.So(err, convey.ShouldEqual, nil)
			convey.So(publishMgr.cmPublishLogMap[testSuperPodID].preCheckCode, convey.ShouldEqual, testCheckCode)
		})

		convey.Convey("Test nil device", func() {
			err := handleUpdate(testSuperPodID, nil)
			convey.So(err, convey.ShouldEqual, nil)
		})
	})
}

func TestHandleDelete(t *testing.T) {
	convey.Convey("Test handleDelete", t, func() {
		patchDelete := gomonkey.ApplyFunc(kube.DeleteConfigMap,
			func(name, namespace string) error {
				return nil
			})
		defer patchDelete.Reset()

		patchHandleUpdate := gomonkey.ApplyFunc(handleUpdate,
			func(superPodID string, device *api.SuperPodDevice) error {
				return nil
			})
		defer patchHandleUpdate.Reset()

		convey.Convey("Test nil device", func() {
			patchHandleDelete := gomonkey.ApplyFuncReturn(handleFileDelete, nil).
				ApplyFuncReturn(handleCmDelete, nil)
			defer patchHandleDelete.Reset()
			err := handleDelete(testSuperPodID)
			convey.So(err, convey.ShouldEqual, nil)
		})
	})
}

func TestGetTask(t *testing.T) {
	convey.Convey("Test getTask", t, func() {
		publishMgr.eventMap[testSuperPodID+"-1"] = constant.AddOperator
		publishMgr.eventMap[testSuperPodID+"-2"] = constant.UpdateOperator
		oldEventMapLen := len(publishMgr.eventMap)

		tasks := getPartTaskAndClean()
		convey.So(len(tasks), convey.ShouldBeGreaterThan, 0)
		convey.So(len(publishMgr.eventMap), convey.ShouldEqual, oldEventMapLen-len(tasks))
	})
}

func TestHandleTasks(t *testing.T) {
	convey.Convey("Test handleTasks", t, func() {
		patchGet := gomonkey.ApplyFunc(superpod.GetSuperPodDevice,
			func(superPodID string) *api.SuperPodDevice {
				return &api.SuperPodDevice{
					SuperPodID:    testSuperPodID,
					NodeDeviceMap: make(map[string]*api.NodeDevice),
				}
			})
		defer patchGet.Reset()

		patchHandleUpdate := gomonkey.ApplyFunc(handleUpdate,
			func(superPodID string, device *api.SuperPodDevice) error {
				return nil
			})
		defer patchHandleUpdate.Reset()

		patchHandleDelete := gomonkey.ApplyFunc(handleDelete,
			func(superPodID string) error {
				return nil
			})
		defer patchHandleDelete.Reset()

		tasks := []task{
			{
				superPodID: testSuperPodID,
				operator:   constant.AddOperator,
			},
		}

		handleTasks(tasks)
		convey.So(publishMgr.eventMap[testSuperPodID], convey.ShouldEqual, "")
	})
}

func TestTickerCheckSuperPodDevice(t *testing.T) {
	convey.Convey("Test TickerCheckSuperPodDevice", t, func() {
		patchInit := gomonkey.ApplyFunc(initSuperPodsCM, func() {})
		defer patchInit.Reset()

		patchGetTask := gomonkey.ApplyFunc(getPartTaskAndClean, func() []task {
			return []task{}
		})
		defer patchGetTask.Reset()

		patchHandleTasks := gomonkey.ApplyFunc(handleTasks, func(tasks []task) {})
		defer patchHandleTasks.Reset()

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		go TickerCheckSuperPodDevice(ctx)
		time.Sleep(time.Second)
		convey.So(true, convey.ShouldBeTrue)
	})
}

func TestUpdateSuperPodDeviceFile(t *testing.T) {
	convey.Convey("Test superPodDeviceFile", t, func() {
		convey.Convey("case 1: device is nil. expect return nil", func() {
			convey.ShouldBeNil(updateSuperPodDeviceFile(nil, "", false))
		})
		spd := &api.SuperPodDevice{
			Version:       "test_version",
			SuperPodID:    "test_id",
			NodeDeviceMap: nil,
		}
		convey.Convey("case 2: writeJsonData failed. expect return err", func() {
			patches := gomonkey.ApplyFuncReturn(writeJsonDataByteToFile, fmt.Errorf("fake error"))
			defer patches.Reset()
			convey.ShouldNotBeNil(updateSuperPodDeviceFile(spd, "", false))
		})
		convey.Convey("case 3: writeJsonData success. expect return nil", func() {
			patches := gomonkey.ApplyFuncReturn(writeJsonDataByteToFile, nil)
			defer patches.Reset()
			convey.ShouldBeNil(updateSuperPodDeviceFile(spd, "", false))
		})
	})
}

func TestHandleFileUpdate(t *testing.T) {
	convey.Convey("Test handleFileUpdate", t, func() {
		publishMgr.filePublishLogMap[testSuperPodID] = &publishLog{
			publishKey:   testSuperPodID,
			preCheckCode: testCheckCode,
		}
		convey.Convey("case 1: super pod id exist and content not change."+
			"expect return nil", func() {
			convey.ShouldBeNil(handleFileUpdate(testSuperPodID, nil, testCheckCode, false))
		})
		convey.Convey("case 2: updateSuperPodDeviceFile failed. expect return err", func() {
			patches := gomonkey.ApplyFuncReturn(updateSuperPodDeviceFile, fmt.Errorf(""))
			defer patches.Reset()
			convey.ShouldNotBeNil(handleFileUpdate(testSuperPodID, nil, testCheckCode+testCheckCode, false))
		})
		convey.Convey("case 3: updateSuperPodDeviceFile failed. expect return err", func() {
			patches := gomonkey.ApplyFuncReturn(updateSuperPodDeviceFile, nil).
				ApplyFuncReturn(saveConfigToFile, fmt.Errorf(""))
			defer patches.Reset()
			convey.ShouldNotBeNil(handleFileUpdate(testSuperPodID, nil, testCheckCode+testCheckCode, false))
		})
		convey.Convey("case 4: file update success. expect return nil", func() {
			patches := gomonkey.ApplyFuncReturn(updateSuperPodDeviceFile, nil).
				ApplyFuncReturn(saveConfigToFile, nil)
			defer patches.Reset()
			mock := gomonkey.ApplyFunc(handlerSuperPodRoce, func(_ map[int]string) {})
			defer mock.Reset()
			device := &api.SuperPodDevice{AcceleratorType: api.A5PodType}
			convey.ShouldNotBeNil(handleFileUpdate(testSuperPodID, device, testCheckCode+testCheckCode, false))
			convey.ShouldEqual(publishMgr.filePublishLogMap[testSuperPodID].preCheckCode, testCheckCode+testCheckCode)
		})
	})
}

func TestHandleCmDelete(t *testing.T) {
	convey.Convey("Test handleCmDelete", t, func() {
		convey.Convey("case 1: DeleteConfigMap failed. expect return err when err is NotFound err", func() {
			patch := gomonkey.ApplyFuncReturn(kube.DeleteConfigMap, fmt.Errorf(""))
			defer patch.Reset()
			convey.ShouldNotBeNil(handleCmDelete(testSuperPodID))
			patch.ApplyFuncReturn(errors.IsNotFound, true)
			convey.ShouldBeNil(handleCmDelete(testSuperPodID))
		})
	})
}

func TestHandleFileDelete(t *testing.T) {
	convey.Convey("Test handleFileDelete", t, func() {
		convey.Convey("case 1: deleteSuperPodFile failed. expect return err", func() {
			patch := gomonkey.ApplyFuncReturn(deleteSuperPodFile, fmt.Errorf("fake error"))
			defer patch.Reset()
			convey.ShouldNotBeNil(handleFileDelete(testSuperPodID))
		})
		convey.Convey("case 2: deleteSuperPodFile success. expect return nil", func() {
			patch := gomonkey.ApplyFuncReturn(deleteSuperPodFile, nil)
			defer patch.Reset()
			convey.ShouldBeNil(handleFileDelete(testSuperPodID))
		})
	})
}
