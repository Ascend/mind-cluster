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

#include "backup_file_view.h"

using namespace ock::bg::backup;

TEST(TestBackupFileView, Initialize)
{
    FileMeta meta;
    meta.inode = 1;
    meta.mtime = { 0, 0 };
    meta.lastBackupTime = { 0, 0 };

    BackupFileView view;
    std::string path = "/opt/aaa";
    FileMeta old;
    auto ret = view.AddFile(path, meta, old);
    ASSERT_EQ(ret, true);

    ret = view.AddFile(path, meta, old);
    ASSERT_EQ(ret, false);

    ret = view.GetFile(path, old);
    ASSERT_EQ(ret, true);

    ret = view.RefreshBackupTime(path, meta.inode);
    ASSERT_EQ(ret, true);

    ret = view.UpdateFile(path, meta.inode, meta);
    ASSERT_EQ(ret, true);

    ret = view.RemoveFile(path, meta.inode);
    ASSERT_EQ(ret, true);
}