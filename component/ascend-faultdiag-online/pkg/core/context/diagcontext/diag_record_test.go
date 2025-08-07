/*
Copyright(C)2025. Huawei Technologies Co.,Ltd. All rights reserved.

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
Package diagcontext some test case for the diag record.
*/
package diagcontext

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"ascend-faultdiag-online/pkg/utils/slicetool"
)

func TestNewDiagRecordStore(t *testing.T) {
	assert.NotNil(t, NewDiagRecordStore())
}

func TestUpdateRecord(t *testing.T) {
	// 现有的诊断记录
	diagItem := &DiagItem{
		Name:           "npu check",
		Interval:       10,
		Rules:          []*DiagRule{},
		CustomRules:    []*CustomRule{},
		ConditionGroup: &ConditionGroup{},
		Description:    "npu chip diagnosis",
	}
	var diagRes []*MetricDiagRes
	nowDiagRecordCount := 80
	for i := 0; i < nowDiagRecordCount; i++ {
		diagRes = append(diagRes, &MetricDiagRes{})
	}
	store := &DiagRecordStore{
		diagRecordItemMap: map[string]*DiagRecordItem{
			"npu check": {diagItem, diagRes},
		},
	}
	assert.Equal(t, nowDiagRecordCount, len(store.diagRecordItemMap[diagItem.Name].DiagRecords))

	// 新增诊断记录中异常的诊断记录
	var diagResList []*MetricDiagRes
	abnormalCount := 40
	for i := 0; i < abnormalCount; i++ {
		diagResList = append(diagResList, &MetricDiagRes{IsAbnormal: true})
	}
	abnormalRes := slicetool.Filter(diagResList, func(res *MetricDiagRes) bool {
		return res.IsAbnormal
	})
	assert.Equal(t, abnormalCount, len(abnormalRes))

	// 更新后的诊断记录
	updateDiagRecordCount := 100
	store.UpdateRecord(diagItem, abnormalRes)
	// 现有的诊断记录清除overflowSize个后 + 新增诊断记录中异常的诊断记录。即：[20:80] + 40
	assert.Equal(t, updateDiagRecordCount, len(store.diagRecordItemMap[diagItem.Name].DiagRecords))

}
