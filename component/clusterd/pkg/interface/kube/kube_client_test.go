// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package kube a series of kube test function
package kube

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
)

const (
	testCMName  = "test-cm-name"
	testPodName = "test-pod-name"
	testNS      = "default"
	testKey1    = "key1"
	testKey2    = "key2"
	testKey3    = "key3"
	testValue1  = "value1"
	testValue2  = "value2"
	testValue3  = "value3"
	testName    = "test-name"
	testJobType = "test-job-type"
)

var (
	cm  *v1.ConfigMap
	pod *v1.Pod
)

func TestPatchCMData(t *testing.T) {
	createCM(t)
	defer func() {
		if err := DeleteConfigMap(testCMName, testNS); err != nil {
			t.Error(err)
		}
	}()
	convey.Convey("test func 'PatchCMData' success", t, func() {

		// data from null to 'key1:value1, key2:value2'
		patchData := map[string]string{testKey1: testValue1, testKey2: testValue2}
		newCM, err := PatchCMData(testCMName, testNS, patchData)
		expData := map[string]string{testKey1: testValue1, testKey2: testValue2}
		convey.So(err, convey.ShouldBeNil)
		convey.So(newCM.Data, convey.ShouldResemble, expData)

		// data from 'key1:value1, key2:value2' to 'key1:value1, key2:value1, key3:value3'
		patchData = map[string]string{testKey2: testValue1, testKey3: testValue3}
		newCM, err = PatchCMData(testCMName, testNS, patchData)
		expData = map[string]string{testKey1: testValue1, testKey2: testValue1, testKey3: testValue3}
		convey.So(err, convey.ShouldBeNil)
		convey.So(newCM.Data, convey.ShouldResemble, expData)
	})

	convey.Convey("test func 'PatchCMData' failed, marshal data failed", t, func() {
		p1 := gomonkey.ApplyFuncReturn(json.Marshal, nil, testErr)
		defer p1.Reset()
		patchData := map[string]string{testKey1: testValue1}
		_, err := PatchCMData(testCMName, testNS, patchData)
		expErr := fmt.Errorf("marshal cm data failed")
		convey.So(err, convey.ShouldResemble, expErr)
	})
}

func TestCheckNodeExist(t *testing.T) {
	convey.Convey("Test checkNodeExist function", t, func() {
		convey.Convey("When JSON unmarshal fails, should return false and log error", func() {
			// Mock the error log
			patch := gomonkey.ApplyFunc(hwlog.RunLog.Errorf, func(format string, args ...interface{}) {
				// Verify that the error message contains the expected text
				convey.So(format, convey.ShouldContainSubstring, "unmarshal fault info failed")
			})
			defer patch.Reset()

			result := checkNodeExist("invalidJsonData", "node1")
			convey.So(result, convey.ShouldBeFalse)
		})
		convey.Convey("When node exists in fault info, should return true", func() {
			// Prepare valid JSON data with node in FaultNodes
			faultInfos := map[int]api.SuperPodFaultInfos{0: {FaultNodes: sets.NewString("node1", "node2")}}
			jsonData, err := json.Marshal(faultInfos)
			if err != nil {
				fmt.Printf("marshal fault info failed, error: %v", err)
			}
			result := checkNodeExist(string(jsonData), "node1")
			convey.So(result, convey.ShouldBeTrue)
		})
		convey.Convey("When node does not exist in fault info, should return false and log info", func() {
			// Mock the info log
			patch := gomonkey.ApplyFunc(hwlog.RunLog.Infof, func(format string, args ...interface{}) {
				// Verify that the info message contains the expected text
				convey.So(format, convey.ShouldContainSubstring, "node %s not in configmap fault-job-info")
			})
			defer patch.Reset()
			// Prepare valid JSON data with node not in FaultNodes
			faultInfos := map[int]api.SuperPodFaultInfos{0: {FaultNodes: sets.NewString("node1", "node2")}}
			jsonData, err := json.Marshal(faultInfos)
			if err != nil {
				fmt.Printf("marshal fault info failed, error: %v", err)
			}
			result := checkNodeExist(string(jsonData), "node3")
			convey.So(result, convey.ShouldBeFalse)
		})
	})
}

func createCM(t *testing.T) {
	cm = &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      testCMName,
			Namespace: testNS,
		},
		Data: nil,
	}
	_, err := CreateConfigMap(cm)
	if err != nil {
		t.Errorf("create test cm failed, error: %v", err)
	}
}

