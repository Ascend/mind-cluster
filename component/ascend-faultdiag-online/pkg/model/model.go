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

// Package model is common struct
package model

// HcclJson is the chip info the training job uses
type HcclJson struct {
	// ServerList is a node info list which traning job uses
	ServerList []struct {
		// ServerId is the ip of node
		ServerId string `json:"server_id"`
		// ServerSn is the series number of node
		ServerSn string `json:"server_sn"`
		// Device is a list including rank info
		Device []struct {
			// RankId is the rank id which traning job uses
			RankId string `json:"rank_id"`
		} `json:"device"`
	} `json:"server_list"`
}

// JobSummary is a struct for cm data in job-summary
type JobSummary struct {
	// JobName name of the slow node detection job
	JobName string
	// JobId uniqe id of each job, got from job-summary
	JobId string
	// Namespace is the namespace of the job
	Namespace string
	// JobStatus is the status of the job, pending/running/complete/failed
	JobStatus string
	// HcclJson is the chip info the training job using
	HcclJson HcclJson
	// Operator add/delete
	Operator string
}

// RescheduleReason the reason why training job reschedule
type RescheduleReason struct {
	// RescheduleReason the reason why training job reschedule
	RescheduleReason string `jons:"RescheduleReason"`
	// PodName the pod name of training job in
	PodName string `json:"PodName"`
	// NodeName the node name of training job in
	NodeName string `json:"NodeName"`
	// NodeRankIndex the rank index of the node
	NodeRankIndex string `json:"NodeRankIndex"`
}

// RescheduleRecords the record struct of reschedule
type RescheduleRecords struct {
	// LogFileFormatTime log file format time
	LogFileFormatTime string `json:"LogFileFormatTime"`
	// RescheduleTimeStamp reschedule timestamp
	RescheduleTimeStamp int64 `json:"RescheduleTimeStamp"`
	// ReasonOfTask reason why training job reschedule
	ReasonOfTask []RescheduleReason `json:"ReasonOfTask"`
}

// RescheduleData the reschedule data struct for training job
type RescheduleData struct {
	// JobId including namespace/jobName-jobId
	JobId string `json:"jobID"` // sample:default/default-test-mindspore-f4121ec4-590e-4cdc-a422-ac256b898659
	// TotalRescheduleTimes the total reschedule count
	TotalRescheduleTimes int `json:"TotalRescheduleTimes"`
	// RescheduleRecords the records of reschedule
	RescheduleRecords []RescheduleRecords `json:"RescheduleRecords"`
}
