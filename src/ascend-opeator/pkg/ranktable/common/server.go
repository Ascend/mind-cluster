/*
Copyright(C) 2024. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package common is common function or object of ranktable.
*/

package common

// Server to hccl
type Server struct {
	DeviceList  []*Device `json:"device"`    // device list in each server
	ServerID    string    `json:"server_id"` // server id, represented by ip address
	ContainerIP string    `json:"container_ip,omitempty"`
}

// Instance is for annotation
type Instance struct { // Instance
	PodName    string `json:"pod_name"`  // pod Name
	ServerID   string `json:"server_id"` // serverdId
	SuperPodId int32  `json:"super_pod_id"`
	Devices    []Dev  `json:"devices"` // dev
}

// Dev to hccl
type Dev struct {
	DeviceID      string `json:"device_id"` // hccl deviceId
	DeviceIP      string `json:"device_ip"` // hccl deviceIp
	SuperDeviceID string `json:"super_device_id,omitempty"`
}

// Device in hccl.json
type Device struct {
	Dev
	RankID string `json:"rank_id"`
}
