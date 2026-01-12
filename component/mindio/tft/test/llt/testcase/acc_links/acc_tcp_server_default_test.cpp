/*
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

#include "acc_file_validator.h"
#define private public
#include "acc_tcp_server_default.h"
#undef private

using namespace ock::acc;

class AccTcpServerDefaultTest : public testing::Test {
public:
    void SetUp() override;
    void TearDown() override;
};

void AccTcpServerDefaultTest::SetUp() {}

void AccTcpServerDefaultTest::TearDown()
{
    GlobalMockObject::verify();
}

TEST_F(AccTcpServerDefaultTest, server_repeat_start_stop)
{
    AccTcpServerDefault server{};
    server.StopAfterFork();

    server.sslHelper_ = AccMakeRef<AccTcpSslHelper>();
    server.started_.store(true);
    AccTcpServerOptions opts;
    AccTlsOption tlsOpts;

    auto ret = server.Start(opts, tlsOpts);
    ASSERT_EQ(0, ret);
    server.StopAfterFork();
}

TEST_F(AccTcpServerDefaultTest, server_load_dl_is_symlink)
{
    AccTcpServerDefault server{};

    MOCKER(FileValidator::IsSymlink).stubs().will(returnValue(true));
    auto ret = server.LoadDynamicLib("test");
    ASSERT_EQ(-1, ret);
}

TEST_F(AccTcpServerDefaultTest, server_load_dl_realpath_failed)
{
    AccTcpServerDefault server{};

    MOCKER(FileValidator::IsSymlink).stubs().will(returnValue(false));
    MOCKER(FileValidator::Realpath).stubs().will(returnValue(false));
    auto ret = server.LoadDynamicLib("test");
    ASSERT_EQ(-1, ret);
}

TEST_F(AccTcpServerDefaultTest, server_load_dl_is_not_dir)
{
    AccTcpServerDefault server{};

    MOCKER(FileValidator::IsSymlink).stubs().will(returnValue(false));
    MOCKER(FileValidator::Realpath).stubs().will(returnValue(true));
    MOCKER(FileValidator::IsDir).stubs().will(returnValue(false));
    auto ret = server.LoadDynamicLib("test");
    ASSERT_EQ(-1, ret);
}

TEST_F(AccTcpServerDefaultTest, server_load_dl_normal)
{
    AccTcpServerDefault server{};

    MOCKER(FileValidator::IsSymlink).stubs().will(returnValue(false));
    MOCKER(FileValidator::Realpath).stubs().will(returnValue(true));
    MOCKER(FileValidator::IsDir).stubs().will(returnValue(true));
    MOCKER(OpenSslApiWrapper::Load).stubs().will(returnValue(0));
    auto ret = server.LoadDynamicLib("test");
    ASSERT_EQ(0, ret);
}

TEST_F(AccTcpServerDefaultTest, validate_option_listenIp_empty)
{
    AccTcpServerDefault server{};
    AccTcpServerOptions opts;
    opts.enableListener = true;
    opts.listenIp = "";
    server.options_ = opts;
    auto ret = server.ValidateOptions();
    ASSERT_EQ(-4, ret);
}

TEST_F(AccTcpServerDefaultTest, validate_option_max_world_size_zero)
{
    AccTcpServerDefault server{};
    AccTcpServerOptions opts;
    opts.enableListener = false;
    opts.workerCount = 1;
    opts.workerStartCpuId = 0;
    opts.linkSendQueueSize = UNO_48;
    opts.keepaliveIdleTime = 1;
    opts.keepaliveProbeTimes = 1;
    opts.keepaliveProbeInterval = 1;
    opts.maxWorldSize = 0;

    server.options_ = opts;
    auto ret = server.ValidateOptions();
    ASSERT_EQ(-4, ret);
}

TEST_F(AccTcpServerDefaultTest, validate_option_tls_opts_invalid)
{
    AccTcpServerDefault server{};
    AccTcpServerOptions opts;
    opts.enableListener = false;
    opts.workerCount = 1;
    opts.workerStartCpuId = 0;
    opts.linkSendQueueSize = UNO_48;
    opts.keepaliveIdleTime = 1;
    opts.keepaliveProbeTimes = 1;
    opts.keepaliveProbeInterval = 1;
    opts.maxWorldSize = 1;

    server.options_ = opts;
    MOCKER(AccCommonUtil::CheckTlsOptions).stubs().will(returnValue(-1));

    auto ret = server.ValidateOptions();
    ASSERT_EQ(-4, ret);
}

TEST_F(AccTcpServerDefaultTest, start_workers_failed)
{
    AccTcpServerDefault server{};
    AccTcpServerOptions opts;
    opts.enableListener = false;
    opts.workerCount = 1;
    opts.workerStartCpuId = 0;
    opts.linkSendQueueSize = UNO_48;
    opts.keepaliveIdleTime = 1;
    opts.keepaliveProbeTimes = 1;
    opts.keepaliveProbeInterval = 1;
    opts.maxWorldSize = 1;

    AccTcpWorkerOptions workerOptions;
    AccTcpWorkerPtr worker = new (std::nothrow) AccTcpWorker(workerOptions);

    server.options_ = opts;
    server.workers_.emplace_back(worker);

    auto ret = server.StartWorkers();
    ASSERT_EQ(-4, ret);
}

TEST_F(AccTcpServerDefaultTest, already_start_delay_cleanup)
{
    AccTcpServerDefault server{};
    server.delayCleanup_ = new (std::nothrow) AccTcpLinkDelayCleanup();

    auto ret = server.StartDelayCleanup();
    ASSERT_TRUE(ret);
}

TEST_F(AccTcpServerDefaultTest, start_delay_cleanup_failed)
{
    AccTcpServerDefault server{};
    union MockerHelper {
        int32_t (AccTcpLinkDelayCleanup::*start)();
        int32_t (*mockStart)(AccTcpLinkDelayCleanup *self);
    };
    MockerHelper helper{};
    helper.start = &AccTcpLinkDelayCleanup::Start;
    MOCKCPP_NS::mockAPI("&AccTcpLinkDelayCleanup::Start", helper.mockStart)
        .stubs().will(returnValue(-1));

    auto ret = server.StartDelayCleanup();
    ASSERT_EQ(-1, ret);
}

TEST_F(AccTcpServerDefaultTest, handle_new_connection_failed)
{
    AccTcpServerDefault server{};
    AccTcpServerOptions opts;
    AccConnReq req;
    AccTcpLinkComplexDefaultPtr newLink = new (std::nothrow) AccTcpLinkComplexDefault(0, "", 0);

    server.options_ = opts;
    // version not equal
    req.version = 1;
    auto ret = server.HandleNewConnection(req, newLink);
    ASSERT_EQ(-1, ret);

    // worker empty
    req.version = 0;
    ret = server.HandleNewConnection(req, newLink);
    ASSERT_EQ(-1, ret);
}

TEST_F(AccTcpServerDefaultTest, generate_ssl_ctx_failed)
{
    AccTcpServerDefault server{};
    AccTlsOption opts;
    opts.enableTls = true;
    server.tlsOption_ = opts;

    MOCKER(OpenSslApiWrapper::TlsMethod).stubs().will(returnValue(static_cast<const SSL_METHOD*>(nullptr)));
    MOCKER(OpenSslApiWrapper::SslCtxNew).stubs().will(returnValue(static_cast<SSL_CTX*>(nullptr)));
    auto ret = server.GenerateSslCtx();
    ASSERT_EQ(-3, ret);
}

TEST_F(AccTcpServerDefaultTest, create_validate_ssl_link_failed)
{
    AccTcpServerDefault server{};
    AccTlsOption opts;
    opts.enableTls = true;
    server.tlsOption_ = opts;
    SSL *ssl = nullptr;
    int fd = -1;
    MOCKER(AccTcpSslHelper::NewSslLink).stubs().will(returnValue(-1));

    auto ret = server.CreateSSLLink(ssl, fd);
    ASSERT_EQ(-2, ret);
    server.ValidateSSLLink(ssl, fd);
}