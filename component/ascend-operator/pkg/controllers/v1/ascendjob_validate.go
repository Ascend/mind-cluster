/*
Copyright(C) 2023. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package controllers is using for reconcile AscendJob.
*/

package controllers

import (
	"context"
	"errors"
	"fmt"

	commonv1 "github.com/kubeflow/common/pkg/apis/common/v1"
	"huawei.com/npu-exporter/v5/common-utils/hwlog"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	mindxdlv1 "ascend-operator/pkg/api/v1"
)

func (r *ASJobReconciler) validateV1ReplicaSpecs(job *mindxdlv1.AscendJob,
	specs map[commonv1.ReplicaType]*commonv1.ReplicaSpec) error {
	if r == nil {
		return errors.New("nil pointer")
	}

	frame, err := mindxdlv1.GetJobFramework(job)
	if err != nil {
		hwlog.RunLog.Errorf("AscendJob<%s> framework label is not set, err: %s", job.Name, err)
		r.recorder.Event(job, corev1.EventTypeWarning, "FrameworkLabelError", "framework label is not set")
		return err
	}

	if specs == nil {
		errMsg := "jobspec is not valid"
		r.recorder.Event(job, corev1.EventTypeWarning, "SpecsError", errMsg)
		return fmt.Errorf(errMsg)
	}

	if job.Spec.SuccessPolicy != nil &&
		*job.Spec.SuccessPolicy != mindxdlv1.SuccessPolicyDefault &&
		*job.Spec.SuccessPolicy != mindxdlv1.SuccessPolicyAllWorkers {
		err = fmt.Errorf(`job success policy is invalid, it must be one of <"", AllWorkers>`)
		hwlog.RunLog.Errorf("AscendJob<%s> success policy is invalid", job.Name)
		r.recorder.Event(job, corev1.EventTypeWarning, "SuccessPolicyError", err.Error())
		return err
	}

	if r.Config.EnableGangScheduling && job.Spec.RunPolicy.SchedulingPolicy != nil {
		queueName := job.Spec.RunPolicy.SchedulingPolicy.Queue
		if _, getErr := r.VolcanoClientSet.SchedulingV1beta1().Queues().Get(context.TODO(), queueName,
			metav1.GetOptions{}); getErr != nil {
			hwlog.RunLog.Errorf("get job<%s> queue failed", job.Name)
			r.recorder.Event(job, corev1.EventTypeWarning, "QueueGetFailed", getErr.Error())
			return getErr
		}
	}

	switch frame {
	case mindxdlv1.MindSporeFrameworkName:
		return r.validateMSReplicaSpecs(job, specs)
	case mindxdlv1.PytorchFrameworkName:
		return r.validatePTReplicaSpecs(job, specs)
	case mindxdlv1.TensorflowFrameworkName:
		return r.validateTFReplicaSpecs(job, specs)
	default:
		err = fmt.Errorf("framework<%s> is not supported, must be one of <mindspore, pytorch, tensorflow>", frame)
		r.recorder.Event(job, corev1.EventTypeWarning, "FrameworkLabelError", err.Error())
		return err
	}
}

