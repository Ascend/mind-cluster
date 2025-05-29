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
	Role        string // The role of the node.
	ServerRank  string // The server rank of the node.
	ProcessRank string // The process rank of the node.
}

// Message represents a message structure.
type Message struct {
	Uuid    string    // The unique identifier of the message.
	BizType string    // The business type of the message.
	Src     *Position // The source position of the message.
	Dst     *Position // The destination position of the message.
	Body    string    // The body content of the message.
}

// Ack represents an acknowledgment structure.
type Ack struct {
	Uuid string    // The unique identifier of the acknowledgment.
	Code uint32    // The response code of the acknowledgment.
	Src  *Position // The source position of the acknowledgment.
}

// TaskNetConfig represents the network configuration of a task.
type TaskNetConfig struct {
	Pos          Position   // The position of the task node.
	ListenAddr   string     // The listening address of the task node.
	UpstreamAddr string     // The upstream address of the task node.
	ServerTLS    bool       // Whether to enable server TLS.
	ClientTLS    bool       // Whether to enable client TLS.
	TlsConf      *TLSConfig // The TLS configuration.
}

// TLSConfig represents the TLS configuration.
type TLSConfig struct {
	CA        string // The certificate authority file path.
	ServerKey string // The server private key file path.
	ServerCrt string // The server certificate file path.
	ClientKey string // The client private key file path.
	ClientCrt string // The client certificate file path.
}
