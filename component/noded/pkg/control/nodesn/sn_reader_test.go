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

// Package control for the node controller test
package nodesn

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"github.com/u-root/u-root/pkg/ipmi"

	"ascend-common/common-utils/hwlog"
)

var (
	testSNByte      = []byte{0x47, 0x48, 0x49, 0x50}
	testEmptySNByte = []byte{0x47, 0x48}
)

const testSN = "IP"

func init() {
	logConfig := &hwlog.LogConfig{
		OnlyToStdout: true,
	}
	if err := hwlog.InitRunLogger(logConfig, context.Background()); err != nil {
		fmt.Printf("init hwlog failed, %v\n", err)
	}
}

func TestSNReader(t *testing.T) {
	convey.Convey("test GetNodeSN", t, func() {
		convey.Convey("when Open ipmi success and get serial numbers success, err should be nil", func() {
			var p1 = gomonkey.ApplyFuncReturn(ipmi.Open, &ipmi.IPMI{}, nil)
			defer p1.Reset()
			var p2 = gomonkey.ApplyMethodReturn(&ipmi.IPMI{}, "RawCmd", testSNByte, nil)
			defer p2.Reset()
			var p3 = gomonkey.ApplyMethodReturn(&ipmi.IPMI{}, "Close", nil)
			defer p3.Reset()
			sn, err := GetNodeSN()
			convey.So(err, convey.ShouldBeNil)
			convey.So(sn, convey.ShouldEqual, testSN)
		})
		convey.Convey("when Open ipmi failed and get serial numbers success, err should not be nil", func() {
			var p1 = gomonkey.ApplyFuncReturn(ipmi.Open, &ipmi.IPMI{}, errors.New("test error"))
			defer p1.Reset()
			var p2 = gomonkey.ApplyMethodReturn(&ipmi.IPMI{}, "RawCmd", testSNByte, nil)
			defer p2.Reset()
			var p3 = gomonkey.ApplyMethodReturn(&ipmi.IPMI{}, "Close", nil)
			defer p3.Reset()
			_, err := GetNodeSN()
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("when Open ipmi success and get serial numbers empty "+
			"and get manufacturer name empty, err should not be nil and sn should be empty", func() {
			var p1 = gomonkey.ApplyFuncReturn(ipmi.Open, &ipmi.IPMI{}, nil)
			defer p1.Reset()
			var p2 = gomonkey.ApplyMethodReturn(&ipmi.IPMI{}, "RawCmd",
				testEmptySNByte, nil)
			defer p2.Reset()
			var p3 = gomonkey.ApplyMethodReturn(&ipmi.IPMI{}, "Close", nil)
			defer p3.Reset()
			sn, err := GetNodeSN()
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(sn, convey.ShouldBeEmpty)
		})
	})
}
