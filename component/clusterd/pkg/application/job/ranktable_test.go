// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package job a series of job function
package job

import (
	"encoding/json"
	"fmt"
	"strconv"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"clusterd/pkg/common/constant"
)

const (
	mockRankTableStatusInit    = `{"Status":"initializing"}`
	mockRankTableStatusInvalid = `{"Status":"invalid"}`
	mockDefaultDeviceIP        = "127.0.0.1"
)

// TestCheckDeviceInfo test CheckDeviceInfo
func TestCheckDeviceInfo(t *testing.T) {
	convey.Convey("test CheckDeviceInfo", t, func() {
		instance := mockInstance()
		convey.Convey("serverID parse failed", func() {
			boolCheck := CheckDeviceInfo(instance)
			convey.So(boolCheck, convey.ShouldBeFalse)
		})
		instance.ServerID = mockDefaultDeviceIP
		convey.Convey("device num is zero", func() {
			boolCheck := CheckDeviceInfo(instance)
			convey.So(boolCheck, convey.ShouldBeFalse)
		})
		instance.Devices = mockDevice()
		convey.Convey("deviceID convert failed", func() {
			patch := gomonkey.ApplyFunc(strconv.Atoi, func(s string) (int, error) {
				return 0, fmt.Errorf("string converted to type int failed")
			})
			defer patch.Reset()
			boolCheck := CheckDeviceInfo(instance)
			convey.So(boolCheck, convey.ShouldBeFalse)
		})
		convey.Convey("deviceIP parse failed", func() {
			instance.Devices[0].DeviceIP = ""
			boolCheck := CheckDeviceInfo(instance)
			convey.So(boolCheck, convey.ShouldBeFalse)
		})
	})
}

