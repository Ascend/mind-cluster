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
Package slownode 特性
*/
package slownode

import "ascend-faultdiag-online/pkg/model/enum"

// NodeSlowNodeAlgoResult 节点侧的结果模型
type NodeSlowNodeAlgoResult struct {
	// DegradationLevel 劣化百分点
	DegradationLevel string `json:"degradationLevel"`
	// IsSlow 是否存在慢节点
	IsSlow int `json:"isSlow"`
	// JobName 任务名称
	JobName string `json:"jobName"`
	// JobId
	JobId string `json:"jobId"`
	// NodeRank 节点IP地址
	NodeRank string `json:"nodeRank"`
	// SlowCalculateRanks 慢计算卡
	SlowCalculateRanks []int `json:"slowCalculateRanks"`
	// SlowCommunicationDomains 慢通信域
	SlowCommunicationDomains [][]int `json:"slowCommunicationDomains"`
	// SlowHostNodes hosts侧慢
	SlowHostNodes []string `json:"slowHostNodes"`
	// SlowIORanks 慢IO卡
	SlowIORanks []int `json:"slowIORanks"`
	// SlowSendRanks 慢send卡的globalrank
	SlowSendRanks []int `json:"slowSendRanks"`
}

func (r *NodeSlowNodeAlgoResult) GetJobName() string { return r.JobName }

// DataParseResult the data parse result data struct for node/cluster
type DataParseResult struct {
	// JobName 任务名称
	JobName string `json:"jobName"`
	// JobId
	JobId string `json:"jobId"`
	// IsFinished whether the data parse finished or not
	IsFinished bool `json:"isFinished"`
	// FinishedTime the data parse finished time, sample: 1745567190000
	FinishedTime int `json:"finishedTime"`
	// StepCount is the step data steptime.csv
	StepCount int `json:"stepCount"`
}

// ClusterSlowNodeAlgoResult 集群中心侧的结果模型
type ClusterSlowNodeAlgoResult struct {
	// DegradationLevel 劣化百分点
	DegradationLevel string `json:"degradationLevel"`
	// IsSlow 是否存在慢节点
	IsSlow int `json:"isSlow"`
	// JobName 任务名称
	JobName string `json:"jobName"`
	// JobId
	JobId string `json:"jobId"`
	// NodeRanks 节点IP地址列表
	NodeRanks []string `json:"nodeRanks"`
	// SlowCalculateRanks 慢计算卡
	SlowCalculateRanks []int `json:"slowCalculateRanks"`
	// SlowCommunicationDomains 慢通信域
	SlowCommunicationDomains [][]int `json:"slowCommunicationDomains"`
	// SlowCommunicationRanks  慢通信卡
	SlowCommunicationRanks []int `json:"slowCommunicationRanks"`
	// SlowHostNodes hosts侧慢
	SlowHostNodes []string `json:"slowHostNodes"`
	// SlowIORanks 慢IO卡
	SlowIORanks []int `json:"slowIORanks"`
}

// NodeDataProfilingResult FD-ON in node if the initial profiling finished or not
type NodeDataProfilingResult struct {
	// JobName 任务名称
	JobName string `json:"jobName"`
	// JobId
	JobId string `json:"jobId"`
	// FinishedInitialProfiling initial profiling finished sign: true -> finished false: unfinished
	FinishedInitialProfiling bool `json:"finishedInitialProfiling"`
	// FinishedTime is the timestamp of initial profiling finished
	FinishedTime int64 `json:"finishedTime"`
	// NodeIP the ip address of node
	NodeIP string `json:"nodeIP"`
	// ParallelGroupInfo the paralle group json data
	ParallelGroupInfo map[string]any `json:"ParallelGroupInfo"`
}

func (r *NodeDataProfilingResult) GetJobName() string { return r.JobName }

// SlowNodeInputBase slow node input base
type SlowNodeInputBase struct {
	// JobName 检测任务名称
	JobName string `json:"jobName,omitempty"`
	// JobId
	JobId string `json:"jobId"`
	// NormalNumber 计算初始阈值（正常数量）
	NormalNumber int `json:"normalNumber,omitempty"`
	// NSigma 使用多少个σ计算上下界
	NSigma int `json:"nSigma,omitempty"`
	// DegradationPercentage 阈值（劣化百分比，0.3表示劣化了30%）
	DegradationPercentage float64 `json:"degradationPercentage,omitempty"`
	// NConsecAnomaliesSignifySlow 连续出现多少次异常才检测（例如：5次）
	NConsecAnomaliesSignifySlow int `json:"nConsecAnomaliesSignifySlow,omitempty"`
	// NSecondsDoOneDetection 聚类后，两个类别之间的距离阈值，mean1/mean2 > 1.3
	NSecondsDoOneDetection int `json:"nSecondsDoOneDetection,omitempty"`
	// ClusterMeanDistance 多长时间检测一次（单位：秒）
	ClusterMeanDistance float64 `json:"clusterMeanDistance,omitempty"`
	// CardOneNode 一个节点的卡片数量（例如：8张卡）
	CardOneNode int `json:"cardOneNode,omitempty"`
}

// SlowNodeAlgoInput slow node input model
type SlowNodeAlgoInput struct {
	// FilePath 节点/集群根目录
	FilePath string `json:"filePath,omitempty"`
	SlowNodeInputBase
}

// SlowNodeJob slow node job conf
type SlowNodeJob struct {
	SlowNodeInputBase
	// SlowNode 特性开关，0 关闭，1 开启
	SlowNode int `json:"slowNode"`
	// Namespace
	Namespace string `json:"jobNamespace"`
}

func (r *SlowNodeJob) GetJobName() string { return r.JobName }

// DataParseInput is the input model for data parse
type DataParseInput struct {
	// FilePath is the file path for data parse reading the db, csv files
	FilePath string `json:"filePath"`
	// JobName is the job name
	JobName string `json:"jobName"`
	// JobId
	JobId string `json:"jobId"`
	// Traffic
	Traffic int64 `json:"traffic"`
	// ParallelGroupPath for cluster only
	ParallelGroupPath []string `json:"parallelGroupPath"`
}

// SlownodeInput is the input model for slow node algo&data parse
type SlowNodeInput struct {
	// EventType is the type of input data, slow node algo or data parse
	EventType enum.SlowNodeEventType
	// SlowNodeAlgoInput is the input model for slow node algo
	SlowNodeAlgoInput SlowNodeAlgoInput
	// DataParseInput is the input model for data parse
	DataParseInput DataParseInput
}

// Input is an input data in the execute function of so.
type Input struct {
	// Command what the operation the so package should do: start/stop
	Command enum.Command `json:"command"`
	// Target cluster/node
	Target enum.DeployMode `json:"target"`
	// Model is the data which so package will process
	Model any `json:"model"`
	// Func is the uintptr of the callback func, using uint replace uintptr
	//the so package will call this func if there is callback data
	Func uint `json:"func"`
	// EventType is the enum of execution type, slow node or dataParse
	EventType enum.SlowNodeEventType `json:"eventType"`
}

// ApiRes is a struct for response from so API.
type ApiRes struct {
	// Status
	Status string `json:"status"`
	// Msg
	Msg string `json:"msg"`
	// Data
	Data any `json:"data"`
}
