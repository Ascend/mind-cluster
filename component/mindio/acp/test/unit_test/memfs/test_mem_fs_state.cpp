/*
* Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.
*/
#include <gtest/gtest.h>

#define private public
#include "mem_fs_state.h"
#undef private


using namespace ock::memfs;

namespace {

TEST(TestMemFsState, test_memfs_set_and_get_state)
{
    auto &state = MemfsState::Instance();
    state.SetState(MemfsStateCode::PREPARING);
    auto result = state.GetState();
    ASSERT_EQ(MemfsStateCode::PREPARING, result.first);

    state.SetState(MemfsStateCode::STARTING);
    result = state.GetState();
    ASSERT_EQ(MemfsStateCode::STARTING, result.first);

    state.SetState(MemfsStateCode::RUNNING, MemfsStartProgress::FIFTY_PERCENT);
    result = state.GetState();
    ASSERT_EQ(MemfsStateCode::RUNNING, result.first);
    ASSERT_EQ(MemfsStartProgress::FIFTY_PERCENT, result.second);

    state.SetState(MemfsStateCode::PRE_EXITING);
    result = state.GetState();
    ASSERT_EQ(MemfsStateCode::PRE_EXITING, result.first);

    state.SetState(MemfsStateCode::EXITING);
    result = state.GetState();
    ASSERT_EQ(MemfsStateCode::EXITING, result.first);

    state.SetState(MemfsStateCode::EXITED);
    result = state.GetState();
    ASSERT_EQ(MemfsStateCode::EXITED, result.first);
}

TEST(TestMemFsState, test_get_memfs_state)
{
    ASSERT_EQ("PREPARING", MemfsState::GetMemfsState(MemfsStateCode::PREPARING));
    ASSERT_EQ("STARTING", MemfsState::GetMemfsState(MemfsStateCode::STARTING));
    ASSERT_EQ("RUNNING", MemfsState::GetMemfsState(MemfsStateCode::RUNNING));
    ASSERT_EQ("PRE_EXITING", MemfsState::GetMemfsState(MemfsStateCode::PRE_EXITING));
    ASSERT_EQ("EXITING", MemfsState::GetMemfsState(MemfsStateCode::EXITING));
    ASSERT_EQ("EXITED", MemfsState::GetMemfsState(MemfsStateCode::EXITED));
}
}