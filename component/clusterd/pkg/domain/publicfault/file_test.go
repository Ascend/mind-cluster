// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package publicfault test for public fault file util
package publicfault

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
)

var (
	testFaultCfg = `
{
  "publicFaultCode": {
    "NotHandleFaultCodes":[
      "code1"
    ],
    "SubHealthFaultCodes":[
      "code2"
    ],
    "SeparateNPUCodes":[
      "code3"
    ]
  },
  "publicFaultResource": [
    "CCAE", "fd-online", "pingmesh"
  ]
}
`
)

func TestLoadPubFaultCfgFromFile(t *testing.T) {
	prepareFaultCfg(t)
	convey.Convey("test func LoadPubFaultCfgFromFile success", t, testLoadConfig)
	convey.Convey("test func LoadPubFaultCfgFromFile failed, load file error", t, testLoadConfigErrLoadFile)
	convey.Convey("test func LoadPubFaultCfgFromFile failed, unmarshal error", t, testLoadConfigErrUnmarshal)
}

func prepareFaultCfg(t *testing.T) {
	const mode644 = 0644
	err := os.WriteFile(testFilePath, []byte(testFaultCfg), mode644)
	if err != nil {
		t.Error(err)
	}
}

func testLoadConfig() {
	err := LoadPubFaultCfgFromFile(testFilePath)
	convey.So(err, convey.ShouldBeNil)
}

func testLoadConfigErrLoadFile() {
	p1 := gomonkey.ApplyFuncReturn(os.ReadFile, nil, testErr)
	defer p1.Reset()
	err := LoadPubFaultCfgFromFile(testFilePath)
	convey.So(err, convey.ShouldResemble, fmt.Errorf("load fault config from <%s> failed", testFilePath))
}

func testLoadConfigErrUnmarshal() {
	p1 := gomonkey.ApplyFuncReturn(json.Unmarshal, testErr)
	defer p1.Reset()
	err := LoadPubFaultCfgFromFile(testFilePath)
	convey.So(err, convey.ShouldResemble, fmt.Errorf("unmarshal from <%s> failed", path.Base(testFilePath)))
}
