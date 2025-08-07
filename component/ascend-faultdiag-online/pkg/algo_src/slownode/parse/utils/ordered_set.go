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

// Package utils provides some common utils
package utils

// OrderedIDSet 有序ID集合，支持按顺序存储和随机删除
type OrderedIDSet struct {
	ids   []int64         // 按顺序存储ID的切片
	idMap map[int64]int64 // 记录ID到切片索引的映射
}

// NewOrderedIDSet 创建新的有序ID集合
func NewOrderedIDSet() *OrderedIDSet {
	return &OrderedIDSet{
		ids:   make([]int64, 0),
		idMap: make(map[int64]int64),
	}
}

// Add 添加ID到集合（保持顺序，重复ID会被忽略）
func (s *OrderedIDSet) Add(id int64) bool {
	// 检查ID是否已存在
	if _, exists := s.idMap[id]; exists {
		return false
	}

	// 添加到切片末尾并更新映射
	s.ids = append(s.ids, id)
	s.idMap[id] = int64(len(s.ids) - 1)
	return true
}

// Remove 随机删除指定ID（保持剩余元素顺序）（传入值，通过值删除）
func (s *OrderedIDSet) Remove(id int64) bool {
	// 检查ID是否存在
	idx, exists := s.idMap[id]
	if !exists {
		return false
	}

	s.ids = append(s.ids[:idx], s.ids[idx+1:]...)
	delete(s.idMap, id)

	return true
}

// GetByIndex 按索引获取ID（保持顺序）(索引 --> 值)
func (s *OrderedIDSet) GetByIndex(index int64) (int64, bool) {
	if index < 0 || index >= int64(len(s.ids)) {
		return 0, false
	}
	return s.ids[index], true
}

// Contains 检查ID是否存在
func (s *OrderedIDSet) Contains(id int64) bool {
	_, exists := s.idMap[id]
	return exists
}

// Size 获取集合大小
func (s *OrderedIDSet) Size() int64 {
	return int64(len(s.ids))
}

// ToSlice 转换为切片（保持顺序）
func (s *OrderedIDSet) ToSlice() []int64 {
	result := make([]int64, len(s.ids))
	copy(result, s.ids)
	return result
}

// Clear 清空集合
func (s *OrderedIDSet) Clear() {
	s.ids = s.ids[:0]
	s.idMap = make(map[int64]int64)
}
