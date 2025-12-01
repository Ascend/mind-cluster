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
	"net"
	"sync"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

func TestResetStatisticData(t *testing.T) {
	convey.Convey("test IcmpPingExecutor method updateStatisticData", t, func() {
		convey.Convey("01-should reset to initial value when function called", func() {
			stopChan := make(chan struct{})
			operator := NewOperator("127.0.0.2", "127.0.0.1", 1)
			executor := NewIcmpPingExecutor(stopChan, 0, operator)
			executor.statistics.SucPktNum = 1
			executor.resetStatisticData()
			convey.So(executor.statistics.SucPktNum, convey.ShouldEqual, 0)
		})
	})
}

func TestUpdateStatisticData(t *testing.T) {
	convey.Convey("test IcmpPingExecutor method updateStatisticData", t, func() {
		convey.Convey("01-should no change when send time is invalid", func() {
			executor := &IcmpPingExecutor{
				recvPktNum: 0,
				statistics: &IcmpPingMeshStatistics{
					Rtt:        make([]int64, 0),
					ActionTime: make([]int64, 0),
				},
				statisticsLock: &sync.RWMutex{},
			}
			sendTime := "xxx"
			executor.updateStatisticData(sendTime, nil, "")
			convey.So(executor.recvPktNum, convey.ShouldEqual, 0)
			convey.So(len(executor.statistics.Rtt), convey.ShouldEqual, 0)
			convey.So(len(executor.statistics.ActionTime), convey.ShouldEqual, 0)
		})
		convey.Convey("02-should increment when send time is valid", func() {
			executor := &IcmpPingExecutor{
				recvPktNum: 0,
				statistics: &IcmpPingMeshStatistics{
					Rtt:        make([]int64, 0),
					ActionTime: make([]int64, 0),
				},
				statisticsLock: &sync.RWMutex{},
			}
			pkt := &IcmpPacketInfo{
				ReceiveTime: time.Now(),
			}
			sendTime := "2006-01-02 15:04:05.0000000"
			executor.updateStatisticData(sendTime, pkt, "")
			convey.So(executor.recvPktNum, convey.ShouldEqual, 1)
			convey.So(len(executor.statistics.Rtt), convey.ShouldEqual, 1)
			convey.So(len(executor.statistics.ActionTime), convey.ShouldEqual, 1)
		})
	})
}

func TestCalcStatisticData(t *testing.T) {
	convey.Convey("test IcmpPingExecutor method calcStatisticData", t, func() {
		convey.Convey("01-should be init value when data is empty", func() {
			executor := &IcmpPingExecutor{
				waitingCache:   NewWaitingPktCache(),
				recvPktNum:     0,
				statistics:     NewIcmpPingMeshStatistics("A", "B"),
				statisticsLock: &sync.RWMutex{},
			}
			executor.calcStatisticData("")
			convey.So(executor.statistics.AvgTime, convey.ShouldEqual, -1)
			convey.So(executor.statistics.MaxTime, convey.ShouldEqual, -1)
			convey.So(executor.statistics.MinTime, convey.ShouldEqual, -1)
		})

		convey.Convey("02-should increment when data is valid", func() {
			executor := &IcmpPingExecutor{
				waitingCache:   NewWaitingPktCache(),
				recvPktNum:     0,
				statistics:     NewIcmpPingMeshStatistics("A", "B"),
				statisticsLock: &sync.RWMutex{},
			}
			executor.statistics.Rtt = []int64{4, 3, 0, 1, 2}
			executor.statistics.SucPktNum = int64(len(executor.statistics.Rtt))
			executor.calcStatisticData("")
			expectedMax := 4
			convey.So(executor.statistics.MaxTime, convey.ShouldEqual, expectedMax)
			convey.So(executor.statistics.MinTime, convey.ShouldEqual, 0)
			expectedDelta := 0.001
			expectedAvg := 2.000
			convey.So(executor.statistics.AvgTime, convey.ShouldAlmostEqual, expectedAvg, expectedDelta)
		})
	})
}

