// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package job a series of job function
package job

import (
	"sync"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"volcano.sh/apis/pkg/client/clientset/versioned"
)

const (
	dryRun           = false
	displayStatistic = false
	cmCheckInterval  = 2
	cmCheckTimeout   = 10

	// EventAdd event type add
	EventAdd = "add"
	// EventUpdate event type updategi
	EventUpdate = "update"
	// EventDelete event type delete
	EventDelete = "delete"
	// ConfigmapPrefix is the job summary prefix
	ConfigmapPrefix = "job-summary"
	// ConfigmapLabel is the label for job summary configmap
	ConfigmapLabel = "outside-job-info"
	// ConfigmapWholeLabel is the label for job summary configmap
	ConfigmapWholeLabel = "outside-job-info=true"
	// ConfigmapKey is the configmap key
	ConfigmapKey = "hccl.json"
	// JobName is the job name
	JobName = "job_name"
	// ConfigmapOperator operator key
	ConfigmapOperator = "operator"
	// OperatorAdd add
	OperatorAdd = "add"
	// OperatorDelete delete
	OperatorDelete = "delete"
	// DataValue init value
	DataValue = `{"status":"initializing"}`
	// ConfigmapCompleted completed
	ConfigmapCompleted = "complete"
	// ConfigmapInitializing initializing
	ConfigmapInitializing = "initializing"
	cmDataMaxMemory       = 1024 * 1024
	// JobId job id
	JobId = "job_id"
	// DeleteTime is the time constant for deleting operator
	DeleteTime = "deleteTime"
	// AddTime is the time constant for adding operator
	AddTime = "time"
	// FrameWork is the key for framework
	FrameWork = "framework"
	// deleteCMInterval make sure at least the specific time passed to delete job summary cm
	deleteCMInterval = 300
	// deleteCMCyclicTime is the cyclic time to delete job summary cm
	deleteCMCyclicTime = 60

	// Key910 910key
	Key910 = "ring-controller.atlas"
	// Val910 value
	Val910 = "ascend-910"
	// Val910B 910b
	Val910B = "ascend-910b"
	// A910ResourceName 910 resource name
	A910ResourceName = "huawei.com/Ascend910"
	// PodDeviceKey device key
	PodDeviceKey = "ascend.kubectl.kubernetes.io/ascend-910-configuration"
	// PodRankIndexKey rank index key
	PodRankIndexKey = "hccl/rankIndex"
	// A800MaxChipNum max number
	A800MaxChipNum = 16
	// InvalidNPUNum invalid number
	InvalidNPUNum = -1
	// BuildStatInterval interval
	BuildStatInterval = 30 * time.Second
	// BitSize32 bit size
	BitSize32 = 32
	// Decimal 10
	Decimal = 10
	// maxRankIndex max rank index
	maxRankIndex = 10000
	// PodLabelKey is the key for pod frame label
	PodLabelKey = "app"
	// defaultResyncTime is the default sync time
	defaultResyncTime = 30
	// JobStatus is the default sync time
	JobStatus = "job_status"
	// StatusJobRunning is the running job status
	StatusJobRunning = "running"
	// StatusJobPending is the pending job status
	StatusJobPending = "pending"
	// StatusJobFail is the failed job status
	StatusJobFail = "failed"
	// StatusJobSucceed is the succeed job status
	StatusJobSucceed = "complete"
	// StatusJobDelete is used when we delete pod
	StatusJobDelete = "delete"
	// PhaseJobRunning is k8s running status
	PhaseJobRunning = "Running"
	// PhaseJobPending is k8s pending status
	PhaseJobPending = "Pending"
	// PhaseJobSucceed is k8s success status
	PhaseJobSucceed = "Succeeded"
	// PhaseJobFail is k8s failed status
	PhaseJobFail        = "Failed"
	cmCutNumKey         = "total"
	cmIndex             = "cm_index"
	deviceNumThresholds = 8000
	torTag              = "isSharedTor"
	sharedTor           = "1"
	torIpTag            = "sharedTorIp"
	vcJobKind           = "Job"
	acJobMasterSuffix   = "-master-0"
	masterAddrKey       = "masterAddr"
	ptFramework         = "pytorch"
)

// Config controller init configure
type Config struct {
	DryRun           bool
	DisplayStatistic bool
	CmCheckInterval  int
	CmCheckTimeout   int
}

// Agent for all businessWorkers
type Agent struct {
	Config        *Config
	BsWorker      map[string]PodWorker
	podsInformer  cache.SharedIndexInformer
	podsIndexer   cache.Indexer
	KubeClientSet kubernetes.Interface
	RwMutex       sync.RWMutex
	vcClient      *versioned.Clientset
}

type jobModel struct {
	key string
	Info
	replicas int32
	devices  *v1.ResourceList
}

// Worker : controller for each job, list/watch corresponding pods and build configmap rank table
type Worker struct {
	WorkerInfo
	Info
}

// Info : Job Worker Info
type Info struct {
	Namespace         string
	JobName           string
	PGName            string
	JobUid            string
	PGUid             string
	PGLabels          map[string]string
	Key               string
	Version           int32
	JobType           string
	CreationTimestamp metav1.Time
}

// WorkerInfo : normal Worker info
type WorkerInfo struct {
	clientSet         kubernetes.Interface
	vcClient          *versioned.Clientset
	JobType           string
	CmMutex, statMu   sync.Mutex
	dryRun            bool
	statSwitch        chan struct{}
	podIndexer        cache.Indexer
	CMName            string
	CMData            RankTabler
	statStopped       bool
	rankIndex         int
	cachedPodNum      int32
	cachePodMap       map[string]*v1.Pod
	jobReplicasTotal  int32
	succeedPodNum     int32
	podSchedulerCache []string
	SharedTorIp       []string
}

type podIdentifier struct {
	namespace string
	name      string
	jobName   string
	eventType string
	jobId     string
	UID       string
}

// RankTable to hccl
type RankTable struct {
	RankTableStatus
	ServerList      []*ServerHccl       `json:"server_list"`
	ServerCount     string              `json:"server_count"`
	Version         string              `json:"version"`
	Total           int                 `json:"total"`
	UnHealthyNode   map[string][]string `json:"-"`
	UnHealthyDevice map[string]string   `json:"-"`
}

// ServerHccl to hccl
type ServerHccl struct {
	DeviceList []*Device `json:"device"`
	ServerID   string    `json:"server_id"`
	PodID      string    `json:"-"`
	ServerName string    `json:"server_name"`
}

// Device to hccl with rankId
type Device struct {
	DeviceID string `json:"device_id"`
	DeviceIP string `json:"device_ip"`
	RankID   string `json:"rank_id"` // rank id
}

// RankTableStatus to hccl
type RankTableStatus struct {
	Status string `json:"status"`
}

// Instance to hccl
type Instance struct {
	Devices  []Device `json:"devices"`
	PodName  string   `json:"pod_name"`
	ServerID string   `json:"server_id"`
}

var (
	// ModelFramework is the framework value
	ModelFramework string
)

// JobServerInfoMap to store job server info
type JobServerInfoMap struct {
	InfoMap     map[string]map[string]ServerHccl
	UceTolerate map[string]bool
}
