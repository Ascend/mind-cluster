package plugin

import (
	"testing"

	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
)

const (
	fakeServerNum     = 32
	fakeSliceNum      = 4
	fakeJobId         = " test"
	fakeEnableNodeNum = 18
)

func fakeNormalTorList(enableNodeNum int, jobUid api.JobID) *TorList {
	var tmpTors = []*Tor{}
	taskNodeNum := 0
	for i := 0; i < fakeServerNum; i++ {
		tmpTor := &Tor{}
		for j := 0; j < fakeSliceNum; j++ {
			if taskNodeNum < enableNodeNum {
				tmpTor.Servers = append(tmpTor.Servers, &Server{CurrentJob: &jobUid, SliceId: j})
				taskNodeNum++
				continue
			}
			tmpTor.Servers = append(tmpTor.Servers, &Server{SliceId: j})
		}
		tmpTors = append(tmpTors, tmpTor)
	}
	return &TorList{Tors: tmpTors}
}

type getLogicTorsAndFullTorNumTest struct {
	name           string
	taskRow        int
	taskColumn     int
	sliceNum       int
	wantFullTorNum int
}

func buildGetLogicTorsAndFullTorNumTestCase() []getLogicTorsAndFullTorNumTest {
	return []getLogicTorsAndFullTorNumTest{
		{
			name:           "01 will return nil 0 when SliceId is over 128",
			sliceNum:       util.MaxSliceNum + 1,
			wantFullTorNum: 0,
		},
		{
			name:           "02 will return nil 0 when taskRow is too large",
			sliceNum:       fakeSliceNum,
			taskRow:        util.NPUIndex5,
			taskColumn:     util.NPUIndex1,
			wantFullTorNum: 0,
		},
		{
			name:           "03 will return nil 0 when not enough logic tor",
			sliceNum:       fakeSliceNum,
			taskRow:        util.NPUIndex4,
			taskColumn:     util.NPUIndex1,
			wantFullTorNum: 4,
		},
	}
}

func TestTorListGetLogicTorsAndFullTorNum(t *testing.T) {
	tl := fakeNormalTorList(fakeEnableNodeNum, fakeJobId)
	for _, tt := range buildGetLogicTorsAndFullTorNumTestCase() {
		t.Run(tt.name, func(t *testing.T) {
			_, got1 := tl.GetLogicTorsAndFullTorNum(fakeJobId, tt.taskColumn, tt.taskRow, tt.sliceNum)
			if got1 != tt.wantFullTorNum {
				t.Errorf("GetLogicTorsAndFullTorNum() got1 = %v, want %v", got1, tt.wantFullTorNum)
			}
		})
	}
}
