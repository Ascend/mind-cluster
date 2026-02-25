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

// Package dcmiv2 this for dcmi v2 manager
package dcmiv2

// #cgo CFLAGS: -I${SRCDIR}/../dcmi
// #cgo LDFLAGS: -ldl
/*
    #include <stddef.h>
    #include <dlfcn.h>
    #include <stdlib.h>
    #include <stdio.h>
    #include "dcmi_interface_api.h"
    #include "dcmi_interface_api_v2.h"

    static void *dcmiHandle;
    #define SO_NOT_FOUND  -99999
    #define FUNCTION_NOT_FOUND  -99998
    #define SUCCESS  0
    #define ERROR_UNKNOWN  -99997
    #define CALL_FUNC(name,...) if(name##_func==NULL){return FUNCTION_NOT_FOUND;}return name##_func(__VA_ARGS__);

    static int (*dcmiv2_init_func)();
    static int dcmiv2_init_new(){
        CALL_FUNC(dcmiv2_init)
    }

    static int (*dcmiv2_get_device_info_func)(int dev_id, enum dcmi_main_cmd main_cmd,
        unsigned int sub_cmd,void *buf, unsigned int *size);
    static int dcmiv2_get_device_info(int dev_id, enum dcmi_main_cmd main_cmd, unsigned int sub_cmd, void *buf,
        unsigned int *size){
        CALL_FUNC(dcmiv2_get_device_info,dev_id,main_cmd,sub_cmd,buf,size)
    }

    static int (*dcmiv2_get_device_type_func)(int dev_id,enum dcmi_unit_type *device_type);
    static int dcmiv2_get_device_type(int dev_id,enum dcmi_unit_type *device_type){
        CALL_FUNC(dcmiv2_get_device_type,dev_id,device_type)
    }

    static int (*dcmiv2_get_device_health_func)(int dev_id, unsigned int *health);
    static int dcmiv2_get_device_health(int dev_id, unsigned int *health){
        CALL_FUNC(dcmiv2_get_device_health,dev_id,health)
   }

    static int (*dcmiv2_get_device_utilization_rate_func)(int dev_id, int input_type, unsigned int *utilization_rate);
    static int dcmiv2_get_device_utilization_rate(int dev_id, int input_type, unsigned int *utilization_rate){
        CALL_FUNC(dcmiv2_get_device_utilization_rate,dev_id,input_type,utilization_rate)
    }

    static int (*dcmiv2_get_device_temperature_func)(int dev_id, int *temperature);
    static int dcmiv2_get_device_temperature(int dev_id, int *temperature){
        CALL_FUNC(dcmiv2_get_device_temperature,dev_id,temperature)
    }

    static int (*dcmiv2_get_device_voltage_func)(int dev_id, unsigned int *voltage);
    static int dcmiv2_get_device_voltage(int dev_id, unsigned int *voltage){
        CALL_FUNC(dcmiv2_get_device_voltage,dev_id,voltage)
    }

    static int (*dcmiv2_get_device_power_info_func)(int dev_id, int *power);
    static int dcmiv2_get_device_power_info(int dev_id, int *power){
        CALL_FUNC(dcmiv2_get_device_power_info,dev_id,power)
    }

    static int (*dcmiv2_get_device_frequency_func)(int dev_id, enum dcmi_freq_type input_type,
        unsigned int *frequency);
    static int dcmiv2_get_device_frequency(int dev_id, enum dcmi_freq_type input_type, unsigned int *frequency){
        CALL_FUNC(dcmiv2_get_device_frequency,dev_id,input_type,frequency)
    }

    static int (*dcmiv2_get_device_hbm_info_func)(int dev_id, struct dcmi_hbm_info *hbm_info);
    static int dcmiv2_get_device_hbm_info(int dev_id, struct dcmi_hbm_info *hbm_info){
        CALL_FUNC(dcmiv2_get_device_hbm_info,dev_id,hbm_info)
    }

    static int (*dcmiv2_get_device_errorcode_func)(int dev_id, int *error_count,
        unsigned int *error_code_list, unsigned int list_len);
    static int dcmiv2_get_device_errorcode(int dev_id, int *error_count,
        unsigned int *error_code_list, unsigned int list_len){
        CALL_FUNC(dcmiv2_get_device_errorcode,dev_id,error_count,error_code_list,list_len)
    }

    static int (*dcmiv2_get_device_chip_info_func)(int dev_id, struct dcmi_chip_info_v2 *chip_info);
    static int dcmiv2_get_device_chip_info(int dev_id, struct dcmi_chip_info_v2 *chip_info){
        CALL_FUNC(dcmiv2_get_device_chip_info,dev_id,chip_info)
    }

    static int (*dcmiv2_get_chip_phyid_from_dev_id_func)(unsigned int dev_id, unsigned int *phyid);
    static int dcmiv2_get_chip_phyid_from_dev_id(unsigned int dev_id, unsigned int *phyid){
        CALL_FUNC(dcmiv2_get_chip_phyid_from_dev_id,dev_id,phyid)
    }

    static int (*dcmiv2_get_dev_id_from_chip_phyid_func)(unsigned int phyid, unsigned int *dev_id);
    static int dcmiv2_get_dev_id_from_chip_phyid(unsigned int phyid, unsigned int *dev_id){
        CALL_FUNC(dcmiv2_get_dev_id_from_chip_phyid,phyid,dev_id)
    }

    static int (*dcmiv2_get_device_ip_func)(int dev_id, enum dcmi_port_type input_type, int port_id,
        struct dcmi_ip_addr *ip, struct dcmi_ip_addr *mask);
    static int dcmiv2_get_device_ip(int dev_id, enum dcmi_port_type input_type, int port_id,
        struct dcmi_ip_addr *ip, struct dcmi_ip_addr *mask){
        CALL_FUNC(dcmiv2_get_device_ip,dev_id,input_type,port_id,ip,mask)
    }

    static int (*dcmiv2_get_device_network_health_func)(int dev_id, enum dcmi_rdfx_detect_result *result);
    static int dcmiv2_get_device_network_health(int dev_id, enum dcmi_rdfx_detect_result *result){
        CALL_FUNC(dcmiv2_get_device_network_health,dev_id,result)
    }

    static int (*dcmiv2_get_device_list_func)(int *device_num, int *device_list, int list_len);
    static int dcmiv2_get_device_list(int *device_num, int *device_list, int list_len){
        CALL_FUNC(dcmiv2_get_device_list,device_num,device_list,list_len)
    }

    static int (*dcmiv2_get_card_elabel_func)(int dev_id, struct dcmi_elabel_info *elabel_info);
    static int dcmiv2_get_card_elabel(int dev_id, struct dcmi_elabel_info *elabel_info){
        CALL_FUNC(dcmiv2_get_card_elabel,dev_id,elabel_info)
    }

    static int (*dcmiv2_set_device_reset_func)(int dev_id, enum dcmi_reset_channel channel_type);
    static int dcmiv2_set_device_reset(int dev_id, enum dcmi_reset_channel channel_type){
        CALL_FUNC(dcmiv2_set_device_reset,dev_id,channel_type)
    }

    static int (*dcmiv2_get_device_outband_channel_state_func)(int dev_id, int* channel_state);
    static int dcmiv2_get_device_outband_channel_state(int dev_id, int* channel_state){
        CALL_FUNC(dcmiv2_get_device_outband_channel_state,dev_id,channel_state)
    }

    static int (*dcmiv2_pre_reset_soc_func)(int dev_id);
    static int dcmiv2_pre_reset_soc(int dev_id){
        CALL_FUNC(dcmiv2_pre_reset_soc,dev_id)
    }

    static int (*dcmiv2_rescan_soc_func)(int dev_id);
    static int dcmiv2_rescan_soc(int dev_id){
        CALL_FUNC(dcmiv2_rescan_soc,dev_id)
    }

    static int (*dcmiv2_get_device_boot_status_func)(int dev_id, enum dcmi_boot_status *boot_status);
    static int dcmiv2_get_device_boot_status(int dev_id, enum dcmi_boot_status *boot_status){
        CALL_FUNC(dcmiv2_get_device_boot_status,dev_id,boot_status)
    }

    void goEventFaultCallBack(struct dcmi_dms_fault_event);
    static void event_handler(struct dcmi_event *fault_event) {
        goEventFaultCallBack(fault_event->event_t.dms_event);
    }

    static int (*dcmiv2_subscribe_fault_event_func)(int dev_id, struct dcmi_event_filter filter,
        void (*f_name)(struct dcmi_event *fault_event));
    static int dcmiv2_subscribe_fault_event(int dev_id, struct dcmi_event_filter filter){
        CALL_FUNC(dcmiv2_subscribe_fault_event,dev_id,filter,event_handler)
    }

    static int (*dcmiv2_get_device_die_func)(int dev_id, enum dcmi_die_type input_type,
        struct dcmi_die_id *die_id);
    static int dcmiv2_get_device_die(int dev_id, enum dcmi_die_type input_type, struct dcmi_die_id *die_id){
        CALL_FUNC(dcmiv2_get_device_die,dev_id,input_type,die_id)
    }

    static int (*dcmiv2_get_device_resource_info_func)(int dev_id, struct dcmi_proc_mem_info *proc_info,
        int *proc_num);
    static int dcmiv2_get_device_resource_info(int dev_id, struct dcmi_proc_mem_info *proc_info, int *proc_num){
        CALL_FUNC(dcmiv2_get_device_resource_info,dev_id,proc_info,proc_num)
    }

    static int (*dcmiv2_get_device_pcie_info_func)(int dev_id, struct dcmi_pcie_info_all *pcie_info);
    static int dcmiv2_get_device_pcie_info(int dev_id, struct dcmi_pcie_info_all *pcie_info){
        CALL_FUNC(dcmiv2_get_device_pcie_info,dev_id,pcie_info)
    }

    static int (*dcmiv2_get_device_board_info_func)(int dev_id, struct dcmi_board_info *board_info);
    static int dcmiv2_get_device_board_info(int dev_id, struct dcmi_board_info *board_info){
        CALL_FUNC(dcmiv2_get_device_board_info,dev_id,board_info)
    }

    static int (*dcmiv2_get_pcie_link_bandwidth_info_func)(int dev_id,
        struct dcmi_pcie_link_bandwidth_info *pcie_link_bandwidth_info);
    static int dcmiv2_get_pcie_link_bandwidth_info(int dev_id,
        struct dcmi_pcie_link_bandwidth_info *pcie_link_bandwidth_info){
        CALL_FUNC(dcmiv2_get_pcie_link_bandwidth_info,dev_id,pcie_link_bandwidth_info)
    }

    static int (*dcmiv2_get_dcmi_version_func)(char *dcmi_ver, int buf_size);
    static int dcmiv2_get_dcmi_version(char *dcmi_ver, int buf_size){
        CALL_FUNC(dcmiv2_get_dcmi_version,dcmi_ver,buf_size)
    }

    static int (*dcmiv2_get_device_ecc_info_func)(int dev_id, enum dcmi_device_type input_type,
        struct dcmi_ecc_info *device_ecc_info);
    static int dcmiv2_get_device_ecc_info(int dev_id, enum dcmi_device_type input_type,
        struct dcmi_ecc_info *device_ecc_info){
        CALL_FUNC(dcmiv2_get_device_ecc_info,dev_id,input_type,device_ecc_info)
    }

    static int (*dcmiv2_get_mainboard_id_func)(int dev_id, unsigned int *mainboard_id);
    static int dcmiv2_get_mainboard_id(int dev_id, unsigned int *mainboard_id){
        CALL_FUNC(dcmiv2_get_mainboard_id,dev_id,mainboard_id)
    }

    static int (*dcmiv2_start_ub_ping_mesh_func)(int dev_id, int count,
        struct dcmi_ub_ping_mesh_operate *ubping_mesh);
    static int dcmiv2_start_ub_ping_mesh(int dev_id, int count,
        struct dcmi_ub_ping_mesh_operate *ubping_mesh){
        CALL_FUNC(dcmiv2_start_ub_ping_mesh, dev_id, count, ubping_mesh)
    }

    static int (*dcmiv2_stop_ub_ping_mesh_func)(int dev_id, int task_id);
    static int dcmiv2_stop_ub_ping_mesh(int dev_id, int task_id){
        CALL_FUNC(dcmiv2_stop_ub_ping_mesh, dev_id, task_id)
    }

    static int (*dcmiv2_get_ub_ping_mesh_info_func)(int dev_id, int task_id,
                struct dcmi_ub_ping_mesh_info *ub_ping_mesh_reply, int mesh_reply_size, int *count);
    static int dcmiv2_get_ub_ping_mesh_info(int dev_id, int task_id,
                struct dcmi_ub_ping_mesh_info *ub_ping_mesh_reply, int mesh_reply_size, int *count){
        CALL_FUNC(dcmiv2_get_ub_ping_mesh_info, dev_id, task_id, ub_ping_mesh_reply, mesh_reply_size, count)
    }

    static int (*dcmiv2_get_ub_ping_mesh_state_func)(int dev_id, int task_id, unsigned int *state);
    static int dcmiv2_get_ub_ping_mesh_state(int dev_id, int task_id, unsigned int *state){
        CALL_FUNC(dcmiv2_get_ub_ping_mesh_state, dev_id, task_id, state)
    }

    static int (*dcmiv2_get_urma_device_cnt_func)(int dev_id, unsigned int *dev_cnt);
    static int dcmiv2_get_urma_device_cnt(int dev_id, unsigned int *dev_cnt) {
        CALL_FUNC(dcmiv2_get_urma_device_cnt, dev_id, dev_cnt)
    }

    static int (*dcmiv2_get_eid_list_by_urma_dev_index_func)(int dev_id, unsigned int dev_index,
                dcmi_urma_eid_info_t *eid_list, unsigned int *eid_cnt);
    static int dcmiv2_get_eid_list_by_urma_dev_index(int dev_id, unsigned int dev_index,
                dcmi_urma_eid_info_t *eid_list, unsigned int *eid_cnt) {
        CALL_FUNC(dcmiv2_get_eid_list_by_urma_dev_index, dev_id, dev_index, eid_list, eid_cnt)
    }


    // load .so files and functions
    static int dcmiInit_dl(const char* dcmiLibPath){
        if (dcmiLibPath == NULL) {
            fprintf (stderr,"lib path is null\n");
            return SO_NOT_FOUND;
        }
        dcmiHandle = dlopen(dcmiLibPath,RTLD_LAZY | RTLD_GLOBAL);
        if (dcmiHandle == NULL){
            fprintf (stderr,"%s\n",dlerror());
            return SO_NOT_FOUND;
        }
        dcmiv2_init_func = dlsym(dcmiHandle,"dcmiv2_init");
        dcmiv2_get_device_info_func = dlsym(dcmiHandle,"dcmiv2_get_device_info");
        dcmiv2_get_device_type_func = dlsym(dcmiHandle,"dcmiv2_get_device_type");
        dcmiv2_get_device_health_func = dlsym(dcmiHandle,"dcmiv2_get_device_health");
        dcmiv2_get_device_utilization_rate_func = dlsym(dcmiHandle,"dcmiv2_get_device_utilization_rate");
        dcmiv2_get_device_temperature_func = dlsym(dcmiHandle,"dcmiv2_get_device_temperature");
        dcmiv2_get_device_voltage_func = dlsym(dcmiHandle,"dcmiv2_get_device_voltage");
        dcmiv2_get_device_power_info_func = dlsym(dcmiHandle,"dcmiv2_get_device_power_info");
        dcmiv2_get_device_frequency_func = dlsym(dcmiHandle,"dcmiv2_get_device_frequency");
        dcmiv2_get_device_hbm_info_func = dlsym(dcmiHandle,"dcmiv2_get_device_hbm_info");
        dcmiv2_get_device_errorcode_func = dlsym(dcmiHandle,"dcmiv2_get_device_errorcode");
        dcmiv2_get_device_chip_info_func = dlsym(dcmiHandle,"dcmiv2_get_device_chip_info");
        dcmiv2_get_chip_phyid_from_dev_id_func = dlsym(dcmiHandle,"dcmiv2_get_chip_phyid_from_dev_id");
        dcmiv2_get_dev_id_from_chip_phyid_func = dlsym(dcmiHandle,"dcmiv2_get_dev_id_from_chip_phyid");
        dcmiv2_get_device_ip_func = dlsym(dcmiHandle,"dcmiv2_get_device_ip");
        dcmiv2_get_device_network_health_func = dlsym(dcmiHandle,"dcmiv2_get_device_network_health");
        dcmiv2_get_device_list_func = dlsym(dcmiHandle,"dcmiv2_get_device_list");
        dcmiv2_get_card_elabel_func = dlsym(dcmiHandle,"dcmiv2_get_card_elabel");
        dcmiv2_set_device_reset_func = dlsym(dcmiHandle,"dcmiv2_set_device_reset");
        dcmiv2_get_device_outband_channel_state_func = dlsym(dcmiHandle,"dcmiv2_get_device_outband_channel_state");
        dcmiv2_pre_reset_soc_func = dlsym(dcmiHandle,"dcmiv2_pre_reset_soc");
        dcmiv2_rescan_soc_func = dlsym(dcmiHandle,"dcmiv2_rescan_soc");
        dcmiv2_get_device_boot_status_func = dlsym(dcmiHandle,"dcmiv2_get_device_boot_status");
        dcmiv2_subscribe_fault_event_func = dlsym(dcmiHandle,"dcmiv2_subscribe_fault_event");
        dcmiv2_get_device_die_func = dlsym(dcmiHandle, "dcmiv2_get_device_die");
        dcmiv2_get_device_resource_info_func = dlsym(dcmiHandle, "dcmiv2_get_device_resource_info");
        dcmiv2_get_device_pcie_info_func = dlsym(dcmiHandle, "dcmiv2_get_device_pcie_info");
        dcmiv2_get_device_board_info_func = dlsym(dcmiHandle, "dcmiv2_get_device_board_info");
        dcmiv2_get_pcie_link_bandwidth_info_func = dlsym(dcmiHandle, "dcmiv2_get_pcie_link_bandwidth_info");
        dcmiv2_get_dcmi_version_func = dlsym(dcmiHandle,"dcmiv2_get_dcmi_version");
        dcmiv2_get_device_ecc_info_func = dlsym(dcmiHandle,"dcmiv2_get_device_ecc_info");
        dcmiv2_get_mainboard_id_func = dlsym(dcmiHandle, "dcmiv2_get_mainboard_id");
        dcmiv2_get_urma_device_cnt_func = dlsym(dcmiHandle, "dcmiv2_get_urma_device_cnt");
        dcmiv2_get_eid_list_by_urma_dev_index_func = dlsym(dcmiHandle, "dcmiv2_get_eid_list_by_urma_dev_index");
        dcmiv2_start_ub_ping_mesh_func = dlsym(dcmiHandle,"dcmiv2_start_ub_ping_mesh");
        dcmiv2_stop_ub_ping_mesh_func = dlsym(dcmiHandle,"dcmiv2_stop_ub_ping_mesh");
        dcmiv2_get_ub_ping_mesh_info_func = dlsym(dcmiHandle,"dcmiv2_get_ub_ping_mesh_info");
        dcmiv2_get_ub_ping_mesh_state_func = dlsym(dcmiHandle,"dcmiv2_get_ub_ping_mesh_state");
        return SUCCESS;
    }

    static int dcmiShutDown(void){
        if (dcmiHandle == NULL) {
            return SUCCESS;
        }
        return (dlclose(dcmiHandle) ? ERROR_UNKNOWN : SUCCESS);
    }
*/
import "C"
import (
	"fmt"
	"math"
	"time"
	"unsafe"

	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	"ascend-common/devmanager/common"
	"ascend-common/devmanager/dcmi"
)

