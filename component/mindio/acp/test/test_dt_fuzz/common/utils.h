/*
* Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved.
*/

#ifndef OCKIO_UTILS_H
#define OCKIO_UTILS_H
void InitLogger();
constexpr uint32_t FUZZ_TEST_SECONDS = 3 * 3600;
constexpr uint32_t FUZZ_TEST_TIMES = 50000000;
constexpr uint32_t MAX_BUFFER_TO_WRITE = 2 * 1024 * 1024;
#endif // OCKIO_UTILS_H
