/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved.
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
#ifndef OCK_DFS_MEMFS_LOGGER_H
#define OCK_DFS_MEMFS_LOGGER_H

#include "memfs_out_logger.h"

#include "error_code.h"

using namespace ock::memfs;

#define MFS_LOG_DEBUG(ARGS) DAGGER_LOG_DEBUG(MFS, ARGS)
#define MFS_LOG_INFO(ARGS) DAGGER_LOG_INFO(MFS, ARGS)
#define MFS_LOG_WARN(ARGS) DAGGER_LOG_WARN(MFS, ARGS)
#define MFS_LOG_ERROR(ARGS) DAGGER_LOG_ERROR(MFS, ARGS)

#endif // OCK_DFS_MEMFS_LOGGER_H
