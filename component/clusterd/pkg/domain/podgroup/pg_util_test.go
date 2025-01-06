// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

//go:build !race

// Package podgroup a series of podgroup test function
package podgroup

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"volcano.sh/apis/pkg/apis/scheduling/v1beta1"

	"clusterd/pkg/common/constant"
)

func TestGetJobKeyByPG(t *testing.T) {
	convey.Convey("test GetJobKeyByPG", t, func() {
		pgDemo1 := getDemoPodGroup(pgName1, pgNameSpace, jobUid1)
		convey.Convey("when pg is nil, should return empty", func() {
			convey.So(GetJobKeyByPG(nil), convey.ShouldEqual, "")
		})
		convey.Convey("when pg is exists, owner is exists, should return jobId", func() {
			convey.So(GetJobKeyByPG(pgDemo1), convey.ShouldEqual, jobUid1)
		})
		convey.Convey("when pg is exists, owner is not exists, should return empty", func() {
			pgDemo1.OwnerReferences = []v1.OwnerReference{}
			convey.So(GetJobKeyByPG(pgDemo1), convey.ShouldEqual, "")
		})
	})
}

func TestGetJobNameByPG(t *testing.T) {
	convey.Convey("test GetJobNameByPG", t, func() {
		pgDemo1 := getDemoPodGroup(pgName1, pgNameSpace, jobUid1)
		convey.Convey("when pg is nil, should return empty", func() {
			convey.So(GetJobNameByPG(nil), convey.ShouldEqual, "")
		})
		convey.Convey("when pg is exists, owner is exists, should return jobName", func() {
			convey.So(GetJobNameByPG(pgDemo1), convey.ShouldEqual, pgName1)
		})
		convey.Convey("when pg is exists, owner is not exists, should return empty", func() {
			pgDemo1.OwnerReferences = []v1.OwnerReference{}
			convey.So(GetJobNameByPG(pgDemo1), convey.ShouldEqual, "")
		})
	})
}

func TestGetJobTypeByPG(t *testing.T) {
	convey.Convey("test GetJobTypeByPG", t, func() {
		pgDemo1 := getDemoPodGroup(pgName1, pgNameSpace, jobUid1)
		convey.Convey("when pg is nil, should return empty", func() {
			convey.So(GetJobTypeByPG(nil), convey.ShouldEqual, "")
		})
		convey.Convey("when pg is exists, owner is exists, should return job type", func() {
			convey.So(GetJobTypeByPG(pgDemo1), convey.ShouldEqual, vcJobKey)
		})
		convey.Convey("when pg is exists, owner is not exists, should return empty", func() {
			pgDemo1.OwnerReferences = []v1.OwnerReference{}
			convey.So(GetJobTypeByPG(pgDemo1), convey.ShouldEqual, "")
		})
	})
}

func TestGetModelFramework(t *testing.T) {
	convey.Convey("test GetModelFramework", t, func() {
		pgDemo1 := getDemoPodGroup(pgName1, pgNameSpace, jobUid1)
		convey.Convey("when pg is nil, should return empty", func() {
			convey.So(GetModelFramework(nil), convey.ShouldEqual, "")
		})
		convey.Convey("when pg is exists, label is exists, should return frameWork", func() {
			convey.So(GetModelFramework(pgDemo1), convey.ShouldEqual, ptFrameWork)
		})
		convey.Convey("when pg is exists, label is not exists, should return empty", func() {
			pgDemo1.Labels = map[string]string{}
			convey.So(GetModelFramework(pgDemo1), convey.ShouldEqual, "")
		})
	})
}

func TestJudgeUceByJobKey(t *testing.T) {
	convey.Convey("test JudgeUceByJobKey", t, func() {
		pgDemo1 := getDemoPodGroup(pgName1, pgNameSpace, jobUid1)
		convey.Convey("when pg is nil, should return false", func() {
			convey.So(JudgeUceByJobKey(jobUid1), convey.ShouldBeFalse)
		})
		convey.Convey("when pg is exists, process-recover-enable is not exists, should return false",
			func() {
				SavePodGroup(pgDemo1)
				defer DeletePodGroup(pgDemo1)
				convey.So(JudgeUceByJobKey(jobUid1), convey.ShouldBeFalse)
			},
		)
		convey.Convey("when pg is exists, process-recover-enable is exists, recover-strategy is not exists, "+
			"should return false",
			func() {
				pgDemo1.Labels[constant.ProcessRecoverEnableLabel] = constant.ProcessRecoverEnable
				SavePodGroup(pgDemo1)
				defer DeletePodGroup(pgDemo1)
				convey.So(JudgeUceByJobKey(jobUid1), convey.ShouldBeFalse)
			},
		)
		convey.Convey("when pg is exists, process-recover-enable is exists, "+
			"recover-strategy is exists, but not equals retry, should return false",
			func() {
				pgDemo1.Labels[constant.ProcessRecoverEnableLabel] = constant.ProcessRecoverEnable
				pgDemo1.Annotations[constant.RecoverStrategies] = ""
				SavePodGroup(pgDemo1)
				defer DeletePodGroup(pgDemo1)
				convey.So(JudgeUceByJobKey(jobUid1), convey.ShouldBeFalse)
			},
		)
		convey.Convey("when pg is exists, process-recover-enable is exists, "+
			"recover-strategy is exists, and equals retry, should return true",
			func() {
				pgDemo1.Labels[constant.ProcessRecoverEnableLabel] = constant.ProcessRecoverEnable
				pgDemo1.Annotations[constant.RecoverStrategies] = constant.ProcessRetryStrategyName
				SavePodGroup(pgDemo1)
				defer DeletePodGroup(pgDemo1)
				convey.So(JudgeUceByJobKey(jobUid1), convey.ShouldBeTrue)
			},
		)
	})
}

func TestJudgeIsRunningByJobKey(t *testing.T) {
	convey.Convey("test JudgeIsRunningByJobKey", t, func() {
		pgDemo1 := getDemoPodGroup(pgName1, pgNameSpace, jobUid1)
		convey.Convey("when pg is nil, should return false", func() {
			convey.So(JudgeIsRunningByJobKey(jobUid1), convey.ShouldBeFalse)
		})
		convey.Convey("when pg is exists but not running, should return false", func() {
			SavePodGroup(pgDemo1)
			defer DeletePodGroup(pgDemo1)
			convey.So(JudgeIsRunningByJobKey(jobUid1), convey.ShouldBeFalse)
		})
		convey.Convey("when pg is exists and running, should return true", func() {
			pgDemo1.Status.Phase = v1beta1.PodGroupRunning
			SavePodGroup(pgDemo1)
			defer DeletePodGroup(pgDemo1)
			convey.So(JudgeIsRunningByJobKey(jobUid1), convey.ShouldBeTrue)
		})
	})
}

func TestGetPGFromCacheOrPod(t *testing.T) {
	convey.Convey("test GetPGFromCacheOrPod, when pg is exists, should return job name", t, func() {
		pgDemo1 := getDemoPodGroup(pgName1, pgNameSpace, jobUid1)
		SavePodGroup(pgDemo1)
		defer DeletePodGroup(pgDemo1)
		jobName, pgName, namespace := GetPGFromCacheOrPod(jobUid1)
		convey.So(jobName, convey.ShouldEqual, pgName1)
		convey.So(pgName, convey.ShouldEqual, pgName1)
		convey.So(namespace, convey.ShouldEqual, pgNameSpace)
	})
}
