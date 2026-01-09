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
#include <fstream>
#include <sys/xattr.h>
#include <gtest/gtest.h>
#include <mockcpp/mokc.h>

#include "mem_fs_constants.h"
#include "pacific_adapter.h"
#include "test_file_utils.h"
#include "under_fs_factory.h"

using namespace ock::ufs;

namespace {

class UfsPacificAdapterTest : public testing::Test {
public:
    static void SetUpTestSuite();
    static void TearDownTestSuite();

public:
    void SetUp() override;
    void TearDown() override;

protected:
    void ClearMockPath();
    static ssize_t GetXattrMock(const char *path, const char *name, void *value, size_t size);

protected:
    static utils::ByteBuffer buffer;
    static std::string rootPath;
    static BaseFileService *ufs;
    static utils::ByteBuffer emptyBuffer;
};

utils::ByteBuffer UfsPacificAdapterTest::buffer;
std::string UfsPacificAdapterTest::rootPath;
BaseFileService *UfsPacificAdapterTest::ufs;
utils::ByteBuffer UfsPacificAdapterTest::emptyBuffer;

template <class Iter> void FillWithRandomValues(Iter start, Iter end, int min, int max)
{
    static std::random_device rd;  // you only need to initialize it once
    static std::mt19937 mte(rd()); // this is a relative big object to create

    std::uniform_int_distribution<int> dist(min, max);
    std::generate(start, end, [&]() { return dist(mte); });
}

void UfsPacificAdapterTest::SetUpTestSuite()
{
    std::string name = ".pacific_adapter_unit-" + std::to_string(getpid()) + "-";
    name.append(std::to_string(std::chrono::steady_clock::now().time_since_epoch().count()));
    rootPath = "/tmp/" + name + "/";
}

void UfsPacificAdapterTest::TearDownTestSuite()
{
    std::string command = "rm -rf " + rootPath;
    system(command.c_str());
    rootPath.clear();
}

void UfsPacificAdapterTest::SetUp()
{
    auto ret = mkdir(rootPath.c_str(), 0755);
    if (ret != 0 && errno != EEXIST) {
        EXPECT_EQ(0, ret) << "create root path failed :" << errno << " : " << strerror(errno);
    }
    ufs = new (std::nothrow) ock::ufs::PacificAdapter(rootPath);
    ASSERT_TRUE(ufs != nullptr);
}

void UfsPacificAdapterTest::TearDown()
{
    delete ufs;
    ufs = nullptr;
    ClearMockPath();
}

void UfsPacificAdapterTest::ClearMockPath()
{
    TestFileUtils::RemoveDirectory(rootPath);
    GlobalMockObject::verify();
}

ssize_t UfsPacificAdapterTest::GetXattrMock(const char *path, const char *name, void *value, size_t size)
{
    if (!buffer.Valid()) {
        errno = ENODATA;
        return -1;
    }

    if (buffer.Capacity() > XATTR_SIZE_MAX) {
        errno = E2BIG;
        return -1;
    }

    if (value == nullptr) {
        return static_cast<ssize_t>(buffer.Capacity());
    }

    if (buffer.Capacity() > size) {
        errno = ERANGE;
        return -1;
    }

    buffer.Offset(0);
    buffer.Read(static_cast<uint8_t *>(value), buffer.Capacity());
    return static_cast<ssize_t>(buffer.Capacity());
}

TEST_F(UfsPacificAdapterTest, open_to_write_not_exist)
{
    std::string fileName = "open_to_read_not_exist.txt";
    auto ret = ufs->PutFile(fileName, 0, FileMode(0644), emptyBuffer);
    ASSERT_EQ(-1, ret);
    EXPECT_EQ(ENOENT, errno);
}

TEST_F(UfsPacificAdapterTest, create_excl_file_already_exist)
{
    std::string fileName = "create_excl_file_already_exist.txt";
    auto ret = ufs->PutFile(fileName, FileMode(0644), emptyBuffer);
    ASSERT_EQ(0, ret);

    ret = ufs->PutFile(fileName, O_CREAT | O_EXCL, FileMode(0644), emptyBuffer);
    ASSERT_EQ(-1, ret);
    EXPECT_EQ(EEXIST, errno);
}

TEST_F(UfsPacificAdapterTest, repeat_create_file_no_excl)
{
    std::string fileName = "repeat_create_file_no_excl.txt";
    auto ret = ufs->PutFile(fileName, O_CREAT, FileMode(0644), emptyBuffer);
    ASSERT_EQ(0, ret);

    ret = ufs->PutFile(fileName, O_CREAT, FileMode(0644), emptyBuffer);
    ASSERT_EQ(0, ret);

    ret = ufs->PutFile(fileName, O_CREAT, FileMode(0644), emptyBuffer);
    ASSERT_EQ(0, ret);
}

TEST_F(UfsPacificAdapterTest, write_small_data_read_check)
{
    std::string fileName = "write_small_data_read_check.dat";
    utils::ByteBuffer inputData(4096U);
    FillWithRandomValues(inputData.Data(), inputData.Data() + inputData.Capacity(), 0, 255);

    auto ret = ufs->PutFile(fileName, FileMode(0644), inputData);
    ASSERT_EQ(0, ret);

    utils::ByteBuffer outputData;
    ret = ufs->GetFile(fileName, outputData);
    ASSERT_EQ(0, ret);
    ASSERT_EQ(inputData.Capacity(), outputData.Capacity());
    ASSERT_EQ(outputData.Offset(), outputData.Capacity());
    ASSERT_TRUE(memcmp(inputData.Data(), outputData.Data(), inputData.Capacity()) == 0);
}

TEST_F(UfsPacificAdapterTest, write_small_data_read_range_check)
{
    std::string fileName = "write_small_data_read_range_check.dat";
    utils::ByteBuffer inputData(256U * 1024U);
    FillWithRandomValues(inputData.Data(), inputData.Data() + inputData.Capacity(), 0, 255);

    auto ret = ufs->PutFile(fileName, FileMode(0644), inputData);
    ASSERT_EQ(0, ret);

    utils::ByteBuffer outputData;
    auto r1st = 12345U;
    auto r1len = 4400U;
    ret = ufs->GetFile(fileName, outputData, FileRange{ r1st, r1len });
    ASSERT_EQ(0, ret);
    ASSERT_EQ(r1len, outputData.Capacity());
    ASSERT_EQ(outputData.Offset(), outputData.Capacity());
    ASSERT_TRUE(memcmp(inputData.Data() + r1st, outputData.Data(), r1len) == 0);
}

TEST_F(UfsPacificAdapterTest, healthy_check_simple)
{
    auto success = ufs->HealthyCheck();
    EXPECT_TRUE(success);
}

TEST_F(UfsPacificAdapterTest, healthy_check_for_invalid_ufs)
{
    PacificAdapter invalidUfs("/tmp/aaaaa/bbbbb/ccccc/ddddd");
    auto success = invalidUfs.HealthyCheck();
    EXPECT_FALSE(success);
}

TEST_F(UfsPacificAdapterTest, file_lock_simple)
{
    std::string filename = "file_lock_simple.lock";

    auto ret = ufs->PutFile(filename, FileMode{ 0644 }, emptyBuffer);
    ASSERT_EQ(0, ret) << "create file : " << filename << " failed : " << strerror(errno);

    auto locker = ufs->GetFileLock(filename);
    ASSERT_TRUE(locker != nullptr);

    ret = locker->Lock();
    ASSERT_EQ(0, ret) << "lock file : " << filename << " failed : " << strerror(errno);

    ret = locker->Unlock();
    ASSERT_EQ(0, ret) << "unlock file : " << filename << " failed : " << strerror(errno);
}

TEST_F(UfsPacificAdapterTest, file_lock_reentrant)
{
    std::string filename = "file_lock_reentrant.lock";

    auto ret = ufs->PutFile(filename, FileMode{ 0644 }, emptyBuffer);
    ASSERT_EQ(0, ret) << "create file : " << filename << " failed : " << strerror(errno);

    auto locker = ufs->GetFileLock(filename);
    ASSERT_TRUE(locker != nullptr);

    ret = locker->Lock();
    ASSERT_EQ(0, ret) << "lock file : " << filename << " failed : " << strerror(errno);

    ret = locker->TryLock();
    ASSERT_EQ(0, ret) << "lock file : " << filename << " failed : " << strerror(errno);

    int childError = 0;
    std::thread child([&ret, &childError, locker]() {
        ret = locker->TryLock();
        childError = errno;
    });
    child.join();
    ASSERT_EQ(0, ret) << "lock file : " << filename << " failed : " << strerror(childError);

    ret = locker->Unlock();
    ASSERT_EQ(0, ret) << "unlock file : " << filename << " failed : " << strerror(errno);
}

TEST_F(UfsPacificAdapterTest, list_file_in_dir_simple)
{
    std::string dirName = "list_file_in_dir_simple";

    auto ret = ufs->CreateDirectory(dirName, FileMode{ 0755 });
    ASSERT_EQ(0, ret) << "create dir : " << dirName << " failed : " << strerror(errno);

    std::string filePrefix = "test_aaa_";
    const auto fileCount = 10;
    const auto beginNumber = 1000;
    for (auto i = 0; i < fileCount; i++) {
        auto number = beginNumber + i;
        auto fullName = dirName;
        fullName.append("/").append(filePrefix).append(std::to_string(number));
        ret = ufs->PutFile(fullName, FileMode{ 0600 }, emptyBuffer);
        ASSERT_EQ(0, ret) << "create file : " << fullName << " failed : " << strerror(errno);
    }

    ListFileResult result;
    ret = ufs->ListFiles(dirName, result);
    ASSERT_EQ(0, ret) << "list files in dir : " << dirName << " failed : " << strerror(errno);

    ASSERT_TRUE(result.marker->Finished());
    result.files.sort([](const ListFileItem &item1, const ListFileItem &item2) { return item1.name < item2.name; });
    auto it = result.files.begin();
    for (auto i = 0; i < fileCount; i++, ++it) {
        ASSERT_TRUE(it != result.files.end()) << "not found items : " << i;
        auto number = beginNumber + i;
        auto fileName = filePrefix + std::to_string(number);
        EXPECT_EQ(fileName, it->name);
    }
}

TEST_F(UfsPacificAdapterTest, list_file_in_dir_large_num_files)
{
    std::string dirName = "list_file_in_dir_large_num_files";

    auto ret = ufs->CreateDirectory(dirName, FileMode{ 0755 });
    ASSERT_EQ(0, ret) << "create dir : " << dirName << " failed : " << strerror(errno);

    std::string filePrefix = "test_bbb_";
    const auto fileCount = 8000;
    const auto beginNumber = 100000;
    for (auto i = 0; i < fileCount; i++) {
        auto number = beginNumber + i;
        auto fullName = dirName;
        fullName.append("/").append(filePrefix).append(std::to_string(number));
        ret = ufs->PutFile(fullName, FileMode{ 0600 }, emptyBuffer);
        ASSERT_EQ(0, ret) << "create file : " << fullName << " failed : " << strerror(errno);
    }

    auto resultCount = 0;
    auto pageCount = 0;
    std::list<std::string> files;
    std::shared_ptr<ListFilePageMarker> marker = nullptr;
    ListFileResult result;

    do {
        if (marker == nullptr) {
            ret = ufs->ListFiles(dirName, result);
        } else {
            ret = ufs->ListFiles(dirName, result, marker);
        }
        ASSERT_EQ(0, ret) << "list files in dir : " << dirName << "page : " << pageCount << " failed : " <<
            strerror(errno);

        std::for_each(result.files.begin(), result.files.end(), [&resultCount, &files](const ListFileItem &item) {
            files.push_back(item.name);
            resultCount++;
        });
        pageCount++;
        marker = result.marker;
    } while (marker != nullptr && !marker->Finished());

    ASSERT_EQ(fileCount, resultCount);
    files.sort();
    auto it = files.begin();
    for (auto i = 0; i < fileCount; i++, ++it) {
        ASSERT_TRUE(it != files.end()) << "not found items : " << i;
        auto number = beginNumber + i;
        auto fileName = filePrefix + std::to_string(number);
        EXPECT_EQ(fileName, *it);
    }
}

TEST_F(UfsPacificAdapterTest, move_file_simple)
{
    std::string filename = "move_file_simple.txt";
    auto ret = ufs->PutFile(filename, FileMode{ 0644 }, emptyBuffer);
    ASSERT_EQ(0, ret) << "create file : " << filename << " failed : " << strerror(errno);

    FileMeta oldMeta;
    ret = ufs->GetFileMeta(filename, oldMeta);
    ASSERT_EQ(0, ret) << "get file : " << filename << " metadata failed : " << strerror(errno);

    std::string newName = "move_file_simple_new_name.txt";
    ret = ufs->MoveFile(filename, newName);
    ASSERT_EQ(0, ret) << "move file : " << filename << " failed : " << strerror(errno);

    FileMeta newMeta;
    ret = ufs->GetFileMeta(newName, newMeta);
    ASSERT_EQ(0, ret) << "get file : " << newName << " metadata failed : " << strerror(errno);

    EXPECT_EQ(oldMeta.meta["st_ino"], newMeta.meta["st_ino"]);
    EXPECT_EQ(oldMeta.meta["st_mode"], newMeta.meta["st_mode"]);
    EXPECT_EQ(oldMeta.meta["st_mtime"], newMeta.meta["st_mtime"]);
    EXPECT_EQ("1", newMeta.meta["st_nlink"]);

    ret = ufs->GetFileMeta(filename, oldMeta);
    ASSERT_EQ(-1, ret);
    ASSERT_EQ(ENOENT, errno);
}

TEST_F(UfsPacificAdapterTest, link_copy_file_simple)
{
    std::string filename = "link_copy_file_simple.txt";
    auto ret = ufs->PutFile(filename, FileMode{ 0644 }, emptyBuffer);
    ASSERT_EQ(0, ret) << "create file : " << filename << " failed : " << strerror(errno);

    FileMeta oldMeta;
    ret = ufs->GetFileMeta(filename, oldMeta);
    ASSERT_EQ(0, ret) << "get file : " << filename << " metadata failed : " << strerror(errno);

    std::string newName = "mlink_copy_file_simple_new_name.txt";
    ret = ufs->CopyFile(filename, newName);
    ASSERT_EQ(0, ret) << "move file : " << filename << " failed : " << strerror(errno);
    ASSERT_EQ("1", oldMeta.meta["st_nlink"]);

    FileMeta newMeta1;
    FileMeta newMeta2;
    ret = ufs->GetFileMeta(filename, newMeta1);
    ASSERT_EQ(0, ret) << "get file : " << filename << " metadata failed : " << strerror(errno);

    ret = ufs->GetFileMeta(newName, newMeta2);
    ASSERT_EQ(0, ret) << "get file : " << newName << " metadata failed : " << strerror(errno);

    EXPECT_EQ(oldMeta.meta["st_ino"], newMeta1.meta["st_ino"]);
    EXPECT_EQ(oldMeta.meta["st_ino"], newMeta2.meta["st_ino"]);
    EXPECT_EQ(oldMeta.meta["st_mode"], newMeta1.meta["st_mode"]);
    EXPECT_EQ(oldMeta.meta["st_mode"], newMeta2.meta["st_mode"]);
    EXPECT_EQ(oldMeta.meta["st_mtime"], newMeta1.meta["st_mtime"]);
    EXPECT_EQ(oldMeta.meta["st_mtime"], newMeta2.meta["st_mtime"]);
    EXPECT_EQ("2", newMeta1.meta["st_nlink"]);
    EXPECT_EQ("2", newMeta2.meta["st_nlink"]);
}

TEST_F(UfsPacificAdapterTest, open_to_write_not_exist_mode_only_wrong)
{
    std::string fileName = "/var/var/var/open_to_read_not_exist.txt";
    auto ret = ufs->PutFile(fileName, FileMode(0644));
    ASSERT_EQ(ret, nullptr);
}

TEST_F(UfsPacificAdapterTest, open_to_write_exist_mode_only)
{
    std::string fileName = "write_small_data_read_check.dat";
    auto ret = ufs->PutFile(fileName, FileMode(0644));
    ASSERT_NE(ret, nullptr);
}

TEST_F(UfsPacificAdapterTest, open_to_write_not_exist_stream_wrong)
{
    std::string fileName = "/var/var/var/open_to_read_not_exist.txt";
    FileInputStream inputstream(fileName, -1, 4096);

    auto ret = ufs->PutFile(fileName, FileMode(0644), inputstream);
    ASSERT_EQ(-1, ret);
}

TEST_F(UfsPacificAdapterTest, write_small_data_read_check_stream)
{
    std::string fileName = "write_small_data_read_check.dat";
    std::string fileName1 = "repeat_create_file_no_excl.txt";
    auto fd = open(fileName1.c_str(), O_CREAT | O_RDWR);
    ASSERT_TRUE(fd > 0);
    FileInputStream inputstream(fileName, fd, 4096);

    auto ret = ufs->PutFile(fileName, FileMode(0644), inputstream);
    ASSERT_EQ(0, ret);

    ret = close(fd);
    ASSERT_EQ(0, ret);
    unlink(fileName1.c_str());
}

TEST_F(UfsPacificAdapterTest, write_small_data_read_check_filename_only)
{
    std::string fileName = "write_small_data_read_check.dat";
    auto ret = ufs->GetFile(fileName);
    ASSERT_EQ(nullptr, ret);
}

TEST_F(UfsPacificAdapterTest, write_small_data_read_check_not_exist)
{
    std::string fileName = "open_to_read_not_exist.txt";
    auto ret0 = ufs->GetFile(fileName);
    ASSERT_EQ(nullptr, ret0);

    utils::ByteBuffer outputData;
    auto ret = ufs->GetFile(fileName, outputData);
    ASSERT_EQ(-1, ret);

    auto r1st = 12345U;
    auto r1len = 4400U;
    auto ret1 = ufs->GetFile(fileName, outputData, FileRange{ r1st, r1len });
    ASSERT_EQ(-1, ret1);
}

TEST_F(UfsPacificAdapterTest, write_small_data_read_check_filename_range)
{
    std::string fileName = "write_small_data_read_check.dat";
    auto r1st = 12345U;
    auto r1len = 4400U;
    auto ret = ufs->GetFile(fileName, FileRange{ r1st, r1len });
    ASSERT_EQ(nullptr, ret);

    r1st = 4400U;
    r1len = 12345U;
    auto ret1 = ufs->GetFile(fileName, FileRange{ r1st, r1len });
    ASSERT_EQ(nullptr, ret1);

    r1st = 1U;
    r1len = 2U;
    auto ret2 = ufs->GetFile(fileName, FileRange{ r1st, r1len });
    ASSERT_EQ(nullptr, ret2);
}

TEST_F(UfsPacificAdapterTest, write_small_data_read_check_outstream)
{
    std::string fileName = "write_small_data_read_check.dat";
    auto fd = open(fileName.c_str(), O_CREAT | O_RDWR);
    ASSERT_TRUE(fd > 0);
    utils::ByteBuffer inputData(4096U);
    FillWithRandomValues(inputData.Data(), inputData.Data() + inputData.Capacity(), 0, 255);

    auto ret = ufs->PutFile(fileName, FileMode(0644), inputData);
    ASSERT_EQ(0, ret);
    FileOutputStream outputstream(fileName, fd, 4096U);

    auto r1st = 1U;
    auto r1len = 12U;
    auto ret1 = ufs->GetFile(fileName, outputstream);
    ASSERT_EQ(0, ret1);
    ret1 = ufs->GetFile(fileName, FileRange{ r1st, r1len }, outputstream);
    ASSERT_EQ(-1, ret1);

    ret = ufs->RemoveFile(fileName);
    ASSERT_EQ(0, ret);
    unlink(fileName.c_str());
}

TEST_F(UfsPacificAdapterTest, write_small_data_read_check_outstream_wrong)
{
    std::string fileName = "write_small_data_read_check.dat";
    std::string fileName1 = "repeat_create_file_no_excl.txt";
    auto fd = open(fileName1.c_str(), O_CREAT | O_RDWR);
    ASSERT_TRUE(fd > 0);
    utils::ByteBuffer inputData(4096U);
    FillWithRandomValues(inputData.Data(), inputData.Data() + inputData.Capacity(), 0, 255);

    auto ret = ufs->PutFile(fileName, FileMode(0644), inputData);
    ASSERT_EQ(0, ret);
    FileOutputStream outputstream(fileName, fd, 4096U);

    auto r1st = 12345U;
    auto r1len = 4400U;
    auto ret1 = ufs->GetFile(fileName, FileRange{ r1st, r1len }, outputstream);
    ASSERT_EQ(0, ret1);

    uint32_t dataSize = 1024;
    uint8_t data[dataSize];
    for (int i = 0; i < dataSize; ++i) {
        data[i] = 'a';
    }
    ret = outputstream.Write(data, dataSize);
    ASSERT_EQ(dataSize, ret);

    uint8_t c = 'c';
    ret = outputstream.Write(c);
    ASSERT_EQ(1, ret);
    ret = outputstream.Close();
    ASSERT_EQ(0, ret);

    ret = ufs->RemoveFile(fileName1);
    ASSERT_EQ(-1, ret);
    unlink(fileName1.c_str());
}

TEST_F(UfsPacificAdapterTest, set_file_meta)
{
    std::map<std::string, std::string> meta;
    std::string fileName = "write_small_data_read_check.dat";
    meta["st_uid"] = std::to_string(111);
    meta["st_gid"] = std::to_string(111);
    auto ret = ufs->SetFileMeta(fileName, meta);
    ASSERT_EQ(0, ret);
}

TEST_F(UfsPacificAdapterTest, create_remove_dir)
{
    std::string dirName = "write_small_data_read_check.dat";
    auto ret = ufs->CreateDirectory(dirName, FileMode{ 0777 });
    ASSERT_EQ(0, ret);

    ret = ufs->RemoveDirectory(dirName);
    ASSERT_EQ(0, ret);
}

TEST_F(UfsPacificAdapterTest, ufs_api_totalsize)
{
    std::string fileName = "write_small_data_read_check.dat";

    auto inputstream = new NullInputStream();

    auto ret = inputstream->TotalSize();
    ASSERT_EQ(-1, ret);

    delete inputstream;
}

TEST_F(UfsPacificAdapterTest, ufs_api_read)
{
    std::string fileName = "write_small_data_read_check.dat";

    auto inputstream = new NullInputStream();

    uint32_t dataSize = 1024;
    uint8_t buf[dataSize];
    auto ret = inputstream->InputStream::Read(buf, dataSize);
    ASSERT_EQ(0, ret);

    delete inputstream;
}

TEST_F(UfsPacificAdapterTest, ufs_api_write)
{
    std::string fileName = "write_small_data_read_check.dat";

    auto fd = open(fileName.c_str(), O_CREAT | O_RDWR);
    ASSERT_TRUE(fd > 0);
    FileOutputStream outputstream(fileName, fd, 1024U);

    uint32_t dataSize = 1024;
    uint8_t buf[dataSize];
    for (int i = 0; i < dataSize; ++i) {
        buf[i] = 'a';
    }
    auto ret = outputstream.OutputStream::Write(buf, dataSize);
    ASSERT_EQ(dataSize, ret);

    ret = outputstream.Close();
    ASSERT_EQ(0, ret);
    unlink(fileName.c_str());
}

TEST_F(UfsPacificAdapterTest, ByteBuffer_test_init_delete)
{
    utils::ByteBuffer *inputData = new utils::ByteBuffer();
    ASSERT_NE(nullptr, inputData);

    uint8_t buf[1024];
    utils::ByteBuffer *inputData1 = new utils::ByteBuffer(buf, 1024);
    ASSERT_NE(nullptr, inputData1);

    uint32_t cap = 1024;
    utils::ByteBuffer *inputData2 = new utils::ByteBuffer(cap);
    ASSERT_NE(nullptr, inputData2);

    delete inputData;
    delete inputData1;
    delete inputData2;
}

TEST_F(UfsPacificAdapterTest, ByteBuffer_test)
{
    uint32_t cap = 1024;
    utils::ByteBuffer *inputData2 = new utils::ByteBuffer(cap);
    ASSERT_NE(nullptr, inputData2);
    uint8_t buf[1024];
    for (int i = 0; i < cap; ++i) {
        buf[i] = 'a';
    }

    auto ret = inputData2->Write(buf, 100);
    ASSERT_EQ(0, ret);

    ret = inputData2->WriteAt(buf, 100, 100);
    ASSERT_EQ(0, ret);

    uint8_t data[1024];
    ret = inputData2->Read(data, 100);
    ASSERT_EQ(0, ret);

    ret = inputData2->ReadAt(data, 100, 100);
    ASSERT_EQ(0, ret);

    auto data1 = inputData2->Data();
    for (int j = 0; j < 100; ++j) {
        ASSERT_EQ(data1[j], data[j]);
    }
    auto capacity = inputData2->Capacity();
    ASSERT_EQ(1024, capacity);
    auto offset = inputData2->Offset();
    ASSERT_EQ(200, offset);

    delete inputData2;
}

TEST_F(UfsPacificAdapterTest, FileAcl_empty)
{
    std::string fileName = "test_file_acl_" + std::to_string(__LINE__) + ".tmp";
    auto ret = ufs->PutFile(fileName, FileMode(0644), emptyBuffer);
    ASSERT_EQ(0, ret) << "Put file(" << fileName << ") failed: " << errno << " : " << strerror(errno);

    FileMeta meta;
    ret = ufs->GetFileMeta(fileName, meta);
    ASSERT_EQ(0, ret) << "Get file(" << fileName << ") meta failed: " << errno << " : " << strerror(errno);
    ASSERT_TRUE(meta.acl.users.empty());
    ASSERT_TRUE(meta.acl.groups.empty());
}

TEST_F(UfsPacificAdapterTest, FileAcl_size_less_head)
{
    std::string fileName = "test_file_acl_" + std::to_string(__LINE__) + ".tmp";
    auto ret = ufs->PutFile(fileName, FileMode(0644), emptyBuffer);
    ASSERT_EQ(0, ret) << "Put file(" << fileName << ") failed: " << errno << " : " << strerror(errno);

    FileMeta meta;
    MOCKER(getxattr).stubs().will(returnValue(static_cast<int>(sizeof(FileAclHeader) - 1)));
    ret = ufs->GetFileMeta(fileName, meta);
    ASSERT_EQ(0, ret) << "Get file(" << fileName << ") meta failed: " << errno << " : " << strerror(errno);
    ASSERT_TRUE(meta.acl.users.empty());
    ASSERT_TRUE(meta.acl.groups.empty());
}

TEST_F(UfsPacificAdapterTest, FileAcl_size_less_value)
{
    std::string fileName = "test_file_acl_" + std::to_string(__LINE__) + ".tmp";
    auto ret = ufs->PutFile(fileName, FileMode(0644), emptyBuffer);
    ASSERT_EQ(0, ret) << "Put file(" << fileName << ") failed: " << errno << " : " << strerror(errno);

    FileMeta meta;
    MOCKER(getxattr).stubs().will(returnValue(static_cast<ssize_t>(sizeof(FileAclHeader) + sizeof(FileAclEntry) - 1)));
    ret = ufs->GetFileMeta(fileName, meta);
    ASSERT_EQ(0, ret) << "Get file(" << fileName << ") meta failed: " << errno << " : " << strerror(errno);
    ASSERT_TRUE(meta.acl.users.empty());
    ASSERT_TRUE(meta.acl.groups.empty());
}

TEST_F(UfsPacificAdapterTest, FileAcl_allocate_buffer_failed)
{
    std::string fileName = "test_file_acl_" + std::to_string(__LINE__) + ".tmp";
    auto ret = ufs->PutFile(fileName, FileMode(0644), emptyBuffer);
    ASSERT_EQ(0, ret) << "Put file(" << fileName << ") failed: " << errno << " : " << strerror(errno);

    FileMeta meta;
    MOCKER(getxattr).stubs().will(returnValue(static_cast<ssize_t>(sizeof(FileAclHeader) + 2 * sizeof(FileAclEntry))));
    union MockerHelper {
        bool (utils::ByteBuffer::*isValid)() const noexcept;
        bool (*mockIsValid)(const utils::ByteBuffer *self) noexcept;
    };

    MockerHelper helper{};
    helper.isValid = &utils::ByteBuffer::Valid;
    MOCKCPP_NS::mockAPI("&utils::ByteBuffer::Valid", helper.mockIsValid).stubs().will(returnValue(false));
    ret = ufs->GetFileMeta(fileName, meta);
    ASSERT_NE(0, ret);
    ASSERT_EQ(ENOMEM, errno);
}

TEST_F(UfsPacificAdapterTest, FileAcl_next_size_changed)
{
    std::string fileName = "test_file_acl_" + std::to_string(__LINE__) + ".tmp";
    auto ret = ufs->PutFile(fileName, FileMode(0644), emptyBuffer);
    ASSERT_EQ(0, ret) << "Put file(" << fileName << ") failed: " << errno << " : " << strerror(errno);

    FileMeta meta;
    MOCKER(getxattr).stubs().will(returnObjectList(static_cast<ssize_t>(sizeof(FileAclHeader) + sizeof(FileAclEntry)),
        static_cast<ssize_t>(sizeof(FileAclHeader) + 2 * sizeof(FileAclEntry))));
    ret = ufs->GetFileMeta(fileName, meta);
    ASSERT_EQ(0, ret) << "Get file(" << fileName << ") meta failed: " << errno << " : " << strerror(errno);
    ASSERT_TRUE(meta.acl.users.empty());
    ASSERT_TRUE(meta.acl.groups.empty());
}

TEST_F(UfsPacificAdapterTest, FileAcl_with_one_user)
{
    std::string fileName = "test_file_acl_" + std::to_string(__LINE__) + ".tmp";
    auto ret = ufs->PutFile(fileName, FileMode(0644), emptyBuffer);
    ASSERT_EQ(0, ret) << "Put file(" << fileName << ") failed: " << errno << " : " << strerror(errno);

    FileMeta meta;
    buffer = utils::ByteBuffer{sizeof(FileAclHeader) + sizeof(FileAclEntry)};
    auto acl = static_cast<FileAclHeader *>(static_cast<void *>(buffer.Data()));
    acl->version = 2;
    acl->entries[0].tag = ock::memfs::MemFsConstants::ACL_TAG_USER;
    acl->entries[0].perm = 7;
    acl->entries[0].id = 8008;

    MOCKER(getxattr).stubs().will(invoke(GetXattrMock));
    ret = ufs->GetFileMeta(fileName, meta);
    ASSERT_EQ(0, ret);
    ASSERT_TRUE(meta.acl.groups.empty());
    ASSERT_EQ(1UL, meta.acl.users.size());
    auto pos = meta.acl.users.begin();

    ASSERT_EQ(acl->entries[0].id, pos->first);
    ASSERT_EQ(acl->entries[0].perm, pos->second);
}

TEST_F(UfsPacificAdapterTest, FileAcl_with_one_group)
{
    std::string fileName = "test_file_acl_" + std::to_string(__LINE__) + ".tmp";
    auto ret = ufs->PutFile(fileName, FileMode(0644), emptyBuffer);
    ASSERT_EQ(0, ret) << "Put file(" << fileName << ") failed: " << errno << " : " << strerror(errno);

    FileMeta meta;
    buffer = utils::ByteBuffer{sizeof(FileAclHeader) + sizeof(FileAclEntry)};
    auto acl = static_cast<FileAclHeader *>(static_cast<void *>(buffer.Data()));
    acl->version = 2;
    acl->entries[0].tag = ock::memfs::MemFsConstants::ACL_TAG_GROUP;
    acl->entries[0].perm = 7;
    acl->entries[0].id = 8008;

    MOCKER(getxattr).stubs().will(invoke(GetXattrMock));
    ret = ufs->GetFileMeta(fileName, meta);
    ASSERT_EQ(0, ret);
    ASSERT_TRUE(meta.acl.users.empty());
    ASSERT_EQ(1UL, meta.acl.groups.size());
    auto pos = meta.acl.groups.begin();

    ASSERT_EQ(acl->entries[0].id, pos->first);
    ASSERT_EQ(acl->entries[0].perm, pos->second);
}

TEST_F(UfsPacificAdapterTest, put_file_with_long_file_name)
{
    std::string fileName(4096, 'a');
    auto ret = ufs->PutFile(fileName, FileMode(0644), emptyBuffer);
    ASSERT_EQ(-1, ret);

    FileInputStream stream{ fileName, 0, 0 };
    ret = ufs->PutFile(fileName, FileMode(0644), stream);
    ASSERT_EQ(-1, ret);

    FileRange range{};
    auto outputStream = ufs->PutFile(fileName, FileMode(0644), range);
    ASSERT_EQ(nullptr, outputStream);
}

TEST_F(UfsPacificAdapterTest, get_file_with_long_file_name)
{
    std::string fileName(4096, 'a');
    FileRange range{};
    auto ret = ufs->GetFile(fileName, emptyBuffer, range);
    ASSERT_EQ(-1, ret);

    auto inputStream = ufs->GetFile(fileName, range);
    ASSERT_EQ(nullptr, inputStream);

    FileOutputStream stream{ fileName, 0, 0 };
    ret = ufs->GetFile(fileName, range, stream);
    ASSERT_EQ(-1, ret);
}

TEST_F(UfsPacificAdapterTest, move_file_abnormal)
{
    std::string src(4096, 'a');
    std::string dst(4096, 'a');

    // src too long
    auto ret = ufs->MoveFile(src, dst);
    ASSERT_EQ(-1, ret);

    // dst too long
    ret = ufs->MoveFile("src", dst);
    ASSERT_EQ(-1, ret);

    // src no exist
    ret = ufs->MoveFile("src", "dst");
    ASSERT_EQ(-1, ret);
}

TEST_F(UfsPacificAdapterTest, copy_file_abnormal)
{
    std::string src(4096, 'a');
    std::string dst(4096, 'a');

    // src too long
    auto ret = ufs->CopyFile(src, dst);
    ASSERT_EQ(-1, ret);

    // dst too long
    ret = ufs->CopyFile("src", dst);
    ASSERT_EQ(-1, ret);

    // src no exist
    ret = ufs->CopyFile("src", "dst");
    ASSERT_EQ(-1, ret);
}

TEST_F(UfsPacificAdapterTest, set_file_meta_with_long_file_name)
{
    std::string fileName(4096, 'a');
    std::map<std::string, std::string> meta;
    auto ret = ufs->SetFileMeta(fileName, meta);
    ASSERT_EQ(-1, ret);
}

TEST_F(UfsPacificAdapterTest, set_file_meta_no_exist)
{
    std::string fileName = "/set_file_meta_no_st_mode";
    std::map<std::string, std::string> meta = { {"st_mode", "600"} };
    auto ret = ufs->SetFileMeta(fileName, meta);
    ASSERT_EQ(-1, ret);
}

TEST_F(UfsPacificAdapterTest, set_file_meta_normal)
{
    std::string fileName = "/set_file_meta_normal";
    std::ofstream{ rootPath + fileName };
    std::map<std::string, std::string> meta = { {"st_mode", "600"} };
    auto ret = ufs->SetFileMeta(fileName, meta);
    ASSERT_EQ(0, ret);
}

TEST_F(UfsPacificAdapterTest, lock_file_no_exist)
{
    std::string fileName = "/try_lock_file_no_exist";
    PacificFileLock lock{ fileName, -1 };
    auto ret = lock.Lock();
    ASSERT_EQ(-1, ret);

    ret = lock.TryLock();
    ASSERT_EQ(-1, ret);

    ret = lock.Unlock();
    ASSERT_EQ(-1, ret);
}
}