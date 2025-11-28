/*
* Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved.
*/
#include <gtest/gtest.h>

#include "inode_evictor.h"
#include "evict_helper.h"

using namespace ock::memfs;

TEST(TestMemfsInodeEvictor, initialize_test)
{
    auto ret = InodeEvictor::GetInstance().Initialize();
    ASSERT_EQ(0, ret);

    InodeEvictor::GetInstance().Destroy();
}

TEST(TestMemfsInodeEvictor, evict_helper_test)
{
    list_head inodeHead { &inodeHead, &inodeHead };
    EvictHelper evictHelper;
    auto ret = evictHelper.Initialize(&inodeHead);
    ASSERT_EQ(0, ret);
    evictHelper.GetLock();
    evictHelper.AddToTail();
}
