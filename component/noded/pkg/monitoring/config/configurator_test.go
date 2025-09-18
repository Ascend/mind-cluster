//go:build !race

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

// Package config for the fault config test

package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"

	"ascend-common/common-utils/utils"
	"nodeD/pkg/common"
	"nodeD/pkg/kubeclient"
)

var configManager *FaultConfigurator

func TestFaultConfigurator(t *testing.T) {
	configManager = NewFaultConfigurator(testK8sClient)
	convey.Convey("test FaultConfigurator method 'Stop'", t, testFaultCfgStop)
	convey.Convey("test FaultConfigurator method 'Init'", t, testFaultCfgInit)
	convey.Convey("test FaultConfigurator method 'AddConfigCM'", t, testAddConfigCM)
	convey.Convey("test FaultConfigurator method 'UpdateConfigCM'", t, testUpdateConfigCM)
	convey.Convey("test FaultConfigurator method 'DeleteConfigCM'", t, testDeleteConfigCM)
	convey.Convey("test FaultConfigurator method 'GetMonitorData'", t, testGetMonitorData)

	convey.Convey("test FaultConfigurator method 'initFaultConfigFromCM'", t, testInitFaultConfigFromCM)
	convey.Convey("test FaultConfigurator method 'loadFaultConfigFromFile'", t, testLoadFaultConfigFromFile)
	convey.Convey("test FaultConfigurator method 'filterAndCheckFaultCodes'", t, testFilterAndCheckFaultCodes)
	convey.Convey("test FaultConfigurator method 'getFaultConfigFromCM'", t, testGetFaultConfigFromCM)
}

func testFaultCfgStop() {
	if configManager == nil {
		panic("configManager is nil")
	}
	configManager.Stop()
	convey.So(<-configManager.stopChan, convey.ShouldResemble, struct{}{})
}

