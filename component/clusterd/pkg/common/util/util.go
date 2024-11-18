// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package util a series of util function
package util

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"os/signal"
	"strconv"

	"huawei.com/npu-exporter/v6/common-utils/hwlog"
)

// NewSignalWatcher create a new signal watcher
func NewSignalWatcher(signals ...os.Signal) chan os.Signal {
	signalChan := make(chan os.Signal, 1)
	for _, sign := range signals {
		signal.Notify(signalChan, sign)
	}
	return signalChan
}

// EqualDataHash get data hashcode and determine equal
func EqualDataHash(checkCode string, data interface{}) bool {
	if len(checkCode) == 0 {
		hwlog.RunLog.Error("checkCode is empty")
		return false
	}
	return MakeDataHash(data) == checkCode
}

// MakeDataHash get data hashcode
func MakeDataHash(data interface{}) string {
	var dataBuffer []byte
	if dataBuffer = marshalData(data); len(dataBuffer) == 0 {
		return ""
	}
	h := sha256.New()
	if _, err := h.Write(dataBuffer); err != nil {
		hwlog.RunLog.Errorf("hash data error: %v", err)
		return ""
	}
	sum := h.Sum(nil)
	return hex.EncodeToString(sum)
}

func marshalData(data interface{}) []byte {
	dataBuffer, err := json.Marshal(data)
	if err != nil {
		hwlog.RunLog.Errorf("marshal data err: %v", err)
		return nil
	}
	return dataBuffer
}

// ObjToString obj to string
func ObjToString(data interface{}) string {
	var dataBuffer []byte
	if dataBuffer = marshalData(data); len(dataBuffer) == 0 {
		return ""
	}
	return string(dataBuffer)
}

// RemoveSliceDuplicateElement remove duplicate element in slice
func RemoveSliceDuplicateElement(languages []string) []string {
	result := make([]string, 0, len(languages))
	temp := map[string]struct{}{}
	for _, item := range languages {
		if _, ok := temp[item]; !ok {
			temp[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

// MaxInt return max between x and y
func MaxInt(x, y int64) int64 {
	if x > y {
		return x
	}
	return y
}

// StringSliceToIntSlice convert string slice to int slice
func StringSliceToIntSlice(strSlice []string) []int {
	var result []int
	for _, str := range strSlice {
		i, err := strconv.Atoi(str)
		if err != nil {
			hwlog.RunLog.Errorf("failed convert str slice to int slice, err: %v", err)
			return nil
		}
		result = append(result, i)
	}
	return result
}
