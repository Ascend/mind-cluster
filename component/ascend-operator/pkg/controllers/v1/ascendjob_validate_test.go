/*
Copyright(C) 2023. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package controllers is using for reconcile Job.
*/

package v1

import (
	"errors"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	commonv1 "github.com/kubeflow/common/pkg/apis/common/v1"
	"github.com/smartystreets/goconvey/convey"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"volcano.sh/apis/pkg/apis/scheduling/v1beta1"

	"ascend-common/api"
	mindxdlv1 "ascend-operator/pkg/api/v1"
	_ "ascend-operator/pkg/testtool"
)

const fakeReplicas2 = 2

func TestValidateJob(t *testing.T) {
	convey.Convey("reconciler validate job", t, func() {
		rc := newCommonReconciler()
		job := newCommonAscendJob()
		job.Spec.ReplicaSpecs = map[commonv1.ReplicaType]*commonv1.ReplicaSpec{}
		convey.Convey("01-valid basic info failed should return err", func() {
			fakeErr := &validateError{
				reason:  "fake reason",
				message: "valid basic info failed",
			}
			patch := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "validateBasicInfo",
				func(_ *ASJobReconciler, _ *mindxdlv1.Job) *validateError {
					return fakeErr
				})
			defer patch.Reset()
			err := rc.validateJob(job)
			convey.So(err, convey.ShouldResemble, fakeErr)
		})
		convey.Convey("02-valid job spec failed should return err", func() {
			fakeErr := &validateError{
				reason:  "fake reason",
				message: "valid job spec failed",
			}
			patch := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "validateSpec", func(_ *ASJobReconciler,
				_ *mindxdlv1.Job, _ map[commonv1.ReplicaType]*commonv1.ReplicaSpec) *validateError {
				return fakeErr
			})
			defer patch.Reset()
			err := rc.validateJob(job)
			convey.So(err, convey.ShouldResemble, fakeErr)
		})
	})
}

