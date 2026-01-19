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
Package utils 提供了一些工具函数
*/
package utils

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"sync"
	"time"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/utils/constants"
)

var (
	mu             sync.Mutex
	startTimestamp int64 = 0
)

const (
	defaultRetryCount = 20
	defaultSleepTime  = 3 * time.Second
	maxRetryCount     = 100
	minSleepTime      = 10 * time.Millisecond
)

// ToFloat64 interface转float64
func ToFloat64(val any, defaultValue float64) float64 {
	switch v := val.(type) {
	case float32:
		return float64(v)
	case float64:
		return v
	case int:
		return float64(v)
	case string:
		float, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return defaultValue
		}
		return float
	default:
		return defaultValue
	}
}

// ToString interface转string
func ToString(val any) string {
	str, ok := val.(string)
	if !ok {
		return ""
	}
	return str
}

// CopyInstance 复制实例
func CopyInstance(src any) (any, error) {
	if src == nil {
		return nil, fmt.Errorf("src cannot be nil")
	}
	srcValue := reflect.ValueOf(src)
	if srcValue.Kind() == reflect.Ptr {
		if srcValue.IsNil() {
			return nil, fmt.Errorf("src ptr cannot be nil")
		}
		srcValue = srcValue.Elem()
	} else {
		return nil, fmt.Errorf("copy instance src is not ptr")
	}
	dst := reflect.New(srcValue.Type())
	dst.Elem().Set(srcValue)
	return dst.Interface(), nil
}

// WriteStartInfo write the current timestamp into file
func WriteStartInfo() {
	mu.Lock()
	defer mu.Unlock()
	if startTimestamp == 0 {
		startTimestamp = time.Now().UnixMilli()
	}
}

// IsRestarted return the pod is restart or not
func IsRestarted() bool {
	// check the file, if start time less than 2 seconds, as restarted
	mu.Lock()
	defer mu.Unlock()
	if startTimestamp == 0 {
		startTimestamp = time.Now().UnixMilli()
		return false
	}
	return time.Now().UnixMilli()-startTimestamp <= constants.RestartInterval
}

// RetryConfig is a config for Retry function
type RetryConfig struct {
	// RetryCount is the max retry count
	RetryCount int
	// SleepTime is the time that retry wait for each calling
	SleepTime time.Duration
}

// Retry is a common function to retry calling the provide f, default cg retryCount: 20, sleepTime: 3s
func Retry[T any](f func() (T, error), cg *RetryConfig) (T, error) {
	if cg == nil {
		// using default config
		cg = &RetryConfig{
			RetryCount: defaultRetryCount,
			SleepTime:  defaultSleepTime,
		}
	}
	var res T
	if f == nil {
		return res, errors.New("retry failed: func is nil")
	}
	if cg.RetryCount > maxRetryCount || cg.SleepTime.Milliseconds() < minSleepTime.Milliseconds() {
		return res, fmt.Errorf("config check failed: excced the max retry count: %d or less than min sleep time: %v",
			maxRetryCount, minSleepTime)
	}
	var err error
	for count := 0; count < cg.RetryCount; count++ {
		res, err = f()
		if err == nil {
			return res, nil
		}
		if count < cg.RetryCount-1 {
			hwlog.RunLog.Warnf("[FD-OL]call failed: %v, retry: %d/%d", err, count+1, cg.RetryCount)
			time.Sleep(cg.SleepTime)
		}
	}
	hwlog.RunLog.Errorf("[FD-OL]reached the max try: %d, and got err: %v", cg.RetryCount, err)
	return res, err
}
