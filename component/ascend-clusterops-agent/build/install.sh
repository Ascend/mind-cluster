#!/usr/bin/env bash
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

# 在用户机上安装 kubectl 插件 (kubectl-ascend_diag / kubectl-clusterops)。
set -euo pipefail

cd "$(dirname "$0")"

chmod +x kubectl-ascend_diag kubectl-clusterops
# 清理旧包安装的 (连字符命名) 残留二进制, 避免 kubectl 解析到错误文件
sudo rm -f /usr/local/bin/kubectl-ascend-diag
# kubectl 插件按 `kubectl <name>` 找 PATH 下 kubectl-<name>; 文件名即插件名, 同名安装即可
sudo cp kubectl-ascend_diag /usr/local/bin/kubectl-ascend_diag
sudo cp kubectl-clusterops /usr/local/bin/kubectl-clusterops

echo "installed. 验证:"
if ! kubectl ascend-diag --help >/dev/null 2>&1 || ! kubectl clusterops --help >/dev/null 2>&1; then
    echo "kubectl 没找到插件, 检查 /usr/local/bin 是否在 PATH" >&2
    exit 1
fi
echo "OK: kubectl ascend-diag / kubectl clusterops 可用"
