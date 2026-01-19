// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package recover a series of service function
package recover

import (
	"errors"
	"strconv"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"volcano.sh/apis/pkg/apis/scheduling/v1beta1"

	"clusterd/pkg/common/constant"
	"clusterd/pkg/interface/grpc/recover"
	"clusterd/pkg/interface/kube"
)

const (
	sliceLength2  = 2
	testPgName    = "test-pg-name"
	testNameSpace = "test-namespace"
	faultString   = "fault1:type1,fault2:type2"
	successString = "success"
)

var mockErr = errors.New("mocked error")

// TestPlatFormStrategy tests the platFormStrategy function.
func TestPlatFormStrategy(t *testing.T) {
	convey.Convey("Testing platFormStrategy", t, func() {
		testPlatFormStrategyCase1(t)
		testPlatFormStrategyCase2(t)
		testPlatFormStrategyCase3(t)
		testPlatFormStrategyCase4(t)
		testPlatFormStrategyCase5(t)
	})
}

// testPlatFormStrategyCase1 tests the scenario where the PodGroup has a
// valid process recover strategy annotation.
// Expected result: The strategy should be equal to the expected strategy, and no error should occur.
func testPlatFormStrategyCase1(t *testing.T) {
	convey.Convey("Test case 1: PodGroup has a valid process recover strategy annotation. "+
		"Expected: strategy should equal ProcessRetryStrategyName, and no error should occur.", func() {
		patches := gomonkey.ApplyFunc(kube.RetryGetPodGroup,
			func(name, namespace string, times int) (*v1beta1.PodGroup, error) {
				return &v1beta1.PodGroup{
					ObjectMeta: v1.ObjectMeta{
						Annotations: map[string]string{
							constant.ProcessRecoverStrategy: constant.ProcessRetryStrategyName,
						},
					},
				}, nil
			})
		defer patches.Reset()

		strategy, err := platFormStrategy(testPgName, testNameSpace, false)
		convey.So(strategy, convey.ShouldEqual, constant.ProcessRetryStrategyName)
		convey.So(err, convey.ShouldBeNil)
	})
}

