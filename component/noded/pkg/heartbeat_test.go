/* Copyright(C) 2023. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package pkg for noded ut
package pkg

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
)

// TestSendHeartbeat test function SendHeartbeat
func TestSendHeartbeat(t *testing.T) {
	convey.Convey("SendHeartbeat test", t, func() {
		patch := gomonkey.ApplyFunc(os.Getenv, func(_ string) string {
			return ubuntuHostName
		})
		defer patch.Reset()
		convey.Convey("kubeClient create failed", func() {
			patches := gomonkey.ApplyFunc(newClientK8s, func() (*kubernetes.Clientset, error) {
				return nil, fmt.Errorf("error")
			})
			defer patches.Reset()
			err := SendHeartbeat(DefaultHeartbeatInterval)
			convey.So(err, convey.ShouldBeError)
		})
		// heartbeatSender create success
		convey.Convey("heartbeatSender create success", func() {
			patches := gomonkey.ApplyFunc(newClientK8s, func() (*kubernetes.Clientset, error) {
				return nil, nil
			})
			defer patches.Reset()
			_, err := newHeartbeatSender(ubuntuHostName, DefaultHeartbeatInterval)
			convey.So(err, convey.ShouldBeNil)
		})
		patchUtil := gomonkey.ApplyFunc(wait.Until, func(f func(), period time.Duration, stopCh <-chan struct{}) {
			return
		})
		defer patchUtil.Reset()
	})
}

// TestCheckNodeName test function checkNodeName
func TestCheckNodeName(t *testing.T) {
	convey.Convey("checkNodeName test", t, func() {
		testCase := []struct {
			caseName string
			nodeName string
		}{
			{"nodeName length is big", strings.Repeat("t", illegalLength)},
			{"nodeName is illegal", "@!tabel$"},
		}
		for _, tCase := range testCase {
			err := checkNodeName(tCase.nodeName)
			convey.So(err, convey.ShouldBeError)
		}
		err := checkNodeName(ubuntuHostName)
		convey.So(err, convey.ShouldBeNil)
	})
}
