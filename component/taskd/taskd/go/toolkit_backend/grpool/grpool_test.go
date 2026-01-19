/*
Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.

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
// package grpool is a Go package that provides a simple and efficient way to manage goroutines.
package grpool

import (
	"context"
	"testing"
	"time"
)

func TestNewPool(t *testing.T) {
	ctx := context.Background()
	workers := uint32(2)
	pool := NewPool(workers, ctx)
	if pool == nil {
		t.Error("Expected new pool to be created, but got nil")
	}
	defer pool.Close()
}

func TestPoolSubmit(t *testing.T) {
	ctx := context.Background()
	pool := NewPool(1, ctx)
	defer pool.Close()

	done := make(chan struct{})
	fn := func(t Task) (interface{}, error) {
		close(done)
		return nil, nil
	}

	pool.Submit(fn)
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Error("Task took too long to complete")
	}
}

func TestPoolGroup(t *testing.T) {
	ctx := context.Background()
	pool := NewPool(1, ctx)
	defer pool.Close()

	group := pool.Group()
	if group == nil {
		t.Error("Expected group to be created, but got nil")
	}
}

func TestPoolClose(t *testing.T) {
	ctx := context.Background()
	pool := NewPool(1, ctx)
	pool.Close()
}
