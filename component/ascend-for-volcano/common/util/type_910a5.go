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

// Package util is using for the total variable.
package util

// for a5 rescheduling
const (
	// ServerIndexKey serverIndex key of node annotations for A5
	ServerIndexKey = "serverIndex"
)

// RankLevel for rank table level info
type RankLevel struct {
	// Level for the level index in rank table
	Level int `json:"level"`
	// Info for all possible infos in the level in rank table, key is the different net type, such as UB/UBG/UBoE/RoCE
	Info map[string]LevelElement `json:"info"`
}

// LevelElement for the concrete level info in rank table
type LevelElement struct {
	NetLayer      int    `json:"net_layer"`       // generate by operator, tentatively increase from 0
	NetInstanceID string `json:"net_instance_id"` // from annotation, relying on super_pod_id field
	NetType       string `json:"net_type"`        // generate by operator, level=0 tentatively empty; level=1,2 clos
	NetAttr       string `json:"net_attr"`        // generate by operator, tentatively empty /
	// generate by operator, tentatively level=0,3 nil; level=1,2 from 9th and 18th eid
	RankAddrList []RankAddrItem `json:"rank_addr_list"`
}

// RankAddrItem for item info in LevelElement
type RankAddrItem struct {
	AddrType string   `json:"addr_type"`
	Addr     string   `json:"addr"`
	Ports    []string `json:"ports"`
	PlaneId  string   `json:"plane_id"`
}
