/* Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
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

package utils

import (
	"crypto/rand"
	"fmt"
	"sync/atomic"
	"time"
)

var uuidFailCounter uint64

func generateUUIDv4() (string, error) {
	uuid := make([]byte, 16)
	if _, err := rand.Read(uuid); err != nil {
		return "", err
	}
	uuid[6] = (uuid[6] & 0x0f) | 0x40 // Version 4
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // Variant is 10
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:16]), nil
}

// NewUUID generates and returns a new version-4 UUID string in canonical
// 8-4-4-4-12 hex format (for example "550e8400-e29b-41d4-a716-446655440000").
func NewUUID() string {
	id, err := generateUUIDv4()
	if err == nil {
		return id
	}
	c := atomic.AddUint64(&uuidFailCounter, 1)
	return fmt.Sprintf("genenrate uuid err-%d-%d", time.Now().UnixNano(), c)
}