func TestCreateCM(t *testing.T) {
	cm = &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      testCMName,
			Namespace: testNS,
		},
		Data: nil,
	}
	defer func() {
		if err := DeleteConfigMap(testCMName, testNS); err != nil {
			t.Error(err)
		}
	}()
	convey.Convey("test func 'CreateConfigMap' success", t, func() {
		resCM, err := CreateConfigMap(cm)
		convey.So(err, convey.ShouldBeNil)
		convey.So(resCM, convey.ShouldResemble, cm)
	})
	convey.Convey("test func 'CreateConfigMap' failed, input cm is nil", t, func() {
		_, err := CreateConfigMap(nil)
		expErr := fmt.Errorf("param cm is nil")
		convey.So(err, convey.ShouldResemble, expErr)
	})
}

func TestUpdateCM(t *testing.T) {
	createCM(t)
	defer func() {
		if err := DeleteConfigMap(testCMName, testNS); err != nil {
			t.Error(err)
		}
	}()
	convey.Convey("test func 'UpdateConfigMap' success", t, func() {
		resCM, err := UpdateConfigMap(cm)
		convey.So(err, convey.ShouldBeNil)
		convey.So(resCM, convey.ShouldResemble, cm)
	})
	convey.Convey("test func 'UpdateConfigMap' failed, input cm is nil", t, func() {
		_, err := UpdateConfigMap(nil)
		expErr := fmt.Errorf("param cm is nil")
		convey.So(err, convey.ShouldResemble, expErr)
	})
}

func TestGetCM(t *testing.T) {
	createCM(t)
	defer func() {
		if err := DeleteConfigMap(testCMName, testNS); err != nil {
			t.Error(err)
		}
	}()
	convey.Convey("test func 'UpdateConfigMap' success", t, func() {
		resCM, err := GetConfigMap(testCMName, testNS)
		convey.So(err, convey.ShouldBeNil)
		convey.So(resCM, convey.ShouldResemble, cm)
	})
}

