/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package rescheduling is using for HuaWei Ascend pin fault rescheduling.
*/
package plugin

import (
	"testing"
)

const (
	nodeTotalCoreNum56 = 56
	nodeTotalCpuNum49  = 49
)

type IsPodWholeCardArgs struct {
	realCardName string
}

type IsPodWholeCardTest struct {
	name string
	args IsPodWholeCardArgs
	want bool
}

func buildIsPodWholeCardTest() []IsPodWholeCardTest {
	tests := []IsPodWholeCardTest{
		{
			name: "01-IsPodWholeCardTest-is whole card",
			args: IsPodWholeCardArgs{
				realCardName: "0,1",
			},
			want: true,
		},
		{
			name: "02-IsPodWholeCardTest-not whold card",
			args: IsPodWholeCardArgs{realCardName: "0-vir04"},
			want: false,
		},
	}
	return tests
}

func TestIsPodWholeCard(t *testing.T) {
	tests := buildIsPodWholeCardTest()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsPodWholeCardFromAscendCore(tt.args.realCardName); got != tt.want {
				t.Errorf("IsPodWholeCard() = %v, want %v", got, tt.want)
			}
		})
	}
}