// DcDriverInterface interface for dcmiv2
type DcDriverInterface interface {
	DcInit() error
	DcShutDown() error
	DcGetDcmiVersion() (string, error)
	DcGetDeviceCount() (int32, error)
	DcGetDeviceHealth(logicID int32) (int32, error)
	DcGetDeviceNetWorkHealth(logicID int32) (uint32, error)
	DcGetDeviceUtilizationRate(logicID int32, devType common.DeviceType) (int32, error)
	DcGetDeviceTemperature(logicID int32) (int32, error)
	DcGetDeviceVoltage(logicID int32) (float32, error)
	DcGetDevicePowerInfo(logicID int32) (float32, error)
	DcGetDeviceFrequency(logicID int32, devType common.DeviceType) (uint32, error)
	DcGetHbmInfo(int32) (*common.HbmInfo, error)
	DcGetDeviceErrorCode(logicID int32) (int32, int64, error)
	DcGetChipInfo(logicID int32) (*common.ChipInfo, error)
	DcGetPhysicIDFromLogicID(int32) (int32, error)
	DcGetLogicIDFromPhysicID(int32) (int32, error)
	DcGetDeviceIPAddress(logicID int32, ipType int32) (string, error)
	DcGetDieID(logicID int32, dcmiDieType dcmi.DieType) (string, error)
	DcGetPCIeBusInfo(logicID int32) (string, error)
	DcGetDeviceList() (int32, []int32, error)
	DcGetDeviceTotalResource(logicID int32) (common.CgoSocTotalResource, error)
	DcGetDeviceFreeResource(logicID int32) (common.CgoSocFreeResource, error)
	DcGetVDeviceInfo(logicID int32) (common.VirtualDevInfo, error)
	DcSetDeviceReset(logicID int32) error
	DcPreResetSoc(logicID int32) error
	DcGetOutBandChannelState(logicID int32) error
	DcSetDeviceResetOutBand(logicID int32) error
	DcRescanSoc(logicID int32) error
	DcGetDeviceBootStatus(int32) (int, error)
	DcGetSuperPodInfo(logicID int32) (common.CgoSuperPodInfo, error)
	DcGetDeviceAllErrorCode(logicID int32) (int32, []int64, error)
	DcSubscribeDeviceFaultEvent(logicID int32) error
	DcSetFaultEventCallFunc(func(common.DevFaultInfo))
	DcGetDevProcessInfo(logicID int32) (*common.DevProcessInfo, error)
	DcGetDeviceBoardInfo(logicID int32) (common.BoardInfo, error)
	DcGetPCIEBandwidth(logicID int32, profilingTime int) (common.PCIEBwStat, error)
	DcGetDeviceEccInfo(logicID int32, inputType common.DcmiDeviceType) (*common.ECCInfo, error)
	DcGetSioInfo(logicID int32) (common.SioCrcErrStatisticInfo, error)
	DcGetDeviceMainBoardInfo(logicID int32) (uint32, error)
	DcGetCardElabel(logicID int32) (common.ElabelInfo, error)
	DcGetUrmaDeviceCount(logicID int32) (int32, error)
	DcGetUrmaDevEidList(logicID int32, urmaDevIndex int32) (*common.UrmaDeviceInfo, error)
	DcGetUrmaDevEidListAll(logicID int32) ([]common.UrmaDeviceInfo, error)
	DcStartUbPingMesh(logicID int32, operate common.HccspingMeshOperate) error
	DcStopUbPingMesh(logicID int32, taskID uint) error
	DcGetUbPingMeshInfo(int32, uint, int) (*common.HccspingMeshInfo, error)
	DcGetUbPingMeshState(logicID int32, taskID uint) (int, error)
}

