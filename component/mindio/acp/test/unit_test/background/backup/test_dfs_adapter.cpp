/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved.
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
#include <gtest/gtest.h>

#include "dfs_adapter.h"
#include "ufs_api.h"
#include "pacific_adapter.h"

using namespace ock::bg::backup;
using namespace ock::ufs;

TEST(TestDfsAdapter, Initialize)
{
    auto marker = std::make_shared<FinishedMarker>();
    DfsAdapter dfs;
    const std::string path = "/home/xxx";
    int flags;
    FileMode mode(0);
    utils::ByteBuffer dataBuffer;
    FileRange range;
    ListFileResult listRet;

    dfs.HealthyCheck();
    dfs.PutFile(path, mode, dataBuffer);
    dfs.PutFile(path, flags, mode, dataBuffer);

    dfs.GetFile(path, range);
    dfs.GetFile(path, dataBuffer, range);

    dfs.MoveFile("/home/aaa", "/home/bbb");
    dfs.CopyFile("/home/aaa", "/home/bbb");
    dfs.RemoveFile(path);
    dfs.CreateDirectory(path, mode);
    dfs.RemoveDirectory(path);

    dfs.ListFiles(path, listRet);
    dfs.ListFiles(path, listRet, marker);

    FileMeta meta;
    dfs.GetFileMeta(path, meta);
    std::map<std::string, std::string> metaMap;
    metaMap.emplace("aaa", "bbb");
    dfs.SetFileMeta(path, metaMap);
    dfs.GetFileLock(path);
}