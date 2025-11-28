/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
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
#include "bmm.h"

namespace ock {
namespace memfs {
MResult MemFsBMM::Initialize(const MemFsBMMOptions &opt)
{
    std::lock_guard<std::mutex> guard(mMutex);
    if (mInited) {
        return MFS_OK;
    }

    auto result = mBmmPool.Initialize(opt);
    if (result != MFS_OK) {
        return result;
    }

    mInited = true;
    return MFS_OK;
}

void MemFsBMM::UnInitialize()
{
    std::lock_guard<std::mutex> guard(mMutex);
    if (!mInited) {
        return;
    }

    mBmmPool.UnInitialize();
    mInited = false;
}
}
}