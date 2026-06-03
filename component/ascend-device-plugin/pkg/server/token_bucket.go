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

package server

import (
	"sync"
	"time"

	"ascend-common/common-utils/hwlog"
)

const (
	tokenRefillInterval = 6 * time.Hour
	tokenMaxCount       = 3
)

type TokenBucket struct {
	mu         sync.Mutex
	tokens     int
	lastRefill time.Time
}

func NewTokenBucket() *TokenBucket {
	return &TokenBucket{
		tokens:     tokenMaxCount,
		lastRefill: time.Now(),
	}
}

func (tb *TokenBucket) refill() {
	now := time.Now()
	if now.Sub(tb.lastRefill) >= tokenRefillInterval {
		tb.tokens = tokenMaxCount
		tb.lastRefill = now
	}
}

func (tb *TokenBucket) HasToken() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.refill()
	return tb.tokens > 0
}

func (tb *TokenBucket) Consume() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.refill()
	if tb.tokens <= 0 {
		hwlog.RunLog.Debugf("token bucket empty")
		return false
	}
	tb.tokens--
	hwlog.RunLog.Infof("consume token, tokens: %d", tb.tokens)
	return true
}

func (tb *TokenBucket) GetTokens() int {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.refill()
	return tb.tokens
}

type TokenBucketMgr struct {
	buckets sync.Map
}

func NewTokenBucketMgr() *TokenBucketMgr {
	return &TokenBucketMgr{}
}

func (m *TokenBucketMgr) getOrCreateBucket(logicID int32) *TokenBucket {
	val, _ := m.buckets.LoadOrStore(logicID, NewTokenBucket())
	return val.(*TokenBucket)
}

func (m *TokenBucketMgr) HasToken(logicID int32) bool {
	return m.getOrCreateBucket(logicID).HasToken()
}

func (m *TokenBucketMgr) ConsumeToken(logicID int32) bool {
	hwlog.RunLog.Debugf("consume token for logicID: %d", logicID)
	return m.getOrCreateBucket(logicID).Consume()
}

func (m *TokenBucketMgr) GetTokens(logicID int32) int {
	return m.getOrCreateBucket(logicID).GetTokens()
}
