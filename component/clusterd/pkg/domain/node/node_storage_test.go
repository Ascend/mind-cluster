// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package node test for funcs about node
package node

import (
	"encoding/json"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"ascend-common/api"
	"clusterd/pkg/common/util"
)

const (
	nodeSN1       = "nodeSN1"
	nodeName1     = "node1"
	nodeIp1       = "10.10.10.10"
	nodeName2     = "node2"
	superPodIDStr = "0"
	notExistName  = "node exist node name"

	devName0       = "Ascend910-0"
	devPhyID0      = "0"
	ip0            = "192.168.1.0"
	devName1       = "Ascend910-1"
	ip1            = "192.168.1.1"
	devPhyID1      = "1"
	superPodID     = 0
	invalidDevName = "invalid device name"
	serverID2      = "2"
)

var (
	node          *v1.Node
	baseDeviceMap = map[string]*api.NpuBaseInfo{
		devName0: {
			IP:            ip0,
			SuperDeviceID: superPodID,
		},
		devName1: {
			IP:            ip1,
			SuperDeviceID: superPodID,
		},
	}
)

func resetCache() {
	cache = nodeCache{
		nodeInfoCache:      make(map[string]nodeInfo),
		nodeSNAndNameCache: make(map[string]string),
	}
}

func TestSaveNodeToCache(t *testing.T) {
	resetCache()

	convey.Convey("test func SaveNodeToCache failed, node is nil", t, func() {
		SaveNodeToCache(nil)
		convey.So(len(cache.nodeInfoCache), convey.ShouldEqual, 0)
		convey.So(len(cache.nodeSNAndNameCache), convey.ShouldEqual, 0)
	})

	convey.Convey("test func SaveNodeToCache", t, func() {
		SaveNodeToCache(node)
		convey.So(len(cache.nodeInfoCache), convey.ShouldEqual, 1)
		convey.So(len(cache.nodeSNAndNameCache), convey.ShouldEqual, 1)
		info := nodeInfo{
			nodeName:     nodeName1,
			nodeSN:       nodeSN1,
			nodeIP:       nodeIp1,
			superPodID:   superPodIDStr,
			baseDevInfos: baseDeviceMap,
			nodeDevice: &api.NodeDevice{
				NodeName:  nodeName1,
				DeviceMap: map[string]string{devPhyID0: superPodIDStr, devPhyID1: superPodIDStr},
			},
		}
		convey.So(cache.nodeInfoCache[nodeName1], convey.ShouldResemble, info)
		convey.So(cache.nodeSNAndNameCache[nodeSN1], convey.ShouldEqual, nodeName1)
	})
}

func TestDeleteNodeFromCache(t *testing.T) {
	resetCache()
	convey.Convey("test func DeleteNodeFromCache", t, func() {
		SaveNodeToCache(node)
		convey.So(len(cache.nodeInfoCache), convey.ShouldEqual, 1)
		convey.So(len(cache.nodeSNAndNameCache), convey.ShouldEqual, 1)
		DeleteNodeFromCache(node)
		convey.So(len(cache.nodeInfoCache), convey.ShouldEqual, 0)
		convey.So(len(cache.nodeSNAndNameCache), convey.ShouldEqual, 0)
	})
	convey.Convey("test func DeleteNodeFromCache failed, node is nil", t, func() {
		SaveNodeToCache(node)
		convey.So(len(cache.nodeInfoCache), convey.ShouldEqual, 1)
		convey.So(len(cache.nodeSNAndNameCache), convey.ShouldEqual, 1)
		DeleteNodeFromCache(nil)
		convey.So(len(cache.nodeInfoCache), convey.ShouldEqual, 1)
		convey.So(len(cache.nodeSNAndNameCache), convey.ShouldEqual, 1)
	})
}

func TestGetNodeNameBySN(t *testing.T) {
	resetCache()
	convey.Convey("test func GetNodeNameBySN success", t, func() {
		SaveNodeToCache(node)
		name, exist := GetNodeNameBySN(nodeSN1)
		convey.So(name, convey.ShouldEqual, nodeName1)
		convey.So(exist, convey.ShouldBeTrue)
	})
	convey.Convey("test func GetNodeNameBySN failed, node sn does not exist", t, func() {
		SaveNodeToCache(node)
		name, exist := GetNodeNameBySN(notExistName)
		convey.So(name, convey.ShouldEqual, "")
		convey.So(exist, convey.ShouldBeFalse)
	})
}

func TestGetNodeSNByName(t *testing.T) {
	resetCache()
	convey.Convey("test func GetNodeSNByName success", t, func() {
		SaveNodeToCache(node)
		sn := GetNodeSNByName(nodeName1)
		convey.So(sn, convey.ShouldEqual, nodeSN1)
	})
	convey.Convey("test func GetNodeSNByName failed, node sn does not exist", t, func() {
		SaveNodeToCache(node)
		sn := GetNodeSNByName(notExistName)
		convey.So(sn, convey.ShouldEqual, "")
	})
}

