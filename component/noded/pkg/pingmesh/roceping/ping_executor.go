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
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"

	"ascend-common/common-utils/hwlog"
)

// IcmpPingExecutor for icmp ping executor struct
type IcmpPingExecutor struct {
	stopCh         chan struct{}
	TaskId         uint32
	LinkPath       string
	policy         *IcmpPingMeshOperate
	statistics     *IcmpPingMeshStatistics
	statisticsLock *sync.RWMutex
	conn           *icmp.PacketConn
	receiveCh      chan string
	sequenceId     int
	curUid         string
	waitingCache   *WaitingPktCache
	sendPktNum     int64
	recvPktNum     int64
}

// NewIcmpPingExecutor create IcmpPingExecutor instance
func NewIcmpPingExecutor(stopCh chan struct{}, taskID uint32, operate *IcmpPingMeshOperate) *IcmpPingExecutor {
	return &IcmpPingExecutor{
		stopCh:         stopCh,
		TaskId:         taskID,
		LinkPath:       fmt.Sprintf("%s->%s", operate.SrcAddr, operate.DstAddr),
		policy:         operate,
		statisticsLock: &sync.RWMutex{},
		statistics:     NewIcmpPingMeshStatistics(operate.SrcAddr, operate.DstAddr),
		sequenceId:     0,
		waitingCache:   NewWaitingPktCache(),
	}
}

func (e *IcmpPingExecutor) resetStatisticData() {
	e.statisticsLock.Lock()
	defer e.statisticsLock.Unlock()
	e.statistics.SucPktNum = 0
	e.statistics.FailPktNum = 0
	e.statistics.MaxTime = -1
	e.statistics.MinTime = -1
	e.statistics.AvgTime = -1
	e.statistics.Rtt = make([]int64, 0)
	e.statistics.ActionTime = make([]int64, 0)
	e.statistics.TP95Time = -1
	e.statistics.ReplyStatNum = 0
	e.statistics.PingTotalNum = 0
}

func (e *IcmpPingExecutor) updateStatisticData(sendTimeStr string, pkt *IcmpPacketInfo, logHeader string) {
	logHeader = fmt.Sprintf("%s updateStatisticData", logHeader)
	sendTime, errParse := time.ParseInLocation(specialTimeFormat, sendTimeStr, time.Local)
	if errParse != nil {
		hwlog.RunLog.Errorf("%s parse send time string failed, err: %v", logHeader, errParse)
		return
	}
	rtt := pkt.ReceiveTime.Sub(sendTime)
	e.statisticsLock.Lock()
	defer e.statisticsLock.Unlock()
	e.statistics.Rtt = append(e.statistics.Rtt, rtt.Milliseconds())
	e.statistics.ActionTime = append(e.statistics.ActionTime, sendTime.UnixMilli())
	e.statistics.SucPktNum++
	e.recvPktNum++
}

func (e *IcmpPingExecutor) calcStatisticData(logHeader string) {
	e.statisticsLock.Lock()
	defer e.statisticsLock.Unlock()
	e.statistics.TP95Time = calcTP95Value(e.statistics.Rtt)
	totalTime := int64(0)
	for _, rtt := range e.statistics.Rtt {
		if e.statistics.MaxTime == -1 || e.statistics.MaxTime < rtt {
			e.statistics.MaxTime = rtt
		}
		if e.statistics.MinTime == -1 || e.statistics.MinTime > rtt {
			e.statistics.MinTime = rtt
		}
		totalTime += rtt
	}
	e.statistics.ReplyStatNum = e.sendPktNum - e.statistics.PingTotalNum
	e.statistics.PingTotalNum = e.sendPktNum
	failedCnt, detailItems := e.waitingCache.calcMayLostPktCnt()
	if failedCnt > 0 {
		for uid, vList := range detailItems {
			hwlog.RunLog.Debugf("%s packet is timed out, deleted from waiting cache, uid=%s, seqIds: %v", logHeader,
				uid, vList)
		}
	}
	e.statistics.FailPktNum = failedCnt
	if e.statistics.SucPktNum != 0 {
		e.statistics.AvgTime = float64(totalTime) / float64(e.statistics.SucPktNum)
	} else {
		e.statistics.AvgTime = -1
	}
	hwlog.RunLog.Debugf("%s result period: totalPingNum=%d, replyStatNum=%d, succPktNum=%d, failPktNum=%d, "+
		"MaxTime=%d, MinTime=%d, AvgTime=%f, TP95Time=%d, Rtt=%v, ActionTime=%v",
		logHeader, e.statistics.PingTotalNum, e.statistics.ReplyStatNum,
		e.statistics.SucPktNum, e.statistics.FailPktNum,
		e.statistics.MaxTime, e.statistics.MinTime, e.statistics.AvgTime, e.statistics.TP95Time,
		e.statistics.Rtt, e.statistics.ActionTime)
}

