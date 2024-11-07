/*
Copyright(C) 2024. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package common is common function or object of ranktable.
*/

package common

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"sync"

	"huawei.com/npu-exporter/v5/common-utils/hwlog"
	corev1 "k8s.io/api/core/v1"

	mindxdlv1 "ascend-operator/pkg/api/v1"
	"ascend-operator/pkg/ranktable/generator"
	"ascend-operator/pkg/ranktable/utils"
)

const (
	defaultPerm   = 0644
	rankTableFile = "hccl.json"
)

// BaseGenerator is the base struct for ranktable generator.
type BaseGenerator struct {
	dir  string
	path string

	servers    *sync.Map
	rankTabler generator.RankTableGenerator

	Status      utils.RankTableStatus `json:"status"`
	ServerList  []*Server             `json:"server_list" json:"server_list,omitempty"`   // hccl_json server list
	ServerCount string                `json:"server_count" json:"server_count,omitempty"` // hccl_json server count
	Version     string                `json:"version" json:"version,omitempty"`           // hccl_json version
}

// NewBaseGenerator is the constructor for BaseGenerator.
func NewBaseGenerator(job *mindxdlv1.AscendJob, version string, r generator.RankTableGenerator) *BaseGenerator {
	rankTableDir := utils.GenRankTableDir(job)
	return &BaseGenerator{
		dir:        rankTableDir,
		path:       rankTableDir + "/" + rankTableFile,
		servers:    &sync.Map{},
		rankTabler: r,
		Status:     utils.InitialRTStatus,
		ServerList: []*Server{},
		Version:    version,
	}
}

// SetStatus is used to set the status of ranktable.
func (r *BaseGenerator) SetStatus(status utils.RankTableStatus) {
	r.Status = status
}

// GetStatus is used to get the status of ranktable.
func (r *BaseGenerator) GetStatus() utils.RankTableStatus {
	return r.Status
}

// WriteToFile is used to write ranktable to file.
func (r *BaseGenerator) WriteToFile() error {
	hwlog.RunLog.Infof("start write info into file: %s", r.path)
	if err := func() error {
		f, err := os.OpenFile(r.path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, defaultPerm)
		if err != nil {
			return err
		}
		defer f.Close()
		rtStr, err := r.ToString()
		if err != nil {
			return err
		}
		if _, err = f.WriteString(rtStr); err != nil {
			return err
		}
		return nil
	}(); err != nil {
		return err
	}

	if err := os.Chmod(r.path, defaultPerm); err != nil {
		return err
	}
	hwlog.RunLog.Infof("write info into file: %s, and change mod to 666 success", r.path)
	return nil
}

// DeleteFile is used to delete ranktable file.
func (r *BaseGenerator) DeleteFile() error {
	hwlog.RunLog.Infof("delete file(%s)", r.path)
	rmErr := os.Remove(r.path)
	if rmErr != nil {
		return fmt.Errorf("failed to remove file(%s): %v", r.path, rmErr)
	}
	return nil
}

// ToString is used to get the string of ranktable.
func (r *BaseGenerator) ToString() (string, error) {
	b, err := json.Marshal(r.rankTabler)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// AddPod is used to add pod to ranktable.
func (r *BaseGenerator) AddPod(pod *corev1.Pod) error {
	deviceInfo := pod.Annotations[utils.PodDeviceKey]
	var instance Instance
	if err := json.Unmarshal([]byte(deviceInfo), &instance); err != nil {
		hwlog.RunLog.Errorf("unmarshal pod(%s/%s) deviceInfo(%s) failed: %v", pod.Namespace, pod.Name,
			deviceInfo, err)
		return err
	}

	if pod.Status.PodIP == "" {
		hwlog.RunLog.Errorf("pod(%s/%s) ip is empty", pod.Namespace, pod.Name)
		return fmt.Errorf("pod(%s/%s) ip is empty", pod.Namespace, pod.Name)
	}

	rankIndex, err := strconv.Atoi(pod.Annotations[utils.PodRankKey])
	if err != nil {
		hwlog.RunLog.Errorf("parse pod(%s/%s) rankIndex(%s) failed: %v", pod.Namespace, pod.Name,
			pod.Annotations[utils.PodRankKey], err)
		return err
	}
	hwlog.RunLog.Debugf("instance: %v", instance)
	server := &Server{
		ServerID:    instance.ServerID,
		ContainerIP: pod.Status.PodIP,
		DeviceList:  make([]*Device, 0),
	}
	rankFactor := len(instance.Devices)

	for index, device := range instance.Devices {
		var serverDevice Device
		serverDevice.DeviceID = device.DeviceID
		serverDevice.DeviceIP = device.DeviceIP
		if r.Version == "1.2" {
			serverDevice.SuperDeviceID = device.SuperDeviceID
		}
		serverDevice.RankID = strconv.Itoa(rankIndex*rankFactor + index)
		server.DeviceList = append(server.DeviceList, &serverDevice)
	}
	if len(server.DeviceList) < 1 {
		return fmt.Errorf("%s/%s get deviceList failed", pod.Namespace, pod.Name)
	}

	r.servers.Store(pod.UID, server)
	return nil
}

// DeletePod is used to delete pod from ranktable.
func (r *BaseGenerator) DeletePod(pod *corev1.Pod) error {
	r.servers.Delete(pod.UID)
	if r.GetStatus() == utils.InitialRTStatus {
		return nil
	}
	r.SetStatus(utils.InitialRTStatus)
	if err := r.WriteToFile(); err != nil {
		hwlog.RunLog.Errorf("failed to write ranktable to file, err: %v", err)
		r.SetStatus(utils.CompletedRTStatus)
	}
}

// GatherServerList is used to gather server list.
func (r *BaseGenerator) GatherServerList() {
	r.ServerList = make([]*Server, 0)
	r.servers.Range(func(key, value interface{}) bool {
		r.ServerList = append(r.ServerList, value.(*Server))
		return true
	})

	sort.Slice(r.ServerList, func(i, j int) bool {
		iRankID, iErr := strconv.Atoi(r.ServerList[i].DeviceList[0].RankID)
		jRankID, jErr := strconv.Atoi(r.ServerList[j].DeviceList[0].RankID)
		if iErr != nil || jErr != nil {
			return false
		}
		return iRankID < jRankID
	})
	r.ServerCount = strconv.Itoa(len(r.ServerList))
}
