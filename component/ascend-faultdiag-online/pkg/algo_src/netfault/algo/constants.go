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

// Package algo 网络连通性检测算法
package algo

// 算法常用字符
const (
	normalIntervalChar = "-"
	dotIntervalChar    = "."
	layerIntervalChar  = "#"
	portIntervalChar   = ":"
	objectIntervalChar = "_"
	fuzzyAlarmFlagChar = "--"
)

// 算法常量字符串
const (
	a3NpuTypeConstant  = "A3"
	a5NpuTypeConstant  = "A5"
	sameAxisConstant   = "same_axis"
	crossAxisConstant  = "cross_axis"
	bothAxisConstant   = "both_axis"
	fromConstant       = "from"
	toConstant         = "to"
	ipConstant         = "ip"
	portConstant       = "port"
	layerConstant      = "layer"
	layer1Constant     = "layer_1"
	layer2Constant     = "layer_2"
	layer3Constant     = "layer_3"
	layer4Constant     = "layer_4"
	superPodConstant   = "SuperPod"
	roceSwitchConstant = "ROCESwitch"
)

// 算法输入参数
const (
	argsPeriod          = "period"
	argsSPeriod         = "suppressedPeriod"
	argsPingObjType     = "pingObjType"
	argsServerIdMap     = "serverIdMap"
	argsAxisStrategy    = "axisStrategy"
	argsSuperPodJobFlag = "superPodJobFlag"
	argsSuperPodArr     = "superPodArr"
	argsNpuType         = "npu_type"
	argsNpu2Npu         = "npu_npu"
	argsNpu2NetPlane    = "npu_netplane"
	argsNpu2SuperPod    = "npu_superpod"
	argsRackName        = "RackName"
	argsSlotName        = "SlotName"
	argsNpuNumber       = "NpuNumber"
	argsNpuIp           = "IP"
	argsNetPlaneId      = "NetPlaneId"
)

// 算法输出参数
const (
	pingListConstant  = "pingList"
	srcTypeConstant   = "srcType"
	dstTypeConstant   = "dstType"
	srcIdConstant     = "srcId"
	dstIdConstant     = "dstId"
	srcCardPhyId      = "srcCardPhyId"
	dstCardPhyId      = "dstCardPhyId"
	srcAddrConstant   = "srcAddr"
	dstAddrConstant   = "dstAddr"
	pktSizeConstant   = "pktSize"
	levelConstant     = "level"
	faultTypeConstant = "faultType"
)

// csv文件列名
const (
	pingTaskIDConstant  = "pingTaskId"
	fromLayerConstant   = "fromLayer"
	toLayerConstant     = "toLayer"
	maxDelayConstant    = "maxDelay"
	minDelayConstant    = "minDelay"
	avgDelayConstant    = "avgDelay"
	maxLoseRateConstant = "maxLossRate"
	minLoseRateConstant = "minLossRate"
	avgLoseRateConstant = "avgLossRate"
	timestampConstant   = "timestamp"
)

// 告警输出
const (
	descriptionConstant = "description"
	informationConstant = "information"
	taskIDConstant      = "taskId"
	rootCauseConstant   = "rootCause"
	nSlotConstant       = "NSlot"
	npuConstant         = "npu"
	cpuConstant         = "cpu"
	unionConstant       = "union"
	rackConstant        = "Rack"
	nodeConstant        = "Node"
	l1Constant          = "L1"
	l2Constant          = "L2"
)

// 数值常量定义
const (
	sampleNum         = 64
	pktSizeNum        = 28
	minimumColNum     = 3
	baseEvenNum       = 2
	baseSegmentNum    = 2
	baseNpuNum        = 8
	basePercentNum    = 100
	millisecondNum    = 1000
	lossThreshold     = 80
	saveLenNum        = 10
	superPodNum       = 3
	defaultPingPeriod = 15
	defaultSPeriod    = 10
	aggregatePathNum  = 3
	coefficientNum    = 3
	fromToNum         = 2
)

// 拨测对象枚举类型
const (
	// IpType ip类型
	IpType = 0
	// SdidType sdid类型
	SdidType = 1
	// EidType eid类型
	EidType = 2
)

// 根因对象枚举类型
const (
	npuType          = 0
	cpuType          = 1
	unionType        = 2
	rackNetplaneType = 3
	workNodeType     = 4
	l1NetplaneType   = 5
	l2NetplaneType   = 6
	superPodType     = 7
	roceSwitchType   = 8
)

// 告警级别枚举类型
const (
	minorType    = 0
	majorType    = 1
	criticalType = 2
)

// 故障检测枚举类型
const (
	delayType      = 0
	lossRateType   = 1
	disconnectType = 2
)

// DataFrame 表示数据框，包含列名、每列的数据（Chains）和行数
type DataFrame struct {
	columnNames []string
	chains      map[string]any
	rowNum      int
}

// Group 表示一个分组，包含分组的键和对应的子数据框
type Group struct {
	key       string
	groupData *DataFrame
}

// DataFrameGroupBy 表示分组结果，包含所有分组的映射
type DataFrameGroupBy struct {
	groupNums int
	groups    []*Group
}

// AiPingStrategy 拨测策略结构体
type AiPingStrategy struct {
	npuNpuList []string
	chainList  map[string][]string
	pingList   []string
	layersIps  map[string]any
	dfGrouped  *DataFrameGroupBy
	pingDict   map[string]any
}

// NetDetect 算法对象结构体（暴露给controller使用）
type NetDetect struct {
	curSuperPodId        string
	curNpuType           string
	curServerIdMap       map[string]string
	curFullPingFlag      bool
	curOpenQueueFlag     bool
	curSuperPodJobFlag   bool
	curSuperPodArr       []string
	curAxisStrategy      string
	curTopo              []string
	curPingPeriod        int
	curSuppressedPeriod  int
	curPingObjType       int
	curDetectParams      map[string]any
	curNpuInfo           map[string]NpuInfo
	curSlideWindows      []map[string]any
	curConsumedQueue     []map[string]any
	curSlideWindowsMaxTs int64
	pathIndex            map[string][]map[string]any
}

// NpuInfo npu信息结构体（暴露给controller使用）
type NpuInfo struct {
	// SuperPodName 超节点号
	SuperPodName string
	// RackName 框号
	RackName string
	// OsName os号
	OsName string
	// SlotName 板号
	SlotName string
	// NpuNumber npu物理id
	NpuNumber int
	// NpuNumber npu ip
	IP string
	// NpuNumber npu所属网络平面id
	NetPlaneId string
}

// algo里的全局变量
var (
	globalHistoryAlarms     = make(map[string]any, superPodNum)
	globalRootCauseEventNpu = make(map[string]string, superPodNum)
	globalRootCauseEventCpu = make(map[string]string, superPodNum)
	globalDetectTypes       = []string{avgLoseRateConstant, avgDelayConstant}
	globalPathKeys          = []string{pingTaskIDConstant, srcTypeConstant, srcAddrConstant,
		dstTypeConstant, dstAddrConstant}
)
