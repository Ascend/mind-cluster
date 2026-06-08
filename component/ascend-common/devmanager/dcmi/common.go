/* Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
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

package dcmi

// #include "dcmi_interface_api.h"
import "C"
import (
	"fmt"
	"unsafe"

	"ascend-common/devmanager/common"
)

// FaultEventCallFunc dcmi v1&v2 share a FaultEventCallFunc
var FaultEventCallFunc func(common.DevFaultInfo) = nil
var (
	dcmiErrMap = map[int32]string{
		-8001:  "The input parameter is incorrect",
		-8002:  "Permission error",
		-8003:  "The memory interface operation failed",
		-8004:  "The security function failed to be executed",
		-8005:  "Internal errors",
		-8006:  "Response timed out",
		-8007:  "Invalid deviceID",
		-8008:  "The device does not exist",
		-8009:  "ioctl returns failed",
		-8010:  "The message failed to be sent",
		-8011:  "Message reception failed",
		-8012:  "Not ready yet,please try again",
		-8013:  "This API is not supported in containers",
		-8014:  "The file operation failed",
		-8015:  "Reset failed",
		-8016:  "Reset cancels",
		-8017:  "Upgrading",
		-8020:  "Device resources are occupied",
		-8022:  "Partition consistency check,inconsistent partitions were found",
		-8023:  "The configuration information does not exist",
		-8255:  "Device ID/function is not supported",
		-99997: "dcmi shutdown failed",
		-99998: "The called function is missing,please upgrade the driver",
		-99999: "dcmi libdcmi.so failed to load",
	}
)

func convertUCharToCharArr(cgoArr [MaxChipNameLen]C.uchar) []byte {
	var charArr []byte
	for _, v := range cgoArr {
		if v == 0 {
			break
		}
		charArr = append(charArr, byte(v))
	}
	return charArr
}

func convertUrmaDeviceInfo(eidInfoListPtr *C.dcmi_urma_eid_info_t, eidCnt C.uint) (*common.UrmaDeviceInfo, error) {
	if eidInfoListPtr == nil {
		return nil, fmt.Errorf("input parameter eidInfoListPtr is nil")
	}

	eidCount := uint(eidCnt)
	if eidCount > common.EidNumMax {
		return nil, fmt.Errorf("urma device count is %d out of range [0, %d]", eidCount, common.EidNumMax)
	}

	eidInfoList := (*[common.EidNumMax]C.dcmi_urma_eid_info_t)(unsafe.Pointer(eidInfoListPtr))
	urmaDevInfo := common.UrmaDeviceInfo{EidCount: eidCount, EidInfos: make([]common.UrmaEidInfo, eidCount)}
	for j := 0; j < int(eidCount); j++ {
		eidInfo := common.UrmaEidInfo{
			Eid: common.Eid{
				Raw: eidInfoList[j].eid,
			},
			EidIndex: uint(eidInfoList[j].eid_index),
		}
		urmaDevInfo.EidInfos[j] = eidInfo
	}

	return &urmaDevInfo, nil
}

func buildUbPingMeshCArray(ops []common.UBPingMeshOperate, cOpsPtr *C.struct_dcmi_ub_ping_mesh_operate, size int) (
	*C.struct_dcmi_ub_ping_mesh_operate, error) {

	cOps := (*[maxCArraySize]C.struct_dcmi_ub_ping_mesh_operate)(unsafe.Pointer(cOpsPtr))[:size:size]

	for idx, op := range ops {
		for i := 0; i < common.EidByteSize; i++ {
			cOps[idx].src_eid[i] = C.char(op.SrcEID.Raw[i])
		}
		for i := 0; i < len(op.DstEIDList); i++ {
			for j := 0; j < common.EidByteSize; j++ {
				cOps[idx].dst_eid_list[i][j] = C.char(op.DstEIDList[i].Raw[j])
			}
		}
		cOps[idx].dst_num = C.int(op.DstNum)
		cOps[idx].pkt_size = C.int(op.PktSize)
		cOps[idx].pkt_send_num = C.int(op.PktSendNum)
		cOps[idx].pkt_interval = C.int(op.PktInterval)
		cOps[idx].timeout = C.int(op.Timeout)
		cOps[idx].task_interval = C.int(op.TaskInterval)
		cOps[idx].task_id = C.int(op.TaskID)
	}

	return cOpsPtr, nil
}

// fillStatsFromCInfo for Ascend950
func fillStatsFromCInfo(info *common.UBPingMeshInfo, cInfo *C.struct_dcmi_ub_ping_mesh_info) {
	info.SucPktNum = make([]uint, common.UbPingMeshMaxNum)
	info.FailPktNum = make([]uint, common.UbPingMeshMaxNum)
	info.MaxTime = make([]int, common.UbPingMeshMaxNum)
	info.MinTime = make([]int, common.UbPingMeshMaxNum)
	info.AvgTime = make([]int, common.UbPingMeshMaxNum)
	info.Tp95Time = make([]int, common.UbPingMeshMaxNum)
	info.ReplyStatNum = make([]int, common.UbPingMeshMaxNum)
	info.PingTotalNum = make([]int, common.UbPingMeshMaxNum)

	for k := 0; k < common.UbPingMeshMaxNum; k++ {
		info.SucPktNum[k] = uint(cInfo.suc_pkt_num[k])
		info.FailPktNum[k] = uint(cInfo.fail_pkt_num[k])
		info.MaxTime[k] = int(cInfo.max_time[k])
		info.MinTime[k] = int(cInfo.min_time[k])
		info.AvgTime[k] = int(cInfo.avg_time[k])
		info.Tp95Time[k] = int(cInfo.tp95_time[k])
		info.ReplyStatNum[k] = int(cInfo.reply_stat_num[k])
		info.PingTotalNum[k] = int(cInfo.ping_total_num[k])
	}
	info.OccurTime = uint(cInfo.occur_time)
}

func convertBaseResource(cBaseResource C.struct_dcmi_base_resource) common.CgoBaseResource {
	baseResource := common.CgoBaseResource{
		Token:       uint64(cBaseResource.token),
		TokenMax:    uint64(cBaseResource.token_max),
		TaskTimeout: uint64(cBaseResource.task_timeout),
		VfgID:       uint32(cBaseResource.vfg_id),
		VipMode:     uint8(cBaseResource.vip_mode),
	}
	return baseResource
}

func convertComputingResource(cComputingResource C.struct_dcmi_computing_resource) common.CgoComputingResource {
	computingResource := common.CgoComputingResource{
		Aic:                float32(cComputingResource.aic),
		Aiv:                float32(cComputingResource.aiv),
		Dsa:                uint16(cComputingResource.dsa),
		Rtsq:               uint16(cComputingResource.rtsq),
		Acsq:               uint16(cComputingResource.acsq),
		Cdqm:               uint16(cComputingResource.cdqm),
		CCore:              uint16(cComputingResource.c_core),
		Ffts:               uint16(cComputingResource.ffts),
		Sdma:               uint16(cComputingResource.sdma),
		PcieDma:            uint16(cComputingResource.pcie_dma),
		MemorySize:         uint64(cComputingResource.memory_size),
		EventID:            uint32(cComputingResource.event_id),
		NotifyID:           uint32(cComputingResource.notify_id),
		StreamID:           uint32(cComputingResource.stream_id),
		ModelID:            uint32(cComputingResource.model_id),
		TopicScheduleAicpu: uint16(cComputingResource.topic_schedule_aicpu),
		HostCtrlCPU:        uint16(cComputingResource.host_ctrl_cpu),
		HostAicpu:          uint16(cComputingResource.host_aicpu),
		DeviceAicpu:        uint16(cComputingResource.device_aicpu),
		TopicCtrlCPUSlot:   uint16(cComputingResource.topic_ctrl_cpu_slot),
	}
	return computingResource
}

func convertMediaResource(cMediaResource C.struct_dcmi_media_resource) common.CgoMediaResource {
	mediaResource := common.CgoMediaResource{
		Jpegd: float32(cMediaResource.jpegd),
		Jpege: float32(cMediaResource.jpege),
		Vpc:   float32(cMediaResource.vpc),
		Vdec:  float32(cMediaResource.vdec),
		Pngd:  float32(cMediaResource.pngd),
		Venc:  float32(cMediaResource.venc),
	}
	return mediaResource
}

func convertSocTotalResource(cSocTotalResource C.struct_dcmi_soc_total_resource) common.CgoSocTotalResource {
	socTotalResource := common.CgoSocTotalResource{
		VDevNum:   uint32(cSocTotalResource.vdev_num),
		VfgNum:    uint32(cSocTotalResource.vfg_num),
		VfgBitmap: uint32(cSocTotalResource.vfg_bitmap),
		Base:      convertBaseResource(cSocTotalResource.base),
		Computing: convertComputingResource(cSocTotalResource.computing),
		Media:     convertMediaResource(cSocTotalResource.media),
	}
	for i := uint32(0); i < uint32(cSocTotalResource.vdev_num) && i < DcmiMaxVdevNum; i++ {
		socTotalResource.VDevID = append(socTotalResource.VDevID, uint32(cSocTotalResource.vdev_id[i]))
	}
	return socTotalResource
}

func convertSuperPodInfo(cSuperPodInfo C.struct_dcmi_spod_info) common.CgoSuperPodInfo {
	superPodInfo := common.CgoSuperPodInfo{
		SdId:         uint32(cSuperPodInfo.sdid),
		ScaleType:    uint32(cSuperPodInfo.scale_type),
		SuperPodId:   uint32(cSuperPodInfo.super_pod_id),
		ServerId:     uint32(cSuperPodInfo.server_id),
		RackId:       uint32(cSuperPodInfo.chassis_id),
		SuperPodType: uint32(cSuperPodInfo.super_pod_type),
	}

	for i := uint32(0); i < DcmiSpodReserveLen; i++ {
		superPodInfo.Reserve = append(superPodInfo.Reserve, uint32(cSuperPodInfo.reserve[i]))
	}

	return superPodInfo
}

func convertSocFreeResource(cSocFreeResource C.struct_dcmi_soc_free_resource) common.CgoSocFreeResource {
	socFreeResource := common.CgoSocFreeResource{
		VfgNum:    uint32(cSocFreeResource.vfg_num),
		VfgBitmap: uint32(cSocFreeResource.vfg_bitmap),
		Base:      convertBaseResource(cSocFreeResource.base),
		Computing: convertComputingResource(cSocFreeResource.computing),
		Media:     convertMediaResource(cSocFreeResource.media),
	}
	return socFreeResource
}

func convertSioInfoStruct(sPodSioInfo C.struct_dcmi_sio_crc_err_statistic_info) common.SioCrcErrStatisticInfo {
	cgoSPodSioInfo := common.SioCrcErrStatisticInfo{
		TxErrCnt: int64(sPodSioInfo.tx_error_count),
		RxErrCnt: int64(sPodSioInfo.rx_error_count),
	}
	for i := uint32(0); i < DcmiMaxReserveNum; i++ {
		cgoSPodSioInfo.Reserved = append(cgoSPodSioInfo.Reserved, uint32(sPodSioInfo.reserved[i]))
	}
	return cgoSPodSioInfo
}

// Define a safe function to get address offsets (for cleanCode)
func getAddrWithOffset(addr unsafe.Pointer, length, offset uintptr) (unsafe.Pointer, error) {
	if offset > length {
		return nil, fmt.Errorf("offset(%d) is invalid, length(%d)", offset, length)
	}
	return (unsafe.Pointer)(uintptr(addr) + offset), nil
}