func calcTP95Value(arr []int64) int64 {
	arrLen := len(arr)
	if arrLen == 0 {
		return -1
	}
	newArr := make([]int64, arrLen)
	copy(newArr, arr)
	sort.Slice(newArr, func(i, j int) bool {
		return newArr[i] < newArr[j]
	})
	const hundred = 100
	const ninetyFive = 95
	return newArr[arrLen*ninetyFive/hundred]
}

func (e *IcmpPingExecutor) startPingTask(wg *sync.WaitGroup) {
	defer wg.Done()
	logHeader := fmt.Sprintf("[%s][ping task#%d][period(%d) ticker]", e.LinkPath, e.TaskId, e.policy.TaskInterval)
	hwlog.RunLog.Infof("%s period ticker started", logHeader)
	ticker := time.NewTicker(time.Duration(e.policy.TaskInterval) * time.Second)
	defer func() {
		ticker.Stop()
		hwlog.RunLog.Infof("%s stopped", logHeader)
		if e.conn == nil {
			hwlog.RunLog.Infof("%s conn is nil, no need close", logHeader)
			return
		}
		closeErr := e.conn.Close()
		if closeErr != nil {
			hwlog.RunLog.Errorf("%s close connection err: %v", logHeader, closeErr)
			return
		}
		hwlog.RunLog.Infof("%s close icmp connection success", logHeader)
	}()
	if err := e.init(logHeader); err != nil {
		hwlog.RunLog.Errorf("init ping executor failed before action, err : %v", err)
		return
	}
	const pktRecvChanBuffSize = 3
	recvChan := make(chan IcmpPacketInfo, pktRecvChanBuffSize)
	wg.Add(1)
	go e.receiveIcmpPacket(recvChan, wg)
	wg.Add(1)
	go e.processIcmpPacketTask(recvChan, wg)
	for {
		select {
		case <-e.stopCh:
			hwlog.RunLog.Infof("%s received stop signal, stop startPingTask, sendPktNum=%d", logHeader, e.sendPktNum)
			return
		case <-ticker.C:
			e.sendPackets(logHeader)
		}
	}
}

func (e *IcmpPingExecutor) sendPackets(logHeader string) {
	hwlog.RunLog.Debugf("%s will send %d packets in sequence", logHeader, e.policy.PktSendNum)
	for i := 0; i < e.policy.PktSendNum; i++ {
		if err := e.sendIcmpPacket(i, logHeader); err != nil {
			hwlog.RunLog.Warnf("%s send icmp packet failed, err: %v", logHeader, err)
		}
		time.Sleep(time.Duration(e.policy.PktInterval) * time.Millisecond)
	}
}

