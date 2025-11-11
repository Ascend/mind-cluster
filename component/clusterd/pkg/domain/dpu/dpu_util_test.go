// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

// Package dpu a series of dpu test function
package dpu

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"ascend-common/api"
	"clusterd/pkg/common/constant"
)

const (
	length1 = 1
	length2 = 2
)

var (
	mockNPUToDpusMap = map[string][]string{
		"0": {"enps0", "enps2"},
		"7": {"enps1", "enps3"},
	}
	mockDpuList = constant.DpuCMDataList{
		{Name: "dpu1"},
		{Name: "dpu2"},
	}
)

func generateParseDpuInfoSuccessTestCM() *v1.ConfigMap {
	dataBytesList, err := json.Marshal(mockDpuList)
	if err != nil {
		return nil
	}
	dataBytesMap, err := json.Marshal(mockNPUToDpusMap)
	if err != nil {
		return nil
	}
	cm := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-cm",
		},
		Data: map[string]string{
			api.DpuInfoCMBusTypeKey:      "ub",
			api.DpuInfoCMDataKey:         string(dataBytesList),
			api.DpuInfoCMNpuToDpusMapKey: string(dataBytesMap),
		},
	}
	return cm
}

func TestParseDpuInfoCM(t *testing.T) {
	convey.Convey("testParseDpuInfoCM", t, func() {
		convey.Convey("Test parse dpu info success", func() {
			cm := generateParseDpuInfoSuccessTestCM()
			result, err := ParseDpuInfoCM(cm)
			convey.So(err, convey.ShouldBeNil)
			convey.So(result.CmName, convey.ShouldEqual, "test-cm")
			convey.So(result.DPUList, convey.ShouldResemble, mockDpuList)
			convey.So(result.NpuToDpusMap, convey.ShouldResemble, mockNPUToDpusMap)
		})

		convey.Convey("Test parse dpu info errors", func() {
			testParseDpuInfoCMInputNotConfigMap()
			testParseDpuInfoCMMissingDataKey()
			testParseDpuInfoCMInvalidDataJSON()
			testParseDpuInfoCMMissingBusTypeKey()
			testParseDpuInfoCMMissingNpuToDpusMapKey()
			testParseDpuInfoCMInvalidNpuToDpusMapJSON()
		})
	})
}

func TestDeepCopy(t *testing.T) {
	convey.Convey("TestDeepCopy", t, func() {
		convey.Convey("DeepCopy returns nil when input is nil", func() {
			result := DeepCopy(nil)
			convey.So(result, convey.ShouldBeNil)
		})

		convey.Convey("DeepCopy returns a deep copy of DpuInfoCM", func() {
			original := &constant.DpuInfoCM{
				CmName: "cm1",
				DPUList: constant.DpuCMDataList{
					{Name: "dpu1"},
					{Name: "dpu2"},
				},
			}
			newCopy := DeepCopy(original)
			convey.So(newCopy, convey.ShouldNotBeNil)
			convey.So(newCopy, convey.ShouldResemble, original)
			// Modify the copy and ensure original is not affected
			newCopy.CmName = "cm2"
			newCopy.DPUList[0].Name = "dpu3"
			convey.So(original.CmName, convey.ShouldEqual, "cm1")
			convey.So(original.DPUList[0].Name, convey.ShouldEqual, "dpu1")
		})
	})
}

func TestGetSafeData(t *testing.T) {
	convey.Convey("TestGetSafeData", t, func() {
		convey.Convey("Empty input returns empty slice", func() {
			result := GetSafeData(nil)
			convey.So(result, convey.ShouldResemble, []string{})
			result = GetSafeData(map[string]*constant.DpuInfoCM{})
			convey.So(result, convey.ShouldResemble, []string{})
		})

		convey.Convey("Input size <= safeDpuCMSize returns single string", func() {
			const mockDpuSize = 10
			dpuCMInfos := generateDpuCMInfos(mockDpuSize)
			result := GetSafeData(dpuCMInfos)
			convey.So(len(result), convey.ShouldEqual, length1)
			// Should contain all keys
			for i := 0; i < 10; i++ {
				convey.So(result[0], convey.ShouldContainSubstring, fmt.Sprintf("cm-%d", i))
			}
		})

		convey.Convey("Input size > safeDpuCMSize splits into multiple strings", func() {
			origSafeDpuCMSize := safeDpuCMSize
			// Prepare 1001 entries
			dpuCMInfos := generateDpuCMInfos(origSafeDpuCMSize + 1)
			result := GetSafeData(dpuCMInfos)
			convey.So(len(result), convey.ShouldEqual, length2)

			// First chunk: safeDpuCMSize, second chunk: 1
			convey.So(strings.Count(result[0], "CmName"), convey.ShouldEqual, origSafeDpuCMSize)
			convey.So(strings.Count(result[1], "CmName"), convey.ShouldEqual, length1)
		})

		convey.Convey("Input size exactly safeDpuCMSize returns one chunk", func() {
			dpuCMInfos := generateDpuCMInfos(safeDpuCMSize)
			result := GetSafeData(dpuCMInfos)
			convey.So(len(result), convey.ShouldEqual, length1)
			for i := 0; i < safeDpuCMSize; i++ {
				convey.So(result[0], convey.ShouldContainSubstring, fmt.Sprintf("cm-%d", i))
			}
		})
	})
}

