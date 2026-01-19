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

// Package config is used for file reading and writing, as well as data processing.
package config

import "os"

// TargetRankDir 匹配npu ID目录 id为群集内唯一
const TargetRankDir string = `^\d+$`

// JobLevelTopologyFileName 集群侧任务级完整并行域topo信息
const JobLevelTopologyFileName string = "parallel_group_global.json"

// NodeDetectionResultDirName 节点侧任务级检测结果存放路径
const NodeDetectionResultDirName string = "NodeLevelDetectionResult"

// NodeJobDetectionResultFileName 正则表达式匹配节点侧检测结果文件
const NodeJobDetectionResultFileName string = `^(\d+\.\d+\.\d+\.\d+)_Result\.json$`

// CallAlgoInterval 调用算法的间隔，即使本次读文件失败，也会自动调用下一次算法检测
const CallAlgoInterval int = 30

// DegradationPercent 劣化百分比
const DegradationPercent float64 = 100

// MaxNpuLinkNumsInDomain 并行域中npu最大连接数
const MaxNpuLinkNumsInDomain int = 2

/* 文件或目录不存在时重试次数,3min，间隔1s */
const fileNotExistRetryNums int = 180

// NormalNumberUpper 上限
const NormalNumberUpper int = 50

// DefaultNormalNumber 默认值
const DefaultNormalNumber int = 15

/* 并行域信息中字段 */
const parallelGroupName string = "group_name"

/* 并行域信息中字段 */
const parallelGlobalRanks string = "global_ranks"

/* 并行域信息中cp字段 */
const parallelCpField string = "cp"

/* 并行域信息中ep字段 */
const parallelEpField string = "exp"

/* 节点侧任务级npu卡并行域信息 */
const rankTopofileName string = "parallel_group.json"

/* host IP环境变量名称 */
const xdlIpField string = "XDL_IP"

/* 文件或目录可读权限 */
const readMode os.FileMode = 0400

// AlgoInputConfig 入参检测算法的配置解析结构
type AlgoInputConfig struct {
	// DetectionLevel 检测级别
	DetectionLevel string `json:"detectionLevel"`
	// FilePath 文件路径（数据源）
	FilePath string `json:"filePath"`
	// JobName 任务名称
	JobName string `json:"jobId"`
	// NormalNumber 计算初始阈值（正常数量）
	NormalNumber int `json:"normalNumber"`
	// Nsigma 使用多少个σ计算上下界
	Nsigma int `json:"nSigma"`
	// DegradationPercentage 阈值（劣化百分比，0.3表示劣化了30%）
	DegradationPercentage float64 `json:"degradationPercentage"`
	// NconsecAnomaliesSignifySlow 连续出现多少次异常才检测（例如：5次）
	NconsecAnomaliesSignifySlow int `json:"nConsecAnomaliesSignifySlow"`
	// ClusterMeanDistance 聚类后，两个类别之间的距离阈值，mean1/mean2 > 1.3
	ClusterMeanDistance float64 `json:"clusterMeanDistance"`
	// NsecondsOneDetection 多长时间检测一次（单位：秒）
	NsecondsOneDetection int `json:"nSecondsDoOneDetection"`
	// CardsOneNode 一个节点的卡片数量（例如：8张卡）
	CardsOneNode int `json:"cardOneNode"`
	// RankIds 当前任务所使用的npuId
	RankIds []string `json:"rankIds"`
}

// ClusterJobResult 集群侧任务级检测结果数据结构
type ClusterJobResult struct {
	// IsSlow 是否存在慢节点
	IsSlow int `json:"isSlow"`
	// DegradationLevel 劣化百分点
	DegradationLevel string `json:"degradationLevel"`
	// JobName 任务名称
	JobName string `json:"jobId"`
	// SlowCalculateRanks 慢计算卡
	SlowCalculateRanks []int `json:"slowCalculateRanks"`
	// SlowCommunicationDomains 慢通信域
	SlowCommunicationDomains [][]int `json:"slowCommunicationDomains"`
	// SlowHostNodes hosts侧慢卡
	SlowHostNodes []string `json:"slowHostNodes"`
	// SlowIONodes 慢IO卡的节点
	SlowIORanks []int `json:"slowIORanks"`
	// SlowCommunicationRanks  慢通信卡
	SlowCommunicationRanks []int `json:"slowCommunicationRanks"`
}

