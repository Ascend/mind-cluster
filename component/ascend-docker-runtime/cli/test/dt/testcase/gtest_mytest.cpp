/*
 * Copyright(C) 2022. Huawei Technologies Co.,Ltd. All rights reserved.
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
#include <string>
#include <iostream>
#include <climits>
#include <sys/mount.h>
#include <unistd.h>

#include "gtest/gtest.h"
#include "securec.h"
#include "mockcpp/mockcpp.hpp"

#include "../../../src/basic.h"

using namespace std;
using namespace testing;

typedef char *(*ParseFileLine)(char *, const char *);
extern "C" int IsStrEqual(const char *s1, const char *s2);
extern "C" int GetNsPath(const int pid, const char *nsType, char *buf, size_t bufSize);
extern "C" int snprintf_s(char *strDest, size_t destMax, size_t count, const char *format, ...);
extern "C" int open(const char *path, int flags);
extern "C" int close(int fd);
extern "C" int stat(const char *file_name, struct stat *buf);
extern "C" int mount(const char *source, const char *target,
                     const char *filesystemtype, unsigned long mountflags, const void *data);
extern "C" int Mount(const char *src, const char *dst);
extern "C" int rmdir(const char *pathname);
extern "C" int EnterNsByFd(int fd, int nsType);
extern "C" bool StrHasPrefix(const char *str, const char *prefix);
extern "C" int GetNsPath(const int pid, const char *nsType, char *buf, size_t bufSize);
extern "C" int GetSelfNsPath(const char *nsType, char *buf, size_t bufSize);
extern "C" int EnterNsByPath(const char *path, int nsType);
extern "C" int CheckDirExists(char *dir, int len);
extern "C" int GetParentPathStr(const char *path, char *parent, size_t bufSize);
extern "C" int MakeDirWithParent(const char *path, mode_t mode);
extern "C" int MountDir(const char *rootfs, const char *file, unsigned long reMountRwFlag);
extern "C" int SetupContainer(struct CmdArgs *args);
extern "C" int Process(int argc, char **argv);
extern "C" int DoFileMounting(const char *rootfs, const struct MountList *list);
extern "C" int DoMounting(const struct ParsedConfig *config);
extern "C" int DoDirectoryMounting(const char *rootfs, const struct MountList *list);
extern "C" int DoPrepare(const struct CmdArgs *args, struct ParsedConfig *config);
extern "C" int ParseRuntimeOptions(const char *options);
extern "C" bool IsOptionNoDrvSet();
extern "C" bool IsVirtual();
extern "C" int MakeMountPoints(const char *path, mode_t mode);
extern "C" int LogLoop(const char *filename);
extern "C" bool TakeNthWord(char **pLine, unsigned int n, char **word);
extern "C" bool CheckRootDir(char **pLine);
extern "C" bool GetFileSubsetAndCheck(const char *basePath, const size_t basePathLen);
extern "C" bool CheckExistsFile(const char* filePath, const size_t filePathLen,
                                const size_t maxFileSizeMb, const bool checkWgroup);
extern "C" int VerifyPathInfo(const struct PathInfo* pathInfo);
extern "C" bool CheckExternalFile(const char* filePath, const size_t filePathLen,
                                  const size_t maxFileSizeMb, const bool checkOwner);
extern "C" int LogLoop(const char* filename);
extern "C" void Logger(const char *msg, int level, int screen);
extern "C" int EnterNsByPath(const char *path, int nsType);
extern "C" bool CheckOpenedFile(FILE* fp, const long maxSize, const bool checkOwner);

#ifdef GOOGLE_TEST
extern "C" STATIC int MkDir(const char *dir, mode_t mode);
extern "C" STATIC void WriteLogFile(const char* filename, const long maxSize, const char* buffer, unsigned bufferSize);
extern "C" STATIC long GetLogSize(const char* filename);
extern "C" STATIC int CreateLog(const char* filename);
extern "C" STATIC int GetCurrentLocalTime(char* buffer, int length);
extern "C" STATIC bool LogConvertStorage(const char* filename, const long maxSize);
extern "C" STATIC void DivertAndWrite(const char *logPath, const char *msg, const int level);
extern "C" STATIC long GetLogSizeProcess(const char* path);
extern "C" STATIC bool CheckSrcFile(const char *src);
extern "C" STATIC int MountFile(const char *rootfs, const char *filepath);
extern "C" STATIC bool OptionsCmdArgParser(struct CmdArgs *args, const char *arg);
extern "C" STATIC bool MountFileCmdArgParser(struct CmdArgs *args, const char *arg);
extern "C" STATIC bool MountDirCmdArgParser(struct CmdArgs *args, const char *arg);
extern "C" STATIC bool LinkCheckCmdArgParser(const char *argv);
#endif

int stub_setns(int fd, int nstype)
{
    return 0;
}

int Stub_GetNsPath_Failed(const int pid, const char *nsType, char *buf, size_t bufSize)
{
    return -1;
}

int Stub_GetSelfNsPath_Failed(const char *nsType, char *buf, size_t bufSize)
{
    return -1;
}

int Stub_EnterNsByFd_Success(int fd, int nsType)
{
    return 0;
}

int Stub_EnterNsByFd_Failed(int fd, int nsType)
{
    return -1;
}

int stub_open_success(const char *path, int flags)
{
    return 0;
}

int stub_open_failed(const char *path, int flags)
{
    return -1;
}

int stub_close_success(int fd)
{
    return 0;
}

int stub_MkDir_success(const char *dir, mode_t mode)
{
    return 0;
}

int stub_MkDir_failed(const char *dir, mode_t mode)
{
    return -1;
}

int stub_mount_success(const char *source, const char *target,
                       const char *filesystemtype, unsigned long mountflags, const void *data)
{
    return 0;
}

int stub_Mount_success(const char *src, const char *dst)
{
    return 0;
}

int stub_Mount_failed(const char *src, const char *dst)
{
    return -1;
}

int stub_mount_failed(const char *source, const char *target,
                      const char *filesystemtype, unsigned long mountflags, const void *data)
{
    return -1;
}

int stub_mount_src_nil_failed(const char *source, const char *target,
                              const char *filesystemtype, unsigned long mountflags, const void *data)
{
    if (source == nullptr) {
        return -1;
    }
    return 0;
}

int stub_stat_success(const char *file_name, struct stat *buf)
{
    return 0;
}

int stub_stat_failed(const char *file_name, struct stat *buf)
{
    return -1;
}

int Stub_MountDevice_Success(const char *rootfs, const char *deviceName)
{
    return 0;
}

int Stub_MountDevice_Failed(const char *rootfs, const char *deviceName)
{
    return -1;
}

int Stub_MountDir_Success(const char *rootfs, const char *file, unsigned long reMountRwFlag)
{
    return 0;
}

int Stub_MountDir_Failed(const char *rootfs, const char *file, unsigned long reMountRwFlag)
{
    return -1;
}

int Stub_CheckDirExists_Success(char *dir, int len)
{
    return 0;
}

int Stub_MakeDirWithParent_Success(const char *path, mode_t mode)
{
    return 0;
}

int Stub_MakeDirWithParent_Failed(const char *path, mode_t mode)
{
    return -1;
}

int Stub_CheckDirExists_Failed(char *dir, int len)
{
    return -1;
}

int Stub_EnterNsByPath_Success(const char *path, int nsType)
{
    return 0;
}

int Stub_EnterNsByPath_Failed(const char *path, int nsType)
{
    return -1;
}

int Stub_DoDeviceMounting_Success(const char *rootfs, const char *device_name, const size_t ids[], size_t idsNr)
{
    return 0;
}

int Stub_DoDeviceMounting_Failed(const char *rootfs, const char *device_name, const size_t ids[], size_t idsNr)
{
    return -1;
}

int Stub_DoCtrlDeviceMounting_Success(const char *rootfs)
{
    return 0;
}

int Stub_DoCtrlDeviceMounting_Failed(const char *rootfs)
{
    return -1;
}

int Stub_DoDirectoryMounting_Success(const char *rootfs, const struct MountList *list)
{
    return 0;
}

int Stub_DoDirectoryMounting_Failed(const char *rootfs, const struct MountList *list)
{
    return -1;
}

int Stub_DoFileMounting_Success(const char *rootfs, const struct MountList *list)
{
    return 0;
}

int Stub_DoFileMounting_Failed(const char *rootfs, const struct MountList *list)
{
    return -1;
}

int Stub_DoMounting_Success(const struct ParsedConfig *config)
{
    return 0;
}

int Stub_DoMounting_Failed(const struct ParsedConfig *config)
{
    return -1;
}

int Stub_SetupCgroup_Success(const struct ParsedConfig *config)
{
    return 0;
}

int Stub_SetupCgroup_Failed(const struct ParsedConfig *config)
{
    return 0;
}

int Stub_SetupContainer_Success(struct CmdArgs *args)
{
    return 0;
}

int Stub_SetupContainer_Failed(struct CmdArgs *args)
{
    return -1;
}

int Stub_SetupDeviceCgroup_Success(FILE *cgroupAllow, const char *devPath)
{
    return 0;
}

int Stub_SetupDeviceCgroup_Failed(FILE *cgroupAllow, const char *devPath)
{
    return -1;
}

int Stub_SetupDriverCgroup_Fail(FILE *cgroupAllow)
{
    return -1;
}

int Stub_SetupDriverCgroup_Success(FILE *cgroupAllow)
{
    return 0;
}

int Stub_DoPrepare_Failed(const struct CmdArgs *args, struct ParsedConfig *config)
{
    return -1;
}

int Stub_DoPrepare_Success(const struct CmdArgs *args, struct ParsedConfig *config)
{
    return 0;
}

int Stub_ParseFileByLine_Success(char *buffer, int bufferSize, ParseFileLine fn, const char *filepath)
{
    return 0;
}

int Stub_GetCgroupPath_Success(int pid, char *effPath, const size_t maxSize)
{
    return 0;
}

bool Stub_IsOptionNoDrvSet_True()
{
    return true;
}

bool Stub_IsOptionNoDrvSet_False()
{
    return false;
}

bool Stub_CheckExistsFile_Success()
{
    return true;
}

int Stub_LogLoop_Failed(const char* filename)
{
    return -1;
}

bool Stub_CheckExternalFile_Failed(const char* filePath, const size_t filePathLen,
    const size_t maxFileSizeMb, const bool checkOwner)
{
    return false;
}

int Stub_MakeMountPoints_Failed(const char *path, mode_t mode)
{
    return -1;
}

int Stub_MakeMountPoints_Success(const char *path, mode_t mode)
{
    return 0;
}

class Test_Fhho : public Test
{
protected:
    static void SetUpTestCase()
    {
        cout << "TestSuite测试套事件：在第一个testcase之前执行" << endl;
    }
    static void TearDownTestCase()
    {
        cout << "TestSuite测试套事件：在最后一个testcase之后执行" << endl;
    }
    // 如果想在相同的测试套中设置两种事件，那么可以写在一起，运行就看到效果了
    virtual void SetUp()
    {
        cout << "TestSuite测试用例事件：在每个testcase之前执行" << endl;
    }
    virtual void TearDown()
    {
        cout << "TestSuite测试用例事件：在每个testcase之后执行" << endl;
    }
};

TEST_F(Test_Fhho, ClassEQ)
{
    int pid = 1;
    const char *nsType = "mnt";
    char buf[100] = {0x0};
    int bufSize = 100;
    int ret = GetNsPath(pid, nsType, buf, 100);
    EXPECT_LE(0, ret);
}

TEST_F(Test_Fhho, GetNsPathCaseOne)
{
    int pid = 1;
    const char *nsType = NULL;
    char buf[100] = {0x0};
    int bufSize = 100;
    int ret = GetNsPath(pid, nsType, buf, bufSize);
    EXPECT_LE(ret, 0);
}

TEST_F(Test_Fhho, StatusOne)
{
    int pid = 1;
    int nsType = 1;
    MOCKER(setns)
        .stubs()
        .will(invoke(stub_setns));
    int ret = EnterNsByFd(pid, nsType);
    GlobalMockObject::verify();
    EXPECT_LE(0, ret);
}

TEST_F(Test_Fhho, StatusTwo)
{
    // The test does not have a file handle into the namespace
    int pid = 1;
    int nsType = 1;
    int ret = EnterNsByFd(pid, nsType);
    EXPECT_LE(-1, ret);
}

TEST_F(Test_Fhho, StatusOne1)
{
    char containerNsPath[BUF_SIZE] = {0};
    int nsType = 1;
    MOCKER(open).stubs().will(invoke(stub_open_success));
    int ret = EnterNsByPath(containerNsPath, nsType);
    GlobalMockObject::verify();
    EXPECT_LE(-1, ret);
}

TEST_F(Test_Fhho, StatusTwo1)
{
    // The test has no path into the namespace
    char containerNsPath[BUF_SIZE] = {0};
    int nsType = 1;
    int ret = EnterNsByPath(containerNsPath, nsType);
    EXPECT_LE(-1, ret);
}

TEST_F(Test_Fhho, EnterNsByPathCase3)
{
    const char* containerNsPath = NULL;
    int nsType = 1;
    int ret = EnterNsByPath(containerNsPath, nsType);
    EXPECT_LE(-1, ret);
}

TEST_F(Test_Fhho, EnterNsByPathCase4)
{
    char containerNsPath[BUF_SIZE] = {0};
    int nsType = 1;
    MOCKER(open).stubs().will(invoke(stub_open_failed));
    int ret = EnterNsByPath(containerNsPath, nsType);
    GlobalMockObject::verify();
    EXPECT_LE(-1, ret);
}

TEST_F(Test_Fhho, EnterNsByPathCase5)
{
    char containerNsPath[BUF_SIZE] = {0};
    int nsType = 1;
    MOCKER(open).stubs().will(invoke(stub_open_success));
    MOCKER(EnterNsByFd).stubs().will(invoke(Stub_EnterNsByFd_Failed));
    int ret = EnterNsByPath(containerNsPath, nsType);
    GlobalMockObject::verify();
    EXPECT_LE(-1, ret);
}

TEST_F(Test_Fhho, EnterNsByPathCase6)
{
    char containerNsPath[BUF_SIZE] = {0};
    int nsType = 1;
    MOCKER(open).stubs().will(invoke(stub_open_success));
    MOCKER(EnterNsByFd).stubs().will(invoke(Stub_EnterNsByFd_Success));
    int ret = EnterNsByPath(containerNsPath, nsType);
    GlobalMockObject::verify();
    EXPECT_LE(0, ret);
}

TEST_F(Test_Fhho, GetNsPathAndGetSelfNsPath)
{
    char containerNsPath[BUF_SIZE] = {0};
    int containerPid = 1;
    EXPECT_LE(0, GetNsPath(containerPid, "mnt", containerNsPath, BUF_SIZE));
    char nsPath[BUF_SIZE] = {0};
    EXPECT_LE(0, GetSelfNsPath("mnt", nsPath, BUF_SIZE));
}

TEST_F(Test_Fhho, StatusOneDoDirectoryMounting)
{
    MOCKER(MountDir).stubs().will(invoke(Stub_MountDir_Failed));
    struct MountList list = {0};
    list.count = 1;
    char *rootfs = "/home";
    int ret = DoDirectoryMounting(rootfs, &list);
    GlobalMockObject::verify();
    EXPECT_EQ(-1, ret);
}

TEST_F(Test_Fhho, StatusTwoDoDirectoryMounting)
{
    MOCKER(MountDir).stubs().will(invoke(Stub_MountDir_Success));
    struct MountList list = {0};
    list.count = 3;
    char *rootfs = "/home";
    int ret = DoDirectoryMounting(rootfs, &list);
    GlobalMockObject::verify();
    EXPECT_EQ(0, ret);
}

TEST_F(Test_Fhho, StatusThreeDoDirectoryMounting)
{
    int ret = DoDirectoryMounting(nullptr, nullptr);
    EXPECT_EQ(-1, ret);
}

TEST_F(Test_Fhho, StatusOneCheckDirExists)
{
    // Test directory exists
    char *dir = "/home";
    int len = strlen(dir);
    int ret = CheckDirExists(dir, len);
    EXPECT_EQ(0, ret);
}

TEST_F(Test_Fhho, StatusTwoCheckDirExists)
{
    // Test directory does not exist
    char *dir = "/home/notexist";
    int len = strlen(dir);
    int ret = CheckDirExists(dir, len);
    EXPECT_EQ(-1, ret);
}

TEST_F(Test_Fhho, StatusOneCheckDirExists1)
{
    // Test get path parent directory
    char *path = "/usr/bin";
    char parent[BUF_SIZE] = {0};
    int ret = GetParentPathStr(path, parent, BUF_SIZE);
    EXPECT_EQ(0, ret);
}

TEST_F(Test_Fhho, StatusOneCheckDirExists11)
{
    // Test get path parent directory
    char *path = nullptr;
    char parent[BUF_SIZE] = {0};
    int ret = GetParentPathStr(path, parent, BUF_SIZE);
    EXPECT_EQ(-1, ret);
}

TEST_F(Test_Fhho, StatusOneMakeDirWithParent)
{
    // The test create directory contains the parent directory
    mode_t mode = 0755;
    char parentDir[BUF_SIZE] = {0};
    int ret = MakeDirWithParent(parentDir, mode);
    EXPECT_EQ(0, ret);
}

TEST_F(Test_Fhho, MakeMountPoints1)
{
    // The test create directory contains the parent directory
    mode_t mode = 0755;
    char *path = "/home";
    int ret = MakeMountPoints(path, mode);
    EXPECT_EQ(-1, ret);
}

TEST_F(Test_Fhho, LogLoopSuccess)
{
    // The test create directory contains the parent directory
    char *filename = "/home/var/log/sys.log";
    int ret = LogLoop(filename);
    EXPECT_EQ(-1, ret);
}

TEST_F(Test_Fhho, StatusTwoMakeDirWithParent)
{
    mode_t mode = 0755;
    char parentDir[BUF_SIZE] = {0};
    MOCKER(CheckDirExists).stubs().will(invoke(Stub_CheckDirExists_Success));
    int ret = MakeDirWithParent(parentDir, mode);
    GlobalMockObject::verify();
    EXPECT_EQ(0, ret);
}

#ifdef GOOGLE_TEST
STATIC long Stub_GetLogSize_Failed(const char* filename)
{
    return -1;
}

STATIC long Stub_GetLogSize_Success(const char* filename)
{
    return strlen(filename);
}

STATIC bool Stub_CheckSrcFile_Success(const char *src)
{
    return true;
}

STATIC int Stub_MountFile_Failed(const char *rootfs, const char *filepath)
{
    return -1;
}

TEST_F(Test_Fhho, MkDirtestsuccess)
{
    // The test create directory contains the parent directory
    mode_t mode = 0755;
    char *dir = "/home";
    int ret = MkDir(dir, mode);
    EXPECT_EQ(-1, ret);
}

TEST_F(Test_Fhho, StatusThreeMakeDirWithParent)
{
    char *pathData = "/path/abc/abcd";
    mode_t mode = 0755;
    char *path = NULL;
    path = strdup(pathData);
    MOCKER(CheckDirExists).stubs().will(invoke(Stub_CheckDirExists_Failed));
    MOCKER(MkDir).stubs().will(invoke(stub_MkDir_success));
    int ret = MakeDirWithParent(path, mode);
    ret = MakeDirWithParent(path, mode);
    GlobalMockObject::verify();
    EXPECT_EQ(0, ret);
}

TEST_F(Test_Fhho, StatusThreeMountDir)
{
    MOCKER(CheckDirExists).stubs().will(invoke(Stub_CheckDirExists_Failed));
    MOCKER(MkDir).stubs().will(invoke(stub_MkDir_failed));
    char *rootfs = "/rootfs";
    unsigned long reMountRwFlag = MS_BIND | MS_REMOUNT | MS_RDONLY | MS_NOSUID | MS_NOEXEC;
    int ret = MountDir(rootfs, "/home", reMountRwFlag);
    GlobalMockObject::verify();
    EXPECT_EQ(-1, ret);
}

FILE* CreateTempFile(const char* content, size_t size, char **tempFilename)
{
    char filename[] = "/tmp/test_XXXXXX";
    int fd = mkstemp(filename);
    if (fd == -1) return NULL;

    if (write(fd, content, size) != (ssize_t)size) {
        close(fd);
        unlink(filename);
        return NULL;
    }
    lseek(fd, 0, SEEK_SET);
    *tempFilename = strdup(filename);
    return fdopen(fd, "r+");
}

void CloseFile(FILE* fp, char *tempFilename)
{
    if (fp) {
        char filename[256];
        snprintf_s(filename, sizeof(filename), sizeof(filename)-1, "/proc/self/fd/%d", fileno(fp));
        fclose(fp);
        unlink(filename);
    }
    unlink(tempFilename);
    free(tempFilename);
    tempFilename = NULL;
}

TEST_F(Test_Fhho, WriteLogFileTestForFileIsNull)
{
    char *tempFilename = NULL;
    FILE* fp = CreateTempFile(NULL, 0, &tempFilename);
    WriteLogFile(tempFilename, 0, NULL, 0);
    long contentSize = GetLogSizeProcess(tempFilename);
    EXPECT_EQ(0, contentSize);
    CloseFile(fp, tempFilename);
}

TEST_F(Test_Fhho, CreateLogTestForFilenameIsNull)
{
    char *tempFilename = NULL;
    int ret = CreateLog(tempFilename);
    EXPECT_EQ(-1, ret);
}

TEST_F(Test_Fhho, CreateLogTestSuccess)
{
    char *tempFilename = NULL;
    FILE* fp = CreateTempFile(NULL, 0, &tempFilename);
    char* dest = strdup(tempFilename);
    CloseFile(fp, tempFilename);
    int ret = CreateLog(dest);
    EXPECT_EQ(0, ret);
    CloseFile(NULL, dest);
}

TEST_F(Test_Fhho, GetLogSizeForFilenameIsNull)
{
    char *tempFilename = NULL;
    int ret = GetLogSize(tempFilename);
    EXPECT_EQ(-1, ret);
}

TEST_F(Test_Fhho, GetCurrentLocalTimeForFilenameIsNull)
{
    char *buffer = NULL;
    int ret = GetCurrentLocalTime(buffer, 0);
    EXPECT_EQ(-1, ret);
}

TEST_F(Test_Fhho, LogConvertStorageTestForGetLogSizeFailed)
{
    MOCKER(GetLogSize).stubs().will(invoke(Stub_GetLogSize_Failed));
    char *tempFilename = "/tmp/test.txt";
    bool ret = LogConvertStorage(tempFilename, 0);
    GlobalMockObject::verify();
    EXPECT_EQ(false, ret);
}

TEST_F(Test_Fhho, LogConvertStorageTestForLogLoopFailed)
{
    MOCKER(GetLogSize).stubs().will(invoke(Stub_GetLogSize_Success));
    MOCKER(LogLoop).stubs().will(invoke(Stub_LogLoop_Failed));
    char *tempFilename = "/tmp/test.txt";
    bool ret = LogConvertStorage(tempFilename, 5);
    GlobalMockObject::verify();
    EXPECT_EQ(false, ret);
}

TEST_F(Test_Fhho, DivertAndWriteTestForInsufficientBuffer)
{
    char *tempFilename = "/tmp/test.txt";
    unlink(tempFilename);
    char msg[1025];
    memset_s(msg, sizeof(msg), 'a', sizeof(msg));
    DivertAndWrite(tempFilename, msg, LEVEL_WARN);
    int ret = GetLogSizeProcess(tempFilename);
    EXPECT_EQ(0, ret);
}

TEST_F(Test_Fhho, CheckSrcFileForCheckFileFailed)
{
    MOCKER(CheckExternalFile).stubs().will(invoke(Stub_CheckExternalFile_Failed));
    char *tempFilename = NULL;
    FILE* fp = CreateTempFile(NULL, 0, &tempFilename);
    bool ret = CheckSrcFile(tempFilename);
    CloseFile(fp, tempFilename);
    GlobalMockObject::verify();
    EXPECT_EQ(false, ret);
}

TEST_F(Test_Fhho, MountFileForNullArgs)
{
    int ret = MountFile(NULL, NULL);
    EXPECT_EQ(-1, ret);
}

TEST_F(Test_Fhho, MountFileForAssembleFilePathFailed)
{
    char *rootfs = "/rootfs";
    char filepath[BUF_SIZE];
    memset_s(filepath, sizeof(filepath), 'a', sizeof(filepath));
    int ret = MountFile(rootfs, filepath);
    EXPECT_EQ(-1, ret);
}

TEST_F(Test_Fhho, MountFileForStatFailed)
{
    MOCKER(stat).stubs().will(invoke(stub_stat_failed));
    char *rootfs = "/rootfs";
    char *filepath = "/tmp/test.txt";
    int ret = MountFile(rootfs, filepath);
    GlobalMockObject::verify();
    EXPECT_EQ(0, ret);
}

TEST_F(Test_Fhho, MountFileForMakeMountPointsFailed)
{
    MOCKER(stat).stubs().will(invoke(stub_stat_success));
    MOCKER(MakeMountPoints).stubs().will(invoke(Stub_MakeMountPoints_Failed));
    char *rootfs = "/rootfs";
    char *filepath = "/tmp/test.txt";
    int ret = MountFile(rootfs, filepath);
    GlobalMockObject::verify();
    EXPECT_EQ(-1, ret);
}

TEST_F(Test_Fhho, MountFileForMountFailed)
{
    MOCKER(stat).stubs().will(invoke(stub_stat_success));
    MOCKER(MakeMountPoints).stubs().will(invoke(Stub_MakeMountPoints_Success));
    MOCKER(Mount).stubs().will(invoke(stub_Mount_failed));
    char *rootfs = "/rootfs";
    char *filepath = "/tmp/test.txt";
    int ret = MountFile(rootfs, filepath);
    GlobalMockObject::verify();
    EXPECT_EQ(-1, ret);
}

TEST_F(Test_Fhho, OptionsCmdArgParserForNullArgs)
{
    bool ret = OptionsCmdArgParser(NULL, NULL);
    EXPECT_EQ(false, ret);
}

TEST_F(Test_Fhho, OptionsCmdArgParserForCpyFailed)
{
    struct CmdArgs args;
    char arg[BUF_SIZE+2];
    memset_s(arg, sizeof(arg), 'a', sizeof(arg));
    bool ret = OptionsCmdArgParser(&args, arg);
    EXPECT_EQ(false, ret);
}

TEST_F(Test_Fhho, OptionsCmdArgParserForCmpFailed)
{
    struct CmdArgs args;
    char *arg = "invalid value";
    bool ret = OptionsCmdArgParser(&args, arg);
    EXPECT_EQ(false, ret);
}

TEST_F(Test_Fhho, OptionsCmdArgParserSuccess)
{
    struct CmdArgs args;
    char *arg = "NODRV";
    bool ret = OptionsCmdArgParser(&args, arg);
    EXPECT_EQ(true, ret);
}

TEST_F(Test_Fhho, MountFileCmdArgParserForNullArgs)
{
    bool ret = MountFileCmdArgParser(NULL, NULL);
    EXPECT_EQ(false, ret);
}

TEST_F(Test_Fhho, MountFileCmdArgParserForFileCountExceed)
{
    struct CmdArgs args;
    args.files.count = MAX_MOUNT_NR + 1;
    char *arg = "NODRV";
    bool ret = MountFileCmdArgParser(&args, arg);
    EXPECT_EQ(false, ret);
}

TEST_F(Test_Fhho, MountFileCmdArgParserForCpyFailed)
{
    struct CmdArgs args;
    memset_s(&args, sizeof(struct CmdArgs), 0, sizeof(struct CmdArgs));
    char arg[PATH_MAX+2];
    memset_s(arg, sizeof(arg), 'a', sizeof(arg));
    bool ret = MountFileCmdArgParser(&args, arg);
    EXPECT_EQ(false, ret);
}

TEST_F(Test_Fhho, MountFileCmdArgParserForCheckFileLegalityFailed)
{
    struct CmdArgs args;
    memset_s(&args, sizeof(struct CmdArgs), 0, sizeof(struct CmdArgs));
    char *arg = "/tmp/##.txt";
    bool ret = MountFileCmdArgParser(&args, arg);
    EXPECT_EQ(false, ret);
}

TEST_F(Test_Fhho, MountDirCmdArgParserForNullArgs)
{
    bool ret = MountDirCmdArgParser(NULL, NULL);
    EXPECT_EQ(false, ret);
}

TEST_F(Test_Fhho, MountDirCmdArgParserForFileCountExceed)
{
    struct CmdArgs args;
    args.dirs.count = MAX_MOUNT_NR + 1;
    char *arg = "NODRV";
    bool ret = MountDirCmdArgParser(&args, arg);
    EXPECT_EQ(false, ret);
}

TEST_F(Test_Fhho, MountDirCmdArgParserForCpyFailed)
{
    struct CmdArgs args;
    memset_s(&args, sizeof(struct CmdArgs), 0, sizeof(struct CmdArgs));
    char arg[PATH_MAX+2];
    memset_s(arg, sizeof(arg), 'a', sizeof(arg));
    bool ret = MountDirCmdArgParser(&args, arg);
    EXPECT_EQ(false, ret);
}

TEST_F(Test_Fhho, MountDirCmdArgParserForCheckFileLegalityFailed)
{
    struct CmdArgs args;
    memset_s(&args, sizeof(struct CmdArgs), 0, sizeof(struct CmdArgs));
    char *arg = "/tmp/##/test";
    bool ret = MountDirCmdArgParser(&args, arg);
    EXPECT_EQ(false, ret);
}

TEST_F(Test_Fhho, LinkCheckCmdArgParserForNullArgs)
{
    bool ret = LinkCheckCmdArgParser(NULL);
    EXPECT_EQ(false, ret);
}

TEST_F(Test_Fhho, LinkCheckCmdArgParserForNotAllowLink)
{
    char *allowLink = "False";
    bool ret = LinkCheckCmdArgParser(allowLink);
    EXPECT_EQ(true, ret);
}

TEST_F(Test_Fhho, LinkCheckCmdArgParserForInvalidAllowLink)
{
    char *allowLink = "InvalidValue";
    bool ret = LinkCheckCmdArgParser(allowLink);
    EXPECT_EQ(false, ret);
}

#endif

TEST_F(Test_Fhho, StatusTwoMountDir)
{
    MOCKER(CheckDirExists).stubs().will(invoke(Stub_CheckDirExists_Failed));
    MOCKER(MakeDirWithParent).stubs().will(invoke(Stub_MakeDirWithParent_Failed));
    char *rootfs = "/rootfs";
    unsigned long reMountRwFlag = MS_BIND | MS_REMOUNT | MS_RDONLY | MS_NOSUID | MS_NOEXEC;
    int ret = MountDir(rootfs, "/home", reMountRwFlag);
    GlobalMockObject::verify();
    EXPECT_EQ(-1, ret);
}

TEST_F(Test_Fhho, StatusFourMountDir)
{
    MOCKER(CheckDirExists).stubs().will(invoke(Stub_CheckDirExists_Failed));
    MOCKER(MakeDirWithParent).stubs().will(invoke(Stub_MakeDirWithParent_Success));
    MOCKER(mount).stubs().will(invoke(stub_mount_failed));
    char *rootfs = "/rootfs";
    unsigned long reMountRwFlag = MS_BIND | MS_REMOUNT | MS_RDONLY | MS_NOSUID | MS_NOEXEC;
    int ret = MountDir(rootfs, "/home", reMountRwFlag);
    GlobalMockObject::verify();
    EXPECT_EQ(-1, ret);
}

TEST_F(Test_Fhho, StatusFiveMountDir)
{
    MOCKER(stat).stubs().will(invoke(stub_stat_failed));
    MOCKER(MakeDirWithParent).stubs().will(invoke(Stub_MakeDirWithParent_Success));
    MOCKER(Mount).stubs().will(invoke(stub_Mount_success));
    char *rootfs = "/rootfs";
    unsigned long reMountRwFlag = MS_BIND | MS_REMOUNT | MS_RDONLY | MS_NOSUID | MS_NOEXEC;
    int ret = MountDir(rootfs, "/dev/random", reMountRwFlag);
    GlobalMockObject::verify();
    EXPECT_EQ(0, ret);
}

TEST_F(Test_Fhho, StatusSixMountDir)
{
    unsigned long reMountRwFlag = MS_BIND | MS_REMOUNT | MS_RDONLY | MS_NOSUID | MS_NOEXEC;
    int ret = MountDir(nullptr, nullptr, reMountRwFlag);
    EXPECT_EQ(-1, ret);
}

TEST_F(Test_Fhho, SetupContainerTestForArgsIsNull)
{
    int ret = SetupContainer(nullptr);
    EXPECT_EQ(-1, ret);
}

TEST_F(Test_Fhho, StatusOneSetupContainer)
{
    struct CmdArgs args;
    (void)strcpy_s(args.rootfs, sizeof(args.rootfs), "/home");
    args.pid = 1;
    MOCKER(DoPrepare).stubs().will(invoke(Stub_DoPrepare_Failed));
    int ret = SetupContainer(&args);
    GlobalMockObject::verify();
    EXPECT_EQ(-1, ret);
}

TEST_F(Test_Fhho, SetupContainerForEnterNsByPathFailed)
{
    struct CmdArgs args;
    (void)strcpy_s(args.rootfs, sizeof(args.rootfs), "/home");
    args.pid = 1;
    MOCKER(DoPrepare).stubs().will(invoke(Stub_DoPrepare_Success));
    MOCKER(EnterNsByPath).stubs().will(invoke(Stub_EnterNsByPath_Failed));
    int ret = SetupContainer(&args);
    GlobalMockObject::verify();
    EXPECT_EQ(-1, ret);
}

TEST_F(Test_Fhho, SetupContainerForDoMountingFailed)
{
    struct CmdArgs args;
    (void)strcpy_s(args.rootfs, sizeof(args.rootfs), "/home");
    args.pid = 1;
    MOCKER(DoPrepare).stubs().will(invoke(Stub_DoPrepare_Success));
    MOCKER(EnterNsByPath).stubs().will(invoke(Stub_EnterNsByPath_Success));
    MOCKER(DoMounting).stubs().will(invoke(Stub_DoMounting_Failed));
    int ret = SetupContainer(&args);
    GlobalMockObject::verify();
    EXPECT_EQ(-1, ret);
}

TEST_F(Test_Fhho, SetupContainerForEnterNsByFdFailed)
{
    struct CmdArgs args;
    (void)strcpy_s(args.rootfs, sizeof(args.rootfs), "/home");
    args.pid = 1;
    MOCKER(DoPrepare).stubs().will(invoke(Stub_DoPrepare_Success));
    MOCKER(EnterNsByPath).stubs().will(invoke(Stub_EnterNsByPath_Success));
    MOCKER(DoMounting).stubs().will(invoke(Stub_DoMounting_Success));
    MOCKER(EnterNsByFd).stubs().will(invoke(Stub_EnterNsByFd_Failed));
    int ret = SetupContainer(&args);
    GlobalMockObject::verify();
    EXPECT_EQ(-1, ret);
}

TEST_F(Test_Fhho, SetupContainerSuccess)
{
    struct CmdArgs args;
    (void)strcpy_s(args.rootfs, sizeof(args.rootfs), "/home");
    args.pid = 1;
    MOCKER(DoPrepare).stubs().will(invoke(Stub_DoPrepare_Success));
    MOCKER(EnterNsByPath).stubs().will(invoke(Stub_EnterNsByPath_Success));
    MOCKER(DoMounting).stubs().will(invoke(Stub_DoMounting_Success));
    MOCKER(EnterNsByFd).stubs().will(invoke(Stub_EnterNsByFd_Success));
    int ret = SetupContainer(&args);
    GlobalMockObject::verify();
    EXPECT_EQ(0, ret);
}

TEST_F(Test_Fhho, StatusOneProcess)
{
    // test parameter is null
    int argc = 0;
    char **argv = NULL;
    int ret = Process(argc, argv);
    EXPECT_EQ(-1, ret);
}

TEST_F(Test_Fhho, StatusTwoProcess)
{
    // Test the correct options
    const int argc = 7;
    const char *argvData[argc] = {"ascend-docker-cli", "--allow-link", "True", "--pid", "123", "--rootfs", "/home"};
    MOCKER(SetupContainer).stubs().will(invoke(Stub_SetupContainer_Success));
    int ret = Process(argc, const_cast<char **>(argvData));
    EXPECT_EQ(0, ret);
}

TEST_F(Test_Fhho, StatusThreeProcess)
{
    // Test error options
    const int argc = 7;
    const char *argvData[argc] = {"ascend-docker-cli", "--evices", "1,2", "--idd", "123", "--ootfs", "/home"};
    int ret = Process(argc, const_cast<char **>(argvData));
    EXPECT_EQ(-1, ret);
}

TEST_F(Test_Fhho, StatusFourProcess)
{
    const int argc = 7;
    const char *argvData[argc] = {"ascend-docker-cli", "--evices", "1,2", "--idd", "123", "--ootfs", "/home"};
    MOCKER(SetupContainer).stubs().will(invoke(Stub_SetupContainer_Success));
    int ret = Process(argc, const_cast<char **>(argvData));
    GlobalMockObject::verify();
    EXPECT_EQ(-1, ret);
}

TEST_F(Test_Fhho, StatusFiveProcess)
{
    const int argc = 11;
    const char *argvData[argc] = {"ascend-docker-cli", "--pid", "123", "--rootfs",
                                  "/home", "--options", "base", "--mount-file", "a.list", "--mount-dir", "/home/code"};
    MOCKER(SetupContainer).stubs().will(invoke(Stub_SetupContainer_Success));
    int ret = Process(argc, const_cast<char **>(argvData));
    GlobalMockObject::verify();
    EXPECT_EQ(-1, ret);
}

TEST_F(Test_Fhho, StatusSixProcess)
{
    const int argc = 11;
    const char *argvData[argc] = {"ascend-docker-cli", "--pid", "123", "--rootfs",
                                  "/home", "--opt", "base", "--mount-f", "a.list", "--mount-dir", "/root/sxv"};
    MOCKER(SetupContainer).stubs().will(invoke(Stub_SetupContainer_Success));
    int ret = Process(argc, const_cast<char **>(argvData));
    GlobalMockObject::verify();
    EXPECT_EQ(-1, ret);
}

TEST_F(Test_Fhho, StatusSevenProcess)
{
    const int argc = 11;
    const char *argvData[argc] = {"ascend-docker-cli", "--ops", "--pid", "123",
                                  "--rootfs", "/home", "base", "--mounle", "a.list", "--mount-dir", "/home/code"};
    MOCKER(SetupContainer).stubs().will(invoke(Stub_SetupContainer_Success));
    int ret = Process(argc, const_cast<char **>(argvData));
    GlobalMockObject::verify();
    EXPECT_EQ(-1, ret);
}

TEST_F(Test_Fhho, StatusEightProcess)
{
    const int argc = 2048;
    char *argvData[argc] = {"ascend-docker-cli", "--ops", "--pid", "123",
                            "--rootfs", "/home", "base", "--mounle", "a.list", "--mount-dir", "/home/code"};
    int ret = Process(argc, argvData);
    EXPECT_EQ(-1, ret);
}

TEST_F(Test_Fhho, StatusOneParseRuntimeOptions)
{
    // Test the right options
    const char options[BUF_SIZE] = "1,2";
    // Options is the parameter value of -o
    int ret = ParseRuntimeOptions(options);
    EXPECT_EQ(0, ret);
}

TEST_F(Test_Fhho, StatusTwoParseRuntimeOptions)
{
    int ret = ParseRuntimeOptions(NULL);
    EXPECT_EQ(0, ret);
}

TEST_F(Test_Fhho, StatusThreeDoPrepare)
{
    MOCKER(GetNsPath).stubs().will(invoke(Stub_GetNsPath_Failed));
    struct CmdArgs args;
    (void)strcpy_s(args.rootfs, sizeof(args.rootfs), "/home");
    args.pid = 1;
    struct ParsedConfig config;
    int ret = DoPrepare(&args, &config);
    GlobalMockObject::verify();
    EXPECT_EQ(-1, ret);
}

TEST_F(Test_Fhho, GetFileSubsetAndCheckOne)
{
    bool ret = GetFileSubsetAndCheck("", 10);
    EXPECT_EQ(false, ret);
}

TEST_F(Test_Fhho, GetFileSubsetAndCheckTwo)
{
    bool ret = GetFileSubsetAndCheck("./", 10);
    EXPECT_EQ(true, ret);
}

TEST_F(Test_Fhho, GetFileSubsetAndCheckThree)
{
    bool ret = GetFileSubsetAndCheck(nullptr, 0);
    EXPECT_EQ(false, ret);
}

TEST_F(Test_Fhho, CheckExistsFileOne)
{
    bool ret = CheckExistsFile("", -1, 1, false);
    EXPECT_EQ(true, ret);
}

TEST_F(Test_Fhho, CheckExistsFileTwo)
{
    bool ret = CheckExistsFile("./gtest_mytest.cpp", strlen("./gtest_mytest.cpp"), 1, false);
    EXPECT_EQ(true, ret);
}

TEST_F(Test_Fhho, CheckExistsFileThree)
{
    bool ret = CheckExistsFile(nullptr, 0, 1, false);
    EXPECT_EQ(false, ret);
}

TEST_F(Test_Fhho, VerifyPathInfoOne)
{
    int ret = VerifyPathInfo(NULL);
    EXPECT_EQ(-1, ret);
}

TEST_F(Test_Fhho, GetParentPathStrOne)
{
    int ret = GetParentPathStr(NULL, NULL, 0);
    EXPECT_EQ(-1, ret);
}

TEST_F(Test_Fhho, GetParentPathStrTwo)
{
    int ret = GetParentPathStr(".", "../", 0);
    EXPECT_EQ(0, ret);
}

TEST_F(Test_Fhho, CheckExternalFileOne)
{
    int ret = CheckExternalFile("./main.cpp", strlen("./main.cpp"), 10, false);
    EXPECT_EQ(0, ret);
}

TEST_F(Test_Fhho, LogLoopOne)
{
    int ret = LogLoop(NULL);
    EXPECT_EQ(-1, ret);
}

TEST_F(Test_Fhho, LogLoopTwo)
{
    int ret = LogLoop("../test_log_not_exist.log");
    EXPECT_EQ(-1, ret);
}

TEST_F(Test_Fhho, LogLoopThree)
{
    const int levelDebug = 3;
    const int screenNo = 0;
    Logger("develop test", levelDebug, screenNo);
    int ret = LogLoop("test_log.log");
    EXPECT_EQ(-1, ret);
}

TEST_F(Test_Fhho, IsOptionNoDrvSetOne)
{
    const char options[BUF_SIZE] = "1,2";
    ParseRuntimeOptions(options);
    bool ret = IsOptionNoDrvSet();
    EXPECT_EQ(false, ret);
}

TEST_F(Test_Fhho, IsVirtualOne)
{
    const char options[BUF_SIZE] = "1,2";
    ParseRuntimeOptions(options);
    bool ret = IsVirtual();
    EXPECT_EQ(false, ret);
}

TEST_F(Test_Fhho, CheckOpenedFileOne)
{
    EXPECT_FALSE(CheckOpenedFile(nullptr, BUF_SIZE, false));
}

TEST_F(Test_Fhho, CheckOpenedFileTwo)
{
    const long maxSize = 5;
    char *tempFilename = nullptr;
    FILE* fp = CreateTempFile("1234567890", 10, &tempFilename);
    EXPECT_FALSE(CheckOpenedFile(fp, maxSize, false));
    CloseFile(fp, tempFilename);
}

TEST_F(Test_Fhho, CheckOpenedFileThree)
{
    const long maxSize = 1024;
    char *tempFilename = nullptr;
    FILE* fp = CreateTempFile("test", 4, &tempFilename);
    EXPECT_TRUE(CheckOpenedFile(fp, maxSize, false));
    CloseFile(fp, tempFilename);
}

TEST_F(Test_Fhho, MountTestForSrcIsNull)
{
    int ret = Mount(nullptr, nullptr);
    EXPECT_EQ(-1, ret);
}

TEST_F(Test_Fhho, MountTestForMountFailedCaseOne)
{
    MOCKER(CheckSrcFile).stubs().will(invoke(Stub_CheckSrcFile_Success));
    MOCKER(mount).stubs().will(invoke(stub_mount_failed));
    char *src = "/tmp/test.txt";
    char *dst = "/tmp/test.txt";
    int ret = Mount(src, dst);
    GlobalMockObject::verify();
    EXPECT_EQ(-1, ret);
}

TEST_F(Test_Fhho, MountTestForMountFailedCaseTwo)
{
    MOCKER(CheckSrcFile).stubs().will(invoke(Stub_CheckSrcFile_Success));
    MOCKER(mount).stubs().will(invoke(stub_mount_src_nil_failed));
    char *src = "/tmp/test.txt";
    char *dst = "/tmp/test.txt";
    int ret = Mount(src, dst);
    GlobalMockObject::verify();
    EXPECT_EQ(-1, ret);
}

TEST_F(Test_Fhho, MountTestForMountFailedCaseThree)
{
    MOCKER(CheckSrcFile).stubs().will(invoke(Stub_CheckSrcFile_Success));
    MOCKER(mount).stubs().will(invoke(stub_mount_success));
    char *src = "/tmp/test.txt";
    char *dst = "/tmp/test.txt";
    int ret = Mount(src, dst);
    GlobalMockObject::verify();
    EXPECT_EQ(0, ret);
}

TEST_F(Test_Fhho, DoFileMountingForArgsIsNull)
{
    int ret = DoFileMounting(nullptr, nullptr);
    EXPECT_EQ(-1, ret);
}

TEST_F(Test_Fhho, DoFileMountingForMountFileFailed)
{
    MOCKER(MountFile).stubs().will(invoke(Stub_MountFile_Failed));
    char *rootfs = "/rootfs";
    struct MountList list = {0};
    list.count = 1;
    int ret = DoFileMounting(rootfs, &list);
    GlobalMockObject::verify();
    EXPECT_EQ(-1, ret);
}

TEST_F(Test_Fhho, DoFileMountingForMountListIsEmpty)
{
    char *rootfs = "/rootfs";
    struct MountList list = {0};
    int ret = DoFileMounting(rootfs, &list);
    EXPECT_EQ(0, ret);
}

TEST_F(Test_Fhho, DoMountingForArgsIsNull)
{
    int ret = DoMounting(nullptr);
    EXPECT_EQ(-1, ret);
}

TEST_F(Test_Fhho, DoMountingForOptionNoDrvSet)
{
    MOCKER(IsOptionNoDrvSet).stubs().will(invoke(Stub_IsOptionNoDrvSet_True));
    struct ParsedConfig config = {0};
    int ret = DoMounting(&config);
    GlobalMockObject::verify();
    EXPECT_EQ(0, ret);
}

TEST_F(Test_Fhho, DoMountingForMountFileFailed)
{
    MOCKER(IsOptionNoDrvSet).stubs().will(invoke(Stub_IsOptionNoDrvSet_False));
    struct ParsedConfig config = {0};
    int ret = DoMounting(&config);
    GlobalMockObject::verify();
    EXPECT_EQ(-1, ret);
}

TEST_F(Test_Fhho, DoMountingForMountDirsFailed)
{
    MOCKER(IsOptionNoDrvSet).stubs().will(invoke(Stub_IsOptionNoDrvSet_False));
    MOCKER(DoFileMounting).stubs().will(invoke(Stub_DoFileMounting_Success));
    struct ParsedConfig config = {0};
    int ret = DoMounting(&config);
    GlobalMockObject::verify();
    EXPECT_EQ(-1, ret);
}