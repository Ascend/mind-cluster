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
Package diagnose 提供诊断功能，用于检查系统或应用程序的健康状况。
*/
package diagnose

import (
	"ascend-faultdiag-online/pkg/context/diagcontext"
	"ascend-faultdiag-online/pkg/diagnose/metricdiag"
)

// DefaultDiagItems 默认的诊断项列表
func DefaultDiagItems() []*diagcontext.DiagItem {
	var diagItems []*diagcontext.DiagItem
	diagItems = append(diagItems, metricdiag.GetBandWidthDiagItems()...)
	return diagItems
}
