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
#include "u_mount.h"

#include <errno.h>
#include <limits.h>
#include <stdlib.h>
#include <sys/mount.h>
#include <sys/stat.h>

#include "basic.h"
#include "logger.h"
#include "options.h"
#include "securec.h"
#include "utils.h"

STATIC bool CheckSrcFile(const char *src)
{
    struct stat fileStat;
    if (lstat(src, &fileStat) != 0)
    {
        return false;  // src file to be mounted does not exist
    }
    if ((S_ISREG(fileStat.st_mode) != 0) || (S_ISDIR(fileStat.st_mode) != 0))
    {                                            // only regular files and directories are checked
        const size_t maxFileSzieMb = 10 * 1024;  // max 10 G
        if (!CheckExternalFile(src, strlen(src), maxFileSzieMb, false))
        {
            char *str = FormatLogMessage("failed to mount src: %s.", src);
            Logger(str, LEVEL_ERROR, SCREEN_YES);
            free(str);
            str = NULL;
            return false;
        }
    }
    if (S_ISDIR(fileStat.st_mode) != 0)
    {  // for directories, recursively check the file subset
        if (!GetFileSubsetAndCheck(src, strlen(src)))
        {
            char *str = FormatLogMessage("Check file subset failed: %s.", src);
            Logger(str, LEVEL_ERROR, SCREEN_YES);
            free(str);
            str = NULL;
            return false;
        }
    }
    return true;
}

int Mount(const char *src, const char *dst)
{
    if (src == NULL || dst == NULL)
    {
        Logger("src pointer or dst pointer is null!", LEVEL_ERROR, SCREEN_YES);
        return -1;
    }

    static const unsigned long MOUNT_FLAGS = MS_BIND;
    static const unsigned long REMOUNT_FLAGS = MS_BIND | MS_REMOUNT | MS_RDONLY | MS_NOSUID;
    if (!CheckSrcFile(src))
    {
        return -1;
    }
    int ret = mount(src, dst, NULL, MOUNT_FLAGS, NULL);
    if (ret < 0)
    {
        Logger("failed to mount src.", LEVEL_ERROR, SCREEN_YES);
        return -1;
    }

    ret = mount(NULL, dst, NULL, REMOUNT_FLAGS, NULL);
    if (ret < 0)
    {
        Logger("failed to re-mount. dst.", LEVEL_ERROR, SCREEN_YES);
        return -1;
    }

    return 0;
}

STATIC int MountFile(const char *rootfs, const char *src, const char *dstInContainer)
{
    if (rootfs == NULL || src == NULL || dstInContainer == NULL)
    {
        Logger("rootfs, src, dstInContainer pointer is null!", LEVEL_ERROR, SCREEN_YES);
        return -1;
    }

    int ret;
    char dst[BUF_SIZE] = {0};

    // fall back to same-path mount when dst is empty
    const char *rel = (dstInContainer[0] != '\0') ? dstInContainer : src;
    if (strlen(rootfs) + strlen(rel) >= BUF_SIZE)
    {
        char *str = FormatLogMessage("rootfs plus mount path %s exceeds the assembly buffer %d.", rel, BUF_SIZE);
        Logger(str, LEVEL_ERROR, SCREEN_YES);
        free(str);
        return -1;
    }
    ret = sprintf_s(dst, BUF_SIZE, "%s%s", rootfs, rel);
    if (ret < 0)
    {
        Logger("failed to assemble file mounting path.", LEVEL_ERROR, SCREEN_YES);
        return -1;
    }

    struct stat srcStat;
    ret = stat(src, &srcStat);
    if (ret < 0)
    {
        return 0;
    }

    ret = MakeMountPoints(dst, srcStat.st_mode);
    if (ret < 0)
    {
        Logger("failed to create mount dst file.", LEVEL_ERROR, SCREEN_YES);
        return -1;
    }

    ret = Mount(src, dst);
    if (ret < 0)
    {
        Logger("failed to mount dev.", LEVEL_ERROR, SCREEN_YES);
        return -1;
    }

    return 0;
}