func (e *IcmpPingExecutor) init(logHeader string) error {
	if e.policy.DstIpAddr == nil {
		dst, err := net.ResolveIPAddr("ip4", e.policy.DstAddr)
		if err != nil {
			hwlog.RunLog.Errorf("%s resolve dst ip addr %s failed, err: %v", logHeader, e.policy.DstAddr, err)
			return err
		}
		hwlog.RunLog.Infof("%s resolve dest addr success, DstAddr: %s", logHeader, e.policy.DstAddr)
		e.policy.DstIpAddr = dst
	}
	if e.conn == nil {
		conn, err := icmp.ListenPacket("ip4:icmp", e.policy.SrcAddr)
		if err != nil {
			hwlog.RunLog.Errorf("%s listen packet to src ip addr %s failed, err: %v", logHeader,
				e.policy.SrcAddr, err)
			return err
		}
		hwlog.RunLog.Infof("%s create icmp connect success: srcAddr=%s", logHeader, e.policy.SrcAddr)
		e.conn = conn
	}
	if len(e.curUid) == 0 {
		uid, errUid := e.generateRandomUID(logHeader)
		if errUid != nil {
			hwlog.RunLog.Errorf("%s generate uid failed, err:%v", logHeader, errUid)
			return errUid
		}
		e.curUid = uid
		hwlog.RunLog.Infof("%s current packet uid is %v", logHeader, uid)
	}
	return nil
}

func (e *IcmpPingExecutor) sendIcmpPacket(index int, logHeader string) error {
	logHeader = fmt.Sprintf("%s[package#%d]", logHeader, index)
	data, wb, err := e.genEchoMsgBytes(logHeader)
	if err != nil {
		hwlog.RunLog.Errorf("%s generate echo msg bytes failed, err:%s", logHeader, err)
		return err
	}
	if e.conn == nil {
		hwlog.RunLog.Errorf("%s conn is empty, can not send msg", logHeader)
		return errors.New("conn is empty")
	}
	e.waitingCache.addPktSeqIdToWaitingSet(data.Uid, data.SeqId, time.Now().UnixMilli())
	hwlog.RunLog.Debugf("%s add sending pkt to waiting cache success: uid=%s, icmp_seq=%d dstAddr=%s", logHeader,
		data.Uid, data.SeqId, e.policy.DstAddr)
	if _, errWrite := e.conn.WriteTo(wb, e.policy.DstIpAddr); errWrite != nil {
		hwlog.RunLog.Errorf("%s send msg pkt(uid=%s, seqId=%d) to dstAddr %s failed, err:%s", logHeader, data.Uid,
			data.SeqId, e.policy.DstAddr, errWrite)
		if errDelete := e.waitingCache.deletePacketInWaitingSet(data); errDelete != nil {
			hwlog.RunLog.Errorf("%s delete pkt(uid=%s, seqId=%d) to dstAddr %s from waiting cache failed, err:%s",
				logHeader, data.Uid, data.SeqId, e.policy.DstAddr, errDelete)
			return errWrite
		}
		hwlog.RunLog.Infof("%s delete pkt(uid=%s, seqId=%d) to dstAddr %s from waiting cache success",
			logHeader, data.Uid, data.SeqId, e.policy.DstAddr)
		return errWrite
	}
	hwlog.RunLog.Debugf("%s send icmp echo request success: uid=%s, icmp_seq=%d dstAddr=%s", logHeader, data.Uid,
		data.SeqId, e.policy.DstAddr)
	e.sequenceId++
	e.sendPktNum++
	if e.sequenceId > maxIcmpSequenceId {
		e.sequenceId = 0
		uid, errUid := e.generateRandomUID(logHeader)
		if errUid != nil {
			hwlog.RunLog.Errorf("%s generate uid failed, err:%v", logHeader, errUid)
			return errUid
		}
		hwlog.RunLog.Infof("%s pkt uid will change from %s to %s", logHeader, e.curUid, uid)
		e.curUid = uid
	}
	return nil
}

