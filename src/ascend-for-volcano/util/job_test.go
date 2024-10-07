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
	"reflect"
	"testing"

	"volcano.sh/volcano/pkg/scheduler/api"
)

type testIsSelectorMeetJobArgs struct {
	jobSelectors map[string]string
	conf         map[string]string
}

type testIsSelectorMeetJobTest struct {
	name string
	args testIsSelectorMeetJobArgs
	want bool
}

func buildTestIsSelectorMeetJobTest() []testIsSelectorMeetJobTest {
	tests := []testIsSelectorMeetJobTest{
		{
			name: "01-IsSelectorMeetJob nil jobSelector test.",
			args: testIsSelectorMeetJobArgs{jobSelectors: nil, conf: nil},
			want: true,
		},
		{
			name: "02-IsSelectorMeetJob conf no job selector test.",
			args: testIsSelectorMeetJobArgs{jobSelectors: map[string]string{"haha": "test"}, conf: nil},
			want: false,
		},
		{
			name: "03-IsSelectorMeetJob jobSelector no have conf test.",
			args: testIsSelectorMeetJobArgs{jobSelectors: map[string]string{"haha": "test"},
				conf: map[string]string{"haha": "what"}},
			want: false,
		},
	}
	return tests
}

// TestIsSelectorMeetJob test IsSelectorMeetJob.
func TestIsSelectorMeetJob(t *testing.T) {
	tests := buildTestIsSelectorMeetJobTest()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsSelectorMeetJob(tt.args.jobSelectors, tt.args.conf); got != tt.want {
				t.Errorf("IsSelectorMeetJob() = %v, want %v", got, tt.want)
			}
		})
	}
}

type testIsVJobTest struct {
	name string
	nJob *NPUJob
	want bool
}

func buildTestIsVJobTest() []testIsVJobTest {
	tests := []testIsVJobTest{
		{
			name: "01-IsVJob nJob.ReqNPUName nil test.",
			nJob: &NPUJob{},
			want: false,
		},
		{
			name: "02-IsVjob nJob.ReqNPUName > 2 test.",
			nJob: &NPUJob{
				ReqNPUName: AscendNPUCore,
			},
			want: true,
		},
	}
	return tests
}

func TestIsVJob(t *testing.T) {
	tests := buildTestIsVJobTest()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.nJob.IsVJob(); got != tt.want {
				t.Errorf("Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

type SetVJobTypeTest struct {
	name string
	nJob *NPUJob
	want int
}

func buildSetVJobTypeTestCase() []SetVJobTypeTest {
	return []SetVJobTypeTest{
		{
			name: "01-TestSetVJobType JobTypeUnknown status",
			nJob: &NPUJob{
				Tasks: map[api.TaskID]NPUTask{},
				VJob:  &VJob{},
			},
			want: JobTypeUnknown,
		},
		{
			name: "02-TestSetVJobType JobTypeWhole status",
			nJob: &NPUJob{
				Tasks: map[api.TaskID]NPUTask{"task01": {ReqNPUNum: 1, VTask: &VTask{Type: JobTypeWhole}}},
				VJob:  &VJob{},
			},
			want: JobTypeWhole,
		},
	}
}

func TestSetVJobType(t *testing.T) {
	tests := buildSetVJobTypeTestCase()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.nJob.SetJobType()
			if got := tt.nJob.Type; !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

type testSetVJobStatusByInfTest struct {
	name  string
	nJob  *NPUJob
	vcJob *api.JobInfo
	want  string
}

func buildTestSetVJobStatusByInfTest() []testSetVJobStatusByInfTest {
	tests := []testSetVJobStatusByInfTest{
		{
			name:  "01-test SetJobStatusByInf ",
			nJob:  &NPUJob{VJob: &VJob{}},
			vcJob: &api.JobInfo{PodGroup: &api.PodGroup{}},
			want:  "",
		},
	}
	return tests
}

func TestSetVJobStatusByInf(t *testing.T) {
	tests := buildTestSetVJobStatusByInfTest()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.nJob.SetJobStatusByInf(tt.vcJob)
			if got := tt.nJob.Status; got != tt.want {
				t.Errorf("Name() = %v, want %v", got, tt.want)
			}
		})
	}
}
