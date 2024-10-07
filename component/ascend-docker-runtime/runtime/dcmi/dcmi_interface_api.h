/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2022. All rights reserved.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
#ifndef __DCMI_INTERFACE_API_H__
#define __DCMI_INTERFACE_API_H__
#include <stddef.h>
#define _GNU_SOURCE
#include <link.h>
#include <dlfcn.h>
#include <limits.h>
#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#ifdef __cplusplus
#if __cplusplus
extern "C" {
#endif
#endif /* __cplusplus */

void *g_dcmiHandle;
#define SO_NOT_FOUND  (-99999)
#define FUNCTION_NOT_FOUND  (-99998)
#define SUCCESS  (0)
#define ERROR_UNKNOWN  (-99997)
#define SO_NOT_CORRECT  (-99996)
#define CALL_FUNC(name, ...) if (name == NULL) {return FUNCTION_NOT_FOUND;}return name(__VA_ARGS__)
#define DCMI_VDEV_FOR_RESERVE (32)
#define MAX_CHIP_NAME_LEN (32)
struct DcmiCreateVdevOut {
    unsigned int vdevId;
    unsigned int pcieBus;
    unsigned int pcieDevice;
    unsigned int pcieFunc;
    unsigned int vfgId;
    unsigned char reserved[DCMI_VDEV_FOR_RESERVE];
};
struct DcmiCreateVdevResStru {
    unsigned int vdevId;
    unsigned int vfgId;
    char templateName[32];
    unsigned char reserved[64];
};

struct DcmiChipInfo {
unsigned char chipType[MAX_CHIP_NAME_LEN];
unsigned char chipName[MAX_CHIP_NAME_LEN];
unsigned char chipVer[MAX_CHIP_NAME_LEN];
unsigned int aicoreCnt;
};

// dcmi
int (*g_dcmiInitFunc)(void);
static int DcmiInit(void)
{
    CALL_FUNC(g_dcmiInitFunc);
}

int (*g_dcmiGetCardNumListFunc)(int *cardNum, int *cardList, int listLength);
static int DcmiGetCardNumList(int *cardNum, int *cardList, int listLength)
{
    CALL_FUNC(g_dcmiGetCardNumListFunc, cardNum, cardList, listLength);
}

int (*g_dcmiGetDeviceNumInCardFunc)(int cardId, int *deviceNum);
static int DcmiGetDeviceNumInCard(int cardId, int *deviceNum)
{
    CALL_FUNC(g_dcmiGetDeviceNumInCardFunc, cardId, deviceNum);
}

int (*g_dcmiGetDeviceLogicIdFunc)(int *deviceLogicId, int cardId, int deviceId);
static int DcmiGetDeviceLogicId(int *deviceLogicId, int cardId, int deviceId)
{
    CALL_FUNC(g_dcmiGetDeviceLogicIdFunc, deviceLogicId, cardId, deviceId);
}

int (*g_dcmiCreateVdeviceFunc)(int cardId, int deviceId,
                               struct DcmiCreateVdevResStru *vdev,
                               struct DcmiCreateVdevOut *out);
static int DcmiCreateVdevice(int cardId, int deviceId,
                      struct DcmiCreateVdevResStru *vdev,
                      struct DcmiCreateVdevOut *out)
{
    CALL_FUNC(g_dcmiCreateVdeviceFunc, cardId, deviceId, vdev, out);
}

int (*g_dcmiSetDestroyVdeviceFunc)(int cardId, int deviceId, unsigned int vDevid);
static int DcmiSetDestroyVdevice(int cardId, int deviceId, unsigned int vDevid)
{
    CALL_FUNC(g_dcmiSetDestroyVdeviceFunc, cardId, deviceId, vDevid);
}

int (*g_dcmiGetDeviceLogicidFromPhyidFunc)(unsigned int phyid, unsigned int *logicid);
static int DcmiGetDeviceLogicidFromPhyid(unsigned int phyid, unsigned int *logicid)
{
    CALL_FUNC(g_dcmiGetDeviceLogicidFromPhyidFunc, phyid, logicid);
}

int (*g_dcmiGetProductTypeFunc)(int cardId, int deviceId, char *productTypeStr, int bufSize);
static int DcmiGetProductType(int cardId, int deviceId, char *productTypeStr, int bufSize)
{
    CALL_FUNC(g_dcmiGetProductTypeFunc, cardId, deviceId, productTypeStr, bufSize);
}

int (*g_dcmiGetDeviceChipInfoFunc)(int cardId, int deviceId, struct DcmiChipInfo *chipInfo);
static int DcmiGetDeviceChipInfo(int cardId, int deviceId, struct DcmiChipInfo *chipInfo)
{
    CALL_FUNC(g_dcmiGetDeviceChipInfoFunc, cardId, deviceId, chipInfo);
}

// load .so files and functions
static int DcmiInitDl(char *dlPath)
{
    g_dcmiHandle = dlopen("libdcmi.so", RTLD_LAZY | RTLD_GLOBAL);
    if (g_dcmiHandle == NULL) {
        fprintf(stderr, "%s\n", dlerror());
        return SO_NOT_FOUND;
    }
    struct link_map *pLinkMap;
    int ret = dlinfo(g_dcmiHandle, RTLD_DI_LINKMAP, &pLinkMap);
    if (ret != 0) {
        fprintf(stderr, "dlinfo sofile failed :%s\n", dlerror());
        return SO_NOT_CORRECT;
    }

    size_t pathSize = strlen(pLinkMap->l_name);
    for (int i = 0; i < pathSize && i < PATH_MAX; i++) {
        dlPath[i] = pLinkMap->l_name[i];
    }

    g_dcmiInitFunc = dlsym(g_dcmiHandle, "dcmi_init");

    g_dcmiGetCardNumListFunc = dlsym(g_dcmiHandle, "dcmi_get_card_num_list");

    g_dcmiGetDeviceNumInCardFunc = dlsym(g_dcmiHandle, "dcmi_get_device_num_in_card");

    g_dcmiGetDeviceLogicIdFunc = dlsym(g_dcmiHandle, "dcmi_get_device_logic_id");

    g_dcmiCreateVdeviceFunc = dlsym(g_dcmiHandle, "dcmi_create_vdevice");

    g_dcmiSetDestroyVdeviceFunc = dlsym(g_dcmiHandle, "dcmi_set_destroy_vdevice");

    g_dcmiGetDeviceLogicidFromPhyidFunc = dlsym(g_dcmiHandle, "dcmi_get_device_logicid_from_phyid");

    g_dcmiGetProductTypeFunc = dlsym(g_dcmiHandle, "dcmi_get_product_type");

    g_dcmiGetDeviceChipInfoFunc = dlsym(g_dcmiHandle, "dcmi_get_device_chip_info");

    return SUCCESS;
}

static int DcmiShutDown(void)
{
    if (g_dcmiHandle == NULL) {
        return SUCCESS;
    }
    return ((dlclose(g_dcmiHandle) != SUCCESS) ? ERROR_UNKNOWN : SUCCESS);
}

#ifdef __cplusplus
#if __cplusplus
}
#endif
#endif /* __cplusplus */

#endif /* __DCMI_INTERFACE_API_H__ */
