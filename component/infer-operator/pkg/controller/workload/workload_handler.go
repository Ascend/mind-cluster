/*
Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.

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

package workload

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"infer-operator/pkg/api/v1"
	"infer-operator/pkg/common"
)

type WorkLoadHandlerFactory struct {
	workloadHandlerMap map[string]WorkLoadHandler
}

func NewWorkLoadHandlerFactory() *WorkLoadHandlerFactory {
	return &WorkLoadHandlerFactory{
		workloadHandlerMap: make(map[string]WorkLoadHandler),
	}
}

func (factory *WorkLoadHandlerFactory) Register(gvk schema.GroupVersionKind, handler WorkLoadHandler) error {
	if _, ok := factory.workloadHandlerMap[gvk.String()]; ok {
		return fmt.Errorf("duplicate workload handler for GVK %s", gvk)
	}
	factory.workloadHandlerMap[gvk.String()] = handler
	return nil
}

func (factory *WorkLoadHandlerFactory) GetWorkLoadHandler(gvk schema.GroupVersionKind) (WorkLoadHandler, error) {
	handler, ok := factory.workloadHandlerMap[gvk.String()]
	if !ok {
		return nil, fmt.Errorf("can not find workload handler for GVK %s", gvk)
	}
	return handler, nil
}

type WorkLoadHandler interface {
	// CheckOrCreateWorkLoad checks if the workload exists and creates it if not
	CheckOrCreateWorkLoad(ctx context.Context, instanceSet *v1.InstanceSet, indexer common.InstanceIndexer) error
	// DeleteExtraWorkLoad deletes workloads that exceed the specified index limit
	DeleteExtraWorkLoad(ctx context.Context, indexer common.InstanceIndexer, indexLimit int) error
	// GetWorkLoadReadyReplicas returns the number of ready replicas of the workload
	GetWorkLoadReadyReplicas(ctx context.Context, indexer common.InstanceIndexer) (int, error)
	// Validate checks if the workload specification is valid
	Validate(spec runtime.RawExtension) error
	// GetReplicas retrieves the number of replicas from the workload specification
	GetReplicas(spec runtime.RawExtension) (int32, error)
	// ListWorkLoad list workloads via selector with filter
	ListWorkLoad(ctx context.Context, selectLabels map[string]string, namespace string, filters ...WorkLoadFilter) ([]WorkLoadInterface, error)
	// DeleteWorkLoad fetches workloads match selector and deletes those filtered by filters
	DeleteWorkLoad(ctx context.Context, selectLabels map[string]string, namespace string, filters ...WorkLoadFilter) error
	// UpdateWorkLoad updates workloads match selector and filters with updater function
	UpdateWorkLoad(ctx context.Context, selectLabels map[string]string, namespace string, updater WorkloadUpdater, filters ...WorkLoadFilter) error
}

// WorkLoadInterface defines the interface for workload objects
type WorkLoadInterface interface {
	// GetWorkLoadObjMeta get the object meta of the workload
	GetWorkLoadObjMeta() metav1.ObjectMeta
	// SetWorkLoadObjMeta set the object meta of the workload
	SetWorkLoadObjMeta(metav1.ObjectMeta)
	// GetWorkLoadReplicas returns the number of ready replicas of the workload
	GetWorkLoadReplicas() int32
	// IsWorkLoadReady returns true if the workload is ready
	IsWorkLoadReady() bool
}

// WorkLoadFilter return true if the workload meets the filter condition
type WorkLoadFilter func(workLoad WorkLoadInterface) bool

// WorkloadUpdater updates the workload
type WorkloadUpdater func(workLoad WorkLoadInterface)