func testParseDpuInfoCMInputNotConfigMap() {
	convey.Convey("Input is not a ConfigMap", func() {
		const mockInput = "123"
		result, err := ParseDpuInfoCM(mockInput)
		convey.So(result, convey.ShouldBeNil)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func testParseDpuInfoCMMissingDataKey() {
	convey.Convey("ConfigMap missing DpuInfoCMDataKey", func() {
		cm := &v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Name: "cm-missing-data"},
			Data: map[string]string{
				api.DpuInfoCMBusTypeKey:      "ub",
				api.DpuInfoCMNpuToDpusMapKey: "{}",
			},
		}
		result, err := ParseDpuInfoCM(cm)
		convey.So(result, convey.ShouldBeNil)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, api.DpuInfoCMDataKey)
	})
}

func testParseDpuInfoCMInvalidDataJSON() {
	convey.Convey("ConfigMap DpuInfoCMDataKey has invalid JSON", func() {
		cm := &v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Name: "cm-bad-json"},
			Data: map[string]string{
				api.DpuInfoCMDataKey:         "{bad json}",
				api.DpuInfoCMBusTypeKey:      "ub",
				api.DpuInfoCMNpuToDpusMapKey: "{}",
			},
		}
		result, err := ParseDpuInfoCM(cm)
		convey.So(result, convey.ShouldBeNil)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "unmarshal data failed")
	})
}

func testParseDpuInfoCMMissingBusTypeKey() {
	convey.Convey("ConfigMap missing DpuInfoCMBusTypeKey", func() {
		cm := &v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Name: "cm-missing-bus"},
			Data: map[string]string{
				api.DpuInfoCMDataKey:         "[]",
				api.DpuInfoCMNpuToDpusMapKey: "{}",
			},
		}
		result, err := ParseDpuInfoCM(cm)
		convey.So(result, convey.ShouldBeNil)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, api.DpuInfoCMBusTypeKey)
	})
}

func testParseDpuInfoCMMissingNpuToDpusMapKey() {
	convey.Convey("ConfigMap missing DpuInfoCMNpuToDpusMapKey", func() {
		cm := &v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Name: "cm-missing-map"},
			Data: map[string]string{
				api.DpuInfoCMDataKey:    "[]",
				api.DpuInfoCMBusTypeKey: "ub",
			},
		}
		result, err := ParseDpuInfoCM(cm)
		convey.So(result, convey.ShouldBeNil)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, api.DpuInfoCMNpuToDpusMapKey)
	})
}

func testParseDpuInfoCMInvalidNpuToDpusMapJSON() {
	convey.Convey("ConfigMap DpuInfoCMNpuToDpusMapKey has invalid JSON", func() {
		cm := &v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Name: "cm-bad-map-json"},
			Data: map[string]string{
				api.DpuInfoCMDataKey:         "[]",
				api.DpuInfoCMBusTypeKey:      "ub",
				api.DpuInfoCMNpuToDpusMapKey: "{bad json}",
			},
		}
		result, err := ParseDpuInfoCM(cm)
		convey.So(result, convey.ShouldBeNil)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "unmarshal data failed")
	})
}

func generateDpuCMInfos(size int) map[string]*constant.DpuInfoCM {
	dpuCMInfos := make(map[string]*constant.DpuInfoCM)
	for i := 0; i < size; i++ {
		name := fmt.Sprintf("cm-%d", i)
		dpuCMInfos[name] = &constant.DpuInfoCM{
			CmName: name,
			DPUList: constant.DpuCMDataList{
				{Name: fmt.Sprintf("dpu-%d", i)},
			},
		}
	}
	return dpuCMInfos
}
