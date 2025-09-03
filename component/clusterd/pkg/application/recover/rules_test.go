// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

// Package recover a series of state machine rules test function
package recover

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
)

func TestGetRules(t *testing.T) {
	convey.Convey("Test getRules", t, func() {
		ctl := &EventController{}
		convey.Convey("01-test getPreRules, should return slice", func() {
			preRules := ctl.getPreRules()
			convey.So(len(preRules) > 0, convey.ShouldBeTrue)
		})
		convey.Convey("02-test getFixRules, should return slice", func() {
			fixRules := ctl.getFixRules()
			convey.So(len(fixRules) > 0, convey.ShouldBeTrue)
		})
		convey.Convey("03-test getAfterRules, should return slice", func() {
			afterRules := ctl.getAfterRules()
			convey.So(len(afterRules) > 0, convey.ShouldBeTrue)
		})
		convey.Convey("04-test getBaseRules, should return slice", func() {
			baseRules := ctl.getBaseRules()
			convey.So(len(baseRules) > 0, convey.ShouldBeTrue)
		})
		convey.Convey("05-test getExtendPreRules, should return slice", func() {
			baseRules := ctl.getExtendPreRules()
			convey.So(len(baseRules) > 0, convey.ShouldBeTrue)
		})
		convey.Convey("06-test getOMRules, should return slice", func() {
			baseRules := ctl.getOMRules()
			convey.So(len(baseRules) > 0, convey.ShouldBeTrue)
		})
		convey.Convey("06-test getSwitchNicRules, should return slice", func() {
			baseRules := ctl.getSwitchNicRules()
			convey.So(len(baseRules) > 0, convey.ShouldBeTrue)
		})
		convey.Convey("07-test getStressTestRules, should return slice", func() {
			baseRules := ctl.getStressTestRules()
			convey.So(len(baseRules) > 0, convey.ShouldBeTrue)
		})
		convey.Convey("08-test getHotSwitchRules, should return slice", func() {
			baseRules := ctl.getHotSwitchRules()
			convey.So(len(baseRules) > 0, convey.ShouldBeTrue)
		})
	})
}

func TestEventControllerGetDPScaleStrategyRules(t *testing.T) {
	tests := []struct {
		name string
		want int
	}{
		{
			name: "get dp rules success",
			want: 19,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctl := &EventController{}
			assert.Equalf(t, tt.want, len(ctl.getDPScaleStrategyRules()), "getDPScaleStrategyRules()")
		})
	}
}
