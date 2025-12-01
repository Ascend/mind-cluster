/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
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

#include "test_hlog.h"
#include "service_configure.h"

using namespace ock::common::config;

static const auto MAX_LOG_FILE_SIZE = 10 * 1024 * 1024;
static const auto MAX_LOG_FILE_COUNT = 40;

void TestHlog::SetUpTestSuite() {}

void TestHlog::TearDownTestSuite()
{
    LoggerInitialize();
}

void TestHlog::SetUp()
{
    Hlog::DestroyLogInstances();
}

void TestHlog::TearDown()
{
    Hlog::DestroyLogInstances();
}

int TestHlog::LoggerInitialize()
{
    static int times = 0;

    times++;
    std::string path = ServiceConfigure::GetInstance().GetWorkPath();
    if (path.empty()) {
        std::cerr << "get local service path failed:" << std::endl;
        return -1;
    }

    path.append("/log");
    auto ret = mkdir(path.c_str(), 0755);
    if (ret != 0 && errno != EEXIST) {
        std::cerr << "create logger path failed : " << strerror(errno) << std::endl;
        return -1;
    }

    auto auditLogPath = path;
    auditLogPath.append("/").append(std::to_string(times)).append("-ock-memfs-audit.log");
    std::cout << "create audit log" << std::endl;
    ret = Hlog::CreateInstanceAudit(auditLogPath.c_str(), MAX_LOG_FILE_SIZE, MAX_LOG_FILE_COUNT);
    if (ret != 0) {
        std::cerr << "create HLOG AuditLog instance failed : " << ret << std::endl;
        return -1;
    }

    auto logPath = path;
    logPath.append("/").append(std::to_string(times)).append("-ock-memfs.log");
    std::cout << "create log" << std::endl;
    ret = Hlog::CreateInstance(1, 0, logPath.c_str(), MAX_LOG_FILE_SIZE, MAX_LOG_FILE_COUNT);
    if (ret != 0) {
        std::cerr << "create HLOG instance failed : " << ret << std::endl;
        return -1;
    }

    std::cout << "create log success" << std::endl;
    return 0;
}

using namespace ock::hlog;

