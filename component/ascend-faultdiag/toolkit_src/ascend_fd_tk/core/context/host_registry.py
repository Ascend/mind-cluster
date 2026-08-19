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

"""HostInfo 代际注册表基础设施。

独立模块（不合并到 register.py），避免 host.py 与 host_a5.py 之间的循环导入，
以及避免与 register.py 内 Analyzer/Collector 等 import 形成跨层循环：
    host.py -> register.py -> Analyzer -> ClusterInfoCache -> host.py

各代际模型模块（host.py / host_a5.py / 未来的 host_a6.py）从此处导入
register_host_info 装饰器，注册各自的 HostInfo 子类。

消费方（LoadCache 等）通过 HOST_INFO_REGISTRY 按代际标识查找反序列化类，
新增代际无需修改消费方代码，符合开闭原则。
"""

from typing import Callable, Dict, Type

from ascend_fd_tk.core.common.json_obj import JsonObj


HOST_INFO_REGISTRY: Dict[str, Type[JsonObj]] = {}


def register_host_info(generation: str) -> Callable[[Type[JsonObj]], Type[JsonObj]]:
    """HostInfo 代际注册装饰器。

    Args:
        generation: 代际标识字符串，如 "A3" / "A5" / "A6"
    """

    def wrap(cls: Type[JsonObj]) -> Type[JsonObj]:
        HOST_INFO_REGISTRY[generation] = cls
        return cls

    return wrap