const (
	dcmiLibraryName = "libdcmi.so"
)

var faultEventCallFunc func(common.DevFaultInfo) = nil
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

const maxCArraySize = 1 << 30 // 1 Gi elements; practical upper bound for C array mapping for A5

// DcManager for manager dcmi interface
type DcManager struct{}

// DcInit load symbol and initialize dcmi
func (d *DcManager) DcInit() error {
	dcmiLibPath, err := utils.GetDriverLibPath(dcmiLibraryName)
	if err != nil {
		return err
	}
	cDcmiTemplateName := C.CString(dcmiLibPath)
	defer C.free(unsafe.Pointer(cDcmiTemplateName))
	if retCode := C.dcmiInit_dl(cDcmiTemplateName); retCode != C.SUCCESS {
		return fmt.Errorf("dcmi lib load failed, error code: %d", int32(retCode))
	}
	if retCode := C.dcmiv2_init_new(); retCode != C.SUCCESS {
		return fmt.Errorf("dcmiv2 init failed, error code: %d", int32(retCode))
	}
	return nil
}

// DcShutDown clean the dynamically loaded resource
func (d *DcManager) DcShutDown() error {
	if retCode := C.dcmiShutDown(); retCode != C.SUCCESS {
		return fmt.Errorf("dcmi shut down failed, error code: %d", int32(retCode))
	}
	return nil
}

