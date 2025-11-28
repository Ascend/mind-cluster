/*
* Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved.
*/

#ifndef OCKIO_TEST_MEMFS_API_H
#define OCKIO_TEST_MEMFS_API_H
#include "gtest/gtest.h"

class MemFsApiDtFuzz : public testing::Test {
public:
    static void SetUpTestSuite();
    static void TearDownTestSuite();
};
#endif // OCKIO_TEST_MEMFS_API_H
