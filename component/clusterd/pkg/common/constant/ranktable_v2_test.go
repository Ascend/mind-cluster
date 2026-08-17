// Copyright (c) Huawei Technologies Co., Ltd. 2026-2026. All rights reserved.

// Package constant tests for rank table v2.0 network info selection policy
package constant

import (
	"testing"
)

func set(keys ...string) map[string]struct{} {
	m := make(map[string]struct{}, len(keys))
	for _, k := range keys {
		m[k] = struct{}{}
	}
	return m
}

func TestGetNetInfo(t *testing.T) {
	cases := []struct {
		name    string
		types   map[string]struct{}
		custom  string
		want    NetInfo
		wantErr bool
	}{
		{"default empty", set(), "", NetInfo{}, false},
		{"default ROCE", set(PortAddrTypeRoCE), "", NetInfo{PortAddrTypeRoCE, ScaleOutTypeRoCE, RankAddrTypeIP}, false},
		{"default UBoE", set(PortAddrTypeUBoE), "", NetInfo{PortAddrTypeUBoE, ScaleOutTypeUBoE, RankAddrTypeIP}, false},
		{"default UBG", set(PortAddrTypeUBG), "", NetInfo{PortAddrTypeUBG, ScaleOutTypeUBoE, RankAddrTypeEID}, false},
		{"default UBoE+UBG prefers UBoE", set(PortAddrTypeUBoE, PortAddrTypeUBG), "", NetInfo{PortAddrTypeUBoE, ScaleOutTypeUBoE, RankAddrTypeIP}, false},
		{"default ROCE+UBoE+UBG prefers ROCE", set(PortAddrTypeRoCE, PortAddrTypeUBoE, PortAddrTypeUBG), "", NetInfo{PortAddrTypeRoCE, ScaleOutTypeRoCE, RankAddrTypeIP}, false},
		{"custom ROCE with ROCE", set(PortAddrTypeRoCE), ScaleOutTypeRoCE, NetInfo{PortAddrTypeRoCE, ScaleOutTypeRoCE, RankAddrTypeIP}, false},
		{"custom ROCE with empty types", set(), ScaleOutTypeRoCE, NetInfo{}, false},
		{"invalid custom", set(PortAddrTypeRoCE), "XXX", NetInfo{}, true},
		{"custom UBOE with UBG", set(PortAddrTypeUBG), ScaleOutTypeUBoE, NetInfo{PortAddrTypeUBG, ScaleOutTypeUBoE, RankAddrTypeEID}, false},
		{"empty custom with UBG uses default", set(PortAddrTypeUBG), "", NetInfo{PortAddrTypeUBG, ScaleOutTypeUBoE, RankAddrTypeEID}, false},
		{"custom label uses custom path", set(PortAddrTypeRoCE), ScaleOutTypeRoCE, NetInfo{PortAddrTypeRoCE, ScaleOutTypeRoCE, RankAddrTypeIP}, false},
		{"invalid custom propagates error", set(PortAddrTypeRoCE), "XXX", NetInfo{}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := GetNetInfo(tc.types, tc.custom)
			if (err != nil) != tc.wantErr || got != tc.want {
				t.Fatalf("got %+v, err %v, want %+v, wantErr %v", got, err, tc.want, tc.wantErr)
			}
		})
	}
}