// DcGetDcmiVersion return dcmi version
func (d *DcManager) DcGetDcmiVersion() (string, error) {
	cDcmiVer := C.CString(string(make([]byte, dcmi.DcmiVersionLen)))
	defer C.free(unsafe.Pointer(cDcmiVer))
	if retCode := C.dcmiv2_get_dcmi_version(cDcmiVer, dcmi.DcmiVersionLen+1); int32(retCode) != common.Success {
		return "", fmt.Errorf("get dcmi version failed, errCode: %d", int32(retCode))
	}
	return C.GoString(cDcmiVer), nil
}

// DcGetDeviceCount get device count
func (d *DcManager) DcGetDeviceCount() (int32, error) {
	devNum, _, err := d.DcGetDeviceList()
	if err != nil {
		return common.RetError, fmt.Errorf("get device count failed, error: %v", err)
	}
	return devNum, nil
}

// DcGetDeviceHealth get device health
func (d *DcManager) DcGetDeviceHealth(logicID int32) (int32, error) {
	if !common.IsValidLogicIDOrPhyID(logicID) {
		return common.RetError, fmt.Errorf("logicID(%d) is invalid", logicID)
	}
	var health C.uint
	if retCode := C.dcmiv2_get_device_health(C.int(logicID), &health); int32(retCode) != common.Success {
		return common.RetError, fmt.Errorf("get device (logicID: %d) health state failed, ret "+
			"code: %d, health code: %d", logicID, int32(retCode), int64(health))
	}
	if common.IsGreaterThanOrEqualInt32(int64(health)) {
		return common.RetError, fmt.Errorf("get wrong health state , device (logicID: %d) "+
			"health: %d", logicID, int64(health))
	}
	return int32(health), nil
}

