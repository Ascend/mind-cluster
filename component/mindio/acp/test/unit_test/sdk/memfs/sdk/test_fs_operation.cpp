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
#include <ostream>
#include <memory>

#include "fs_operation.h"

using namespace ock::memfs;

template <class OStream> OStream &operator << (OStream &stream, const FileWriteParam &param)
{
    stream << "FileWriteParam(buffer=" << (void *)param.buffers << ", index=" << param.index << ", fileOff=" <<
        param.fileOffset << ", pageIndex=" << param.pageIndex << ", pageOffset=" << param.pageOffset << ", size=" <<
        param.writeSize << " )";
    return stream;
}

namespace {

std::string ToString(const FileWriteParam &param)
{
    std::stringstream ss;
    ss << param;
    return ss.str();
}

TEST(TestFsOperation, mutil_write_split_to_slice_test)
{
    std::vector<uint64_t> sizes = { 1564122, 14748, 3137022, 26600, 1971320, 31297, 1592554, 2635143, 3484030, 529090 };
    auto totalSize = 0UL;
    std::for_each(sizes.begin(), sizes.end(), [&totalSize](uint64_t s) { totalSize += s; });
    auto totalBuffer = new char[totalSize];
    std::unique_ptr<char[]> autoRelease(totalBuffer);

    char *ptr = totalBuffer;
    std::vector<Buffer> buffers;
    for (auto size : sizes) {
        buffers.emplace_back(ptr, size);
        ptr += size;
    }

    auto sliceSize = 1048576;
    auto sliceCount = 15;
    auto fp = MakeRef<MemFsFileOpInfo>();
    ParallelRwContext context{ fp, sliceCount };
    const std::vector<FileWriteParam> param{};
    auto taskParams = context.GenerateRwParameters(param, buffers, totalSize, sliceSize);
    ASSERT_EQ(15UL, taskParams.size());

    auto totalOffset = 0UL;
    auto i = 0UL;
    for (auto &task : taskParams) {
        std::cout << "task = " << task << std::endl;
        ASSERT_EQ(&buffers, task.buffers) << ToString(task);
        ASSERT_EQ(i, task.index) << ToString(task);
        ASSERT_EQ(totalOffset, task.fileOffset) << ToString(task);
        totalOffset += task.writeSize;
        i++;
        ASSERT_LT(task.pageIndex, 10UL) << ToString(task);
        ASSERT_LT(task.pageOffset, buffers[task.pageIndex].size) << ToString(task);
    }
}

TEST(TestFsOperation, mutil_read_split_to_slice_test)
{
    std::vector<uint64_t> sizes = { 1564122, 14748, 3137022, 26600, 1971320, 31297, 1592554, 2635143, 3484030, 529090 };
    auto totalSize = 0UL;
    std::for_each(sizes.begin(), sizes.end(), [&totalSize](uint64_t s) { totalSize += s; });
    auto totalBuffer = new char[totalSize];
    std::unique_ptr<char[]> autoRelease(totalBuffer);

    char *ptr = totalBuffer;
    uint64_t start = 0UL;
    auto sliceSize = 1048576;
    auto sliceCount = 15;

    std::vector<ReadBuffer> buffers;
    for (auto size : sizes) {
        buffers.emplace_back(ptr, start, size);
        ptr += size;
        start += sliceSize;
    }
    auto fp = MakeRef<MemFsFileOpInfo>();
    ParallelRwContext context{ fp, sliceCount };
    const std::vector<FileReadParam> param{};
    auto taskParams = context.GenerateRwParameters(param, buffers, totalSize, sliceSize);
    ASSERT_EQ(15UL, taskParams.size());

    auto totalOffset = 0UL;
    auto i = 0UL;
    for (auto &task : taskParams) {
        ASSERT_EQ(&buffers, task.buffers) << "buffer=" << (void *)task.buffers;
        ASSERT_EQ(i, task.index) << "index=" << task.index;
        ASSERT_EQ(totalOffset, task.fileOffset) << "fileOffset=" << task.fileOffset;
        totalOffset += task.readSize;
        i++;
        ASSERT_LT(task.pageIndex, 10UL) << "pageIndex=" << task.pageIndex;
        ASSERT_LT(task.pageOffset, buffers[task.pageIndex].size) << "pageOffset=" << task.pageOffset;
    }
}
}