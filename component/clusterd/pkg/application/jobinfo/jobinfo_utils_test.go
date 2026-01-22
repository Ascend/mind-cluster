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

// Package jobinfo is used to return job info by subscribe

package jobinfo

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"

	"clusterd/pkg/common/constant"
	"clusterd/pkg/interface/grpc/job"
)

const (
	jobName1           = "job1"
	jobName2           = "job2"
	jobName3           = "job3"
	jobNameSpace       = "default"
	three              = 3
	four               = 4
	five               = 5
	seven              = 7
	jobAddTime         = 1630000000
	statusJobCompleted = "complete"
)

// TestSendJobInfoSignal test send signal
func TestSendJobInfoSignal(t *testing.T) {
	convey.Convey("test SendJobInfoSignal function", t, func() {
		originalChan := jobUpdateChan
		defer func() { jobUpdateChan = originalChan }()

		convey.Convey("test  SendJobInfoSignal while chan not blocking", func() {
			jobUpdateChan = make(chan job.JobSummarySignal, 1)
			signal := job.JobSummarySignal{JobId: "test-job-1"}

			SendJobInfoSignal(signal)

			convey.So(len(jobUpdateChan), convey.ShouldEqual, 1)
			received := <-jobUpdateChan
			convey.So(received.JobId, convey.ShouldEqual, signal.JobId)
		})
	})
}

// TestSendJobInfoSignalDrop test drop msg
func TestSendJobInfoSignalDrop(t *testing.T) {
	convey.Convey("test while jobUpdateChan full", t, func() {
		jobUpdateChan = make(chan job.JobSummarySignal, jobUpdateChanCache)

		for i := 0; i < jobUpdateChanCache; i++ {
			SendJobInfoSignal(job.JobSummarySignal{JobId: fmt.Sprintf("job-%d", i)})
		}

		convey.Convey("11 signal go default", func() {
			signal := job.JobSummarySignal{JobId: "overflow-job"}
			SendJobInfoSignal(signal)
			convey.So(len(jobUpdateChan), convey.ShouldEqual, jobUpdateChanCache)
		})
		convey.Reset(func() {
			close(jobUpdateChan)
		})
	})
}

func mockJobInfoWithNPU(npuNumPerServer ...int) constant.JobInfo {
	serverList := make([]constant.ServerHccl, 0, len(npuNumPerServer))
	for i, npuNum := range npuNumPerServer {
		deviceList := make([]constant.Device, 0, npuNum)
		for j := 0; j < npuNum; j++ {
			deviceList = append(deviceList, constant.Device{
				DeviceID: strconv.Itoa(j),
				DeviceIP: "192.168.0.1",
				RankID:   strconv.Itoa(j),
			})
		}
		serverList = append(serverList, constant.ServerHccl{
			ServerID:   strconv.Itoa(i),
			HostIp:     "192.168.0.1",
			DeviceList: deviceList,
		})
	}

	return constant.JobInfo{
		Key:         jobName1,
		Name:        jobName1,
		NameSpace:   jobNameSpace,
		Framework:   ptFramework,
		Status:      statusJobCompleted,
		AddTime:     time.Now().Unix(),
		TotalCmNum:  1,
		SharedTorIp: "10.0.0.1",
		MasterAddr:  "10.0.0.2",
		JobRankTable: constant.RankTable{
			ServerList: serverList,
		},
	}
}

func TestCalcJobNPUNum(t *testing.T) {
	convey.Convey("Test calcJobNPUNum", t, func() {
		convey.Convey("no ServerList, NPU count is 0", func() {
			jobInfo := constant.JobInfo{
				JobRankTable: constant.RankTable{ServerList: []constant.ServerHccl{}},
			}
			npuNum := calcJobNPUNum(jobInfo)
			convey.So(npuNum, convey.ShouldEqual, 0)
		})
		convey.Convey("single server with multiple devices, NPU count equals the total number of devices",
			func() {
				jobInfo := mockJobInfoWithNPU(five)
				npuNum := calcJobNPUNum(jobInfo)
				convey.So(npuNum, convey.ShouldEqual, five)
			})
		convey.Convey("multiple servers and multiple devices, with the total number of NPUs "+
			"being the sum across all devices", func() {
			jobInfo := mockJobInfoWithNPU(three, four)
			npuNum := calcJobNPUNum(jobInfo)
			convey.So(npuNum, convey.ShouldEqual, seven)
		})
		convey.Convey("server has no Device, NPU count is 0", func() {
			jobInfo := mockJobInfoWithNPU(0, 0)
			npuNum := calcJobNPUNum(jobInfo)
			convey.So(npuNum, convey.ShouldEqual, 0)
		})
	})
}

// TestBuildJobSignalFromJobInfo test build signal
func TestBuildJobSignalFromJobInfo(t *testing.T) {
	convey.Convey("test BuildJobSignalFromJobInfo", t, func() {
		convey.Convey("PyTorch init status", func() {
			jobInfo := constant.JobInfo{
				Key:        jobName1,
				Name:       jobName1,
				NameSpace:  jobNameSpace,
				Framework:  ptFramework,
				Status:     "running",
				AddTime:    jobAddTime,
				TotalCmNum: 1,
			}

			signal := BuildJobSignalFromJobInfo(jobInfo, constant.DefaultHcclJson, constant.AddOperator)

			convey.So(signal.JobId, convey.ShouldEqual, jobInfo.Key)
			convey.So(signal.JobName, convey.ShouldEqual, jobInfo.Name)
			convey.So(signal.Namespace, convey.ShouldEqual, jobInfo.NameSpace)
			convey.So(signal.FrameWork, convey.ShouldEqual, jobInfo.Framework)
			convey.So(signal.JobStatus, convey.ShouldEqual, jobInfo.Status)
			convey.So(signal.Time, convey.ShouldEqual, strconv.Itoa(int(jobInfo.AddTime)))
			convey.So(signal.CmIndex, convey.ShouldEqual, "0")
			convey.So(signal.Operator, convey.ShouldEqual, constant.AddOperator)
			convey.So(signal.Total, convey.ShouldEqual, strconv.Itoa(jobInfo.TotalCmNum))
			convey.So(signal.HcclJson, convey.ShouldEqual, constant.DefaultHcclJson)
			convey.So(signal.SharedTorIp, convey.ShouldEqual, jobInfo.SharedTorIp)
			convey.So(signal.MasterAddr, convey.ShouldEqual, jobInfo.MasterAddr)
			convey.So(signal.DeleteTime, convey.ShouldBeEmpty)
		})
	})
}