func callDcmiGetDeviceNetworkHealth(logicID int32, result chan<- common.DeviceNetworkHealth) {
	var healthCode C.enum_dcmi_rdfx_detect_result
	rCode := C.dcmiv2_get_device_network_health(C.int(logicID), &healthCode)
	result <- common.DeviceNetworkHealth{HealthCode: uint32(healthCode), RetCode: int32(rCode)}
}

// DcGetDeviceNetWorkHealth get device network health by logicID
func (d *DcManager) DcGetDeviceNetWorkHealth(logicID int32) (uint32, error) {
	if !common.IsValidLogicIDOrPhyID(logicID) {
		return common.UnRetError, fmt.Errorf("logicID(%d) is invalid", logicID)
	}

	result := make(chan common.DeviceNetworkHealth, 1)
	go callDcmiGetDeviceNetworkHealth(logicID, result)
	select {
	case res := <-result:
		if res.RetCode != common.Success {
			return common.UnRetError, fmt.Errorf("get device network healthCode failed, logicID(%d),"+
				" ret code: %d, health code: %d", logicID, res.RetCode, res.HealthCode)
		}

		if int32(res.HealthCode) < 0 || int32(res.HealthCode) > int32(math.MaxInt8) {
			return common.UnRetError, fmt.Errorf("get wrong device network healthCode, logicID(%d),"+
				" error healthCode: %d", logicID, int32(res.HealthCode))
		}

		return res.HealthCode, nil
	// dcmiv2_get_device_network_health is occasionally blocked for a long time, because of retrying,
	// after the card dropped. This method is used to interrupt the execution of the dcmi interface,
	// if invoking time excceeds 1 second.
	case <-time.After(common.DcmiApiTimeout * time.Second):
		return common.UnRetError, fmt.Errorf("accessing dcmiv2_get_device_network_health interface timeout, "+
			"logicID(%d)", logicID)
	}
}

