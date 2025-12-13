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

	"github.com/smartystreets/goconvey/convey"

	"clusterd/pkg/common/constant"
	"clusterd/pkg/interface/grpc/job"
)

const (
	jobAddTime = 1630000000
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

// TestBuildJobSignalFromJobInfo test build signal
func TestBuildJobSignalFromJobInfo(t *testing.T) {
	convey.Convey("test BuildJobSignalFromJobInfo", t, func() {
		convey.Convey("PyTorch init status", func() {
			jobInfo := constant.JobInfo{
				Key:         "job-1",
				Name:        "test-job",
				NameSpace:   "default",
				Framework:   ptFramework,
				Status:      "running",
				AddTime:     jobAddTime,
				TotalCmNum:  1,
				SharedTorIp: "192.168.1.1",
				MasterAddr:  "192.168.1.2:29500",
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
