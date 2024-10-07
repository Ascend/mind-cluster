/*
Copyright(C)2020-2023. Huawei Technologies Co.,Ltd. All rights reserved.

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
Package util is using for the total variable.
*/
package util

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type IsConfigMapChangedArgs struct {
	k8s       kubernetes.Interface
	cm        *v1.ConfigMap
	cmName    string
	nameSpace string
}

type IsConfigMapChangedTest struct {
	name string
	args IsConfigMapChangedArgs
	want bool
}

func buildIsConfigMapChangedTestCase() []IsConfigMapChangedTest {
	tests := []IsConfigMapChangedTest{
		{
			name: "01-GetConfigMapWithRetry will return true when cm is not in a session",
			args: IsConfigMapChangedArgs{k8s: nil, nameSpace: "vcjob", cmName: "cm01"},
			want: true,
		},
	}
	return tests
}

func TestIsConfigMapChanged(t *testing.T) {
	config := &rest.Config{}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return
	}
	tests := buildIsConfigMapChangedTestCase()
	for _, tt := range tests {
		tt.args.k8s = client
		t.Run(tt.name, func(t *testing.T) {
			if got := IsConfigMapChanged(tt.args.k8s, tt.args.cm, tt.args.cmName, tt.args.nameSpace); got != tt.want {
				t.Errorf("IsConfigMapChanged() = %v, want %v", got, tt.want)
			}
		})
	}
}

type CreateOrUpdateConfigMapArgs struct {
	k8s       kubernetes.Interface
	cm        *v1.ConfigMap
	cmName    string
	nameSpace string
}

type CreateOrUpdateConfigMapTest struct {
	name    string
	args    CreateOrUpdateConfigMapArgs
	err     metav1.StatusReason
	wantErr bool
}

func buildCreateOrUpdateConfigMapTestCase() []CreateOrUpdateConfigMapTest {
	tests := []CreateOrUpdateConfigMapTest{
		{
			name:    "01-GetConfigMapWithRetry will return true when cm is not in a session",
			args:    CreateOrUpdateConfigMapArgs{k8s: nil, nameSpace: "vcjob", cmName: "cm01", cm: &v1.ConfigMap{}},
			err:     "AlreadyExists",
			wantErr: true,
		},
		{
			name:    "02-GetConfigMapWithRetry will return true when cm is not in a session",
			args:    CreateOrUpdateConfigMapArgs{k8s: nil, nameSpace: "vcjob", cmName: "cm01", cm: &v1.ConfigMap{}},
			err:     "",
			wantErr: true,
		},
	}
	return tests
}

func TestCreateOrUpdateConfigMap(t *testing.T) {
	config := &rest.Config{}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return
	}
	tests := buildCreateOrUpdateConfigMapTestCase()
	for _, tt := range tests {
		tt.args.k8s = client
		tt.args.cm.ObjectMeta.Namespace = tt.args.nameSpace

		t.Run(tt.name, func(t *testing.T) {

			patch := gomonkey.ApplyFunc(errors.ReasonForError, func(err error) metav1.StatusReason {
				return tt.err
			})

			defer patch.Reset()

			if err := CreateOrUpdateConfigMap(tt.args.k8s, tt.args.cm, tt.args.cmName,
				tt.args.nameSpace); (err != nil) != tt.wantErr {
				t.Errorf("CreateOrUpdateConfigMap() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type UpdateConfigmapIncrementallyArgs struct {
	kubeClient kubernetes.Interface
	ns         string
	name       string
	newData    map[string]string
}

type UpdateConfigmapIncrementallyTest struct {
	name    string
	args    UpdateConfigmapIncrementallyArgs
	cm      *v1.ConfigMap
	err     error
	want    map[string]string
	wantErr bool
}

func buildUpdateConfigmapIncrementallyTestCase01() UpdateConfigmapIncrementallyTest {
	test01 := UpdateConfigmapIncrementallyTest{
		name:    "01-UpdateConfigmapIncrementally will return err when newDate is nil",
		args:    UpdateConfigmapIncrementallyArgs{},
		cm:      nil,
		want:    nil,
		wantErr: true,
	}
	return test01
}

func buildUpdateConfigmapIncrementallyTestCase02() UpdateConfigmapIncrementallyTest {
	test02 := UpdateConfigmapIncrementallyTest{
		name:    "02-UpdateConfigmapIncrementally will return err when get configMap failed",
		args:    UpdateConfigmapIncrementallyArgs{newData: map[string]string{"ascend01": "data01"}},
		cm:      &v1.ConfigMap{Data: map[string]string{"ascend": "data"}},
		err:     fmt.Errorf("config map already exist"),
		want:    map[string]string{"ascend01": "data01"},
		wantErr: true,
	}
	return test02
}

func buildUpdateConfigmapIncrementallyTestCase03() UpdateConfigmapIncrementallyTest {
	test02 := UpdateConfigmapIncrementallyTest{
		name:    "03-UpdateConfigmapIncrementally will return nil when newDate is nil",
		args:    UpdateConfigmapIncrementallyArgs{newData: map[string]string{"ascend01": "data01"}},
		cm:      &v1.ConfigMap{Data: map[string]string{"ascend": "data"}},
		want:    map[string]string{"ascend": "data", "ascend01": "data01"},
		wantErr: false,
	}
	return test02
}

func buildUpdateConfigmapIncrementallyTestCase() []UpdateConfigmapIncrementallyTest {
	tests := []UpdateConfigmapIncrementallyTest{
		buildUpdateConfigmapIncrementallyTestCase01(),
		buildUpdateConfigmapIncrementallyTestCase02(),
		buildUpdateConfigmapIncrementallyTestCase03(),
	}
	return tests
}

func TestUpdateConfigmapIncrementally(t *testing.T) {
	tests := buildUpdateConfigmapIncrementallyTestCase()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			patch := gomonkey.ApplyFunc(GetConfigMapWithRetry, func(client kubernetes.Interface,
				namespace string, cmName string) (*v1.ConfigMap, error) {
				return tt.cm, tt.err
			})
			defer patch.Reset()

			got, err := UpdateConfigmapIncrementally(tt.args.kubeClient, tt.args.ns, tt.args.name, tt.args.newData)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateConfigmapIncrementally() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UpdateConfigmapIncrementally() got = %v, want %v", got, tt.want)
			}
		})
	}
}
