// Copyright 2026 Huawei Technologies Co., Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package resources

import (
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"
)

func TestNewSignalNotifierWithSignals(t *testing.T) {
	convey.Convey("When NewSignalNotifier is called with signals", t, func() {
		notifier := NewSignalNotifier(syscall.SIGINT, syscall.SIGTERM)

		convey.Convey("Then it should return a valid SignalNotifier", func() {
			convey.So(notifier, convey.ShouldNotBeNil)

			sn, ok := notifier.(*signalNotifier)
			convey.So(ok, convey.ShouldBeTrue)
			convey.So(len(sn.signals), convey.ShouldEqual, 2)
		})
	})
}

func TestNewSignalNotifierEmptySignals(t *testing.T) {
	convey.Convey("When NewSignalNotifier is called with no signals", t, func() {
		notifier := NewSignalNotifier()

		convey.Convey("Then it should return a valid SignalNotifier", func() {
			convey.So(notifier, convey.ShouldNotBeNil)

			sn, ok := notifier.(*signalNotifier)
			convey.So(ok, convey.ShouldBeTrue)
			convey.So(len(sn.signals), convey.ShouldEqual, 0)
		})
	})
}

func TestSignalNotifierNotifyWithSignals(t *testing.T) {
	convey.Convey("When Notify is called with registered signals", t, func() {
		notifier := NewSignalNotifier(syscall.SIGINT, syscall.SIGTERM)

		sigChan := notifier.Notify()

		convey.Convey("Then it should return a non-nil channel", func() {
			convey.So(sigChan, convey.ShouldNotBeNil)
		})

		convey.Convey("Then it should receive sent signals", func() {
			go func() {
				time.Sleep(10 * time.Millisecond)
				process, _ := os.FindProcess(os.Getpid())
				_ = process.Signal(syscall.SIGINT)
			}()

			select {
			case sig := <-sigChan:
				convey.So(sig, convey.ShouldEqual, syscall.SIGINT)
			case <-time.After(2 * time.Second):
				convey.So(false, convey.ShouldBeTrue)
			}
		})
	})
}

func TestSignalNotifierNotifyEmptySignals(t *testing.T) {
	convey.Convey("When Notify is called with empty signals", t, func() {
		notifier := NewSignalNotifier()

		sigChan := notifier.Notify()

		convey.Convey("Then it should return a non-nil channel", func() {
			convey.So(sigChan, convey.ShouldNotBeNil)
		})

		convey.Convey("Then the channel should not block on close", func() {
			close(sigChan)
			convey.So(true, convey.ShouldBeTrue)
		})
	})
}

func TestSignalNotifierNotifyMultipleCalls(t *testing.T) {
	convey.Convey("When Notify is called multiple times", t, func() {
		notifier := NewSignalNotifier(syscall.SIGTERM)

		sigChan1 := notifier.Notify()
		sigChan2 := notifier.Notify()

		convey.Convey("Then it should return independent channels", func() {
			convey.So(sigChan1, convey.ShouldNotBeNil)
			convey.So(sigChan2, convey.ShouldNotBeNil)
			convey.So(sigChan1 == sigChan2, convey.ShouldBeFalse)
		})
	})
}
