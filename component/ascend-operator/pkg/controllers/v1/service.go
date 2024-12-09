/*
Copyright(C) 2023. Huawei Technologies Co.,Ltd. All rights reserved.

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

/*
Package controllers is using for reconcile AscendJob.
*/

package v1

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	commonv1 "github.com/kubeflow/common/pkg/apis/common/v1"
	"github.com/kubeflow/common/pkg/controller.v1/common"
	"github.com/kubeflow/common/pkg/util/labels"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"ascend-common/common-utils/hwlog"
	mindxdlv1 "ascend-operator/pkg/api/v1"
)

func (r *ASJobReconciler) genService(job metav1.Object, rtype commonv1.ReplicaType,
	spec *commonv1.ReplicaSpec) (*corev1.Service, error) {
	servicePorts, err := r.genServicePorts(spec)
	if err != nil {
		return nil, err
	}

	labels := r.genServiceLabels(job, rtype, "0")
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:   common.GenGeneralName(job.GetName(), strings.ToLower(string(rtype)), "0"),
			Labels: labels,
			OwnerReferences: []metav1.OwnerReference{
				*r.GenOwnerReference(job),
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports:    servicePorts,
		},
	}

	return service, nil
}

func (r *ASJobReconciler) genServicePorts(spec *commonv1.ReplicaSpec) ([]corev1.ServicePort, error) {
	ports, err := r.GetPortsFromJob(spec)
	if err != nil {
		return nil, err
	}
	var servicePorts []corev1.ServicePort
	// Add service ports to headless service
	for name, port := range ports {
		svcPort := corev1.ServicePort{Name: name, Port: port}
		servicePorts = append(servicePorts, svcPort)
	}
	return servicePorts, nil
}

func (r *ASJobReconciler) genServiceLabels(job metav1.Object, rtype commonv1.ReplicaType,
	index string) map[string]string {
	rt := strings.ToLower(string(rtype))
	// Append ReplicaTypeLabelDeprecated and ReplicaIndexLabelDeprecated labels.
	labelMap := r.GenLabels(job.GetName())
	labels.SetReplicaType(labelMap, rt)
	labels.SetReplicaIndexStr(labelMap, index)
	return labelMap
}

func (r *ASJobReconciler) getClusterDSvcIp() string {
	clusterdSvcIp := r.getIpFromSvcName(mindxServiceName, mindxServiceNamespace, mindxDefaultServerDomain)
	hwlog.RunLog.Infof("get ClusterD service ip = %s", clusterdSvcIp)
	return clusterdSvcIp
}

func (r *ASJobReconciler) getMngSvcIpAndPort(job *mindxdlv1.AscendJob, frame string,
	rtype commonv1.ReplicaType) (string, string, error) {
	if frame == mindxdlv1.MindSporeFrameworkName && len(job.Spec.ReplicaSpecs) == 1 &&
		rtype == mindxdlv1.ReplicaTypeWorker {
		return "", "", nil
	}

	svc, err := r.getOrCreateSvc(job)
	if err != nil {
		return "", "", err
	}
	svcIp, svcPort := getServiceIpAndPort(svc)
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

func (r *ASJobReconciler) getIpFromSvcName(svcName, svcNamespace, defaultDomain string) string {
	svcClient := r.KubeClientSet.CoreV1().Services(svcNamespace)
	service, err := svcClient.Get(context.Background(), svcName, metav1.GetOptions{})
	if err != nil {
		hwlog.RunLog.Warnf("get service cluster ip error: %v, use default service domain: %s", err, defaultDomain)
		return defaultDomain
	}
	return service.Spec.ClusterIP
}

func getServiceIpAndPort(service *corev1.Service) (string, string) {
	if service == nil {
		return "", ""
	}
	schedulerPort := ""
	for _, port := range service.Spec.Ports {
		if port.Name == mindxdlv1.DefaultPortName {
			schedulerPort = strconv.Itoa(int(port.Port))
			break
		}
	}
	return service.Spec.ClusterIP, schedulerPort
}
