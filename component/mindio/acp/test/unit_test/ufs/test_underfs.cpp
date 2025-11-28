/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.
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

#include "pacific_adapter.h"
#include "under_fs_factory.h"
#include "under_fs_manager.h"
#include "service_configure.h"

using namespace ock::common::config;
using namespace ock::ufs;

class TestUnderFS : public testing::Test {
public:
    void SetUp() override;
    void TearDown() override;

protected:
    static BaseFileService *ufs;
};

BaseFileService *TestUnderFS::ufs;

void TestUnderFS::SetUp()
{
    ufs = new (std::nothrow) ock::ufs::PacificAdapter("/tmp/test_underfs");
    ASSERT_TRUE(ufs != nullptr);
}

void TestUnderFS::TearDown()
{
    delete ufs;
    ufs = nullptr;
}

TEST_F(TestUnderFS, test_underfs_factory_default)
{
    std::shared_ptr<PacificAdapter> base = std::make_shared<PacificAdapter>("/tmp/test_underfs");
    UnderFsFactory::GetInstance().SetDefault(base);

    auto ret = UnderFsFactory::GetInstance().GetDefault();
    ASSERT_EQ(base, ret);
}

TEST_F(TestUnderFS, test_underfs_factory)
{
    std::shared_ptr<PacificAdapter> base = std::make_shared<PacificAdapter>("/tmp/test_underfs");
    UnderFsFactory::GetInstance().Set("1", base);

    auto ret = UnderFsFactory::GetInstance().Get("1");
    ASSERT_EQ(base, ret);
}

TEST_F(TestUnderFS, test_underfs_manager_noinst)
{
    auto &ufsConfig = ServiceConfigure::GetInstance().GetUnderFileSystemConfig();
    ASSERT_NE(nullptr, &ufsConfig);
    std::shared_ptr<PacificAdapter> base = std::make_shared<PacificAdapter>("/tmp/test_underfs");
    UnderFsFactory::GetInstance().Set(ufsConfig.defaultName, base);

    auto ret = UnderFsManager::GetInstance().Initialize();
    ASSERT_EQ(0, ret);

    UnderFsManager::GetInstance().Destroy();
}

TEST_F(TestUnderFS, test_underfs_manager)
{
    auto &ufsConfig = ServiceConfigure::GetInstance().GetUnderFileSystemConfig();
    ASSERT_NE(nullptr, &ufsConfig);
    auto ufs = const_cast<UnderFsConfig *>(&ufsConfig);
    ASSERT_NE(nullptr, ufs);
    std::string name = "pacific";
    std::string mount = "mount_path";
    std::map<std::string, std::string> opt;

    opt.insert(std::pair<std::string, std::string>(mount, name));
    opt.insert(std::pair<std::string, std::string>(name, mount));
    struct UnderFsInstance inst = {.name = name, .type = name, .options = opt};
    ufs->defaultName = name;
    ufs->instances[inst.name] = inst;

    std::shared_ptr<PacificAdapter> base = std::make_shared<PacificAdapter>("/tmp/test_underfs");
    UnderFsFactory::GetInstance().Set(ufs->defaultName, base);

    auto ret = UnderFsManager::GetInstance().Initialize();
    ASSERT_EQ(0, ret);

    UnderFsManager::GetInstance().Destroy();
}
