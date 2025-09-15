// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

//go:build !race

// Package podgroup a series of podgroup test function
package podgroup

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"volcano.sh/apis/pkg/apis/scheduling/v1beta1"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
)

func init() {
	hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, context.Background())
}

const (
	pgName1      = "job1"
	pgNameSpace  = "default"
	jobUid1      = "123"
	vcJobKey     = "job"
	ptFrameWork  = "pytorch"
	ResourceName = api.ResourceNamePrefix + api.Ascend910
)

func TestSavePodGroup(t *testing.T) {
	convey.Convey("test SavePodGroup, when pg info is nil", t, func() {
		oldLen := len(pgManager.pgMap)
		SavePodGroup(nil)
		newLen := len(pgManager.pgMap)
		convey.So(oldLen, convey.ShouldEqual, newLen)
	})
	convey.Convey("test SavePodGroup, when pg cache is exists, should get podGroup", t, func() {
		pgDemo1 := getDemoPodGroup(pgName1, pgNameSpace, jobUid1)
		SavePodGroup(pgDemo1)
		defer DeletePodGroup(pgDemo1)
		pgInfo := GetPodGroup(jobUid1)
		convey.So(pgInfo.Name, convey.ShouldEqual, pgName1)
	})
}

func getDemoPodGroup(pgName, nameSpace, jobUid string) *v1beta1.PodGroup {
	podGroupInfo := &v1beta1.PodGroup{}
	podGroupInfo.Name = pgName
	podGroupInfo.Namespace = nameSpace
	isControl := true
	owner := v1.OwnerReference{
		Name:       pgName,
		Controller: &isControl,
		Kind:       vcJobKey,
		UID:        types.UID(jobUid)}
	podGroupInfo.SetOwnerReferences([]v1.OwnerReference{owner})
	labelMap := map[string]string{frameWorkKey: ptFrameWork}
	podGroupInfo.Labels = labelMap
	podGroupInfo.Annotations = make(map[string]string)
	podGroupInfo.Spec.MinResources = &corev1.ResourceList{
		ResourceName: {},
	}
	return podGroupInfo
}

func TestCheckPodGroupExist(t *testing.T) {
	convey.Convey("Test CheckPodGroupExist function", t, func() {
		convey.Convey("When PodGroup exists, should return true", func() {
			pg := getDemoPodGroup("test-pg", "test-namespace", "test-job-key")
			SavePodGroup(pg)
			defer DeletePodGroup(pg)

			exists := CheckPodGroupExist(GetJobKeyByPG(pg))
			convey.So(exists, convey.ShouldBeTrue)
		})
	})
}