func (d *DcManager) DcGetDeviceUtilizationRate(logicID int32, devType common.DeviceType) (int32, error) {
	if !common.IsValidLogicIDOrPhyID(logicID) {
		return common.RetError, fmt.Errorf("logicID(%d) is invalid", logicID)
	}
	var rate C.uint
	if retCode := C.dcmiv2_get_device_utilization_rate(C.int(logicID), C.int(devType.Code),
		&rate); int32(retCode) != common.Success {
		return common.RetError,
			buildDcmiErr(logicID, fmt.Sprintf("utilization (name: %v, code:%d)", devType.Name,
				devType.Code), retCode)
	}
	if !common.IsValidUtilizationRate(uint32(rate)) {
		return common.RetError, fmt.Errorf("get wrong device (logicID: %d) "+
			"utilization (name: %v, code:%d): %d", logicID, devType.Name, devType.Code, uint32(rate))
	}
	return int32(rate), nil
}

func (d *DcManager) DcGetDeviceTemperature(logicID int32) (int32, error) {
	if !common.IsValidLogicIDOrPhyID(logicID) {
		return common.RetError, fmt.Errorf("logicID(%d) is invalid", logicID)
	}
	var temp C.int
	if retCode := C.dcmiv2_get_device_temperature(C.int(logicID), &temp); int32(retCode) != common.Success {
		return common.RetError, fmt.Errorf("get device (logicID: %d) temperature failed, error "+
			"code is : %d", logicID, int32(retCode))
	}
	parsedTemp := int32(temp)
	if parsedTemp < int32(common.DefaultTemperatureWhenQueryFailed) {
		return common.RetError, fmt.Errorf("get wrong device temperature, devcie (logicID: %d), "+
			"temperature: %d", logicID, parsedTemp)
	}
	return parsedTemp, nil
}

