/*
 * Copyright (c) 2026. Huawei Technologies Co., Ltd. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package common for general collector
package common

import (
	"context"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/npu-exporter/v6/collector/container"
)

func TestBuildSchedule(t *testing.T) {
	convey.Convey("should skip nil collector when building schedule", t, func() {
		schedule := buildSchedule([]MetricsCollector{nil})
		convey.So(len(schedule.entries), convey.ShouldEqual, 0)
	})
}

func TestWaitForNextSignal(t *testing.T) {
	convey.Convey("should return wakeByContext when context cancelled", t, func() {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		result := waitForNextSignal(ctx, time.Hour, nil)
		convey.So(result, convey.ShouldEqual, wakeByContext)
	})

	convey.Convey("should return wakeByConfigReload when reload signal received", t, func() {
		ch := make(chan struct{}, 1)
		ch <- struct{}{}
		result := waitForNextSignal(context.Background(), time.Hour, ch)
		convey.So(result, convey.ShouldEqual, wakeByConfigReload)
	})

	convey.Convey("should return wakeByTimer when timer expires", t, func() {
		result := waitForNextSignal(context.Background(), 10*time.Millisecond, nil)
		convey.So(result, convey.ShouldEqual, wakeByTimer)
	})
}

type fakeCollector struct {
	name string
}

func (f *fakeCollector) Describe(_ chan<- *prometheus.Desc)               {}
func (f *fakeCollector) IsSupported(_ *NpuCollector) bool                 { return true }
func (f *fakeCollector) PreCollect(_ *NpuCollector, _ []HuaWeiAIChip)     {}
func (f *fakeCollector) CollectToCache(_ *NpuCollector, _ []HuaWeiAIChip) {}
func (f *fakeCollector) PostCollect(_ *NpuCollector)                      {}
func (f *fakeCollector) UpdatePrometheus(_ chan<- prometheus.Metric, _ *NpuCollector,
	_ map[int32]container.DevicesInfo, _ []HuaWeiAIChip) {
}
func (f *fakeCollector) UpdateTelegraf(_ map[string]map[string]interface{}, _ *NpuCollector,
	_ map[int32]container.DevicesInfo, _ []HuaWeiAIChip) map[string]map[string]interface{} {
	return nil
}

func TestScheduleUpdateNext(t *testing.T) {
	convey.Convey("TestScheduleUpdateNext", t, func() {
		convey.Convey("should update nextRun for ran collectors", func() {
			now := time.Now()
			collector := &fakeCollector{name: "k1"}
			schedule := collectorSchedule{entries: []scheduleEntry{
				{cacheKey: "k1", collector: collector, interval: 5 * time.Second, nextRun: now},
			}}
			ran := []MetricsCollector{collector}
			schedule.updateNext(ran, now)
			convey.So(schedule.entries[0].nextRun.After(now), convey.ShouldBeTrue)
		})

		convey.Convey("should set nextRun to far future when interval is collectOnce", func() {
			now := time.Now()
			collector := &fakeCollector{name: "k1"}
			schedule := collectorSchedule{entries: []scheduleEntry{
				{cacheKey: "k1", collector: collector, interval: CollectOnceInterval(), nextRun: now},
			}}
			ran := []MetricsCollector{collector}
			schedule.updateNext(ran, now)
			convey.So(schedule.entries[0].nextRun.Equal(farFutureTime), convey.ShouldBeTrue)
		})

		convey.Convey("should skip empty ran", func() {
			schedule := collectorSchedule{}
			schedule.updateNext(nil, time.Now())
		})
	})
}

func TestScheduleNextWaitDuration(t *testing.T) {
	convey.Convey("TestScheduleNextWaitDuration", t, func() {
		convey.Convey("should return 1 second when entries empty", func() {
			schedule := collectorSchedule{}
			convey.So(schedule.nextWaitDuration(), convey.ShouldEqual, time.Second)
		})

		convey.Convey("should return default interval when all entries in far future", func() {
			schedule := collectorSchedule{entries: []scheduleEntry{
				{collector: &fakeCollector{name: "k1"}, nextRun: farFutureTime},
			}}
			convey.So(schedule.nextWaitDuration(), convey.ShouldEqual, defaultGroupInterval)
		})
	})
}

func TestNotifyConfigReload(t *testing.T) {
	convey.Convey("should notify subscribers via subscribeConfigReload", t, func() {
		ch := subscribeConfigReload()
		defer unsubscribeConfigReload(ch)
		NotifyConfigReload()
		select {
		case <-ch:
		case <-time.After(time.Second):
			t.Fatal("expected reload signal")
		}
	})
}

const testCacheKey = "test-key"

func TestSetAndGetCollectorInterval(t *testing.T) {
	convey.Convey("TestSetAndGetCollectorInterval", t, func() {
		convey.Convey("should store and return interval when valid", func() {
			SetCollectorInterval(testCacheKey, 5*time.Second)
			defer collectorIntervalMap.Delete(testCacheKey)
			convey.So(GetCollectorInterval(testCacheKey, time.Second), convey.ShouldEqual, 5*time.Second)
		})

		convey.Convey("should return collectOnce interval when stored", func() {
			SetCollectorInterval(testCacheKey, CollectOnceInterval())
			defer collectorIntervalMap.Delete(testCacheKey)
			convey.So(GetCollectorInterval(testCacheKey, time.Second), convey.ShouldEqual, CollectOnceInterval())
		})

		convey.Convey("should ignore empty cache key", func() {
			SetCollectorInterval("", 5*time.Second)
			convey.So(GetCollectorInterval("", time.Second), convey.ShouldEqual, time.Second)
		})

		convey.Convey("should ignore zero or negative interval", func() {
			SetCollectorInterval(testCacheKey, -5*time.Second)
			defer collectorIntervalMap.Delete(testCacheKey)
			convey.So(GetCollectorInterval(testCacheKey, time.Second), convey.ShouldEqual, time.Second)
		})

		convey.Convey("should return fallback when key not exists", func() {
			collectorIntervalMap.Delete("not-exist")
			convey.So(GetCollectorInterval("not-exist", 3*time.Second), convey.ShouldEqual, 3*time.Second)
		})
	})
}

func TestCollectOnceInterval(t *testing.T) {
	convey.Convey("should return -1 second when CollectOnceInterval is called", t, func() {
		convey.So(CollectOnceInterval(), convey.ShouldEqual, -1*time.Second)
	})
}

func TestDisabledInterval(t *testing.T) {
	convey.Convey("should return -2 seconds when DisabledInterval is called", t, func() {
		convey.So(DisabledInterval(), convey.ShouldEqual, -2*time.Second)
	})
}
