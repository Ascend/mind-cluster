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
	"context"
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/hwlog"
)

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	err := hwlog.InitRunLogger(&hwLogConfig, context.Background())
	if err != nil {
		return
	}
}

func TestTokenBucketHasToken(t *testing.T) {
	convey.Convey("test TokenBucket HasToken", t, func() {
		convey.Convey("new bucket should have tokens", func() {
			tb := NewTokenBucket()
			convey.So(tb.HasToken(), convey.ShouldBeTrue)
			convey.So(tb.GetTokens(), convey.ShouldEqual, tokenMaxCount)
		})
		convey.Convey("bucket with zero tokens should return false", func() {
			tb := NewTokenBucket()
			for i := 0; i < tokenMaxCount; i++ {
				tb.Consume()
			}
			convey.So(tb.HasToken(), convey.ShouldBeFalse)
		})
	})
}

func TestTokenBucketConsume(t *testing.T) {
	convey.Convey("test TokenBucket Consume", t, func() {
		convey.Convey("consume should decrease tokens", func() {
			tb := NewTokenBucket()
			convey.So(tb.Consume(), convey.ShouldBeTrue)
			convey.So(tb.GetTokens(), convey.ShouldEqual, tokenMaxCount-1)
		})
		convey.Convey("consume on empty bucket should return false", func() {
			tb := NewTokenBucket()
			for i := 0; i < tokenMaxCount; i++ {
				tb.Consume()
			}
			convey.So(tb.Consume(), convey.ShouldBeFalse)
		})
	})
}

func TestTokenBucketRefill(t *testing.T) {
	convey.Convey("test TokenBucket refill", t, func() {
		convey.Convey("refill after interval should restore tokens", func() {
			tb := NewTokenBucket()
			for i := 0; i < tokenMaxCount; i++ {
				tb.Consume()
			}
			tb.lastRefill = time.Now().Add(-tokenRefillInterval - time.Second)
			convey.So(tb.HasToken(), convey.ShouldBeTrue)
			convey.So(tb.GetTokens(), convey.ShouldEqual, tokenMaxCount)
		})
		convey.Convey("refill before interval should not restore tokens", func() {
			tb := NewTokenBucket()
			tb.Consume()
			tb.lastRefill = time.Now().Add(-time.Hour)
			convey.So(tb.GetTokens(), convey.ShouldEqual, tokenMaxCount-1)
		})
	})
}

func TestTokenBucketMgrHasToken(t *testing.T) {
	convey.Convey("test TokenBucketMgr HasToken", t, func() {
		convey.Convey("new mgr should have tokens for any logicID", func() {
			mgr := NewTokenBucketMgr()
			convey.So(mgr.HasToken(0), convey.ShouldBeTrue)
			convey.So(mgr.HasToken(1), convey.ShouldBeTrue)
		})
		convey.Convey("different logicIDs should have independent buckets", func() {
			mgr := NewTokenBucketMgr()
			for i := 0; i < tokenMaxCount; i++ {
				mgr.ConsumeToken(0)
			}
			convey.So(mgr.HasToken(0), convey.ShouldBeFalse)
			convey.So(mgr.HasToken(1), convey.ShouldBeTrue)
		})
	})
}

func TestTokenBucketMgrConsumeToken(t *testing.T) {
	convey.Convey("test TokenBucketMgr ConsumeToken", t, func() {
		convey.Convey("consume should decrease tokens for specific logicID", func() {
			mgr := NewTokenBucketMgr()
			convey.So(mgr.ConsumeToken(0), convey.ShouldBeTrue)
			convey.So(mgr.GetTokens(0), convey.ShouldEqual, tokenMaxCount-1)
		})
		convey.Convey("consume on exhausted bucket should return false", func() {
			mgr := NewTokenBucketMgr()
			for i := 0; i < tokenMaxCount; i++ {
				mgr.ConsumeToken(0)
			}
			convey.So(mgr.ConsumeToken(0), convey.ShouldBeFalse)
		})
	})
}

func TestTokenBucketGetTokens(t *testing.T) {
	convey.Convey("test TokenBucket GetTokens", t, func() {
		convey.Convey("new bucket should return max tokens", func() {
			tb := NewTokenBucket()
			convey.So(tb.GetTokens(), convey.ShouldEqual, tokenMaxCount)
		})
		convey.Convey("after partial consume should return remaining tokens", func() {
			tb := NewTokenBucket()
			tb.Consume()
			tb.Consume()
			convey.So(tb.GetTokens(), convey.ShouldEqual, tokenMaxCount-2)
		})
		convey.Convey("after full consume should return zero", func() {
			tb := NewTokenBucket()
			for i := 0; i < tokenMaxCount; i++ {
				tb.Consume()
			}
			convey.So(tb.GetTokens(), convey.ShouldEqual, 0)
		})
		convey.Convey("refill should reset tokens to max", func() {
			tb := NewTokenBucket()
			for i := 0; i < tokenMaxCount; i++ {
				tb.Consume()
			}
			tb.lastRefill = time.Now().Add(-tokenRefillInterval - time.Second)
			convey.So(tb.GetTokens(), convey.ShouldEqual, tokenMaxCount)
		})
	})
}

func TestTokenBucketMgrGetTokens(t *testing.T) {
	convey.Convey("test TokenBucketMgr GetTokens", t, func() {
		convey.Convey("new logicID should return max tokens", func() {
			mgr := NewTokenBucketMgr()
			convey.So(mgr.GetTokens(0), convey.ShouldEqual, tokenMaxCount)
			convey.So(mgr.GetTokens(1), convey.ShouldEqual, tokenMaxCount)
		})
		convey.Convey("after consume should return remaining tokens", func() {
			mgr := NewTokenBucketMgr()
			mgr.ConsumeToken(0)
			mgr.ConsumeToken(0)
			convey.So(mgr.GetTokens(0), convey.ShouldEqual, tokenMaxCount-2)
		})
		convey.Convey("independent buckets should not affect each other", func() {
			mgr := NewTokenBucketMgr()
			for i := 0; i < tokenMaxCount; i++ {
				mgr.ConsumeToken(0)
			}
			convey.So(mgr.GetTokens(0), convey.ShouldEqual, 0)
			convey.So(mgr.GetTokens(1), convey.ShouldEqual, tokenMaxCount)
		})
	})
}
