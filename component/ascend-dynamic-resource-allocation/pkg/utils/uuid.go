/*
 * Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * 		http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package utils

import (
	"math/rand"

	"github.com/google/uuid"
)

func generateUUIDs(seed string, count int) []string {
	rand := rand.New(rand.NewSource(hash(seed)))

	uuids := make([]string, count)
	for i := 0; i < count; i++ {
		charset := make([]byte, consts.CharsetLength)
		rand.Read(charset)
		uuid, _ := uuid.FromBytes(charset)
		uuids[i] = consts.NpuPrefix + uuid.String()
	}

	return uuids
}

func hash(s string) int64 {
	h := int64(0)
	for _, c := range s {
		// Hash calculation using prime number 31 to get seed
		h = consts.HashPrime*h + int64(c)
	}
	return h
}
