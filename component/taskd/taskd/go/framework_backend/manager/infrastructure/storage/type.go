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

// Package storage for taskd manager backend data type
package storage

import (
	"time"

	"taskd/toolkit_backend/net/common"
)

// BaseMessage base message in manager
type BaseMessage struct {
	Header MsgHeader
	Body   MsgBody
}

// MsgHeader message header
type MsgHeader struct {
	BizType   string
	Uuid      string
	Src       *common.Position
	Timestamp time.Time
}

// MsgBody message body
type MsgBody struct {
	MsgType   string            `json:"msg_type"`
	Code      int32             `json:"code"`
	Message   string            `json:"message"`
	Extension map[string]string `json:"extension"`
}
