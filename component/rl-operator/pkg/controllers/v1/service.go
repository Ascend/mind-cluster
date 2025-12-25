/*
Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"ascend-common/common-utils/hwlog"
	mindxdlv1 "rl-operator/pkg/api/v1"
	"rl-operator/pkg/common"
)

func (r *RayHeadReconciler) ReconcileService(rayHead *mindxdlv1.RayHead) error {
	name := common.GetServiceName(rayHead.GetNamespace(), rayHead.GetName())
	_, err := r.GetSvcFromApiserver(name, rayHead.GetNamespace())
	if err == nil {
		hwlog.RunLog.Debugf("get service %s/%s success", rayHead.GetNamespace(), name)
		return nil
	}

	if errors.IsNotFound(err) {
		newSvc, gerr := r.genService(rayHead)
		if gerr != nil {
			return gerr
		}
		_, err = r.CreateService(rayHead.GetNamespace(), newSvc)
		if err != nil {
			return err
		}
		hwlog.RunLog.Infof("create service %s/%s success", rayHead.GetNamespace(), name)
		return nil
	}
	return err
}

func (r *RayHeadReconciler) genService(rayHead *mindxdlv1.RayHead) (*corev1.Service, error) {
	servicePorts, err := r.genServicePorts(rayHead)
	if err != nil {
		return nil, err
	}

	label := r.GenLabels(rayHead, common.RayHeadIdentification)
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      common.GetServiceName(rayHead.Namespace, rayHead.Name),
			Namespace: rayHead.GetNamespace(),
			Labels:    label,
			OwnerReferences: []metav1.OwnerReference{
				*r.GenOwnerReference(rayHead),
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: label,
			Ports:    servicePorts,
		},
	}

	return service, nil
}

func (r *RayHeadReconciler) genServicePorts(rayHead *mindxdlv1.RayHead) ([]corev1.ServicePort, error) {
	labels := rayHead.GetLabels()
	var servicePorts []corev1.ServicePort
	// Add service ports to headless service
	for _, key := range common.ServicePortKeys {
		port, _ := strconv.Atoi(labels[key])
		key = strings.ToLower(key)
		svcPort := corev1.ServicePort{Name: key, Port: int32(port)}
		servicePorts = append(servicePorts, svcPort)
	}
	return servicePorts, nil
}
