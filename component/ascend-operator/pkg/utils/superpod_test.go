/*
Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
*/

// Package utils is common utils
package utils

import (
	"strconv"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	commonv1 "github.com/kubeflow/common/pkg/apis/common/v1"
	"github.com/smartystreets/goconvey/convey"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"ascend-common/api"
	v1 "ascend-operator/pkg/api/v1"
)

const (
	nodeNumber2 = 2
	nodeNumber4 = 4
	npuNumber8  = 8
)

func TestGetLogicSuperPodNodes(t *testing.T) {
	convey.Convey("TestGetLogicSuperPodNodes", t, func() {
		convey.Convey("01-when spBlock is smaller than chipsPerNode, should return 1", func() {
			spBlock, chipsPerNode := 1, 2
			res := getLogicSuperPodNodes(spBlock, chipsPerNode)
			convey.So(res, convey.ShouldEqual, 1)
		})
		convey.Convey("02-when spBlock is bigger than chipsPerNode, should return quotient", func() {
			expected := 2
			spBlock, chipsPerNode := 4, 2
			res := getLogicSuperPodNodes(spBlock, chipsPerNode)
			convey.So(res, convey.ShouldEqual, expected)
		})
	})
}

func TestGetLogicSuperPodId(t *testing.T) {
	convey.Convey("TestGetLogicSuperPodId", t, func() {
		convey.Convey("01-when spBlock is 0, should return 0", func() {
			res := GetLogicSuperPodId(1, 0, 0)
			convey.So(res, convey.ShouldEqual, 0)
		})
		convey.Convey("02-when pod rank is 1, spBlock is 1 and chipsPerNode is 1, should return 1", func() {
			res := GetLogicSuperPodId(1, 1, 1)
			convey.So(res, convey.ShouldEqual, 1)
		})
	})
}

func TestGetSpBlock(t *testing.T) {
	convey.Convey("TestGetSpBlock", t, func() {
		convey.Convey("01-job is nil will return 0", func() {
			res := GetSpBlock(nil)
			convey.So(res, convey.ShouldEqual, 0)
		})
		convey.Convey("02-job without sp-block annotation should return 0", func() {
			res := GetSpBlock(newCommonAscendJob())
			convey.So(res, convey.ShouldEqual, 0)
		})
		convey.Convey("03-job with invalid sp-block annotation should return 0", func() {
			job := newCommonAscendJob()
			job.Annotations[AnnoKeyOfSuperPod] = "xx"
			res := GetSpBlock(newCommonAscendJob())
			convey.So(res, convey.ShouldEqual, 0)
		})
		convey.Convey("04-job with valid sp-block annotation should return 1", func() {
			job := newCommonAscendJob()
			job.Annotations[AnnoKeyOfSuperPod] = "1"
			res := GetSpBlock(job)
			convey.So(res, convey.ShouldEqual, 1)
		})
		convey.Convey("05-job with multilevel return 1", func() {
			job := newCommonAscendJob()
			patches := gomonkey.ApplyFuncReturn(IsMultiLevelJob, true).
				ApplyFuncReturn(getSpBlockFromAffinityConfig, 1)
			defer patches.Reset()
			res := GetSpBlock(job)
			convey.So(res, convey.ShouldEqual, 1)
		})
	})
}

func TestGetSpBlockNum(t *testing.T) {
	convey.Convey("TestGetSpBlockNum", t, func() {
		convey.Convey("01-job is nil will return 0", func() {
			res := GetSpBlockNum(nil)
			convey.So(res, convey.ShouldEqual, 0)
		})
		convey.Convey("02-job without annotations will return 0", func() {
			job := &v1.AscendJob{}
			res := GetSpBlockNum(job)
			convey.So(res, convey.ShouldEqual, 0)
		})
		convey.Convey("03-job with sp-block 0 will return 0", func() {
			job := newCommonAscendJob()
			job.Annotations[AnnoKeyOfSuperPod] = "0"
			res := GetSpBlockNum(job)
			convey.So(res, convey.ShouldEqual, 0)
		})
		convey.Convey("04-job with valid sp-block should return correct value", func() {
			replicas := int32(2)
			job := newCommonAscendJob()
			job.Annotations[AnnoKeyOfSuperPod] = "2"
			job.Spec.ReplicaSpecs = map[commonv1.ReplicaType]*commonv1.ReplicaSpec{
				"Worker": {
					Replicas: &replicas,
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Resources: corev1.ResourceRequirements{
										Requests: corev1.ResourceList{
											api.HuaweiAscend910: resource.MustParse("4"),
										},
									},
								},
							},
						},
					},
				},
			}
			const expectedReplicaNum = 4
			res := GetSpBlockNum(job)
			convey.So(res, convey.ShouldEqual, expectedReplicaNum) // 2 replicas * 4 devices / 2 spBlock = 4
		})
	})
}

