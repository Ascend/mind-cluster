/*
Copyright(C) 2023. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package controllers is using for reconcile AscendJob.
*/

package v1

import (
	"context"
	"fmt"

	commonv1 "github.com/kubeflow/common/pkg/apis/common/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"volcano.sh/apis/pkg/apis/scheduling/v1beta1"

	"ascend-common/common-utils/hwlog"
	mindxdlv1 "ascend-operator/pkg/api/v1"
)

func (r *ASJobReconciler) validateJob(job *mindxdlv1.AscendJob) *validateError {
	if r == nil {
		return &validateError{
			reason:  "ArgumentError",
			message: "nil pointer",
		}
	}

	var err *validateError
	defer func() {
		if err != nil {
			r.recorder.Event(job, corev1.EventTypeWarning, err.reason, err.message)
		}
	}()

	if err = r.validateBasicInfo(job); err != nil {
		return err
	}

	err = r.validateSpec(job, job.Spec.ReplicaSpecs)
	return err
}

func (r *ASJobReconciler) validateBasicInfo(job *mindxdlv1.AscendJob) *validateError {
	if job.Spec.ReplicaSpecs == nil {
		return &validateError{
			reason:  "SpecsError",
			message: "job spec is not valid",
		}
	}

	if job.Spec.SuccessPolicy != nil &&
		*job.Spec.SuccessPolicy != mindxdlv1.SuccessPolicyDefault &&
		*job.Spec.SuccessPolicy != mindxdlv1.SuccessPolicyAllWorkers {
		return &validateError{
			reason:  "SuccessPolicyError",
			message: `job success policy is invalid, it must be one of <"", AllWorkers>`,
		}
	}

	if r.Config.EnableGangScheduling && job.Spec.RunPolicy.SchedulingPolicy != nil {
		queueName := job.Spec.RunPolicy.SchedulingPolicy.Queue
		queue := v1beta1.Queue{}
		if err := r.Get(context.TODO(), types.NamespacedName{Name: queueName}, &queue); err != nil {
			return &validateError{
				reason:  "QueueGetFailed",
				message: err.Error(),
			}
		}
	}

	return nil
}

func (r *ASJobReconciler) validateSpec(job *mindxdlv1.AscendJob,
	specs map[commonv1.ReplicaType]*commonv1.ReplicaSpec) *validateError {
	frame, err := mindxdlv1.GetJobFramework(job)
	if err != nil {
		return &validateError{
			reason:  "FrameworkLabelError",
			message: err.Error(),
		}
	}

	hwlog.RunLog.Debugf("validate framework<%s> replica specs", frame)
	return checkReplicaSpecs(frame, specs)
}

func validContainerNum(rType commonv1.ReplicaType, spec *commonv1.ReplicaSpec) *validateError {
	if spec == nil || len(spec.Template.Spec.Containers) == 0 {
		return &validateError{
			reason:  "ReplicaTypeError",
			message: fmt.Sprintf("jobSpec is not valid: containers definition expected in %v", rType),
		}
	}
	return nil
}

func checkReplicaSpecs(frame string, specs map[commonv1.ReplicaType]*commonv1.ReplicaSpec) *validateError {
	hasLeader := false
	for rType, value := range specs {
		if *value.Replicas == 0 {
			return &validateError{
				reason:  "ReplicaTypeError",
				message: fmt.Sprintf("replicaType<%s> replicas is 0", rType),
			}
		}
		if ve := validContainerNum(rType, value); ve != nil {
			return ve
		}

		if ve := validateReplicaType(frame, rType); ve != nil {
			return ve
		}

		if rType != mindxdlv1.ReplicaTypeWorker {
			hasLeader = true
			if err := validateLeader(rType, value); err != nil {
				return err
			}
		}

		if err := validateContainer(rType, value); err != nil {
			return err
		}
	}

	if !hasLeader {
		if frame != mindxdlv1.MindSporeFrameworkName {
			return &validateError{
				reason:  "ReplicaTypeError",
				message: fmt.Sprintf("ReplicaType is not valid: there need 1 leader replica-type"),
			}
		}
		if jobTotalRequest(specs) > 1 {
			return &validateError{
				reason: "ReplicaTypeError",
				message: fmt.Sprintf("replicaType is not valid: schdeuler not found, " +
					"but need 1 while req npu more than 1"),
			}
		}
	}

	return nil
}

func validateContainer(rType commonv1.ReplicaType, spec *commonv1.ReplicaSpec) *validateError {
	validReq := func(ct corev1.Container) *validateError {
		isScheduler := rType == mindxdlv1.MindSporeReplicaTypeScheduler
		resReq := getContainerResourceReq(ct)
		if resReq == 0 && !isScheduler {
			return &validateError{
				reason:  "ContainerError",
				message: fmt.Sprintf("replicaType<%s> req npu<%d> is invalid, it can not be 0", rType, resReq),
			}
		}
		return nil
	}

	for _, container := range spec.Template.Spec.Containers {
		if container.Image == "" {
			return &validateError{
				reason: "ContainerError",
				message: fmt.Sprintf("replicaType is not valid: Image is undefined in the container of %v",
					rType),
			}
		}
		if container.Name != mindxdlv1.DefaultContainerName {
			continue
		}
		return validReq(container)
	}

	return &validateError{
		reason: "ContainerError",
		message: fmt.Sprintf("replicaType is not valid: There is no container named %s in %v",
			mindxdlv1.DefaultContainerName, rType),
	}
}

func getValidReplicaType(frame string) []commonv1.ReplicaType {
	switch frame {
	case mindxdlv1.MindSporeFrameworkName:
		return []commonv1.ReplicaType{
			mindxdlv1.MindSporeReplicaTypeScheduler,
			mindxdlv1.ReplicaTypeWorker,
		}
	case mindxdlv1.PytorchFrameworkName:
		return []commonv1.ReplicaType{
			mindxdlv1.PytorchReplicaTypeMaster,
			mindxdlv1.ReplicaTypeWorker,
		}
	case mindxdlv1.TensorflowFrameworkName:
		return []commonv1.ReplicaType{
			mindxdlv1.TensorflowReplicaTypeChief,
			mindxdlv1.ReplicaTypeWorker,
		}
	default:
		return nil
	}
}

func validateReplicaType(frame string, rType commonv1.ReplicaType) *validateError {
	replicaTypes := getValidReplicaType(frame)
	for _, t := range replicaTypes {
		if rType == t {
			return nil
		}
	}

	return &validateError{
		reason:  "ReplicaTypeError",
		message: fmt.Sprintf("replicaType is %v but must be one of %v", rType, replicaTypes),
	}
}

func validateLeader(rtype commonv1.ReplicaType, spec *commonv1.ReplicaSpec) *validateError {
	if spec.Replicas != nil && *spec.Replicas != 1 {
		return &validateError{
			reason:  "ReplicaTypeError",
			message: fmt.Sprintf("replicaType<%v> replicas is invalid, it must be only 1", rtype),
		}
	}
	return nil
}

func jobTotalRequest(specs map[commonv1.ReplicaType]*commonv1.ReplicaSpec) int {
	totalResRequest := 0
	for rType, value := range specs {
		if rType == mindxdlv1.ReplicaTypeWorker {
			totalResRequest += getReplicaSpecRequestRes(value) * int(specReplicas(value))
		}
	}
	return totalResRequest
}

func getReplicaSpecRequestRes(spec *commonv1.ReplicaSpec) int {
	for _, container := range spec.Template.Spec.Containers {
		if container.Name != mindxdlv1.DefaultContainerName {
			continue
		}
		return getContainerResourceReq(container)
	}
	return 0
}
