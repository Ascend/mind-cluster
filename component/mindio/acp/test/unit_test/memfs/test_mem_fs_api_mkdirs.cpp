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
#include <mockcpp/mockcpp.hpp>

#include <map>
#include <atomic>

#include "mem_file_system.h"
#include "memfs_api.h"

using namespace ock::memfs;

struct StatMock {
    uint64_t inode;
    mode_t mode;

    StatMock(uint64_t i, mode_t m) noexcept : inode { i }, mode { m } {}
};

class TestMemFsApiMkdirs : public testing::Test {
public:
    void SetUp() override;
    void TearDown() override;

protected:
    static int MakeDirectoryMock(MemFileSystem *self, const std::string &path, mode_t mode,
        uint64_t &outInode) noexcept;
    static int GetMetaMock(MemFileSystem *self, const std::string &path, struct stat &metadata) noexcept;
    static int AddFileMock(const std::string &path, mode_t mode) noexcept;
    static void EnableMocks() noexcept;

protected:
    static int mockingMkdirErrorNum;
    static int mockingGetStatErrorNum;
    static std::atomic<uint64_t> inodeGen;
    static std::map<std::string, StatMock> filesMock;
};

int TestMemFsApiMkdirs::mockingMkdirErrorNum = 0;
int TestMemFsApiMkdirs::mockingGetStatErrorNum = 0;
std::atomic<uint64_t> TestMemFsApiMkdirs::inodeGen { 0x2000UL };
std::map<std::string, StatMock> TestMemFsApiMkdirs::filesMock;

void TestMemFsApiMkdirs::SetUp()
{
    mockingMkdirErrorNum = 0;
    mockingGetStatErrorNum = 0;
    filesMock.clear();
    filesMock.emplace("/", StatMock { 0x1000UL, __S_IFDIR | 0755 });
    MemFsApi::Initialize();
}

void TestMemFsApiMkdirs::TearDown()
{
    GlobalMockObject::verify();
    filesMock.clear();
    mockingMkdirErrorNum = 0;
    mockingGetStatErrorNum = 0;
    MemFsApi::Destroy();
}

int TestMemFsApiMkdirs::MakeDirectoryMock(MemFileSystem *self, const std::string &path, mode_t mode,
    uint64_t &outInode) noexcept
{
    if (mockingMkdirErrorNum != 0) {
        errno = mockingMkdirErrorNum;
        return -1;
    }

    auto pos = filesMock.find(path);
    if (pos != filesMock.end()) {
        errno = EEXIST;
        return -1;
    }

    outInode = inodeGen.fetch_add(1UL);
    filesMock.emplace(path, StatMock { outInode, __S_IFDIR | (mode & 0777) });
    return 0;
}

int TestMemFsApiMkdirs::GetMetaMock(MemFileSystem *self, const std::string &path, struct stat &metadata) noexcept
{
    if (mockingGetStatErrorNum != 0) {
        errno = mockingGetStatErrorNum;
        return -1;
    }

    auto pos = filesMock.find(path);
    if (pos == filesMock.end()) {
        errno = ENOENT;
        return -1;
    }

    metadata.st_mode = pos->second.mode;
    metadata.st_ino = pos->second.inode;
    return 0;
}

int TestMemFsApiMkdirs::AddFileMock(const std::string &path, mode_t mode) noexcept
{
    auto pos = filesMock.find(path);
    if (pos != filesMock.end()) {
        errno = EEXIST;
        return -1;
    }

    filesMock.emplace(path, StatMock { inodeGen.fetch_add(1UL), __S_IFREG | (mode & 0777) });
    return 0;
}

union MockerHelper {
    int (MemFileSystem::*realMkdir)(const std::string &path, mode_t mode, uint64_t &outInode) noexcept;
    int (*mockMkdir)(MemFileSystem *self, const std::string &path, mode_t mode, uint64_t &outInode) noexcept;
    int (MemFileSystem::*realGetMeta)(const std::string &path, struct stat &metadata) noexcept;
    int (*mockGetMeta)(MemFileSystem *self, const std::string &path, struct stat &metadata) noexcept;
};

void TestMemFsApiMkdirs::EnableMocks() noexcept
{
    MockerHelper helper {};
    helper.realMkdir = &MemFileSystem::MakeDirectory;
    MOCKCPP_NS::mockAPI("&MemFileSystem::MakeDirectory", helper.mockMkdir).stubs().will(invoke(MakeDirectoryMock));

    helper.realGetMeta = &MemFileSystem::GetMeta;
    MOCKCPP_NS::mockAPI("&MemFileSystem::GetMeta", helper.mockGetMeta).stubs().will(invoke(GetMetaMock));
}

