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

// Package roceping for ping by icmp in RoCE mesh net between super pods in A5
package roceping

import (
	"encoding/json"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/fake"

	"ascend-common/api"
	"ascend-common/api/slownet"
	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	"nodeD/pkg/common"
	"nodeD/pkg/kubeclient"
	"nodeD/pkg/pingmesh/types"
)

const (
	superPodID    = 1
	serverIndex   = 1
	rackID        = 1
	devType       = ""
	label800IA5x8 = "850-SuperPod-Atlas-8"
	label900POD   = "950-SuperPod-Atlas-8"
)

func TestCalcAvgLossRate(t *testing.T) {
	convey.Convey("test calcAvgLossRate", t, func() {
		convey.Convey("01-should return zero when suc and fail is zero", func() {
			ret := calcAvgLossRate(0, 0)
			convey.So(ret, convey.ShouldEqual, "0.000")
		})
		convey.Convey("02-should return one when suc is zero and fail is not zero", func() {
			ret := calcAvgLossRate(0, 1)
			convey.So(ret, convey.ShouldEqual, "1.000")
		})
		convey.Convey("03-should return not zero when suc and fail is not zero", func() {
			ret := calcAvgLossRate(1, 1)
			convey.So(ret, convey.ShouldEqual, "0.500")
		})
	})
}

func TestCalcTP95Value(t *testing.T) {
	convey.Convey("test calcTP95Value", t, func() {
		convey.Convey("01-should return -1 when input arr is nil", func() {
			ret := calcTP95Value(nil)
			convey.So(ret, convey.ShouldEqual, -1)
		})
		convey.Convey("02-should return element at index 0 when len(arr)=1", func() {
			ret := calcTP95Value([]int64{5})
			expected := 5
			convey.So(ret, convey.ShouldEqual, expected)
		})
		convey.Convey("03-should return element at index 1 when len(arr)=2", func() {
			ret := calcTP95Value([]int64{6, 5})
			expected := 6
			convey.So(ret, convey.ShouldEqual, expected)
		})
		convey.Convey("04-should return element at index 2 when len(arr)=3", func() {
			ret := calcTP95Value([]int64{6, 5, 4})
			expected := 6
			convey.So(ret, convey.ShouldEqual, expected)
		})
		convey.Convey("05-should return element at index 3 when len(arr)=4", func() {
			ret := calcTP95Value([]int64{6, 5, 4, 3})
			expected := 6
			convey.So(ret, convey.ShouldEqual, expected)
		})
	})
}

func TestCalcAppendModeAndOpenFlag(t *testing.T) {
	convey.Convey("test PingManager method calcAppendModeAndOpenFlag", t, func() {
		convey.Convey("01-should return false when time period is small", func() {
			m := &PingManager{}
			expectedOpenFlag := os.O_RDWR | os.O_CREATE | os.O_TRUNC
			appendMode, openFlag := m.calcAppendModeAndOpenFlag()
			convey.So(appendMode, convey.ShouldBeFalse)
			convey.So(openFlag, convey.ShouldEqual, expectedOpenFlag)
		})
		convey.Convey("02-should return true when time period is big enough", func() {
			lastSaveTime := time.Now().UnixMilli() - savePeriodMillSec + 1
			m := &PingManager{lastSaveTime: lastSaveTime}
			expectedOpenFlag := os.O_WRONLY | os.O_APPEND | os.O_CREATE
			appendMode, openFlag := m.calcAppendModeAndOpenFlag()
			convey.So(appendMode, convey.ShouldBeTrue)
			convey.So(openFlag, convey.ShouldEqual, expectedOpenFlag)
			convey.So(m.lastSaveTime, convey.ShouldEqual, lastSaveTime)
		})
	})
}

