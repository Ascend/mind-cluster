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

// Package common defines common constants and types used by the toolkit backend.
package common

const (
    // OK indicates that the operation is successful.
    OK = 0

    // Client error codes
    // NilMessage indicates that the message is nil.
    NilMessage = 4000
    // NilHeader indicates that the message header is nil.
    NilHeader = 4001
    // NilPosition indicates that the destination or source position is nil.
    NilPosition = 4002
    // DstRoleIllegal indicates that the destination role is illegal.
    DstRoleIllegal = 4003
    // DstSrvRankIllegal indicates that the destination server rank is illegal.
    DstSrvRankIllegal = 4004
    // DstProcessRankIllegal indicates that the destination process rank is illegal.
    DstProcessRankIllegal = 4005
    // DstTypeIllegal indicates that the destination type is illegal.
    DstTypeIllegal = 4006
    // ClientErr indicates a client error.
    ClientErr = 4999

    // Server error codes
    // RecvBufNil indicates that the receive buffer is nil.
    RecvBufNil = 5000
    // RecvBufBusy indicates that the receive buffer is busy.
    RecvBufBusy = 5001
    // NoRoute indicates that there is no route.
    NoRoute = 5002
    // ExceedMaxRegistryNum indicates that the maximum registry number has been exceeded.
    ExceedMaxRegistryNum = 5003
    // ServerErr indicates a server error.
    ServerErr = 5999

    // Network error codes
    // NetworkSendLost indicates that the network send is lost.
    NetworkSendLost = 6000
    // NetworkAckLost indicates that the network ACK is lost.
    NetworkAckLost = 6001
    // NetStreamNotInited indicates that the network stream is not initialized.
    NetStreamNotInited = 6002
    // NetErr indicates a network error.
    NetErr = 6999
)