int MountDir(const char *rootfs, const char *src, const char *dstInContainer)
{
    if (rootfs == NULL || src == NULL || dstInContainer == NULL)
    {
        Logger("rootfs, src, dstInContainer pointer is null!", LEVEL_ERROR, SCREEN_YES);
        return -1;
    }

    int ret;
    char dst[BUF_SIZE] = {0};

    // fall back to same-path mount when dst is empty
    const char *rel = (dstInContainer[0] != '\0') ? dstInContainer : src;
    if (strlen(rootfs) + strlen(rel) >= BUF_SIZE)
    {
        char *str = FormatLogMessage("rootfs plus mount path %s exceeds the assembly buffer %d.", rel, BUF_SIZE);
        Logger(str, LEVEL_ERROR, SCREEN_YES);
        free(str);
        return -1;
    }
    ret = sprintf_s(dst, BUF_SIZE, "%s%s", rootfs, rel);
    if (ret < 0)
    {
        Logger("failed to assemble dir mounting path.", LEVEL_ERROR, SCREEN_YES);
        return -1;
    }

    struct stat srcStat;
    ret = stat(src, &srcStat);
    if (ret < 0)
    {
        return 0;
    }

    ret = MakeDirWithParent(dst, DEFAULT_DIR_MODE);
    if (ret < 0)
    {
        char *str = FormatLogMessage("failed to make dir: %s.", dst);
        Logger(str, LEVEL_ERROR, SCREEN_YES);
        free(str);
        str = NULL;
        return -1;
    }

    ret = Mount(src, dst);
    if (ret < 0)
    {
        char *str = FormatLogMessage("failed to mount src dir: %s to dst dir: %s.", src, dst);
        Logger(str, LEVEL_ERROR, SCREEN_YES);
        free(str);
        str = NULL;
        return -1;
    }

    return 0;
}

int DoDirectoryMounting(const char *rootfs, const struct MountList *list)
{
    if (rootfs == NULL || list == NULL)
    {
        Logger("rootfs, list pointer is null!", LEVEL_ERROR, SCREEN_YES);
        return -1;
    }

    int ret;
    for (unsigned int i = 0; i < list->count; i++)
    {
        ret = MountDir(rootfs, list->list[i].src, list->list[i].dst);
        if (ret < 0)
        {
            Logger("failed to do directory mounting", LEVEL_ERROR, SCREEN_YES);
            return -1;
        }
    }

    return 0;
}

int DoFileMounting(const char *rootfs, const struct MountList *list)
{
    if (rootfs == NULL || list == NULL)
    {
        Logger("rootfs, list pointer is null!", LEVEL_ERROR, SCREEN_YES);
        return -1;
    }

    int ret;
    for (unsigned int i = 0; i < list->count; i++)
    {
        ret = MountFile(rootfs, list->list[i].src, list->list[i].dst);
        if (ret < 0)
        {
            Logger("failed to do file mounting for.", LEVEL_ERROR, SCREEN_YES);
            return -1;
        }
    }

    return 0;
}

int DoMounting(const struct ParsedConfig *config)
{
    if (config == NULL)
    {
        Logger("config pointer is null!", LEVEL_ERROR, SCREEN_YES);
        return -1;
    }

    int ret;

    if (IsOptionNoDrvSet())
    {
        return 0;
    }

    ret = DoFileMounting(config->rootfs, config->files);
    if (ret < 0)
    {
        Logger("failed to mount files.", LEVEL_ERROR, SCREEN_YES);
        return -1;
    }

    ret = DoDirectoryMounting(config->rootfs, config->dirs);
    if (ret < 0)
    {
        Logger("failed to do mount directories.", LEVEL_ERROR, SCREEN_YES);
        return -1;
    }

    return 0;
}
