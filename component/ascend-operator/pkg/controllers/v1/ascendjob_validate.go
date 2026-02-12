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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"volcano.sh/apis/pkg/apis/scheduling/v1beta1"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	mindxdlv1 "ascend-operator/pkg/api/v1"
	"ascend-operator/pkg/utils"
)

func (r *ASJobReconciler) validateJob(job *mindxdlv1.AscendJob) *validateError {
	if r == nil {
		return &validateError{
			reason:  argumentErrorReason,
			message: "nil pointer",
		}
	}

	var err *validateError
	defer func() {
		if err != nil {
			r.recorder.Event(job, corev1.EventTypeWarning, err.reason, err.message)
		}
	}()

	if scaleError := r.scaler.ValidJob(job); scaleError != nil {
		err = &validateError{
			reason:  invalidScalingConfigReason,
			message: scaleError.Error(),
		}
		return err
	}

	// 910a5 branch check the scaleout-type label
	if scaleOutTypeError := utils.CheckAcJobScaleOutTypeLabel(job); scaleOutTypeError != nil {
		return &validateError{
			reason:  invalidScaleOutConfigReason,
			message: scaleOutTypeError.Error(),
		}
	}

	if err = r.validateBasicInfo(job); err != nil {
		return err
	}

	err = r.validateSpec(job, job.Spec.ReplicaSpecs)
	return err
}

func (r *ASJobReconciler) validateBasicInfo(job *mindxdlv1.AscendJob) *validateError {
	if job.Spec.ReplicaSpecs == nil {
		return &validateError{
			reason:  invalidSpecsReason,
			message: "replicaSpecs is not set, please modify your job",
		}
	}

	if job.Spec.SuccessPolicy != nil &&
		*job.Spec.SuccessPolicy != mindxdlv1.SuccessPolicyDefault &&
		*job.Spec.SuccessPolicy != mindxdlv1.SuccessPolicyAllWorkers {
		return &validateError{
			reason:  invalidSuccessPolicyReason,
			message: `job success policy is invalid, it must be one of <"", AllWorkers>`,
		}
	}

	if r.Config.EnableGangScheduling && job.Spec.RunPolicy.SchedulingPolicy != nil {
		queueName := job.Spec.RunPolicy.SchedulingPolicy.Queue
		if _, err := r.getQueueFromApiserver(queueName); err != nil {
			return &validateError{
				reason:  invalidQueueReason,
				message: fmt.Sprintf("check queue<%s> failed, err: %v, maybe it is not create", queueName, err),
			}
		}
	}

	return nil
}

func (r *ASJobReconciler) getQueueFromApiserver(queueName string) (*v1beta1.Queue, error) {
	return r.VolcanoClientSet.SchedulingV1beta1().Queues().Get(context.TODO(), queueName, metav1.GetOptions{})
}

func (r *ASJobReconciler) validateSpec(job *mindxdlv1.AscendJob,
	specs map[commonv1.ReplicaType]*commonv1.ReplicaSpec) *validateError {
	frame, err := mindxdlv1.GetJobFramework(job)
	if err != nil {
		return &validateError{
			reason:  invalidFrameworkReason,
			message: err.Error(),
		}
	}

	hwlog.RunLog.Debugf("validate framework<%s> replica specs", frame)
	return checkReplicaSpecs(frame, specs)
}

func validContainerNum(rType commonv1.ReplicaType, spec *commonv1.ReplicaSpec) *validateError {
	if spec == nil || len(spec.Template.Spec.Containers) == 0 {
		return &validateError{
			reason:  invalidReplicaSpecReason,
			message: fmt.Sprintf("%s replicaSpec is not valid: containers is undefined", rType),
		}
	}
	return nil
}

func validateReplicas(rtype commonv1.ReplicaType, spec *commonv1.ReplicaSpec) *validateError {
	if spec.Replicas == nil {
		return nil
	}
	if *spec.Replicas < 0 {
		return &validateError{
			reason: invalidReplicaSpecReason,
			message: fmt.Sprintf("%s replicaSpec is not valid: replicas can not be negative num, but got %d",
				rtype, *spec.Replicas),
		}
	}
	if *spec.Replicas > maxReplicas {
		return &validateError{
			reason: invalidReplicaSpecReason,
			message: fmt.Sprintf("%s replicaSpec is not valid: replicas can not be larger than %d, but got %d",
				rtype, maxReplicas, *spec.Replicas),
		}
	}
	return nil
}

func checkReplicaSpecs(frame string, specs map[commonv1.ReplicaType]*commonv1.ReplicaSpec) *validateError {
	hasLeader := false
	for rType, value := range specs {
		if ve := validateReplicaType(frame, rType); ve != nil {
			return ve
		}

		if value == nil {
			return &validateError{
				reason:  invalidReplicaSpecReason,
				message: fmt.Sprintf("%s replicaSpec is not set, please modify your job", rType),
			}
		}

		if ve := validateReplicas(rType, value); ve != nil {
			return ve
		}

		if ve := validContainerNum(rType, value); ve != nil {
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
				reason: invalidReplicaTypeReason,
				message: "replicaType is not valid: there need 1 leader replicaType, Master for pytorch," +
					" Chief of tensorflow",
			}
		}
		if jobTotalRequest(specs) > 1 {
			return &validateError{
				reason:  invalidReplicaSpecReason,
				message: "replicaSpec is not valid: when scheduler not found, the req num must be 1",
			}
		}
	}

	return nil
}

func validateContainer(rType commonv1.ReplicaType, spec *commonv1.ReplicaSpec) *validateError {
	hasDefaultContainer := false
	for _, container := range spec.Template.Spec.Containers {
		if container.Name != api.DefaultContainerName {
			continue
		}
		if container.Image == "" {
			return &validateError{
				reason: invalidContainerReason,
				message: fmt.Sprintf("%s replicaSpec is not valid: Image is undefined in the container of %s",
					rType, api.DefaultContainerName),
			}
		}

		hasDefaultContainer = true
	}

	if hasDefaultContainer {
		return nil
	}

	return &validateError{
		reason: invalidContainerReason,
		message: fmt.Sprintf("%s replicaSpec is not valid: There is no container named %s",
			rType, api.DefaultContainerName),
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
		reason:  invalidReplicaTypeReason,
		message: fmt.Sprintf("replicaType is %v but must be one of %v", rType, replicaTypes),
	}
}

func validateLeader(rtype commonv1.ReplicaType, spec *commonv1.ReplicaSpec) *validateError {
	if spec.Replicas != nil && *spec.Replicas != 1 {
		return &validateError{
			reason:  invalidReplicaSpecReason,
			message: fmt.Sprintf("%s replicaSpec is not valid, the replicas must be only 1", rtype),
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
		if container.Name != api.DefaultContainerName {
			continue
		}
		return getContainerResourceReq(container)
	}
	return 0
}
