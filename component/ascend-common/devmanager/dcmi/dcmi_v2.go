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

	static int (*dcmiv2_get_all_device_count_func)(int *all_device_count);
    static int dcmiv2_get_all_device_count(int *all_device_count){
        CALL_FUNC(dcmiv2_get_all_device_count,all_device_count)
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
    static int dcmiv2Init_dl(const char* dcmiLibPath){
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
		dcmiv2_get_all_device_count_func = dlsym(dcmiHandle,"dcmiv2_get_all_device_count");
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
	"errors"
	"fmt"
	"math"
	"net"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	"ascend-common/devmanager/common"
)

// DcV2DriverInterface interface for dcmiv2
type DcV2DriverInterface interface {
	DcInit() error
	DcShutDown() error
	DcGetDcmiVersion() (string, error)
	DcGetAllDeviceCount() (int32, error)
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
	DcGetDieID(logicID int32, dcmiDieType DieType) (string, error)
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

// DcV2Manager for manager dcmi interface
type DcV2Manager struct{}

// check if struct DcV2Manager implements all methods of interface DcV2DriverInterface in build stage
var _ DcV2DriverInterface = &DcV2Manager{}

// DcInit load symbol and initialize dcmi
func (d *DcV2Manager) DcInit() error {
	dcmiLibPath, err := utils.GetDriverLibPath(dcmiLibraryName)
	if err != nil {
		return err
	}
	cDcmiTemplateName := C.CString(dcmiLibPath)
	defer C.free(unsafe.Pointer(cDcmiTemplateName))
	if retCode := C.dcmiv2Init_dl(cDcmiTemplateName); retCode != C.SUCCESS {
		return fmt.Errorf("dcmi lib load failed, error code: %d", int32(retCode))
	}
	if retCode := C.dcmiv2_init_new(); retCode != C.SUCCESS {
		return fmt.Errorf("dcmiv2 init failed, error code: %d", int32(retCode))
	}
	return nil
}

// DcShutDown clean the dynamically loaded resource
func (d *DcV2Manager) DcShutDown() error {
	if retCode := C.dcmiShutDown(); retCode != C.SUCCESS {
		return fmt.Errorf("dcmi shut down failed, error code: %d", int32(retCode))
	}
	return nil
}

// DcGetDcmiVersion return dcmi version
func (d *DcV2Manager) DcGetDcmiVersion() (string, error) {
	cDcmiVer := C.CString(string(make([]byte, DcmiVersionLen)))
	defer C.free(unsafe.Pointer(cDcmiVer))
	if retCode := C.dcmiv2_get_dcmi_version(cDcmiVer, DcmiVersionLen+1); int32(retCode) != common.Success {
		return "", fmt.Errorf("get dcmi version failed, errCode: %d", int32(retCode))
	}
	return C.GoString(cDcmiVer), nil
}

// DcGetAllDeviceCount get all device count
func (d *DcV2Manager) DcGetAllDeviceCount() (int32, error) {
	var allDeviceCount C.int
	if retCode := C.dcmiv2_get_all_device_count(&allDeviceCount); int32(retCode) != common.Success {
		return 0, fmt.Errorf("get all device count failed, errCode: %d", int32(retCode))
	}
	return int32(allDeviceCount), nil
}

// DcGetDeviceHealth get device health
func (d *DcV2Manager) DcGetDeviceHealth(logicID int32) (int32, error) {
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

func callDcmiV2GetDeviceNetworkHealth(logicID int32, result chan<- common.DeviceNetworkHealth) {
	var healthCode C.enum_dcmi_rdfx_detect_result
	rCode := C.dcmiv2_get_device_network_health(C.int(logicID), &healthCode)
	result <- common.DeviceNetworkHealth{HealthCode: uint32(healthCode), RetCode: int32(rCode)}
}

// DcGetDeviceNetWorkHealth get device network health by logicID
func (d *DcV2Manager) DcGetDeviceNetWorkHealth(logicID int32) (uint32, error) {
	if !common.IsValidLogicIDOrPhyID(logicID) {
		return common.UnRetError, fmt.Errorf("logicID(%d) is invalid", logicID)
	}

	result := make(chan common.DeviceNetworkHealth, 1)
	go callDcmiV2GetDeviceNetworkHealth(logicID, result)
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

// DcGetDeviceUtilizationRate get device utilization rate
func (d *DcV2Manager) DcGetDeviceUtilizationRate(logicID int32, devType common.DeviceType) (int32, error) {
	if !common.IsValidLogicIDOrPhyID(logicID) {
		return common.RetError, fmt.Errorf("logicID(%d) is invalid", logicID)
	}
	var rate C.uint
	if retCode := C.dcmiv2_get_device_utilization_rate(C.int(logicID), C.int(devType.Code),
		&rate); int32(retCode) != common.Success {
		return common.RetError,
			buildDcmiV2Err(logicID, fmt.Sprintf("utilization (name: %v, code:%d)", devType.Name,
				devType.Code), retCode)
	}
	if !common.IsValidUtilizationRate(uint32(rate)) {
		return common.RetError, fmt.Errorf("get wrong device (logicID: %d) "+
			"utilization (name: %v, code:%d): %d", logicID, devType.Name, devType.Code, uint32(rate))
	}
	return int32(rate), nil
}

// DcGetDeviceTemperature get device temperature
func (d *DcV2Manager) DcGetDeviceTemperature(logicID int32) (int32, error) {
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
func (d *DcV2Manager) DcGetDeviceVoltage(logicID int32) (float32, error) {
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
func (d *DcV2Manager) DcGetDevicePowerInfo(logicID int32) (float32, error) {
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
func (d *DcV2Manager) DcGetDeviceFrequency(logicID int32, devType common.DeviceType) (uint32, error) {
	if !common.IsValidLogicIDOrPhyID(logicID) {
		return common.UnRetError, fmt.Errorf("logicID(%d) is invalid", logicID)
	}
	var cFrequency C.uint
	if retCode := C.dcmiv2_get_device_frequency(C.int(logicID), C.enum_dcmi_freq_type(devType.Code),
		&cFrequency); int32(retCode) != common.Success {
		return common.UnRetError,
			buildDcmiV2Err(logicID, fmt.Sprintf("frequency (name: %v, code:%d)", devType.Name, devType.Code), retCode)
	}
	// check whether cFrequency is too big
	if common.IsGreaterThanOrEqualInt32(int64(cFrequency)) || int64(cFrequency) < 0 {
		return common.UnRetError, fmt.Errorf("frequency value out of range [0, int32),logicID(%d), "+
			"frequency (name: %v, code:%d): %d", logicID, devType.Name, devType.Code, int64(cFrequency))
	}
	return uint32(cFrequency), nil
}

// DcGetHbmInfo get HBM information
func (d *DcV2Manager) DcGetHbmInfo(logicID int32) (*common.HbmInfo, error) {
	if !common.IsValidLogicIDOrPhyID(logicID) {
		return nil, fmt.Errorf("logicID(%d) is invalid", logicID)
	}
	var cHbmInfo C.struct_dcmi_hbm_info
	if retCode := C.dcmiv2_get_device_hbm_info(C.int(logicID), &cHbmInfo); int32(retCode) != common.Success {
		return nil, buildDcmiV2Err(logicID, "high bandwidth memory info", retCode)
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
func (d *DcV2Manager) DcGetDeviceErrorCode(logicID int32) (int32, int64, error) {
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

// DcGetChipInfo get chip info
func (d *DcV2Manager) DcGetChipInfo(logicID int32) (*common.ChipInfo, error) {
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
func (d *DcV2Manager) DcGetPhysicIDFromLogicID(logicID int32) (int32, error) {
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
func (d *DcV2Manager) DcGetLogicIDFromPhysicID(physicID int32) (int32, error) {
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

// DcGetDeviceIPAddress get device ip addresses
func (d *DcV2Manager) DcGetDeviceIPAddress(logicID int32, ipType int32) (string, error) {
	if !common.IsValidLogicIDOrPhyID(logicID) {
		return "", fmt.Errorf("logicID(%d) is invalid", logicID)
	}
	var portType C.enum_dcmi_port_type = 1
	var portID C.int
	var ipAddress C.struct_dcmi_ip_addr
	var maskAddress C.struct_dcmi_ip_addr
	if ipType == IpAddrTypeV6 {
		ipAddress.ip_type = IpAddrTypeV6
	}
	rCode := C.dcmiv2_get_device_ip(C.int(logicID), portType, portID, &ipAddress, &maskAddress)
	if int32(rCode) != common.Success {
		return "", fmt.Errorf("get device IP address failed, logicID(%d), error code: %d", logicID, int32(rCode))
	}
	if ipType == IpAddrTypeV6 {
		return d.buildIPv6Addr(ipAddress)
	}
	return d.buildIPv4Addr(ipAddress)
}

func (d *DcV2Manager) buildIPv4Addr(ipAddress C.struct_dcmi_ip_addr) (string, error) {
	deviceIP := make([]string, 0, net.IPv4len)
	for key, val := range ipAddress.u_addr {
		if key >= net.IPv4len {
			break
		}
		deviceIP = append(deviceIP, fmt.Sprintf("%v", val))
	}
	if netIP := net.ParseIP(strings.Join(deviceIP, ".")); netIP != nil {
		return netIP.String(), nil
	}
	return "", fmt.Errorf("the device IPv4 address is invalid, value: %v", deviceIP)
}

func (d *DcV2Manager) buildIPv6Addr(ipAddress C.struct_dcmi_ip_addr) (string, error) {
	deviceIP := make([]byte, 0, net.IPv6len)
	for key, val := range ipAddress.u_addr {
		if key >= net.IPv6len {
			break
		}
		deviceIP = append(deviceIP, byte(val))
	}
	if netIP := net.IP(deviceIP); netIP != nil {
		return netIP.String(), nil
	}
	return "", fmt.Errorf("the device IPv6 address is invalid, value: %v", deviceIP)
}

// DcGetDieID get die id
func (d *DcV2Manager) DcGetDieID(logicID int32, dcmiDieType DieType) (string, error) {
	if !common.IsValidLogicIDOrPhyID(logicID) {
		return "", fmt.Errorf("logicID(%d) is invalid", logicID)
	}

	if dcmiDieType != VDIE && dcmiDieType != NDIE {
		return "", fmt.Errorf("dcmi die type can only be one of %d or %d", VDIE, NDIE)
	}

	var dieIDObj C.struct_dcmi_die_id
	if retCode := C.dcmiv2_get_device_die(C.int(logicID),
		C.enum_dcmi_die_type(dcmiDieType), &dieIDObj); int32(retCode) != common.Success {
		return "", buildDcmiV2Err(logicID, "chip die ID", retCode)
	}

	dieIDStr := make([]string, DieIDCount)
	hwlog.RunLog.Debugf("logicID(%d) get die type(%d) value %v", logicID, dcmiDieType, dieIDObj.soc_die)
	for i := 0; i < DieIDCount; i++ {
		s := strconv.FormatUint(uint64(dieIDObj.soc_die[i]), HexBase)
		// Each part of the die id consists of 8 characters, and if the length is not enough,
		// zero is added at the beginning
		dieIDStr[i] = fmt.Sprintf("%08s", s)
	}
	return strings.ToUpper(strings.Join(dieIDStr, "-")), nil
}

// DcGetDeviceList get device id list
func (d *DcV2Manager) DcGetDeviceList() (int32, []int32, error) {
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

func buildDcmiV2Err(logicID int32, msg string, errCode C.int) error {
	errDesc, ok := dcmiErrMap[int32(errCode)]
	if !ok {
		errDesc = "unknown error code"
	}
	return fmt.Errorf("logicID(%d):get %s info failed,error code: %v,error desc: %v",
		logicID, msg, errCode, errDesc)
}

// DcGetUrmaDeviceCount get urma device count
func (d *DcV2Manager) DcGetUrmaDeviceCount(logicID int32) (int32, error) {
	if !common.IsValidLogicIDOrPhyID(logicID) {
		return common.RetError, fmt.Errorf("logicID(%d) is invalid", logicID)
	}
	var cnt C.uint
	if retCode := C.dcmiv2_get_urma_device_cnt(C.int(logicID), &cnt); retCode != common.Success {
		return common.RetError, fmt.Errorf("dcmi get urma device count failed logicID(%d) error "+
			"code: %d", logicID, int32(retCode))
	}
	return int32(cnt), nil
}

// DcGetUrmaDevEidList get urma device index EID info
func (d *DcV2Manager) DcGetUrmaDevEidList(logicID int32, urmaDevIndex int32) (*common.UrmaDeviceInfo, error) {
	if !common.IsValidLogicIDOrPhyID(logicID) {
		return nil, fmt.Errorf("logicID(%d) is invalid", logicID)
	}
	if urmaDevIndex < 0 || urmaDevIndex >= MaxUrmaDevCnt {
		return nil, fmt.Errorf("urma device index is %d out of range [0, %d), logicID(%d)",
			urmaDevIndex, MaxUrmaDevCnt, logicID)
	}

	var eidInfoList [common.EidNumMax]C.dcmi_urma_eid_info_t
	eidInfoListPtr := (*C.dcmi_urma_eid_info_t)(&eidInfoList[0])
	eidCnt := C.uint(common.EidNumMax)
	if ret := C.dcmiv2_get_eid_list_by_urma_dev_index(C.int(logicID), C.uint(urmaDevIndex), eidInfoListPtr,
		&eidCnt); ret != common.Success {
		return nil, fmt.Errorf("dcmi get urma device info failed, logicID(%d) index(%d) ret(%d)",
			logicID, urmaDevIndex, int(ret))
	}

	info, err := convertUrmaDeviceInfo(eidInfoListPtr, eidCnt)
	if err != nil {
		return nil, fmt.Errorf("convert urma device info failed, logicID(%d) index(%d), err: %v",
			logicID, urmaDevIndex, err)
	}
	return info, nil
}

// DcGetUrmaDevEidListAll get all urma device eid list
func (d *DcV2Manager) DcGetUrmaDevEidListAll(logicID int32) ([]common.UrmaDeviceInfo, error) {
	feCnt, err := d.DcGetUrmaDeviceCount(logicID)
	if err != nil {
		return []common.UrmaDeviceInfo{}, err
	}

	if feCnt > MaxUrmaDevCnt || feCnt < 0 {
		return []common.UrmaDeviceInfo{}, fmt.Errorf("urma device number is %d, out of range [0, %d], "+
			"logicID(%d)", feCnt, MaxUrmaDevCnt, logicID)
	}

	infos := make([]common.UrmaDeviceInfo, feCnt)
	for index := int32(0); index < feCnt; index++ {
		eidInfo, err := d.DcGetUrmaDevEidList(logicID, index)
		if err != nil || eidInfo == nil {
			return []common.UrmaDeviceInfo{}, err
		}
		infos[index] = *eidInfo
	}
	return infos, nil
}

// DcStartUbPingMesh start ub ping mesh
func (d *DcV2Manager) DcStartUbPingMesh(logicID int32, operate common.HccspingMeshOperate) error {
	ops := operate.UBPingMeshOperateList
	if len(ops) == 0 {
		return fmt.Errorf("no UB ping mesh operations provided")
	}

	size := len(ops)
	cOpsPtr := C.malloc(C.size_t(size) * C.size_t(unsafe.Sizeof(C.struct_dcmi_ub_ping_mesh_operate{})))
	if cOpsPtr == nil {
		return errors.New("failed to allocate memory for UB ping mesh C array")
	}
	defer C.free(cOpsPtr)

	cOpsPtrStruct := (*C.struct_dcmi_ub_ping_mesh_operate)(cOpsPtr)
	cOpsPtrStruct, err := buildUbPingMeshCArray(ops, cOpsPtrStruct, size)
	if err != nil {
		return err
	}

	if retCode := C.dcmiv2_start_ub_ping_mesh(C.int(logicID), C.int(len(ops)),
		cOpsPtrStruct); retCode != common.Success {
		return fmt.Errorf("dcmi start ub ping mesh failed logicID(%d) error code: %d",
			logicID, int32(retCode))
	}

	return nil
}

// DcGetUbPingMeshInfo get ub ping mesh info
func (d *DcV2Manager) DcGetUbPingMeshInfo(logicID int32, taskID uint, meshReplySize int) (*common.HccspingMeshInfo,
	error) {
	if meshReplySize <= 0 {
		return nil, fmt.Errorf("meshReplySize must be > 0")
	}

	cInfosPtr := C.malloc(C.size_t(meshReplySize) * C.size_t(unsafe.Sizeof(C.struct_dcmi_ub_ping_mesh_info{})))
	if cInfosPtr == nil {
		return nil, fmt.Errorf("failed to allocate memory")
	}
	defer C.free(cInfosPtr)

	var count C.int
	hwlog.RunLog.Debugf("get: the logicID %d, taskID %d, meshReplySize %d", logicID, taskID, meshReplySize)
	if retCode := C.dcmiv2_get_ub_ping_mesh_info(
		C.int(logicID), C.int(taskID), (*C.struct_dcmi_ub_ping_mesh_info)(cInfosPtr),
		C.int(meshReplySize), &count,
	); retCode != common.Success {
		return nil, fmt.Errorf("dcmi get ub ping mesh info failed logicID(%d) error code: %d",
			logicID, int32(retCode))
	}

	cInfos := (*[maxCArraySize]C.struct_dcmi_ub_ping_mesh_info)(cInfosPtr)[:meshReplySize:meshReplySize]
	ubList := convertCInfoToGoSlice(cInfos, int(count))

	return &common.HccspingMeshInfo{UBPingMeshInfoList: ubList}, nil
}

// DcStopUbPingMesh stops UB ping mesh (no change needed)
func (d *DcV2Manager) DcStopUbPingMesh(logicID int32, taskID uint) error {
	if !common.IsValidLogicIDOrPhyID(logicID) {
		return fmt.Errorf("logicID(%d) is invalid", logicID)
	}
	if !common.IsValidTaskID(taskID) {
		return fmt.Errorf("taskID(%d) is invalid", taskID)
	}
	if retCode := C.dcmiv2_stop_ub_ping_mesh(C.int(logicID), C.int(taskID)); retCode != common.Success {
		return fmt.Errorf("dcmi stop ub ping mesh failed logicID(%d) error code: %d", logicID, int32(retCode))
	}
	return nil
}

// DcGetUbPingMeshState gets UB ping mesh state (no change needed)
func (d *DcV2Manager) DcGetUbPingMeshState(logicID int32, taskID uint) (int, error) {
	if !common.IsValidLogicIDOrPhyID(logicID) {
		return common.RetError, fmt.Errorf("logicID(%d) is invalid", logicID)
	}
	if !common.IsValidTaskID(taskID) {
		return common.RetError, fmt.Errorf("taskID(%d) is invalid", taskID)
	}
	var state C.uint
	if retCode := C.dcmiv2_get_ub_ping_mesh_state(C.int(logicID), C.int(taskID), &state); retCode != common.Success {
		return common.RetError, fmt.Errorf("dcmi get ub ping mesh state failed logicID(%d) "+
			"error code: %d", logicID, int32(retCode))
	}
	return int(state), nil
}

func convertCgoCharArrayToString(cgoArr [DcmiVDevResNameLen]C.char) string {
	var charArr []rune
	for _, v := range cgoArr {
		if v == 0 {
			break
		}
		charArr = append(charArr, rune(v))
	}
	return string(charArr)
}

func convertVDevQueryInfoV2(cVDevQueryInfo C.struct_dcmi_vdev_query_info) common.CgoVDevQueryInfo {
	name := convertCgoCharArrayToString(cVDevQueryInfo.name)
	vDevQueryInfo := common.CgoVDevQueryInfo{
		Name:            name,
		Status:          uint32(cVDevQueryInfo.status),
		IsContainerUsed: uint32(cVDevQueryInfo.is_container_used),
		Vfid:            uint32(cVDevQueryInfo.vfid),
		VfgID:           uint32(cVDevQueryInfo.vfg_id),
		ContainerID:     uint64(cVDevQueryInfo.container_id),
		Base:            convertBaseResource(cVDevQueryInfo.base),
		Computing:       convertComputingResource(cVDevQueryInfo.computing),
		Media:           convertMediaResource(cVDevQueryInfo.media),
	}
	return vDevQueryInfo
}

func convertVDevQueryStruV2(cVDevQueryStru C.struct_dcmi_vdev_query_stru) common.CgoVDevQueryStru {
	vDevQueryStru := common.CgoVDevQueryStru{
		VDevID:    uint32(cVDevQueryStru.vdev_id),
		QueryInfo: convertVDevQueryInfoV2(cVDevQueryStru.query_info),
	}
	return vDevQueryStru
}

// DcGetDeviceVDevResource get virtual device resource info
func (d *DcV2Manager) DcGetDeviceVDevResource(logicID int32, vDevID uint32) (common.CgoVDevQueryStru, error) {
	if !common.IsValidLogicIDOrPhyID(logicID) {
		return common.CgoVDevQueryStru{}, fmt.Errorf("logicID(%d) is invalid", logicID)
	}
	var cMainCmd = C.enum_dcmi_main_cmd(MainCmdVDevMng)
	subCmd := VmngSubCmdGetVDevResource
	var vDevResource C.struct_dcmi_vdev_query_stru
	size := C.uint(unsafe.Sizeof(vDevResource))
	vDevResource.vdev_id = C.uint(vDevID)
	if retCode := C.dcmiv2_get_device_info(C.int(logicID), cMainCmd, C.uint(subCmd),
		unsafe.Pointer(&vDevResource), &size); int32(retCode) != common.Success {
		return common.CgoVDevQueryStru{}, fmt.Errorf("get device info failed, error is: %d", int32(retCode))
	}
	return convertVDevQueryStruV2(vDevResource), nil
}

// DcGetDeviceTotalResource get device total resource info
func (d *DcV2Manager) DcGetDeviceTotalResource(logicID int32) (common.CgoSocTotalResource, error) {
	if !common.IsValidLogicIDOrPhyID(logicID) {
		return common.CgoSocTotalResource{}, fmt.Errorf("logicID(%d) or deviceID(%d) is invalid", logicID)
	}
	var cMainCmd = C.enum_dcmi_main_cmd(MainCmdVDevMng)
	subCmd := VmngSubCmdGetTotalResource
	var totalResource C.struct_dcmi_soc_total_resource
	size := C.uint(unsafe.Sizeof(totalResource))
	if retCode := C.dcmiv2_get_device_info(C.int(logicID), cMainCmd, C.uint(subCmd),
		unsafe.Pointer(&totalResource), &size); int32(retCode) != common.Success {
		return common.CgoSocTotalResource{}, fmt.Errorf("get device info failed, error is: %d", int32(retCode))
	}
	if uint32(totalResource.vdev_num) > DcmiMaxVdevNum {
		return common.CgoSocTotalResource{}, fmt.Errorf("get error virtual quantity: %d",
			uint32(totalResource.vdev_num))
	}

	return convertSocTotalResource(totalResource), nil
}

// DcGetSuperPodInfo get device total resource info
func (d *DcV2Manager) DcGetSuperPodInfo(logicID int32) (common.CgoSuperPodInfo, error) {
	if !common.IsValidLogicIDOrPhyID(logicID) {
		return common.CgoSuperPodInfo{}, fmt.Errorf("logicID(%d) is invalid", logicID)
	}

	var unitType C.enum_dcmi_unit_type
	if retCode := C.dcmiv2_get_device_type(C.int(logicID), &unitType); int32(retCode) != common.Success {
		return common.CgoSuperPodInfo{}, fmt.Errorf("get device type failed, error is: %d", int32(retCode))
	}
	if int32(unitType) != common.NpuType {
		return common.CgoSuperPodInfo{}, fmt.Errorf("not support unit type: %d", int32(unitType))
	}

	var cMainCmd = C.enum_dcmi_main_cmd(MainCmdChipInf)
	subCmd := CinfSubCmdGetSPodInfo
	var sPodInfo C.struct_dcmi_spod_info
	size := C.uint(unsafe.Sizeof(sPodInfo))
	if retCode := C.dcmiv2_get_device_info(C.int(logicID), cMainCmd, C.uint(subCmd),
		unsafe.Pointer(&sPodInfo), &size); int32(retCode) != common.Success {
		return common.CgoSuperPodInfo{}, fmt.Errorf("get super pod info failed, error is: %d", int32(retCode))
	}

	return convertSuperPodInfo(sPodInfo), nil
}

// DcGetDeviceFreeResource get device free resource info
func (d *DcV2Manager) DcGetDeviceFreeResource(logicID int32) (common.CgoSocFreeResource, error) {
	if !common.IsValidLogicIDOrPhyID(logicID) {
		return common.CgoSocFreeResource{}, fmt.Errorf("logicID(%d) is invalid", logicID)
	}
	var cMainCmd = C.enum_dcmi_main_cmd(MainCmdVDevMng)
	subCmd := VmngSubCmdGetFreeResource
	var freeResource C.struct_dcmi_soc_free_resource
	size := C.uint(unsafe.Sizeof(freeResource))
	if retCode := C.dcmiv2_get_device_info(C.int(logicID), cMainCmd, C.uint(subCmd),
		unsafe.Pointer(&freeResource), &size); int32(retCode) != common.Success {
		return common.CgoSocFreeResource{}, fmt.Errorf("get device info failed, error is: %d", int32(retCode))
	}
	return convertSocFreeResource(freeResource), nil
}

// DcGetVDevActivityInfo get vir device activity info by virtual device id
func (d *DcV2Manager) DcGetVDevActivityInfo(logicID int32, vDevID uint32) (common.VDevActivityInfo, error) {
	if !common.IsValidLogicIDOrPhyID(logicID) {
		return common.VDevActivityInfo{}, fmt.Errorf("logicID(%d) is invalid", logicID)
	}
	if !common.IsValidVDevID(vDevID) {
		return common.VDevActivityInfo{}, fmt.Errorf("vDevID(%d) invalid", vDevID)
	}
	var cMainCmd = C.enum_dcmi_main_cmd(MainCmdVDevMng)
	subCmd := VmngSubCmdGetVDevActivity
	var vDevActivityInfo C.struct_dcmi_vdev_query_stru
	size := C.uint(unsafe.Sizeof(vDevActivityInfo))
	vDevActivityInfo.vdev_id = C.uint(vDevID)
	if retCode := C.dcmiv2_get_device_info(C.int(logicID), cMainCmd, C.uint(subCmd),
		unsafe.Pointer(&vDevActivityInfo), &size); int32(retCode) != common.Success {
		return common.VDevActivityInfo{}, fmt.Errorf("retCode: %d", int32(retCode))
	}
	totalMemSize := uint64(vDevActivityInfo.query_info.computing.vdev_memory_total)
	usedMemSize := totalMemSize - uint64(vDevActivityInfo.query_info.computing.vdev_memory_free)
	if usedMemSize < 0 {
		return common.VDevActivityInfo{}, errors.New("used memory value abnormal")
	}
	return common.VDevActivityInfo{
		VDevID:         vDevID,
		VDevAiCoreRate: uint32(vDevActivityInfo.query_info.computing.vdev_aicore_utilization),
		VDevTotalMem:   totalMemSize,
		VDevUsedMem:    usedMemSize,
		IsVirtualDev:   true,
	}, nil
}

// DcGetVDeviceInfo get vdevice resource info
func (d *DcV2Manager) DcGetVDeviceInfo(logicID int32) (common.VirtualDevInfo, error) {
	if !common.IsValidLogicIDOrPhyID(logicID) {
		return common.VirtualDevInfo{}, fmt.Errorf("logicID(%d) is invalid", logicID)
	}
	var unitType C.enum_dcmi_unit_type
	if retCode := C.dcmiv2_get_device_type(C.int(logicID), &unitType); int32(retCode) != common.Success {
		return common.VirtualDevInfo{}, fmt.Errorf("get device type failed, error is: %d", int32(retCode))
	}
	if int32(unitType) != common.NpuType {
		return common.VirtualDevInfo{}, fmt.Errorf("not support unit type: %d", int32(unitType))
	}

	cgoDcmiSocTotalResource, err := d.DcGetDeviceTotalResource(logicID)
	if err != nil {
		return common.VirtualDevInfo{}, fmt.Errorf("get device total resource failed, error is: %v", err)
	}

	cgoDcmiSocFreeResource, err := d.DcGetDeviceFreeResource(logicID)
	if err != nil {
		return common.VirtualDevInfo{}, fmt.Errorf("get device free resource failed, error is: %v", err)
	}
	dcmiVDevInfo := common.VirtualDevInfo{
		TotalResource: cgoDcmiSocTotalResource,
		FreeResource:  cgoDcmiSocFreeResource,
	}
	for _, vDevID := range cgoDcmiSocTotalResource.VDevID {
		cgoVDevQueryStru, err := d.DcGetDeviceVDevResource(logicID, vDevID)
		if err != nil {
			return common.VirtualDevInfo{}, fmt.Errorf("get device virtual resource failed, error is: %v", err)
		}
		dcmiVDevInfo.VDevInfo = append(dcmiVDevInfo.VDevInfo, cgoVDevQueryStru)
		vDevActivityInfo, err := d.DcGetVDevActivityInfo(logicID, vDevID)
		if err != nil {
			hwlog.RunLog.Warnf("get cur vDev's activity info failed, err: %s", err)
			continue
		}
		vDevActivityInfo.VDevAiCore = float64(cgoVDevQueryStru.QueryInfo.Computing.Aic)
		dcmiVDevInfo.VDevActivityInfo = append(dcmiVDevInfo.VDevActivityInfo, vDevActivityInfo)
	}
	return dcmiVDevInfo, nil
}

// DcSetDeviceReset set device reset
func (d *DcV2Manager) DcSetDeviceReset(logicID int32) error {
	var channelType C.enum_dcmi_reset_channel = C.INBAND_CHANNEL
	return d.setDeviceReset(logicID, channelType)
}

// DcGetOutBandChannelState get out band channel state
func (d *DcV2Manager) DcGetOutBandChannelState(logicID int32) error {
	var channelState C.int
	errCode := C.dcmiv2_get_device_outband_channel_state(C.int(logicID), &channelState)
	if errCode != common.Success {
		return fmt.Errorf("get out band channel state error, errCode: %v", errCode)
	}
	if channelState != common.ChannelStateOk {
		return fmt.Errorf("chip reset not support, channel state: %v", channelState)
	}
	return nil
}

// DcPreResetSoc pre reset soc
func (d *DcV2Manager) DcPreResetSoc(logicID int32) error {
	errCode := C.dcmiv2_pre_reset_soc(C.int(logicID))
	if errCode != common.Success {
		return fmt.Errorf("pre reset failed, cardID: %v, errCode: %v", logicID, errCode)
	}
	return nil
}

// DcSetDeviceResetOutBand set device reset out band
func (d *DcV2Manager) DcSetDeviceResetOutBand(logicID int32) error {
	var channelType C.enum_dcmi_reset_channel = C.OUTBAND_CHANNEL
	return d.setDeviceReset(logicID, channelType)
}

func (d *DcV2Manager) setDeviceReset(logicID int32, channelType C.enum_dcmi_reset_channel) error {
	if !common.IsValidLogicIDOrPhyID(logicID) {
		return fmt.Errorf("logicID(%d) is invalid", logicID)
	}
	if errCode := C.dcmiv2_set_device_reset(C.int(logicID), channelType); errCode != 0 {
		return fmt.Errorf("cardID(%d) hot reset errCode: %v", logicID, errCode)
	}
	return nil
}

// DcRescanSoc rescan soc
func (d *DcV2Manager) DcRescanSoc(logicID int32) error {
	errCode := C.dcmiv2_rescan_soc(C.int(logicID))
	if errCode != common.Success {
		return fmt.Errorf("fail to rescan chip logicID %d, errCode: %v", logicID, errCode)
	}
	return nil
}

// DcGetDeviceBootStatus get NPU boot status
func (d *DcV2Manager) DcGetDeviceBootStatus(logicID int32) (int, error) {
	if !common.IsValidLogicIDOrPhyID(logicID) {
		return common.RetError, fmt.Errorf("input invalid logicID: %d", logicID)
	}
	var bootStatus C.enum_dcmi_boot_status = C.DCMI_BOOT_STATUS_FINISH
	if errCode := C.dcmiv2_get_device_boot_status(C.int(logicID), &bootStatus); errCode != 0 {
		return common.RetError, fmt.Errorf("device boot status errCode: %v", errCode)
	}
	return int(bootStatus), nil
}

// DcGetDeviceAllErrorCode get device all error code info
func (d *DcV2Manager) DcGetDeviceAllErrorCode(logicID int32) (int32, []int64, error) {
	if !common.IsValidLogicIDOrPhyID(logicID) {
		return common.RetError, nil, fmt.Errorf("logicID(%d) is invalid", logicID)
	}
	var errCount C.int
	var errCodeArray [common.MaxErrorCodeCount]C.uint
	retCode := C.dcmiv2_get_device_errorcode(C.int(logicID), &errCount, &errCodeArray[0], common.MaxErrorCodeCount)

	var health C.uint
	healthRetCode := C.dcmiv2_get_device_health(C.int(logicID), &health)
	if int32(retCode) != common.Success && int32(healthRetCode) != common.DeviceNotReadyErrCode {
		return common.RetError, nil, fmt.Errorf("failed to obtain the device errorcode based on logicID("+
			"%d), error code: %d, error count: %d", logicID, int32(retCode), int32(errCount))
	}

	errCodes := make([]int64, 0, len(errCodeArray))
	for _, errCode := range errCodeArray {
		if int64(errCode) != 0 {
			errCodes = append(errCodes, int64(errCode))
		}
	}

	if int32(healthRetCode) == common.DeviceNotReadyErrCode {
		hwlog.RunLog.Errorf("device errorcode v2 ret code: %d, device health ret code: %d, device not ready, "+
			"maybe a card drop fault occurred on logicID(%d)", int32(retCode), int32(healthRetCode), logicID)
		errCount += 1
		errCodes = append(errCodes, common.CardDropFaultCode)
	}

	if int32(errCount) < 0 || int32(errCount) > common.MaxErrorCodeCount {
		return common.RetError, nil, fmt.Errorf("get wrong errorcode count, "+
			"logicID(%d), errorcode count: %d", logicID, int32(errCount))
	}

	return int32(errCount), errCodes, nil
}

// DcSubscribeDeviceFaultEvent subscribe device fault event
func (d *DcV2Manager) DcSubscribeDeviceFaultEvent(logicID int32) error {
	if FaultEventCallFunc == nil {
		return errors.New("callFunc is invalid, can't start subscribe")
	}

	var filter C.struct_dcmi_event_filter
	if rCode := C.dcmiv2_subscribe_fault_event(C.int(logicID), filter); int32(rCode) != common.Success {
		return fmt.Errorf("subscribe fault event failed, logicID(%d), error code: %d", logicID, int32(rCode))
	}
	return nil
}

// DcSetFaultEventCallFunc set fault event call back func
func (d *DcV2Manager) DcSetFaultEventCallFunc(businessFunc func(common.DevFaultInfo)) {
	FaultEventCallFunc = businessFunc
}

// DcGetDevProcessInfo get device process info
func (d *DcV2Manager) DcGetDevProcessInfo(logicID int32) (*common.DevProcessInfo, error) {
	if !common.IsValidLogicIDOrPhyID(logicID) {
		return nil, fmt.Errorf("logicID(%d) is invalid", logicID)
	}

	var procList [common.MaxProcNum]C.struct_dcmi_proc_mem_info
	var procNum C.int

	if retCode := C.dcmiv2_get_device_resource_info(C.int(logicID), &procList[0],
		&procNum); int32(retCode) != common.Success {
		return nil, buildDcmiV2Err(logicID, "device resource", retCode)
	}

	if int32(procNum) < 0 || int32(procNum) > common.MaxProcNum {
		return nil, fmt.Errorf("get invalid proccess num (%d), logicID(%d)", int32(procNum), logicID)
	}

	info, err := d.convertToDevResourceInfo(procList, int32(procNum))
	if err != nil {
		return nil, err
	}
	return info, nil
}

func (d *DcV2Manager) convertToDevResourceInfo(procList [common.MaxProcNum]C.struct_dcmi_proc_mem_info,
	procNum int32) (*common.DevProcessInfo, error) {
	if procNum < 0 || procNum > common.MaxProcNum {
		return nil, fmt.Errorf("process num %v is not within in the range [0~%v]", procNum, common.MaxProcNum)
	}

	info := new(common.DevProcessInfo)
	if procNum == 0 {
		return info, nil
	}

	info.ProcNum = procNum
	for i := int32(0); i < procNum; i++ {
		proc := common.DevProcInfo{
			Pid:      int32(procList[i].proc_id),
			MemUsage: float64(procList[i].proc_mem_usage) / common.UnitMB, // convert byte to MB
		}
		info.DevProcArray = append(info.DevProcArray, proc)
	}

	return info, nil
}

// DcGetPCIeBusInfo get pcie bus info
func (d *DcV2Manager) DcGetPCIeBusInfo(logicID int32) (string, error) {
	if !common.IsValidLogicIDOrPhyID(logicID) {
		return "", fmt.Errorf("logicID(%d) is invalid", logicID)
	}
	var pcieInfo C.struct_dcmi_pcie_info_all
	if retCode := C.dcmiv2_get_device_pcie_info(C.int(logicID), &pcieInfo); int32(retCode) != common.Success {
		return "", buildDcmiV2Err(logicID, "pcie bus", retCode)
	}
	info := fmt.Sprintf("%04X:%02X:%02X.%-4X", int32(pcieInfo.domain), uint32(pcieInfo.bdf_busid),
		uint32(pcieInfo.bdf_deviceid), uint32(pcieInfo.bdf_funcid))
	hwlog.RunLog.Debugf("pcie bus info is: '%s'", info)
	return strings.TrimRight(info, " "), nil
}

// DcGetDeviceBoardInfo get device board info
func (d *DcV2Manager) DcGetDeviceBoardInfo(logicID int32) (common.BoardInfo, error) {
	if !common.IsValidLogicIDOrPhyID(logicID) {
		return common.BoardInfo{}, fmt.Errorf("logicID(%d)is invalid", logicID)
	}

	var cBoardInfo C.struct_dcmi_board_info
	if retCode := C.dcmiv2_get_device_board_info(C.int(logicID), &cBoardInfo); int32(retCode) != common.Success {
		return common.BoardInfo{}, buildDcmiV2Err(logicID, "board info", retCode)
	}

	return common.BoardInfo{
		BoardId: uint32(cBoardInfo.board_id),
		PcbId:   uint32(cBoardInfo.pcb_id),
		BomId:   uint32(cBoardInfo.bom_id),
		SlotId:  uint32(cBoardInfo.slot_id),
	}, nil
}

// DcGetPCIEBandwidth get pcie bandwidth
func (d *DcV2Manager) DcGetPCIEBandwidth(logicID int32, profilingTime int) (common.PCIEBwStat, error) {
	if !common.IsValidLogicIDOrPhyID(logicID) {
		return common.PCIEBwStat{}, fmt.Errorf("logicID(%d) is invalid", logicID)
	}
	var dcmiPCIEBandwidth C.struct_dcmi_pcie_link_bandwidth_info
	var pcieBandwidth common.PCIEBwStat
	dcmiPCIEBandwidth.profiling_time = C.int(profilingTime)
	retCode := C.dcmiv2_get_pcie_link_bandwidth_info(C.int(logicID), &dcmiPCIEBandwidth)
	if int32(retCode) != common.Success {
		return pcieBandwidth, buildDcmiV2Err(logicID, "PCIEBandwidth", retCode)
	}

	pcieBandwidth.PcieRxPBw = d.convertPcieBw(dcmiPCIEBandwidth.rx_p_bw)
	pcieBandwidth.PcieRxNPBw = d.convertPcieBw(dcmiPCIEBandwidth.rx_np_bw)
	pcieBandwidth.PcieRxCPLBw = d.convertPcieBw(dcmiPCIEBandwidth.rx_cpl_bw)

	pcieBandwidth.PcieTxPBw = d.convertPcieBw(dcmiPCIEBandwidth.tx_p_bw)
	pcieBandwidth.PcieTxNPBw = d.convertPcieBw(dcmiPCIEBandwidth.tx_np_bw)
	pcieBandwidth.PcieTxCPLBw = d.convertPcieBw(dcmiPCIEBandwidth.tx_cpl_bw)

	return pcieBandwidth, nil
}

func (d *DcV2Manager) convertPcieBw(pcieBwArr [AgentdrvProfDataNum]C.uint) common.PcieStatValue {
	return common.PcieStatValue{
		PcieMinBw: int32(pcieBwArr[0]),
		PcieMaxBw: int32(pcieBwArr[1]),
		PcieAvgBw: int32(pcieBwArr[AgentdrvProfDataNum-1]),
	}
}

// DcGetDeviceEccInfo get device ecc info
func (d *DcV2Manager) DcGetDeviceEccInfo(logicID int32, inputType common.DcmiDeviceType) (*common.ECCInfo, error) {
	if !common.IsValidLogicIDOrPhyID(logicID) {
		return nil, fmt.Errorf("logicID(%d) is invalid", logicID)
	}
	dcmiDeviceType, err := d.getInputType(inputType)
	if err != nil {
		return nil, err
	}
	var deviceEccInfo C.struct_dcmi_ecc_info
	if retCode := C.dcmiv2_get_device_ecc_info(C.int(logicID), dcmiDeviceType,
		&deviceEccInfo); retCode != 0 {
		return nil, buildDcmiV2Err(logicID, "dcmi device ECC", retCode)
	}
	eccInfo := &common.ECCInfo{
		EnableFlag:                int32(deviceEccInfo.enable_flag),
		SingleBitErrorCnt:         int64(deviceEccInfo.single_bit_error_cnt),
		DoubleBitErrorCnt:         int64(deviceEccInfo.double_bit_error_cnt),
		TotalSingleBitErrorCnt:    int64(deviceEccInfo.total_single_bit_error_cnt),
		TotalDoubleBitErrorCnt:    int64(deviceEccInfo.total_double_bit_error_cnt),
		SingleBitIsolatedPagesCnt: int64(deviceEccInfo.single_bit_isolated_pages_cnt),
		DoubleBitIsolatedPagesCnt: int64(deviceEccInfo.double_bit_isolated_pages_cnt),
	}
	return eccInfo, nil
}

// DcGetSioInfo get sio info
func (d *DcV2Manager) DcGetSioInfo(logicID int32) (common.SioCrcErrStatisticInfo, error) {
	if !common.IsValidLogicIDOrPhyID(logicID) {
		return common.SioCrcErrStatisticInfo{}, fmt.Errorf("logicID(%d) is invalid", logicID)
	}
	var cMainCmd = C.enum_dcmi_main_cmd(MainCmdSio)
	subCmd := SioSubCmdCrcErrStatistics
	var sioInfo C.struct_dcmi_sio_crc_err_statistic_info
	// Use a secure function to get the address (for cleanCode)
	addr, err := getAddrWithOffset(unsafe.Pointer(&sioInfo), unsafe.Sizeof(sioInfo), 0)
	if err != nil {
		return common.SioCrcErrStatisticInfo{}, fmt.Errorf("get sioInfo addr failed, error is: %v", err)
	}
	size := C.uint(unsafe.Sizeof(sioInfo))
	if retCode := C.dcmiv2_get_device_info(C.int(logicID), cMainCmd, C.uint(subCmd),
		addr, &size); int32(retCode) != common.Success {
		return common.SioCrcErrStatisticInfo{}, buildDcmiV2Err(logicID, "super pod sio", retCode)
	}
	return convertSioInfoStruct(sioInfo), nil
}

var goInputTypeToCgoDeviceType = map[common.DcmiDeviceType]C.enum_dcmi_device_type{
	common.DcmiDeviceTypeDDR:  C.DCMI_DEVICE_TYPE_DDR,
	common.DcmiDeviceTypeSRAM: C.DCMI_DEVICE_TYPE_SRAM,
	common.DcmiDeviceTypeHBM:  C.DCMI_DEVICE_TYPE_HBM,
	common.DcmiDeviceTypeNPU:  C.DCMI_DEVICE_TYPE_NPU,
	common.DcmiDeviceTypeNONE: C.DCMI_DEVICE_TYPE_NONE,
}

func (d *DcV2Manager) getInputType(inputType common.DcmiDeviceType) (C.enum_dcmi_device_type, error) {
	if val, exist := goInputTypeToCgoDeviceType[inputType]; exist {
		return val, nil
	}
	return C.DCMI_DEVICE_TYPE_NONE, fmt.Errorf("invalid input type for getting device ecc info")
}

// DcGetDeviceMainBoardInfo get device main board info
func (d *DcV2Manager) DcGetDeviceMainBoardInfo(logicID int32) (uint32, error) {
	if !common.IsValidLogicIDOrPhyID(logicID) {
		return 0, fmt.Errorf("logicID(%d) is invalid", logicID)
	}
	var cMainBoardId C.uint
	if retCode := C.dcmiv2_get_mainboard_id(C.int(logicID), &cMainBoardId); int32(retCode) != common.Success {
		return 0, buildDcmiV2Err(logicID, "mainBoardId", retCode)
	}
	return uint32(cMainBoardId), nil
}

// DcGetCardElabel get device elabel info
func (d *DcV2Manager) DcGetCardElabel(logicID int32) (common.ElabelInfo, error) {
	if !common.IsValidLogicIDOrPhyID(logicID) {
		return common.ElabelInfo{}, fmt.Errorf("logicID(%d) is invalid", logicID)
	}
	var elabelInfo C.struct_dcmi_elabel_info
	if retCode := C.dcmiv2_get_card_elabel(C.int(logicID), &elabelInfo); int32(retCode) != common.Success {
		return common.ElabelInfo{}, fmt.Errorf("logicID(%d): get elabel info failed, error code: %v", logicID, retCode)
	}
	return common.ElabelInfo{
		ProductName:      C.GoString(&elabelInfo.product_name[0]),
		Model:            C.GoString(&elabelInfo.model[0]),
		Manufacturer:     C.GoString(&elabelInfo.manufacturer[0]),
		ManufacturerDate: C.GoString(&elabelInfo.manufacturer_date[0]),
		SerialNumber:     C.GoString(&elabelInfo.serial_number[0]),
	}, nil
}