// DcGetDeviceVoltage the accuracy is 0.01v.
func (d *DcManager) DcGetDeviceVoltage(logicID int32) (float32, error) {
	if !common.IsValidLogicIDOrPhyID(logicID) {
		return common.RetError, fmt.Errorf("logicID(%d) is invalid", logicID)
	}
	var vol C.uint
	if retCode := C.dcmiv2_get_device_voltage(C.int(logicID), &vol); int32(retCode) != common.Success {
		return common.RetError, fmt.Errorf("failed to obtain the voltage based on logicID(%d) "+
			", error code: %d", logicID, int32(retCode))
	}
	// the voltage's value is error if it's greater than or equal to MaxInt32
	if common.IsGreaterThanOrEqualInt32(int64(vol)) {
		return common.RetError, fmt.Errorf("voltage value out of range(max is int32), "+
			"logicID(%d), voltage: %d", logicID, int64(vol))
	}

	return float32(vol) * common.ReduceOnePercent, nil
}

// DcGetDevicePowerInfo the accuracy is 0.1w, the result like: 8.2 by dcmiv2 api
func (d *DcManager) DcGetDevicePowerInfo(logicID int32) (float32, error) {
	if !common.IsValidLogicIDOrPhyID(logicID) {
		return common.RetError, fmt.Errorf("logicID(%d) is invalid", logicID)
	}
	var cpower C.int
	if retCode := C.dcmiv2_get_device_power_info(C.int(logicID), &cpower); int32(retCode) != common.Success {
		return common.RetError, fmt.Errorf("failed to obtain the power based on logicID(%d)"+
			", error code: %d", logicID, int32(retCode))
	}
	parsedPower := float32(cpower)
	if parsedPower < 0 {
		return common.RetError, fmt.Errorf("get wrong device power, logicID(%d) , power: %f", logicID, parsedPower)
	}
	return parsedPower * common.ReduceTenth, nil
}

// DcGetDeviceFrequency get device frequency, unit MHz
func (d *DcManager) DcGetDeviceFrequency(logicID int32, devType common.DeviceType) (uint32, error) {
	if !common.IsValidLogicIDOrPhyID(logicID) {
		return common.UnRetError, fmt.Errorf("logicID(%d) is invalid", logicID)
	}
	var cFrequency C.uint
	if retCode := C.dcmiv2_get_device_frequency(C.int(logicID), C.enum_dcmi_freq_type(devType.Code),
		&cFrequency); int32(retCode) != common.Success {
		return common.UnRetError,
			buildDcmiErr(logicID, fmt.Sprintf("frequency (name: %v, code:%d)", devType.Name, devType.Code), retCode)
	}
	// check whether cFrequency is too big
	if common.IsGreaterThanOrEqualInt32(int64(cFrequency)) || int64(cFrequency) < 0 {
		return common.UnRetError, fmt.Errorf("frequency value out of range [0, int32),logicID(%d), "+
			"frequency (name: %v, code:%d): %d", logicID, devType.Name, devType.Code, int64(cFrequency))
	}
	return uint32(cFrequency), nil
}

// DcGetHbmInfo get HBM information
func (d *DcManager) DcGetHbmInfo(logicID int32) (*common.HbmInfo, error) {
	if !common.IsValidLogicIDOrPhyID(logicID) {
		return nil, fmt.Errorf("logicID(%d) is invalid", logicID)
	}
	var cHbmInfo C.struct_dcmi_hbm_info
	if retCode := C.dcmiv2_get_device_hbm_info(C.int(logicID), &cHbmInfo); int32(retCode) != common.Success {
		return nil, buildDcmiErr(logicID, "high bandwidth memory info", retCode)
	}
	hbmTemp := int32(cHbmInfo.temp)
	if hbmTemp < 0 {
		return nil, fmt.Errorf("get wrong device HBM temporary, logicID(%d), HBM.temp: %d", logicID, hbmTemp)
	}
	return &common.HbmInfo{
		MemorySize:        uint64(cHbmInfo.memory_size),
		Frequency:         uint32(cHbmInfo.freq),
		Usage:             uint64(cHbmInfo.memory_usage),
		Temp:              hbmTemp,
		BandWidthUtilRate: uint32(cHbmInfo.bandwith_util_rate)}, nil
}

// DcGetDeviceErrorCode get error code info of device by logicID
func (d *DcManager) DcGetDeviceErrorCode(logicID int32) (int32, int64, error) {
	if !common.IsValidLogicIDOrPhyID(logicID) {
		return common.RetError, common.RetError, fmt.Errorf("logicID(%d) is invalid", logicID)
	}
	var errCount C.int
	var errCodeArray [common.MaxErrorCodeCount]C.uint
	if retCode := C.dcmiv2_get_device_errorcode(C.int(logicID), &errCount, &errCodeArray[0],
		common.MaxErrorCodeCount); int32(retCode) != common.Success {
		return common.RetError, common.RetError, fmt.Errorf("failed to obtain the device errorcode based on "+
			"logicID(%d), error code: %d, error count: %d", logicID, int32(retCode), int32(errCount))
	}
	if int32(errCount) < 0 || int32(errCount) > common.MaxErrorCodeCount {
		return common.RetError, common.RetError, fmt.Errorf("get wrong errorcode count, "+
			"logicID(%d), errorcode count: %d", logicID, int32(errCount))
	}
	return int32(errCount), int64(errCodeArray[0]), nil
}