// testPlatFormStrategyCase2 tests the scenario where the PodGroup has
// no valid process recover strategy annotation.
// Expected result: An error should occur.
func testPlatFormStrategyCase2(t *testing.T) {
	convey.Convey("Test case 2: PodGroup has no valid process recover strategy annotation. "+
		"Expected: an error should occur.", func() {
		patches := gomonkey.ApplyFunc(kube.RetryGetPodGroup,
			func(name, namespace string, times int) (*v1beta1.PodGroup, error) {
				return &v1beta1.PodGroup{}, nil
			})
		defer patches.Reset()

		_, err := platFormStrategy(testPgName, testNameSpace, false)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

// testPlatFormStrategyCase3 tests the scenario where getting the PodGroup fails.
// Expected result: An error should occur.
func testPlatFormStrategyCase3(t *testing.T) {
	convey.Convey("Test case 3: Failed to get the PodGroup. "+
		"Expected: an error should occur.", func() {
		patches := gomonkey.ApplyFunc(kube.RetryGetPodGroup,
			func(name, namespace string, times int) (*v1beta1.PodGroup, error) {
				return nil, mockErr
			})
		defer patches.Reset()

		_, err := platFormStrategy(testPgName, testNameSpace, false)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

// testPlatFormStrategyCase4 tests the scenario where the PodGroup has an
// invalid process recover strategy annotation.
// Expected result: An error should occur.
func testPlatFormStrategyCase4(t *testing.T) {
	convey.Convey("Test case 4: PodGroup has an invalid process recover strategy annotation. "+
		"Expected: an error should occur.", func() {
		patches := gomonkey.ApplyFunc(kube.RetryGetPodGroup,
			func(name, namespace string, times int) (*v1beta1.PodGroup, error) {
				return &v1beta1.PodGroup{
					ObjectMeta: v1.ObjectMeta{
						Annotations: map[string]string{
							constant.ProcessRecoverStrategy: "invalid-strategy",
						},
					},
				}, nil
			})
		defer patches.Reset()

		_, err := platFormStrategy(testPgName, testNameSpace, false)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

// testPlatFormStrategyCase5 tests the scenario where the PodGroup has a
// valid process exit strategy annotation and the flag is true.
// Expected result: No error should occur.
func testPlatFormStrategyCase5(t *testing.T) {
	convey.Convey("Test case 5: PodGroup has a valid process exit strategy annotation and the flag is true. "+
		"Expected: no error should occur.", func() {
		patches := gomonkey.ApplyFunc(kube.RetryGetPodGroup,
			func(name, namespace string, times int) (*v1beta1.PodGroup, error) {
				return &v1beta1.PodGroup{
					ObjectMeta: v1.ObjectMeta{
						Annotations: map[string]string{
							constant.ProcessRecoverStrategy: constant.ProcessExitStrategyName,
						},
					},
				}, nil
			})
		defer patches.Reset()

		_, err := platFormStrategy(testPgName, testNameSpace, true)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestWaitPlatFormStrategyReady(t *testing.T) {
	convey.Convey("Testing WaitPlatFormStrategyReady", t, func() {
		patches := gomonkey.ApplyFunc(platFormStrategy,
			func(name, namespace string, confirmState bool) (string, error) {
				return constant.ProcessRetryStrategyName, nil
			})
		defer patches.Reset()

		strategy, err := WaitPlatFormStrategyReady(testPgName, testNameSpace)
		convey.So(strategy, convey.ShouldEqual, constant.ProcessRetryStrategyName)
		convey.So(err, convey.ShouldBeNil)
	})
}

// TestUpdateProcessConfirmFault tests the UpdateProcessConfirmFault function.
func TestUpdateProcessConfirmFault(t *testing.T) {
	convey.Convey("Testing UpdateProcessConfirmFault", t, func() {
		testUpdateProcessConfirmFaultCase1(t)
		testUpdateProcessConfirmFaultCase2(t)
		testUpdateProcessConfirmFaultCase3(t)
		testUpdateProcessConfirmFaultCase4(t)
	})
}

// testUpdateProcessConfirmFaultCase1 tests the scenario where
// getting the PodGroup and patching annotations succeed.
// Expected result: No error should occur.
func testUpdateProcessConfirmFaultCase1(t *testing.T) {
	convey.Convey("Test case 1: Getting the PodGroup and patching annotations succeed. "+
		"Expected: No error should occur.", func() {
		patches := gomonkey.ApplyFunc(kube.RetryGetPodGroup,
			func(name, namespace string, times int) (*v1beta1.PodGroup, error) {
				return &v1beta1.PodGroup{}, nil
			}).
			ApplyFunc(kube.RetryPatchPodGroupAnnotations,
				func(name, namespace string, times int, annotations map[string]string) (*v1beta1.PodGroup, error) {
					return &v1beta1.PodGroup{}, nil
				})
		defer patches.Reset()

		err := UpdateProcessConfirmFault(testPgName, testNameSpace, []*pb.FaultRank{})
		convey.So(err, convey.ShouldBeNil)
	})
}

// testUpdateProcessConfirmFaultCase2 tests the scenario where getting the PodGroup fails.
// Expected result: An error should occur.
func testUpdateProcessConfirmFaultCase2(t *testing.T) {
	convey.Convey("Test case 2: Getting the PodGroup fails. "+
		"Expected: An error should occur.", func() {
		patches := gomonkey.ApplyFunc(kube.RetryGetPodGroup,
			func(name, namespace string, times int) (*v1beta1.PodGroup, error) {
				return nil, mockErr
			})
		defer patches.Reset()

		err := UpdateProcessConfirmFault(testPgName, testNameSpace, []*pb.FaultRank{})
		convey.So(err, convey.ShouldNotBeNil)
	})
}

// testUpdateProcessConfirmFaultCase3 tests the scenario where the PodGroup already has a
// ProcessConfirmFaultKey annotation.
// Expected result: An error should occur.
func testUpdateProcessConfirmFaultCase3(t *testing.T) {
	convey.Convey("Test case 3: The PodGroup already has a ProcessConfirmFaultKey annotation. "+
		"Expected: An error should occur.", func() {
		patches := gomonkey.ApplyFunc(kube.RetryGetPodGroup,
			func(name, namespace string, times int) (*v1beta1.PodGroup, error) {
				return &v1beta1.PodGroup{
					ObjectMeta: v1.ObjectMeta{
						Annotations: map[string]string{
							constant.ProcessConfirmFaultKey: faultString,
						},
					},
				}, nil
			}).ApplyFunc(kube.RetryPatchPodGroupAnnotations,
			func(name, namespace string, times int, annotations map[string]string) (*v1beta1.PodGroup, error) {
				return &v1beta1.PodGroup{}, nil
			})
		defer patches.Reset()

		err := UpdateProcessConfirmFault(testPgName, testNameSpace, []*pb.FaultRank{})
		convey.So(err, convey.ShouldNotBeNil)
	})
}

// testUpdateProcessConfirmFaultCase4 tests the scenario where patching the PodGroup annotations fails.
// Expected result: An error should occur.
func testUpdateProcessConfirmFaultCase4(t *testing.T) {
	convey.Convey("Test case 4: Patching the PodGroup annotations fails. "+
		"Expected: An error should occur.", func() {
		patches := gomonkey.ApplyFunc(kube.RetryGetPodGroup,
			func(name, namespace string, times int) (*v1beta1.PodGroup, error) {
				return &v1beta1.PodGroup{}, nil
			}).
			ApplyFunc(kube.RetryPatchPodGroupAnnotations,
				func(name, namespace string, times int, annotations map[string]string) (*v1beta1.PodGroup, error) {
					return &v1beta1.PodGroup{}, mockErr
				})
		defer patches.Reset()

		err := UpdateProcessConfirmFault(testPgName, testNameSpace, []*pb.FaultRank{})
		convey.So(err, convey.ShouldNotBeNil)
	})
}

// TestPullProcessResultFault tests the pullProcessResultFault function.
func TestPullProcessResultFault(t *testing.T) {
	convey.Convey("Testing pullProcessResultFault", t, func() {
		testPullProcessResultFaultCase1()
		testPullProcessResultFaultCase2()
		testPullProcessResultFaultCase3()
		testPullProcessResultFaultCase4()
		testPullProcessResultFaultCase5()
	})
}

// testPullProcessResultFaultCase1 tests the scenario where the fault ranks are successfully retrieved.
// Expected result: The resultRanks and confirmRanks slices should have length 2, and no error should occur.
func testPullProcessResultFaultCase1() {
	convey.Convey("Test case 1: Successful retrieval of fault ranks. "+
		"Expected: resultRanks and confirmRanks slices should have length 2, and no error should occur.", func() {
		patches := gomonkey.ApplyFunc(kube.RetryGetPodGroup,
			func(name, namespace string, times int) (*v1beta1.PodGroup, error) {
				return &v1beta1.PodGroup{
					ObjectMeta: v1.ObjectMeta{
						Annotations: map[string]string{
							constant.ProcessResultFaultKey:  faultString,
							constant.ProcessConfirmFaultKey: faultString,
						},
					},
				}, nil
			})
		defer patches.Reset()

		resultRanks, confirmRanks, err :=
			pullProcessResultFault(testPgName, testNameSpace)
		convey.So(resultRanks, convey.ShouldHaveLength, sliceLength2)
		convey.So(confirmRanks, convey.ShouldHaveLength, sliceLength2)
		convey.So(err, convey.ShouldBeNil)
	})
}

// testPullProcessResultFaultCase2 tests the scenario where getting the PodGroup fails.
// Expected result: An error should occur.
func testPullProcessResultFaultCase2() {
	convey.Convey("Test case 2: Failed to get PodGroup. "+
		"Expected: An error should occur.", func() {
		patches := gomonkey.ApplyFunc(kube.RetryGetPodGroup,
			func(name, namespace string, times int) (*v1beta1.PodGroup, error) {
				return nil, mockErr
			})
		defer patches.Reset()

		_, _, err := pullProcessResultFault(testPgName, testNameSpace)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

// testPullProcessResultFaultCase3 tests the scenario where the PodGroup has no annotations.
// Expected result: An error should occur.
func testPullProcessResultFaultCase3() {
	convey.Convey("Test case 3: PodGroup has no annotations. "+
		"Expected: An error should occur.", func() {
		patches := gomonkey.ApplyFunc(kube.RetryGetPodGroup,
			func(name, namespace string, times int) (*v1beta1.PodGroup, error) {
				return &v1beta1.PodGroup{
					ObjectMeta: v1.ObjectMeta{
						Annotations: nil,
					},
				}, nil
			})
		defer patches.Reset()

		_, _, err := pullProcessResultFault(testPgName, testNameSpace)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

// testPullProcessResultFaultCase4 tests the scenario where the PodGroup has only ProcessConfirmFaultKey.
// Expected result: An error should occur.
func testPullProcessResultFaultCase4() {
	convey.Convey("Test case 4: PodGroup has only ProcessConfirmFaultKey. "+
		"Expected: An error should occur.", func() {
		patches := gomonkey.ApplyFunc(kube.RetryGetPodGroup,
			func(name, namespace string, times int) (*v1beta1.PodGroup, error) {
				return &v1beta1.PodGroup{
					ObjectMeta: v1.ObjectMeta{
						Annotations: map[string]string{
							constant.ProcessConfirmFaultKey: faultString,
						},
					},
				}, nil
			})
		defer patches.Reset()

		_, _, err := pullProcessResultFault(testPgName, testNameSpace)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

// testPullProcessResultFaultCase5 tests the scenario where the PodGroup has only ProcessResultFaultKey.
// Expected result: No error should occur.
func testPullProcessResultFaultCase5() {
	convey.Convey("Test case 5: PodGroup has only ProcessResultFaultKey. "+
		"Expected: No error should occur.", func() {
		patches := gomonkey.ApplyFunc(kube.RetryGetPodGroup,
			func(name, namespace string, times int) (*v1beta1.PodGroup, error) {
				return &v1beta1.PodGroup{
					ObjectMeta: v1.ObjectMeta{
						Annotations: map[string]string{
							constant.ProcessResultFaultKey: faultString,
						},
					},
				}, nil
			})
		defer patches.Reset()

		_, _, err := pullProcessResultFault(testPgName, testNameSpace)
		convey.So(err, convey.ShouldBeNil)
	})
}

// TestWaitProcessResultFault tests the WaitProcessResultFault function.
func TestWaitProcessResultFault(t *testing.T) {
	convey.Convey("Testing WaitProcessResultFault", t, func() {
		patches := gomonkey.ApplyFunc(pullProcessResultFault,
			func(name, namespace string) ([]*pb.FaultRank, []*pb.FaultRank, error) {
				return []*pb.FaultRank{}, []*pb.FaultRank{}, nil
			})
		defer patches.Reset()

		convey.Convey("Test case: Successful waiting for process result fault. "+
			"Expected: No error should occur.", func() {
			_, err := WaitProcessResultFault(testPgName, testNameSpace)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestRankTableReady tests the rankTableReady function.
func TestRankTableReady(t *testing.T) {
	convey.Convey("Testing rankTableReady", t, func() {
		testRankTableReadyCase1()
		testRankTableReadyCase2()
		testRankTableReadyCase3()
		testRankTableReadyCase4()
		testRankTableReadyCase5()
	})
}

// testRankTableReadyCase1 tests the scenario where the rank table is ready.
// Expected result: The rankTableReady function should return a true str.
func testRankTableReadyCase1() {
	convey.Convey("Test case 1: Rank table is ready. "+
		"Expected: The rankTableReady function should return an empty str and a nil err.", func() {
		patches := gomonkey.ApplyFunc(kube.RetryGetPodGroup,
			func(name, namespace string, times int) (*v1beta1.PodGroup, error) {
				return &v1beta1.PodGroup{
					ObjectMeta: v1.ObjectMeta{
						Annotations: map[string]string{
							constant.RankTableReadyKey: strconv.FormatBool(true),
						},
					},
				}, nil
			})
		defer patches.Reset()

		ready := rankTableReady(testPgName, testNameSpace)
		convey.So(ready, convey.ShouldEqual, strconv.FormatBool(true))
	})
}

// testRankTableReadyCase2 tests the scenario where the rank table is not ready.
// Expected result: The rankTableReady function should return a false str.
func testRankTableReadyCase2() {
	convey.Convey("Test case 2: Rank table is not ready. "+
		"Expected: The rankTableReady function should return false.", func() {
		patches := gomonkey.ApplyFunc(kube.RetryGetPodGroup,
			func(name, namespace string, times int) (*v1beta1.PodGroup, error) {
				return &v1beta1.PodGroup{
					ObjectMeta: v1.ObjectMeta{
						Annotations: map[string]string{
							constant.RankTableReadyKey: strconv.FormatBool(false),
						},
					},
				}, nil
			})
		defer patches.Reset()

		ready := rankTableReady(testPgName, testNameSpace)
		convey.So(ready, convey.ShouldEqual, strconv.FormatBool(false))
	})
}

// testRankTableReadyCase3 tests the scenario where getting the PodGroup fails.
// Expected result: The rankTableReady function should return an empty str.
func testRankTableReadyCase3() {
	convey.Convey("Test case 3: Failed to get PodGroup. "+
		"Expected: The rankTableReady function should return false.", func() {
		patches := gomonkey.ApplyFunc(kube.RetryGetPodGroup,
			func(name, namespace string, times int) (*v1beta1.PodGroup, error) {
				return nil, mockErr
			})
		defer patches.Reset()

		ready := rankTableReady(testPgName, testNameSpace)
		convey.So(ready, convey.ShouldEqual, "")
	})
}

// testRankTableReadyCase4 tests the scenario where the PodGroup has no annotations.
// Expected result: The rankTableReady function should return an empty str.
func testRankTableReadyCase4() {
	convey.Convey("Test case 4: PodGroup has no annotations. "+
		"Expected: The rankTableReady function should return false.", func() {
		patches := gomonkey.ApplyFunc(kube.RetryGetPodGroup,
			func(name, namespace string, times int) (*v1beta1.PodGroup, error) {
				return &v1beta1.PodGroup{
					ObjectMeta: v1.ObjectMeta{
						Annotations: nil,
					},
				}, nil
			})
		defer patches.Reset()

		ready := rankTableReady(testPgName, testNameSpace)
		convey.So(ready, convey.ShouldEqual, "")
	})
}

// testRankTableReadyCase5 tests the scenario where RankTable annotation value invalid.
// Expected result: The rankTableReady function should return a wrong str.
func testRankTableReadyCase5() {
	convey.Convey("Test case 5: PodGroup has no annotations. "+
		"Expected: The rankTableReady function should return false.", func() {
		patches := gomonkey.ApplyFunc(kube.RetryGetPodGroup,
			func(name, namespace string, times int) (*v1beta1.PodGroup, error) {
				return &v1beta1.PodGroup{
					ObjectMeta: v1.ObjectMeta{
						Annotations: map[string]string{
							constant.RankTableReadyKey: "wrong",
						},
					},
				}, nil
			})
		defer patches.Reset()

		ready := rankTableReady(testPgName, testNameSpace)
		convey.So(ready, convey.ShouldNotEqual, "")
	})
}

// TestWaitRankTableReady tests the WaitRankTableReady function return normal.
func TestWaitRankTableReady(t *testing.T) {
	convey.Convey("Testing WaitRankTableReady", t, func() {
		patches := gomonkey.ApplyFunc(rankTableReady, func(name, namespace string) string {
			return "true"
		})
		defer patches.Reset()
		convey.Convey("Test case: Successful waiting for rank table to be ready. "+
			"Expected: No error should occur.", func() {
			err := WaitRankTableReady(testPgName, testNameSpace)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestWaitRankTableNotReady tests the WaitRankTableReady function return abnormal.
func TestWaitRankTableNotReady(t *testing.T) {
	convey.Convey("Testing WaitRankTableNotReady", t, func() {
		patches := gomonkey.ApplyFunc(rankTableReady, func(name, namespace string) string {
			return "false"
		})
		defer patches.Reset()
		convey.Convey("Test case: waiting for rank table to be ready failed. "+
			"Expected: error should occur.", func() {
			err := WaitRankTableReady(testPgName, testNameSpace)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}
