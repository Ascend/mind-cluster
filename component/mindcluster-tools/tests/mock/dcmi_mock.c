/*
# Perform test mindcluster_tools
# Copyright(C) Huawei Technologies Co.,Ltd. 2025. All rights reserved.
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
# http://www.apache.org/licenses/LICENSE-2.0
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
*/

#include <stdio.h>
#include <stdlib.h>
#include <string.h>


#define NPU_COUNT (8)
#define MOCK_URMA_DEV_CNT (4)
#define DCMI_URMA_EID_SIZE (16)
#define DCMI_URMA_EID_MAX_COUNT (32)
#define MAX_URMA_DEV_CNT (8)
#define PRODUCT_TYPE_CNT (3)
#define MAX_EID_CNT (32)
#define DEVICE_ID_MAX (1)
#define PRODUCT_TYPE_ENV "PRODUCT_TYPE"

#define DLL_PUBLIC __attribute__((visibility("default")))

enum product_type {
    SUPER_POD_1D = 1,
    SUPER_POD_2D = 2,
    SERVER       = 3,
};

typedef struct {
    enum product_type pt;
    int npu_count;
}NPU_COUNT_MAP;

static NPU_COUNT_MAP g_npu_count[PRODUCT_TYPE_CNT] = {
    {SUPER_POD_1D, 3},
    {SUPER_POD_2D, 3},
    {SERVER,       3},
};

typedef struct {
    int dev_index;
    int eid_cnt;
    unsigned char eid[MAX_EID_CNT][DCMI_URMA_EID_SIZE];
}urma_device;

typedef struct {
    int product_type;
    urma_device dev[MAX_URMA_DEV_CNT];
}NPU_EID_MAP;

static NPU_EID_MAP g_eid_map[PRODUCT_TYPE_CNT] = {
    {SUPER_POD_1D,
        {
            {0, 11, {
                        {0,0,0,0,0,0,0,0x20,0,0x10,0,0,0xdf,0,0,0x01},
                        {0,0,0,0,0,0,0,0x20,0,0x10,0,0,0xdf,0,0,0x02},
                        {0,0,0,0,0,0,0,0x20,0,0x10,0,0,0xdf,0,0,0x03},
                        {0,0,0,0,0,0,0,0x20,0,0x10,0,0,0xdf,0,0,0x04},
                        {0,0,0,0,0,0,0,0x20,0,0x10,0,0,0xdf,0,0,0x05},
                        {0,0,0,0,0,0,0,0x20,0,0x10,0,0,0xdf,0,0,0x06},
                        {0,0,0,0,0,0,0,0x20,0,0x10,0,0,0xdf,0,0,0x07},
                        {0,0,0,0,0,0,0,0x20,0,0x10,0,0,0xdf,0,0,0x08},
                        {0,0,0,0,0,0,0,0x20,0,0x10,0,0,0xdf,0,0,0x09},
                        {0,0,0,0,0,0,0,0x20,0,0x10,0,0,0xdf,0,0,0xb5},
                        {0,0,0,0,0,0,0,0x20,0,0x10,0,0,0xdf,0,0,0xb7},
                    }
            },
            {0, 2, {
                        {0,0,0,0,0,0,0,0x20,0,0x10,0,0,0xdf,0,0,0xb6},
                        {0,0,0,0,0,0,0,0x20,0,0x10,0,0,0xdf,0,0,0xb8},
                   }
            },
        }
    },
    {
        SUPER_POD_2D,
        {
            {0, 20, {
                        {0,0,0,0,0,0,0,0x20,0,0x10,0,0,0xdf,0,0,0x01},
                        {0,0,0,0,0,0,0,0x20,0,0x10,0,0,0xdf,0,0,0x02},
                        {0,0,0,0,0,0,0,0x20,0,0x10,0,0,0xdf,0,0,0x03},
                        {0,0,0,0,0,0,0,0x20,0,0x10,0,0,0xdf,0,0,0x04},
                        {0,0,0,0,0,0,0,0x20,0,0x10,0,0,0xdf,0,0,0x05},
                        {0,0,0,0,0,0,0,0x20,0,0x10,0,0,0xdf,0,0,0x06},
                        {0,0,0,0,0,0,0,0x20,0,0x10,0,0,0xdf,0,0,0x07},
                        {0,0,0,0,0,0,0,0x20,0,0x10,0,0,0xdf,0,0,0x08},
                        {0,0,0,0,0,0,0,0x20,0,0x10,0,0,0xdf,0,0,0x09},
                        {0,0,0,0,0,0,0,0x20,0,0x10,0,0,0xdf,0,0,0x0a},
                        {0,0,0,0,0,0,0,0x20,0,0x10,0,0,0xdf,0,0,0x0b},
                        {0,0,0,0,0,0,0,0x20,0,0x10,0,0,0xdf,0,0,0x0c},
                        {0,0,0,0,0,0,0,0x20,0,0x10,0,0,0xdf,0,0,0x0d},
                        {0,0,0,0,0,0,0,0x20,0,0x10,0,0,0xdf,0,0,0x0e},
                        {0,0,0,0,0,0,0,0x20,0,0x10,0,0,0xdf,0,0,0x0f},
                        {0,0,0,0,0,0,0,0x20,0,0x10,0,0,0xdf,0,0,0x10},
                        {0,0,0,0,0,0,0,0x20,0,0x10,0,0,0xdf,0,0,0x11},
                        {0,0,0,0,0,0,0,0x20,0,0x10,0,0,0xdf,0,0,0x12},
                        {0,0,0,0,0,0,0,0x20,0,0x10,0,0,0xdf,0,0,0xb5},
                        {0,0,0,0,0,0,0,0x20,0,0x10,0,0,0xdf,0,0,0xb7},
                   }
            },
            {0, 2, {
                        {0,0,0,0,0,0,0,0x20,0,0x10,0,0,0xdf,0,0,0xb6},
                        {0,0,0,0,0,0,0,0x20,0,0x10,0,0,0xdf,0,0,0xb8},
                   }
            },
        }
    },
};


