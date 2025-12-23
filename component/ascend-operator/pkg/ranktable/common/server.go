/*
Copyright(C) 2024. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package common is common function or object of ranktable.
*/
package common

import "ascend-common/api"

// Server to hccl
type Server struct {
	DeviceList   []*Device `json:"device"` // device list in each server
	Hardware     string    `json:"hardware_type,omitempty"`
	ServerID     string    `json:"server_id"` // server id, represented by ip address
	HostIP       string    `json:"host_ip"`
	ContainerIP  string    `json:"container_ip,omitempty"`
	SuperPodRank string    `json:"-"`
	SuperPodID   string    `json:"-"`
}

// Instance is for annotation
type Instance struct { // Instance
	PodName    string `json:"pod_name"`  // pod Name
	ServerID   string `json:"server_id"` // serverdId
	ServerIP   string `json:"server_ip"` // server ip for A5
	HostIp     string `json:"host_ip"`   // hostIp
	SuperPodId int32  `json:"super_pod_id"`
	Devices    []Dev  `json:"devices"`      // dev
	RackID     int32  `json:"rack_id"`      // Rack id for A5
	SeverIndex string `json:"server_index"` // sever index for A5
}

// Dev to hccl
type Dev struct {
	DeviceID      string `json:"device_id"` // hccl deviceId
	DeviceIP      string `json:"device_ip"` // hccl deviceIp
	SuperDeviceID string `json:"super_device_id,omitempty"`
	// rank level info in rank table for A5
	LevelList []api.RankLevel `json:"levelList,omitempty"`
}

// Device in hccl.json
type Device struct {
	Dev
	RankID     string `json:"rank_id"`
	RackID     int32  `json:"rack_id,omitempty"`      // rack id for 910A5
	ServerID   string `json:"server_id,omitempty"`    // server id for 910A5
	SuperPodId int32  `json:"super_pod_id,omitempty"` // Super pod id for version 1.2
}