namespace {

TEST_F(TestMemFsApiMkdirs, create_deep_path_simple)
{
    std::vector<std::string> items { "aaa", "bbb", "ccc", "ddd", "eee" };
    std::string fullPath;
    std::for_each(items.begin(), items.end(),
        [&fullPath](const std::string &item) { fullPath.append("/").append(item); });

    EnableMocks();
    auto ret = MemFsApi::CreateDirectoryWithParents(fullPath, 0755);
    EXPECT_EQ(0, ret) << "create failed : " << errno << " : " << strerror(errno);

    fullPath.clear();
    struct stat statBuf {};
    for (auto &item : items) {
        fullPath.append("/").append(item);
        ret = MemFsApi::GetMeta(fullPath, statBuf);
        EXPECT_EQ(0, ret) << "check path (" << fullPath << ") failed : " << errno << " : " << strerror(errno);
    }
}

TEST_F(TestMemFsApiMkdirs, create_deep_path_exist_some)
{
    std::vector<std::string> items { "aaa", "bbb", "ccc", "ddd", "eee" };
    std::string fullPath;
    std::for_each(items.begin(), items.end(),
        [&fullPath](const std::string &item) { fullPath.append("/").append(item); });

    EnableMocks();
    auto ret = MemFsApi::CreateDirectoryWithParents(fullPath, 0755);
    EXPECT_EQ(0, ret) << "create failed : " << errno << " : " << strerror(errno);

    fullPath.append("/more-path");
    ret = MemFsApi::CreateDirectoryWithParents(fullPath, 0755);
    EXPECT_EQ(0, ret) << "create failed : " << errno << " : " << strerror(errno);
}

TEST_F(TestMemFsApiMkdirs, create_deep_path_exist_file)
{
    std::vector<std::string> items { "aaa", "bbb", "ccc", "ddd", "eee" };
    std::string fullPath;
    std::for_each(items.begin(), items.end(),
        [&fullPath](const std::string &item) { fullPath.append("/").append(item); });

    EnableMocks();
    auto ret = MemFsApi::CreateDirectoryWithParents(fullPath, 0755);
    EXPECT_EQ(0, ret) << "create failed : " << errno << " : " << strerror(errno);

    fullPath.append("/file");
    ret = AddFileMock(fullPath, 0644);
    EXPECT_EQ(0, ret);

    fullPath.append("/should_failed");
    ret = MemFsApi::CreateDirectoryWithParents(fullPath, 0755);
    ASSERT_NE(0, ret);
    EXPECT_EQ(ENOTDIR, errno) << "failed type: " << errno << " : " << strerror(errno);
}

TEST_F(TestMemFsApiMkdirs, create_deep_path_failed_specified)
{
    std::vector<std::string> items { "aaa", "bbb", "ccc", "ddd", "eee" };
    std::string fullPath;
    std::for_each(items.begin(), items.end(),
        [&fullPath](const std::string &item) { fullPath.append("/").append(item); });

    EnableMocks();
    mockingMkdirErrorNum = EAGAIN;
    auto ret = MemFsApi::CreateDirectoryWithParents(fullPath, 0755);
    ASSERT_NE(0, ret);
    EXPECT_EQ(EAGAIN, errno) << "failed type: " << errno << " : " << strerror(errno);
}

TEST_F(TestMemFsApiMkdirs, create_deep_path_stat_failed_specified)
{
    std::vector<std::string> items { "aaa", "bbb", "ccc", "ddd", "eee" };
    std::string fullPath;
    std::for_each(items.begin(), items.end(),
        [&fullPath](const std::string &item) { fullPath.append("/").append(item); });

    EnableMocks();
    mockingGetStatErrorNum = EPERM;
    auto ret = MemFsApi::CreateDirectoryWithParents(fullPath, 0755);
    ASSERT_NE(0, ret);
    EXPECT_EQ(EPERM, errno) << "failed type: " << errno << " : " << strerror(errno);
}

TEST_F(TestMemFsApiMkdirs, create_deep_path_others_create)
{
    std::vector<std::string> items { "aaa", "bbb", "ccc", "ddd", "eee" };
    std::string fullPath;
    std::for_each(items.begin(), items.end(),
        [&fullPath](const std::string &item) { fullPath.append("/").append(item); });

    EnableMocks();
    mockingGetStatErrorNum = ENOENT;
    mockingMkdirErrorNum = EEXIST;
    auto ret = MemFsApi::CreateDirectoryWithParents(fullPath, 0755);
    EXPECT_EQ(0, ret) << "create failed : " << errno << " : " << strerror(errno);
}
}