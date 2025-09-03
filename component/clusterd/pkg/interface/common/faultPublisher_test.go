// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package common a series of common interface
package common

import (
	"reflect"
	"testing"
)

type mockFaultPublisher struct {
	isSubscribed bool
}

func (m *mockFaultPublisher) IsSubscribed(topic, subscriber string) bool {
	return m.isSubscribed
}

func TestSetPublisher(t *testing.T) {
	t.Run("SetValidPublisher", func(t *testing.T) {
		mockPub := &mockFaultPublisher{isSubscribed: true}
		SetPublisher(mockPub)

		if Publisher == nil {
			t.Error("Expected Publisher to be set, but got nil")
		}

		if !reflect.DeepEqual(Publisher, mockPub) {
			t.Error("Expected Publisher to be the same as mockPub")
		}

		if !Publisher.IsSubscribed("testTopic", "testSubscriber") {
			t.Error("Expected mock publisher to return true for IsSubscribed")
		}
	})

	t.Run("SetNilPublisher", func(t *testing.T) {
		SetPublisher(nil)

		if Publisher != nil {
			t.Error("Expected Publisher to be nil, but got a value")
		}
	})

	t.Run("ReplacePublisher", func(t *testing.T) {
		mockPub1 := &mockFaultPublisher{isSubscribed: true}
		mockPub2 := &mockFaultPublisher{isSubscribed: false}

		SetPublisher(mockPub1)
		if Publisher == nil || Publisher.IsSubscribed("", "") != true {
			t.Error("First publisher not set correctly")
		}

		SetPublisher(mockPub2)
		if Publisher == nil || Publisher.IsSubscribed("", "") != false {
			t.Error("Publisher not replaced correctly")
		}
	})
}