func (e *IcmpPingExecutor) genEchoMsgBytes(logHeader string) (*IcmpBodyData, []byte, error) {
	sendTime := time.Now().Format(specialTimeFormat)
	data := IcmpBodyData{
		Uid:      e.curUid,
		SendTime: sendTime,
		SeqId:    e.sequenceId,
	}
	dataBytes, errMarshal := json.Marshal(data)
	if errMarshal != nil {
		hwlog.RunLog.Errorf("%s icmp body data Marshal failed, err:%s", logHeader, errMarshal)
		return nil, nil, errMarshal
	}
	const pidMask = 0xffff
	msg := icmp.Message{
		Type: ipv4.ICMPTypeEcho, Code: 0,
		Body: &icmp.Echo{
			ID:   os.Getpid() & pidMask,
			Seq:  e.sequenceId,
			Data: dataBytes,
		},
	}
	msgBytes, err := msg.Marshal(nil)
	if err != nil {
		hwlog.RunLog.Errorf("%s icmp echo msg Marshal failed, err:%s", logHeader, err)
		return nil, nil, err
	}
	return &data, msgBytes, nil
}

func (e *IcmpPingExecutor) generateRandomUID(logHeader string) (string, error) {
	logHeader = fmt.Sprintf("%s[generateRandomUID]", logHeader)
	const uidByteLen = 16
	uidBytes := make([]byte, uidByteLen)
	n, err := rand.Read(uidBytes)
	if err != nil {
		hwlog.RunLog.Errorf("%s read random bytes failed, err: %v", logHeader, err)
		return "", err
	}
	if n != uidByteLen {
		hwlog.RunLog.Errorf("%s read random bytes len is not enough, len: %v", logHeader, n)
		return "", errors.New("random bytes is not enough")
	}
	return hex.EncodeToString(uidBytes), nil
}

func (e *IcmpPingExecutor) processIcmpPacketTask(recvChan chan IcmpPacketInfo, wg *sync.WaitGroup) {
	defer wg.Done()
	logHeader := fmt.Sprintf("[%s][processIcmpPacket goroutine task#%d]", e.LinkPath, e.TaskId)
	processCnt := uint64(0)
	for {
		select {
		case <-e.stopCh:
			hwlog.RunLog.Infof("%s received stop signal, stop processIcmpPacket, recvPktNum=%d, waitingForReply: %v",
				logHeader, e.recvPktNum, e.waitingCache)
			return
		case pkt := <-recvChan:
			hwlog.RunLog.Debugf("%s received len(bytes)=%d, processCnt=%d, pkt=%v",
				logHeader, len(pkt.Data), processCnt, pkt)
			e.processIcmpPacket(pkt, logHeader)
			hwlog.RunLog.Debugf("%s processCnt=%d done", logHeader, processCnt)
			processCnt++
		}
	}
}

func (e *IcmpPingExecutor) processIcmpPacket(pkt IcmpPacketInfo, logHeader string) {
	rm, err := icmp.ParseMessage(ipv4.ICMPTypeEchoReply.Protocol(), pkt.Data)
	if err != nil {
		hwlog.RunLog.Errorf("%s parse message from receive channel failed, err: %v", logHeader, err)
		return
	}
	switch rm.Type {
	case ipv4.ICMPTypeEchoReply:
		if pkt.Peer.String() != e.policy.DstAddr {
			return
		}
		dataInfo, errParse := e.parseMsgBody(rm)
		if errParse != nil {
			hwlog.RunLog.Errorf("%s parse echo reply body failed, err: %v", logHeader, errParse)
			return
		}
		hwlog.RunLog.Debugf("%s packet(uid=%s, seqId=%d) is received", logHeader, dataInfo.Uid, dataInfo.SeqId)
		if err = e.waitingCache.isPacketValid(dataInfo); err != nil {
			hwlog.RunLog.Errorf("%s packet(uid=%s, seqId=%d) is invalid, err: %v", logHeader, dataInfo.Uid,
				dataInfo.SeqId, err)
			return
		}
		hwlog.RunLog.Debugf("%s packet(uid=%s, seqId=%d) is valid by waiting cache", logHeader, dataInfo.Uid,
			dataInfo.SeqId)
		if errDelete := e.waitingCache.deletePacketInWaitingSet(dataInfo); errDelete != nil {
			hwlog.RunLog.Warnf("%s delete pkt(uid=%s, seqId=%d) in waiting set failed, err: %v", logHeader,
				dataInfo.Uid, dataInfo.SeqId, errDelete)
			return
		}
		hwlog.RunLog.Debugf("%s delete packet(uid=%s, seqId=%d) from waiting cache success", logHeader, dataInfo.Uid,
			dataInfo.SeqId)
		e.updateStatisticData(dataInfo.SendTime, &pkt, logHeader)
		hwlog.RunLog.Debugf("%s update statistic data by packet(uid=%s, seqId=%d) success", logHeader, dataInfo.Uid,
			dataInfo.SeqId)
	default:
		hwlog.RunLog.Debugf("%s received unexpected icmp message type: %v", logHeader, rm.Type)
	}
}

