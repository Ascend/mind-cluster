// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package util a series of util function
package util

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"os"
	"os/signal"
	"strconv"
	"time"

	"ascend-common/common-utils/hwlog"
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

func Abs[T int64 | int](x T) T {
	if x < 0 {
		return -x
	}
	return x
}

func DeleteStringSliceItem(slice []string, item string) []string {
	newSlice := make([]string, 0)
	for _, val := range slice {
		if val == item {
			continue
		}
		newSlice = append(newSlice, val)
	}
	return newSlice
}

// ReadableMsTime return more readable time from msec
func ReadableMsTime(msTime int64) string {
	return time.UnixMilli(msTime).Format("2006-01-02 15:04:05")
}

// DeepCopy for object using gob
func DeepCopy(dst, src interface{}) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(src); err != nil {
		return err
	}
	return gob.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(dst)
}