// TestUnmarshalToRankTable test UnmarshalToRankTable
func TestUnmarshalToRankTable(t *testing.T) {
	convey.Convey("test UnmarshalToRankTable", t, func() {
		rankTableStatus := &RankTableStatus{Status: ""}
		convey.Convey("status is invalid", func() {
			err := rankTableStatus.UnmarshalToRankTable(mockRankTableStatusInvalid)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("status is initializing", func() {
			err := rankTableStatus.UnmarshalToRankTable(mockRankTableStatusInit)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("json length is invalid", func() {
			jsonBytes := make([]byte, cmDataMaxMemory+1)
			err := rankTableStatus.UnmarshalToRankTable(string(jsonBytes))
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("json formatting is invalid", func() {
			jsonBytes := ""
			err := rankTableStatus.UnmarshalToRankTable(jsonBytes)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestCachePodInfo test CachePodInfo
func TestCachePodInfo(t *testing.T) {
	convey.Convey("test CachePodInfo", t, func() {
		rankTable := mockRankTableInit()
		pod := mockPod()
		instance := mockInstance()
		var rankIndex *int
		rankIndex = new(int)
		*rankIndex = 1
		convey.Convey("check device info failed", func() {
			err := rankTable.CachePodInfo(pod, *instance, rankIndex)
			convey.So(err, convey.ShouldNotBeNil)
		})

		mockCheckDeviceInfo := gomonkey.ApplyFunc(CheckDeviceInfo, func(_ *Instance) bool {
			return true
		})
		defer mockCheckDeviceInfo.Reset()
		convey.Convey("when pod is already cached", func() {
			server1 := &ServerHccl{
				ServerID:   "192.168.1.1",
				ServerName: "Server1",
				DeviceList: []*Device{
					{DeviceID: "0", DeviceIP: "192.168.1.10", RankID: "0"},
				},
			}
			server2 := &ServerHccl{
				ServerID:   "192.168.1.2",
				ServerName: "Server2",
				DeviceList: []*Device{
					{DeviceID: "1", DeviceIP: "192.168.1.11", RankID: "1"},
				},
			}
			rankTable.ServerList = []*ServerHccl{server1, server2}
			err := rankTable.CachePodInfo(pod, *instance, rankIndex)
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("with rankFactor more than A800MaxChipNum", func() {
			instance.Devices = make([]Device, A800MaxChipNum+1)
			err := rankTable.CachePodInfo(pod, *instance, rankIndex)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("with device list undefined", func() {
			err := rankTable.CachePodInfo(pod, *instance, rankIndex)
			convey.So(err, convey.ShouldNotBeNil)
		})

		instance.Devices = mockDevice()
		convey.Convey("pod cached succeed", func() {
			err := rankTable.CachePodInfo(pod, *instance, rankIndex)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestGetJobHealthy(t *testing.T) {
	convey.Convey("test GetJobHealthy", t, func() {
		rankTable := mockRankTableInit()
		rankTable.UnHealthyNode = make(map[string][]string)
		rankTable.UnHealthyDevice = make(map[string]string)
		convey.Convey("node and device is healthy", func() {
			healthy, ranks := rankTable.GetJobHealthy()
			convey.So(healthy, convey.ShouldEqual, true)
			convey.So(len(ranks), convey.ShouldEqual, 0)
		})
		rankTable.UnHealthyNode["node1"] = []string{"rank1"}
		convey.Convey("node is not healthy and device is healthy", func() {
			healthy, ranks := rankTable.GetJobHealthy()
			convey.So(healthy, convey.ShouldEqual, false)
			convey.So(len(ranks), convey.ShouldEqual, 1)
		})
		rankTable.UnHealthyDevice["device1"] = "rank1"
		convey.Convey("node is not healthy and device is not healthy", func() {
			healthy, ranks := rankTable.GetJobHealthy()
			convey.So(healthy, convey.ShouldEqual, false)
			convey.So(len(ranks), convey.ShouldEqual, 1)
		})
	})
}

func TestSetJobNodeHealthy(t *testing.T) {
	convey.Convey("test SetJobNodeHealthy", t, func() {
		rankTable := mockRankTableInit()
		rankTable.UnHealthyNode = make(map[string][]string)
		device1 := &Device{
			DeviceID: "device_id_0",
			DeviceIP: "device_ip_0",
			RankID:   "rank_id_0",
		}
		device2 := &Device{
			DeviceID: "device_id_1",
			DeviceIP: "device_ip_1",
			RankID:   "rank_id_1",
		}
		rankTable.ServerList = []*ServerHccl{{
			DeviceList: []*Device{device1, device2},
			ServerID:   "",
			PodID:      "",
			ServerName: "node0",
				}}
		convey.Convey("set node status case node not exist", func() {
			rankTable.SetJobNodeHealthy("node1", true)
			convey.So(len(rankTable.UnHealthyNode), convey.ShouldEqual, 0)
		})
		convey.Convey("set node status case node exist with unhealthy status", func() {
			rankTable.SetJobNodeHealthy("node0", false)
			convey.So(len(rankTable.UnHealthyNode), convey.ShouldEqual, 1)
			convey.So(len(rankTable.UnHealthyNode["node0"]), convey.ShouldEqual, 2)
		})
		convey.Convey("set node status case node exist with healthy status", func() {
			rankTable.SetJobNodeHealthy("node0", true)
			convey.So(len(rankTable.UnHealthyNode), convey.ShouldEqual, 0)
		})
	})
}

func TestSetJobDeviceHealthy(t *testing.T) {
	convey.Convey("test SetJobDeviceHealthy", t, func() {
		rankTable := mockRankTableInit()
		rankTable.UnHealthyDevice = make(map[string]string)
		rankTable.ServerList = []*ServerHccl{{
			DeviceList: []*Device{{
				DeviceID: "0",
				DeviceIP: "",
				RankID:   "",
			}},
			ServerID:   "",
			PodID:      "",
			ServerName: "node0",
				}}
		convey.Convey("set device status case node not exist", func() {
			rankTable.SetJobDeviceHealthy("node1", constant.AscendDevPrefix+"0", "")
			convey.So(len(rankTable.UnHealthyDevice), convey.ShouldEqual, 0)
		})
		convey.Convey("set device status case node exist with networkUnhealthy device", func() {
			rankTable.SetJobDeviceHealthy("node0", constant.AscendDevPrefix+"0", "")
			convey.So(len(rankTable.UnHealthyDevice), convey.ShouldEqual, 1)
		})
		convey.Convey("set device status case node exist with unhealthy device", func() {
			rankTable.SetJobDeviceHealthy("node0", "", constant.AscendDevPrefix+"0")
			convey.So(len(rankTable.UnHealthyDevice), convey.ShouldEqual, 1)
		})
		convey.Convey("set device status case node neither exist with networkUnhealthy device nor unhealthy device", func() {
			rankTable.UnHealthyDevice["node0:"+constant.AscendDevPrefix+"0"] = ""
			rankTable.SetJobDeviceHealthy("node0", "", "")
			convey.So(len(rankTable.UnHealthyDevice), convey.ShouldEqual, 1)
		})
	})
}

func mockRankTableInit() *RankTable {
	return &RankTable{
		RankTableStatus: RankTableStatus{Status: ConfigmapInitializing},
		ServerList:      []*ServerHccl{},
		ServerCount:     "",
		Version:         "",
	}
}

func mockRankTableWithLength1() *RankTable {
	return &RankTable{
		ServerList: []*ServerHccl{
			{
				ServerID:   "server1",
				ServerName: "Server One",
				DeviceList: []*Device{
					{DeviceID: "device1", DeviceIP: "192.168.0.1", RankID: "0"},
				},
			},
		},
	}
}

func mockInstance() *Instance {
	return &Instance{
		Devices:  make([]Device, 0),
		PodName:  "",
		ServerID: "",
	}
}

func mockDevice() []Device {
	device := Device{
		DeviceID: "1",
		DeviceIP: "192.168.30.31",
		RankID:   "0",
	}
	return []Device{device}
}

type testCase struct {
	description  string
	rankTable    RankTable
	expectedJson []string
}

func mockRankTableWithParam() RankTable {
	return RankTable{
		RankTableStatus: RankTableStatus{Status: "ok"},
		ServerList: []*ServerHccl{
			{
				ServerID:   "server1",
				ServerName: "Server One",
				DeviceList: []*Device{
					{DeviceID: "device1", DeviceIP: "192.168.0.1", RankID: "0"},
				},
			},
		},
		ServerCount: "1",
		Version:     "1.0",
		Total:       1,
	}
}

func mockTestCases() []testCase {
	testCases := []testCase{
		{
			description:  "serverList length is 0",
			rankTable:    RankTable{},
			expectedJson: nil,
		},
		{
			description: "server count less than threshold, returns json slice with single element",
			rankTable:   mockRankTableWithParam(),
			expectedJson: []string{`{"status":"ok","server_list":[{"device":[{"device_id":"device1","device_ip"` +
				`:"192.168.0.1","rank_id":"0"}],"server_id":"server1","server_name":"Server One"}],` +
				`"server_count":"1","version":"1.0","total":1}`},
		},
	}
	return testCases
}

// TestRankTableGetHccLJsonSlice test case for RankTableGetHccLJsonSlice
func TestRankTableGetHccLJsonSlice(t *testing.T) {
	testCases := mockTestCases()
	convey.Convey("Testing GetHccLJsonSlice", t, func() {
		for _, testCase := range testCases {
			convey.Convey(testCase.description, func() {
				result := testCase.rankTable.GetHccLJsonSlice()
				fmt.Printf("TestRankTableGetHccLJsonSlice result=%v", result)
				convey.So(result, convey.ShouldResemble, testCase.expectedJson)
			})
		}
		convey.Convey("server count less than threshold, returns error", func() {
			patch := gomonkey.ApplyFunc(json.Marshal, func(v interface{}) ([]byte, error) {
				return nil, fmt.Errorf("err")
			})
			defer patch.Reset()
			rankTable := mockRankTableWithParam()
			result := rankTable.GetHccLJsonSlice()
			convey.So(result, convey.ShouldResemble, nil)
		})
	})
}

// TestGetPodNum test case for GetPodNum
func TestGetPodNum(t *testing.T) {
	convey.Convey("Testing GetPodNum", t, func() {
		rankTable := mockRankTableInit()
		rankTable.ServerList = []*ServerHccl{{
			DeviceList: []*Device{{
				DeviceID: "0",
				DeviceIP: "",
				RankID:   "",
			}},
			ServerID:   "",
			PodID:      "",
			ServerName: "node0",
				}}
		convey.Convey("Get pod num successfully", func() {
			result := rankTable.GetPodNum()
			convey.Convey("It should return the ServerList num", func() {
				convey.So(result, convey.ShouldEqual, 1)
			})
		})
	})
}

// TestGetFirstServerIp test case for GetFirstServerIp
func TestGetFirstServerIp(t *testing.T) {
	convey.Convey("Given a RankTable with ServerList", t, func() {
		server1 := &ServerHccl{
			ServerID:   "192.168.1.1",
			ServerName: "Server1",
			DeviceList: []*Device{
				{DeviceID: "0", DeviceIP: "192.168.1.10", RankID: "0"},
			},
		}
		server2 := &ServerHccl{
			ServerID:   "192.168.1.2",
			ServerName: "Server2",
			DeviceList: []*Device{
				{DeviceID: "1", DeviceIP: "192.168.1.11", RankID: "1"},
			},
		}
		rankTable := &RankTable{
			ServerList: []*ServerHccl{server1, server2},
		}

		convey.Convey("When calling GetFirstServerIp", func() {
			result := rankTable.GetFirstServerIp()
			convey.Convey("It should return the first server's ID", func() {
				convey.So(result, convey.ShouldEqual, "192.168.1.1")
			})
		})

		convey.Convey("When ServerList is empty", func() {
			emptyRankTable := &RankTable{
				ServerList: []*ServerHccl{{}},
			}
			result := emptyRankTable.GetFirstServerIp()
			convey.Convey("It should return an empty string", func() {
				convey.So(result, convey.ShouldEqual, "")
			})
		})

		convey.Convey("When rankTable is nil", func() {
			nilRankTable := new(RankTable)
			result := nilRankTable.GetFirstServerIp()
			convey.Convey("It should return an empty string", func() {
				convey.So(result, convey.ShouldEqual, "")
			})
		})
	})
}

func TestRemovePodInfo(t *testing.T) {
	convey.Convey("test RemovePodInfo", t, func() {
		rankTable := mockRankTableWithLength1()
		err := rankTable.RemovePodInfo(mockNamespace, mockPodUID1)
		convey.So(err, convey.ShouldNotBeNil)
		err = rankTable.RemovePodInfo(mockNamespace, "")
		convey.So(err, convey.ShouldBeNil)
	})
}
