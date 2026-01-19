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

// Package algo for net fault detection algorithm
package algo

// PingItem ping info item
type PingItem struct {
	SrcType      int    `json:"srcType"`
	DstType      int    `json:"dstType"`
	PktSize      int    `json:"pktSize"`
	SrcCardPhyId int    `json:"srcCardPhyId"`
	SrcAddr      string `json:"srcAddr"`
	DstAddr      string `json:"dstAddr"`
}

// PingListInfo ping list info
type PingListInfo struct {
	PingList []PingItem `json:"pingList"`
}