func TestValidateReplicas(t *testing.T) {
	convey.Convey("Test validateReplicas", t, func() {
		convey.Convey("Case 1: Replicas is nil should return nil", func() {
			spec := &commonv1.ReplicaSpec{
				Replicas: nil,
			}
			err := validateReplicas(spec)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("Case 2: Replicas < 0 should return error", func() {
			replica := int32(-1)
			spec := &commonv1.ReplicaSpec{
				Replicas: &replica,
			}
			expectedErr := &validateError{
				reason:  "ReplicaTypeError",
				message: "jobSpec is not valid: replicas can not be negative num, but got -1",
			}

			err := validateReplicas(spec)
			convey.So(err, convey.ShouldResemble, expectedErr)
		})

		convey.Convey("Case 3: Replicas > maxReplicas should return error", func() {
			replica := int32(maxReplicas + 1)
			spec := &commonv1.ReplicaSpec{
				Replicas: &replica,
			}
			expectedErr := &validateError{
				reason:  "ReplicaTypeError",
				message: "jobSpec is not valid: replicas can not be larger than 15000, but got 15001",
			}

			err := validateReplicas(spec)
			convey.So(err, convey.ShouldResemble, expectedErr)
		})

		convey.Convey("Case 4: Valid replicas (e.g. 0) should return nil", func() {
			replica := int32(0)
			spec := &commonv1.ReplicaSpec{
				Replicas: &replica,
			}
			err := validateReplicas(spec)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestValidateBasicInfo test validateBasicInfo
func TestValidateBasicInfo(t *testing.T) {
	convey.Convey("reconciler valid basic info", t, func() {
		rc := &ASJobReconciler{}
		job := newCommonAscendJob()
		convey.Convey("01-job ReplicaSpecs is nil, should return err", func() {
			err := rc.validateBasicInfo(job)
			convey.So(err, convey.ShouldResemble, &validateError{
				reason:  "SpecsError",
				message: "job spec is not valid",
			})
		})
		job.Spec.ReplicaSpecs = map[commonv1.ReplicaType]*commonv1.ReplicaSpec{}
		convey.Convey("02-job SuccessPolicy is invalid, should return err", func() {
			fakeSuccessPolicy := mindxdlv1.SuccessPolicy("xxx")
			job.Spec.SuccessPolicy = &fakeSuccessPolicy
			err := rc.validateBasicInfo(job)
			convey.So(err, convey.ShouldResemble, &validateError{
				reason:  "SuccessPolicyError",
				message: `job success policy is invalid, it must be one of <"", AllWorkers>`,
			})
		})
		convey.Convey("03-job queue is not exist, should return err", func() {
			rc.Config.EnableGangScheduling = true
			schedulingPolicy := &commonv1.SchedulingPolicy{
				Queue: "XXX",
			}
			job.Spec.RunPolicy.SchedulingPolicy = schedulingPolicy
			patch := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "getQueueFromApiserver",
				func(_ *ASJobReconciler, _ string) (*v1beta1.Queue, error) {
					return nil, errors.New("not found")
				})
			defer patch.Reset()
			err := rc.validateBasicInfo(job)
			convey.So(err, convey.ShouldResemble, &validateError{
				reason:  "QueueGetFailed",
				message: "not found",
			})
		})
		convey.Convey("04-job queue is exist, should return nil", func() {
			err := rc.validateBasicInfo(job)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestValidateSpec test validateSpec
func TestValidateSpec(t *testing.T) {
	convey.Convey("reconciler valid job spec", t, func() {
		rc := &ASJobReconciler{}
		job := newCommonAscendJob()
		spec := make(map[commonv1.ReplicaType]*commonv1.ReplicaSpec)
		convey.Convey("01-job framework label is not set, should return err", func() {
			err := rc.validateSpec(job, spec)
			convey.So(err, convey.ShouldResemble, &validateError{
				reason:  "FrameworkLabelError",
				message: "framework label is not set",
			})
		})
		convey.Convey("02-job framework label is invalid, should return err", func() {
			job.Labels = map[string]string{mindxdlv1.FrameworkKey: "xxx"}
			err := rc.validateSpec(job, spec)
			convey.So(err, convey.ShouldResemble, &validateError{
				reason:  "FrameworkLabelError",
				message: "framework label<xxx> is not in map[mindspore:{} pytorch:{} tensorflow:{}]",
			})
		})
		convey.Convey("03-job framework label is valid, should return nil", func() {
			job.Labels = map[string]string{mindxdlv1.FrameworkKey: "mindspore"}
			patch := gomonkey.ApplyFunc(checkReplicaSpecs, func(_ string,
				_ map[commonv1.ReplicaType]*commonv1.ReplicaSpec) *validateError {
				return nil
			})
			defer patch.Reset()
			err := rc.validateSpec(job, spec)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestJobTotalRequest test jobTotalRequest
func TestJobTotalRequest(t *testing.T) {
	convey.Convey("get ms job total require npu", t, func() {
		convey.Convey("01-job require 8 npu, should return 8", func() {
			replicas := int32(1)
			expectResult := 8
			specs := map[commonv1.ReplicaType]*commonv1.ReplicaSpec{
				mindxdlv1.MindSporeReplicaTypeScheduler: nil,
				mindxdlv1.ReplicaTypeWorker: {
					Replicas: &replicas,
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{Name: api.DefaultContainerName}},
						},
					},
				},
			}
			patch1 := gomonkey.ApplyFunc(getContainerResourceReq, func(_ corev1.Container) int {
				return expectResult
			})
			defer patch1.Reset()
			res := jobTotalRequest(specs)
			convey.So(res, convey.ShouldEqual, expectResult)
		})
	})
}

// TestValidContainerNum test validContainerNum
func TestValidContainerNum(t *testing.T) {
	convey.Convey("validContainerNum", t, func() {
		rtype := mindxdlv1.ReplicaTypeWorker
		convey.Convey("01-spec is nil will return err", func() {
			err := validContainerNum(rtype, nil)
			convey.So(err, convey.ShouldResemble, &validateError{
				reason:  "ReplicaTypeError",
				message: fmt.Sprintf("jobSpec is not valid: containers definition expected in %v", rtype),
			})
		})
		convey.Convey("02-spec's container will return err", func() {
			spec := newCommonSpec()
			spec.Template.Spec.Containers = []corev1.Container{}
			err := validContainerNum(rtype, spec)
			convey.So(err, convey.ShouldResemble, &validateError{
				reason:  "ReplicaTypeError",
				message: fmt.Sprintf("jobSpec is not valid: containers definition expected in %v", rtype),
			})
		})
		convey.Convey("03-spec's container is valid will return nil", func() {
			spec := newCommonSpec()
			err := validContainerNum(rtype, spec)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestCheckReplicaSpecs01(t *testing.T) {
	convey.Convey("01-checkReplicaSpecs", t, func() {
		frame := mindxdlv1.MindSporeFrameworkName
		specs := map[commonv1.ReplicaType]*commonv1.ReplicaSpec{}
		convey.Convey("02-spec without container will return err", func() {
			spec := newCommonSpec()
			spec.Replicas = newReplicas(1)
			spec.Template.Spec.Containers = []corev1.Container{}
			specs[mindxdlv1.ReplicaTypeWorker] = spec
			err := checkReplicaSpecs(frame, specs)
			convey.So(err, convey.ShouldResemble, &validateError{
				reason: "ReplicaTypeError",
				message: fmt.Sprintf("jobSpec is not valid: containers definition expected in %v",
					mindxdlv1.ReplicaTypeWorker),
			})
		})
		convey.Convey("03-spec with invalid replica type should return error", func() {
			spec := newCommonSpec()
			spec.Replicas = newReplicas(1)
			delete(specs, mindxdlv1.ReplicaTypeWorker)
			fakeType := commonv1.ReplicaType("fake")
			specs[fakeType] = spec
			err := checkReplicaSpecs(frame, specs)
			convey.So(err, convey.ShouldResemble, &validateError{
				reason: "ReplicaTypeError",
				message: fmt.Sprintf("replicaType is %v but must be one of %v", fakeType, []commonv1.ReplicaType{
					mindxdlv1.MindSporeReplicaTypeScheduler,
					mindxdlv1.ReplicaTypeWorker,
				}),
			})
		})
	})
}

func TestCheckReplicaSpecs02(t *testing.T) {
	convey.Convey("02-checkReplicaSpecs", t, func() {
		frame := mindxdlv1.PytorchFrameworkName
		specs := map[commonv1.ReplicaType]*commonv1.ReplicaSpec{}
		spec := newCommonSpec()
		convey.Convey("04-leader with invalid replicas should return error", func() {
			spec.Replicas = newReplicas(fakeReplicas2)
			specs[mindxdlv1.PytorchReplicaTypeMaster] = spec
			err := checkReplicaSpecs(frame, specs)
			convey.So(err, convey.ShouldResemble, &validateError{
				reason:  "ReplicaTypeError",
				message: "replicaType<Master> replicas is invalid, it must be only 1",
			})
		})
		spec.Replicas = newReplicas(1)
		ct := newCommonContainer()
		ct.Image = "xxx"
		ct.Name = api.DefaultContainerName
		convey.Convey("06-pytorch  without leader replicas should return error", func() {
			err := checkReplicaSpecs(frame, specs)
			convey.So(err, convey.ShouldResemble, &validateError{
				reason:  "ReplicaTypeError",
				message: fmt.Sprintf("ReplicaType is not valid: there need 1 leader replica-type"),
			})
		})
	})
}

func TestCheckReplicaSpecs03(t *testing.T) {
	convey.Convey("03-checkReplicaSpecs", t, func() {
		spec := newCommonSpec()
		spec.Replicas = newReplicas(1)
		specs := map[commonv1.ReplicaType]*commonv1.ReplicaSpec{}
		ct := newCommonContainer()
		ct.Image = "xxx"
		ct.Name = api.DefaultContainerName
		spec.Template.Spec.Containers[0] = ct
		specs[mindxdlv1.ReplicaTypeWorker] = spec
		frame := mindxdlv1.MindSporeFrameworkName
		convey.Convey("07-pytorch  without leader replicas should return error", func() {
			err := checkReplicaSpecs(frame, specs)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("08-pytorch  without leader replicas should return error", func() {
			specs[mindxdlv1.ReplicaTypeWorker].Template.Spec.Containers[0].Resources.Requests = corev1.ResourceList{
				fakeResourceName: resource.MustParse("2"),
			}
			specs[mindxdlv1.ReplicaTypeWorker].Template.Spec.Containers[0].Resources.Limits = corev1.ResourceList{
				fakeResourceName: resource.MustParse("2"),
			}
			err := checkReplicaSpecs(frame, specs)
			convey.So(err, convey.ShouldResemble, &validateError{
				reason: "ReplicaTypeError",
				message: fmt.Sprintf("replicaType is not valid: schdeuler not found, " +
					"but need 1 while req npu more than 1"),
			})
		})
	})
}

// TestValidateContainer test validateContainer
func TestValidateContainer(t *testing.T) {
	convey.Convey("TestValidateContainer", t, func() {
		spec := newCommonSpec()
		rtype := mindxdlv1.ReplicaTypeWorker
		container := newCommonContainer()
		convey.Convey("01-spec without container  should return error", func() {
			spec.Template.Spec.Containers = []corev1.Container{}
			err := validateContainer(rtype, spec)
			convey.So(err, convey.ShouldResemble, &validateError{
				reason: "ContainerError",
				message: fmt.Sprintf("replicaType is not valid: There is no container named %s in %v",
					api.DefaultContainerName, rtype),
			})
		})
		convey.Convey("02-container without setting image should return nil", func() {
			spec.Template.Spec.Containers[0] = container
			err := validateContainer(rtype, spec)
			convey.So(err, convey.ShouldResemble, &validateError{
				reason: "ContainerError",
				message: fmt.Sprintf("replicaType is not valid: Image is undefined in the container of %v",
					rtype),
			})
		})
		convey.Convey("03-spec without ascend container should return nil", func() {
			container.Image = "fake-image"
			spec.Template.Spec.Containers[0] = container
			err := validateContainer(rtype, spec)
			convey.So(err, convey.ShouldResemble, &validateError{
				reason: "ContainerError",
				message: fmt.Sprintf("replicaType is not valid: There is no container named %s in %v",
					api.DefaultContainerName, rtype),
			})
		})
	})
}

// TestGetValidReplicaType test getValidReplicaType
func TestGetValidReplicaType(t *testing.T) {
	convey.Convey("GetValidReplicaType", t, func() {
		convey.Convey("01-wrong frame should return nil", func() {
			frame := "fakeFrame"
			res := getValidReplicaType(frame)
			convey.So(res, convey.ShouldBeNil)
		})
		convey.Convey("02-mindspore frame should return valid rtype", func() {
			frame := mindxdlv1.MindSporeFrameworkName
			res := getValidReplicaType(frame)
			convey.So(res, convey.ShouldResemble, []commonv1.ReplicaType{
				mindxdlv1.MindSporeReplicaTypeScheduler,
				mindxdlv1.ReplicaTypeWorker})
		})
		convey.Convey("03-pytorch frame should return valid rtype", func() {
			frame := mindxdlv1.PytorchFrameworkName
			res := getValidReplicaType(frame)
			convey.So(res, convey.ShouldResemble, []commonv1.ReplicaType{
				mindxdlv1.PytorchReplicaTypeMaster,
				mindxdlv1.ReplicaTypeWorker})

		})
		convey.Convey("04-tensorflow frame should return valid rtype", func() {
			frame := mindxdlv1.TensorflowFrameworkName
			res := getValidReplicaType(frame)
			convey.So(res, convey.ShouldResemble, []commonv1.ReplicaType{
				mindxdlv1.TensorflowReplicaTypeChief,
				mindxdlv1.ReplicaTypeWorker,
			})
		})
	})
}

// TestValidateReplicaType test validateReplicaType
func TestValidateReplicaType(t *testing.T) {
	convey.Convey("TestValidateReplicaType", t, func() {
		convey.Convey("01-wrong frame should return error", func() {
			frame := "fakeFrame"
			rtype := mindxdlv1.ReplicaTypeWorker
			res := validateReplicaType(frame, rtype)
			convey.So(res, convey.ShouldNotBeNil)
		})
		convey.Convey("02-right frame and wrong rtype should return error", func() {
			frame := mindxdlv1.MindSporeFrameworkName
			rtype := "fakeReplicaType"
			res := validateReplicaType(frame, commonv1.ReplicaType(rtype))
			convey.So(res, convey.ShouldNotBeNil)
		})
		convey.Convey("03-right frame and right rtype should return nil", func() {
			frame := mindxdlv1.MindSporeFrameworkName
			rtype := mindxdlv1.ReplicaTypeWorker
			res := validateReplicaType(frame, rtype)
			convey.So(res, convey.ShouldBeNil)
		})
	})
}

// TestValidateLeader test validateLeader
func TestValidateLeader(t *testing.T) {
	convey.Convey("validateLeader", t, func() {
		rtype := mindxdlv1.PytorchReplicaTypeMaster
		spec := newCommonSpec()
		convey.Convey("01-nil replicas should return nil", func() {
			res := validateLeader(rtype, spec)
			convey.So(res, convey.ShouldBeNil)
		})
		convey.Convey("02-1 replicas should return nil", func() {
			spec.Replicas = newReplicas(1)
			res := validateLeader(rtype, spec)
			convey.So(res, convey.ShouldBeNil)
		})
		convey.Convey("03-2 replicas should return err", func() {
			const mockReplicas = 2
			spec.Replicas = newReplicas(mockReplicas)
			res := validateLeader(rtype, spec)
			convey.So(res, convey.ShouldResemble, &validateError{
				reason:  "ReplicaTypeError",
				message: fmt.Sprintf("replicaType<%v> replicas is invalid, it must be only 1", rtype),
			})
		})
	})
}

// TestGetReplicaSpecRequestRes test getReplicaSpecRequestRes
func TestGetReplicaSpecRequestRes(t *testing.T) {
	convey.Convey("GetReplicaSpecRequestRes", t, func() {
		spec := newCommonSpec()
		convey.Convey("01-spec without container should return 0 npu", func() {
			res := getReplicaSpecRequestRes(spec)
			convey.So(res, convey.ShouldEqual, 0)
		})
		container := newCommonContainer()
		convey.Convey("02-spec with out default container should return 0 npu", func() {
			spec.Template.Spec.Containers[0] = container
			res := getReplicaSpecRequestRes(spec)
			convey.So(res, convey.ShouldEqual, 0)
		})
		convey.Convey("03-spec with default container should return 1 npu", func() {
			container.Name = api.DefaultContainerName
			spec.Template.Spec.Containers[0] = container
			res := getReplicaSpecRequestRes(spec)
			convey.So(res, convey.ShouldEqual, 1)
		})
	})
}
