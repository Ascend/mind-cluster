/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
#include "acc_common_util.h"
#include "acc_includes.h"
#include "acc_tcp_server_default.h"

namespace ock {
namespace acc {
AccDecryptHandler AccTcpServer::decryptHandler_ = nullptr;

AccTcpServerPtr AccTcpServer::Create()
{
    auto server = AccMakeRef<AccTcpServerDefault>();
    if (server.Get() == nullptr) {
        LOG_ERROR("Failed to create AccTcpserverDefault, probably out of memory");
        return nullptr;
    }

    return server.Get();
}

void AccTcpServer::RegisterDecryptHandler(const AccDecryptHandler &h)
{
    ASSERT_RET_VOID(h != nullptr);
    ASSERT_RET_VOID(decryptHandler_ == nullptr);
    decryptHandler_ = h;
}
}  // namespace acc
}  // namespace ock