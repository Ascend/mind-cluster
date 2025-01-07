/*
Copyright(C)2022-2025. Huawei Technologies Co.,Ltd. All rights reserved.

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
Package util is using for the total variable.
*/
package util

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"k8s.io/api/core/v1"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/config"
)

const (
	two   = 2
	three = 3
	four  = 4
	five  = 5
)

func TestChangeTopToIntArray(t *testing.T) {
	type args struct {
		topStr         string
		npuCardPreName string
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		{
			name: "01-ChangeTopToIntArray get int array",
			args: args{
				topStr:         fmt.Sprintf("%s0,%s1", NPU310PCardNamePre, NPU310PCardNamePre),
				npuCardPreName: NPU310PCardNamePre,
			},
			want: []int{0, 1},
		},
		{
			name: "02-ChangeTopToIntArray topStr is empty",
			args: args{
				topStr:         "",
				npuCardPreName: NPU310PCardNamePre,
			},
			want: []int{},
		},
		{
			name: "03-ChangeTopToIntArray string to int error",
			args: args{
				topStr:         fmt.Sprintf("%s0ab", NPU310PCardNamePre),
				npuCardPreName: NPU310PCardNamePre,
			},
			want: nil,
		},
		{
			name: "04-ChangeTopToIntArray get int array",
			args: args{
				topStr: fmt.Sprintf("%s0,%s1,%s2,%s3,%s4,%s5", NPU310PCardNamePre, NPU310PCardNamePre,
					NPU310PCardNamePre, NPU310PCardNamePre, NPU310PCardNamePre, NPU310PCardNamePre),
				npuCardPreName: NPU310PCardNamePre,
			},
			want: []int{0, 1, two, three, four, five},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ChangeTopToIntArray(tt.args.topStr, tt.args.npuCardPreName); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ChangeTopToIntArray() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsMapHasNPUResource(t *testing.T) {
	type args struct {
		resMap  map[v1.ResourceName]float64
		npuName string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "01-ChangeTopToIntArray has npu resource",
			args: args{
				resMap:  map[v1.ResourceName]float64{NPU910CardName: 1},
				npuName: NPU910CardName,
			},
			want: true,
		},
		{
			name: "02-ChangeTopToIntArray not exist npu resource",
			args: args{
				resMap:  map[v1.ResourceName]float64{},
				npuName: NPU910CardName,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsMapHasNPUResource(tt.args.resMap, tt.args.npuName); got != tt.want {
				t.Errorf("IsMapHasNPUResource() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChangeIntArrToStr(t *testing.T) {
	type args struct {
		top            []int
		npuCardPreName string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "01-ChangeIntArrToStr get int array",
			args: args{
				top:            []int{0, 1},
				npuCardPreName: NPU310CardNamePre,
			},
			want: fmt.Sprintf("%s0,%s1", NPU310CardNamePre, NPU310CardNamePre),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ChangeIntArrToStr(tt.args.top, tt.args.npuCardPreName); got != tt.want {
				t.Errorf("ChangeIntArrToStr() = %v, want %v", got, tt.want)
			}
		})
	}
}

type configFromSchedulerArgs struct {
	configKey      string
	configurations []config.Configuration
}

type configFromSchedulerConfigMapTest struct {
	name    string
	args    configFromSchedulerArgs
	want    *config.Configuration
	wantErr bool
}

const (
	configKey = "test"
)

func buildConfigFromSchedulerConfigMapTest() []configFromSchedulerConfigMapTest {
	return []configFromSchedulerConfigMapTest{
		{
			name: "01-GetConfigFromSchedulerConfigMap no configurations in scheduler configmap",
			args: configFromSchedulerArgs{
				configKey:      configKey,
				configurations: []config.Configuration{},
			},
			wantErr: true,
		},
		{
			name: "02-GetConfigFromSchedulerConfigMap get the configurations by name",
			args: configFromSchedulerArgs{
				configKey:      configKey,
				configurations: []config.Configuration{{Name: configKey}},
			},
			want:    &config.Configuration{Name: configKey},
			wantErr: false,
		},
		{
			name: "03-GetConfigFromSchedulerConfigMap cannot get configurations by name",
			args: configFromSchedulerArgs{
				configKey:      configKey,
				configurations: []config.Configuration{{Name: configKey + "0"}},
			},
			wantErr: true,
		},
		{
			name: "04-GetConfigFromSchedulerConfigMap compatible with old versions",
			args: configFromSchedulerArgs{
				configKey:      CMSelectorKey,
				configurations: []config.Configuration{{Name: configKey + "0"}, {Name: configKey + "1"}},
			},
			want:    nil,
			wantErr: true,
		},
	}
}

func TestGetConfigFromSchedulerConfigMap(t *testing.T) {
	tests := buildConfigFromSchedulerConfigMapTest()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetConfigFromSchedulerConfigMap(tt.args.configKey, tt.args.configurations)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetConfigFromSchedulerConfigMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetConfigFromSchedulerConfigMap() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMax(t *testing.T) {
	type args struct {
		x int
		y int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "01-Max return y when x < y",
			args: args{
				x: 0,
				y: 1,
			},
			want: 1,
		},
		{
			name: "02-Max return x when x > y",
			args: args{
				x: 1,
				y: 0,
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Max(tt.args.x, tt.args.y); got != tt.want {
				t.Errorf("Max() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMin(t *testing.T) {
	type args struct {
		x int
		y int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "01-Min return x when x < y",
			args: args{
				x: 0,
				y: 1,
			},
			want: 0,
		},
		{
			name: "02-Min return y when x > y",
			args: args{
				x: 1,
				y: 0,
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Min(tt.args.x, tt.args.y); got != tt.want {
				t.Errorf("Min() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsSliceContain(t *testing.T) {
	type args struct {
		keyword     interface{}
		targetSlice interface{}
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "01-IsSliceContain targetSlice is nil",
			args: args{
				keyword:     1,
				targetSlice: nil,
			},
			want: false,
		},
		{
			name: "02-IsSliceContain targetSlice is not slice or array",
			args: args{
				keyword:     1,
				targetSlice: 1,
			},
			want: false,
		},
		{
			name: "03-IsSliceContain slice contains keyword",
			args: args{
				keyword:     1,
				targetSlice: []int{0, 1},
			},
			want: true,
		},
		{
			name: "04-IsSliceContain slice not contains keyword",
			args: args{
				keyword:     1,
				targetSlice: []int{0},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsSliceContain(tt.args.keyword, tt.args.targetSlice); got != tt.want {
				t.Errorf("IsSliceContain() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRemoveSliceDuplicateElement(t *testing.T) {
	type args struct {
		languages []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "01-RemoveSliceDuplicateElement remove duplicate",
			args: args{
				languages: []string{"Go", "Python", "Python", "Go", "C++", "Java"},
			},
			want: []string{"Go", "Python", "C++", "Java"},
		},
		{
			name: "02-RemoveSliceDuplicateElement empty slice",
			args: args{
				languages: []string{},
			},
			want: []string{},
		},
		{
			name: "03-RemoveSliceDuplicateElement slice is nil",
			args: args{
				languages: nil,
			},
			want: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RemoveSliceDuplicateElement(tt.args.languages); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RemoveSliceDuplicateElement() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertErrSliceToError(t *testing.T) {
	type args struct {
		reErrors []error
	}
	tests := []struct {
		name string
		args args
		want error
	}{
		{
			name: "01-ConvertErrSliceToError",
			args: args{reErrors: []error{errors.New("test0"), errors.New("test1")}},
			want: errors.New("test0 test1"),
		},
		{
			name: "02-ConvertErrSliceToError",
			args: args{reErrors: []error{}},
			want: nil,
		},
		{
			name: "03-ConvertErrSliceToError",
			args: args{reErrors: nil},
			want: nil,
		},
		{
			name: "04-ConvertErrSliceToError",
			args: args{reErrors: []error{nil, nil}},
			want: nil,
		},
		{
			name: "05-ConvertErrSliceToError",
			args: args{reErrors: []error{nil, errors.New("test0")}},
			want: errors.New("test0"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ConvertErrSliceToError(tt.args.reErrors)
			if tt.want != nil && err == nil {
				t.Errorf("ConvertErrSliceToError() error = %v, wantErr %v", err, tt.want)
			}
			if tt.want == nil && err != nil {
				t.Errorf("ConvertErrSliceToError() error = %v, wantErr %v", err, tt.want)
			}
			if tt.want != nil && err != nil && tt.want.Error() != err.Error() {
				t.Errorf("ConvertErrSliceToError() error = %v, wantErr %v", err, tt.want)
			}
		})
	}
}

func TestChangeNodesToNodeMaps(t *testing.T) {
	type args struct {
		nodes []*api.NodeInfo
	}
	tests := []struct {
		name string
		args args
		want map[string]*api.NodeInfo
	}{
		{
			name: "01-ChangeNodesToNodeMaps",
			args: args{nodes: []*api.NodeInfo{{Name: "node0"}, {Name: "node1"}}},
			want: map[string]*api.NodeInfo{"node0": {Name: "node0"}, "node1": {Name: "node1"}},
		},
		{
			name: "02-ChangeNodesToNodeMaps empty slice",
			args: args{nodes: []*api.NodeInfo{}},
			want: nil,
		},
		{
			name: "03-ChangeNodesToNodeMaps nil slice",
			args: args{nodes: nil},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ChangeNodesToNodeMaps(tt.args.nodes); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ChangeNodesToNodeMaps() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetNpuNameFromJobRequire(t *testing.T) {
	type args struct {
		npuName string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "01-GetNpuNameFromJobRequire get npu name",
			args: args{npuName: AscendNPUCore},
			want: NPU310PCardName,
		},
		{
			name: "02-GetNpuNameFromJobRequire get npu name",
			args: args{npuName: NPU310CardName},
			want: NPU310CardName,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetNpuNameFromJobRequire(tt.args.npuName); got != tt.want {
				t.Errorf("GetNpuNameFromJobRequire() = %v, want %v", got, tt.want)
			}
		})
	}
}

type superPodInfoArgs struct {
	key            string
	configurations []config.Configuration
}

type superPodInfoArgsTest struct {
	name    string
	args    superPodInfoArgs
	want    int
	wantErr bool
}

func buildSuperPodInfoTest() []superPodInfoArgsTest {
	return []superPodInfoArgsTest{
		{
			name: "01-getSuperPodInfoFromConfig get super pod size",
			args: superPodInfoArgs{key: sizeOfSuperPodKey,
				configurations: []config.Configuration{
					{Name: CMInitParamKey, Arguments: map[string]string{sizeOfSuperPodKey: "1"}}},
			},
			want: 1,
		},
		{
			name: "02-getSuperPodInfoFromConfig get super pod size",
			args: superPodInfoArgs{key: reserveNodesKey,
				configurations: []config.Configuration{
					{Name: CMInitParamKey, Arguments: map[string]string{reserveNodesKey: "1"}}},
			},
			want: 1,
		},
		{
			name: "03-getSuperPodInfoFromConfig error",
			args: superPodInfoArgs{key: reserveNodesKey,
				configurations: []config.Configuration{},
			},
			wantErr: true,
		},
		{
			name: "04-getSuperPodInfoFromConfig reserveNodesKey not exist",
			args: superPodInfoArgs{key: reserveNodesKey,
				configurations: []config.Configuration{
					{Name: CMInitParamKey, Arguments: map[string]string{"abcd": "1"}}},
			},
			wantErr: true,
		},
		{
			name: "5-getSuperPodInfoFromConfig not number",
			args: superPodInfoArgs{key: reserveNodesKey,
				configurations: []config.Configuration{
					{Name: CMInitParamKey, Arguments: map[string]string{reserveNodesKey: "1xx"}}},
			},
			wantErr: true,
		},
		{
			name: "06-getSuperPodInfoFromConfig less than zero",
			args: superPodInfoArgs{key: reserveNodesKey,
				configurations: []config.Configuration{
					{Name: CMInitParamKey, Arguments: map[string]string{reserveNodesKey: "-1"}}},
			},
			wantErr: true,
		},
	}
}

func TestGetSuperPodInfoFromConfig(t *testing.T) {
	tests := buildSuperPodInfoTest()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getSuperPodInfoFromConfig(tt.args.key, tt.args.configurations)
			if (err != nil) != tt.wantErr {
				t.Errorf("getSuperPodInfoFromConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getSuperPodInfoFromConfig() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckStrInSlice(t *testing.T) {
	type args struct {
		str   string
		slice []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "01-CheckStrInSlice in slice",
			args: args{str: "a", slice: []string{"a", "b", "c"}},
			want: true,
		},
		{
			name: "02-CheckStrInSlice not in slice",
			args: args{str: "d", slice: []string{"a", "b", "c"}},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckStrInSlice(tt.args.str, tt.args.slice); got != tt.want {
				t.Errorf("CheckStrInSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeepCopyCmData(t *testing.T) {
	type args struct {
		cmData map[string]string
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "01-DeepCopyCmData",
			args: args{cmData: map[string]string{"a": "a", "b": "b"}},
			want: map[string]string{"a": "a", "b": "b"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DeepCopyCmData(tt.args.cmData); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DeepCopyCmData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsNodeReady(t *testing.T) {
	type args struct {
		node *v1.Node
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "01-IsNodeReady ready",
			args: args{node: &v1.Node{Status: v1.NodeStatus{Conditions: []v1.NodeCondition{
				{Type: v1.NodeReady, Status: v1.ConditionTrue}}}}},
			want: true,
		},
		{
			name: "02-IsNodeReady not ready",
			args: args{node: &v1.Node{Status: v1.NodeStatus{Conditions: []v1.NodeCondition{
				{Type: v1.NodeReady, Status: v1.ConditionFalse}}}}},
			want: false,
		},
		{
			name: "03-IsNodeReady not ready",
			args: args{node: &v1.Node{Status: v1.NodeStatus{Conditions: []v1.NodeCondition{
				{Type: v1.NodeReady, Status: v1.ConditionUnknown}}}}},
			want: false,
		},
		{
			name: "04-IsNodeReady not ready",
			args: args{node: &v1.Node{Status: v1.NodeStatus{Conditions: []v1.NodeCondition{
				{Type: v1.NodeMemoryPressure, Status: v1.ConditionUnknown}}}}},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNodeReady(tt.args.node); got != tt.want {
				t.Errorf("IsNodeReady() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRemoveCommonElement(t *testing.T) {
	type args struct {
		s1 []int
		s2 []int
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		{
			name: "01-RemoveCommonElement",
			args: args{s1: []int{0, 1}, s2: []int{0, 1}},
			want: []int{},
		},
		{
			name: "02-RemoveCommonElement",
			args: args{s1: []int{0, 1}, s2: []int{1}},
			want: []int{0},
		},
		{
			name: "03-RemoveCommonElement",
			args: args{s1: []int{0, 1}, s2: []int{three}},
			want: []int{0, 1},
		},
		{
			name: "04-RemoveCommonElement nil",
			args: args{s1: []int{0, 1}, s2: nil},
			want: []int{0, 1},
		},
		{
			name: "05-RemoveCommonElement nil",
			args: args{s1: nil, s2: []int{0, 1}},
			want: []int{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RemoveCommonElement(tt.args.s1, tt.args.s2); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RemoveCommonElement() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVResourceAdd(t *testing.T) {
	tests := []struct {
		name string
		res  VResource
		add  VResource
		want VResource
	}{
		{
			name: "01-Add return aicore=1 and aicpu=1",
			res:  VResource{Aicore: 0, Aicpu: 0},
			add:  VResource{Aicore: 1, Aicpu: 1},
			want: VResource{Aicore: 1, Aicpu: 1},
		},
		{
			name: "02-Add return aicore=2 and aicpu=2",
			res:  VResource{Aicore: 1, Aicpu: 1},
			add:  VResource{Aicore: 1, Aicpu: 1},
			want: VResource{Aicore: two, Aicpu: two},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.res.Add(tt.add)
			if !reflect.DeepEqual(tt.res, tt.want) {
				t.Errorf("VResource_Add() = %v, want %v", tt.res, tt.want)
			}
		})
	}
}

func TestVResourceSub(t *testing.T) {
	tests := []struct {
		name string
		res  VResource
		sub  VResource
		want VResource
	}{
		{
			name: "01-Sub return aicore=1 and aicpu=1",
			res:  VResource{Aicore: 1, Aicpu: 1},
			sub:  VResource{Aicore: 1, Aicpu: 1},
			want: VResource{Aicore: 0, Aicpu: 0},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.res.Sub(tt.sub)
			if !reflect.DeepEqual(tt.res, tt.want) {
				t.Errorf("VResource_Sub() = %v, want %v", tt.res, tt.want)
			}
		})
	}
}

func TestVResourceBeGreater(t *testing.T) {
	tests := []struct {
		name string
		res  VResource
		vr   VResource
		want bool
	}{
		{
			name: "01-BeGreater return true when res less than vr",
			res:  VResource{Aicore: 0, Aicpu: 0},
			vr:   VResource{Aicore: 1, Aicpu: 1},
			want: false,
		},
		{
			name: "02-BeGreater return false when res greater than vr",
			res:  VResource{Aicore: 1, Aicpu: 1},
			vr:   VResource{Aicore: 0, Aicpu: 0},
			want: true,
		},
		{
			name: "03-BeGreater return true when res equal to vr",
			res:  VResource{Aicore: 1, Aicpu: 1},
			vr:   VResource{Aicore: 1, Aicpu: 1},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.res.BeGreater(tt.vr); got != tt.want {
				t.Errorf("BeGreater() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetSizeOfSuperPod(t *testing.T) {
	tests := []struct {
		name    string
		conf    []config.Configuration
		want    int
		wantErr bool
	}{
		{
			name:    "01-GetSizeOfSuperPod return error when conf is empty",
			conf:    []config.Configuration{},
			want:    0,
			wantErr: true,
		},
		{
			name:    "02-GetSizeOfSuperPod return error when conf is nil",
			conf:    nil,
			want:    0,
			wantErr: true,
		},
		{
			name:    "03-GetSizeOfSuperPod return error when conf not exist init-params",
			conf:    []config.Configuration{{Name: "test"}},
			want:    0,
			wantErr: true,
		},
		{
			name: "04-GetSizeOfSuperPod return 1 when conf exist init-params",
			conf: []config.Configuration{{Name: CMInitParamKey, Arguments: map[string]string{sizeOfSuperPodKey: "1"}}},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetSizeOfSuperPod(tt.conf)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSizeOfSuperPod() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetSizeOfSuperPod() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetReserveNodes(t *testing.T) {
	tests := []struct {
		name    string
		conf    []config.Configuration
		want    int
		wantErr bool
	}{
		{
			name:    "01-GetReserveNodes return error when conf is empty",
			conf:    []config.Configuration{},
			want:    0,
			wantErr: true,
		},
		{
			name:    "02-GetReserveNodes return error when conf is nil",
			conf:    nil,
			want:    0,
			wantErr: true,
		},
		{
			name:    "03-GetReserveNodes return error when conf not exist init-params",
			conf:    []config.Configuration{{Name: "test"}},
			want:    0,
			wantErr: true,
		},
		{
			name: "04-GetReserveNodes return 1 when conf exist init-params",
			conf: []config.Configuration{{Name: CMInitParamKey, Arguments: map[string]string{reserveNodesKey: "1"}}},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetReserveNodes(tt.conf)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetReserveNodes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetReserveNodes() got = %v, want %v", got, tt.want)
			}
		})
	}
}
