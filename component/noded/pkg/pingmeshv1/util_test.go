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

/*
Package pingmeshv1 is using for checking hccs network
*/

package pingmeshv1

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"nodeD/pkg/pingmeshv1/types"
)

func TestGenerateJobUID(t *testing.T) {
	type testCase struct {
		name      string
		config    *types.HccspingMeshConfig
		destAddrs map[string]types.SuperDeviceIDs
		expectErr error
	}
	testCases := []testCase{
		{
			name:   "Normal case",
			config: &types.HccspingMeshConfig{},
			destAddrs: map[string]types.SuperDeviceIDs{
				"node1": map[string]string{"1": "111"},
			},
			expectErr: nil,
		},
		{
			name:      "Nil config",
			config:    nil,
			destAddrs: map[string]types.SuperDeviceIDs{},
			expectErr: nil,
		},
	}

	convey.Convey("Testing generateJobUID function", t, func() {
		for _, tc := range testCases {
			convey.Convey(tc.name, func() {
				uid, err := generateJobUID(tc.config, tc.destAddrs)
				convey.So(err, convey.ShouldBeNil)
				cfg, err := json.Marshal(tc.config)
				convey.So(err, convey.ShouldBeNil)
				address, err := json.Marshal(tc.destAddrs)
				convey.So(err, convey.ShouldBeNil)
				address = append(address, cfg...)
				hasher := sha256.New()
				_, err = hasher.Write(address)
				convey.So(err, convey.ShouldBeNil)
				expectedUID := hex.EncodeToString(hasher.Sum(nil))
				convey.So(uid, convey.ShouldEqual, expectedUID)
			})
		}
	})
}
