// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

//go:build !race

// Package podgroup a series of podgroup test function
package podgroup

import (
	"errors"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"volcano.sh/apis/pkg/apis/scheduling/v1beta1"

	"ascend-common/api"
	"clusterd/pkg/common/constant"
)

const (
	testJobId        = "test-job-id"
	testFramework    = "pytorch"
	emptyFramework   = ""
	testPodGroupName = "test-pg"
	firstRetryIndex  = 0
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
	convey.Convey("test JudgeRetryByJobKey", t, func() {
		pgDemo1 := getDemoPodGroup(pgName1, pgNameSpace, jobUid1)
		convey.Convey("when pg is nil, should return false", func() {
			convey.So(JudgeRetryByJobKey(jobUid1), convey.ShouldBeFalse)
		})
		convey.Convey("when pg is exists, process-recover-enable is not exists, should return false",
			func() {
				SavePodGroup(pgDemo1)
				defer DeletePodGroup(pgDemo1)
				convey.So(JudgeRetryByJobKey(jobUid1), convey.ShouldBeFalse)
			},
		)
		convey.Convey("when pg is exists, process-recover-enable is exists, recover-strategy is not exists, "+
			"should return false",
			func() {
				pgDemo1.Labels[constant.ProcessRecoverEnableLabel] = constant.ProcessRecoverEnable
				SavePodGroup(pgDemo1)
				defer DeletePodGroup(pgDemo1)
				convey.So(JudgeRetryByJobKey(jobUid1), convey.ShouldBeFalse)
			},
		)
		convey.Convey("when pg is exists, process-recover-enable is exists, "+
			"recover-strategy is exists, but not equals retry, should return false",
			func() {
				pgDemo1.Labels[constant.ProcessRecoverEnableLabel] = constant.ProcessRecoverEnable
				pgDemo1.Annotations[constant.RecoverStrategies] = ""
				SavePodGroup(pgDemo1)
				defer DeletePodGroup(pgDemo1)
				convey.So(JudgeRetryByJobKey(jobUid1), convey.ShouldBeFalse)
			},
		)
		convey.Convey("when pg is exists, process-recover-enable is exists, "+
			"recover-strategy is exists, and equals retry, should return true",
			func() {
				pgDemo1.Labels[constant.ProcessRecoverEnableLabel] = constant.ProcessRecoverEnable
				pgDemo1.Annotations[constant.RecoverStrategies] = constant.ProcessRetryStrategyName
				SavePodGroup(pgDemo1)
				defer DeletePodGroup(pgDemo1)
				convey.So(JudgeRetryByJobKey(jobUid1), convey.ShouldBeTrue)
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

func TestGetResourceType(t *testing.T) {
	convey.Convey("test TestGetResourceType success", t, func() {
		pgDemo1 := getDemoPodGroup(pgName1, pgNameSpace, jobUid1)
		resourceType := GetResourceType(pgDemo1)
		convey.So(resourceType, convey.ShouldEqual, api.Ascend910)
	})
	convey.Convey("test TestGetResourceType fail", t, func() {
		pgDemo1 := getDemoPodGroup(pgName1, pgNameSpace, jobUid1)
		pgDemo1.Spec.MinResources = nil
		resourceType := GetResourceType(pgDemo1)
		convey.So(resourceType, convey.ShouldEqual, "")
	})
}

func TestJudgeRestartProcessByJobKey(t *testing.T) {
	convey.Convey("test JudgeRestartProcessByJobKey", t, func() {
		pgDemo1 := getDemoPodGroup(pgName1, pgNameSpace, jobUid1)
		convey.Convey("when pg is exists, process-recover-enable is exists, "+
			"recover-strategy is exists, and equals recover-in-place, should return true",
			func() {
				pgDemo1.Labels[constant.ProcessRecoverEnableLabel] = constant.ProcessRecoverEnable
				pgDemo1.Annotations[constant.RecoverStrategies] = constant.ProcessRecoverInPlaceStrategyName
				SavePodGroup(pgDemo1)
				defer DeletePodGroup(pgDemo1)
				convey.So(JudgeRestartProcessByJobKey(jobUid1), convey.ShouldBeTrue)
			},
		)
	})
}

func TestGetOwnerRefByPG(t *testing.T) {
	convey.Convey("test GetOwnerRefByPG", t, func() {
		pgDemo1 := getDemoPodGroup(pgName1, pgNameSpace, jobUid1)
		convey.Convey("get owner reference success", func() {
			isControl := true
			expOwner := v1.OwnerReference{
				Name:       pgName1,
				Controller: &isControl,
				Kind:       vcJobKey,
				UID:        types.UID(jobUid1)}
			owner, err := GetOwnerRefByPG(pgDemo1)
			convey.So(err, convey.ShouldBeNil)
			convey.So(owner, convey.ShouldResemble, expOwner)
		})
		convey.Convey("get owner reference failed, info is nil", func() {
			_, err := GetOwnerRefByPG(nil)
			expErr := errors.New("pg info is nil")
			convey.So(err, convey.ShouldResemble, expErr)
		})
		convey.Convey("get owner reference failed, pg don't have controller", func() {
			pgDemo1.OwnerReferences = []v1.OwnerReference{}
			_, err := GetOwnerRefByPG(pgDemo1)
			expErr := errors.New("pg don't have controller")
			convey.So(err, convey.ShouldResemble, expErr)
		})
	})
}

func TestGetSubHealthStrategyByJobKey(t *testing.T) {
	convey.Convey("test GetSubHealthStrategyByJobKey", t, func() {
		pgDemo1 := getDemoPodGroup(pgName1, pgNameSpace, jobUid1)
		convey.Convey("when pg is nil, should return default strategy", func() {
			convey.So(GetSubHealthStrategyByJobKey(jobUid1), convey.ShouldEqual, constant.SubHealthyIngore)
		})
		convey.Convey("when pg is exists, subHealthyStrategy is not exists, should return default strategy", func() {
			SavePodGroup(pgDemo1)
			defer DeletePodGroup(pgDemo1)
			convey.So(GetSubHealthStrategyByJobKey(jobUid1), convey.ShouldEqual, constant.SubHealthyIngore)
		})
		convey.Convey("when pg is exists, subHealthyStrategy is exists, should return config value", func() {
			pgDemo1.Labels[constant.SubHealthyStrategy] = constant.SubHealthyGraceExit
			SavePodGroup(pgDemo1)
			defer DeletePodGroup(pgDemo1)
			convey.So(GetSubHealthStrategyByJobKey(jobUid1), convey.ShouldEqual, constant.SubHealthyGraceExit)
		})
	})
}

type retryGetFrameworkTestCase struct {
	name           string
	jobId          string
	podGroups      []v1beta1.PodGroup
	frameworks     []string
	expectedResult string
}

func TestRetryGetFramework(t *testing.T) {
	testCases := getRetryGetFrameworkTestCases()
	for _, tc := range testCases {
		convey.Convey(tc.name, t, func() {
			patches := setupPatches(tc)
			defer patches.Reset()
			result := RetryGetFramework(tc.jobId)
			convey.So(result, convey.ShouldEqual, tc.expectedResult)
		})
	}
}

func getRetryGetFrameworkTestCases() []retryGetFrameworkTestCase {
	emptyPG := v1beta1.PodGroup{}
	pg := getDemoPodGroup(testPodGroupName, pgNameSpace, testJobId)
	pg.Labels = map[string]string{}
	pgWithoutFramework := getDemoPodGroup(testPodGroupName, pgNameSpace, testJobId)
	pgWithoutFramework.Labels = map[string]string{}
	pgWithFramework := getDemoPodGroup(testPodGroupName, pgNameSpace, testJobId)
	return []retryGetFrameworkTestCase{
		{name: "should return framework when get framework success at first retry",
			jobId: testJobId,
			podGroups: []v1beta1.PodGroup{
				*getDemoPodGroup(testPodGroupName, pgNameSpace, testJobId),
			},
			frameworks:     []string{testFramework},
			expectedResult: testFramework},
		{name: "should return framework when get framework success at second retry",
			jobId: testJobId,
			podGroups: []v1beta1.PodGroup{
				v1beta1.PodGroup{},
				*getDemoPodGroup(testPodGroupName, pgNameSpace, testJobId),
			},
			frameworks:     []string{emptyFramework, testFramework},
			expectedResult: testFramework},
		{name: "should return empty when podGroup not exists after all retries",
			jobId: testJobId,
			podGroups: []v1beta1.PodGroup{
				emptyPG, emptyPG, emptyPG, emptyPG,
			},
			frameworks: []string{
				emptyFramework, emptyFramework, emptyFramework, emptyFramework,
			},
			expectedResult: emptyFramework},
		{name: "should return empty when framework not exists after all retries",
			jobId:     testJobId,
			podGroups: []v1beta1.PodGroup{*pg, *pg, *pg, *pg},
			frameworks: []string{
				emptyFramework, emptyFramework, emptyFramework, emptyFramework,
			},
			expectedResult: emptyFramework},
		{name: "should return framework when podGroup exists after retries and framework found",
			jobId: testJobId,
			podGroups: []v1beta1.PodGroup{
				v1beta1.PodGroup{},
				*pgWithoutFramework,
				*pgWithFramework,
			},
			frameworks:     []string{emptyFramework, emptyFramework, testFramework},
			expectedResult: testFramework},
	}
}

func setupPatches(tc retryGetFrameworkTestCase) *gomonkey.Patches {
	patches := gomonkey.NewPatches()
	callCount := firstRetryIndex
	patches.ApplyFunc(GetPodGroup, func(jobKey string) v1beta1.PodGroup {
		if callCount < len(tc.podGroups) {
			result := tc.podGroups[callCount]
			callCount++
			return result
		}
		return tc.podGroups[len(tc.podGroups)-1]
	})

	frameworkCallCount := firstRetryIndex
	patches.ApplyFunc(GetModelFramework, func(info *v1beta1.PodGroup) string {
		if frameworkCallCount < len(tc.frameworks) {
			result := tc.frameworks[frameworkCallCount]
			frameworkCallCount++
			return result
		}
		return tc.frameworks[len(tc.frameworks)-1]
	})

	patches.ApplyFuncReturn(time.Sleep)
	return patches
}
