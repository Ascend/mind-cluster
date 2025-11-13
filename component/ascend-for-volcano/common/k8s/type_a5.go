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

// Package k8s is using for the k8s operation
package k8s

import "sync"

// DpuInfosFromCmWithMutex DPU infos from configmap with mutex
type DpuInfosFromCmWithMutex struct {
	sync.Mutex
	Dpus map[string]DpuCMInfo
}

// DpuListItem one DPU info struct
type DpuListItem struct {
	Name      string `json:"Name"`
	Operstate string `json:"Operstate"`
	Deviceid  string `json:"DeviceID"`
	Vendor    string `json:"VendorID"`
}

// DpuCMInfo  one node dpu info struct from clusterd
type DpuCMInfo struct {
	DpuList         []DpuListItem       `json:"DPUList"`
	BusType         string              `json:"BusType"`
	CacheUpdateTime int64               `json:"UpdateTime"`
	NpuToDpusMap    map[string][]string `json:"NpuToDpusMap"`
}
