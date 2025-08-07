/* Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Pakcage cluster is a DT collection for func in job_summary
package cluster

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"ascend-faultdiag-online/pkg/model/slownode"
)

func TestConvertCMToJobSummarySuccess(t *testing.T) {
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Namespace: "test-ns", Name: "test-cm"},
		Data: map[string]string{
			"job_id":     "job-123",
			"job_name":   "test-job",
			"job_status": "running",
			"hccl.json": `{
				"server_list": [
					{
						"server_id": "192.168.1.1",
						"server_sn": "sn-001",
						"device": [{"rank_id": "0"},{"rank_id": "1"}]
					},
					{
						"server_id": "192.168.1.2",
						"server_sn": "sn-002",
						"device": [{"rank_id": "2"}]
					}
				]
			}`,
		},
	}
	want := slownode.JobSummary{
		Namespace: "test-ns",
		JobStatus: "running",
		Servers: []slownode.Server{
			{
				Sn:      "sn-001",
				Ip:      "192.168.1.1",
				RankIds: []string{"0", "1"},
			},
			{
				Sn:      "sn-002",
				Ip:      "192.168.1.2",
				RankIds: []string{"2"},
			},
		},
	}
	want.JobId = "job-123"
	want.JobName = "test-job"

	got, err := convertCMToJobSummary(configMap)
	assert.Nil(t, err)
	assert.Equal(t, want, *got)
}

func TestConvertCMToJobSummaryMissingRequiredFields(t *testing.T) {
	tests := []struct {
		name        string
		configMap   *corev1.ConfigMap
		errContains string
	}{
		{
			name: "Missing jobId",
			configMap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Namespace: "test-ns", Name: "test-cm"},
				Data:       map[string]string{"job_name": "test-job", "job_status": "running", "hccl.json": "{}"},
			},
			errContains: "ConfigMap test-ns/test-cm does not contain job_id",
		},
		{
			name: "Missing jobName",
			configMap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Namespace: "test-ns", Name: "test-cm"},
				Data:       map[string]string{"job_id": "job-123", "job_status": "running", "hccl.json": "{}"},
			},
			errContains: "ConfigMap test-ns/test-cm does not contain job_name",
		},
		{
			name: "Missing jobStatus",
			configMap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Namespace: "test-ns", Name: "test-cm"},
				Data:       map[string]string{"job_id": "job-123", "job_name": "test-job", "hccl.json": "{}"},
			},
			errContains: "ConfigMap test-ns/test-cm does not contain job_status",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := convertCMToJobSummary(tt.configMap)
			assert.NotNil(t, err)
			assert.Contains(t, err.Error(), tt.errContains)
		})
	}
}

func TestConvertCMToJobSummaryHCCLJSONErrors(t *testing.T) {
	tests := []struct {
		name        string
		configMap   *corev1.ConfigMap
		errContains string
	}{
		{
			name: "Invalid HCCL JSON",
			configMap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test-ns",
					Name:      "test-cm",
				},
				Data: map[string]string{
					"job_id":     "job-123",
					"job_name":   "test-job",
					"job_status": "running",
					"hccl.json":  `}`,
				},
			},
			errContains: "failed to unmarshal HCCL data",
		},
		{
			name: "Empty HCCL data",
			configMap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test-ns",
					Name:      "test-cm",
				},
				Data: map[string]string{
					"job_id":     "job-123",
					"job_name":   "test-job",
					"job_status": "running",
					"hccl.json":  "",
				},
			},
			errContains: "ConfigMap test-ns/test-cm does not contain hccl.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := convertCMToJobSummary(tt.configMap)
			assert.NotNil(t, err)
			assert.Contains(t, err.Error(), tt.errContains)
		})
	}
}
