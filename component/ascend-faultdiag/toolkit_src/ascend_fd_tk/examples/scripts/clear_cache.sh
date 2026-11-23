#!/usr/bin/env bash
# -*- coding:utf-8 -*-
# Copyright 2026 Huawei Technologies Co., Ltd
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# ==============================================================================
# 清除缓存目录脚本

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
CACHE_DIR="${SCRIPT_DIR}/../../cache"

if [ -d "$CACHE_DIR" ]; then
    echo "正在清理缓存目录：$CACHE_DIR"

    # 删除cache目录下的所有文件和目录
    for item in "$CACHE_DIR"/*; do
        if [ -e "$item" ]; then
            rm -rf "$item" 2>/dev/null
        fi
    done

    # 重新创建bmc_dump_cache和parse_cache空文件夹
    mkdir -p "$CACHE_DIR/host_dump_cache" 2>/dev/null
    mkdir -p "$CACHE_DIR/bmc_dump_cache" 2>/dev/null
    mkdir -p "$CACHE_DIR/switch_cli_output_cache" 2>/dev/null


    echo "缓存目录清理完成，已重新创建空文件夹"
else
    echo "缓存目录不存在：$CACHE_DIR"
fi
