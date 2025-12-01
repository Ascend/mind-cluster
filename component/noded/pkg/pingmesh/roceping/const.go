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

// Package roceping for ping by icmp in RoCE mesh net between super pods in A5
package roceping

const (
	readTimeout         = 1
	readBuffSize        = 1500
	waitTimeoutMilliSec = 1000

	maxRetryTimes        = 24
	waitTimesForGenerate = 5

	savePeriodMillSec = 45 * 1000
	rasNetSubPath     = "cluster"
	roceSubPath       = "super-pod-roce"

	digitalBase            = 10
	float64FormatType      = 'f'
	float64FormatPrecision = 3
	float64BitSize         = 64

	specialTimeFormat = "2006-01-02 15:04:05.0000000"

	maxIcmpSequenceId = (1 << 16) - 1

	acceleratorTypeKey       = "accelerator-type" // acceleratorTypeKey 加速标签
	labelPrefix900SuperPodA5 = "900superpod-a5"   // A5超节点形态加速标签前缀小写格式
)
