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

#include "ipc_message.h"
#include "fs_server.h"

using namespace ock::memfs;

class TestFsServer : public testing::Test {};

TEST_F(TestFsServer, InputPathValid_simple)
{
    std::string path = "/hello/world.txt";
    std::string realpath;
    auto valid = ShellFSServer::InputPathValid(path, realpath);

    EXPECT_TRUE(valid) << "path : " << path;
}

TEST_F(TestFsServer, InputPathValid_path_end_with_slash)
{
    std::string path = "/hello/world.txt/";
    std::string realpath;
    auto valid = ShellFSServer::InputPathValid(path, realpath);

    EXPECT_TRUE(valid) << "path : " << path;
}

TEST_F(TestFsServer, InputPathValid_path_no_start_slash)
{
    std::string path = "hello/world.txt";
    std::string realpath;
    auto valid = ShellFSServer::InputPathValid(path, realpath);

    EXPECT_TRUE(valid) << "path : " << path;
}

TEST_F(TestFsServer, InputPathValid_path_repeated_slash)
{
    std::string path = "/hello//world.txt";
    std::string realpath;
    auto valid = ShellFSServer::InputPathValid(path, realpath);

    EXPECT_TRUE(valid) << "path : " << path;
}

TEST_F(TestFsServer, InputPathValid_path_contain_b)
{
    std::string path = "/hello/wo\brld.txt";
    std::string realpath;
    auto valid = ShellFSServer::InputPathValid(path, realpath);

    EXPECT_FALSE(valid) << "path : " << path;
}

TEST_F(TestFsServer, InputPathValid_path_contain_n)
{
    std::string path = "/hello/wo\nrld.txt";
    std::string realpath;
    auto valid = ShellFSServer::InputPathValid(path, realpath);

    EXPECT_FALSE(valid) << "path : " << path;
}

TEST_F(TestFsServer, InputPathValid_path_contain_r)
{
    std::string path = "/hello/wo\rrld.txt";
    std::string realpath;
    auto valid = ShellFSServer::InputPathValid(path, realpath);

    EXPECT_FALSE(valid) << "path : " << path;
}

TEST_F(TestFsServer, InputPathValid_path_contain_dot)
{
    std::string path = "/hello/./world.txt";
    std::string realpath;
    auto valid = ShellFSServer::InputPathValid(path, realpath);

    EXPECT_FALSE(valid) << "path : " << path;
}

TEST_F(TestFsServer, InputPathValid_path_contain_2dot)
{
    std::string path = "/hello/../world.txt";
    std::string realpath;
    auto valid = ShellFSServer::InputPathValid(path, realpath);

    EXPECT_FALSE(valid) << "path : " << path;
}

TEST_F(TestFsServer, InputPathValid_path_hidden_file)
{
    std::string path = "/hello/.world.txt";
    std::string realpath;
    auto valid = ShellFSServer::InputPathValid(path, realpath);

    EXPECT_TRUE(valid) << "path : " << path;
}

TEST_F(TestFsServer, InputPathValid_path_hidden_dir)
{
    std::string path = "/.hello/world.txt";
    std::string realpath;
    auto valid = ShellFSServer::InputPathValid(path, realpath);

    EXPECT_TRUE(valid) << "path : " << path;
}

TEST_F(TestFsServer, InputPathValid_path_contains_chinese)
{
    std::string path = "/hello/这里有中文world.txt";
    std::string realpath;
    auto valid = ShellFSServer::InputPathValid(path, realpath);

    EXPECT_TRUE(valid) << "path : " << path;
}

TEST_F(TestFsServer, PrintableString_simple)
{
    std::string path = "/hello/world.txt";
    auto result = PrintableString(path);

    EXPECT_EQ(path, result);
}

TEST_F(TestFsServer, PrintableString_bnr)
{
    std::string path = "/hello/\b\n\rworld.txt";
    std::string expect = "/hello/\\0x08\\0x0a\\0x0dworld.txt";
    auto result = PrintableString(path);

    EXPECT_EQ(expect, result);
}

TEST_F(TestFsServer, PrintableString_chinese)
{
    std::string path = "/hello/中文_world.txt";
    std::string expect = "/hello/\\0xe4\\0xb8\\0xad\\0xe6\\0x96\\0x87_world.txt";
    auto result = PrintableString(path);

    EXPECT_EQ(expect, result);
}