/*
Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.

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

// Package common defines common constants and types used by the toolkit backend.
package common

// Position represents the position information of a node.
type Position struct {
	Role        string `json:"role"`         // The role of the node.
	ServerRank  string `json:"server_rank"`  // The server rank of the node.
	ProcessRank string `json:"process_rank"` // The process rank of the node.
}

// Message represents a message structure.
type Message struct {
	Uuid    string    `json:"uuid"`     // The unique identifier of the message.
	BizType string    `json:"biz_type"` // The business type of the message.
	Src     *Position `json:"src"`      // The source position of the message.
	Dst     *Position `json:"dst"`      // The destination position of the message.
	Body    string    `json:"body"`     // The body content of the message.
}

// Ack represents an acknowledgment structure.
type Ack struct {
	Uuid string    `json:"uuid"` // The unique identifier of the acknowledgment.
	Code uint32    `json:"code"` // The response code of the acknowledgment.
	Src  *Position `json:"src"`  // The source position of the acknowledgment.
}

// TaskNetConfig represents the network configuration of a task.
type TaskNetConfig struct {
	Pos          Position   `json:"pos"`           // The position of the task node.
	ListenAddr   string     `json:"listen_addr"`   // The listening address of the task node.
	UpstreamAddr string     `json:"upstream_addr"` // The upstream address of the task node.
	EnableTls    bool       `json:"enable_tls"`    // Whether to enable TLS.
	TlsConf      *TLSConfig `json:"tls_conf"`      // The TLS configuration.
}

// TLSConfig represents the TLS configuration.
type TLSConfig struct {
	CA        string `json:"ca"`         // The certificate authority file path.
	ServerKey string `json:"server_key"` // The server private key file path.
	ServerCrt string `json:"server_crt"` // The server certificate file path.
	ClientKey string `json:"client_key"` // The client private key file path.
	ClientCrt string `json:"client_crt"` // The client certificate file path.
}
