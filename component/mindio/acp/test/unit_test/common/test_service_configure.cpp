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

#include <gtest/gtest.h>
#include <mockcpp/mokc.h>

#include "service_configure.h"

using namespace ock::memfs;
using namespace ock::common::config;

namespace {

class TestServiceConfigure : public testing::Test {
public:
    void SetUp() override;
    void TearDown() override;
};

void TestServiceConfigure::SetUp() {}

void TestServiceConfigure::TearDown()
{
    GlobalMockObject::verify();
}

TEST_F(TestServiceConfigure, test_initialize_should_success)
{
    int ret = ServiceConfigure::GetInstance().Initialize();
    ASSERT_EQ(0, ret);
}

TEST_F(TestServiceConfigure, init_with_empty_path)
{
    auto &instance = ServiceConfigure::GetInstance();
    auto workPath = instance.GetWorkPath();
    ASSERT_NE("", workPath);

    instance.SetWorkPath("");

    auto ret = instance.Initialize();
    ASSERT_EQ(-1, ret);
    instance.SetWorkPath(workPath);
}

TEST_F(TestServiceConfigure, init_create_config_obj_failed)
{
    auto &instance = ServiceConfigure::GetInstance();

    MOCKER(Configuration::GetInstance<MemFsConfigure>)
        .stubs().will(returnValue(static_cast<ConfigurationPtr>(nullptr)));
    auto ret = instance.Initialize();
    ASSERT_EQ(-1, ret);
}

TEST_F(TestServiceConfigure, init_check_config_file_failed)
{
    auto &instance = ServiceConfigure::GetInstance();

    MOCKER(realpath).stubs().will(returnValue(static_cast<char *>(nullptr)));
    auto ret = instance.Initialize();
    ASSERT_EQ(-1, ret);
}

TEST_F(TestServiceConfigure, init_read_config_failed)
{
    auto &instance = ServiceConfigure::GetInstance();

    MOCKER(Configuration::ReadConf<MemFsConfigure>).stubs().will(returnValue(false));
    auto ret = instance.Initialize();
    ASSERT_EQ(-1, ret);
}

TEST_F(TestServiceConfigure, init_valid_config_failed)
{
    std::vector<std::string> mockErrors = { "01", "02" };
    auto &instance = ServiceConfigure::GetInstance();

    union MockerHelper {
        std::vector<std::string> (Configuration::*validate)(ValidatorTag tag);
        std::vector<std::string> (*mockValidate)(Configuration *self, ValidatorTag tag);
    };
    MockerHelper helper{};
    helper.validate = &Configuration::Validate;
    MOCKCPP_NS::mockAPI("&Configuration::Validate", helper.mockValidate).stubs().will(returnValue(mockErrors));
    auto ret = instance.Initialize();
    ASSERT_EQ(-1, ret);
}
}