func TestIcmpPingExecutorInit(t *testing.T) {
	convey.Convey("test IcmpPingExecutor method init", t, func() {
		convey.Convey("01-should return err when resolve dst ip failed", func() {
			e := &IcmpPingExecutor{
				policy: NewOperator("a", "b", 1),
			}
			patch := gomonkey.ApplyFunc(net.ResolveIPAddr, func(network, address string) (*net.IPAddr, error) {
				return nil, errors.New("resolve ip addr failed")
			})
			defer patch.Reset()
			err := e.init("")
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("02-should return err when listen src ip failed", func() {
			e := &IcmpPingExecutor{
				policy: NewOperator("a", "b", 1),
			}
			patch := gomonkey.ApplyFunc(net.ResolveIPAddr, func(network, address string) (*net.IPAddr, error) {
				return &net.IPAddr{}, nil
			})
			defer patch.Reset()
			patch2 := gomonkey.ApplyFunc(icmp.ListenPacket, func(network, address string) (*icmp.PacketConn, error) {
				return nil, errors.New("listen packet failed")
			})
			defer patch2.Reset()
			err := e.init("")
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("03-should return err when listen src ip failed", func() {
			e := &IcmpPingExecutor{
				policy: NewOperator("192.168.0.1", "192.168.0.1", 1),
			}
			patch := gomonkey.ApplyFunc(net.ResolveIPAddr, func(network, address string) (*net.IPAddr, error) {
				return &net.IPAddr{}, nil
			})
			defer patch.Reset()
			patch2 := gomonkey.ApplyFunc(icmp.ListenPacket, func(network, address string) (*icmp.PacketConn, error) {
				return nil, nil
			})
			defer patch2.Reset()
			err := e.init("")
			convey.So(err, convey.ShouldBeNil)
			convey.So(e.curUid, convey.ShouldNotBeNil)
		})
	})
}

func TestSendPackets(t *testing.T) {
	convey.Convey("test IcmpPingExecutor method sendPackets", t, func() {
		convey.Convey("01-should add waiting cache size when send success", func() {
			e := &IcmpPingExecutor{
				waitingCache: NewWaitingPktCache(),
				conn:         &icmp.PacketConn{},
				policy:       NewOperator("a", "b", 1),
			}
			patch := gomonkey.ApplyMethod(e.conn, "WriteTo",
				func(c *icmp.PacketConn, b []byte, dst net.Addr) (int, error) {
					return 1, nil
				})
			defer patch.Reset()
			convey.So(e.waitingCache.Len(), convey.ShouldEqual, 0)
			e.sendPackets("")
			convey.So(e.waitingCache.Len(), convey.ShouldEqual, 1)
		})

		convey.Convey("02-should keep waiting cache size when send failed", func() {
			e := &IcmpPingExecutor{
				waitingCache: NewWaitingPktCache(),
				conn:         &icmp.PacketConn{},
				policy:       NewOperator("a", "b", 1),
			}
			patch := gomonkey.ApplyMethod(e.conn, "WriteTo",
				func(c *icmp.PacketConn, b []byte, dst net.Addr) (int, error) {
					return 0, errors.New("write to peer failed")
				})
			defer patch.Reset()
			convey.So(e.waitingCache.Len(), convey.ShouldEqual, 0)
			e.sendPackets("")
			convey.So(e.waitingCache.Len(), convey.ShouldEqual, 0)
		})
	})
}

func TestSendIcmpPacket(t *testing.T) {
	convey.Convey("test IcmpPingExecutor method sendIcmpPacket", t, func() {
		convey.Convey("01-should return err when conn is empty", func() {
			e := &IcmpPingExecutor{}
			err := e.sendIcmpPacket(0, "")
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("02-should return err when conn is empty", func() {
			e := &IcmpPingExecutor{
				waitingCache: NewWaitingPktCache(),
				conn:         &icmp.PacketConn{},
				policy:       NewOperator("a", "b", 1),
			}
			patch := gomonkey.ApplyMethod(e.conn, "WriteTo",
				func(c *icmp.PacketConn, b []byte, dst net.Addr) (int, error) {
					return 1, nil
				})
			defer patch.Reset()
			err := e.sendIcmpPacket(0, "")
			convey.So(err, convey.ShouldBeNil)
			convey.So(e.sendPktNum, convey.ShouldEqual, 1)
		})
	})
}

func TestProcessIcmpPacket(t *testing.T) {
	convey.Convey("test IcmpPingExecutor method processIcmpPacket", t, func() {
		testProcessIcmpPacketShouldFail()
		convey.Convey("03-should recvPktNum increment when pkt data type echo reply", func() {
			srcAddr := "127.0.0.1"
			dstAddr := "127.0.0.2"
			e := &IcmpPingExecutor{
				waitingCache:   NewWaitingPktCache(),
				policy:         NewOperator(dstAddr, srcAddr, 1),
				statistics:     NewIcmpPingMeshStatistics(srcAddr, dstAddr),
				statisticsLock: &sync.RWMutex{},
			}
			body := &IcmpBodyData{
				Uid: "870f890e9daf16dfaa28fdebd459626d", SeqId: 18, SendTime: "2025-07-08 22:08:53.8215865",
			}
			bodyBytes, errMarshal := json.Marshal(body)
			convey.So(errMarshal, convey.ShouldBeNil)
			curTime, err := time.ParseInLocation(specialTimeFormat, body.SendTime, time.Local)
			convey.So(err, convey.ShouldBeNil)
			patch := gomonkey.ApplyFunc(icmp.ParseMessage, func(proto int, b []byte) (*icmp.Message, error) {
				msg := &icmp.Message{
					Type: ipv4.ICMPTypeEchoReply, Code: 0, Checksum: 53140,
					Body: &icmp.Echo{ID: 23688, Seq: 18, Data: bodyBytes},
				}
				return msg, nil
			})
			defer patch.Reset()
			peer := net.IPAddr{IP: net.ParseIP(dstAddr)}
			pkt := IcmpPacketInfo{
				Data:        nil,
				Peer:        &peer,
				ReceiveTime: time.Now(),
			}
			e.waitingCache.addPktSeqIdToWaitingSet(body.Uid, body.SeqId, curTime.UnixMilli())
			e.processIcmpPacket(pkt, "")
			convey.So(e.recvPktNum, convey.ShouldEqual, 1)
		})
	})
}

func testProcessIcmpPacketShouldFail() {
	convey.Convey("01-should recvPktNum not increment when input is invalid", func() {
		e := &IcmpPingExecutor{}
		pkt := IcmpPacketInfo{}
		e.processIcmpPacket(pkt, "")
		convey.So(e.recvPktNum, convey.ShouldEqual, 0)
	})
	convey.Convey("02-should recvPktNum not increment when pkt data type is not echo reply", func() {
		e := &IcmpPingExecutor{}
		_, pktBytes, err := e.genEchoMsgBytes("")
		convey.So(err, convey.ShouldBeNil)
		pkt := IcmpPacketInfo{
			Data: pktBytes,
		}
		e.processIcmpPacket(pkt, "")
		convey.So(e.recvPktNum, convey.ShouldEqual, 0)
	})
}

type customTimeOutErr struct {
	errMsg string
}

func (r *customTimeOutErr) Timeout() bool {
	return true
}
func (r *customTimeOutErr) Error() string {
	return r.errMsg
}

func TestIsTimeoutErr(t *testing.T) {
	convey.Convey("test IcmpPingExecutor method isTimeoutErr", t, func() {
		convey.Convey("01-should return false when err is nil", func() {
			e := &IcmpPingExecutor{}
			ret := e.isTimeoutErr(nil)
			convey.So(ret, convey.ShouldBeFalse)
		})
		convey.Convey("02-should return false when err is not net.OpError", func() {
			e := &IcmpPingExecutor{}
			ret := e.isTimeoutErr(errors.New("normal error"))
			convey.So(ret, convey.ShouldBeFalse)
		})
		convey.Convey("03-should return true when err is net.OpError with timeout err", func() {
			e := &IcmpPingExecutor{}
			err := &net.OpError{
				Err: &customTimeOutErr{
					errMsg: "do some thing timeout",
				},
			}
			ret := e.isTimeoutErr(err)
			convey.So(ret, convey.ShouldBeTrue)
		})
	})
}

func TestReceiveIcmpPacket(t *testing.T) {
	convey.Convey("test IcmpPingExecutor method receiveIcmpPacket", t, func() {
		convey.Convey("01-should not receive pkt when stop chan is closed", func() {
			e := &IcmpPingExecutor{
				conn:       &icmp.PacketConn{},
				stopCh:     make(chan struct{}),
				recvPktNum: 0,
			}
			wg := &sync.WaitGroup{}
			recvChan := make(chan IcmpPacketInfo, 1)
			wg.Add(1)
			go e.receiveIcmpPacket(recvChan, wg)
			close(e.stopCh)
			wg.Wait()
			convey.So(e.recvPktNum, convey.ShouldEqual, 0)
		})
		convey.Convey("01-should receive pkt when conn is valid", func() {
			e := &IcmpPingExecutor{
				conn:       &icmp.PacketConn{},
				stopCh:     make(chan struct{}),
				recvPktNum: 0,
			}
			patch := gomonkey.ApplyMethodReturn(e.conn, "SetReadDeadline", nil)
			defer patch.Reset()
			patchRead := gomonkey.ApplyMethodReturn(e.conn, "ReadFrom", 1, &net.IPAddr{IP: net.ParseIP("127.0.0.1")},
				nil)
			defer patchRead.Reset()
			wg := &sync.WaitGroup{}
			recvChan := make(chan IcmpPacketInfo, 1)
			wg.Add(1)
			go e.receiveIcmpPacket(recvChan, wg)
			pkt := <-recvChan
			close(e.stopCh)
			wg.Wait()
			convey.So(len(pkt.Data), convey.ShouldEqual, 1)
		})
	})
}

func TestGatherCsvRecord(t *testing.T) {
	convey.Convey("test IcmpPingExecutor method gatherCsvRecord", t, func() {
		srcAddr := "127.0.0.1"
		dstAddr := "127.0.0.2"
		convey.Convey("01-should send data to channel when all is ok", func() {
			e := &IcmpPingExecutor{
				conn:           &icmp.PacketConn{},
				stopCh:         make(chan struct{}),
				recvPktNum:     0,
				policy:         NewOperator(dstAddr, srcAddr, 1),
				statistics:     NewIcmpPingMeshStatistics(srcAddr, dstAddr),
				statisticsLock: &sync.RWMutex{},
			}
			sendCh := make(chan statisticData, 1)
			convey.So(len(sendCh), convey.ShouldEqual, 0)
			e.gatherCsvRecord("", sendCh)
			convey.So(len(sendCh), convey.ShouldEqual, 1)
		})
		convey.Convey("02-should not send data to channel when stopCh is closed", func() {
			e := &IcmpPingExecutor{
				conn:           &icmp.PacketConn{},
				stopCh:         make(chan struct{}),
				recvPktNum:     0,
				policy:         NewOperator(dstAddr, srcAddr, 1),
				statistics:     NewIcmpPingMeshStatistics(srcAddr, dstAddr),
				statisticsLock: &sync.RWMutex{},
			}
			sendCh := make(chan statisticData)
			convey.So(len(sendCh), convey.ShouldEqual, 0)
			close(e.stopCh)
			e.gatherCsvRecord("", sendCh)
			convey.So(len(sendCh), convey.ShouldEqual, 0)
		})
		convey.Convey("03-should not send data to channel when json.Marshal failed", func() {
			e := &IcmpPingExecutor{
				conn:           &icmp.PacketConn{},
				stopCh:         make(chan struct{}),
				recvPktNum:     0,
				policy:         NewOperator(dstAddr, srcAddr, 1),
				statistics:     NewIcmpPingMeshStatistics(srcAddr, dstAddr),
				statisticsLock: &sync.RWMutex{},
			}
			sendCh := make(chan statisticData, 1)
			convey.So(len(sendCh), convey.ShouldEqual, 0)
			patch := gomonkey.ApplyFuncReturn(json.Marshal, nil, errors.New("json.Marshal failed"))
			defer patch.Reset()
			e.gatherCsvRecord("", sendCh)
			convey.So(len(sendCh), convey.ShouldEqual, 0)
		})
	})
}

func TestGetPingResultInfo(t *testing.T) {
	convey.Convey("test IcmpPingExecutor method getPingResultInfo", t, func() {
		srcAddr := "127.0.0.1"
		dstAddr := "127.0.0.2"
		convey.Convey("01-should get data from channel when all is ok", func() {
			e := &IcmpPingExecutor{
				conn:           &icmp.PacketConn{},
				stopCh:         make(chan struct{}),
				recvPktNum:     0,
				waitingCache:   NewWaitingPktCache(),
				policy:         NewOperator(dstAddr, srcAddr, 1),
				statistics:     NewIcmpPingMeshStatistics(srcAddr, dstAddr),
				statisticsLock: &sync.RWMutex{},
			}
			wg := &sync.WaitGroup{}
			sendCh := make(chan statisticData, 1)
			convey.So(len(sendCh), convey.ShouldEqual, 0)
			wg.Add(1)
			go e.getPingResultInfo(wg, sendCh)
			const waitTime = 11
			time.Sleep(waitTime * time.Second)
			close(e.stopCh)
			wg.Wait()
			convey.So(len(sendCh), convey.ShouldEqual, 1)
		})
		convey.Convey("02-should get non data from channel when ticker no triggered", func() {
			e := &IcmpPingExecutor{
				conn:           &icmp.PacketConn{},
				stopCh:         make(chan struct{}),
				recvPktNum:     0,
				waitingCache:   NewWaitingPktCache(),
				policy:         NewOperator(dstAddr, srcAddr, 1),
				statistics:     NewIcmpPingMeshStatistics(srcAddr, dstAddr),
				statisticsLock: &sync.RWMutex{},
			}
			wg := &sync.WaitGroup{}
			sendCh := make(chan statisticData, 1)
			convey.So(len(sendCh), convey.ShouldEqual, 0)
			wg.Add(1)
			go e.getPingResultInfo(wg, sendCh)
			close(e.stopCh)
			wg.Wait()
			convey.So(len(sendCh), convey.ShouldEqual, 0)
		})
	})
}