func TestGetNodeSNByIP(t *testing.T) {
	resetCache()
	convey.Convey("test func GetNodeIPAndSNMap success", t, func() {
		SaveNodeToCache(node)
		ipSnMap := GetNodeIPAndSNMap()
		expMap := map[string]string{nodeIp1: nodeSN1}
		convey.So(ipSnMap, convey.ShouldResemble, expMap)
	})
}

func TestGetNodeDeviceAndSuperPodID(t *testing.T) {
	resetCache()
	convey.Convey("test func GetNodeDeviceAndSuperPodID", t, testGetNodeDevAndSpID)
	convey.Convey("test func GetNodeDeviceAndSuperPodID success, node info is not in cache but can get normally",
		t, testGetNodeDevAndSpIDNotInCache)
	convey.Convey("test func GetNodeDeviceAndSuperPodID failed, node is nil", t, testGetNodeDevAndSpIDNilNode)
	convey.Convey("test func GetNodeDeviceAndSuperPodID failed, node name does not exist", t, testGetNodeDevAndSpIDNotExist)
	convey.Convey("test func GetNodeDeviceAndSuperPodID failed, dev is nil", t, testGetNodeDevAndSpIDNilDev)
	convey.Convey("test func GetNodeDeviceAndSuperPodID failed, deep copy error", t, testGetNodeDevAndSpIDErrDeepCp)
	convey.Convey("test func GetNodeDeviceAndSuperPodIDA5", t, testGetNodeDevAndSpIDNotInCacheA5)
}

func testGetNodeDevAndSpID() {
	SaveNodeToCache(node)
	nodeDev, spID := GetNodeDeviceAndSuperPodID(node)
	expDev := &api.NodeDevice{
		NodeName:  nodeName1,
		DeviceMap: map[string]string{devPhyID0: superPodIDStr, devPhyID1: superPodIDStr},
	}
	convey.So(nodeDev, convey.ShouldResemble, expDev)
	convey.So(spID, convey.ShouldEqual, superPodIDStr)

	// external cannot modify internal data
	nodeDev.NodeName = notExistName
	nodeDev, _ = GetNodeDeviceAndSuperPodID(node)
	convey.So(nodeDev, convey.ShouldResemble, expDev)
}

func testGetNodeDevAndSpIDNotInCache() {
	baseDevInfo, err := json.Marshal(baseDeviceMap)
	if err != nil {
		return
	}
	node2 := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: nodeName2,
			Annotations: map[string]string{
				api.NodeSNAnnotation: nodeSN1,
				superPodIDKey:        superPodIDStr,
				baseDevInfoAnno:      string(baseDevInfo),
			},
		},
	}
	nodeDev, spID := GetNodeDeviceAndSuperPodID(node2)
	expDev := &api.NodeDevice{
		NodeName:  nodeName2,
		DeviceMap: map[string]string{devPhyID0: superPodIDStr, devPhyID1: superPodIDStr},
	}
	convey.So(nodeDev, convey.ShouldResemble, expDev)
	convey.So(spID, convey.ShouldEqual, superPodIDStr)
}

func testGetNodeDevAndSpIDNotInCacheA5() {
	baseDevInfo, err := json.Marshal(baseDeviceMap)
	if err != nil {
		return
	}
	node2 := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: nodeName2,
			Annotations: map[string]string{
				api.NodeSNAnnotation: nodeSN1,
				superPodIDKey:        superPodIDStr,
				baseDevInfoAnno:      string(baseDevInfo),
				serverTypeKey:        api.VersionA5,
			},
		},
	}
	nodeDev, spID := GetNodeDeviceAndSuperPodID(node2)
	expDev := &api.NodeDevice{
		NodeName:   nodeName2,
		ServerType: api.VersionA5,
		DeviceMap:  map[string]string{devPhyID0: superPodIDStr, devPhyID1: superPodIDStr},
		NpuInfoMap: map[string]*api.NpuInfo{
			devPhyID0: {
				PhyId: devPhyID0,
			},
			devPhyID1: {
				PhyId: devPhyID1,
			},
		},
	}
	convey.So(nodeDev, convey.ShouldResemble, expDev)
	convey.So(spID, convey.ShouldEqual, superPodIDStr)
}

func testGetNodeDevAndSpIDNilNode() {
	nodeDev, spID := GetNodeDeviceAndSuperPodID(nil)
	convey.So(nodeDev, convey.ShouldBeNil)
	convey.So(spID, convey.ShouldEqual, "")
}

func testGetNodeDevAndSpIDNotExist() {
	notExistNode := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: notExistName,
		},
	}
	nodeDev, spID := GetNodeDeviceAndSuperPodID(notExistNode)
	convey.So(nodeDev, convey.ShouldBeNil)
	convey.So(spID, convey.ShouldEqual, "")
}

func testGetNodeDevAndSpIDNilDev() {
	notExistNode := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: notExistName,
		},
	}
	SaveNodeToCache(notExistNode)
	nodeDev, spID := GetNodeDeviceAndSuperPodID(notExistNode)
	convey.So(nodeDev, convey.ShouldBeNil)
	convey.So(spID, convey.ShouldEqual, "")
}

