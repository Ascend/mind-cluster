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
#include <mockcpp/mokc.h>

#include "bmm.h"

using namespace ock::memfs;

namespace {

class TestBmm : public testing::Test {
public:
    void SetUp() override;
    void TearDown() override;
};

void TestBmm::SetUp() {}

void TestBmm::TearDown()
{
    GlobalMockObject::verify();
}

TEST_F(TestBmm, test_repeat_init_and_uninit)
{
    MemFsBMM bmm{};
    MemFsBMMOptions opt{};

    union MockerHelper {
        int32_t (MemFsBmmPool::*initialize)(const MemFsBMMOptions &opt) noexcept;
        int32_t (*mockInitialize)(MemFsBmmPool *self, const MemFsBMMOptions &opt) noexcept;

        void (MemFsBmmPool::*unInitialize)() noexcept;
        void (*mockUnInitialize)(MemFsBmmPool *self) noexcept;
    };
    MockerHelper helper{};
    helper.initialize = &MemFsBmmPool::Initialize;
    auto mocker = MOCKCPP_NS::mockAPI("&MemFsBmmPool::Initialize", helper.mockInitialize);
    mocker.defaults().will(returnValue(0));

    auto ret = bmm.Initialize(opt);
    ASSERT_EQ(0, ret);
    ret = bmm.Initialize(opt);
    ASSERT_EQ(0, ret);

    bmm.UnInitialize();
    bmm.UnInitialize();
}

TEST_F(TestBmm, test_bmm_init_with_pool_error)
{
    MemFsBMM bmm{};
    MemFsBMMOptions opt{};

    union MockerHelper {
        int32_t (MemFsBmmPool::*initialize)(const MemFsBMMOptions &opt) noexcept;
        int32_t (*mockInitialize)(MemFsBmmPool *self, const MemFsBMMOptions &opt) noexcept;
    };
    MockerHelper helper{};
    helper.initialize = &MemFsBmmPool::Initialize;
    auto mocker = MOCKCPP_NS::mockAPI("&MemFsBmmPool::Initialize", helper.mockInitialize);
    mocker.defaults().will(returnValue(1));

    auto ret = bmm.Initialize(opt);
    ASSERT_EQ(1, ret);

    bmm.UnInitialize();
}
}