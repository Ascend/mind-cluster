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
#include "acc_includes.h"
#include "acc_tcp_request_context.h"

namespace ock {
namespace acc {
Result AccTcpRequestContext::Reply(int16_t result, const AccDataBufferPtr &d) const
{
    ASSERT_RETURN(d.Get() != nullptr, ACC_INVALID_PARAM);
    ASSERT_RETURN(link_.Get() != nullptr, ACC_LINK_ERROR);
    if (UNLIKELY(!link_->Established())) {
        LOG_ERROR("Failed to send reply message with message type " << header_.type <<
                  ", seqlo " << header_.seqNo << " as the link is broken");
        return ACC_LINK_ERROR;
    }
    AccMsgHeader replyHeader(header_.type, result, d->DataLen(), header_.seqNo);
    return link_->EnqueueAndModifyEpoll(replyHeader, d, nullptr);
}
}  // namespace acc
}  // namespace ock