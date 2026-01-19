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
#include <stdbool.h>
#include <stdio.h>
#include <stdlib.h>
#include <ctype.h>
#include <string.h>
#include <errno.h>
#include <limits.h>
#include <sys/stat.h>
#include <unistd.h>
#include <libgen.h>
#include <link.h>
#include <dlfcn.h>
#include "securec.h"
#include "basic.h"
#include "logger.h"
#include "utils.h"

#define DCMI_INIT "dcmi_init"
#define DCMI_SET_DESTROY_VDEVICE "dcmi_set_destroy_vdevice"
#define ROOT_UID 0
#define DECIMAL 10
#define DESTROY_PARAMS_NUM 4
#define PARAMS_SECOND 1
#define PARAMS_THIRD 2
#define PARAMS_FOURTH 3
#define ID_MAX 65535

static bool ShowExceptionInfo(const char* exceptionInfo)
{
    Logger(exceptionInfo, LEVEL_ERROR, SCREEN_YES);
    return false;
}

static bool CheckFileOwner(const struct stat fileStat, const bool checkOwner)
{
    if (checkOwner) {
        if ((fileStat.st_uid != ROOT_UID) && (fileStat.st_uid != geteuid())) { // 操作文件owner非root/自己
            return ShowExceptionInfo("Please check the folder owner!");
        }
    }
    return true;
}

static bool CheckParentDir(char* buf, const size_t bufLen, struct stat fileStat, const bool checkOwner)
{
    if (buf == NULL) {
        return false;
    }
    for (int iLoop = 0; iLoop < PATH_MAX; iLoop++) {
        if (!CheckFileOwner(fileStat, checkOwner)) {
            return false;
        }
        if ((fileStat.st_mode & S_IWOTH) != 0) { // 操作文件对other用户可写
            return ShowExceptionInfo("Please check the write permission!");
        }
        if ((strcmp(buf, "/") == 0) || (strstr(buf, "/") == NULL)) {
            break;
        }
        if (strcmp(dirname(buf), ".") == 0) {
            break;
        }
        if (stat(buf, &fileStat) != 0) {
            return false;
        }
    }
    return true;
}

static bool CheckLegality(const char* resolvedPath, const size_t resolvedPathLen,
    const unsigned long long maxFileSzieMb, const bool checkOwner)
{
    const unsigned long long maxFileSzieB = maxFileSzieMb * 1024 * 1024;
    char buf[PATH_MAX] = {0};
    if (strncpy_s(buf, sizeof(buf), resolvedPath, resolvedPathLen) != EOK) {
        return false;
    }
    struct stat fileStat;
    if ((stat(buf, &fileStat) != 0) ||
        ((S_ISREG(fileStat.st_mode) == 0) && (S_ISDIR(fileStat.st_mode) == 0))) {
        return ShowExceptionInfo("resolvedPath does not exist or is not a file!");
    }
    if (fileStat.st_size >= maxFileSzieB) { // 文件大小超限
        return ShowExceptionInfo("fileSize out of bounds!");
    }
    return CheckParentDir(buf, PATH_MAX, fileStat, checkOwner);
}

static bool IsAValidChar(const char c)
{
    if (isalnum(c) != 0) {
        return true;
    }
    // ._-/~为合法字符
    if ((c == '.') || (c == '_') ||
        (c == '-') || (c == '/') || (c == '~')) {
        return true;
    }
    return false;
}

static bool CheckFileName(const char* filePath, const size_t filePathLen)
{
    int iLoop;
    if ((filePathLen > PATH_MAX) || (filePathLen <= 0)) { // 长度越界
        return ShowExceptionInfo("filePathLen out of bounds!");
    }
    for (iLoop = 0; iLoop < filePathLen; iLoop++) {
        if (!IsAValidChar(filePath[iLoop])) { // 非法字符
            return ShowExceptionInfo("filePath has an illegal character!");
        }
    }
    return true;
}

static bool CheckAExternalFile(const char* filePath, const size_t filePathLen,
    const size_t maxFileSzieMb, const bool checkOwner)
{
    if (filePath == NULL) {
        return false;
    }
    if (!CheckFileName(filePath, filePathLen)) {
        return false;
    }
    char resolvedPath[PATH_MAX] = {0};
    if (realpath(filePath, resolvedPath) == NULL && errno != ENOENT) {
        return ShowExceptionInfo("realpath failed!");
    }
    if (strstr(resolvedPath, filePath) == NULL) { // 存在软链接
        return ShowExceptionInfo("filePath has a soft link!");
    }
    return CheckLegality(resolvedPath, strlen(resolvedPath), maxFileSzieMb, checkOwner);
}

static bool DeclareDcmiApiAndCheck(void **handle)
{
    *handle = dlopen("libdcmi.so", RTLD_LAZY);
    if (*handle == NULL) {
        Logger("dlopen failed.", LEVEL_ERROR, SCREEN_YES);
        return false;
    }
    struct link_map *pLinkMap;
    int ret = dlinfo(*handle, RTLD_DI_LINKMAP, &pLinkMap);
    if (ret == 0) {
        const size_t maxFileSzieMb = 10; // max 10 mb
        if (!CheckAExternalFile(pLinkMap->l_name, strlen(pLinkMap->l_name), maxFileSzieMb, true)) {
            Logger("check sofile failed.", LEVEL_ERROR, SCREEN_YES);
            return false;
        }
    } else {
        Logger("dlinfo sofile failed.", LEVEL_ERROR, SCREEN_YES);
        return false;
    }

    return true;
}

