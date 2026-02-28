package manualfault

import (
	"sort"
	"strings"
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/conf"
	"clusterd/pkg/domain/faultdomain"
	"clusterd/pkg/domain/manualfault"
	pod2 "clusterd/pkg/domain/pod"
)

const (
	defaultFaultWindowHours = 24
	defaultFaultThreshold   = 3
	defaultFaultFreeHours   = 48
)

var (
	validPolicy = conf.ManuallySeparatePolicy{
		Enabled: true,
		Separate: struct {
			FaultWindowHours int `yaml:"fault_window_hours"`
			FaultThreshold   int `yaml:"fault_threshold"`
		}{
			FaultWindowHours: defaultFaultWindowHours,
			FaultThreshold:   defaultFaultThreshold,
		},
		Release: struct {
			FaultFreeHours int `yaml:"fault_free_hours"`
		}{
			FaultFreeHours: defaultFaultFreeHours,
		},
	}
)

func TestProcessor(t *testing.T) {
	preparePodStorage()
	convey.Convey("test func 'Process', manually separate npu enabled is false", t, testCloseManualSep)
	convey.Convey("test func 'Process', input type is invalid", t, testInputIsNil)
	convey.Convey("test func 'Process', load from manual cm", t, testLoadFromManualCm)
	convey.Convey("test func 'Process', when have job, update job fault mgr", t, testWhenHaveJob)
	convey.Convey("test func 'Process', when have increment fault", t, testIncrementFault)
	convey.Convey("test func 'Process', when fault is same", t, testSameFault)
}

func testCloseManualSep() {
	validPolicy1 := validPolicy
	validPolicy1.Enabled = false
	conf.SetManualSeparatePolicy(validPolicy1)
	resetRelatedCache()
	ori := constant.OneConfigmapContent[*constant.AdvanceDeviceFaultCm]{
		AllConfigmap:    faultdomain.GetAdvanceFaultCm[*constant.AdvanceDeviceFaultCm](oriDevInfo1),
		UpdateConfigmap: nil,
	}
	ManualFaultProcessor.Process(ori)
	convey.So(manualfault.FaultCmInfo.Len(), convey.ShouldEqual, 0)
}

func testInputIsNil() {
	conf.SetManualSeparatePolicy(validPolicy)
	resetRelatedCache()
	res := ManualFaultProcessor.Process(nil)
	convey.So(res, convey.ShouldBeNil)
}

func addManualInfo(advanceFaultCm map[string]*constant.AdvanceDeviceFaultCm) map[string]*constant.AdvanceDeviceFaultCm {
	var newAdvancedFaultCm = make(map[string]*constant.AdvanceDeviceFaultCm)
	for node, advancedCm := range advanceFaultCm {
		deviceFaultMap := make(map[string][]constant.DeviceFault)
		for devName, faults := range advancedCm.FaultDeviceList {
			var newFaults []constant.DeviceFault
			for _, fault := range faults {
				fault.FaultCode = strings.Replace(fault.FaultCode, " ", "", -1)
				codes := strings.Split(fault.FaultCode, ",")
				for _, code := range codes {
					if code == "" && fault.FaultLevel == constant.ManuallySeparateNPU {
						fault.FaultCode = constant.ManuallySeparateNPU
						faultTimeAndLevel := constant.FaultTimeAndLevel{
							FaultTime:  constant.UnknownFaultTime,
							FaultLevel: constant.ManuallySeparateNPU,
						}
						fault.FaultTimeAndLevelMap = map[string]constant.FaultTimeAndLevel{
							constant.ManuallySeparateNPU: faultTimeAndLevel,
						}
					}
				}
				newFaults = append(newFaults, fault)
			}
			deviceFaultMap[devName] = append(deviceFaultMap[devName], newFaults...)
		}
		advancedCm.FaultDeviceList = deviceFaultMap
		newAdvancedFaultCm[node] = advancedCm
	}
	return newAdvancedFaultCm
}

func testLoadFromManualCm() {
	conf.SetManualSeparatePolicy(validPolicy)
	resetRelatedCache()
	nodeInfo := getDemoNodeInfo()
	manualfault.FaultCmInfo.SetNodeInfo(nodeInfo)

	content := constant.OneConfigmapContent[*constant.AdvanceDeviceFaultCm]{
		AllConfigmap:    faultdomain.GetAdvanceFaultCm[*constant.AdvanceDeviceFaultCm](oriDevInfo1),
		UpdateConfigmap: nil,
	}
	resContent := ManualFaultProcessor.Process(content).(constant.OneConfigmapContent[*constant.AdvanceDeviceFaultCm])
	sortDeviceFaultList(resContent.AllConfigmap)
	res := addManualInfo(resContent.AllConfigmap)
	want := faultdomain.GetAdvanceFaultCm[*constant.AdvanceDeviceFaultCm](expDeviceInfo1)
	sortDeviceFaultList(want)

	convey.So(res, convey.ShouldResemble, want)
}

func testWhenHaveJob() {
	conf.SetManualSeparatePolicy(validPolicy)
	resetRelatedCache()
	pod2.SavePod(podDemo1)
	pod2.SavePod(podDemo2)
	defer pod2.DeletePod(podDemo1)
	defer pod2.DeletePod(podDemo2)

	content := constant.OneConfigmapContent[*constant.AdvanceDeviceFaultCm]{
		AllConfigmap:    faultdomain.GetAdvanceFaultCm[*constant.AdvanceDeviceFaultCm](oriDevInfo1),
		UpdateConfigmap: nil,
	}
	_ = ManualFaultProcessor.Process(content).(constant.OneConfigmapContent[*constant.AdvanceDeviceFaultCm])
}