func (r *ASJobReconciler) validateMSReplicaSpecs(job *mindxdlv1.AscendJob,
	specs map[commonv1.ReplicaType]*commonv1.ReplicaSpec) error {
	hwlog.RunLog.Debugf("validate framework<%s> replica specs", mindxdlv1.MindSporeFrameworkName)
	foundScheduler := 0
	totalResRequest := 0
	for rType, value := range specs {
		if value == nil || len(value.Template.Spec.Containers) == 0 {
			err := fmt.Errorf("JobSpec is not valid: containers definition expected in %v", rType)
			r.recorder.Event(job, corev1.EventTypeWarning, "ReplicaTypeError", err.Error())
			return err
		}

		validReplicaTypes := []commonv1.ReplicaType{
			mindxdlv1.MindSporeReplicaTypeScheduler,
			mindxdlv1.ReplicaTypeWorker,
		}
		isValidReplicaType := false
		for _, t := range validReplicaTypes {
			if rType == t {
				isValidReplicaType = true
				break
			}
		}
		if !isValidReplicaType {
			err := fmt.Errorf("mindspore replicaType is %v but must be one of %v", rType, validReplicaTypes)
			r.recorder.Event(job, corev1.EventTypeWarning, "ReplicaTypeError", err.Error())
			return err
		}
		if rType == mindxdlv1.MindSporeReplicaTypeScheduler {
			foundScheduler++
			if value.Replicas != nil && *value.Replicas != 1 {
				err := fmt.Errorf("mindspore replicaType<%v> replicas is invalid, it must be only 1", rType)
				r.recorder.Event(job, corev1.EventTypeWarning, "ReplicaTypeError", err.Error())
				return err
			}
		}

		replicas := int32(0)
		if value.Replicas == nil {
			replicas = 1
		} else {
			replicas = *value.Replicas
		}
		// Make sure the image is defined in the container.
		numNamedMindSpore := 0
		resReq := 0
		for _, container := range value.Template.Spec.Containers {
			if container.Image == "" {
				err := fmt.Errorf("JobSpec is not valid: Image is undefined in the container of %v", rType)
				r.recorder.Event(job, corev1.EventTypeWarning, "ContainerError", err.Error())
				return err
			}
			if container.Name != mindxdlv1.DefaultContainerName {
				continue
			}
			numNamedMindSpore++
			resReq = getContainerResourceReq(container)
			if resReq != 0 && rType == mindxdlv1.MindSporeReplicaTypeScheduler {
				err := fmt.Errorf("mindspore replicaType<%s> req npu<%d> is invalid, it must be 0", rType, resReq)
				r.recorder.Event(job, corev1.EventTypeWarning, "ReplicaTypeError", err.Error())
				return err
			}
			if resReq == 0 && rType == mindxdlv1.ReplicaTypeWorker {
				err := fmt.Errorf("mindspore replicaType<%s> req npu<%d> is invalid", rType, resReq)
				r.recorder.Event(job, corev1.EventTypeWarning, "ReplicaTypeError", err.Error())
				return err
			}

		}
		totalResRequest += resReq * int(replicas)
		if numNamedMindSpore == 0 {
			err := fmt.Errorf("mindspore replicaType is not valid: There is no container named %s in %v",
				mindxdlv1.DefaultContainerName, rType)
			r.recorder.Event(job, corev1.EventTypeWarning, "ReplicaTypeError", err.Error())
			return err
		}
	}
	if foundScheduler > 1 {
		err := fmt.Errorf("mindspore replicaType is not valid: %d Scheduler found, it must be 1", foundScheduler)
		r.recorder.Event(job, corev1.EventTypeWarning, "ReplicaTypeError", err.Error())
		return err
	}

	if foundScheduler == 0 && totalResRequest > 1 {
		err := fmt.Errorf("mindspore replicaType is not valid: %d schdeuler found, "+
			"but need 1 while req npu more than 1", foundScheduler)
		r.recorder.Event(job, corev1.EventTypeWarning, "ReplicaTypeError", err.Error())
		return err
	}

	return nil
}

func (r *ASJobReconciler) validatePTReplicaSpecs(job *mindxdlv1.AscendJob,
	specs map[commonv1.ReplicaType]*commonv1.ReplicaSpec) error {
	hwlog.RunLog.Debugf("validate framework<%s> replica specs", mindxdlv1.PytorchFrameworkName)

	foundMaster := 0
	for rType, value := range specs {
		if value == nil || len(value.Template.Spec.Containers) == 0 {
			err := fmt.Errorf("JobSpec is not valid: containers definition expected in %v", rType)
			r.recorder.Event(job, corev1.EventTypeWarning, "ReplicaTypeError", err.Error())
			return err
		}
		validReplicaTypes := []commonv1.ReplicaType{
			mindxdlv1.PytorchReplicaTypeMaster,
			mindxdlv1.ReplicaTypeWorker,
		}
		isValidReplicaType := false
		for _, t := range validReplicaTypes {
			if t == rType {
				isValidReplicaType = true
				break
			}
		}

		if !isValidReplicaType {
			err := fmt.Errorf("pytorch replicaType is %v but must be one of %v", rType, validReplicaTypes)
			r.recorder.Event(job, corev1.EventTypeWarning, "ReplicaTypeError", err.Error())
			return err
		}

		if rType == mindxdlv1.PytorchReplicaTypeMaster {
			foundMaster++
			if value.Replicas != nil && *value.Replicas != 1 {
				err := fmt.Errorf("pytorch replicaType<%v> replicas is invalid, it must be only 1", rType)
				r.recorder.Event(job, corev1.EventTypeWarning, "ReplicaTypeError", err.Error())
				return err
			}
		}

		defaultContainerPresent := false
		for _, container := range value.Template.Spec.Containers {
			if container.Image == "" {
				err := fmt.Errorf("pytorch replicaType is not valid: Image is undefined in the container of %s",
					rType)
				r.recorder.Event(job, corev1.EventTypeWarning, "ContainerError", err.Error())
				return err
			}
			if container.Name != mindxdlv1.DefaultContainerName {
				continue
			}
			defaultContainerPresent = true
			resReq := getContainerResourceReq(container)
			if resReq == 0 {
				err := fmt.Errorf("pytorch replicaType<%s> req npu<%d> is invalid, it can not be 0", rType, resReq)
				r.recorder.Event(job, corev1.EventTypeWarning, "ContainerError", err.Error())
				return err
			}
		}
		if !defaultContainerPresent {
			err := fmt.Errorf("pytorch ReplicaType is not valid: there is no container named %s in %s",
				mindxdlv1.DefaultContainerName, rType)
			r.recorder.Event(job, corev1.EventTypeWarning, "ReplicaTypeError", err.Error())
			return err
		}
	}
	if foundMaster != 1 {
		err := fmt.Errorf("pytorch ReplicaType is not valid: there must be only 1 Master replica-type")
		r.recorder.Event(job, corev1.EventTypeWarning, "ReplicaTypeError", err.Error())
		return err
	}
	return nil
}