// ClusterDetectionResult 集群侧任务级检测上报和存储数据格式
type ClusterDetectionResult map[string]map[string]ClusterJobResult

// NodeJobResult 定义节点侧任务级检测结果数据结构
type NodeJobResult struct {
	// IsSlow 是否存在变慢
	IsSlow int `json:"isSlow"`
	// DegradationLevel 劣化程度
	DegradationLevel string `json:"degradationLevel"`
	// JobName 检测任务的名称
	JobName string `json:"jobId"`
	// NodeRank 当前节点IP
	NodeRank string `json:"nodeRank"`
	// SlowCalculateRanks 节点级慢计算卡
	SlowCalculateRanks []int `json:"slowCalculateRanks"`
	// SlowCommunicationDomains 节点级慢通信域
	SlowCommunicationDomains [][]int `json:"slowCommunicationDomains"`
	// SlowSendRanks 节点级slow send
	SlowSendRanks []int `json:"slowSendRanks"`
	// SlowHostNodes 当前节点是否为慢节点
	SlowHostNodes []string `json:"slowHostNodes"`
	// SlowIORanks 节点级慢IO卡
	SlowIORanks []int `json:"slowIORanks"`
}

// NodeDetectionResult 节点侧任务级检测上报和存储结果数据格式
type NodeDetectionResult map[string]map[string]NodeJobResult

// DetectionConfig 存储检测系统的配置
type DetectionConfig struct {
	// DetectionLevel 检测级别
	DetectionLevel string `json:"detectionLevel"`

	// SharedFilePath 共享文件路径（结果落盘地址）
	SharedFilePath string `json:"sharedFilePath"`

	// LocalFilePath 本地文件路径（数据源）
	LocalFilePath string `json:"localFilePath"`

	// NormalNumber 计算初始阈值（正常数量）
	NormalNumber int `json:"normalNumber"`

	// Nsigma 使用多少个σ计算上下界
	Nsigma int `json:"nSigma"`

	// DegradationPercentage 阈值（劣化百分比，0.3表示劣化了30%）
	DegradationPercentage float64 `json:"degradationPercentage"`

	// NconsecAnomaliesSignifySlow 连续出现多少次异常才检测（例如：5次）
	NconsecAnomaliesSignifySlow int `json:"nConsecAnomaliesSignifySlow"`

	// ClusterMeanDistance 聚类后，两个类别之间的距离阈值，mean1/mean2 > 1.3
	ClusterMeanDistance float64 `json:"clusterMeanDistance"`

	// NsecondsOneDetection 多长时间检测一次（单位：秒）
	NsecondsOneDetection int `json:"nSecondsDoOneDetection"`

	// CardsOneNode 一个节点的卡片数量（例如：8张卡）
	CardsOneNode int `json:"cardOneNode"`
}

// DataParseModel the model definition of data parse
type DataParseModel struct {
	// FilePath need to parsed file saved path
	FilePath string `json:"filePath"`
	// JobName the unique name of a job
	JobName string `json:"jobName"`
	// jobId the unique id of a job
	JobId string `json:"jobId"`
	// Traffic 通信量
	Traffic int64 `json:"traffic"`
	// ParallelGroupPath 集群侧并行域文件路径
	ParallelGroupPath []string `json:"parallelGroupPath"`
	// RankIds 当前任务所使用的npuId
	RankIds []string `json:"rankIds"`
	// JobStartTime 任务开始时间
	JobStartTime int64 `json:"-"`
}