func convertUCharToCharArr(cgoArr [dcmi.MaxChipNameLen]C.uchar) []byte {
	var charArr []byte
	for _, v := range cgoArr {
		if v == 0 {
			break
		}
		charArr = append(charArr, byte(v))
	}
	return charArr
}

func (d *DcManager) DcGetChipInfo(logicID int32) (*common.ChipInfo, error) {
	if !common.IsValidLogicIDOrPhyID(logicID) {
		return nil, fmt.Errorf("logicID(%d) is invalid", logicID)
	}
	var chipInfo C.struct_dcmi_chip_info_v2
	chip := &common.ChipInfo{}
	if rCode := C.dcmiv2_get_device_chip_info(C.int(logicID), &chipInfo); int32(rCode) != common.Success {
		hwlog.RunLog.Debugf("get device ChipInfo information failed, logicID(%d),"+
			" error code: %d", logicID, int32(rCode))
		return nil, fmt.Errorf("get device ChipInfo information failed, logicID(%d),"+
			" error code: %d", logicID, int32(rCode))
	}
	chip.Name = string(convertUCharToCharArr(chipInfo.chip_name))
	chip.Type = string(convertUCharToCharArr(chipInfo.chip_type))
	chip.Version = string(convertUCharToCharArr(chipInfo.chip_ver))
	chip.AICoreCnt = int(chipInfo.aicore_cnt)
	chip.NpuName = string(convertUCharToCharArr(chipInfo.npu_name))
	if !common.IsValidChipInfo(chip) {
		return nil, fmt.Errorf("get device ChipInfo information failed, chip info is empty,"+
			" logicID(%d)", logicID)
	}

	return chip, nil
}

// DcGetPhysicIDFromLogicID get physicID from logicID
func (d *DcManager) DcGetPhysicIDFromLogicID(logicID int32) (int32, error) {
	if !common.IsValidLogicIDOrPhyID(logicID) {
		return common.RetError, fmt.Errorf("logicID(%d) is invalid", logicID)
	}
	var physicID C.uint
	if rCode := C.dcmiv2_get_chip_phyid_from_dev_id(C.uint(logicID), &physicID); int32(rCode) != common.Success {
		return common.RetError, fmt.Errorf("get physic id from logicID(%d) failed, error code: %d", logicID, int32(rCode))
	}
	if !common.IsValidLogicIDOrPhyID(int32(physicID)) {
		return common.RetError, fmt.Errorf("get wrong physicID(%d) from logicID(%d)", uint32(physicID), logicID)
	}
	return int32(physicID), nil
}

// DcGetLogicIDFromPhysicID get logicID from physicID
func (d *DcManager) DcGetLogicIDFromPhysicID(physicID int32) (int32, error) {
	if !common.IsValidLogicIDOrPhyID(physicID) {
		return common.RetError, fmt.Errorf("physicID(%d) is invalid", physicID)
	}
	var logicID C.uint
	if rCode := C.dcmiv2_get_dev_id_from_chip_phyid(C.uint(physicID), &logicID); int32(rCode) != common.Success {
		return common.RetError, fmt.Errorf("get logicID from physicID(%d) failed, error code: %d",
			physicID, int32(rCode))
	}

	if !common.IsValidLogicIDOrPhyID(int32(logicID)) {
		return common.RetError, fmt.Errorf("get wrong logicID(%d) from physicID(%d)", uint32(logicID), physicID)
	}
	return int32(logicID), nil
}

// DcGetDeviceList get device id list
func (d *DcManager) DcGetDeviceList() (int32, []int32, error) {
	var ids [common.HiAIMaxCardNum]C.int
	var dNum C.int
	if retCode := C.dcmiv2_get_device_list(&ids[0], &dNum, common.HiAIMaxCardNum); int32(retCode) != common.Success {
		return common.RetError, nil, fmt.Errorf("get device list failed, error code: %d", int32(retCode))
	}
	// checking device's quantity
	if dNum <= 0 || dNum > common.HiAIMaxCardNum {
		return common.RetError, nil, fmt.Errorf("get error device quantity: %d", int32(dNum))
	}
	var deviceNum = int32(dNum)
	var i int32
	var deviceIDList []int32
	for i = 0; i < deviceNum; i++ {
		deviceID := int32(ids[i])
		if deviceID < 0 {
			hwlog.RunLog.Errorf("get invalid device ID: %d", deviceID)
			continue
		}
		deviceIDList = append(deviceIDList, deviceID)
	}
	return deviceNum, deviceIDList, nil
}

func buildDcmiErr(logicID int32, msg string, errCode C.int) error {
	errDesc, ok := dcmiErrMap[int32(errCode)]
	if !ok {
		errDesc = "unknown error code"
	}
	return fmt.Errorf("logicID(%d):get %s info failed,error code: %v,error desc: %v",
		logicID, msg, errCode, errDesc)
}
