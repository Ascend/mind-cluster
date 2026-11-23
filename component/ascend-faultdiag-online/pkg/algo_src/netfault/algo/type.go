/* Copyright 2025 Huawei Technologies Co., Ltd.

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
	// SrcType indicates the type of the source end
	SrcType int `json:"srcType"`
	// DstType indicates the type of the destination end
	DstType int `json:"dstType"`
	// PktSize represents the size of the ping packet
	PktSize int `json:"pktSize"`
	// SrcCardPhyId is the physical ID of the source-end card
	SrcCardPhyId int `json:"srcCardPhyId"`
	// SrcAddr is the network address of the source end
	SrcAddr string `json:"srcAddr"`
	// DstAddr is the network address of the destination end
	DstAddr string `json:"dstAddr"`
}

// PingListInfo ping list info
type PingListInfo struct {
	// PingList is a slice that stores multiple Ping info
	PingList []PingItem `json:"pingList"`
}