func testFaultCfgInit() {
	if configManager == nil {
		panic("configManager is nil")
	}

	convey.Convey("test method Init success", func() {
		var p1 = gomonkey.ApplyPrivateMethod(&FaultConfigurator{}, "initFaultConfigFromCM",
			func() error {
				return nil
			})
		defer p1.Reset()
		err := configManager.Init()
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("test method Init success, initFaultConfigFromCM error", func() {
		var p1 = gomonkey.ApplyPrivateMethod(&FaultConfigurator{}, "initFaultConfigFromCM",
			func() error {
				return testErr
			}).ApplyPrivateMethod(&FaultConfigurator{}, "loadFaultConfigFromFile",
			func() error {
				return nil
			})
		defer p1.Reset()
		err := configManager.Init()
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("test method Init failed, loadFaultConfigFromFile error", func() {
		var p1 = gomonkey.ApplyPrivateMethod(&FaultConfigurator{}, "initFaultConfigFromCM",
			func() error {
				return testErr
			}).ApplyPrivateMethod(&FaultConfigurator{}, "loadFaultConfigFromFile",
			func() error {
				return testErr
			})
		defer p1.Reset()
		err := configManager.Init()
		convey.So(err, convey.ShouldResemble, testErr)
	})
}

func testInitFaultConfigFromCM() {
	if configManager == nil {
		panic("configManager is nil")
	}

	convey.Convey("test method initFaultConfigFromCM success", func() {
		var p1 = gomonkey.ApplyMethodReturn(&kubeclient.ClientK8s{}, "GetConfigMap", fakeNodeFaultConfigCM(), nil)
		defer p1.Reset()
		err := configManager.initFaultConfigFromCM()
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("test method initFaultConfigFromCM failed, get cm error", func() {
		var p2 = gomonkey.ApplyMethodReturn(&kubeclient.ClientK8s{}, "GetConfigMap", nil, testErr)
		defer p2.Reset()
		err := configManager.initFaultConfigFromCM()
		convey.So(err, convey.ShouldResemble, testErr)
	})

	convey.Convey("test method initFaultConfigFromCM failed, UpdateConfigCache error", func() {
		var p3 = gomonkey.ApplyMethodReturn(&kubeclient.ClientK8s{}, "GetConfigMap", fakeNodeFaultConfigCM(), nil).
			ApplyPrivateMethod(&FaultConfigurator{}, "getFaultConfigFromCM",
				func(cm *v1.ConfigMap) (*common.FaultConfig, error) {
					return nil, testErr
				})
		defer p3.Reset()
		err := configManager.initFaultConfigFromCM()
		convey.So(err, convey.ShouldResemble, testErr)
	})
}

func testLoadFaultConfigFromFile() {
	data, err := json.Marshal(testFaultConfig)
	if err != nil {
		convey.ShouldBeNil(err)
	}
	errData, err := json.Marshal(testErrFaultConfig)
	if err != nil {
		convey.ShouldBeNil(err)
	}

	convey.Convey("test method loadFaultConfigFromFile success", func() {
		var p1 = gomonkey.ApplyFuncReturn(utils.LoadFile, data, nil)
		defer p1.Reset()
		err = configManager.loadFaultConfigFromFile()
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("test method loadFaultConfigFromFile failed, load file error", func() {
		var p2 = gomonkey.ApplyFuncReturn(utils.LoadFile, data, testErr)
		defer p2.Reset()
		err = configManager.loadFaultConfigFromFile()
		expErr := fmt.Errorf("load local fault config json file failed: %v", testErr)
		convey.So(err, convey.ShouldResemble, expErr)
	})

	convey.Convey("test method loadFaultConfigFromFile failed, unmarshal error", func() {
		var p3 = gomonkey.ApplyFuncReturn(json.Unmarshal, testErr)
		defer p3.Reset()
		err = configManager.loadFaultConfigFromFile()
		expErr := fmt.Errorf("unmarshal fault config byte failed: %v", testErr)
		convey.So(err, convey.ShouldResemble, expErr)
	})

	convey.Convey("test method loadFaultConfigFromFile failed, filterAndCheckFaultCodes error", func() {
		var p4 = gomonkey.ApplyFuncReturn(utils.LoadFile, fakeWrongNodeFaultConfigCMData(), nil)
		defer p4.Reset()
		err = configManager.loadFaultConfigFromFile()
		convey.So(err.Error(), convey.ShouldContainSubstring, "contains illegal character")
	})

	convey.Convey("test method loadFaultConfigFromFile failed, filterAndCheckFaultCodes error, "+
		"fault config is nil", func() {
		var p4 = gomonkey.ApplyFuncReturn(utils.LoadFile, errData, nil)
		defer p4.Reset()
		err = configManager.loadFaultConfigFromFile()
		expErr := errors.New("fault config is nil")
		convey.So(err, convey.ShouldResemble, expErr)
	})
}

func testFilterAndCheckFaultCodes() {
	if configManager == nil {
		panic("configManager is nil")
	}

	// error case
	oldFaultConfig := testFaultConfig
	wrongFaultConfig1 := &common.FaultConfig{FaultTypeCode: &common.FaultTypeCode{}}
	wrongFaultConfig2 := &common.FaultConfig{FaultTypeCode: &common.FaultTypeCode{}}
	wrongFaultConfig3 := &common.FaultConfig{FaultTypeCode: &common.FaultTypeCode{}}
	common.DeepCopyFaultConfig(oldFaultConfig, wrongFaultConfig1)
	common.DeepCopyFaultConfig(oldFaultConfig, wrongFaultConfig2)
	common.DeepCopyFaultConfig(oldFaultConfig, wrongFaultConfig3)
	wrongFaultConfig1.FaultTypeCode.NotHandleFaultCodes = []string{"WrongFaultCodes"}
	wrongFaultConfig2.FaultTypeCode.PreSeparateFaultCodes = []string{"WrongFaultCodes"}
	wrongFaultConfig3.FaultTypeCode.SeparateFaultCodes = []string{"WrongFaultCodes"}

	testCase := []*common.FaultConfig{wrongFaultConfig1, wrongFaultConfig2, wrongFaultConfig3}
	for _, wrongFaultConfig := range testCase {
		err := configManager.filterAndCheckFaultCodes(wrongFaultConfig)
		convey.So(err.Error(), convey.ShouldContainSubstring, "contains illegal character")
	}
}

func testGetFaultConfigFromCM() {
	if configManager == nil {
		panic("configManager is nil")
	}

	convey.Convey("test method getFaultConfigFromCM success", func() {
		_, err := configManager.getFaultConfigFromCM(fakeNodeFaultConfigCM())
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("test method getFaultConfigFromCM failed, cm data error", func() {
		_, err := configManager.getFaultConfigFromCM(&v1.ConfigMap{})
		expErr := fmt.Errorf("can not find the key '%s' in cm", common.FaultConfigKey)
		convey.So(err, convey.ShouldResemble, expErr)
	})

	convey.Convey("test method getFaultConfigFromCM failed, unmarshal error", func() {
		var p1 = gomonkey.ApplyFuncReturn(json.Unmarshal, testErr)
		defer p1.Reset()
		_, err := configManager.getFaultConfigFromCM(fakeNodeFaultConfigCM())
		expErr := fmt.Errorf("unmarshal fault config failed: %v", testErr)
		convey.So(err, convey.ShouldResemble, expErr)
	})

	convey.Convey("test method getFaultConfigFromCM failed, filterAndCheckFaultCodes error", func() {
		_, err := configManager.getFaultConfigFromCM(fakeWrongNodeFaultConfigCM())
		convey.So(err.Error(), convey.ShouldContainSubstring, "contains illegal character")
	})
}

func testAddConfigCM() {
	if configManager == nil {
		panic("configManager is nil")
	}

	convey.Convey("test method AddConfigCM, initFromCMFlag is true", func() {
		configManager.initFromCMFlag = true
		configManager.AddConfigCM(fakeNodeFaultConfigCM())
		convey.So(configManager.initFromCMFlag, convey.ShouldBeFalse)
	})

	convey.Convey("test method AddConfigCM, initFromCMFlag is false", func() {
		configManager.initFromCMFlag = false
		configManager.AddConfigCM(fakeNodeFaultConfigCM())
		convey.So(<-common.GetTrigger(), convey.ShouldResemble, common.ConfigProcess)
	})

	convey.Convey("test method AddConfigCM, input type error", func() {
		configManager.AddConfigCM(nil)
	})

	convey.Convey("test method AddConfigCM, cm data error", func() {
		configManager.initFromCMFlag = false
		configManager.AddConfigCM(fakeWrongNodeFaultConfigCM())
	})
}

func testUpdateConfigCM() {
	if configManager == nil {
		panic("configManager is nil")
	}

	convey.Convey("test method UpdateConfigCM", func() {
		configManager.UpdateConfigCM(nil, fakeNodeFaultConfigCM())
		convey.So(configManager.initFromCMFlag, convey.ShouldBeFalse)
		convey.So(<-common.GetTrigger(), convey.ShouldResemble, common.ConfigProcess)
	})

	convey.Convey("test method UpdateConfigCM, input type error", func() {
		configManager.UpdateConfigCM(nil, nil)
	})

	convey.Convey("test method UpdateConfigCM, cm data error", func() {
		configManager.initFromCMFlag = false
		configManager.UpdateConfigCM(nil, fakeWrongNodeFaultConfigCM())
	})
}

func testDeleteConfigCM() {
	if configManager == nil {
		panic("configManager is nil")
	}
	configManager.DeleteConfigCM(nil)
}

func testGetMonitorData() {
	if configManager == nil {
		panic("configManager is nil")
	}
	faultConfig := &common.FaultConfig{FaultTypeCode: &common.FaultTypeCode{
		NotHandleFaultCodes:   []string{"00000001"},
		PreSeparateFaultCodes: []string{"00000002"},
		SeparateFaultCodes:    []string{"00000003"},
	}}

	configManager.configManager.SetFaultConfig(faultConfig)
	fcInfo := configManager.GetMonitorData()
	convey.So(fcInfo.FaultConfig, convey.ShouldResemble, faultConfig)
}
