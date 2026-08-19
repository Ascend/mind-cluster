/*
Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.

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
Package util is using for the total variable.
*/

package util

import "k8s.io/klog/v2"

// ErrorCollector is used to collect errors
type ErrorCollector struct {
	count      int
	printLimit int
	title      string
	errors     map[string][]string
}

const (
	// DefaultPrintLimit default print limit
	DefaultPrintLimit = 10
)

// NewErrorCollector create a new error collector
func NewErrorCollector(title string, printLimit int) *ErrorCollector {
	return &ErrorCollector{
		count:      0,
		printLimit: printLimit,
		title:      title,
		errors:     map[string][]string{},
	}
}

// Add add an error
func (ec *ErrorCollector) Add(key string, err error) {
	if err == nil {
		return
	}
	ec.count++
	ec.errors[err.Error()] = append(ec.errors[err.Error()], key)
}

// Print print all errors
func (ec *ErrorCollector) Print() {
	if ec.count == 0 {
		return
	}
	klog.V(LogErrorLev).Infof("%s: %d errors found", ec.title, ec.count)
	for errStr, objs := range ec.errors {
		klog.V(LogErrorLev).Infof("err: %s, count: %d, detail: %v...", errStr, len(objs), objs[:Min(len(objs),
			ec.printLimit)])
	}
}
