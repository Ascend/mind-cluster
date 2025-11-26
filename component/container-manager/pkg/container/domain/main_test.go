package domain

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"ascend-common/common-utils/hwlog"
)

const (
	len0 = 0
	len1 = 1
	len2 = 2
	len3 = 3
	len4 = 4

	devId0   = 0
	devId1   = 1
	devId2   = 2
	devId3   = 3
	eventId0 = 0x123
	eventId1 = 0x456

	ctrId0 = "ctr0"
	ctrId1 = "ctr1"
	ctrId2 = "ctr2"
	ctrId3 = "ctr3"
	ctrId4 = "ctr4" // not exist

	testCtrNs = "moby"
)

var testErr = errors.New("test error")

func TestMain(m *testing.M) {
	if err := setup(); err != nil {
		return
	}
	code := m.Run()
	fmt.Printf("exit_code = %v\n", code)
}

func setup() error {
	if err := initLog(); err != nil {
		return err
	}
	NewDevCache([]int32{devId0, devId1, devId2, devId3})
	GetCtrInfo()
	return nil
}

func initLog() error {
	logConfig := &hwlog.LogConfig{
		OnlyToStdout: true,
	}
	if err := hwlog.InitRunLogger(logConfig, context.Background()); err != nil {
		fmt.Printf("init hwlog failed, %v\n", err)
		return errors.New("init hwlog failed")
	}
	return nil
}
