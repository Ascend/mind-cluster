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

// Package server contains unit tests for HwDevManager methods.
package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"Ascend-device-plugin/pkg/common"
	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	apiCommon "ascend-common/devmanager/common"
)

const (
	expectedFeId   = 1
	testPortLength = 2
)

var (
	eid = apiCommon.Eid{
		Raw: [apiCommon.EidByteSize]byte{
			0, 0, 0, 0, 0, 0, 0, expectedFeId << common.FeIdIndexBit,
			0, 0, 0, 0, 0, 0, 0, 0,
		},
	}
	urmaDevInfo = apiCommon.UrmaDeviceInfo{
		EidCount: 1,
		EidInfos: []apiCommon.UrmaEidInfo{
			{
				EidIndex: 1,
				Eid:      eid,
			},
		},
	}
)

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwlog.InitRunLogger(&hwLogConfig, context.Background())
}

// TestGetTopoFileInfoCheckFile test get topo file info (check file)
func TestGetTopoFileInfoCheckFile(t *testing.T) {
	convey.Convey("test HwDevManager method getTopoFileInfo check file", t, func() {
		convey.Convey("01-getTopoPath failed", func() {
			var invalidSuperPodType uint8 = math.MaxUint8
			p := ProductBase{superPodType: invalidSuperPodType}
			_, err := p.getTopoFileInfo()
			convey.So(err.Error(), convey.ShouldEqual,
				fmt.Sprintf("get topo path failed, err:<super pod type:<%d> topo path not exist>",
					invalidSuperPodType),
			)
		})
		convey.Convey("02-stat path failed", func() {
			mock1 := gomonkey.ApplyFunc(os.Stat, func(_ string) (os.FileInfo, error) {
				return nil, errors.New("fake error")
			})
			defer mock1.Reset()
			p := ProductBase{superPodType: 2}
			_, err := p.getTopoFileInfo()
			convey.So(err, convey.ShouldNotBeNil)
		})
		mock1 := gomonkey.ApplyFunc(os.Stat, func(_ string) (os.FileInfo, error) {
			return nil, nil
		})
		defer mock1.Reset()
		convey.Convey("03-read file failed", func() {
			mock2 := gomonkey.ApplyFunc(utils.ReadLimitBytes, func(_ string, _ int) ([]byte, error) {
				return nil, errors.New("fake error")
			})
			defer mock2.Reset()
			p := ProductBase{superPodType: 2}
			_, err := p.getTopoFileInfo()
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestGetTopoFileInfoCheckJson test get topo file info (check json)
func TestGetTopoFileInfoCheckJson(t *testing.T) {
	convey.Convey("test HwDevManager method getTopoFileInfo check json", t, func() {
		mock1 := gomonkey.ApplyFunc(os.Stat, func(_ string) (os.FileInfo, error) { return nil, nil })
		defer mock1.Reset()
		mock2 := gomonkey.ApplyFunc(utils.ReadLimitBytes, func(_ string, _ int) ([]byte, error) {
			return make([]byte, 0), nil
		})
		defer mock2.Reset()
		convey.Convey("04-json valid failed", func() {
			mock3 := gomonkey.ApplyFunc(json.Valid, func(_ []byte) bool { return false })
			defer mock3.Reset()
			p := ProductBase{superPodType: 2}
			_, err := p.getTopoFileInfo()
			convey.So(err, convey.ShouldNotBeNil)
		})
		mock3 := gomonkey.ApplyFunc(json.Valid, func(_ []byte) bool { return true })
		defer mock3.Reset()
		convey.Convey("05-json unmarshal failed", func() {
			mock4 := gomonkey.ApplyFunc(json.Unmarshal, func(_ []byte, _ any) error {
				return errors.New("fake error")
			})
			defer mock4.Reset()
			p := ProductBase{superPodType: 2}
			_, err := p.getTopoFileInfo()
			convey.So(err, convey.ShouldNotBeNil)
		})
		mock4 := gomonkey.ApplyFunc(json.Unmarshal, func(_ []byte, v any) error {
			return nil
		})
		defer mock4.Reset()
		convey.Convey("06-get topo info success", func() {
			p := ProductBase{superPodType: 2}
			topo, _ := p.getTopoFileInfo()
			convey.So(topo, convey.ShouldNotBeNil)
		})
	})
}

// TestGetEidPortMapKey test get eid port map key
func TestGetEidPortMapKey(t *testing.T) {
	convey.Convey("test method getEidPortMapKey", t, func() {
		convey.So(getEidPortMapKey(1, "0"), convey.ShouldEqual, "1_0")
	})
}

// TestGetSuffixAndCheckEid test get suffix and check eid
func TestGetSuffixAndCheckEid(t *testing.T) {
	convey.Convey("test method getSuffixAndCheckEid", t, func() {
		convey.Convey("01-rLevel is 2, return 0", func() {
			x, _ := getSuffixAndCheckEid("", api.RankLevel2)
			convey.So(x, convey.ShouldEqual, 0)
		})
		convey.Convey("02-eid len is 0, return err", func() {
			_, err := getSuffixAndCheckEid("1", 1)
			convey.So(err.Error(), convey.ShouldEqual, "get port list by eid error, eid:<1> len is invalid")
		})
		convey.Convey("03-eid parse int error, return err", func() {
			_, err := getSuffixAndCheckEid("1xx", 1)
			convey.So(err.Error(), convey.ShouldEqual, "get port list by eid error, eid:<1xx> is invalid")
		})
		convey.Convey("04-x value is 0, return err", func() {
			_, err := getSuffixAndCheckEid("100", 1)
			convey.So(err.Error(), convey.ShouldEqual, "get port list by eid error, eid:<100> is invalid")
		})
		convey.Convey("05-x value valid ok, return x", func() {
			x, _ := getSuffixAndCheckEid("1b6", 1)
			convey.So(x, convey.ShouldEqual, common.LogicLowerLimit+1)
		})
	})
}

// TestGetPortListByEid test get port list by eid
func TestGetPortListByEid(t *testing.T) {
	convey.Convey("test method getPortListByEid", t, func() {
		convey.Convey("01-real card not a5, return err", func() {
			oldRealType := common.ParamOption.RealCardType
			common.ParamOption.RealCardType = api.Ascend910
			defer func() {
				common.ParamOption.RealCardType = oldRealType
			}()
			npu := NewNpuBase()
			npu.productInfo = &ProductBase{superPodType: 2}
			_, err := npu.GetPortListByEid(1, "1", 1)
			convey.So(err.Error(), convey.ShouldEqual, "get port list by eid error, device type is not A5")
		})
		oldRealType := common.ParamOption.RealCardType
		common.ParamOption.RealCardType = api.Ascend910A5
		defer func() {
			common.ParamOption.RealCardType = oldRealType
		}()
		convey.Convey("02-get suffix err, return err", func() {
			npu := NewNpuBase()
			npu.productInfo = &ProductBase{superPodType: 2}
			_, err := npu.GetPortListByEid(1, "1", 1)
			convey.So(err.Error(), convey.ShouldEqual, "get port list by eid error, eid:<1> len is invalid")
		})
		convey.Convey("03-hit cache, return ports", func() {
			npu := NewNpuBase()
			npu.eidPortMap["1_1b6"] = []string{"1", "2"}
			npu.productInfo = &ProductBase{superPodType: 2, topoInfo: &TopoInfo{}}
			ports, err := npu.GetPortListByEid(1, "1b6", 1)
			convey.So(err, convey.ShouldBeNil)
			convey.So(len(ports), convey.ShouldEqual, testPortLength)
		})

		convey.Convey("04-get port list, return ports", func() {
			npu := NewNpuBase()
			npu.productInfo = &ProductBase{superPodType: 2}
			mock4 := gomonkey.ApplyPrivateMethod(npu, "getPortsList", func(_ *NpuBase, _ int32, _ string,
				_ int, _ int8) ([]string, error) {
				return []string{"1", "2"}, nil
			})
			defer mock4.Reset()
			ports, err := npu.GetPortListByEid(1, "1b6", 1)
			convey.So(err, convey.ShouldBeNil)
			convey.So(len(ports), convey.ShouldEqual, testPortLength)
		})
	})
}

// TestNpuBaseMethodCreateRankAddrItem test base method
func TestNpuBaseMethodCreateRankAddrItem(t *testing.T) {
	convey.Convey("Test NpuBase method createRankAddrItem", t, func() {
		npu := NewNpuBase()
		eid := apiCommon.Eid{
			Raw: [apiCommon.EidByteSize]byte{
				0, 1, 0, 1, 0, 1, 0, 1,
				0, 1, 0, 1, 0, 1, 0, 1,
			},
		}
		convey.Convey("01-should return empty when netType is invalid", func() {
			actual := npu.createRankAddrItem("xxx", eid, []string{})
			convey.So(actual.AddrType, convey.ShouldBeEmpty)
		})

		convey.Convey("01-should start with EID_ when netType is UB or UBG", func() {
			actual := npu.createRankAddrItem(api.LevelInfoTypeUB, eid, []string{})
			convey.So(actual.AddrType, convey.ShouldEqual, "EID")
		})

		convey.Convey("02-should start with IP_ when netType is UBOE", func() {
			actual := npu.createRankAddrItem(api.LevelInfoTypeUBoE, eid, []string{})
			convey.So(actual.AddrType, convey.ShouldEqual, "IPV4")
		})
	})
}

// TestNpuBaseMethodGetNetTypeForLevel test base method
func TestNpuBaseMethodGetNetTypeForLevel(t *testing.T) {
	convey.Convey("Test NpuBase method getNetTypeForLevel", t, func() {
		npu := NewNpuBase()
		convey.Convey("01-should return topo when level is 0", func() {
			actual := npu.getNetTypeForLevel(api.RankLevel0)
			convey.So(actual, convey.ShouldEqual, api.NetTypeTopo)
		})

		convey.Convey("02-should return topo when level is not 0", func() {
			actual := npu.getNetTypeForLevel(api.RankLevel1)
			convey.So(actual, convey.ShouldEqual, api.NetTypeCLOS)
		})
	})
}

// TestNpuBaseMethodGetFeIDByEid test base method
func TestNpuBaseMethodGetFeIDByEid(t *testing.T) {
	convey.Convey("Test NpuBase method getFeIDByEid", t, func() {
		npu := NewNpuBase()
		convey.Convey("01-should return max uint when eid is nil", func() {
			actual := npu.getFeIDByEid(nil)
			convey.So(actual, convey.ShouldEqual, uint(math.MaxUint))
		})

		convey.Convey("02-should return feID when eid is valid", func() {
			actual := npu.getFeIDByEid(&eid)
			convey.So(actual, convey.ShouldEqual, expectedFeId)
		})
	})
}

// TestNpuBaseMethodGetEidListByFeID test base method
func TestNpuBaseMethodGetEidListByFeID(t *testing.T) {
	convey.Convey("Test NpuBase method getEidListByFeID", t, func() {
		npu := NewNpuBase()
		convey.Convey("01-should return empty when urma dev info is nil", func() {
			actual := npu.getEidListByFeID(expectedFeId, nil)
			convey.So(actual, convey.ShouldBeEmpty)
		})
		convey.Convey("02-should return empty when urma dev info is valid with invalid feId ", func() {
			actual := npu.getEidListByFeID(expectedFeId+1, &urmaDevInfo)
			convey.So(actual, convey.ShouldBeEmpty)
		})

		convey.Convey("03-should return eid when urma dev info is valid with valid feId", func() {
			actual := npu.getEidListByFeID(expectedFeId, &urmaDevInfo)
			convey.So(len(actual), convey.ShouldEqual, 1)
		})
	})
}

// TestNpuBaseMethodGetRankLevelInfoKeyArr test base method
func TestNpuBaseMethodGetRankLevelInfoKeyArr(t *testing.T) {
	convey.Convey("Test NpuBase method getRankLevelInfoKeyArr", t, func() {
		npu := NewNpuBase()
		convey.Convey("01-should return empty when product info is nil", func() {
			npu.productInfo = nil
			actual := npu.getRankLevelInfoKeyArr()
			convey.So(actual, convey.ShouldBeEmpty)
		})
		convey.Convey("02-should return level 2 ubg when product is 2d", func() {
			npu.productInfo = &ProductBase{
				superPodType: common.ProductType2D}
			actual := npu.getRankLevelInfoKeyArr()
			convey.So(len(actual), convey.ShouldEqual, api.RankLevelCnt)
			convey.So(actual[api.RankLevel2], convey.ShouldEqual, api.LevelInfoTypeUBG)
		})
		convey.Convey("03-should return level 2 uboe when product is server", func() {
			npu.productInfo = &ProductBase{
				superPodSize: 0,
				superPodID:   0,
				superPodType: common.ProductTypeServer}
			actual := npu.getRankLevelInfoKeyArr()
			convey.So(len(actual), convey.ShouldEqual, api.RankLevelCnt)
			convey.So(actual[api.RankLevel2], convey.ShouldEqual, api.LevelInfoTypeUBoE)
		})
		convey.Convey("04-should return level 1 empty when product is server without superPodID", func() {
			npu.productInfo = &ProductBase{
				superPodSize: common.InvalidSuperPodSize,
				superPodID:   common.InvalidSuperPodID,
				superPodType: common.ProductTypeServer}
			actual := npu.getRankLevelInfoKeyArr()
			convey.So(len(actual), convey.ShouldEqual, api.RankLevelCnt)
			convey.So(actual[api.RankLevel1], convey.ShouldEqual, api.LevelInfoTypeIgnore)
			convey.So(actual[api.RankLevel2], convey.ShouldEqual, api.LevelInfoTypeUBoE)
		})
		convey.Convey("05-should return level 3 when product is standard card 1p", func() {
			npu.productInfo = &ProductBase{cardType: common.A5300ICardName}
			actual := npu.getRankLevelInfoKeyArr()
			convey.So(len(actual), convey.ShouldEqual, api.RankLevelCnt)
			convey.So(actual[api.RankLevel0], convey.ShouldEqual, api.LevelInfoTypeIgnore)
		})
		convey.Convey("06-should return level 0,3 when product is standard card 4p", func() {
			npu.productInfo = &ProductBase{cardType: common.A54P300ICardName}
			actual := npu.getRankLevelInfoKeyArr()
			convey.So(len(actual), convey.ShouldEqual, api.RankLevelCnt)
			convey.So(actual[api.RankLevel0], convey.ShouldEqual, api.LevelInfoTypeUB)
			convey.So(actual[api.RankLevel1], convey.ShouldEqual, api.LevelInfoTypeIgnore)
			convey.So(actual[api.RankLevel2], convey.ShouldEqual, api.LevelInfoTypeIgnore)
		})
	})
}

// TestNpuBaseMethodGetNetTypeAndFeIDListByRankLevel test base method
func TestNpuBaseMethodGetNetTypeAndFeIDListByRankLevel(t *testing.T) {
	convey.Convey("Test NpuBase method getNetTypeAndFeIDListByRankLevel", t, func() {
		npu := NewNpuBase()
		convey.Convey("01-should return empty info when product info is nil", func() {
			npu.productInfo = nil
			netType, feIdList := npu.getNetTypeAndFeIDListByRankLevel(api.RankLevel0)
			convey.So(netType, convey.ShouldBeEmpty)
			convey.So(feIdList, convey.ShouldBeEmpty)
		})
		convey.Convey("02-should return empty info when rank level is invalid", func() {
			npu.productInfo = &ProductBase{
				superPodSize: 0,
				superPodID:   0,
				superPodType: common.ProductTypeServer,
			}
			netType, feIdList := npu.getNetTypeAndFeIDListByRankLevel(api.RankLevel0 - 1)
			convey.So(netType, convey.ShouldBeEmpty)
			convey.So(feIdList, convey.ShouldBeEmpty)
		})
		convey.Convey("03-should return level 2 ubg when product is 2d", func() {
			npu.productInfo = &ProductBase{
				superPodSize: 0,
				superPodID:   0,
				superPodType: common.ProductType2D,
			}
			netType, feIdList := npu.getNetTypeAndFeIDListByRankLevel(api.RankLevel2)
			convey.So(netType, convey.ShouldEqual, api.LevelInfoTypeUBG)
			convey.So(feIdList, convey.ShouldResemble, []uint{common.UrmaFeId3})
		})
		convey.Convey("04-should return level 2 uboe when product is server", func() {
			npu.productInfo = &ProductBase{
				superPodSize: 0,
				superPodID:   0,
				superPodType: common.ProductTypeServer,
			}
			netType, feIdList := npu.getNetTypeAndFeIDListByRankLevel(api.RankLevel2)
			convey.So(netType, convey.ShouldEqual, api.LevelInfoTypeUBoE)
			convey.So(feIdList, convey.ShouldResemble, []uint{common.UrmaFeId8, common.UrmaFeId9})
		})
	})
}

// TestNpuBaseMethodGetRandAddrByFuncEntityID test base method
func TestNpuBaseMethodGetRandAddrByFuncEntityID(t *testing.T) {
	convey.Convey("Test NpuBase method getRandAddrByFuncEntityID", t, func() {
		npu := NewNpuBase()
		convey.Convey("01-should return nil when urmaDevInfoMap is empty", func() {
			npu.urmaDevInfoMap = nil
			actual := npu.getRandAddrByFuncEntityID(0, 0, "", 0)
			convey.So(actual, convey.ShouldBeEmpty)
		})
		convey.Convey("02-should return nil when GetPortListByEid failed", func() {
			patch := gomonkey.ApplyMethodReturn(npu, "GetPortListByEid", nil,
				errors.New("get port list by eid failed"))
			defer patch.Reset()
			const thePhyID0 = 0
			npu.urmaDevInfoMap[thePhyID0] = []apiCommon.UrmaDeviceInfo{urmaDevInfo}
			actual := npu.getRandAddrByFuncEntityID(thePhyID0, expectedFeId, api.LevelInfoTypeUBG, api.RankLevel2)
			convey.So(actual, convey.ShouldBeEmpty)
		})
		convey.Convey("03-should return rank addr list when all is valid", func() {
			patch := gomonkey.ApplyMethodReturn(npu, "GetPortListByEid", []string{"0/1", "1/0"}, nil)
			defer patch.Reset()
			const thePhyID0 = 0
			npu.urmaDevInfoMap[thePhyID0] = []apiCommon.UrmaDeviceInfo{urmaDevInfo}
			actual := npu.getRandAddrByFuncEntityID(thePhyID0, expectedFeId, api.LevelInfoTypeUBG, api.RankLevel2)
			convey.So(actual, convey.ShouldNotBeEmpty)
		})
	})
}

// TestProductBaseMethodIsServer test base method
func TestProductBaseMethodIsServer(t *testing.T) {
	convey.Convey("Test ProductBase method isServer", t, func() {
		convey.Convey("01-should return false when product is empty", func() {
			var product *ProductBase
			actual := product.isServer()
			convey.So(actual, convey.ShouldBeFalse)
		})

		convey.Convey("02-should return false when product is 2d", func() {
			product := &ProductBase{
				superPodSize: 0,
				superPodID:   0,
				chassisID:    0,
				superPodType: common.ProductType2D,
			}
			actual := product.isServer()
			convey.So(actual, convey.ShouldBeFalse)
		})

		convey.Convey("03-should return true when product is server", func() {
			product := &ProductBase{
				superPodSize: 0,
				superPodID:   0,
				chassisID:    0,
				superPodType: common.ProductTypeServer,
			}
			actual := product.isServer()
			convey.So(actual, convey.ShouldBeTrue)
		})
	})
}

// TestProductBaseMethodIsSuperServer test base method
func TestProductBaseMethodIsSuperServer(t *testing.T) {
	convey.Convey("Test ProductBase method isSuperServer", t, func() {
		convey.Convey("01-should return false when product is empty", func() {
			var product *ProductBase
			actual := product.isSuperServer()
			convey.So(actual, convey.ShouldBeFalse)
		})

		convey.Convey("02-should return false when product is 2d", func() {
			product := &ProductBase{
				superPodSize: 0,
				superPodID:   0,
				chassisID:    0,
				superPodType: common.ProductType2D,
			}
			actual := product.isSuperServer()
			convey.So(actual, convey.ShouldBeFalse)
		})
		convey.Convey("03-should return true when product is server with super pod", func() {
			product := &ProductBase{
				superPodSize: common.InvalidSuperPodSize,
				superPodID:   common.InvalidSuperPodID,
				superPodType: common.ProductTypeServer,
			}
			actual := product.isSuperServer()
			convey.So(actual, convey.ShouldBeFalse)
		})

		convey.Convey("04-should return true when product is super server", func() {
			product := &ProductBase{
				superPodSize: 0,
				superPodID:   0,
				chassisID:    0,
				superPodType: common.ProductTypeServer,
			}
			actual := product.isSuperServer()
			convey.So(actual, convey.ShouldBeTrue)
		})
	})
}

// TestProductBaseMethodIsPodScene test base method
func TestProductBaseMethodIsPodScene(t *testing.T) {
	convey.Convey("Test ProductBase method isPodScene", t, func() {
		convey.Convey("01-should return false when product is empty", func() {
			var product *ProductBase
			actual := product.isPodScene()
			convey.So(actual, convey.ShouldBeFalse)
		})

		convey.Convey("02-should return false when product is 2d", func() {
			product := &ProductBase{
				superPodSize: 0,
				superPodID:   0,
				chassisID:    0,
				superPodType: common.ProductType2D,
			}
			actual := product.isPodScene()
			convey.So(actual, convey.ShouldBeTrue)
		})
		convey.Convey("03-should return false when product is server with super pod", func() {
			product := &ProductBase{
				superPodSize: common.InvalidSuperPodSize,
				superPodID:   common.InvalidSuperPodID,
				superPodType: common.ProductTypeServer,
			}
			actual := product.isPodScene()
			convey.So(actual, convey.ShouldBeFalse)
		})

		convey.Convey("04-should return false when product is super server", func() {
			product := &ProductBase{
				superPodSize: 0,
				superPodID:   0,
				chassisID:    0,
				superPodType: common.ProductTypeServer,
			}
			actual := product.isPodScene()
			convey.So(actual, convey.ShouldBeFalse)
		})
	})
}

type getIdCaseData struct {
	level        int
	superPodSize int
	superPodType int
	superPodID   int
	rackID       int
	serverIndex  int
	serverIP     string
	expected     string
}

// TestGetPortsList test get ports list
func TestGetPortsList(t *testing.T) {
	convey.Convey("Test NpuBase method getPortsList", t, func() {
		npu := NewNpuBase()
		npu.productInfo = &ProductBase{superPodType: 2}
		convey.Convey("01-getPortsList level 2 read topo info err", func() {
			mock1 := gomonkey.ApplyPrivateMethod(npu.productInfo,
				"getTopoFileInfo", func(_ *ProductBase) (*TopoInfo, error) {
					return nil, errors.New("fake error")
				})
			defer mock1.Reset()
			_, err := npu.getPortsList(1, "1", api.RankLevel2)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("02-getPortsList level 2 read topo info success", func() {
			mock1 := gomonkey.ApplyPrivateMethod(npu.productInfo,
				"getTopoFileInfo", func(_ *ProductBase) (*TopoInfo, error) {
					topoInfo := &TopoInfo{EdgeList: []Edge{
						{LocalA: 1, LocalAPorts: []string{"123", "345"}, NetLayer: 2},
					}}
					return topoInfo, nil
				})
			defer mock1.Reset()
			ports, _ := npu.getPortsList(1, "1", api.RankLevel2)
			convey.So(len(ports), convey.ShouldEqual, testPortLength)
		})
		convey.Convey("03-getPortsList level 1 full mesh", func() {
			ports, _ := npu.getPortsList(1, "18c", 1)
			convey.So(ports[0], convey.ShouldEqual, "1/4")
		})
		convey.Convey("04-getPortsList level 1 read topo info", func() {
			mock1 := gomonkey.ApplyPrivateMethod(npu.productInfo,
				"getTopoFileInfo", func(_ *ProductBase) (*TopoInfo, error) {
					topoInfo := &TopoInfo{EdgeList: []Edge{
						{LocalA: 1, LocalAPorts: []string{"0/1", "0/2"}, NetLayer: 1, LinkType: "PEER2NET"},
					}}
					return topoInfo, nil
				})
			defer mock1.Reset()
			ports, _ := npu.getPortsList(1, "1b6", 1)
			convey.So(ports[0], convey.ShouldEqual, "0/1")
		})
	})
}
