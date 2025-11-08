/*
Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
*/

// Package common is common function or object of ranktable. The latest ranktable format of 910A5
package common

import "ascend-common/api"

// Rank for the rank info in rank table
type Rank struct {
	RankID    int                `json:"rank_id"`   // generate by operator
	LocalID   int                `json:"local_id"`  // from annotation, relying on devices.device_id field
	DeviceID  int                `json:"device_id"` // from annotation, relying on devices.device_id field
	LevelList []api.LevelElement `json:"level_list,omitempty"`
	Device    Dev                `json:"-"`
}