func (e *IcmpPingExecutor) parseMsgBody(msg *icmp.Message) (*IcmpBodyData, error) {
	if msg == nil {
		return nil, errors.New("input msg is empty")
	}
	echoReply, ok := msg.Body.(*icmp.Echo)
	if !ok {
		return nil, errors.New("invalid echo reply body type")
	}
	var dataInfo IcmpBodyData
	if errUnmarshal := json.Unmarshal(echoReply.Data, &dataInfo); errUnmarshal != nil {
		return nil, fmt.Errorf("unmarshal echo reply body failed, err: %v", errUnmarshal)
	}
	if echoReply.Seq != dataInfo.SeqId {
		return nil, fmt.Errorf("seqId in icmp echo reply message is not consistent, seq in header : %v, "+
			"seq in body: %v", echoReply.Seq, dataInfo.SeqId)
	}
	return &dataInfo, nil
}

func (e *IcmpPingExecutor) receiveIcmpPacket(recvChan chan IcmpPacketInfo, wg *sync.WaitGroup) {
	defer wg.Done()
	logHeader := fmt.Sprintf("[%s][receiveIcmpPacket goroutine task#%d]", e.LinkPath, e.TaskId)
	recvCnt := uint64(0)
	for {
		select {
		case <-e.stopCh:
			hwlog.RunLog.Infof("%s received stop signal, stop receiveIcmpPacket", logHeader)
			return
		default:
			if err := e.conn.SetReadDeadline(time.Now().Add(time.Duration(readTimeout) * time.Second)); err != nil {
				hwlog.RunLog.Errorf("%s set read deadline failed, err: %v", logHeader, err)
				time.Sleep(time.Second)
				continue
			}
			rb := make([]byte, readBuffSize)
			n, peer, err := e.conn.ReadFrom(rb)
			if err != nil && !e.isTimeoutErr(err) {
				hwlog.RunLog.Errorf("%s receive icmp echo reply failed, received %d bytes, peer: %v, "+
					"err: %v", logHeader, n, peer, err)
				continue
			}
			hwlog.RunLog.Debugf("%s receive icmp echo reply success, received %d bytes, recvCnt=%d",
				logHeader, n, recvCnt)
			if n == 0 {
				hwlog.RunLog.Debugf("%s n == 0, will continue", logHeader)
				continue
			}
			pkt := IcmpPacketInfo{
				Data:        rb[:n],
				ReceiveTime: time.Now(),
				Peer:        peer,
			}
			select {
			case <-e.stopCh:
				hwlog.RunLog.Infof("%s received stop signal, stop send data to receive channel", logHeader)
				return
			case recvChan <- pkt:
				hwlog.RunLog.Debugf("%s send data to receive channel success, recvCnt=%d", logHeader, recvCnt)
				recvCnt++
			}
		}
	}
}

func (e *IcmpPingExecutor) isTimeoutErr(err error) bool {
	var errNet *net.OpError
	ok := errors.As(err, &errNet)
	if !ok {
		return false
	}
	if errNet == nil {
		return false
	}
	return errNet.Timeout()
}

