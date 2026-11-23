/* Copyright 2025 Huawei Technologies Co., Ltd.

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
Package model.
*/
package model

// OpGroupInfo contains information about an operator group
type OpGroupInfo struct {
	// GroupName is the name of the operator communication domain
	GroupName string `json:"group_name"`
	// GroupRank is the ranking of the card within the parallel domain
	GroupRank int64 `json:"group_rank"`
	// GlobalRanks is a slice of rank IDs included in the parallel domain within the node
	GlobalRanks []int64 `json:"global_ranks"`
}

// JsonData contains information from JSON files
type JsonData struct {
	// Kind is the "kind" field of profiling data
	Kind int `json:"Kind"`
	// Flag indicates whether the data is start or end data
	Flag int `json:"Flag"`
	// SourceKind indicates the data source: device or host
	SourceKind int `json:"SourceKind"`
	// Timestamp is the data's timestamp
	Timestamp int64 `json:"Timestamp"`
	// Id is the data identifier
	Id int64 `json:"Id"`
	// MSPTIObjectId is the identifier of the MSPTIObject
	MSPTIObjectId ObjectId `json:"MsptiObjectId"`
	// Name is the "name" field in the JSON
	Name string `json:"Name"`
	// ParseName contains parsed information from the JSON's "name" field
	ParseName IntName `json:"-"`
	// Domain indicates the data type
	Domain string `json:"Domain"`
}

// IntName represents the integer values corresponding to the Name field
type IntName struct {
	// NameId is the identifier of the name
	NameId int64
	// StreamId is the stream identifier
	StreamId int64
	// IntCount is the count value
	IntCount int64
	// IntDataType is the data type
	IntDataType int64
	// IntOpName is the operator name
	IntOpName int64
	// IntGroupName is the group name
	IntGroupName int64
}

// JsonName represents the "Name" field in JSON data
type JsonName struct {
	// StreamId is the stream identifier
	StreamId string `json:"streamId"`
	// Count is the count value
	Count string `json:"count"`
	// DataType is the data type
	DataType string `json:"dataType"`
	// OpName is the operator name
	OpName string `json:"opName"`
	// GroupName is the group name
	GroupName string `json:"groupName"`
}

// ObjectId is a built-in property "ObjectId" in JSON file information
type ObjectId struct {
	// Pt contains Pt information
	Pt Pt `json:"Pt"`
	// Ds contains Ds information
	Ds Ds `json:"Ds"`
}

// Pt is a built-in property "Pt" in JSON file information
type Pt struct {
	// ProcessId is the process identifier
	ProcessId int `json:"ProcessId"`
	// ThreadId is the thread identifier
	ThreadId int `json:"ThreadId"`
}

// Ds is a built-in property "Ds" in JSON file information
type Ds struct {
	// DeviceId is the device identifier
	DeviceId int `json:"DeviceId"`
	// StreamId is the stream identifier
	StreamId int `json:"StreamId"`
}
