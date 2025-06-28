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
Package consts define common variable
*/
package consts

const (
	// ResultRootDir is the root dir of pingmesh result
	ResultRootDir = "/user/mind-cluster/pingmesh"
	// PingMeshConfigCm is the name of pingmesh configmap
	PingMeshConfigCm = "pingmesh-config"
	// IpConfigmapNamePrefix is the prefix of ip configmap name
	IpConfigmapNamePrefix = "super-pod-"
	// PingMeshFaultCmPrefix is the label key of ip configmap
	PingMeshFaultCmPrefix = "pingmesh-fault-"
	// FaultConfigmapLabelValue is the label value of ip configmap
	FaultConfigmapLabelValue = "true"
	// PingMeshConfigLabelKey is the label key of pingmesh configmap
	PingMeshConfigLabelKey = "app"
	// PingMeshConfigLabelValue is the label value of pingmesh configmap
	PingMeshConfigLabelValue = "pingmesh"
	// SuffixOfPingMeshLogFile is the suffix of pingmeshv1 log file
	SuffixOfPingMeshLogFile = ".log"
)
