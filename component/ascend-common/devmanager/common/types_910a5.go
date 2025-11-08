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

// Eid Entity ID which is 128bit and used as Unify BUS address
type Eid struct {
	Raw [EidByteSize]byte
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
