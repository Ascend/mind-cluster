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
Package nslb is using for HuaWei Ascend pin tor affinity.
*/
package nslb

import (
	"k8s.io/apimachinery/pkg/util/sets"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

// TorHandler tor handler is a struct that handles the tor affinity job
type TorHandler struct {
	pluginName   string
	ServerList   []*plugin.Tor
	Job          *plugin.SchedulerJob
	globalTorEnv *plugin.TorList
}

type TorHandlerV1 struct {
	TorHandler
}

type TorHandlerV2 struct {
	oldTorInfos jobUsedTorInfos
	TorHandler
}

type TorSingleLevelHandler struct {
	TorHandler
}

type jobUsedTorInfos struct {
	sharedTorNumToAdd int
	isSingleTorJob    bool
	torBlackList      sets.String
	usedTor           []*plugin.Tor
	unUsedTor         []*plugin.Tor
	serverNums        map[string]int
	usedTorMaps       map[string]*plugin.Tor
}

const (
	podRankIndex               = "hccl/rankIndex"
	maxTorAffinityNodeScore    = float64(200)
	halfTorAffinityNodeScore   = float64(100)
	sharedTorAffinityNodeScore = float64(99)
	// NormalSchema the value of normal tor affinity
	NormalSchema = "normal-schema"
	// NullTag the value means not use tor affinity
	NullTag = "null"
	// SingleLayer the single layer switch value of tor level in configmap
	SingleLayer = "single_layer"
	// TorAffinityKey the key of tor affinity
	TorAffinityKey = "tor-affinity"
	// LargeModelTag the value of large model
	LargeModelTag = "large-model-schema"
	// SharedTorIp shared tor Ip
	SharedTorIp = "sharedTorIp"
)

const (
	// the define of tor attr
	sharedTor    = 1
	exclusiveTor = 2
	freeTor      = 0
	allTor       = -1
	freeTorAnno  = "0"
	// the define of tor is healthy
	healthyTor = 0
	// the define of tor is unhealthy
	unhealthyTor = 1
)

const (
	defaultNSLBVersion   = "1.0"
	oneTor               = 1
	twoTor               = 2
	nslbv2Version        = "2.0"
	descOrder            = "desc"
	ascOrder             = "asc"
	isHealthy            = "isHealthy"
	isSharedTor          = "isSharedTor"
	pluginName           = "torAffinity"
	noneSharedTor        = 0
	fillJobMaxNPUTaskNum = 4
)
