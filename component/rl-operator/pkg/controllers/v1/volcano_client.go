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
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"volcano.sh/apis/pkg/apis/scheduling/v1beta1"
	volcanoclient "volcano.sh/apis/pkg/client/clientset/versioned"

	"ascend-common/common-utils/hwlog"
)

type VolcanoControlInterface interface {
	GetQueue(namespace string) (*v1beta1.Queue, error)
	GetPodGroup(namespace string, pgName string) (*v1beta1.PodGroup, error)
	CreatePodGroup(createPG *v1beta1.PodGroup) (*v1beta1.PodGroup, error)
}

type RealVolcanoControl struct {
	VolcanoClient volcanoclient.Interface
	Recorder      record.EventRecorder
}

func (r *RealVolcanoControl) GetQueue(queueName string) (*v1beta1.Queue, error) {
	return r.VolcanoClient.SchedulingV1beta1().Queues().Get(context.TODO(), queueName, metav1.GetOptions{})
}

func (r *RealVolcanoControl) GetPodGroup(namespace string, pgName string) (*v1beta1.PodGroup, error) {
	return r.VolcanoClient.SchedulingV1beta1().PodGroups(namespace).Get(context.TODO(), pgName,
		metav1.GetOptions{})
}

func (r *RealVolcanoControl) CreatePodGroup(
	createPG *v1beta1.PodGroup) (*v1beta1.PodGroup, error) {
	nameSpace := createPG.GetNamespace()
	createdPodGroup, err := r.VolcanoClient.SchedulingV1beta1().PodGroups(nameSpace).Create(context.TODO(),
		createPG, metav1.CreateOptions{})
	if err != nil {
		return createdPodGroup, fmt.Errorf("unable to create PodGroup: %v", err)
	}
	hwlog.RunLog.Infof("create podGroup %s/%s success", nameSpace, createdPodGroup.Name)
	return createdPodGroup, nil
}