typedef union dcmi_urma_eid {
    unsigned char raw[DCMI_URMA_EID_SIZE];
    struct {
        unsigned long reserved;
        unsigned int prefix;
        unsigned int addr;
    }in4;
    struct {
        unsigned long subnet_prefix;
        unsigned long interface_id;
    }in6;
}dcmi_urma_eid_t;

typedef struct dcmi_urma_eid_info{
    dcmi_urma_eid_t eid;
    unsigned int eid_index;
}dcmi_urma_eid_info_t;


enum dcmi_main_cmd {
    DCMI_MAIN_CMD_DVPP = 0,
    DCMI_MAIN_CMD_CHIP_INF = 12,
    DCMI_MAIN_CMD_MAX
};

typedef enum {
    DCMI_CHIP_INF_SUB_CMD_CHIP_ID,
    DCMI_CHIP_INF_SUB_CMD_SPOD_INFO,
    DCMI_CHIP_INF_SUB_CMD_SPOD_NODE_STATUS,
    DCMI_CHIP_INF_SUB_CMD_MAX = 0xFF,
}DCMI_CHIP_INFO_SUB_CMD;

struct dcmi_spod_info {
    unsigned int sdid;
    unsigned int super_pod_size;
    unsigned int super_pod_id;
    unsigned int server_index;
    unsigned int chassis_id;
    unsigned char super_pod_type;
    unsigned char reserve[27];
};

DLL_PUBLIC int dcmi_init()
{
    return 0;
}

static int get_product_type()
{
    return atoi(getenv(PRODUCT_TYPE_ENV));
}

DLL_PUBLIC int dcmi_get_card_list(int *card_num,
                             int *card_list,
                             int list_len)
{
    *card_num = NPU_COUNT;
    if (list_len < 0) {
        return -1;
    }
    for (int i = 0; i < NPU_COUNT; i++) {
        card_list[i] = i;
    }
    return 0;
}

DLL_PUBLIC int dcmi_get_urma_device_cnt(int card_id,
                             int device_id,
                             unsigned int* dev_cnt)
{
    if (dev_cnt == NULL) {
        return -1;
    }
    if (card_id < 0 || card_id >= NPU_COUNT) {
        return -1;
    }
    if (device_id < 0 || device_id >= DEVICE_ID_MAX) {
        return -1;
    }
    int i = 0;
    for (i = 0; i < PRODUCT_TYPE_CNT; ++i) {
        if (g_npu_count[i].pt == get_product_type()) {
            *dev_cnt = MOCK_URMA_DEV_CNT;
        }
    }
    return 0;
}

DLL_PUBLIC int dcmi_get_eid_list_by_urma_dev_index(int card_id,
                                        int device_id,
                                        unsigned int dev_index,
                                        dcmi_urma_eid_info_t* eid_list,
                                        unsigned int* eid_cnt)
{
    if (card_id < 0 || card_id >= NPU_COUNT) {
        return -1;
    }
    if (device_id < 0 || device_id >= DEVICE_ID_MAX) {
        return -1;
    }
    if (eid_list == NULL || eid_cnt == NULL) {
        return -1;
    }
    int product_type = get_product_type();
    for (int i = 0; i < PRODUCT_TYPE_CNT; ++i) {
        if (g_eid_map[i].product_type == product_type) {
            *eid_cnt = g_eid_map[i].dev->eid_cnt;
            for (int j = 0; j < *eid_cnt; ++j) {
                eid_list[j].eid_index = j;
                for (int k = 0; k < DCMI_URMA_EID_SIZE; ++k) {
                    eid_list[j].eid.raw[k] = g_eid_map[i].dev[dev_index].eid[j][k];
                }
            }
        }
    }
    return 0;
}

DLL_PUBLIC int dcmi_get_all_device_count(int *all_device_count)
{
    if (all_device_count == NULL) {
        return -1;
    }
    *all_device_count = NPU_COUNT;
    return 0;
}

DLL_PUBLIC int dcmi_get_device_id_in_card(int card_id, int *device_id_max, int *mcu_id, int *cpu_id)
{
    if (card_id < 0 || card_id >= NPU_COUNT) {
        return -1;
    }
    if (device_id_max == NULL || mcu_id == NULL || cpu_id == NULL) {
        return - 1;
    }
    *device_id_max = DEVICE_ID_MAX;
    *mcu_id = 0;
    *cpu_id = 0;
    return 0;
}

DLL_PUBLIC int dcmi_get_device_info(
        int card_id,
        int device_id,
        enum dcmi_main_cmd main_cmd,
        unsigned int sub_cmd,
        void *buf ,
        unsigned *size)
{
    if (main_cmd == DCMI_MAIN_CMD_CHIP_INF && sub_cmd == DCMI_CHIP_INF_SUB_CMD_SPOD_INFO) {
        *size = sizeof(struct dcmi_spod_info);
        struct dcmi_spod_info *spinfo = (struct dcmi_spod_info*)buf;
        spinfo->super_pod_id = atoi(getenv("MOCK_SPOD_ID"));
        spinfo->super_pod_size = atoi(getenv("MOCK_SPOD_SIZE"));
        spinfo->chassis_id = atoi(getenv("MOCK_CHASSIS_ID"));
        spinfo->super_pod_type = (unsigned char)get_product_type();
    }
    return 0;
}
