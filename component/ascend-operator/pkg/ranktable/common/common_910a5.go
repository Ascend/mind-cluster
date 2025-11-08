/*
Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
*/

// Package common is common function or object of ranktable.
package common

// GetNeedGenerate get the state of needGenerate
func (r *BaseGenerator) GetNeedGenerate() bool {
	return r.needGenerate
}

// SetNeedGenerate set the state of needGenerate
func (r *BaseGenerator) SetNeedGenerate(needGenerate bool) {
	r.needGenerate = needGenerate
}
