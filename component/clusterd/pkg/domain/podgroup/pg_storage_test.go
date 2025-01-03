// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

//go:build !race

// Package podgroup a series of podgroup test function
package podgroup

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"volcano.sh/apis/pkg/apis/scheduling/v1beta1"

	"ascend-common/common-utils/hwlog"
)

func init() {
	hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, context.Background())
}

const (
	pgName1     = "job1"
	pgNameSpace = "default"
	jobUid1     = "123"
	vcJobKey    = "job"
	ptFrameWork = "pytorch"
)

func TestSavePodGroup(t *testing.T) {
	convey.Convey("test SavePodGroup, when pg cache is exists, should get podGroup ", t, func() {
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
	isControlle := true
	owner := metav1.OwnerReference{
		Name:       pgName,
		Controller: &isControlle,
		Kind:       vcJobKey,
		UID:        types.UID(jobUid)}
	podGroupInfo.SetOwnerReferences([]metav1.OwnerReference{owner})
	labelMap := map[string]string{frameWorkKey: ptFrameWork}
	podGroupInfo.Labels = labelMap
	podGroupInfo.Annotations = make(map[string]string)
	return podGroupInfo
}
