// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package publicfault test for file about public fault
package publicfault

import (
	"os"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/utils"
	"clusterd/pkg/domain/publicfault"
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

func TestTryLoad(t *testing.T) {
	prepareFaultCfg(t)
	convey.Convey("test func tryLoadPubFaultCfgFromFile success", t, testTryLoad)
	convey.Convey("test func tryLoadPubFaultCfgFromFile success, load new file error", t, testTryLoadErrOne)
	convey.Convey("test func tryLoadPubFaultCfgFromFile failed, load files error", t, testTryLoadErr)
}

func prepareFaultCfg(t *testing.T) {
	const mode644 = 0644
	err := os.WriteFile(testFilePath, []byte(testFaultCfg), mode644)
	if err != nil {
		t.Error(err)
	}
}

func testTryLoad() {
	resetPubFaultResource()
	fileData, err := utils.LoadFile(testFilePath)
	convey.So(err, convey.ShouldBeNil)
	p1 := gomonkey.ApplyFuncReturn(utils.LoadFile, fileData, nil)
	defer p1.Reset()
	tryLoadPubFaultCfgFromFile()
	convey.So(publicfault.PubFaultResource, convey.ShouldResemble, []string{"CCAE", "fd-online", "pingmesh"})
}

func testTryLoadErrOne() {
	resetPubFaultResource()
	fileData, err := utils.LoadFile(testFilePath)
	convey.So(err, convey.ShouldBeNil)
	output := []gomonkey.OutputCell{
		{Values: gomonkey.Params{nil, testErr}, Times: 3},
		{Values: gomonkey.Params{fileData, nil}},
	}
	p1 := gomonkey.ApplyFuncSeq(utils.LoadFile, output)
	defer p1.Reset()
	tryLoadPubFaultCfgFromFile()
	convey.So(publicfault.PubFaultResource, convey.ShouldResemble, []string{"CCAE", "fd-online", "pingmesh"})
}

func testTryLoadErr() {
	resetPubFaultResource()
	p1 := gomonkey.ApplyFuncReturn(utils.LoadFile, nil, testErr)
	defer p1.Reset()
	tryLoadPubFaultCfgFromFile()
	convey.So(publicfault.PubFaultResource, convey.ShouldResemble, []string{})
}

func resetPubFaultResource() {
	publicfault.PubFaultResource = []string{}
}
