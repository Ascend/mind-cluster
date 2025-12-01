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
#include "byte_buffer.h"

using namespace ock::ufs::utils;

namespace {

TEST(TestByteBuffer, wrapped_byte_buffer)
{
    uint8_t buf[64];
    ByteBuffer bb{ buf, sizeof(buf) };

    ASSERT_TRUE(bb.Valid());

    auto data = bb.Data();
    ASSERT_TRUE(data == buf);

    auto capacity = bb.Capacity();
    ASSERT_EQ(sizeof(buf), capacity);

    auto offset = bb.Offset();
    ASSERT_EQ(0UL, offset);

    const auto pos = 16UL;
    bb.Offset(pos);
    offset = bb.Offset();
    ASSERT_EQ(pos, offset);

    bb.Offset(1024UL);
    offset = bb.Offset();
    ASSERT_EQ(sizeof(buf), offset);

    bb.Offset(pos);
    bb.AddOffset(60);
    offset = bb.Offset();
    ASSERT_EQ(sizeof(buf), offset);
}
}