func TestDeleteCM(t *testing.T) {
	createCM(t)
	convey.Convey("test func 'DeleteConfigMap' success", t, func() {
		err := DeleteConfigMap(testCMName, testNS)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestCreateOrUpdateCM(t *testing.T) {
	convey.Convey("test func 'CreateOrUpdateConfigMap' success. cm does not exist, create success", t, func() {
		DeleteConfigMap(testCMName, testNS)
		err := CreateOrUpdateConfigMap(testCMName, testNS, nil, nil)
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("test func 'CreateOrUpdateConfigMap' failed. cm does not exist, create error", t, func() {
		DeleteConfigMap(testCMName, testNS)
		p1 := gomonkey.ApplyMethodReturn(GetClientK8s().ClientSet.CoreV1().ConfigMaps(v1.NamespaceAll),
			"Create", &v1.ConfigMap{}, testErr)
		defer p1.Reset()
		err := CreateOrUpdateConfigMap(testCMName, testNS, nil, nil)
		expErr := fmt.Errorf("unable to create ConfigMap: %v", testErr)
		convey.So(err, convey.ShouldResemble, expErr)
	})
	convey.Convey("test func 'CreateOrUpdateConfigMap' success. cm existed, update success", t, func() {
		DeleteConfigMap(testCMName, testNS)
		createCM(t)
		err := CreateOrUpdateConfigMap(testCMName, testNS, nil, nil)
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("test func 'CreateOrUpdateConfigMap' failed. cm existed, update error", t, func() {
		DeleteConfigMap(testCMName, testNS)
		createCM(t)
		// first create, second update: configmaps "test-cm-name" already exists
		p1 := gomonkey.ApplyMethodReturn(GetClientK8s().ClientSet.CoreV1().ConfigMaps(v1.NamespaceAll),
			"Update", nil, testErr)
		defer p1.Reset()
		err := CreateOrUpdateConfigMap(testCMName, testNS, nil, nil)
		convey.So(err, convey.ShouldResemble, testErr)
	})
}

func TestUpdateOrCreateCM(t *testing.T) {
	convey.Convey("test func 'UpdateOrCreateConfigMap' success. cm existed, update success", t, func() {
		DeleteConfigMap(testCMName, testNS)
		createCM(t)
		err := UpdateOrCreateConfigMap(testCMName, testNS, nil, nil)
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("test func 'UpdateOrCreateConfigMap' failed. cm existed, update error", t, func() {
		DeleteConfigMap(testCMName, testNS)
		createCM(t)
		p1 := gomonkey.ApplyMethodReturn(GetClientK8s().ClientSet.CoreV1().ConfigMaps(v1.NamespaceAll),
			"Update", nil, testErr)
		defer p1.Reset()
		err := UpdateOrCreateConfigMap(testCMName, testNS, nil, nil)
		expErr := fmt.Errorf("unable to update ConfigMap: %v", testErr)
		convey.So(err, convey.ShouldResemble, expErr)
	})
	convey.Convey("test func 'UpdateOrCreateConfigMap' success. cm does not exist, create success", t, func() {
		DeleteConfigMap(testCMName, testNS)
		err := UpdateOrCreateConfigMap(testCMName, testNS, nil, nil)
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("test func 'UpdateOrCreateConfigMap' failed. cm does not exist, create error", t, func() {
		DeleteConfigMap(testCMName, testNS)
		p1 := gomonkey.ApplyMethodReturn(GetClientK8s().ClientSet.CoreV1().ConfigMaps(v1.NamespaceAll),
			"Create", &v1.ConfigMap{}, testErr)
		defer p1.Reset()
		err := UpdateOrCreateConfigMap(testCMName, testNS, nil, nil)
		convey.So(err, convey.ShouldResemble, testErr)
	})
}

func createPod(t *testing.T) {
	pod = &v1.Pod{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      testPodName,
			Namespace: testNS,
		},
		Spec:   v1.PodSpec{},
		Status: v1.PodStatus{},
	}
	_, err := GetClientK8s().ClientSet.CoreV1().Pods(testNS).Create(context.TODO(), pod, metav1.CreateOptions{})
	if err != nil {
		t.Errorf("create test pod failed, error: %v", err)
	}
}

func deletePod(t *testing.T) {
	if err := GetClientK8s().ClientSet.CoreV1().Pods(testNS).Delete(context.TODO(),
		testPodName, metav1.DeleteOptions{}); err != nil {
		t.Error(err)
	}
}

func TestPatchPodLabel(t *testing.T) {
	createPod(t)
	defer deletePod(t)
	convey.Convey("test func 'PatchPodLabel' success", t, func() {

		// label from null to 'key1:value1, key2:value2'
		patchData := map[string]string{testKey1: testValue1, testKey2: testValue2}
		newPod, err := PatchPodLabel(testPodName, testNS, patchData)
		expData := map[string]string{testKey1: testValue1, testKey2: testValue2}
		convey.So(err, convey.ShouldBeNil)
		convey.So(newPod.Labels, convey.ShouldResemble, expData)

		// label from 'key1:value1, key2:value2' to 'key1:value1, key2:value1, key3:value3'
		patchData = map[string]string{testKey2: testValue1, testKey3: testValue3}
		newPod, err = PatchPodLabel(testPodName, testNS, patchData)
		expData = map[string]string{testKey1: testValue1, testKey2: testValue1, testKey3: testValue3}
		convey.So(err, convey.ShouldBeNil)
		convey.So(newPod.Labels, convey.ShouldResemble, expData)
	})

	convey.Convey("test func 'PatchPodLabel' failed, marshal data failed", t, func() {
		p1 := gomonkey.ApplyFuncReturn(json.Marshal, nil, testErr)
		defer p1.Reset()
		patchData := map[string]string{testKey1: testValue1}
		_, err := PatchPodLabel(testPodName, testNS, patchData)
		convey.So(err, convey.ShouldResemble, testErr)
	})
}

func TestRetryPatchPodLabels(t *testing.T) {
	convey.Convey("test func 'RetryPatchPodLabels' success", t, func() {
		p1 := gomonkey.ApplyFuncReturn(PatchPodLabel, pod, nil)
		defer p1.Reset()
		err := RetryPatchPodLabels(testPodName, testNS, 1, nil)
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("test func 'RetryPatchPodLabels' failed, PatchPodLabel error", t, func() {
		p1 := gomonkey.ApplyFuncReturn(PatchPodLabel, nil, testErr)
		defer p1.Reset()
		err := RetryPatchPodLabels(testPodName, testNS, 1, nil)
		convey.So(err, convey.ShouldResemble, testErr)
	})
}

func TestGetJobEvent(t *testing.T) {
	convey.Convey("get event error", t, func() {
		p1 := gomonkey.ApplyMethodReturn(GetClientK8s().ClientSet.CoreV1().Events(v1.NamespaceAll),
			"List", nil, testErr)
		defer p1.Reset()
		_, err := GetJobEvent(testNS, testName, testJobType)
		convey.So(err, convey.ShouldResemble, testErr)
	})
	convey.Convey("get event ok", t, func() {
		p1 := gomonkey.ApplyMethodReturn(GetClientK8s().ClientSet.CoreV1().Events(v1.NamespaceAll),
			"List", nil, nil)
		defer p1.Reset()
		_, err := GetJobEvent(testNS, testName, testJobType)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestUpdateFaultJobInfoCmWhenJobDelete(t *testing.T) {
	convey.Convey("test func 'UpdateFaultJobInfoCmWhenJobDelete' success", t, func() {
		testJobId := "test-job-id"
		testCm := &v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      api.FaultJobCmName,
				Namespace: api.ClusterNS,
			},
			Data: map[string]string{
				testJobId: "test-data",
			},
		}
		p1 := gomonkey.ApplyFuncReturn(GetConfigMap, testCm, nil)
		defer p1.Reset()
		p2 := gomonkey.ApplyFuncReturn(UpdateConfigMap, testCm, nil)
		defer p2.Reset()
		RecoverFaultJobInfoCm(testJobId, "")
		_, exists := testCm.Data[testJobId]
		convey.So(exists, convey.ShouldBeFalse)
	})

	convey.Convey("test func 'UpdateFaultJobInfoCmWhenJobDelete' with GetConfigMap error", t, func() {
		p1 := gomonkey.ApplyFuncReturn(GetConfigMap, nil, testErr)
		defer p1.Reset()
		RecoverFaultJobInfoCm("test-job-id", "")
	})

	convey.Convey("test func 'UpdateFaultJobInfoCmWhenJobDelete' with UpdateConfigMap error", t, func() {
		testJobId := "test-job-id"
		testCm := &v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      api.FaultJobCmName,
				Namespace: api.ClusterNS,
			},
			Data: map[string]string{
				testJobId: "test-data",
			},
		}
		p1 := gomonkey.ApplyFuncReturn(GetConfigMap, testCm, nil)
		defer p1.Reset()

		p2 := gomonkey.ApplyFuncReturn(UpdateConfigMap, nil, testErr)
		defer p2.Reset()

		RecoverFaultJobInfoCm(testJobId, "")
	})
}

func TestCreateOrUpdateSuperPodFaultInfo(t *testing.T) {
	const fakeTime = 123456789
	convey.Convey("Test CreateOrUpdateSuperPodFaultInfo", t, func() {
		testJobId := "test-job-id"
		testFaultInfos := map[int]api.SuperPodFaultInfos{
			1: {FaultTimes: fakeTime},
		}
		convey.Convey("Should update existing configmap failed", func() {
			testCm := &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      api.FaultJobCmName,
					Namespace: api.ClusterNS,
				},
				Data: nil,
			}
			patch1 := gomonkey.ApplyFuncReturn(GetConfigMap, testCm, nil)
			defer patch1.Reset()
			patch2 := gomonkey.ApplyFuncReturn(UpdateConfigMap, testCm, errors.New("fake error"))
			defer patch2.Reset()
			patch3 := gomonkey.ApplyFuncReturn(json.Marshal, []byte(`{}`), nil)
			defer patch3.Reset()
			CreateOrUpdateSuperPodFaultInfo(testJobId, testFaultInfos)
			_, exists := testCm.Data[testJobId]
			convey.So(exists, convey.ShouldBeFalse)
		})
		convey.Convey("Should handle GetConfigMap error", func() {
			patch1 := gomonkey.ApplyFuncReturn(GetConfigMap, nil, testErr)
			defer patch1.Reset()
			patch2 := gomonkey.ApplyFunc(hwlog.RunLog.Errorf, func(format string, args ...interface{}) {
				convey.So(format, convey.ShouldContainSubstring, "get configmap fault-job-info err")
			})
			defer patch2.Reset()

			CreateOrUpdateSuperPodFaultInfo(testJobId, testFaultInfos)
		})
	})
}
