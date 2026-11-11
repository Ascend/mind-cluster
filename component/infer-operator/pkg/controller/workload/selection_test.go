/*
Copyright(C) 2026. Huawei Technologies CO.,Ltd. All rights reserved.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"infer-operator/pkg/common"
)

// fakeWorkLoad implements WorkLoadInterface for pure function unit tests.
// It does not depend on a K8s client or informer.
type fakeWorkLoad struct {
	objectMeta metav1.ObjectMeta
	ready      bool
}

func (f *fakeWorkLoad) GetWorkLoadObjMeta() metav1.ObjectMeta {
	return f.objectMeta
}

func (f *fakeWorkLoad) SetWorkLoadObjMeta(objectMeta metav1.ObjectMeta) {
	f.objectMeta = objectMeta
}

func (f *fakeWorkLoad) GetWorkLoadReplicas() int32 {
	return common.DefaultReplicas
}

func (f *fakeWorkLoad) IsWorkLoadReady() bool {
	return f.ready
}

// readyWorkLoad builds a ready instance, index is the value of the instance index label.
func readyWorkLoad(index string) *fakeWorkLoad {
	return &fakeWorkLoad{
		objectMeta: metav1.ObjectMeta{
			Name:   "workload-" + index,
			Labels: map[string]string{common.InstanceIndexLabelKey: index},
		},
		ready: true,
	}
}

// unreadyWorkLoad builds an unready instance.
func unreadyWorkLoad(index string) *fakeWorkLoad {
	return &fakeWorkLoad{
		objectMeta: metav1.ObjectMeta{
			Name:   "workload-" + index,
			Labels: map[string]string{common.InstanceIndexLabelKey: index},
		},
		ready: false,
	}
}

// invalidWorkLoad builds an instance whose index label is missing or illegal.
func invalidWorkLoad(index string, withLabel bool) *fakeWorkLoad {
	meta := metav1.ObjectMeta{Name: "workload-invalid"}
	if withLabel {
		meta.Labels = map[string]string{common.InstanceIndexLabelKey: index}
	}
	return &fakeWorkLoad{objectMeta: meta, ready: true}
}
