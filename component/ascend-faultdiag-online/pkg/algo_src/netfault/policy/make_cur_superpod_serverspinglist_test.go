/* Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
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
// Package policy is used for processing superpod information
package policy

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/algo_src/netfault/algo"
	"ascend-faultdiag-online/pkg/algo_src/netfault/controllerflags"
)

// TestWriteServerIdPingList test for func writeServerIdPingList
func TestWriteServerIdPingList(t *testing.T) {
	var serverId = "1"
	defer os.Remove(serverId)
	convey.Convey("Test writeServerIdPingList", t, func() {
		convey.Convey("should return nil when osInfo is nil", func() {
			superPodPath := "1"
			patch := gomonkey.ApplyFunc(json.MarshalIndent, func(v any, prefix, indent string) ([]byte, error) {
				err := errors.New("error")
				return nil, err
			})
			defer patch.Reset()
			ret := writeServerIdPingList(nil, filepath.Join(superPodPath, serverId), superPodPath)
			convey.So(ret, convey.ShouldBeError)
		})

		convey.Convey("should return nil when create serverFilePath fail", func() {
			superPodPath := "1"
			ret := writeServerIdPingList(nil, filepath.Join(superPodPath, serverId), superPodPath)
			convey.So(ret, convey.ShouldBeError)
		})

		convey.Convey("should return nil when write serverFilePath success", func() {
			superPodPath := "."
			ret := writeServerIdPingList(nil, filepath.Join(superPodPath, serverId), superPodPath)
			convey.So(ret, convey.ShouldBeNil)
		})
	})
}

func TestHandlePingListPartOne(t *testing.T) {
	convey.Convey("Test handlePingList Part One", t, func() {
		convey.Convey("should return empty list when allPingList is empty", func() {
			var allPingList []any
			var expectReturnValue = make([]PingInfo, 0)
			srcIp := "1"
			key := "1"
			ret := handlePingList(allPingList, srcIp, key)
			convey.So(ret, convey.ShouldResemble, expectReturnValue)
		})

		convey.Convey("should return empty list when allPingList has one invalid value", func() {
			allPingList := []any{"1"}
			allPingList[0] = make([]any, 0)
			var expectReturnValue = make([]PingInfo, 0)
			srcIp := "1"
			key := "1"
			ret := handlePingList(allPingList, srcIp, key)
			convey.So(ret, convey.ShouldResemble, expectReturnValue)
		})

		convey.Convey("should return empty list when allPingList has no srcAddr value", func() {
			allPingList := []any{"1"}
			pingListItem := map[string]any{
				"1": "1",
			}
			allPingList[0] = pingListItem
			var expectReturnValue = make([]PingInfo, 0)
			srcIp := "1"
			key := "1"
			ret := handlePingList(allPingList, srcIp, key)
			hwlog.RunLog.Errorf("allPingList is: %v", allPingList)
			convey.So(ret, convey.ShouldResemble, expectReturnValue)
		})
	})
}

func TestHandlePingListPartTwo(t *testing.T) {
	convey.Convey("Test handlePingList Part Two", t, func() {
		convey.Convey("should return empty list when allPingList has one nil srcAddr value", func() {
			allPingList := []any{"1"}
			pingListItem := map[string]any{
				"srcAddr": nil,
			}
			allPingList[0] = pingListItem
			var expectReturnValue = make([]PingInfo, 0)
			srcIp := "1"
			key := "1"
			ret := handlePingList(allPingList, srcIp, key)
			hwlog.RunLog.Errorf("allPingList is: %v", allPingList)
			convey.So(ret, convey.ShouldResemble, expectReturnValue)
		})

		convey.Convey("should return empty list when allPingList has one srcAddr value, but no match with "+
			"srcIp", func() {
			allPingList := []any{"1"}
			pingListItem := map[string]any{
				"srcAddr": "2",
			}
			allPingList[0] = pingListItem
			var expectReturnValue = make([]PingInfo, 0)
			srcIp := "1"
			key := "1"
			ret := handlePingList(allPingList, srcIp, key)
			hwlog.RunLog.Errorf("allPingList is: %v", allPingList)
			convey.So(ret, convey.ShouldResemble, expectReturnValue)
		})

		convey.Convey("should return empty list when allPingList has one srcAddr value with match with srcIp, "+
			"not dstAddr", func() {
			allPingList := []any{"1"}
			pingListItem := map[string]any{
				"srcAddr": "1",
			}
			allPingList[0] = pingListItem
			var expectReturnValue = make([]PingInfo, 0)
			srcIp := "1"
			key := "1"
			ret := handlePingList(allPingList, srcIp, key)
			hwlog.RunLog.Errorf("allPingList is: %v", allPingList)
			convey.So(ret, convey.ShouldResemble, expectReturnValue)
		})
	})
}

func TestHandlePingListPartThree(t *testing.T) {
	convey.Convey("Test handlePingList Part Three", t, func() {
		convey.Convey("should return empty list when allPingList has one srcAddr value with match with srcIp, "+
			"invalid dstAddr", func() {
			allPingList := []any{"1"}
			pingListItem := map[string]any{"srcAddr": "1", "dstAddr": nil}
			allPingList[0] = pingListItem
			var expectReturnValue = make([]PingInfo, 0)
			srcIp := "1"
			key := "1"
			ret := handlePingList(allPingList, srcIp, key)
			hwlog.RunLog.Errorf("allPingList is: %v", allPingList)
			convey.So(ret, convey.ShouldResemble, expectReturnValue)
		})

		convey.Convey("should return empty list when allPingList has one srcAddr value with match with srcIp, "+
			"valid dstAddr, invalid key", func() {
			allPingList := []any{"1"}
			pingListItem := map[string]any{"srcAddr": "1", "dstAddr": "2"}
			allPingList[0] = pingListItem
			srcIp := "1"
			key := "abc"
			ret := handlePingList(allPingList, srcIp, key)
			hwlog.RunLog.Errorf("allPingList is: %v", allPingList)
			convey.So(ret, convey.ShouldBeNil)
		})

		convey.Convey("should return valid list", func() {
			var expectReturnValue = make([]PingInfo, 0)
			expectReturnValue = append(expectReturnValue,
				PingInfo{SrcIp: "1", DstIp: "2", SrcType: 0, DstType: 0, PktSize: 28, SrcCardPhyId: 2})
			allPingList := []any{"1"}
			pingListItem := map[string]any{
				"srcAddr": "1",
				"dstAddr": "2",
			}
			allPingList[0] = pingListItem
			srcIp := "1"
			key := "2"
			ret := handlePingList(allPingList, srcIp, key)
			hwlog.RunLog.Errorf("allPingList is: %v", allPingList)
			convey.So(ret, convey.ShouldResemble, expectReturnValue)
		})
	})
}

func ParseSuperPodInfoJson(jsonStr string) *SuperPodInfo {
	var config SuperPodInfo
	if err := json.Unmarshal([]byte(jsonStr), &config); err != nil {
		hwlog.RunLog.Errorf("parse Json err :%v", err)
		return nil
	}
	return &config
}

func ParsePingListJson(jsonStr string) map[string]any {
	var pingListInfo map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &pingListInfo); err != nil {
		hwlog.RunLog.Errorf("parse json err: %v", err)
		return nil
	}
	return pingListInfo
}

var pingListJson = `
    {
      "pingList": [
        {"srcType": 0,"dstType": 0,"pktSize": 28,"srcCardPhyId": 4,"srcAddr": "1","dstAddr": "2"},
        {"srcType": 0,"dstType": 0,"pktSize": 28,"srcCardPhyId": 4,"srcAddr": "2","dstAddr": "4"},
        {"srcType": 0,"dstType": 0,"pktSize": 28,"srcCardPhyId": 4,"srcAddr": "16","dstAddr": "14"},
        {"srcType": 0,"dstType": 0,"pktSize": 28,"srcCardPhyId": 4,"srcAddr": "17","dstAddr": "18"}
      ]
    }`
var npuJson = `{
       "Version": "A3",
       "SuperPodID": "0",
       "NodeDeviceMap": {
              "work1": {
					 "ServerID":"1",
                     "NodeName": "work1",
                     "DeviceMap": {"1": "1","2": "2"}
                      
              },
              "work2": {
                     "ServerID":"2",
                     "NodeName": "work2",
                     "DeviceMap": {"16": "16","17": "17","18": "18"}
              }
       }
}
`

func TestHandlePingList(t *testing.T) {
	var localhost = "127.0.0.0"
	var dstAddr = "127.0.0.1"
	convey.Convey("testingHanlePingList", t, func() {
		convey.Convey("should return [] when allPingList V not map[string]any", func() {
			allPingList := []any{
				"not map[string]any",
			}
			ret := handlePingList(allPingList, "path", "0")
			fmt.Println(ret)
			convey.So(ret, convey.ShouldResemble, []PingInfo{})
		})

		convey.Convey("should return nil when key atoi err and jump err format", func() {
			allPingList := []any{
				map[string]any{"srcaddr": 0},
				map[string]any{"srcAddr": 0},
				map[string]any{"srcAddr": localhost, "dstaddr": 0},
				map[string]any{"srcAddr": localhost, "dstAddr": 0},
				map[string]any{"srcAddr": localhost, "dstAddr": dstAddr},
			}
			ret := handlePingList(allPingList, localhost, "A")
			convey.So(ret, convey.ShouldBeNil)
		})

		convey.Convey("should return NewPingList when vaild", func() {
			allPingList := []any{
				map[string]any{"srcAddr": localhost, "dstAddr": dstAddr},
			}
			ret := handlePingList(allPingList, localhost, "0")
			fmt.Println(ret)
			convey.So(ret, convey.ShouldNotBeNil)
		})
	})
}

func TestGenSuperPodServersPingList(t *testing.T) {
	convey.Convey("Test GenSuperPodServersPingList when input invaild part I", t, func() {
		convey.Convey("should return getCrrentSuperPodInfo err", func() {
			mockGetCurrentSuperPod := gomonkey.ApplyFunc(getCurrentSuperPodInfo,
				func(_ string, _ *algo.NetDetect) (*SuperPodInfo, map[string]any) {
					return nil, nil
				})
			defer mockGetCurrentSuperPod.Reset()
			controllerflags.IsControllerExited.SetState(false)
			GenSuperPodServersPingList("", nil)
		})
	})
}

func TestSiftFromConfigMapA3PartOne(t *testing.T) {
	defer os.Remove("ping_list_1.json")
	defer os.Remove("ping_list_2.json")
	convey.Convey("test SiftFromConfigMapA3PartOne", t, func() {
		convey.Convey("should create correct file when input valid", func() {
			pingJson := ParsePingListJson(pingListJson)
			npuMap := ParseSuperPodInfoJson(npuJson)
			ret := siftFromConfigMapA3(npuMap, pingJson, "./")
			convey.So(ret, convey.ShouldBeTrue)
		})
		convey.Convey("should return false when NodeDeviceMap format error", func() {
			configMap := &SuperPodInfo{NodeDeviceMap: nil}
			ret := siftFromConfigMapA3(configMap, nil, "path")
			convey.So(ret, convey.ShouldBeFalse)
		})
		convey.Convey("should return false when workInfo format err", func() {
			configMap := &SuperPodInfo{
				NodeDeviceMap: map[string]*NodeDevice{"work": nil},
			}
			ret := siftFromConfigMapA3(configMap, nil, "path")
			convey.So(ret, convey.ShouldBeFalse)
		})

	})
}

func TestSiftFromConfigMapA3PartTwo(t *testing.T) {
	convey.Convey("test siftFromConfigMapA3 part two", t, func() {
		workNode := &NodeDevice{DeviceMap: nil}
		configMap := &SuperPodInfo{
			NodeDeviceMap: map[string]*NodeDevice{
				"work": workNode,
			},
		}
		convey.Convey("should return false when DeviceMap format err", func() {
			ret := siftFromConfigMapA3(configMap, nil, "path")
			convey.So(ret, convey.ShouldBeFalse)
		})

		workNodeWithServerID := &NodeDevice{
			DeviceMap: map[string]string{},
			ServerID:  "1",
		}
		configMapWithServerID := &SuperPodInfo{
			NodeDeviceMap: map[string]*NodeDevice{
				"work": workNodeWithServerID,
			},
		}
		convey.Convey("should return false when ServerID format err", func() {
			ret := siftFromConfigMapA3(configMapWithServerID, nil, "path")
			convey.So(ret, convey.ShouldBeFalse)
		})

		workNodeWithInvalidServerID := &NodeDevice{
			DeviceMap: map[string]string{},
			ServerID:  "err",
		}
		configMapWithInvalidServerID := &SuperPodInfo{
			NodeDeviceMap: map[string]*NodeDevice{
				"work": workNodeWithInvalidServerID,
			},
		}
		convey.Convey("should return false when ServerID Atoi err", func() {
			ret := siftFromConfigMapA3(configMapWithInvalidServerID, nil, "path")
			convey.So(ret, convey.ShouldBeFalse)
		})
	})
}

func TestSftFromConfigMapInterface(t *testing.T) {
	convey.Convey("test func siftFromConfigMapInterface", t, func() {
		convey.Convey("return false when version err ", func() {
			s := &SuperPodInfo{Version: "A"}
			ret := siftFromConfigMapInterface(s, nil, "")
			convey.So(ret, convey.ShouldBeFalse)
		})
		convey.Convey("call a3 when version ", func() {
			s := &SuperPodInfo{Version: DiagVersionA3}
			patch := gomonkey.ApplyFunc(siftFromConfigMapA3, func(_ *SuperPodInfo,
				_ map[string]any, _ string) bool {
				return true
			})
			defer patch.Reset()
			ret := siftFromConfigMapInterface(s, nil, "")
			convey.So(ret, convey.ShouldBeTrue)
		})

	})
}
