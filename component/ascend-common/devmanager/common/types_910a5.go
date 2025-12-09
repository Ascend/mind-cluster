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

// Package common define common types
package common

const (
	// EidNumMax urma device max 32 eid
	EidNumMax = 32
	// EidByteSize eid ->16byte 128bit
	EidByteSize = 16
)

// UbPingMeshMaxNum defines the maximum number of UB ping mesh tasks or destinations.
const (
	UbPingMeshMaxNum = 48
)

// Eid Entity ID which is 128bit and used as Unify BUS address
type Eid struct {
	Raw [EidByteSize]byte
}

// UBPingMeshOperate defines the configuration for a UB ping mesh operation.
type UBPingMeshOperate struct {
	SrcEID       Eid   // Source endpoint ID
	DstEIDList   []Eid // List of destination endpoint IDs
	DstNum       int   // Number of destinations
	PktSize      int   // Size of each ping packet in bytes
	PktSendNum   int   // Number of packets to send per destination
	PktInterval  int   // Interval between sending packets (ms)
	Timeout      int   // Timeout for a single ping reply (ms)
	TaskInterval int   // Interval between consecutive ping mesh tasks (ms)
	TaskID       int   // Unique task ID for tracking
}

// UBPingMeshInfo stores the results/statistics of a UB ping mesh task.
type UBPingMeshInfo struct {
	SrcEIDs      Eid    // Source endpoint ID (supports multiple EIDs if compressed)
	DstEIDList   []Eid  // List of destination endpoint IDs
	SucPktNum    []uint // Number of successfully received packets for each destination
	FailPktNum   []uint // Number of failed packet deliveries for each destination
	MaxTime      []int  // Maximum round-trip time (ms) for each destination
	MinTime      []int  // Minimum round-trip time (ms) for each destination
	AvgTime      []int  // Average round-trip time (ms) for each destination
	Tp95Time     []int  // 95th percentile round-trip time (ms) for each destination
	ReplyStatNum []int  // Number of replies received per destination (possibly categorically)
	PingTotalNum []int  // Total number of pings sent to each destination
	DestNum      int    // Total number of destinations in the result
	OccurTime    uint   // Timestamp when the result was recorded (in ms)
}

// CgoUBAddr ub address
type CgoUBAddr struct {
	PortID int32
	EID    string
	CNA    string
}

// UrmaEidInfo eid info in urma device
type UrmaEidInfo struct {
	Eid      Eid
	EidIndex uint
}

// UrmaDeviceInfo a urma device represent a URMA device
type UrmaDeviceInfo struct {
	EidCount uint
	EidInfos []UrmaEidInfo
}
