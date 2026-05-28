/* Copyright(C) 2021-2023. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package dcmi this for constants
package dcmi

// MainCmd main command enum
type MainCmd uint32

// VDevMngSubCmd virtual device manager sub command
type VDevMngSubCmd uint32

// DieType present chip die type
type DieType int32

// DiePolicyType die policy type
type DiePolicyType uint32

const (
	// DcmiMaxVdevNum is max number of vdevice, value is from driver specification
	DcmiMaxVdevNum = 32
	// DcmiMaxReserveNum is max number of reserve, value is from driver specification
	DcmiMaxReserveNum = 8
	// DcmiVDevResNameLen length of vnpu resource name
	DcmiVDevResNameLen = 16
	// dcmiHccsMaxPcsNum max pcs number for hccs
	dcmiHccsMaxPcsNum = 16

	MaxChipNameLen = 32
	productTypeLen = 64
	DcmiVersionLen = 32

	// MainCmdChipInf main cmd chip inf
	MainCmdChipInf MainCmd = 12
	// MainCmdHccs main cmd of hccs
	MainCmdHccs MainCmd = 16
	// MainCmdVDevMng virtual device manager
	MainCmdVDevMng MainCmd = 52
	// MainCmdSio SIO status between die
	MainCmdSio MainCmd = 56

	// VmngSubCmdGetVDevResource get virtual device resource info
	VmngSubCmdGetVDevResource VDevMngSubCmd = 0
	// VmngSubCmdGetTotalResource get total resource info
	VmngSubCmdGetTotalResource VDevMngSubCmd = 1
	// VmngSubCmdGetFreeResource get free resource info
	VmngSubCmdGetFreeResource VDevMngSubCmd = 2
	// VmngSubCmdGetVDevActivity get vir device activity info
	VmngSubCmdGetVDevActivity VDevMngSubCmd = 5
	// CinfSubCmdGetSPodInfo get super pod info
	CinfSubCmdGetSPodInfo VDevMngSubCmd = 1
	// SioSubCmdCrcErrStatistics get SIO err statistics info
	SioSubCmdCrcErrStatistics VDevMngSubCmd = 0
	// HccsSubCmdGetStatisticInfo get statistic info
	HccsSubCmdGetStatisticInfo VDevMngSubCmd = 3
	// HccsSubCmdGetStatisticInfoU64 get statistic info in u64
	HccsSubCmdGetStatisticInfoU64 VDevMngSubCmd = 5

	// NDIE NDie ID, only Ascend910 has
	NDIE DieType = 0
	// VDIE VDie ID, it can be the uuid of chip
	VDIE DieType = 1
	// DDIE DDie ID, it can be the uuid of A5 chip
	DDIE DieType = 2
	// DieIDCount die id array max length
	DieIDCount = 5

	// IpAddrTypeV6 ip address type of IPv6
	IpAddrTypeV6 = 1
	// HexBase hex base number
	HexBase             = 16
	AgentdrvProfDataNum = 3

	// MultiDieUnion A3 multi-die must be supplied into the container simultaneously
	MultiDieUnion DiePolicyType = 0
	// MultiDieIndep A3 supports single-die independent injection into the container
	MultiDieIndep DiePolicyType = 1
)
