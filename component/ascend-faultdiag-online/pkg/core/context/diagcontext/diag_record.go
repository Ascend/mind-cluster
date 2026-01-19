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
Package diagcontext 包提供了与监控和诊断相关的上下文信息。
*/
package diagcontext

import (
	"sync"

	"ascend-faultdiag-online/pkg/utils/slicetool"
)

const maXDiagItemRecordsSize = 100

// DiagRecordItem 诊断异常记录结构体
type DiagRecordItem struct {
	DiagItem    *DiagItem        // 诊断项
	DiagRecords []*MetricDiagRes // 历史记录
}

// DiagRecordStore 诊断异常存储结构体
type DiagRecordStore struct {
	diagRecordItemMap map[string]*DiagRecordItem // 诊断记录项
	mu                sync.RWMutex               // 读写锁，保证并发安全
}

// NewDiagRecordStore 创建一个新的诊断记录存储实例。
func NewDiagRecordStore() *DiagRecordStore {
	return &DiagRecordStore{diagRecordItemMap: make(map[string]*DiagRecordItem)}
}

// UpdateRecord 更新诊断记录
func (store *DiagRecordStore) UpdateRecord(diagItem *DiagItem, diagResList []*MetricDiagRes) {
	if store == nil {
		return
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	abnormalRes := slicetool.Filter(diagResList, func(res *MetricDiagRes) bool {
		return res.IsAbnormal
	})
	if len(abnormalRes) == 0 {
		return
	}
	recordItem, ok := store.diagRecordItemMap[diagItem.Name]
	if !ok {
		recordItem = &DiagRecordItem{
			DiagItem:    diagItem,
			DiagRecords: make([]*MetricDiagRes, 0),
		}
	}
	if len(recordItem.DiagRecords)+len(abnormalRes) > maXDiagItemRecordsSize {
		overflowSize := len(recordItem.DiagRecords) + len(abnormalRes) - maXDiagItemRecordsSize
		recordItem.DiagRecords = recordItem.DiagRecords[overflowSize:len(recordItem.DiagRecords)]
	}
	recordItem.DiagRecords = append(recordItem.DiagRecords, abnormalRes...)
}

// GetRecordsByDiagName get the diag results by the diag name
func (store *DiagRecordStore) GetRecordsByDiagName(diagName string) []*MetricDiagRes {
	diagRecordItem, ok := store.diagRecordItemMap[diagName]
	if !ok {
		return nil
	}
	return diagRecordItem.DiagRecords
}