func testGetNodeDevAndSpIDErrDeepCp() {
	p1 := gomonkey.ApplyFuncReturn(util.DeepCopy, testErr)
	defer p1.Reset()
	nodeDev, spID := GetNodeDeviceAndSuperPodID(node)
	convey.So(nodeDev, convey.ShouldBeNil)
	convey.So(spID, convey.ShouldEqual, superPodIDStr)
}

func TestGetSuerPodID(t *testing.T) {
	resetCache()
	convey.Convey("test func getSuerPodID failed, empty super pod id", t, func() {
		notExistSpID := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: notExistName,
			},
		}
		spID := getSuerPodID(notExistSpID)
		convey.So(spID, convey.ShouldEqual, "")
	})
}

func TestGetBaseDevInfos(t *testing.T) {
	resetCache()
	convey.Convey("test func getBaseDevInfos failed, empty device info", t, func() {
		notExistDevInfo := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: notExistName,
			},
		}
		baseDevInfos := getBaseDevInfos(notExistDevInfo)
		convey.So(baseDevInfos, convey.ShouldBeNil)
	})

	convey.Convey("test func getBaseDevInfos failed, unmarshal error", t, func() {
		p1 := gomonkey.ApplyFuncReturn(json.Unmarshal, testErr)
		defer p1.Reset()
		baseDevInfos := getBaseDevInfos(node)
		convey.So(baseDevInfos, convey.ShouldBeNil)
	})
}

func TestGetNodeDevice(t *testing.T) {
	resetCache()
	convey.Convey("test func getNodeDevice failed, baseDevInfos is nil", t, func() {
		nodeDev := getNodeDevice(nil, nodeName1, "", "")
		convey.So(nodeDev, convey.ShouldBeNil)
	})

	convey.Convey("test func getNodeDevice failed, illegal device name", t, func() {
		baseDevInfos := map[string]*api.NpuBaseInfo{
			invalidDevName: {
				IP:            ip0,
				SuperDeviceID: superPodID,
			},
		}
		nodeDev := getNodeDevice(baseDevInfos, nodeName1, "", "")
		convey.So(nodeDev, convey.ShouldBeNil)
	})
}

// TestGetServerID test case for func getServerID
func TestGetServerID(t *testing.T) {
	baseDevInfo, err := json.Marshal(baseDeviceMap)
	if err != nil {
		return
	}
	convey.Convey("test func getServerID should return empty when input has no version key", t, func() {
		actual := getServerID(node)
		convey.So(actual, convey.ShouldBeEmpty)
	})

	convey.Convey("test func getServerID should return empty when version value is blank", t, func() {
		node2 := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: nodeName2,
				Annotations: map[string]string{
					api.NodeSNAnnotation: nodeSN1,
					superPodIDKey:        superPodIDStr,
					baseDevInfoAnno:      string(baseDevInfo),
					serverIndexKey:       " ",
				},
			},
		}
		actual := getServerID(node2)
		convey.So(actual, convey.ShouldBeEmpty)
	})

	convey.Convey("test func getServerID should return value when version value is valid", t, func() {
		node2 := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: nodeName2,
				Annotations: map[string]string{
					api.NodeSNAnnotation: nodeSN1,
					superPodIDKey:        superPodIDStr,
					baseDevInfoAnno:      string(baseDevInfo),
					serverIndexKey:       serverID2,
				},
			},
		}
		actual := getServerID(node2)
		convey.So(actual, convey.ShouldEqual, serverID2)
	})
}

// TestGetVersion test case for func getVersion
func TestGetVersion(t *testing.T) {
	baseDevInfo, err := json.Marshal(baseDeviceMap)
	if err != nil {
		return
	}
	convey.Convey("test func getVersion should return empty when input has no version key", t, func() {
		actual := getDeviceType(node)
		convey.So(actual, convey.ShouldBeEmpty)
	})

	convey.Convey("test func getVersion should return empty when version value is blank", t, func() {
		node2 := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: nodeName2,
				Annotations: map[string]string{
					api.NodeSNAnnotation: nodeSN1,
					superPodIDKey:        superPodIDStr,
					baseDevInfoAnno:      string(baseDevInfo),
					serverTypeKey:        "  ",
				},
			},
		}
		actual := getDeviceType(node2)
		convey.So(actual, convey.ShouldBeEmpty)
	})

	convey.Convey("test func getVersion should return value when version value is valid", t, func() {
		const testDevType = "testDevType"
		node2 := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: nodeName2,
				Annotations: map[string]string{
					api.NodeSNAnnotation: nodeSN1,
					superPodIDKey:        superPodIDStr,
					baseDevInfoAnno:      string(baseDevInfo),
					serverTypeKey:        testDevType,
				},
			},
		}
		actual := getDeviceType(node2)
		convey.So(actual, convey.ShouldEqual, testDevType)
	})
}