func (r *ASJobReconciler) validateTFReplicaSpecs(job *mindxdlv1.AscendJob,
	specs map[commonv1.ReplicaType]*commonv1.ReplicaSpec) error {
	hwlog.RunLog.Debugf("validate framework<%s> replica specs", mindxdlv1.TensorflowFrameworkName)

	foundChief := 0
	for rType, value := range specs {
		if value == nil || len(value.Template.Spec.Containers) == 0 {
			err := fmt.Errorf("tensorflow replicaType is not valid: containers definition expected in %v", rType)
			r.recorder.Event(job, corev1.EventTypeWarning, "ReplicaTypeError", err.Error())
			return err
		}

		validReplicaTypes := []commonv1.ReplicaType{
			mindxdlv1.TensorflowReplicaTypeChief,
			mindxdlv1.ReplicaTypeWorker,
		}
		isValidReplicaType := false
		for _, t := range validReplicaTypes {
			if t == rType {
				isValidReplicaType = true
				break
			}
		}

		if !isValidReplicaType {
			err := fmt.Errorf("tensorflow replicaType is %v but must be one of %v", rType, validReplicaTypes)
			r.recorder.Event(job, corev1.EventTypeWarning, "ReplicaTypeError", err.Error())
			return err
		}

		if rType == mindxdlv1.TensorflowReplicaTypeChief {
			foundChief++
			if value.Replicas != nil && *value.Replicas != 1 {
				err := fmt.Errorf("tensorflow replicaType<%v> replicas is invalid, it must be only 1", rType)
				r.recorder.Event(job, corev1.EventTypeWarning, "ReplicaTypeError", err.Error())
				return err
			}
		}
		// Make sure the image is defined in the container.
		defaultContainerPresent := false
		for _, container := range value.Template.Spec.Containers {
			if container.Image == "" {
				err := fmt.Errorf("tensorflow replicaType is not valid: Image is undefined in the container of %v",
					rType)
				r.recorder.Event(job, corev1.EventTypeWarning, "ContainerError", err.Error())
				return err
			}
			if container.Name != mindxdlv1.DefaultContainerName {
				continue
			}
			defaultContainerPresent = true
			resReq := getContainerResourceReq(container)
			if resReq == 0 {
				err := fmt.Errorf("tensorflow replicaType<%s> req npu<%d> is invalid, it can not be 0", rType, resReq)
				r.recorder.Event(job, corev1.EventTypeWarning, "ContainerError", err.Error())
				return err
			}
		}
		// Make sure there has at least one container named "tensorflow".
		if !defaultContainerPresent {
			err := fmt.Errorf("tensorflow replicaType is not valid: There is no container named %s in %v",
				mindxdlv1.DefaultContainerName, rType)
			r.recorder.Event(job, corev1.EventTypeWarning, "ReplicaTypeError", err.Error())
			return err
		}
	}
	if foundChief != 1 {
		err := fmt.Errorf("tensorflow replicaType is not valid: there must be only 1 Chief replica-type")
		r.recorder.Event(job, corev1.EventTypeWarning, "ReplicaTypeError", err.Error())
		return err
	}
	return nil
}