func TestPrepareResultFilePathsErrCase(t *testing.T) {
	convey.Convey("test PingManager method prepareResultFilePaths err case", t, func() {
		convey.Convey("01-should return err when getting root path failed", func() {
			m := &PingManager{}
			_, _, err := m.prepareResultFilePaths(false)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("02-should return err when check path failed", func() {
			m := &PingManager{}
			patch := gomonkey.ApplyFuncReturn(slownet.GetRasNetRootPath, "/", nil)
			defer patch.Reset()
			patchCheck := gomonkey.ApplyFuncReturn(utils.CheckPath, "/", errors.New("path invalid"))
			defer patchCheck.Reset()
			_, _, err := m.prepareResultFilePaths(false)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("03-should return err when remove failed", func() {
			m := &PingManager{}
			patch := gomonkey.ApplyFuncReturn(slownet.GetRasNetRootPath, "/", nil)
			defer patch.Reset()
			patchCheck := gomonkey.ApplyFuncReturn(utils.CheckPath, "/", nil)
			defer patchCheck.Reset()
			patchExist := gomonkey.ApplyFuncReturn(utils.IsLexist, true)
			defer patchExist.Reset()
			patchRemove := gomonkey.ApplyFuncReturn(os.Remove, errors.New("remove failed"))
			defer patchRemove.Reset()
			_, _, err := m.prepareResultFilePaths(false)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("04-should return err when remove failed", func() {
			m := &PingManager{}
			patch := gomonkey.ApplyFuncReturn(slownet.GetRasNetRootPath, "/", nil)
			defer patch.Reset()
			patchCheck := gomonkey.ApplyFuncReturn(utils.CheckPath, "/", nil)
			defer patchCheck.Reset()
			patchExist := gomonkey.ApplyFuncReturn(utils.IsLexist, true)
			defer patchExist.Reset()
			patchRemove := gomonkey.ApplyFuncReturn(os.Remove, nil)
			defer patchRemove.Reset()
			patchRename := gomonkey.ApplyFuncReturn(os.Rename, errors.New("rename failed"))
			defer patchRename.Reset()
			_, _, err := m.prepareResultFilePaths(false)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestPrepareResultFilePathsSuccessCase(t *testing.T) {
	convey.Convey("test PingManager method prepareResultFilePaths success case", t, func() {
		convey.Convey("01-should return nil when all action success", func() {
			m := &PingManager{}
			patch := gomonkey.ApplyFuncReturn(slownet.GetRasNetRootPath, "/", nil)
			defer patch.Reset()
			patchCheck := gomonkey.ApplyFuncReturn(utils.CheckPath, "/", nil)
			defer patchCheck.Reset()
			patchExist := gomonkey.ApplyFuncReturn(utils.IsLexist, true)
			defer patchExist.Reset()
			patchRemove := gomonkey.ApplyFuncReturn(os.Remove, nil)
			defer patchRemove.Reset()
			patchRename := gomonkey.ApplyFuncReturn(os.Rename, nil)
			defer patchRename.Reset()
			csvFile, csvBackFile, err := m.prepareResultFilePaths(false)
			convey.So(err, convey.ShouldBeNil)
			convey.So(csvFile, convey.ShouldNotBeEmpty)
			convey.So(csvBackFile, convey.ShouldNotBeEmpty)
		})
	})
}

func TestIsInPingListRange(t *testing.T) {
	convey.Convey("test PingManager method IsInPingListRange", t, func() {
		convey.Convey("01-should return false when get path failed", func() {
			m := &PingManager{}
			ret := m.IsInPingListRange()
			convey.So(ret, convey.ShouldBeFalse)
		})
		convey.Convey("02-should return false when path invalid timed out", func() {
			m := &PingManager{}
			patch := gomonkey.ApplyFuncReturn(slownet.GetPingListRangePath, "/", nil)
			defer patch.Reset()
			patchCheck := gomonkey.ApplyFuncReturn(slownet.CheckIsExistAndValid, errors.New("path not exist"))
			defer patchCheck.Reset()
			patchSleep := gomonkey.ApplyFunc(time.Sleep, func(duration time.Duration) {})
			defer patchSleep.Reset()
			ret := m.IsInPingListRange()
			convey.So(ret, convey.ShouldBeFalse)
		})
		convey.Convey("03-should return false when read file bytes failed", func() {
			m := &PingManager{}
			patch := gomonkey.ApplyFuncReturn(slownet.GetPingListRangePath, "/", nil)
			defer patch.Reset()
			patchCheck := gomonkey.ApplyFuncReturn(slownet.CheckIsExistAndValid, nil)
			defer patchCheck.Reset()
			patchRead := gomonkey.ApplyFuncReturn(utils.ReadLimitBytes, nil, errors.New("read failed"))
			defer patchRead.Reset()
			ret := m.IsInPingListRange()
			convey.So(ret, convey.ShouldBeFalse)
		})
		convey.Convey("04-should return false when unmarshal bytes failed", func() {
			m := &PingManager{}
			patch := gomonkey.ApplyFuncReturn(slownet.GetPingListRangePath, "/", nil)
			defer patch.Reset()
			patchCheck := gomonkey.ApplyFuncReturn(slownet.CheckIsExistAndValid, nil)
			defer patchCheck.Reset()
			patchRead := gomonkey.ApplyFuncReturn(utils.ReadLimitBytes, nil, nil)
			defer patchRead.Reset()
			ret := m.IsInPingListRange()
			convey.So(ret, convey.ShouldBeFalse)
		})
		testIsInPingListRangeWhenIdNotInRange()
	})
}

func testIsInPingListRangeWhenIdNotInRange() {
	convey.Convey("05-should return false when superPodId not in range", func() {
		m := &PingManager{superPodId: 1}
		patch := gomonkey.ApplyFuncReturn(slownet.GetPingListRangePath, "/", nil)
		defer patch.Reset()
		patchCheck := gomonkey.ApplyFuncReturn(slownet.CheckIsExistAndValid, nil)
		defer patchCheck.Reset()
		patchRead := gomonkey.ApplyFunc(utils.ReadLimitBytes, func(path string, limitLength int) ([]byte,
			error) {
			pingListRange := map[string][]string{"0": {"0", "1"}}
			return json.Marshal(pingListRange)
		})
		defer patchRead.Reset()
		ret := m.IsInPingListRange()
		convey.So(ret, convey.ShouldBeFalse)
	})
	convey.Convey("06-should return false when serverIndex not in range", func() {
		m := &PingManager{superPodId: 0, serverIndex: 1}
		patch := gomonkey.ApplyFuncReturn(slownet.GetPingListRangePath, "/", nil)
		defer patch.Reset()
		patchCheck := gomonkey.ApplyFuncReturn(slownet.CheckIsExistAndValid, nil)
		defer patchCheck.Reset()
		patchRead := gomonkey.ApplyFunc(utils.ReadLimitBytes, func(path string, limitLength int) ([]byte,
			error) {
			pingListRange := map[string][]string{"0": {"0", "2"}}
			return json.Marshal(pingListRange)
		})
		defer patchRead.Reset()
		ret := m.IsInPingListRange()
		convey.So(ret, convey.ShouldBeFalse)
	})
	convey.Convey("07-should return false when superPodID and serverIndex both in range", func() {
		m := &PingManager{superPodId: 0, serverIndex: 1}
		patch := gomonkey.ApplyFuncReturn(slownet.GetPingListRangePath, "/", nil)
		defer patch.Reset()
		patchCheck := gomonkey.ApplyFuncReturn(slownet.CheckIsExistAndValid, nil)
		defer patchCheck.Reset()
		patchRead := gomonkey.ApplyFunc(utils.ReadLimitBytes, func(path string, limitLength int) ([]byte,
			error) {
			pingListRange := map[string][]string{"0": {"0", "1"}}
			return json.Marshal(pingListRange)
		})
		defer patchRead.Reset()
		ret := m.IsInPingListRange()
		convey.So(ret, convey.ShouldBeTrue)
	})
}

func TestWriteToCsv(t *testing.T) {
	convey.Convey("test PingManager method writeToCsv", t, func() {
		m := &PingManager{superPodId: 0, serverIndex: 1}
		record := []string{"pingTaskId", "srcType", "srcAddr", "dstType", "dstAddr", "minDelay", "maxDelay", "avgDelay",
			"minLossRate", "maxLossRate", "avgLossRate", "timestamp"}
		csvFile := "res.csv"
		csvBakFile := "res.csv-bak"
		convey.Convey("01-should failed when file path invalid", func() {
			m.writeToCsv(record)
			_, err := os.Stat(csvFile)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("02-should success when prepare path success", func() {
			patch := gomonkey.ApplyPrivateMethod(m, "prepareResultFilePaths",
				func(appendMode bool) (string, string, error) {
					return csvFile, csvBakFile, nil
				})
			defer patch.Reset()
			m.writeToCsv(record)
			fileInfo, err := os.Stat(csvFile)
			convey.So(err, convey.ShouldBeNil)
			convey.So(fileInfo.Size(), convey.ShouldNotEqual, 0)
			err = os.Remove(csvFile)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestNewPingManager(t *testing.T) {
	testK8sClient := &kubeclient.ClientK8s{
		ClientSet:    fake.NewSimpleClientset(),
		NodeName:     "testNodeName",
		NodeInfoName: common.NodeInfoCMNamePrefix + "testNodeName",
	}
	convey.Convey("test func NewPingManager", t, func() {
		convey.Convey("01-should return nil when k8s client is nil", func() {
			m := NewPingManager(superPodID, rackID, serverIndex, nil, devType)
			convey.So(m, convey.ShouldBeNil)
		})
		convey.Convey("02-should return nil when k8s client get node failed", func() {
			m := NewPingManager(superPodID, rackID, serverIndex, testK8sClient, devType)
			convey.So(m, convey.ShouldBeNil)
		})
		convey.Convey("03-should return not nil when node acceleratorType label not exist", func() {
			patch := gomonkey.ApplyMethod(testK8sClient, "GetNodeWithCache",
				func(ck *kubeclient.ClientK8s) (*v1.Node, error) {
					return &v1.Node{}, nil
				})
			defer patch.Reset()
			m := NewPingManager(superPodID, rackID, serverIndex, testK8sClient, devType)
			convey.So(m, convey.ShouldNotBeNil)
		})
		convey.Convey("04-should return nil when node acceleratorType label invalid", func() {
			patch := gomonkey.ApplyMethod(testK8sClient, "GetNodeWithCache",
				func(ck *kubeclient.ClientK8s) (*v1.Node, error) {
					node := &v1.Node{}
					node.Labels = map[string]string{api.AcceleratorTypeKey: label800IA5x8}
					return node, nil
				})
			defer patch.Reset()
			m := NewPingManager(superPodID, rackID, serverIndex, testK8sClient, devType)
			convey.So(m, convey.ShouldBeNil)
		})
		convey.Convey("05-should return not nil when node acceleratorType label valid", func() {
			patch := gomonkey.ApplyMethod(testK8sClient, "GetNodeWithCache",
				func(ck *kubeclient.ClientK8s) (*v1.Node, error) {
					node := &v1.Node{}
					node.Labels = map[string]string{api.AcceleratorTypeKey: label900POD}
					return node, nil
				})
			defer patch.Reset()
			m := NewPingManager(superPodID, rackID, serverIndex, testK8sClient, devType)
			convey.So(m, convey.ShouldNotBeNil)
		})
	})
}

func TestPingManagerStart(t *testing.T) {
	convey.Convey("test PingManager method Start", t, func() {
		convey.Convey("01-should policy is nil when no cmd received", func() {
			m := &PingManager{
				wg:          &sync.WaitGroup{},
				executors:   make([]*IcmpPingExecutor, 0),
				commandChan: make(chan *types.HccspingMeshPolicy, 1),
				recordChan:  make(chan statisticData),
			}
			stopCh := make(chan struct{})
			go m.Start(stopCh)
			close(stopCh)
			m.wg.Wait()
			convey.So(m.executors, convey.ShouldBeEmpty)
		})
		convey.Convey("02-should policy is not nil when cmd received", func() {
			m := &PingManager{
				wg:          &sync.WaitGroup{},
				executors:   make([]*IcmpPingExecutor, 0),
				commandChan: make(chan *types.HccspingMeshPolicy, 1),
				recordChan:  make(chan statisticData),
			}
			patchCollect := gomonkey.ApplyPrivateMethod(m, "startCollect",
				func(stopCh chan struct{}) {
					defer m.wg.Done()
					return
				})
			defer patchCollect.Reset()
			stopCh := make(chan struct{})
			go m.Start(stopCh)
			policy := &types.HccspingMeshPolicy{
				Config: &types.HccspingMeshConfig{Activate: types.ActivateOn, TaskInterval: 1},
				DestAddrMap: map[string][]types.PingItem{
					"phyId1": {
						{SrcAddr: "127.0.0.1", DstAddr: "127.0.0.2"},
					},
				},
			}
			m.UpdateConfig(policy)
			time.Sleep(time.Millisecond * 1)
			close(stopCh)
			m.wg.Wait()
			if m.curPolicy == nil {
				convey.So(m.executors, convey.ShouldBeEmpty)
			} else {
				convey.So(m.executors, convey.ShouldNotBeEmpty)
			}
		})
	})
}

func TestCheckNodeLabelSupported(t *testing.T) {
	testK8sClient := &kubeclient.ClientK8s{
		ClientSet:    fake.NewSimpleClientset(),
		NodeName:     "testNodeName",
		NodeInfoName: common.NodeInfoCMNamePrefix + "testNodeName",
	}
	convey.Convey("test PingManager method CheckNodeLabelSupported", t, func() {
		convey.Convey("01-should return false when k8s client is nil", func() {
			m := &PingManager{}
			ret := m.CheckNodeLabelSupported()
			convey.So(ret, convey.ShouldBeFalse)
		})
		convey.Convey("02-should return false when get node info from k8s failed", func() {
			m := &PingManager{
				k8sClient: testK8sClient,
			}
			patch := gomonkey.ApplyMethod(testK8sClient, "GetNodeWithCache",
				func(ck *kubeclient.ClientK8s) (*v1.Node, error) {
					return nil, errors.New("get node info from k8s failed")
				})
			defer patch.Reset()
			ret := m.CheckNodeLabelSupported()
			convey.So(ret, convey.ShouldBeFalse)
		})
		testCheckNodeLabelSupportedWithLabelCase(testK8sClient)
	})
}

func testCheckNodeLabelSupportedWithLabelCase(testK8sClient *kubeclient.ClientK8s) {
	convey.Convey("03-should return false when accelerator-type not exist", func() {
		m := &PingManager{
			k8sClient: testK8sClient,
		}
		patch := gomonkey.ApplyMethod(testK8sClient, "GetNodeWithCache",
			func(ck *kubeclient.ClientK8s) (*v1.Node, error) {
				return &v1.Node{}, nil
			})
		defer patch.Reset()
		ret := m.CheckNodeLabelSupported()
		convey.So(ret, convey.ShouldBeFalse)
	})

	convey.Convey("04-should return false when accelerator-type not supported", func() {
		m := &PingManager{
			k8sClient: testK8sClient,
		}
		patch := gomonkey.ApplyMethod(testK8sClient, "GetNodeWithCache",
			func(ck *kubeclient.ClientK8s) (*v1.Node, error) {
				node := &v1.Node{}
				node.Labels = map[string]string{api.AcceleratorTypeKey: label800IA5x8}
				return node, nil
			})
		defer patch.Reset()
		ret := m.CheckNodeLabelSupported()
		convey.So(ret, convey.ShouldBeFalse)
	})

	convey.Convey("05-should return true when accelerator-type valid", func() {
		m := &PingManager{
			k8sClient: testK8sClient,
		}
		patch := gomonkey.ApplyMethod(testK8sClient, "GetNodeWithCache",
			func(ck *kubeclient.ClientK8s) (*v1.Node, error) {
				node := &v1.Node{}
				node.Labels = map[string]string{api.AcceleratorTypeKey: label900POD}
				return node, nil
			})
		defer patch.Reset()
		ret := m.CheckNodeLabelSupported()
		convey.So(ret, convey.ShouldBeTrue)
	})
}

func TestStartCollectCase01(t *testing.T) {
	convey.Convey("test PingManager method startCollect case 01", t, func() {
		convey.Convey("01-should return when stop channel is closed", func() {
			m := &PingManager{
				serverIndex: serverIndex,
				devType:     devType,
				wg:          &sync.WaitGroup{},
				executors:   make([]*IcmpPingExecutor, 0),
				commandChan: make(chan *types.HccspingMeshPolicy, 1),
				recordChan:  make(chan statisticData),
			}
			writeCsvCnt := 0
			patchCsv := gomonkey.ApplyPrivateMethod(m, "writeToCsv", func(record []string) {
				writeCsvCnt++
			})
			defer patchCsv.Reset()
			writeLogCnt := 0
			patchLog := gomonkey.ApplyPrivateMethod(m, "writeToLog", func(record []string) {
				writeLogCnt++
			})
			defer patchLog.Reset()
			m.wg.Add(1)
			stopCh := make(chan struct{})
			go m.startCollect(stopCh)
			close(stopCh)
			m.wg.Wait()
			convey.So(writeCsvCnt, convey.ShouldEqual, 0)
			convey.So(writeLogCnt, convey.ShouldEqual, 0)
		})
	})
}

func TestStartCollectCase02(t *testing.T) {
	convey.Convey("test PingManager method startCollect case 02", t, func() {
		convey.Convey("02-should write data when received data", func() {
			m := &PingManager{
				serverIndex: serverIndex,
				devType:     devType,
				wg:          &sync.WaitGroup{},
				executors:   make([]*IcmpPingExecutor, 0),
				commandChan: make(chan *types.HccspingMeshPolicy, 1),
				recordChan:  make(chan statisticData),
			}
			stopCh := make(chan struct{})
			writeCsvCnt := 0
			patchCsv := gomonkey.ApplyPrivateMethod(m, "writeToCsv", func(record []string) {
				writeCsvCnt++
			})
			defer patchCsv.Reset()
			writeLogCnt := 0
			patchLog := gomonkey.ApplyPrivateMethod(m, "writeToLog", func(record []string) {
				writeLogCnt++
			})
			defer patchLog.Reset()
			e := NewIcmpPingExecutor(stopCh, 0, NewOperator("a", "b", 1))
			m.executors = []*IcmpPingExecutor{e}
			data := statisticData{result: "json data", record: []string{"csv", "record", "data"}}
			patchResult := gomonkey.ApplyPrivateMethod(e, "getPingResultInfo",
				func(wg *sync.WaitGroup, sendCh chan statisticData) {
					defer m.wg.Done()
					m.recordChan <- data
				})
			defer patchResult.Reset()
			m.wg.Add(1)
			go m.startCollect(stopCh)
			time.Sleep(time.Second)
			close(stopCh)
			m.wg.Wait()
			convey.So(writeCsvCnt, convey.ShouldEqual, 1)
			convey.So(writeLogCnt, convey.ShouldEqual, 1)
		})
	})
}

func TestWriteToLog(t *testing.T) {
	convey.Convey("test writeToLog case 1", t, func() {
		m := &PingManager{
			writer: nil,
		}
		m.writeToLog("1")
	})
}

func TestGetCurPolicy(t *testing.T) {
	convey.Convey("test GetCurPolicy case 1", t, func() {
		curPolicy := &types.HccspingMeshPolicy{}
		m := &PingManager{
			curPolicy: curPolicy,
		}
		convey.So(m.GetCurPolicy(), convey.ShouldEqual, curPolicy)
	})
}

func TestSetFileWriter(t *testing.T) {
	convey.Convey("test SetFileWriter case 1", t, func() {
		writer := &hwlog.CustomLogger{}
		m := &PingManager{}
		m.SetFileWriter(writer)
		convey.So(writer, convey.ShouldEqual, m.writer)
	})
}

func TestGetDevType(t *testing.T) {
	convey.Convey("test GetDevType case 1", t, func() {
		devTypeTest := "1"
		m := &PingManager{
			devType: devTypeTest,
		}
		convey.So(m.GetDevType(), convey.ShouldEqual, devTypeTest)
	})
}