func testIncrementFault() {
	conf.SetManualSeparatePolicy(validPolicy)
	resetRelatedCache()
	pod2.SavePod(podDemo1)
	pod2.SavePod(podDemo3)
	pod2.SavePod(podDemo4)
	defer func() {
		pod2.DeletePod(podDemo1)
		pod2.DeletePod(podDemo3)
		pod2.DeletePod(podDemo4)
	}()

	// node1 Ascend910-5, Ascend910-6 and node2 Ascend910-2 have fault
	content := constant.OneConfigmapContent[*constant.AdvanceDeviceFaultCm]{
		AllConfigmap:    faultdomain.GetAdvanceFaultCm[*constant.AdvanceDeviceFaultCm](oriDevInfo2),
		UpdateConfigmap: nil,
	}
	ManualFaultProcessor.nodeDeviceCmMap = faultdomain.GetAdvanceFaultCm[*constant.AdvanceDeviceFaultCm](oriDevInfo1)
	_ = ManualFaultProcessor.Process(content).(constant.OneConfigmapContent[*constant.AdvanceDeviceFaultCm])
}

func testSameFault() {
	conf.SetManualSeparatePolicy(validPolicy)
	resetRelatedCache()
	pod2.SavePod(podDemo1)
	pod2.SavePod(podDemo3)
	defer func() {
		pod2.DeletePod(podDemo1)
		pod2.DeletePod(podDemo3)
	}()
	content := constant.OneConfigmapContent[*constant.AdvanceDeviceFaultCm]{
		AllConfigmap:    faultdomain.GetAdvanceFaultCm[*constant.AdvanceDeviceFaultCm](oriDevInfo1),
		UpdateConfigmap: nil,
	}
	ManualFaultProcessor.nodeDeviceCmMap = faultdomain.GetAdvanceFaultCm[*constant.AdvanceDeviceFaultCm](oriDevInfo1)
	_ = ManualFaultProcessor.Process(content).(constant.OneConfigmapContent[*constant.AdvanceDeviceFaultCm])
}

func sortDeviceFaultList(advanceFaultCm map[string]*constant.AdvanceDeviceFaultCm) {
	for _, advanceDeviceCm := range advanceFaultCm {
		for _, fault := range advanceDeviceCm.FaultDeviceList {
			sort.Slice(fault, func(i, j int) bool {
				return util.MakeDataHash(fault[i]) < util.MakeDataHash(fault[j])
			})
		}
		sort.Strings(advanceDeviceCm.CardUnHealthy)
		sort.Strings(advanceDeviceCm.NetworkUnhealthy)
		sort.Strings(advanceDeviceCm.Recovering)
		sort.Strings(advanceDeviceCm.AvailableDeviceList)
	}
}

func resetRelatedCache() {
	manualfault.InitJobFaultManager(constant.DefaultSlidingWindow)
	manualfault.InitCounter()
	manualfault.InitFaultCmInfo()
	ManualFaultProcessor.nodeDeviceCmMap = make(map[string]*constant.AdvanceDeviceFaultCm)
}

func preparePodStorage() {
	podDemo1 = getDemoPod(node1, podName1, dev0, job1)
	podDemo2 = getDemoPod(node1, podName2, dev5, job2)
	podDemo3 = getDemoPod(node1, podName3, dev6, job1)
	podDemo4 = getDemoPod(node2, podName4, dev2, job1)
}

func TestIsContainFault(t *testing.T) {
	oldFaults := []constant.DeviceFault{
		{
			FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
				code1: {
					FaultLevel:        constant.NotHandleFault,
					FaultReceivedTime: receiveTime1,
				},
				code0: {
					FaultLevel:        constant.NotHandleFault,
					FaultReceivedTime: receiveTime1,
				},
			},
		},
	}
	convey.Convey("test func 'isContainFault', contain new fault", t, func() {
		res := isContainFault(getNewFault1(), oldFaults)
		convey.So(res, convey.ShouldBeTrue)
	})
	convey.Convey("test func 'isContainFault', not contain new fault", t, func() {
		// fault received time is different
		res := isContainFault(getNewFault2(), oldFaults)
		convey.So(res, convey.ShouldBeFalse)

		// fault level is different
		res = isContainFault(getNewFault3(), oldFaults)
		convey.So(res, convey.ShouldBeFalse)

		// not contain fault code
		res = isContainFault(getNewFault4(), oldFaults)
		convey.So(res, convey.ShouldBeFalse)
	})
}

func getNewFault1() constant.DeviceFault {
	return constant.DeviceFault{
		FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
			code1: {
				FaultLevel:        constant.NotHandleFault,
				FaultReceivedTime: receiveTime1,
			},
		},
	}
}

func getNewFault2() constant.DeviceFault {
	return constant.DeviceFault{
		FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
			code1: {
				FaultLevel:        constant.NotHandleFault,
				FaultReceivedTime: receiveTime1 + 1,
			},
		},
	}
}

func getNewFault3() constant.DeviceFault {
	return constant.DeviceFault{
		FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
			code1: {
				FaultLevel:        constant.RestartRequest,
				FaultReceivedTime: receiveTime1,
			},
		},
	}
}

func getNewFault4() constant.DeviceFault {
	return constant.DeviceFault{
		FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
			code2: {
				FaultLevel:        constant.NotHandleFault,
				FaultReceivedTime: receiveTime1,
			},
		},
	}
}
