/*
* Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved.
*/
#include "gtest/gtest.h"
#include "Secodefuzz/secodeFuzz.h"

GTEST_API_ int main(int argc, char **argv)
{
    char *path = "./";
    DT_Set_Report_Path(path);
    testing::InitGoogleTest(&argc, argv);
    return RUN_ALL_TESTS();
}