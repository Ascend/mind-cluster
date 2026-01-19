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

import (
	"fmt"

	"ascend-faultdiag-online/pkg/core/model/enum"
)

// jobBase is the base info for slow node detection job
type jobBase struct {
	// JobName name of the slow node detection job
	JobName string `json:"jobName"`
	// JobId uniqe id of each job, got from job-summary
	JobId string `json:"jobId"`
}

type algoResultBase struct {
	jobBase
	// DegradationLevel 劣化百分点
	DegradationLevel string `json:"degradationLevel"`
	// IsSlow 是否存在慢节点
	IsSlow int `json:"isSlow"`
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

// AlgoResult 节点侧的结果模型
type NodeAlgoResult struct {
	algoResultBase
	Namespace string `json:"namespace"`
	NodeRank  string `json:"nodeRank"`
}

// KeyGenerator return a string combined by namespace and jobName as the key in slowNodeContext
func (r *NodeAlgoResult) KeyGenerator() string {
	return fmt.Sprintf("%s/%s", r.Namespace, r.JobName)
}

// ClusterAlgoResult 集群中心侧的结果模型
type ClusterAlgoResult struct {
	algoResultBase
	NodeRanks []string `json:"nodeRanks"`
}

// DataParseResult the data parse result data struct for node/cluster
type DataParseResult struct {
	jobBase
	// IsFinished whether the data parse finished or not
	IsFinished bool `json:"isFinished"`
	// FinishedTime the data parse finished time, sample: 1745567190000
	FinishedTime int `json:"finishedTime"`
	// StepCount is the step data steptime.csv
	StepCount int `json:"stepCount"`
	// RankIds real rankIds parsed in data parse
	RankIds []string `json:"rankIds"`
}

// NodeDataProfilingResult FD-ON in node if the initial profiling finished or not
type NodeDataProfilingResult struct {
	jobBase
	// Namespace the namespace of the job
	Namespace string `json:"namespace"`
	// FinishedInitialProfiling initial profiling finished sign: true -> finished false: unfinished
	FinishedInitialProfiling bool `json:"finishedInitialProfiling"`
	// FinishedTime is the timestamp of initial profiling finished
	FinishedTime int64 `json:"finishedTime"`
	// NodeIp the ip address of node
	NodeIp string `json:"nodeIp"`
	// ParallelGroupInfo the paralle group json data
	ParallelGroupInfo map[string]any `json:"parallelGroupInfo"`
}

// KeyGenerator return a string combined by namespace and jobName as the key in slowNodeContext
func (r *NodeDataProfilingResult) KeyGenerator() string {
	return fmt.Sprintf("%s/%s", r.Namespace, r.JobName)
}

// InputBase is the slow node input base
type InputBase struct {
	jobBase
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
	// RankIds the available rank ids of the node, e.g. ["0", "1", "2"]
	RankIds []string `json:"rankIds"`
}

// AlgoInput is the slow node input model
type AlgoInput struct {
	// FilePath 节点/集群根目录
	FilePath string `json:"filePath,omitempty"`
	InputBase
}

// Job is the slow node job conf
type Job struct {
	InputBase
	// SlowNode 特性开关，0 关闭，1 开启
	SlowNode int `json:"slowNode"`
	// Namespace
	Namespace string `json:"jobNamespace"`
	// Servers the servers in the job, e.g. [{"sn": "1", "ip": "192.168.0.1", "rankIds": ["0", "1", "2"]}]
	Servers []Server `json:"servers"`
}

// KeyGenerator return a string combined by namespace and jobName as the key in slowNodeContext
func (r *Job) KeyGenerator() string {
	return fmt.Sprintf("%s/%s", r.Namespace, r.JobName)
}

// DataParseInput is the input model for data parse
type DataParseInput struct {
	jobBase
	// FilePath is the file path for data parse reading the db, csv files
	FilePath string `json:"filePath"`
	// Traffic
	Traffic int64 `json:"traffic"`
	// ParallelGroupPath for cluster only
	ParallelGroupPath []string `json:"parallelGroupPath"`
	// RankIds the available rank ids of the node, e.g. ["0", "1", "2"]
	RankIds []string `json:"rankIds"`
}

// ReqInput is the request model for slow node algo&data parse
type ReqInput struct {
	// EventType is the type of input data, slow node algo or data parse
	EventType string
	// SlowNodeAlgoInput is the input model for slow node algo
	AlgoInput AlgoInput
	// DataParseInput is the input model for data parse
	DataParseInput DataParseInput
}

// ApiRes is a struct for response from so API.
type ApiRes struct {
	// Status
	Status enum.ResponseBodyStatus `json:"status"`
	// Msg
	Msg string `json:"msg"`
	// Data
	Data any `json:"data"`
}

// Server is a struct for cm data in job-summary
type Server struct {
	// Sn is the serial number of the server
	Sn string
	// Ip is the ip address of the server
	Ip string
	// RankIds is the rank ids of the server, e.g. ["0", "1", "2"]
	RankIds []string
}

// JobSummary is a struct for cm data in job-summary
type JobSummary struct {
	jobBase
	// Namespace is the namespace of the job
	Namespace string
	// JobStatus is the status of the job, running/complete/failed
	JobStatus string
	// Servers
	Servers []Server
}
