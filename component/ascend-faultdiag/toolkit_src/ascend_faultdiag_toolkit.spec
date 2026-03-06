#!/usr/bin/env python3
# -*- coding: utf-8 -*-
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

# -*- mode: python ; coding: utf-8 -*-

# 确保导入 Tree
from PyInstaller.building.datastruct import Tree


# 手动将 Tree 转换为 datas 需要的格式
def convert_tree_to_datas(source_dir, target_dir):
    tree = Tree(source_dir, prefix=target_dir)
    # Tree 返回 (dest, src, ty)，需要转换为 (src, dest) 格式
    return [(src, os.path.dirname(dest)) for dest, src, ty in tree]


a = Analysis(
    ['ascend_fd_tk\\cli.py'],
    pathex=[],
    binaries=[],
    datas=[
        # 配置文件
        *convert_tree_to_datas('ascend_fd_tk/core/config', 'ascend_fd_tk/core/config'),
        # prase_script要传到远端
        *convert_tree_to_datas('ascend_fd_tk/core/log_parser', 'ascend_fd_tk/core/log_parser'),
        # 注册制需要源文件
        *convert_tree_to_datas('ascend_fd_tk/core/fault_analyzer', 'ascend_fd_tk/core/fault_analyzer'),
        # 注册制需要源文件
        *convert_tree_to_datas('ascend_fd_tk/core/inspection', 'ascend_fd_tk/core/inspection')
    ],
    hiddenimports=[],
    hookspath=[],
    hooksconfig={},
    runtime_hooks=[],
    excludes=[],
    noarchive=False,
    optimize=0,
)
pyz = PYZ(a.pure)

exe = EXE(
    pyz,
    a.scripts,
    a.binaries,
    a.datas,
    [],
    name='ascend-faultdiag-toolkit',
    debug=False,
    bootloader_ignore_signals=False,
    strip=False,
    upx=True,
    upx_exclude=[],
    runtime_tmpdir=None,
    console=True,
    disable_windowed_traceback=False,
    argv_emulation=False,
    target_arch=None,
    codesign_identity=None,
    entitlements_file=None,
    icon=['doc\\picture\\ascend_logo.ico'],
)
