/*
Copyright(C)2025. Huawei Technologies Co.,Ltd. All rights reserved.

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

/*
Package resulthandler is using for handle hccsping mesh result
*/

package resulthandler

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"

	"nodeD/pkg/pingmesh/types"
	_ "nodeD/pkg/testtool"
)

func TestHandle(t *testing.T) {
	convey.Convey("TestHandle", t, func() {
		var expected atomic.Int32
		convey.Convey("01-nil Stop chan will do nothing", func() {
			h := NewAggregatedHandler()
			h.Handle(nil)
			convey.So(expected.Load(), convey.ShouldEqual, 0)
		})
		convey.Convey("02-Stop chan will do nothing", func() {
			h := NewAggregatedHandler(func(*types.HccspingMeshResult) error {
				expected.Add(1)
				return nil
			})
			stopChan := make(chan struct{})
			go h.Handle(stopChan)
			h.Receive(&types.HccspingMeshResult{})
			time.Sleep(time.Second)
			convey.So(expected.Load(), convey.ShouldEqual, 1)
		})
	})
}

func TestHandleDeadlock(t *testing.T) {
	// 1. 初始化目标对象
	h := NewAggregatedHandler(func(*types.HccspingMeshResult) error {
		return nil
	})

	// 2. 用 WaitGroup 跟踪目标 Run 方法是否退出
	stopChan := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		h.Handle(stopChan)
	}()

	// 3. 等待一小段时间，确保主循环进入阻塞状态（队列空时 Get() 阻塞）
	time.Sleep(1 * time.Second)

	// 4. 发送退出信号（关闭 stop 通道）
	close(stopChan)

	// 5. 监控是否超时退出（若超时，视为死锁）
	timeout := time.After(1 * time.Second)
	done := make(chan struct{})
	go func() {
		wg.Wait() // 等待目标 Handle 方法退出
		close(done)
	}()

	select {
	case <-done:
		// 正常退出，无死锁
		assert.True(t, true, "程序正常响应退出信号，无死锁")
	case <-timeout:
		assert.Fail(t, "程序超时未退出，存在死锁（无法响应 stop 信号）")
	}
}
