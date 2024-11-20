/*
Copyright(C) 2023. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package controllers is using for reconcile AscendJob.
*/

package controllers

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	commonv1 "github.com/kubeflow/common/pkg/apis/common/v1"
	"github.com/kubeflow/common/pkg/controller.v1/common"
	"github.com/kubeflow/common/pkg/controller.v1/expectation"
	commonutil "github.com/kubeflow/common/pkg/util"
	utillabels "github.com/kubeflow/common/pkg/util/labels"
	corev1 "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	mindxdlv1 "ascend-operator/pkg/api/v1"
)

// reconcileServices checks and updates services for each given ReplicaSpec.
// It will requeue the job in case of an error while creating/deleting services.
func (r *ASJobReconciler) ReconcileServices(
	job metav1.Object,
	services []*corev1.Service,
	rtype commonv1.ReplicaType,
	spec *commonv1.ReplicaSpec) error {
	if r == nil {
		return errors.New("nil pointer")
	}

	// Convert ReplicaType to lower string.
	rt := strings.ToLower(string(rtype))
	replicas := int(*spec.Replicas)
	// Get all services for the type rt.
	filterServices, err := r.FilterServicesForReplicaType(services, rt)
	if err != nil {
		return err
	}

	// GetServiceSlices will return enough information here to make decision to add/remove/update resources.
	//
	// For example, let's assume we have services with replica-index 0, 1, 2
	// If replica is 4, return a slice with size 4. [[0],[1],[2],[]], a svc with replica-index 3 will be created.
	//
	// If replica is 1, return a slice with size 3. [[0],[1],[2]], svc with replica-index 1 and 2 are out of range and will be deleted.
	serviceSlices := r.GetServiceSlices(filterServices, replicas, commonutil.LoggerForReplica(job, rt))

	for index, serviceSlice := range serviceSlices {
		switch {
		case len(serviceSlice) > 1:
			commonutil.LoggerForReplica(job, rt).Warningf("We have too many services for %s %d", rtype, index)
		case len(serviceSlice) == 0:
			commonutil.LoggerForReplica(job, rt).Infof("need to create new service: %s-%d", rtype, index)
			if err = r.CreateNewService(job, rtype, spec, strconv.Itoa(index)); err != nil {
				return err
			}
		default:
			// Check the status of the current svc.
			svc := serviceSlice[0]

			// check if the index is in the valid range, if not, we should kill the svc
			if index < 0 || index >= replicas {
				err = r.ServiceControl.DeleteService(svc.Namespace, svc.Name, job.(runtime.Object))
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// createNewService creates a new service for the given index and type.
func (r *ASJobReconciler) CreateNewService(job metav1.Object, rtype commonv1.ReplicaType,
	spec *commonv1.ReplicaSpec, index string) error {
	if r == nil {
		return errors.New("nil pointer")
	}

	jobKey, err := common.KeyFunc(job)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("couldn't get key for job object %#v: %v", job, err))
		return err
	}

	rt := strings.ToLower(string(rtype))
	// Append ReplicaTypeLabelDeprecated and ReplicaIndexLabelDeprecated labels.
	labels := r.GenLabels(job.GetName())
	utillabels.SetReplicaType(labels, rt)
	utillabels.SetReplicaIndexStr(labels, index)

	ports, err := r.GetPortsFromJob(spec)
	if err != nil {
		return err
	}

	service := &corev1.Service{
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports:    []corev1.ServicePort{},
		},
	}

	// Add service ports to headless service
	for name, port := range ports {
		svcPort := corev1.ServicePort{Name: name, Port: port}
		service.Spec.Ports = append(service.Spec.Ports, svcPort)
	}

	service.Name = common.GenGeneralName(job.GetName(), rt, index)
	service.Labels = labels
	// Create OwnerReference.
	controllerRef := r.GenOwnerReference(job)

	// Creation is expected when there is no error returned
	expectationServicesKey := expectation.GenExpectationServicesKey(jobKey, rt)
	r.Expectations.RaiseExpectations(expectationServicesKey, 1, 0)

	err = r.ServiceControl.CreateServicesWithControllerRef(job.GetNamespace(), service, job.(runtime.Object), controllerRef)
	if err != nil && k8serr.IsTimeout(err) {
		// Service is created but its initialization has timed out.
		// If the initialization is successful eventually, the
		// controller will observe the creation via the informer.
		// If the initialization fails, or if the service keeps
		// uninitialized for a long time, the informer will not
		// receive any update, and the controller will create a new
		// service when the expectation expires.
		return nil
	} else if err != nil {
		// Since error occurred(the informer won't observe this service),
		// we decrement the expected number of creates
		// and wait until next reconciliation
		r.Expectations.CreationObserved(expectationServicesKey)
		return err
	}
	return nil
}

func (r *ASJobReconciler) getMngSvcIpAndPort(job *mindxdlv1.AscendJob) (string, string, error) {
	services, err := r.GetServicesForJob(job)
	if err != nil || len(services) == 0 {
		return "", "", fmt.Errorf("get job<%s/%s> services failed", job.Namespace, job.Name)
	}

	mngSvc := r.getMangerSvc(services)
	if mngSvc == nil {
		return "", "", fmt.Errorf("get job<%s/%s> chief service failed", job.Namespace, job.Name)
	}

	svcIp, svcPort := getServiceIpAndPort(mngSvc)
	if svcIp == "" || svcPort == "" {
		return "", "", fmt.Errorf("job<%s/%s> chief service Ip<%s> or port<%s> is empty", job.Namespace, job.Name,
			svcIp, svcPort)
	}
	return svcIp, svcPort, nil
}

func (r *ASJobReconciler) getMangerSvc(services []*corev1.Service) *corev1.Service {
	for _, svc := range services {
		if label, ok := svc.Labels[commonv1.ReplicaTypeLabel]; ok &&
			(label == strings.ToLower(string(mindxdlv1.PytorchReplicaTypeMaster)) ||
				label == strings.ToLower(string(mindxdlv1.TensorflowReplicaTypeChief)) ||
				label == strings.ToLower(string(mindxdlv1.MindSporeReplicaTypeScheduler))) {
			return svc
		}
	}
	return nil
}

func getServiceIpAndPort(service *corev1.Service) (string, string) {
	schedulerPort := ""
	for _, port := range service.Spec.Ports {
		if port.Name == mindxdlv1.DefaultPortName {
			schedulerPort = strconv.Itoa(int(port.Port))
			break
		}
	}
	return service.Spec.ClusterIP, schedulerPort
}
