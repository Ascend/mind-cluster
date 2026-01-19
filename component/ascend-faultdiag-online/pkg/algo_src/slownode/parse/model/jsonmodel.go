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

/*
Package model.
*/
package model

// OpGroupInfo 算子group信息
type OpGroupInfo struct {
	GroupName   string  `json:"group_name"`
	GroupRank   int64   `json:"group_rank"`
	GlobalRanks []int64 `json:"global_ranks"`
}

// JsonData json文件信息
type JsonData struct {
	Kind          int      `json:"Kind"`
	Flag          int      `json:"Flag"`
	SourceKind    int      `json:"SourceKind"`
	Timestamp     int64    `json:"Timestamp"`
	Id            int64    `json:"Id"`
	MSPTIObjectId ObjectId `json:"MsptiObjectId"`
	Name          string   `json:"Name"`
	ParseName     IntName  `json:"-"`
	Domain        string   `json:"Domain"`
}

// IntName Name字段对应的int值
type IntName struct {
	NameId       int64
	StreamId     int64
	IntCount     int64
	IntDataType  int64
	IntOpName    int64
	IntGroupName int64
}

// JsonName Json中Name字段
type JsonName struct {
	StreamId  string `json:"streamId"`
	Count     string `json:"count"`
	DataType  string `json:"dataType"`
	OpName    string `json:"opName"`
	GroupName string `json:"groupName"`
}

// ObjectId json文件信息内置属性ObjectId
type ObjectId struct {
	Pt Pt `json:"Pt"`
	Ds Ds `json:"Ds"`
}

// Pt json文件信息内置属性Pt
type Pt struct {
	ProcessId int `json:"ProcessId"`
	ThreadId  int `json:"ThreadId"`
}

// Ds json文件信息内置属性Ds
type Ds struct {
	DeviceId int `json:"DeviceId"`
	StreamId int `json:"StreamId"`
}
