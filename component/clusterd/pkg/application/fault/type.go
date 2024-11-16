package fault

import (
	"clusterd/pkg/application/job"
	"clusterd/pkg/common/constant"
	"sync"
)

// The faultProcessor process the fault information.
type faultProcessor interface {
	process()
}

type baseFaultCenter struct {
	processorList        []faultProcessor
	lastProcessTime      int64
	subscribeChannelList []chan struct{}
	mutex                sync.Mutex
	processPeriod        int64
	jobServerInfoMap     job.JobServerInfoMap
}

// deviceFaultProcessCenter
type deviceFaultProcessCenter struct {
	baseFaultCenter
	mutex   sync.RWMutex
	infoMap map[string]*constant.DeviceInfo
}

// The FaultProcessCenter process the faults
var GlobalFaultProcessCenter *FaultProcessCenter

type FaultProcessCenter struct {
	deviceCenter      *deviceFaultProcessCenter
	nodeCenter        *nodeFaultProcessCenter
	switchCenter      *switchFaultProcessCenter
	notifyProcessChan chan int
}

// AdvanceDeviceCm more structure device info
type AdvanceDeviceCm struct {
	ServerType       string
	CmName           string
	SuperPodID       int32
	ServerIndex      int32
	DeviceList       map[string][]constant.DeviceFault
	CarUnHealthy     []string
	NetworkUnhealthy []string
	UpdateTime       int64
}

// FaultRank fault rank info
type FaultRank struct {
	RankId    string
	FaultCode string
}

// JobFaultInfo job fault rank info
type JobFaultInfo struct {
	JobId     string
	FaultList []FaultRank
}

type jobRankFaultInfoProcessor struct {
	deviceCenter    *deviceFaultProcessCenter
	jobFaultInfoMap map[string]JobFaultInfo
	mutex           sync.RWMutex
}

// nodeFaultProcessCenter
type nodeFaultProcessCenter struct {
	baseFaultCenter
	infoMap map[string]*constant.NodeInfo
	mutex   sync.RWMutex
}

type switchFaultProcessCenter struct {
	baseFaultCenter
	infoMap map[string]*constant.SwitchInfo
	mutex   sync.RWMutex
}

// uceAccompanyFaultProcessor:
// aic aiv fault can be 1) accompanied by uce fault, also can 2) curr alone.
// if 1) aic aiv fault should be filtered. Once find aic fault, check if there is an uce fault 5s ago
// if 2) aic aiv fault should not be retained.
type uceAccompanyFaultProcessor struct {
	deviceCenter *deviceFaultProcessCenter
	// maintain 5s ago device info
	DiagnosisAccompanyTimeout int64
	// nodeName -> deviceName -> faultQue
	uceAccompanyFaultQue map[string]map[string][]constant.DeviceFault
	// uceFaultTime
	uceFaultTime       map[string]map[string]int64
	deviceCmForNodeMap map[string]AdvanceDeviceCm
}

/*
The uceFaultProcessor process uce fault reporting information.
If the device fault is UCE fault, then determine whether the job running on the device can tolerate UCE faults.
If they can tolerate it, the reporting of the UCE fault should be delayed by 10 seconds.
*/
type uceFaultProcessor struct {
	deviceCenter             *deviceFaultProcessCenter
	JobReportRecoverTimeout  int64
	JobReportCompleteTimeout int64

	reportInfo *reportInfosForAllJobs
	// uceJob->jobInfo
	uceDevicesOfUceJob map[string]uceJobInfo
	// node->DeviceName->uceDeviceInfo
	uceDeviceOfNode  map[string]uceNodeInfo
	jobServerInfoMap job.JobServerInfoMap
	nodeDeviceCmMap  map[string]AdvanceDeviceCm
}

// JobId->node->device->report_info
type reportInfosForAllJobs struct {
	InfoMap map[string]map[string]map[string]reportInfo
	RwMutex sync.RWMutex
}

type uceDeviceInfo struct {
	// DeviceName has prefix Ascend910
	DeviceName   string
	FaultTime    int64
	RecoverTime  int64
	CompleteTime int64
}

type uceNodeInfo struct {
	NodeName string
	// DeviceName->DeviceInfo
	DeviceInfo map[string]uceDeviceInfo
}

type uceJobInfo struct {
	// UceNode node->nodeInfo
	JobId   string
	UceNode map[string]uceNodeInfo
}

type reportInfo struct {
	RecoverTime  int64
	CompleteTime int64
}
