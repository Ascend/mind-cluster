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

// Package constants provides some constants code
package constants

const (
	// TypeSeparator character -
	TypeSeparator = "-"
	// ValueSeparator character :
	ValueSeparator = ":"
)

const (
	// MetricBandwidth fixed name: bandwidth
	MetricBandwidth = "bandwidth"
)

const (
	// ApiSeparator character /
	ApiSeparator = "/"
)

const (
	// TaskId the id of pingmesh task
	TaskId = "TaskID"
	// MinLossRate the mini loss rate of network
	MinLossRate = "MinLossRate"
	// MaxLossRate the max loss rate of network
	MaxLossRate = "MaxLossRate"
	// AvgLossRate the avg loss rate of network
	AvgLossRate = "AvgLossRate"
	// MinDelay the min time delay of network
	MinDelay = "MinDelay"
	// MaxDelay the max time delay of network
	MaxDelay = "MaxDelay"
	// AvgDelay the avg time delay of network
	AvgDelay = "AvgDelay"
	// FaultType the fault type of network
	FaultType = "FaultType"
	// SrcID source ip
	SrcID = "SrcID"
	// SrcType the type of source ip
	SrcType = "SrcType"
	// DstID dest ip
	DstID = "DstID"
	// DstType the type of dest ip
	DstType = "DstType"
	// Level the level of fault
	Level = "Level"
	// XdlIpField the ip env variable of node
	XdlIpField = "XDL_IP"
	// PodIP is the ip env variable of cluster
	PodIP = "POD_IP"
	// GrpcPort is the port of Grpc server
	GrpcPort = ":8899"
	// MaxConfigMapNum the top number of config map size allowed created
	MaxConfigMapNum = 20000
	// RestartInterval is the interval judge that the pod is restarted or not, unit is milliseconds
	RestartInterval = 2000
	// OneDaySeconds is the number of seconds in a day
	OneDaySeconds = 24 * 60 * 60
)

const (
	// MaxFileCount the max file number under a path
	MaxFileCount = 10000
	// Size10M size 10M
	Size10M = 10 * 1025 * 1024
	// Size50M size 50M
	Size50M = 50 * 1025 * 1024
	// Size100M size 100M
	Size100M = 100 * 1024 * 1024
	// Size500M size 500M
	Size500M = 500 * 1024 * 1024
)