static void DcmiDlAbnormalExit(void **handle, const char* errorInfo)
{
    Logger(errorInfo, LEVEL_INFO, SCREEN_YES);
    if (*handle != NULL) {
        dlclose(*handle);
        *handle = NULL;
    }
}

static void DcmiDlclose(void **handle)
{
    if (*handle != NULL) {
        dlclose(*handle);
        *handle = NULL;
    }
}

static bool CheckLimitId(const int idValue)
{
    if (idValue < 0 || idValue > ID_MAX) {
        return false;
    }
    return true;
}

static bool GetAndCheckID(const char *argv[], int *cardId,
                          int *deviceId, int *vDeviceId)
{
    const int decimal = 10;
    errno = 0;
    char *endPtr = NULL;
    *cardId = strtol(argv[PARAMS_SECOND], &endPtr, decimal);
    if ((errno != 0) || *endPtr != '\0' || !CheckLimitId(*cardId)) {
        return false;
    }
    *deviceId = strtol(argv[PARAMS_THIRD], &endPtr, decimal);
    if ((errno != 0) || *endPtr != '\0' || !CheckLimitId(*deviceId)) {
        return false;
    }
    *vDeviceId = strtol(argv[PARAMS_FOURTH], &endPtr, decimal);
    if ((errno != 0) || *endPtr != '\0' || !CheckLimitId(*vDeviceId)) {
        return false;
    }
    return true;
}

static bool DcmiInitProcess(void *handle)
{
    if (handle == NULL) {
        return false;
    }
    int (*dcmiInit)(void) = NULL;
    dcmiInit = dlsym(handle, DCMI_INIT);
    if (dcmiInit == NULL) {
        DcmiDlAbnormalExit(&handle, "DeclareDlApi failed");
        return false;
    }
    int ret = dcmiInit();
    if (ret != 0) {
        Logger("dcmiInit failed.", LEVEL_ERROR, SCREEN_YES);
        DcmiDlclose(&handle);
        return false;
    }
    return true;
}

static bool DcmiDestroyProcess(void *handle, const int cardId,
                               const int deviceId, const int vDeviceId)
{
    if (handle == NULL) {
        return false;
    }
    int (*dcmiSetDestroyVdevice)(int, int, int) = NULL;
    dcmiSetDestroyVdevice = dlsym(handle, DCMI_SET_DESTROY_VDEVICE);
    if (dcmiSetDestroyVdevice == NULL) {
        DcmiDlAbnormalExit(&handle, "DeclareDlApi failed");
        return false;
    }
    int ret = dcmiSetDestroyVdevice(cardId, deviceId, vDeviceId);
    if (ret != 0) {
        Logger("dcmiSetDestroyVdevice failed.", LEVEL_ERROR, SCREEN_YES);
        DcmiDlclose(&handle);
        return false;
    }
    return true;
}

static int DestroyEntrance(const char *argv[])
{
    if (argv == NULL) {
        return -1;
    }
    int cardId = 0;
    int deviceId = 0;
    int vDeviceId = 0;
    char *str = FormatLogMessage("start to destroy v-device %d start...", vDeviceId);
    Logger(str, LEVEL_INFO, SCREEN_YES);
    free(str);
    if (!GetAndCheckID(argv, &cardId, &deviceId, &vDeviceId)) {
        return -1;
    }

    void *handle = NULL;
    if (!DeclareDcmiApiAndCheck(&handle)) {
        Logger("Declare dcmi failed.", LEVEL_ERROR, SCREEN_YES);
        return -1;
    }
    if (!DcmiInitProcess(handle)) {
        return -1;
    }
    if (!DcmiDestroyProcess(handle, cardId, deviceId, vDeviceId)) {
        return -1;
    }
    DcmiDlclose(&handle);
    char *strEnd = FormatLogMessage("destroy v-device %d successfully", vDeviceId);
    Logger(strEnd, LEVEL_INFO, SCREEN_YES);
    free(strEnd);
    return 0;
}

static bool EntryCheck(const int argc, const char *argv[])
{
    if (argc != DESTROY_PARAMS_NUM) {
        Logger("destroy params namber error.", LEVEL_ERROR, SCREEN_YES);
        return false;
    }
    for (int iLoop = 1; iLoop < argc; iLoop++) {
        for (size_t jLoop = 0; jLoop < strlen(argv[iLoop]); jLoop++) {
            if (isdigit(argv[iLoop][jLoop]) == 0) {
                return false;
            }
        }
    }
    return true;
}

int main(const int argc, const char *argv[])
{
    if (!EntryCheck(argc, argv)) {
        Logger("destroy params value error.", LEVEL_ERROR, SCREEN_YES);
        return -1;
    }

    return DestroyEntrance(argv);
}