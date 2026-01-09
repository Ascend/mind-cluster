/**
 * Copyright (c) Huawei Technologies Co., Ltd. 2026. All rights reserved.
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


#include "mem_file_system.h"
#include "inode_evictor.h"

using namespace ock::memfs;

namespace {

TEST(TestInodeEvictor, test_initialize)
{
    auto ret = InodeEvictor::GetInstance().Initialize();
    ASSERT_EQ(0, ret);

    InodeEvictor::GetInstance().Destroy();
}

TEST(TestInodeEvictor, test_recycle_inodes)
{
    auto &instance = InodeEvictor::GetInstance();
    auto ret = instance.Initialize();
    ASSERT_EQ(0, ret);
    uint64_t blockSize = 16UL << 20;
    MemFileSystem memFileSystem(blockSize, 2, "test");
    ret = memFileSystem.Initialize();
    ASSERT_EQ(0, ret);

    instance.RecycleInodes(blockSize);
    memFileSystem.Destroy();
    InodeEvictor::GetInstance().Destroy();
}
}