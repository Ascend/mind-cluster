// Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.

// Package common a series of util function
package common

import (
	"bytes"
	"encoding/gob"
)

// DeepCopy for object using gob
// DeepCopy has performance problem, cannot use in Time-sensitive scenario
func DeepCopy(dst, src interface{}) error {
	if src == nil {
		return nil
	}
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(src); err != nil {
		return err
	}
	return gob.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(dst)
}
