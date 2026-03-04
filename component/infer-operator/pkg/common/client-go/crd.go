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

package util

import (
	"context"
	"fmt"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apiv1 "infer-operator/pkg/api/v1"
)

// CRDExists checks if a CRD exists
func CRDExists(ctx context.Context, reader client.Reader, name string) error {
	crd := &v1.CustomResourceDefinition{}
	if err := reader.Get(ctx, client.ObjectKey{Name: name}, crd); err != nil {
		return err
	}

	for _, cond := range crd.Status.Conditions {
		if cond.Type == v1.Established && cond.Status == v1.ConditionTrue {
			return nil
		}
	}
	return fmt.Errorf("crd %s is not established", name)
}

// OwnedByGVK checks if an object is owned by a specific GVK
func OwnedByGVK(obj client.Object, targetGVK schema.GroupVersionKind) bool {
	refs := obj.GetOwnerReferences()
	if len(refs) == 0 {
		return false
	}
	for _, ref := range refs {
		refGV, err := schema.ParseGroupVersion(ref.APIVersion)
		if err != nil {
			continue
		}
		if refGV.Group == targetGVK.Group &&
			refGV.Version == targetGVK.Version &&
			ref.Kind == targetGVK.Kind {
			return true
		}
	}
	return false
}

// GetInferServiceGVK returns the GVK for InferService
func GetInferServiceGVK() schema.GroupVersionKind {
	return schema.FromAPIVersionAndKind(apiv1.GroupVersion.String(), "InferService")
}