namespace {

TEST_F(TestHlog, test_create_instace_error_param_should_fail)
{
    Hlog::gLogger = nullptr;
    int ret = Hlog::CreateInstance(-1, 0, nullptr, 0, 0);
    ASSERT_EQ(HLOG_ERR_INVALID_PARAM, ret);

    Hlog::gLogger = nullptr;
    ret = Hlog::CreateInstance(0, -1, nullptr, 0, 0);
    ASSERT_EQ(HLOG_ERR_INVALID_PARAM, ret);
    Hlog::GetLastErrorMessage();

    Hlog::gLogger = nullptr;
    ret = Hlog::CreateInstance(1, 0, nullptr, 0, 0);
    ASSERT_EQ(HLOG_ERR_INVALID_PARAM, ret);

    Hlog::gLogger = nullptr;
    const char *path = "/tmp/test.log";
    ret = Hlog::CreateInstance(1, 0, path, 0, 0);
    ASSERT_EQ(HLOG_ERR_INVALID_PARAM, ret);

    Hlog::gLogger = nullptr;
    int fileSize = 2 * 1024 * 1024;
    ret = Hlog::CreateInstance(1, 0, path, fileSize, 0);
    ASSERT_EQ(HLOG_ERR_INVALID_PARAM, ret);
}

TEST_F(TestHlog, test_create_audit_instace_error_param_should_fail)
{
    Hlog::gAuditLogger = nullptr;
    int ret = Hlog::CreateInstanceAudit(nullptr, 0, 0);
    ASSERT_EQ(HLOG_ERR_INVALID_PARAM, ret);

    Hlog::gAuditLogger = nullptr;
    const char *path = "/tmp/test.log";
    ret = Hlog::CreateInstanceAudit(path, 0, 0);
    ASSERT_EQ(HLOG_ERR_INVALID_PARAM, ret);

    Hlog::gAuditLogger = nullptr;
    int fileSize = 2 * 1024 * 1024;
    ret = Hlog::CreateInstanceAudit(path, fileSize, 0);
    ASSERT_EQ(HLOG_ERR_INVALID_PARAM, ret);
}

TEST_F(TestHlog, test_message_error_param_should_fail)
{
    int ret = Hlog::LogMessage(0, nullptr, nullptr);
    ASSERT_EQ(HLOG_ERR_INVALID_PARAM, ret);

    const char *prefix = "/tmp/";
    ret = Hlog::LogMessage(0, prefix, nullptr);
    ASSERT_EQ(HLOG_ERR_INVALID_PARAM, ret);

    const char *message = "test log message";
    ret = Hlog::LogMessage(0, prefix, message);
    ASSERT_EQ(HLOG_ERR_NOT_INITIALIZED, ret);
}

TEST_F(TestHlog, test_audit_message_error_param_should_fail)
{
    int ret = Hlog::LogAuditMessage(0, nullptr, nullptr);
    ASSERT_EQ(HLOG_ERR_INVALID_PARAM, ret);

    const char *prefix = "/tmp/";
    ret = Hlog::LogAuditMessage(0, prefix, nullptr);
    ASSERT_EQ(HLOG_ERR_INVALID_PARAM, ret);

    const char *message = "test log message";
    ret = Hlog::LogAuditMessage(0, prefix, message);
    ASSERT_EQ(HLOG_ERR_NOT_INITIALIZED, ret);
}

TEST_F(TestHlog, test_set_pattern_error_param_should_fail)
{
    int ret = Hlog::SetInstancePattern(nullptr);
    ASSERT_EQ(HLOG_ERR_INVALID_PARAM, ret);

    ret = Hlog::SetInstancePattern("[%Y-%m-%d]%v");
    ASSERT_EQ(HLOG_ERR_NOT_INITIALIZED, ret);
}

TEST_F(TestHlog, test_create_instace_should_success)
{
    const char *path = "/tmp/test.log";
    int fileSize = 2 * 1024 * 1024;

    int ret = Hlog::CreateInstance(0, 0, path, fileSize, 5);
    ASSERT_EQ(HLOG_OK, ret);

    ret = Hlog::gLogger->IsDebug();
    ASSERT_EQ(1, ret);
    ret = Hlog::gLogger->IsHigherLevel(1);
    ASSERT_EQ(1, ret);

    Hlog::DestroyLogInstances();
    ret = Hlog::CreateInstance(1, 0, path, fileSize, 5);
    ASSERT_EQ(HLOG_OK, ret);
}


TEST_F(TestHlog, test_create_audit_instace_should_success)
{
    const char *path = "/tmp/test.audit.log";
    int fileSize = 2 * 1024 * 1024;
    Hlog::gAuditLogger = nullptr;
    int ret = Hlog::CreateInstanceAudit(path, fileSize, 5);
    ASSERT_EQ(HLOG_OK, ret);
}

TEST_F(TestHlog, test_message_should_success)
{
    const char *message = "test log message";
    int ret = Hlog::LogMessage(0, "test", message);
    ASSERT_NE(0, ret);
}

TEST_F(TestHlog, test_log_should_success)
{
    const char *message = "test log message";
    int ret = Hlog::gLogger->Log(10, message);
    ASSERT_EQ(HLOG_ERR_INVALID_LEVEL, ret);

    ret = Hlog::gLogger->Log(1, message);
    ASSERT_EQ(HLOG_OK, ret);

    ret = Hlog::gLogger->Log(1, "pre-", message);
    ASSERT_EQ(HLOG_OK, ret);

    ret = Hlog::gLogger->DebugLog("this is debug log");
    ASSERT_EQ(HLOG_OK, ret);

    ret = Hlog::gLogger->AuditLog("this is audit log");
    ASSERT_EQ(HLOG_OK, ret);

    std::string pattern = "[%Y-%m-%d]%v";
    Hlog::gLogger->SetPattern(pattern);
    Hlog::gLogger->FlushAudit();
}
}