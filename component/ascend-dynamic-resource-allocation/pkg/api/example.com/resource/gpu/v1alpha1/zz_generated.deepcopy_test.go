/*
 * Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * 		http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package v1alpha1

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNpuConfigDeepCopyObject(t *testing.T) {
	testCase := NpuConfig{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "example.com/v1alpha1",
			Kind:       "NpuConfig",
		},
		Sharing: &NpuSharing{
			Strategy: TimeSlicingStrategy,
			TimeSlicingConfig: &TimeSlicingConfig{
				Interval: DefaultTimeSlice,
			},
		},
	}
	copyObject := testCase.DeepCopyObject()

	copyObject, ok := copyObject.(*NpuConfig)
	if !ok {
		t.Errorf("Cannot convert object from Copied object")
	}
	if &testCase == copyObject {
		t.Error("testCase and copyObject should be different instances")
	}
	if !reflect.DeepEqual(&testCase, copyObject) {
		t.Error("testCase and copyObject should be deeply equal")
	}
}

func TestNpuSharingDeepCopy(t *testing.T) {
	testCase := &NpuSharing{
		Strategy: TimeSlicingStrategy,
		TimeSlicingConfig: &TimeSlicingConfig{
			Interval: DefaultTimeSlice,
		},
	}
	copyObject := testCase.DeepCopy()
	if testCase == copyObject {
		t.Error("testCase and copyObject should be different instances")
	}
	if !reflect.DeepEqual(testCase, copyObject) {
		t.Error("testCase and copyObject should be deeply equal")
	}
}

func TestSpacePartitioningDeepCopy(t *testing.T) {
	testCase := &SpacePartitioningConfig{
		PartitionCount: 1,
	}
	copyObject := testCase.DeepCopy()
	if testCase == copyObject {
		t.Error("testCase and copyObject should be different instances")
	}
	if !reflect.DeepEqual(testCase, copyObject) {
		t.Error("testCase and copyObject should be deeply equal")
	}
}

func TestTimeSlicingConfigDeepCopy(t *testing.T) {
	testCase := &TimeSlicingConfig{
		Interval: DefaultTimeSlice,
	}
	copyObject := testCase.DeepCopy()
	if testCase == copyObject {
		t.Error("testCase and copyObject should be different instances")
	}
	if !reflect.DeepEqual(testCase, copyObject) {
		t.Error("testCase and copyObject should be deeply equal")
	}
}
