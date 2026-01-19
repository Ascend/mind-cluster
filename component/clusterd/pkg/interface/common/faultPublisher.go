// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package common a series of common function
package common

var Publisher FaultPublisher

// FaultPublisher interface
type FaultPublisher interface {
	IsSubscribed(string, string) bool
}

// SetPublisher set publisher
func SetPublisher(publisher FaultPublisher) {
	Publisher = publisher
}
