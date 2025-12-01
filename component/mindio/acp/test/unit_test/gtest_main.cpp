/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2022-2024. All rights reserved.
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

#include "hlog.h"
#include "common/test_hlog.h"
#include "service_configure.h"

using namespace ock::common::config;
using namespace ock::hlog;

static int GetWorkPath(std::string &configPath)
{
    std::string linkedPath = "/proc/" + std::to_string(getpid()) + "/exe";
    char realPath[PATH_MAX];
    auto size = readlink(linkedPath.c_str(), realPath, sizeof(realPath));
    if (size < 0 || size >= PATH_MAX) {
        return -1;
    }

    realPath[size] = '\0';
    std::string path{ realPath };
    std::string tempPath = std::move(path.substr(0, path.find("/test/build/dfs_hdt")));
    configPath = tempPath + "/output";
    return 0;
}

int main(int argc, char **argv)
{
    std::string configPath;
    int ret = GetWorkPath(configPath);
    if (ret == -1) {
        std::cerr << "Failed get config path.";
        return -1;
    }
    std::cout << "Get config path: " << configPath << std::endl;
    ServiceConfigure::GetInstance().SetWorkPath(configPath);
    TestHlog::LoggerInitialize();
    HLOG_INFO("Begin ockd unit test");
    testing::InitGoogleTest(&argc, argv);
    ret = RUN_ALL_TESTS();
    if (Hlog::gLogger != nullptr) {
        Hlog::gLogger->Flush();
    }
    HLOG_INFO("Finished ockd unit test " << ret);
    Hlog::DestroyLogInstances();

    return ret;
}
