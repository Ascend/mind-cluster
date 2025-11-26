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

// Package app unit tests for workflow
package app

import (
	"context"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"container-manager/pkg/reset/domain"
)

// TestResetMgr_Work tests the Work method
func TestResetMgr_Work(t *testing.T) {
	convey.Convey("Test ResetMgr Work", t, func() {
		r := &ResetMgr{
			resetCache: domain.NewNpuInResetCache(),
			countCache: domain.NewFailedResetCountCache(),
		}
		convey.Convey("When context is done", func() {
			ctx, cancel := context.WithCancel(context.Background())
			// Use channel for test synchronization
			done := make(chan bool)
			go func() {
				r.Work(ctx)
				done <- true
			}()
			// Cancel context should trigger exit
			cancel()

			select {
			case <-done:
				// Normal exit
			case <-time.After(time.Second):
				t.Fatal("Work method did not exit on context cancellation")
			}
		})

		convey.Convey("When ticker triggers processResetWork", func() {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			// Stub processResetWork to count calls
			callCount := 0

			const (
				testCallTimes    = 2
				testLoopInterval = time.Millisecond * 10
				testWaitTime     = time.Millisecond * 30
			)
			patches := gomonkey.ApplyPrivateMethod(r, "processResetWork", func() {
				callCount++
				if callCount >= testCallTimes {
					cancel() // Cancel after two calls
				}
			}).ApplyFuncReturn(time.NewTicker, time.NewTicker(testLoopInterval))
			defer patches.Reset()

			go r.Work(ctx)
			// Wait to ensure calls occur
			time.Sleep(testWaitTime)
			convey.So(callCount, convey.ShouldBeGreaterThanOrEqualTo, 1)
		})
	})
}

// TestResetMgr_ShutDown tests the ShutDown method
func TestResetMgr_ShutDown(t *testing.T) {
	convey.Convey("Test ResetMgr ShutDown", t, func() {
		r := &ResetMgr{}

		// ShutDown method should execute without error
		convey.So(func() { r.ShutDown() }, convey.ShouldNotPanic)
	})
}

// TestResetMgr_Name tests the Name method
func TestResetMgr_Name(t *testing.T) {
	convey.Convey("Test ResetMgr Name", t, func() {
		r := &ResetMgr{}
		name := r.Name()
		convey.So(name, convey.ShouldEqual, resetMgrModuleName)
	})
}

// TestResetMgr_Init tests the Init method
func TestResetMgr_Init(t *testing.T) {
	convey.Convey("Test ResetMgr Init", t, func() {
		r := &ResetMgr{}
		err := r.Init()
		convey.So(err, convey.ShouldBeNil)
	})
}

// TestNewResetMgr tests the NewResetMgr function
func TestNewResetMgr(t *testing.T) {
	convey.Convey("Test NewResetMgr", t, func() {
		module := NewResetMgr()
		convey.So(module, convey.ShouldNotBeNil)

		resetMgr, ok := module.(*ResetMgr)
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(resetMgr.resetCache, convey.ShouldNotBeNil)
		convey.So(resetMgr.countCache, convey.ShouldNotBeNil)
	})
}