func TestGetDevicesPerPod(t *testing.T) {
	convey.Convey("TestGetDevicesPerPod", t, func() {
		convey.Convey("01- get pod npu number correctly", func() {
			containers := []corev1.Container{
				{Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{api.HuaweiAscend910: resource.MustParse(strconv.Itoa(npuNumber8))},
				}},
			}
			res := getDevicesPerPod(containers)
			convey.So(res, convey.ShouldEqual, npuNumber8)
		})
		convey.Convey("02- get 0 when pod has no npu", func() {
			containers := []corev1.Container{{}}
			res := getDevicesPerPod(containers)
			convey.So(res, convey.ShouldEqual, 0)
		})
	})
}

func TestGetSpBlockFromAffinityConfig(t *testing.T) {
	convey.Convey("TestGetSpBlockFromAffinityConfig", t, func() {
		convey.Convey("01-get sp-block number correctly", func() {
			patches := gomonkey.ApplyFuncReturn(getAffinityBlocks, map[string]int{Level1BlockKey: nodeNumber2}).
				ApplyFuncReturn(getDevicesPerPod, npuNumber8)
			defer patches.Reset()
			job := &v1.AscendJob{}
			job.Spec.ReplicaSpecs = map[commonv1.ReplicaType]*commonv1.ReplicaSpec{"Worker": {}}
			res := getSpBlockFromAffinityConfig(job)
			convey.So(res, convey.ShouldEqual, npuNumber8*nodeNumber2)
		})
		convey.Convey("02-when job is nil return 0 ", func() {
			res := getSpBlockFromAffinityConfig(nil)
			convey.So(res, convey.ShouldEqual, 0)
		})
		convey.Convey("03-when affinity block is nil return 0 ", func() {
			patch := gomonkey.ApplyFuncReturn(getAffinityBlocks, nil)
			defer patch.Reset()
			job := &v1.AscendJob{}
			res := getSpBlockFromAffinityConfig(job)
			convey.So(res, convey.ShouldEqual, 0)
		})
		convey.Convey("04-when level config is nil return 0 ", func() {
			patch := gomonkey.ApplyFuncReturn(getAffinityBlocks, map[string]int{})
			defer patch.Reset()
			job := &v1.AscendJob{}
			res := getSpBlockFromAffinityConfig(job)
			convey.So(res, convey.ShouldEqual, 0)
		})
	})
}

func TestGetAffinityBlocks(t *testing.T) {
	const level1BlockKey = "level2"
	convey.Convey("TestGetAffinityBlocks", t, func() {
		job := &v1.AscendJob{}
		convey.Convey("01-get sp-block number correctly", func() {
			job.Annotations = map[string]string{api.AffinityConfigAnnoKey: "level1=2,level2=4"}
			res := getAffinityBlocks(job)
			convey.So(res, convey.ShouldEqual, map[string]int{Level1BlockKey: nodeNumber2, level1BlockKey: nodeNumber4})
		})
		convey.Convey("02-job with empty annotation will return nil", func() {
			job.Annotations = map[string]string{}
			res := getAffinityBlocks(job)
			convey.So(res, convey.ShouldEqual, nil)
		})
		convey.Convey("03-job with invalid config will return nil", func() {
			job.Annotations = map[string]string{api.AffinityConfigAnnoKey: "level1"}
			res := getAffinityBlocks(job)
			convey.So(res, convey.ShouldEqual, nil)
		})
		convey.Convey("04-job with invalid config will return nil", func() {
			job.Annotations = map[string]string{api.AffinityConfigAnnoKey: "level1=a"}
			res := getAffinityBlocks(job)
			convey.So(res, convey.ShouldEqual, nil)
		})
	})
}
