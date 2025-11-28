/*
* Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved.
*/

#include "gtest/gtest.h"
#include "Secodefuzz/secodeFuzz.h"

GTEST_API_ int main(int argc, char **argv)
{
    char *path = "./";
    DT_Set_Report_Path(path);

    testing::InitGoogleTest(&argc, argv);
    auto ret = RUN_ALL_TESTS();
    return ret;
}