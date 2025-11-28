/*
* Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved.
*/

#ifndef OCK_TTP_TEST_C_API_H
#define OCK_TTP_TEST_C_API_H
#include "gtest/gtest.h"
class CApiDtFuzz : public testing::Test {
public:
  static void SetUpTestSuite();
  static void TearDownTestSuite();
};
#endif // OCK_TTP_TEST_C_API_H