func (e *IcmpPingExecutor) getPingResultInfo(wg *sync.WaitGroup, sendCh chan statisticData) {
	defer wg.Done()
	const collectPeriodFactor = 10
	logHeader := fmt.Sprintf("[%s][ping collect task#%d][period(%ds) ticker]", e.LinkPath, e.TaskId,
		e.policy.TaskInterval*collectPeriodFactor)
	ticker := time.NewTicker(time.Duration(e.policy.TaskInterval*collectPeriodFactor) * time.Second)
	defer func() {
		ticker.Stop()
		hwlog.RunLog.Infof("%s stopped", logHeader)
	}()
	for {
		select {
		case <-e.stopCh:
			hwlog.RunLog.Infof("%s received stop signal, stop getPingResultInfo", logHeader)
			return
		case <-ticker.C:
			e.calcStatisticData(logHeader)
			e.gatherCsvRecord(logHeader, sendCh)
			e.resetStatisticData()
		}
	}
}

func (e *IcmpPingExecutor) gatherCsvRecord(logHeader string, sendCh chan statisticData) {
	e.statisticsLock.RLock()
	avgLossRateStr := calcAvgLossRate(e.statistics.SucPktNum, e.statistics.FailPktNum)
	record := []string{
		strconv.Itoa(int(e.TaskId)), // taskID
		strconv.Itoa(0),             // srcType
		e.policy.SrcAddr,            // srcAddr
		strconv.Itoa(0),             // dstType
		e.policy.DstAddr,            // dstAddr
		strconv.FormatInt(e.statistics.MinTime, digitalBase),                                                 // minDelay
		strconv.FormatInt(e.statistics.MaxTime, digitalBase),                                                 // maxDelay
		strconv.FormatFloat(e.statistics.AvgTime, float64FormatType, float64FormatPrecision, float64BitSize), // avgDelay
		avgLossRateStr, // minLossRate use the avgLossRate value
		avgLossRateStr, // maxLossRate use the avgLossRate value
		avgLossRateStr, // avgLossRate
		strconv.FormatInt(time.Now().UnixMilli(), digitalBase), // timestamp use the write time stamp
	}
	ri := resultInfo{
		SourceAddr:   e.statistics.SrcAddr,
		TargetAddr:   e.statistics.DstAddr,
		SucPktNum:    uint(e.statistics.SucPktNum),
		FailPktNum:   uint(e.statistics.FailPktNum),
		MaxTime:      int(e.statistics.MaxTime),
		MinTime:      int(e.statistics.MinTime),
		AvgTime:      int(e.statistics.AvgTime),
		TP95Time:     int(e.statistics.TP95Time),
		ReplyStatNum: int(e.statistics.ReplyStatNum),
		PingTotalNum: int(e.statistics.PingTotalNum),
	}
	e.statisticsLock.RUnlock()
	b, err := json.Marshal(ri)
	if err != nil {
		hwlog.RunLog.Errorf("json marshal error: %v", err)
		return
	}
	data := statisticData{result: string(b), record: record}
	hwlog.RunLog.Debugf("%s begin to write record to receive channel", logHeader)
	select {
	case <-e.stopCh:
		hwlog.RunLog.Infof("%s received stop signal, stop gatherCsvRecord", logHeader)
		return
	case sendCh <- data:
		hwlog.RunLog.Debugf("%s write record to receive channel success", logHeader)
	}
}

func calcAvgLossRate(sucPktNum, failPktNum int64) string {
	var avgLossRate float64
	totalPkgNum := sucPktNum + failPktNum
	if totalPkgNum != 0 {
		avgLossRate = float64(failPktNum) / float64(totalPkgNum)
	}
	avgLossRateStr := strconv.FormatFloat(avgLossRate, float64FormatType, float64FormatPrecision,
		float64BitSize)
	return avgLossRateStr
}
