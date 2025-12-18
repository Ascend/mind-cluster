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

// Package utils provides some utility functions for local storage
package utils

import (
	"sync"
)

// Storage is common struct for save the data
type Storage[T any] struct {
	data sync.Map
}

// Clear clear all the data by pointing to a new sync map
func (s *Storage[T]) Clear() {
	if s == nil {
		return
	}
	s.data = sync.Map{}
}

// Store store the data and value
func (s *Storage[T]) Store(key string, value T) {
	if s == nil {
		return
	}
	s.data.Store(key, value)
}

// Load load the data
func (s *Storage[T]) Load(key string) (T, bool) {
	var res T
	if s == nil {
		return res, false
	}
	data, ok := s.data.Load(key)
	if !ok {
		return res, false
	}
	res, ok = data.(T)
	return res, ok
}

func (s *Storage[T]) Delete(key string) {
	if s == nil {
		return
	}
	s.data.Delete(key)
}

func (s *Storage[T]) Len() int {
	length := 0
	if s == nil {
		return length
	}
	s.data.Range(func(key, value any) bool {
		length++
		return true
	})
	return length
}

func (s *Storage[T]) Range(f func(key string, value T) bool) {
	if s == nil {
		return
	}
	s.data.Range(func(key, value any) bool {
		k, ok := key.(string)
		if !ok {
			return true
		}
		v, ok := value.(T)
		if !ok {
			return true
		}
		return f(k, v)
	})
}

// NewStorage got a new storage instance
func NewStorage[T any]() *Storage[T] {
	return &Storage[T]{}
}
